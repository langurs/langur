// langur/str/escape.go

package str

import (
	"fmt"
	"langur/cpoint"
	"strings"
	"unicode"
)

func EscapeGo(s string) string {
	// escape as Go string
	// format with escape codes, but leave out the quote marks
	st := fmt.Sprintf("%q", s)
	return st[1 : len(st)-1]
}

func Escape(s string) string {
	// escape as langur string, not as Go string
	var sb strings.Builder
	for _, cp := range []rune(s) {
		sb.WriteString(CpEscMaybe(cp))
	}
	return sb.String()
}

func EscapeAll(s string) string {
	// escape as langur string, not as Go string
	var sb strings.Builder
	for _, cp := range []rune(s) {
		sb.WriteString(CpEsc(cp))
	}
	return sb.String()
}

var cpEscapeShortForm = map[rune]string{
	// NOTE: Coordinate with the Lexer.readEscCode() method to ensure both are updated.
	0:    `\0`,
	0x1B: `\e`,
	'\\': `\\`,
	0x09: `\t`,
	0x0A: `\n`,
	0x0D: `\r`,
}

// TODO: error on negative code points
// TODO: cu escapes (UTF-8)

func CpEsc(cp rune) string {
	// convert code point to langur escape code
	// some may differ from Go escape codes
	short, ok := cpEscapeShortForm[cp]
	if !ok {
		return CpEscToLongForm(cp)
	}
	return short
}

func CpEscMaybe(cp rune) string {
	short, ok := cpEscapeShortForm[cp]
	if !ok {
		if unicode.IsPrint(cp) {
			return string(cp)
		}
		return CpEscToLongForm(cp)
	}
	return short
}

func CpEscToLongForm(cp rune) string {
	if cpoint.IsValidCodeUnitEscape(cp) {
		s := PadLeft(RuneToStr(cp, 16), 2, '0')
		return `\x` + s

	} else if cp <= 0xFFFF {
		s := PadLeft(RuneToStr(cp, 16), 4, '0')
		return `\u` + s

	} else {
		s := PadLeft(RuneToStr(cp, 16), 8, '0')
		return `\U` + s
	}
}
