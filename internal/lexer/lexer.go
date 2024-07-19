// langur/lexer/lexer.go

package lexer

import (
	"fmt"
	"langur/cpoint"
	"langur/modes"
	"langur/str"
	"langur/token"
)

// NOTE: Each branch of the lexer is responsible for advancing the code point position.

const (
	INIT_LINE_POSITION = 1
	INIT_LINE          = 1

	IMPLIED_EXPRESSION_TERMINATOR_LITERAL = "(;)"

	identifierMaxBytes = 127
)

type Lexer struct {
	// The Lexer currently works with UTF-8 only.
	input string // input source code string

	bytePosition int  // current byte position (not code point) in source code string
	peekPosition int  // next byte position to start reading at (might be more than 1 byte ahead to read UTF-8)
	cp           rune // current code point
	peekCp       rune // peek to the next code point
	peekCpWidth  int  // code unit width of next code point

	// Are we there yet?
	EOF     bool
	peekEOF bool

	// for sub-lexing
	tokenQueue []token.Token
	tqPosition int

	// Looking back...
	previousToken token.Token

	// added to each token for error reporting
	fileName       string
	line           int
	cpPosition     int // code point position (total)
	cpLinePosition int // code point position on the line

	Modes *modes.CompileModes
}

func New(input, fileName string, m *modes.CompileModes) (lex *Lexer, err error) {
	defer func() {
		if p := recover(); p != nil {
			err = fmt.Errorf("Lexer.New() panic: %s", p)
		}
	}()

	if m == nil {
		m = modes.NewCompileModes()
	}

	lex = &Lexer{input: input,
		fileName:       fileName,
		line:           INIT_LINE,
		cpLinePosition: -1,
		cpPosition:     -1, // -1 as they will be advanced twice to get the Lexer started
		Modes:          m,
	}

	lex.advanceCodePoint()
	lex.advanceCodePoint()

	return
}

func NewWithPosition(
	input, fileName string,
	line, cpPosition, cpLinePosition int,
	m *modes.CompileModes) (lex *Lexer, err error) {

	lex, err = New(input, fileName, m)
	lex.line = line
	lex.cpPosition = cpPosition
	lex.cpLinePosition = cpLinePosition

	return
}

func bug(fnName, description string) {
	// panic for now...
	// panic caught by New() or NextToken()
	panic("Lexer bug: " + description)
}

func (lex *Lexer) queueToken(tok token.Token) {
	lex.tokenQueue = append(lex.tokenQueue, tok)
}

func (lex *Lexer) queueImpliedSemicolon() {
	lex.tokenQueue = append(lex.tokenQueue, impliedSemicolon(lex.line, lex.cpLinePosition))
}

func impliedSemicolon(line, linePosition int) token.Token {
	return token.Token{
		Type: token.SEMICOLON, Literal: IMPLIED_EXPRESSION_TERMINATOR_LITERAL,
		Line: line, LinePosition: linePosition}
}

func lexFrom(
	input, fileName string,
	line, cpPosition, cpLinePosition int,
	until []token.Type,
	m *modes.CompileModes) (
	literal string, tokens []token.Token, err error) {

	var lex *Lexer
	lex, err = NewWithPosition(input, fileName, line, cpPosition, cpLinePosition, m)
	if err != nil {
		return
	}
	if lex.EOF {
		err = fmt.Errorf("Nothing to lex")
		return
	}

	scope := []token.Type{}
	errStr := ""
	var tok token.Token

	for !lex.EOF {
		tok, err = lex.NextToken()
		if err != nil {
			return
		}

		errs := tok.Errors(", ")

		tokens = append(tokens, tok)
		if errs != "" {
			if errStr != "" {
				errStr += "; "
			}
			errStr += errs
		}

		if until != nil {
			if len(scope) == 0 && token.InTypeSlice(tok.Type, until) {
				break
			}
			// push and pop scope for () [] {}
			if tok.Type == token.LPAREN || tok.Type == token.LBRACKET || tok.Type == token.LBRACE {
				scope = append(scope, token.Closer(tok.Type))
			} else if len(scope) > 0 && tok.Type == scope[len(scope)-1] {
				scope = scope[:len(scope)-1]
			}
		}
	}
	if errStr != "" {
		err = fmt.Errorf(errStr)
	}

	literal = lex.input[:lex.bytePosition]
	return
}

func LexString(input, fileName string, m *modes.CompileModes) (tokens []token.Token, err error) {
	_, tokens, err = lexFrom(input, fileName, 1, 1, 1, nil, m)
	return
}

// NextToken() is a long function, but don't panic.  :)
func (lex *Lexer) NextToken() (tok token.Token, err error) {
	defer func() {
		if p := recover(); p != nil {
			err = fmt.Errorf("Lexer.NextToken() panic: %s", p)
		}
	}()

	// Functions other than NextToken() should return errors (and not panic).
	addError := func(tp *token.Token, err error) {
		tp.AddTokenErr(err)
	}

	checkDBop := func() {
		// db operator?
		// ? mark immediately following operator without spacing
		if lex.cp == '?' {
			if token.MayBeUsedAsDbOperator(tok.Type) {
				tok.Code |= token.CODE_DB_OPERATOR
			} else {
				addError(&tok, fmt.Errorf("Illegal DB operator"))
			}
			tok.Literal = tok.Literal + "?"
			lex.advanceCodePoint()
		}
	}

	position := lex.bytePosition
	cpPosition := lex.cpPosition

	// QUEUED TOKENS
	// It must read any queued tokens first.
	if lex.tqPosition < len(lex.tokenQueue) {
		tok = lex.tokenQueue[lex.tqPosition]
		lex.tqPosition++
		lex.previousToken = tok
		return

	} else {
		// Empty the queue and move on.
		lex.tokenQueue = nil
		lex.tqPosition = 0
	}

	// skip whitespace first
	var wsNewlineCount int
	tok.CpDiff, wsNewlineCount, err = lex.skipWhiteSpace()
	tok.NewLinePrecedes = wsNewlineCount != 0
	if err != nil {
		addError(&tok, err)
	}

	tok.Type = token.INVALID
	tok.Code = token.CODE_DEFAULT
	tok.Line = lex.line
	tok.LinePosition = lex.cpLinePosition

	if lex.EOF {
		// end of file
		tok.Literal, tok.Type = "", token.EOF
		// the one case that does not attempt to advance the code point

	} else {
		tok.Literal = string(lex.cp)
		switch lex.cp {

		case '!':
			if lex.peekCp == '=' {
				tok.Literal, tok.Type = "!=", token.NOT_EQUAL
				lex.advanceCodePoint()

			} else {
				tok.Type = token.INVALID
			}
			lex.advanceCodePoint()
			checkDBop()

		case '(':
			tok.Type = token.LPAREN
			lex.advanceCodePoint()

			if tok.CpDiff == 0 &&
				!token.MayPrecedeOpeningParenthesisWithoutSpacing(lex.previousToken) {

				addError(&tok, fmt.Errorf("Cannot use opening parenthesis without spacing after a %s token", token.TypeDescription(lex.previousToken.Type)))
			}

		case ')':
			tok.Type = token.RPAREN
			lex.advanceCodePoint()

		case '<':
			if lex.peekCp == '=' {
				tok.Literal, tok.Type = "<=", token.LT_OR_EQUAL
				lex.advanceCodePoint()
			} else {
				tok.Type = token.LESS_THAN
			}
			lex.advanceCodePoint()
			checkDBop()

		case '>':
			if lex.peekCp == '=' {
				tok.Literal, tok.Type = ">=", token.GT_OR_EQUAL
				lex.advanceCodePoint()
			} else {
				tok.Type = token.GREATER_THAN
			}
			lex.advanceCodePoint()
			checkDBop()

		case '{':
			tok.Type = token.LBRACE
			lex.advanceCodePoint()

		case '}':
			tok.Type = token.RBRACE
			lex.advanceCodePoint()

		case '[':
			tok.Type = token.LBRACKET
			lex.advanceCodePoint()

		case ']':
			tok.Type = token.RBRACKET
			lex.advanceCodePoint()

		case '+':
			if lex.peekCp == '+' {
				// might have future use; better to call it illegal for now
				tok.Literal, tok.Type = "++", token.INVALID
				lex.advanceCodePoint()
			} else {
				tok.Type = token.PLUS
			}
			lex.advanceCodePoint()

		case '-':
			if lex.peekCp == '-' {
				// might have future use; better to call it illegal for now
				// also to not take as sequential negation tokens
				tok.Literal, tok.Type = "--", token.INVALID
				lex.advanceCodePoint()

			} else if lex.peekCp == '>' {
				tok.Literal, tok.Type = "->", token.FORWARD
				lex.advanceCodePoint()

			} else {
				tok.Type = token.MINUS
			}
			lex.advanceCodePoint()

		case '*':
			tok.Type, tok.Literal = token.ASTERISK, "*"
			tok.AddImpliedSemicolonAtNewLine = lex.previousToken.Type == token.MODULE
			lex.advanceCodePoint()

		case '/':
			if lex.peekCp == '/' {
				tok.Literal, tok.Type = `//`, token.DOUBLESLASH
				lex.advanceCodePoint()
			} else {
				tok.Type = token.SLASH
			}
			lex.advanceCodePoint()

		case '\\':
			tok.Type = token.BACKSLASH
			lex.advanceCodePoint()

		case '^':
			if lex.peekCp == '/' {
				tok.Literal, tok.Type = "^/", token.ROOT
				lex.advanceCodePoint()
			} else {
				tok.Type = token.POWER
			}
			lex.advanceCodePoint()

		case ',':
			tok.Type = token.COMMA
			lex.advanceCodePoint()

		case ';':
			tok.Type = token.SEMICOLON
			lex.advanceCodePoint()

		case ':':
			tok.Type = token.COLON
			lex.advanceCodePoint()

		case '=':
			if lex.peekCp == '=' {
				tok.Literal, tok.Type = "==", token.EQUAL
				lex.advanceCodePoint()
			} else {
				tok.Type = token.ASSIGN
			}
			lex.advanceCodePoint()
			checkDBop()

		case '.':
			if lex.peekCp == '.' {
				lex.advanceCodePoint()

				if lex.peekCp == '.' {
					// ... 3 dots
					tok.Literal, tok.Type = "...", token.EXPANSION
					lex.advanceCodePoint()
					lex.advanceCodePoint()

					if lex.cp == '.' {
						addError(&tok, fmt.Errorf("Unexpected 4 dots in row"))
					}

				} else {
					// .. 2 dots
					tok.Literal, tok.Type = "..", token.RANGE
					lex.advanceCodePoint()
				}

			} else if cpoint.IsWordTokenChar(lex.peekCp) {
				// .var
				lex.advanceCodePoint() // past the dot

				tok.Literal, err = lex.readWord()
				if err != nil {
					addError(&tok, err)
				}
				// tok.Type = token.IDENT

				err = fmt.Errorf(".var syntax not supported; use var without dot")
				addError(&tok, err)

			} else {
				tok.Type = token.INVALID
				err = fmt.Errorf("Dot . cannot stand on its own")
				addError(&tok, err)
				lex.advanceCodePoint()
			}

		case '_':
			if cpoint.IsWordTokenChar(lex.peekCp) {
				// _systemIdentifier
				// underscore part of system identifier
				tok.Literal, err = lex.readWord()
				if err != nil {
					addError(&tok, err)
				}
				tok.Type = token.IDENT

			} else {
				tok.Literal, tok.Type = "_", token.NONE
				lex.advanceCodePoint()
			}

		case '~': // tilde
			tok.Type = token.APPEND
			lex.advanceCodePoint()

		case '"':
			tok.Literal, tok.Type, tok.Attachments, _, err =
				lex.readStringLiteral(true, true, false, false, false, false, "")

			if err != nil {
				addError(&tok, err)
			}

		case '\'':
			if token.MayPrecedeShorthandStringIndexing(lex.previousToken.Type) &&
				tok.CpDiff == 0 {

				err = lex.readShortHandStringIndex(&tok)
				if err != nil {
					addError(&tok, err)
				}

			} else {
				tok.Literal, tok.Type, err = lex.readCodePointLiteral(false)
				if err != nil {
					addError(&tok, err)
				}
			}

		default:
			if lex.cp >= '0' && lex.cp <= '9' {
				// starts with a base 10 digit
				// must test for this before testing if IsTokenWordChar()
				var code int
				tok.Literal, tok.Type, code, err = lex.readNumber()
				tok.Code |= code

				if err != nil {
					addError(&tok, err)
				}

			} else if cpoint.IsWordTokenChar(lex.cp) {
				err = lex.readAndInterpretWordToken(&tok, cpPosition)
				if err != nil {
					addError(&tok, err)
					err = nil
				}
				checkDBop()

			} else {
				// Anything not defined is not a valid token.
				tok.Type = token.INVALID
				lex.advanceCodePoint()
			}
		}
	}

	if tok.Type == token.IDENT && len(tok.Literal) > identifierMaxBytes {
		err = fmt.Errorf("Identifier maximum bytes (%d) exceeded (%q)",
			identifierMaxBytes, str.ReformatInput(tok.Literal))
		addError(&tok, err)
	}

	// check for combination operator
	if lex.cp == '=' && lex.peekCp != '=' && token.MayBeUsedOnCombinationOperator(tok.Type) {
		lex.advanceCodePoint()
		tok.Code |= token.CODE_COMBINATION_ASSIGNMENT_OPERATOR
		tok.Literal += "="

	} else {
		// using newline as default expression terminator
		if tok.NewLinePrecedes &&
			token.ImpliedExprTerminatorIfFollowedByNewline(lex.previousToken) {

			// Queue this token to pick up it up next time so we can add our implied expression terminator here.
			lex.queueToken(tok)
			tok = impliedSemicolon(
				lex.previousToken.Line,
				lex.previousToken.LinePosition+len(lex.previousToken.Literal),
			)
		}
	}

	// NOTE: Rather than have a call to l.advanceCodePoint() at the end, with some toggle not to advance sometimes, ...
	// ... we make each branch responsible for advancing the input string position as it reads.
	// It's actually simpler that way.

	// If not every call to NextToken() advances at least one code point, ...
	// ...you could get into an infinite loop, which is a lot of fun.
	if !lex.EOF && lex.bytePosition == position {
		bug("NextToken", "Call to next token failed to advance the token position")
		// panic if we haven't...
		panic("Call to next token failed to advance the token position")
	}

	lex.previousToken = tok
	return
}

func (lex *Lexer) skipLineReturnOnly() bool {
	if lex.cp == '\r' && lex.peekCp == '\n' && cpoint.IsTokenVerticalSpace('\r') {
		lex.advanceCodePoint()
		lex.advanceCodePoint()
		return true
	} else if cpoint.IsTokenVerticalSpace(lex.cp) {
		lex.advanceCodePoint()
		return true
	}
	return false
}

// skip comments and whitespace
// returns the difference in code points
func (lex *Lexer) skipWhiteSpace() (cpDiff, newlineCount int, err error) {
	cpPosition := lex.cpPosition
	commentNewlineCount := 0

	isCommentStart := func() bool {
		return lex.cp == '#' || lex.cp == '/' && lex.peekCp == '*'
	}

	for cpoint.IsTokenSpace(lex.cp) || isCommentStart() {
		if isCommentStart() {
			commentNewlineCount, err = lex.skipComments()
			newlineCount += commentNewlineCount

		} else if cpoint.IsTokenVerticalSpace(lex.cp) {
			if lex.cp == '\r' && lex.peekCp == '\n' {
				// accounting for \r\n as single line return
				lex.advanceCodePoint()
			}
			lex.advanceCodePoint()
			newlineCount++

		} else {
			lex.advanceCodePoint()
		}
	}

	cpDiff = lex.cpPosition - cpPosition
	return
}

func (lex *Lexer) skipComments() (newlineCount int, err error) {
	isInline := false
	position := lex.bytePosition

	if lex.cp == '/' && lex.peekCp == '*' {
		isInline = true
		lex.advanceCodePoint()
	} else if lex.cp != '#' {
		return
	}
	lex.advanceCodePoint()

	any := false

	for {
		if lex.EOF {
			if isInline {
				if err == nil {
					err = fmt.Errorf("Pattern EOF reached without closing inline comment")
				}
			}
			break

		} else if !any && !cpoint.IsAllowedInComments(lex.cp) {
			err = fmt.Errorf("Comment contains character(s) not allowed (%s)", cpoint.Display(lex.cp))
			lex.advanceCodePoint()

		} else if cpoint.IsVerticalSpace(lex.cp) {
			if !cpoint.IsTokenVerticalSpace(lex.cp) {
				err = fmt.Errorf("Non-token vertical space (%s) not allowed in comments", cpoint.Display(lex.cp))
			}

			if lex.cp == '\r' && lex.peekCp == '\n' {
				// accounting for \r\n as single line return
				lex.advanceCodePoint()
			}
			lex.advanceCodePoint()
			newlineCount++

			if !isInline {
				// end of line comment
				break
			}

		} else if lex.cp == '*' && lex.peekCp == '/' {
			lex.advanceCodePoint()
			lex.advanceCodePoint()
			if isInline {
				// end of /* inline comment */
				break
			}
			// keep going; */ does not end single line comment
		} else {
			lex.advanceCodePoint()
		}
	}

	if err == nil && !str.Balanced(lex.input[position:lex.bytePosition]) {
		err = fmt.Errorf("Comment contains unbalanced code points, such as LTR/RTL opening markers without matching closing markers")
	}

	return
}

// Lexer.advanceCodePoint()
// move the Lexer one code point forward
func (lex *Lexer) advanceCodePoint() {
	var width int
	var err error

	startingNextLine := false
	if lex.cp == '\n' {
		// previous code point a newline
		startingNextLine = true
	}

	// Don't waste time decoding the same thing twice....
	lex.cp, width, lex.EOF = lex.peekCp, lex.peekCpWidth, lex.peekEOF

	lex.bytePosition = lex.peekPosition
	lex.peekPosition += width // width == number of code units in UTF-8 string

	// counting code points, not code units, for line position reporting
	lex.cpLinePosition++
	lex.cpPosition++

	if startingNextLine {
		lex.line++
		lex.cpLinePosition = INIT_LINE_POSITION
	}

	// read in the next code point (peek ahead)
	lex.peekCp, lex.peekCpWidth, err = cpoint.Decode(&lex.input, lex.peekPosition)
	lex.peekEOF = cpoint.IsEOF(err)

	if lex.cp == '\r' {
		if err == nil && lex.peekCp != '\n' {
			err = fmt.Errorf("contains carriage return (\\r) not followed by linefeed (\\n)")
		}
	}

	if !lex.peekEOF && err != nil {
		// NOTE: panic to be caught by New() or NextToken()
		panic("Error reading from input string: " + err.Error())
	}
}
