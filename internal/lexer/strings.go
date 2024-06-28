// langur/lexer/strings.go

package lexer

import (
	"bytes"
	"fmt"
	"langur/cpoint"
	"langur/str"
	"langur/token"
	"strings"
)

const allowEndingBlockQuoteMarkerOffBeginningOfLine = true

func (lex *Lexer) readFreeWordList(tok *token.Token) (err error) {
	// free word list, such as fw/abcd jkl wer/
	// which equates to ["abcd", "jkl", "wer"]
	allowEsc := token.InterpretEscapeSequences(tok.Literal)

	modifiers, blockQuoteMarker, err2 := lex.readStringModifiers()
	if err2 != nil {
		err = err2
		return
	}
	any := false
	for _, mod := range modifiers {
		switch mod {
		default:
			err = fmt.Errorf("Invalid free word list modifier: %s",
				str.ReformatInput(mod))
			return
		}
	}

	allowNewLines := true //blockQuoteMarker != ""

	var strs []string
	strs, err = lex.readFreeWordListLiteral(allowEsc, allowNewLines, any, blockQuoteMarker)

	if err != nil {
		return

	} else {
		// having received a slice of strings, we make tokens for the parser to build a list
		// first token: opening LBRACKET
		tok.Literal, tok.Type = "([)", token.LBRACKET

		// the rest of the tokens to be queued up
		for i, s := range strs {
			lex.queueToken(token.Token{Literal: s, Type: token.STRING, Line: tok.Line, LinePosition: tok.LinePosition})
			if i < len(strs)-1 {
				lex.queueToken(token.Token{Literal: "(,)", Type: token.COMMA, Line: tok.Line, LinePosition: tok.LinePosition})
			}
		}
		lex.queueToken(token.Token{Literal: "(])", Type: token.RBRACKET})
	}

	if lex.cp != ',' && blockQuoteMarker != "" {
		lex.queueImpliedSemicolon()
	}

	return
}

func (lex *Lexer) readRe2Regex(tok *token.Token) (err error) {
	// re2 regex literal, such as re(some re2 regex using parentheses as quote marks)
	// re: allowing langur escape codes
	// RE: without langur escape codes
	allowEsc := token.InterpretEscapeSequences(tok.Literal)
	allowNewLines := true

	modifiers, blockQuoteMarker, err2 := lex.readStringModifiers()
	if err2 != nil {
		err = err2
		return
	}

	any := false
	lead := false
	marks := false
	list := ""
	neg := "-"
	maybeInterpolated := true
	escInterpolations := false
	freeSpacing := false

	for _, mod := range modifiers {
		switch mod {
		case "i", "m", "s", "U", "x":
			if strings.Contains(list, mod) || strings.Contains(neg, mod) {
				err = fmt.Errorf("Unexpected repeat of re2 modifier: %s", mod)
				return
			}
			// https://github.com/google/re2/wiki/Syntax
			// x: free-spacing mode added for langur; not a normal part of re2
			list += mod

			if mod == "x" {
				freeSpacing = true
			}

		case "any":
			if any {
				err = fmt.Errorf(`Unexpected repeat of "any" modifier`)
				return
			}
			any = true

		case "lead":
			if lead {
				err = fmt.Errorf(`Unexpected repeat of "lead" modifier`)
				return
			}
			lead = true

		case "esc":
			if !maybeInterpolated {
				err = fmt.Errorf(`Cannot combine "esc" and "ni" modifiers`)
				return
			}
			if escInterpolations {
				err = fmt.Errorf(`Unexpected repeat of "esc" modifier`)
				return
			}
			escInterpolations = true

		case "ni":
			if escInterpolations {
				err = fmt.Errorf(`Cannot combine "esc" and "ni" modifiers`)
				return
			}
			if !maybeInterpolated {
				err = fmt.Errorf(`Unexpected repeat of "ni" modifier`)
				return
			}
			maybeInterpolated = false

		default:
			err = fmt.Errorf("Invalid regex literal/re2 modifier: %s", str.ReformatInput(mod))
			return
		}
	}

	// auto-negate all regex modifiers to make safe for interpolation into a regex literal
	for _, m := range "smiUx" {
		if !strings.ContainsRune(list, rune(m)) && !strings.ContainsRune(neg, rune(m)) {
			neg += string(m)
		}
	}

	if neg != "-" {
		list += neg
	}

	tok.Literal, _, tok.Attachments, _, err =
		lex.readStringLiteral(allowEsc, maybeInterpolated, allowNewLines, any, lead, marks, blockQuoteMarker)

	if err != nil {
		return
	}
	tok.Type = token.REGEX_RE2

	if tok.Literal != "" {
		if escInterpolations {
			tok.Code |= token.CODE_ESC_ALL_INTERPOLATION
		}

		if list != "" {
			plus := ""
			if freeSpacing &&
				len(tok.Literal) > 0 && tok.Literal[len(tok.Literal)-1] != '\n' {
				// for free-spacing mode, get past any line comments to ensure enclosing parenthesis is not eaten by them
				plus = "\n"
			}

			// surround regex pattern with modifiers
			tok.Literal = "(?" + list + ":" + tok.Literal + plus + ")"

			// ... do same with attachments
			if tok.Attachments != nil {
				// add to first attachment (string)
				tok.Attachments[0] = "(?" + list + ":" + tok.Attachments[0].(string)

				// add to last attachment (string)
				tok.Attachments[len(tok.Attachments)-1] =
					tok.Attachments[len(tok.Attachments)-1].(string) + plus + ")"
			}
		}
	}

	if lex.cp != ',' && blockQuoteMarker != "" {
		lex.queueImpliedSemicolon()
	}

	return
}

func (lex *Lexer) readQuotedStringLiteral(tok *token.Token) (err error) {
	// quoted string literal, such as qs(some string using parentheses as quote marks)
	// qs: allowing langur escape codes
	// QS: without langur escape codes
	allowEsc := token.InterpretEscapeSequences(tok.Literal)
	allowNewLines := true

	modifiers, blockQuoteMarker, err2 := lex.readStringModifiers()
	if err2 != nil {
		err = err2
		return
	}

	any := false
	lead := false
	marks := false
	maybeInterpolated := true

	for _, mod := range modifiers {
		switch mod {
		case "any":
			if any {
				err = fmt.Errorf(`Unexpected repeat of "any" modifier`)
				return
			}
			any = true

		case "lead":
			if lead {
				err = fmt.Errorf(`Unexpected repeat of "lead" modifier`)
				return
			}
			lead = true

		case "ni":
			if !maybeInterpolated {
				err = fmt.Errorf(`Unexpected repeat of "ni" modifier`)
				return
			}
			maybeInterpolated = false

		default:
			err = fmt.Errorf("Invalid string modifier: %s", str.ReformatInput(mod))
			return
		}
	}

	tok.Literal, tok.Type, tok.Attachments, _, err =
		lex.readStringLiteral(allowEsc, maybeInterpolated, allowNewLines, any, lead, marks, blockQuoteMarker)

	if err != nil {
		return
	}

	if lex.cp != ',' && blockQuoteMarker != "" {
		lex.queueImpliedSemicolon()
	}

	return
}

func (lex *Lexer) readCodePointLiteral(any bool) (
	tokLit string, tokType token.Type, err error) {

	// read code point literal as integer (no code point type in Langur)
	var tok token.Token

	lex.advanceCodePoint()

	if lex.cp == '\'' {
		err = fmt.Errorf("Empty code point literal")

	} else if lex.cp == '\\' {
		var cpslc []rune
		cpslc, err = lex.readEscCode('\'', false)
		if err == nil {
			tok.Literal, tok.Type = str.IntToStr(int(cpslc[0]), 10), token.INT
		} else {
			tok.Literal = string(cpslc)
		}

	} else {
		if cpoint.IsVerticalSpace(lex.cp) {
			err = fmt.Errorf("Vertical spacing characters not allowed in code point literals; use '%s' instead", str.CpEsc(lex.cp))

		} else if !any && !cpoint.IsAllowedInStringLiterals(lex.cp) {
			err = fmt.Errorf("Code point literal contains bare code point that is not allowed; use '%s' instead", str.CpEsc(lex.cp))
		}

		tok.Literal = str.IntToStr(int(lex.cp), 10)
		tok.Type = token.INT
		lex.advanceCodePoint()
	}

	if lex.cp == '\'' {
		// end of code point literal
		lex.advanceCodePoint()
	} else {
		err = fmt.Errorf("Code point literal not closed properly")
	}

	return tok.Literal, tok.Type, err
}

func (lex *Lexer) readStringModifiers() (
	modifiers []string, blockQuoteMarker string, err error) {
	// not to be confused with readInterpolationModifiers()

	for lex.cp == ':' {
		lex.advanceCodePoint()
		prefix := ""
		if lex.cp == '-' {
			prefix = "-"
			lex.advanceCodePoint()
		}
		mod, _ := lex.readBareWord()
		mod = prefix + mod

		if mod == "block" {
			// block modifier must be followed by one space and the marker
			if lex.cp == ' ' {
				lex.advanceCodePoint()
			} else {
				err = fmt.Errorf("Expected space, then block marker, then newline")
				break
			}
			var containsLetterOrNumber bool
			blockQuoteMarker, containsLetterOrNumber = lex.readBareWord()
			if !containsLetterOrNumber {
				err = fmt.Errorf("Invalid block marker (%s)", str.ReformatInput(blockQuoteMarker))
				break
			}
			if _, isToken := token.Keywords[blockQuoteMarker]; isToken {
				err = fmt.Errorf("Block marker cannot be token (%s)", blockQuoteMarker)
				break
			}

			if !lex.skipLineReturnOnly() {
				err = fmt.Errorf("Expected newline immediately after starting block marker")
			}
			// block must be last modifier
			return

		} else {
			// not a blockquote modifier
			modifiers = append(modifiers, mod)
		}
	}
	return
}

func markerMatches(possibleMarker, marker string) bool {
	if allowEndingBlockQuoteMarkerOffBeginningOfLine {
		return strings.TrimLeft(possibleMarker, " \t") == marker
	}
	return possibleMarker == marker
}

// see also readFreeWordListLiteral()
func (lex *Lexer) readStringLiteral(
	interpretEsc, maybeInterpolated, allowNewLines,
	any, lead, marks bool,
	blockQuoteMarker string) (
	s string, tt token.Type, pieces []interface{}, containsNewline bool, err error) {

	var errs []error
	tt = token.STRING
	position := lex.bytePosition
	trimminglead := lead

	// to get more accurate accounting of line number/position when there's an error in reading a string literal
	addStrErr := func(e string) {
		errs = append(errs, fmt.Errorf("[%d, %d] %s", lex.line, lex.cpLinePosition, e))
	}

	var closingQuote rune = 0
	var ok bool
	var out, piece, possibleMarker bytes.Buffer

	writeCPToPiece := func() {
		out.WriteRune(lex.cp)
		piece.WriteRune(lex.cp)
	}

	if blockQuoteMarker == "" {
		closingQuote, ok = cpoint.QuotedLiteralClosingMark(lex.cp)
		if !ok {
			err = fmt.Errorf("Illegal opening mark")
			return
		}

		if lex.cp == '\\' && (interpretEsc || maybeInterpolated) {
			addStrErr("Cannot use escape codes or interpolation and use the escape character as a string quote mark")
		}

		if marks {
			writeCPToPiece()
		}
		lex.advanceCodePoint() // past the opening mark
	}

	for {
		marker := possibleMarker.String()

		if blockQuoteMarker != "" &&
			markerMatches(marker, blockQuoteMarker) {

			// end of block quote

			if lex.cp == ',' {
				// may have trailing comma and newline (included in list)
				if !cpoint.IsTokenVerticalSpace(lex.peekCp) {
					// confusion to be avoided
					addStrErr(fmt.Sprintf("Blockquote marker %q and comma not followed by newline (potential confusion); fix or choose another blockquote marker", str.ReformatInput(blockQuoteMarker)))
					break
				}

			} else if cpoint.IsTokenVerticalSpace(lex.cp) || lex.EOF {
				lex.advanceCodePoint()

			} else {
				addStrErr(fmt.Sprintf("Blockquote marker %q not followed by comma, newline, or end of program (potential confusion); choose another blockquote marker", str.ReformatInput(blockQuoteMarker)))
				break
			}

			testOut := out.String()
			pieceOut := piece.String()

			// leave out the blockquote marker and newline preceding it
			Lt := len(testOut) - len(marker) - 2
			newLineWidth := 1
			if Lt >= 0 && testOut[Lt] == '\r' {
				// was CR/LF preceding marker
				newLineWidth = 2
			}
			testOut = testOut[:len(testOut)-len(marker)-newLineWidth]
			pieceOut = pieceOut[:len(pieceOut)-len(marker)-newLineWidth]

			out.Reset()
			piece.Reset()
			out.WriteString(testOut)
			piece.WriteString(pieceOut)
			break
		}

		if lex.EOF {
			addStrErr("Pattern EOF reached without closing string literal.")
			break

		} else if trimminglead && cpoint.IsTrimmableLeadingSpace(lex.cp) {
			// skip this code point
			lex.advanceCodePoint()

		} else if maybeInterpolated && lex.cp == '{' && lex.peekCp == '{' {
			// "{{interpolation}}"
			// add piece preceding interpolation
			pieces = append(pieces, piece.String())
			piece.Reset()

			// get interpolated string and tokens
			lit, tokSlc, interpModifiers, err := lex.readInterpolatedSection(
				interpretEsc, allowNewLines, any)

			if err != nil {
				addStrErr("String interpolation error: " + err.Error())
			}
			if len(tokSlc) == 0 {
				addStrErr("String interpolation error: no tokens")
			}
			out.WriteString(lit)

			// add the token slice
			pieces = append(pieces, tokSlc)

			// add the interpolation modifiers
			pieces = append(pieces, interpModifiers)

			possibleMarker.Reset()
			trimminglead = false

		} else if interpretEsc && lex.cp == '\\' {
			var cparr []rune

			// not an interpolation
			cparr, err = lex.readEscCode(closingQuote, true)
			out.WriteString(string(cparr))
			piece.WriteString(string(cparr))

			if err != nil {
				addStrErr(err.Error())
			}
			possibleMarker.Reset()
			trimminglead = false

		} else if !allowNewLines && cpoint.IsVerticalSpace(lex.cp) {
			addStrErr(fmt.Sprintf("Vertical space characters are not allowed in straight quoted string literals.  Use the q'' or Q'' form or use escape code %s instead or use a block quote", str.CpEsc(lex.cp)))
			lex.advanceCodePoint()
			break

		} else if !any && !cpoint.IsAllowedInStringLiterals(lex.cp) {
			addStrErr(fmt.Sprintf("String literal contains character(s) not allowed; use escape %s instead", str.CpEsc(lex.cp)))
			lex.advanceCodePoint()
			break

		} else if blockQuoteMarker == "" && lex.cp == closingQuote {
			if marks {
				writeCPToPiece()
			}
			lex.advanceCodePoint()
			break

		} else {
			if cpoint.IsTokenVerticalSpace(lex.cp) {
				possibleMarker.Reset()
				containsNewline = true
				trimminglead = lead

			} else {
				possibleMarker.WriteRune(lex.cp)
				trimminglead = false
			}

			writeCPToPiece()
			lex.advanceCodePoint()
		}
	}

	if len(pieces) > 0 {
		// add last piece for interpolation
		pieces = append(pieces, piece.String())
	}

	if !str.Balanced(lex.input[position:lex.bytePosition]) {
		errs = append(errs, fmt.Errorf("String literal contains unbalanced code points, such as LTR/RTL opening markers without matching closing markers"))
	}

	if errs != nil {
		err = errs[0]
	}

	s = out.String()
	return
}

// see also readStringLiteral()
func (lex *Lexer) readFreeWordListLiteral(
	interpretEsc, allowNewLines, any bool, blockQuoteMarker string) (
	strs []string, err error) {

	var errs []error

	// to get more accurate accounting of line number/position when there's an error in reading a string literal
	addStrErr := func(e string) {
		errs = append(errs, fmt.Errorf("[%d, %d] %s", lex.line, lex.cpLinePosition, e))
	}

	var closingQuote rune = 0
	var ok bool

	if blockQuoteMarker == "" {
		closingQuote, ok = cpoint.QuotedLiteralClosingMark(lex.cp)
		if !ok {
			return nil, fmt.Errorf("Illegal opening mark")
		}
		if lex.cp == '\\' && interpretEsc {
			addStrErr("Cannot use escape codes and use the escape character as a quote mark")
		}
		lex.advanceCodePoint() // past the opening mark
	}

	var out bytes.Buffer
	idx := -1

	for {
		marker := out.String()

		if blockQuoteMarker != "" &&
			markerMatches(marker, blockQuoteMarker) {

			// end of block quote

			if lex.cp == ',' {
				// may have trailing comma and newline (included in list)
				if !cpoint.IsTokenVerticalSpace(lex.peekCp) {
					// confusion to be avoided
					addStrErr(fmt.Sprintf("Blockquote marker %q and comma not followed by newline (potential confusion); fix or choose another blockquote marker", str.ReformatInput(blockQuoteMarker)))
					break
				}

			} else if cpoint.IsTokenVerticalSpace(lex.cp) || lex.EOF {
				lex.advanceCodePoint()

			} else {
				addStrErr(fmt.Sprintf("Blockquote marker %q not followed by comma, newline, or end of program (potential confusion); choose another blockquote marker", str.ReformatInput(blockQuoteMarker)))
				break
			}

			out.Reset()
			break
		}

		if lex.EOF {
			addStrErr("Pattern EOF reached without closing free word literal.")
			break

		} else if interpretEsc && lex.cp == '\\' {
			cparr, err := lex.readEscCode(closingQuote, true)
			if err != nil {
				addStrErr(err.Error())
			}
			out.WriteString(string(cparr))

		} else if !allowNewLines && cpoint.IsVerticalSpace(lex.cp) {
			addStrErr(fmt.Sprintf("Vertical space characters are not allowed in this free word list literal; use escape %s with w// to represent the vertical space, or use a block quote to spread over multiple lines", str.CpEsc(lex.cp)))
			lex.advanceCodePoint()

		} else if !any && !cpoint.IsAllowedInQuotedWordLiterals(lex.cp) {
			addStrErr(fmt.Sprintf("Free word list literal contains character(s) not allowed; use a regular list or use escape %s with w//", str.CpEsc(lex.cp)))
			lex.advanceCodePoint()

		} else if blockQuoteMarker == "" && lex.cp == closingQuote {
			lex.advanceCodePoint()
			break

		} else if cpoint.IsTokenSpace(lex.cp) {
			if idx < len(strs) && out.String() != "" {
				// add to word string list
				strs = append(strs, out.String())
				out.Reset()
				idx++
			}
			lex.advanceCodePoint()

		} else {
			out.WriteRune(lex.cp)
			lex.advanceCodePoint()
		}
	}

	if errs != nil {
		err = errs[0]
	}

	// add last one
	if out.String() != "" {
		strs = append(strs, out.String())
	}

	return
}

func cpToType(cp rune) token.Type {
	switch cp {
	case ')':
		return token.RPAREN
	case '}':
		return token.RBRACE
	case ']':
		return token.RBRACKET
	case ';':
		return token.SEMICOLON
	case '>':
		return token.GREATER_THAN
	}
	return token.INVALID
}

func (lex *Lexer) readInterpolatedSection(
	interpretEsc, allowNewLines, any bool) (
	literal string, tokens []token.Token, interpModifiers []string, err error) {

	lex.advanceCodePoint()
	lex.advanceCodePoint() // past the {{

	until := []token.Type{token.COLON, token.RBRACE}

	literal, tokens, err = lexFrom(
		lex.input[lex.bytePosition:], lex.fileName, lex.line, lex.cpPosition, lex.cpLinePosition, until, lex.Modes)

	tokCnt := len(tokens)
	if tokCnt == 0 {
		err = fmt.Errorf("Failed to read any tokens in interpolated section")
		return
	}

	// reset position of current lexer after the sublexed tokens
	resetPosition := lex.bytePosition + len(literal)
	for lex.bytePosition < resetPosition {
		if !allowNewLines && cpoint.IsVerticalSpace(lex.cp) {
			err = fmt.Errorf("Newline(s) found in interpolated section (not allowed in this string literal)")
		}
		if !any && !cpoint.IsAllowedInStringLiterals(lex.cp) && err == nil {
			err = fmt.Errorf("Interpolated section contains illegal character (%s)", cpoint.Display(lex.cp))
		}
		lex.advanceCodePoint()
	}
	if lex.bytePosition != resetPosition {
		err = fmt.Errorf("Error resetting lexer position after sublexing interpolated tokens")
	}
	if err != nil {
		return
	}

	switch tokens[tokCnt-1].Type {
	case token.COLON:
		// read interpolation modifiers
		// lexer just generates slice of strings of interpolation modifiers (does not interpret them)
		var modLiteral string
		modLiteral, interpModifiers, err = lex.readInterpolationModifiers(interpretEsc, any)
		literal += modLiteral

	case token.RBRACE:
		// okay

	default:
		err = fmt.Errorf("Failed to close interpolation")
		return
	}

	if lex.cp == '}' {
		lex.advanceCodePoint()
		// past second } of }}

	} else {
		err = fmt.Errorf("Missing second } of }} to close interpolated section")
		return
	}

	tokens = tokens[:tokCnt-1] // remove : or first } closing token
	return
}

func (lex *Lexer) readInterpolationModifiers(interpretEsc, any bool) (
	literal string, interpModifiers []string, err error) {
	// not to be confused with readStringModifiers()

	// NOTE: The lexer doesn't verify interpolation modifiers. They will be verified when they are compiled.

	mod := bytes.Buffer{}
	firstCp := true
	for {
		if lex.EOF {
			err = fmt.Errorf("Pattern EOF reached without closing string interpolation modifiers")
			return
		}
		if cpoint.IsVerticalSpace(lex.cp) {
			err = fmt.Errorf("Vertical space characters not allowed within interpolation modifiers")
			return
		}

		if lex.cp == '(' || lex.cp == '[' {
			closer, _ := cpoint.QuotedLiteralClosingMark(lex.cp)
			mod.WriteRune(lex.cp)

			var s string
			s, _, _, _, err = lex.readStringLiteral(interpretEsc, false, false, any, false, false, "")
			if err != nil {
				err = fmt.Errorf("Error reading enclosed subsection of interpolation modifier: %s", err.Error())
				return
			}
			mod.WriteString(s)
			mod.WriteRune(closer)
		}

		if lex.cp == '}' {
			// first } of 2 }}
			interpModifiers = append(interpModifiers, mod.String())
			lex.advanceCodePoint()
			break

		} else if lex.cp == ':' {
			// more modifiers
			interpModifiers = append(interpModifiers, mod.String())
			lex.advanceCodePoint()
			mod.Reset()
			firstCp = true
		}

		// effectively trim a leading/trailing space by not adding it
		if !(lex.cp == ' ' && (firstCp || lex.peekCp == ':' || lex.peekCp == '}')) {
			mod.WriteRune(lex.cp)
		}

		firstCp = false
		lex.advanceCodePoint()
	}

	literal = strings.Join(interpModifiers, ":")
	return
}

func (lex *Lexer) readEscCode(closingMark rune, forString bool) (
	result []rune, err error) {
	// NOTE: Coordinate with str.cpEscapeShortForm to ensure both are updated.

	// repeated code...
	readCodePointOrCodeUnitEsc := func(digitCount, base int, isCodeUnitEsc bool) {
		var r rune

		// get past u, U, x, or o
		lex.advanceCodePoint()

		// read in allotted code points
		rslc := make([]rune, digitCount)
		for i := range rslc {
			if lex.EOF {
				err = fmt.Errorf("Pattern EOF reached without finishing escape code")
				return

			} else if lex.cp == closingMark {
				err = fmt.Errorf("Closing mark reached without finishing escape code")
				return
			}

			rslc[i] = lex.cp
			lex.advanceCodePoint()
		}

		r, err = str.StrToRune(string(rslc), base)
		if err == nil {
			if isCodeUnitEsc && cpoint.IsValidCodeUnitEscape(r) ||
				!isCodeUnitEsc && cpoint.IsValidCodePoint(r) {

				result = []rune{r}

			} else {
				if isCodeUnitEsc {
					err = fmt.Errorf("Code Unit Escape Code (%s) outside of allowed range", cpoint.DisplayCU(r))

				} else if lex.Modes.WarnOnSurrogateCodes && cpoint.IsSurrogate(r) {
					err = fmt.Errorf("warning: Surrogate (%s) not allowed for Code Point Escape Code", cpoint.Display(r))

				} else {
					err = fmt.Errorf("Code Point Escape Code (%s) outside of allowed range", cpoint.Display(r))
				}
			}

		} else {
			// change error message
			err = fmt.Errorf("Invalid syntax for Escape Code")
		}
	}

	// get past esc char
	lex.advanceCodePoint()

	if lex.EOF {
		err = fmt.Errorf("Pattern EOF reached without finishing escape code")

	} else {
		switch lex.cp {
		case '\\':
			result = []rune{'\\'}
			lex.advanceCodePoint()

		case closingMark:
			if closingMark == 0 {
				err = fmt.Errorf("Cannot escape null character")
			}
			result = []rune{lex.cp}
			lex.advanceCodePoint()

		case '{', '}':
			if !forString {
				err = fmt.Errorf("Escape \\{ or \\} for string only")
			}
			result = []rune{lex.cp}
			lex.advanceCodePoint()

		case '0':
			// no ambiguous bare octal escapes, so this is fine; 1 to 9 not defined so far
			result = []rune{0}
			lex.advanceCodePoint()

		case 'e':
			// escape
			result = []rune{0x1B}
			lex.advanceCodePoint()

		case 't':
			// horizontal tab
			result = []rune{9}
			lex.advanceCodePoint()

		case 'n':
			// line feed
			result = []rune{0x0A}
			lex.advanceCodePoint()

		case 'r':
			// carriage return
			result = []rune{0x0D}
			lex.advanceCodePoint()

		case 'x':
			readCodePointOrCodeUnitEsc(2, 16, true)

		case 'o':
			readCodePointOrCodeUnitEsc(3, 8, true)

		case 'u':
			readCodePointOrCodeUnitEsc(4, 16, false)

		case 'U':
			readCodePointOrCodeUnitEsc(8, 16, false)

		default:
			err = fmt.Errorf("Unknown escape code %s", cpoint.DisplayCU(lex.cp))
			lex.advanceCodePoint()
		}
	}

	if err != nil {
		result = []rune{cpoint.REPLACEMENT}
	}

	return
}
