// langur/lexer/numbers.go

package lexer

import (
	"fmt"
	"langur/cpoint"
	"langur/str"
	"langur/token"
)

func (lex *Lexer) readNumber() (
	tl string, tt token.Type, base int, err error) {

	position := lex.bytePosition

	// unless we find out otherwise...
	tt = token.INT

	// Don't set a base if not explicit basex notation.
	// Context may dictate a different default base (interpreted later).
	base = token.CODE_DEFAULT

	for cpoint.IsWordTokenChar(lex.cp) || lex.cp == '.' {
		if lex.cp == '.' {
			if lex.peekCp == '.' {
				// more than one dot in a row, not a floating point (may be range operator)
				// Don't advance the code point, since we're already on the first dot.
				break
			}

			if tt == token.INT {
				tt = token.FLOAT
			} else {
				tt = token.INVALID
				if err == nil {
					err = fmt.Errorf("Multiple decimal points found in numeric literal")
				}
			}

			lex.advanceCodePoint()

		} else if (lex.cp == 'e' || lex.cp == 'E') && (lex.peekCp == '+' || lex.peekCp == '-') {
			// e-notation requiring + or -
			if tt == token.INT {
				tt = token.FLOAT
			}
			lex.advanceCodePoint()
			lex.advanceCodePoint()

			if err == nil && (lex.cp == '_' || !cpoint.IsWordTokenChar(lex.cp)) {
				err = fmt.Errorf("Missing digits after start of e-notation")
			}

		} else if (lex.cp == 'e' || lex.cp == 'E') &&
			(base != token.CODE_DEFAULT && base < 15 || base == token.CODE_DEFAULT) {

			// Langur requires the plus or minus for e-notation, which we checked for in the "else if" above.
			// We need to check this here, in case the decimal library we're using doesn't require it.
			err = fmt.Errorf("Missing plus or minus after start of e-notation")
			lex.advanceCodePoint()

		} else if err == nil && lex.cp == 'x' {
			// numbers preceding x indicates a base, such as 16xFF or 2x1010
			// read base section
			b := lex.input[position:lex.bytePosition]

			// move past the x
			lex.advanceCodePoint()

			if tt == token.INT {
				base, err = str.StrToInt(b, 10)
				// letting the parser decide if it's a valid base

			} else {
				tt = token.INVALID
				err = fmt.Errorf("Error in base section of numeric literal")
			}

			if err != nil {
				continue
			}

			if !cpoint.IsWordTokenChar(lex.cp) {
				tl, tt = str.IntToStr(base, 10)+"x", token.INVALID
				if err == nil {
					err = fmt.Errorf("Base specifier stump (missing number)")
				}

			} else if tokLit, tokType, tokErr := lex.readNumberOfBase(base); tokErr == nil {
				tl, tt = tokLit, tokType

			} else {
				if err == nil {
					err = tokErr
				}
				tl, tt = str.IntToStr(base, 10)+"x"+tokLit, token.INVALID
			}

			return

		} else {
			// keeps reading whether the token is legal or illegal (grab all identifier characters)
			lex.advanceCodePoint()
		}
	}

	tl = lex.input[position:lex.bytePosition]
	return
}

// called by readNumber() after a base specifier is found
func (lex *Lexer) readNumberOfBase(base int) (tl string, tt token.Type, err error) {
	tl, tt = "", token.INVALID

	if !cpoint.IsWordTokenChar(lex.cp) {
		// This is a bad start.
		err = fmt.Errorf("Not able to read number")
		return
	}

	if base < 2 {
		err = fmt.Errorf("Cannot read number with base less than 2")
		return
	}

	// We read in all the "word" characters whether they are valid or not.
	position := lex.bytePosition
	tt = token.INT

	for cpoint.IsWordTokenChar(lex.cp) || lex.cp == '.' {
		if lex.cp == '.' {
			if lex.peekCp == '.' {
				// more than one dot in a row, not a floating point
				// Don't advance since we're already at the first dot.
				break
			}

			if tt == token.INT {
				tt = token.FLOAT
			} else {
				tt = token.INVALID
				if err == nil {
					err = fmt.Errorf("Multiple decimal points found in numeric literal")
				}
			}
			lex.advanceCodePoint()

		} else if (lex.cp == 'e' || lex.cp == 'E') && (lex.peekCp == '+' || lex.peekCp == '-') {
			// e-notation requiring + or -
			if tt == token.INT {
				tt = token.FLOAT
			}
			lex.advanceCodePoint()
			lex.advanceCodePoint()

			if err == nil && (lex.cp == '_' || !cpoint.IsWordTokenChar(lex.cp)) {
				err = fmt.Errorf("Missing digits after start of e-notation")
			}

		} else if lex.cp == '_' {
			// ok
			lex.advanceCodePoint()

		} else if err == nil && !cpoint.IsDigitInBase(lex.cp, base) {
			tt = token.INVALID
			err = fmt.Errorf("Invalid characters for base used (" + str.IntToStr(base, 10) + ") in numeric literal")
			lex.advanceCodePoint()

		} else {
			lex.advanceCodePoint()
		}
	}

	tl = lex.input[position:lex.bytePosition]
	return
}
