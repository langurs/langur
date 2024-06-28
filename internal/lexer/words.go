// langur/lexer/words.go

package lexer

import (
	"bytes"
	"fmt"
	"langur/cpoint"
	"langur/token"
)

func (lex *Lexer) readAndInterpretWordToken(tok *token.Token, cpPosition int) (err error) {
	// keyword or built-in function
	tok.Literal, err = lex.readWord()
	if err != nil {
		return
	}

	// unless we find out otherwise
	tok.Type = token.IDENT

	// in keywords map?
	tt, ok := token.Keywords[tok.Literal]
	if ok {
		tok.Type = tt

		switch tt {
		case token.QUOTE_STRING:
			err = lex.readQuotedStringLiteral(tok)

		case token.FREE_WORD_LIST:
			err = lex.readFreeWordList(tok)

		case token.REGEX_RE2:
			err = lex.readRe2Regex(tok)

		case token.DATETIME:
			// date-time literal
			// using same quoting mechanism as strings
			// might include interpolation
			tok.Literal, _, tok.Attachments, _, err =
				lex.readStringLiteral(false, true, false, false, false, false, "")

			if err != nil {
				return
			}

		case token.DURATION:
			// duration literal
			// using same quoting mechanism as strings
			// might include interpolation
			tok.Literal, _, tok.Attachments, _, err =
				lex.readStringLiteral(false, true, false, false, false, false, "")

			if err != nil {
				return
			}
			tok.Type = token.DURATION

		case token.ZLS:
			tok.Type, tok.Literal = token.STRING, ""

		case token.RESERVED:
			err = fmt.Errorf("Cannot use reserved token %s", tok.Literal)
		}
	}

	return
}

func (lex *Lexer) readWord() (tl string, err error) {
	position := lex.bytePosition

	for cpoint.IsWordTokenChar(lex.cp) || lex.cp == '.' {
		if lex.cp == '.' && lex.peekCp == '.' {
			// 2 dots in a row
			break
		}

		lex.advanceCodePoint()
	}

	tl = lex.input[position:lex.bytePosition]
	return
}

func (lex *Lexer) readBareWord() (string, bool) {
	var out bytes.Buffer
	containsLetterOrNumber := false
	for {
		if !cpoint.IsWordTokenChar(lex.cp) {
			break
		}
		if lex.cp >= 'a' && lex.cp <= 'z' ||
			lex.cp >= 'A' && lex.cp <= 'Z' ||
			lex.cp >= '0' && lex.cp <= '9' {
			containsLetterOrNumber = true
		}
		out.WriteRune(lex.cp)
		lex.advanceCodePoint()
	}
	return out.String(), containsLetterOrNumber
}

func (lex *Lexer) readShortHandStringIndex(tok *token.Token) (err error) {
	lex.advanceCodePoint() // past single quote mark

	if cpoint.IsWordTokenChar(lex.cp) {
		tok.Literal, err = lex.readWord()
		if err != nil {
			return
		}
		tok.Type = token.STRING

		lex.queueToken(*tok)
		tok.Literal, tok.Type = "(])", token.RBRACKET
		lex.queueToken(*tok)

		tok.Literal, tok.Type = "([)", token.LBRACKET

	} else {
		err = fmt.Errorf("Expected word token characters representing short-hand indexing by string")
	}

	return
}
