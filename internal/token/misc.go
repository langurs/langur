// langur/token/misc.go

package token

import (
	"bytes"
	"fmt"
	"langur/trace"
)

type Token struct {
	Type    Type
	Literal string

	// Code: so far has multiple uses
	// see the token code flags enumeration
	Code int

	// Code2
	// 1. If basex notation is used, this will specify the base for the parser to use to interpret a number, ...
	// ... or can be 0 for base 10.
	Code2 int

	// Attachments
	// 1. string/regex interpolation
	Attachments []interface{}

	// for error reporting
	Where trace.Where

	NewLinePrecedes              bool
	CpDiff                       int  // number of code points from previous token
	AddImpliedSemicolonAtNewLine bool // an override of the norm

	// lexing or parsing errors
	Errs []tokenErr
}

func (tok Token) Copy() Token {
	newTok := Token{
		Type:            tok.Type,
		Literal:         tok.Literal,
		Code:            tok.Code,
		Code2:           tok.Code2,
		Where:           tok.Where.Copy(),
		NewLinePrecedes: tok.NewLinePrecedes,
		CpDiff:          tok.CpDiff,

		AddImpliedSemicolonAtNewLine: tok.AddImpliedSemicolonAtNewLine,
	}

	if len(tok.Attachments) > 0 {
		newTok.Attachments = make([]interface{}, len(tok.Attachments))
		copy(newTok.Attachments, tok.Attachments)
	}

	if len(tok.Errs) > 0 {
		newTok.Errs = make([]tokenErr, len(tok.Errs))
		copy(newTok.Errs, tok.Errs)
	}

	return newTok
}

func (tok Token) NewTokenCopyPosInfo(typ Type, literal string) Token {
	return Token{
		Type:    typ,
		Literal: literal,

		Where:           tok.Where.Copy(),
		NewLinePrecedes: tok.NewLinePrecedes,
		CpDiff:          tok.CpDiff,
	}
}

func InTypeSlice(tokType Type, tokTypes []Type) bool {
	for _, tt := range tokTypes {
		if tokType == tt {
			return true
		}
	}
	return false
}

const (
	// token code flags enumeration
	// 1a. On an operator, this can indicate a database operation (not database op by default).
	// 1b. ... and combination operator such as += *= etc.
	// 2. indicates to implicitly escape interpolations
	// 3. passes object type integer for type token
	// 4. indicate an imaginary number
	// 5. indicate to include fractional seconds in a "now" datetime literal

	// 0x01, 0x02, 0x04, 0x08, ...
	CODE_DEFAULT                         = 0
	CODE_DB_OPERATOR                     = 1 << iota
	CODE_COMBINATION_ASSIGNMENT_OPERATOR
	CODE_ESC_ALL_INTERPOLATION
	CODE_IMAGINARY_NUMBER
	CODE_FRACTIONAL_SECONDS
)

func New(line, linePosition int) Token {
	var tok Token
	tok.Type = INVALID
	tok.Where = trace.NewWhere(line, linePosition)
	tok.Code = CODE_DEFAULT
	tok.Code2 = CODE_DEFAULT
	tok.Attachments = nil
	return tok
}

func bug(s string) {
	panic("Token bug: " + s)
}

type tokenErr struct {
	Err   error
	Count int
}

func (tok Token) Errors(delim string) string {
	var out bytes.Buffer
	for i, e := range tok.Errs {
		out.WriteString(fmt.Sprintf("%d: %s", e.Count, e.Err))
		if i < len(tok.Errs)-1 {
			out.WriteString(delim)
		}
	}
	return out.String()
}

func (t *Token) AddTokenErr(err error) {
	// If the exact error is already accounted for, just add to the count.
	found := false
	if t.Errs != nil {
		for i := range t.Errs {
			if t.Errs[i].Err.Error() == err.Error() {
				t.Errs[i].Count++
				found = true
				break
			}
		}
	}

	if !found {
		t.Errs = append(t.Errs, tokenErr{Err: err, Count: 1})
	}
}

func TypeDescription(tt Type) string {
	if tt == INVALID {
		return "INVALID"
	}
	return string(tt)
}

func (t Token) String() string {
	var out bytes.Buffer

	out.WriteRune('[')
	out.WriteString(t.Where.String())

	// if t.CpDiff > 0 {
	out.WriteString(fmt.Sprintf(" %d", t.CpDiff))
	// }
	out.WriteString(fmt.Sprintf("] %s", TypeDescription(t.Type)))
	out.WriteString(fmt.Sprintf(": %q", t.Literal))

	if t.Code != CODE_DEFAULT {
		out.WriteString(fmt.Sprintf(", Code %d", t.Code))
	}
	if t.Code2 != CODE_DEFAULT {
		out.WriteString(fmt.Sprintf(", Code2 %d", t.Code2))
	}
	if t.Errs != nil {
		out.WriteString("; Errs: [" + t.Errors("; ") + "]")
	}

	return out.String()
}
