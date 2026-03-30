// langur/cpoint/cpoint.go

package cpoint

import (
	"fmt"
	"unicode"
)

const (
	REPLACEMENT rune = 0xFFFD
	MAX         rune = unicode.MaxRune
	codeUnitMax rune = 127
)

var endOfStringIndicator = "EOF"

func Display(cp rune) string {
	if cp > 0xFFFF {
		return fmt.Sprintf("U+%08X", cp)
	}
	return fmt.Sprintf("U+%04X", cp)
}

func DisplayCU(cp rune) string {
	if cp <= codeUnitMax {
		return fmt.Sprintf("%02X", cp)
	}
	return Display(cp)
}

func IsEOF(err error) bool {
	return err != nil && err.Error() == endOfStringIndicator
}

func Lcase(cp rune) rune {
	return unicode.ToLower(cp)
}
func Ucase(cp rune) rune {
	return unicode.ToUpper(cp)
}
func Tcase(cp rune) rune {
	return unicode.ToTitle(cp)
}

func QuotedLiteralClosingMark(openingMark rune) (cp rune, ok bool) {
	switch openingMark {
	case '"', '\'', '/':
		return openingMark, true
	case '(':
		return ')', true
	case '[':
		return ']', true
	// case '<':
	// 	return '>', true
	}
	return 0, false
}

func ClosingMark(openingMark rune) rune {
	switch openingMark {
	case '{':
		return '}'
	case '(':
		return ')'
	case '[':
		return ']'
	case '<':
		return '>'
	}
	return openingMark
}

func InSlice(cp rune, set []rune) bool {
	for _, cp2 := range set {
		if cp2 == cp {
			return true
		}
	}
	return false
}

// currently limited to ASCII range, unless its use can be otherwise clarified
func IsValidCodeUnitEscape(r rune) bool {
	return r >= 0 && r <= codeUnitMax
}

// checks that value is within code point constraints
func IsValidCodePoint(cp rune) bool {
	return cp >= 0 && cp <= MAX && !IsSurrogate(cp)
}

func IsWordTokenChar(cp rune) bool {
	return IsStartingWordTokenChar(cp) || '0' <= cp && cp <= '9' || cp == '_'
}

func IsStartingWordTokenChar(cp rune) bool {
	return 'a' <= cp && cp <= 'z' || 'A' <= cp && cp <= 'Z'
}

func IsDigitInBase(cp rune, base int) bool {
	// up to base 36
	// TODO: update for base 37 to 62
	if base < 2 || base > 36 {
		// bug("IsDigitInBase", "base passed out of range")
		return false
	}

	if cp >= '0' && cp <= '9' {
		return cp-'0' < rune(base)

	} else if cp >= 'a' && cp <= 'z' {
		return cp-'a'+10 < rune(base)

	} else if cp >= 'A' && cp <= 'Z' {
		return cp-'A'+10 < rune(base)
	}
	return false
}

func IsCapitalBase16Digit(cp rune) bool {
	return '0' <= cp && cp <= '9' || 'A' <= cp && cp <= 'F'
}

func Base36ToNumber(cp rune) (n int, err error) {
	// with uppercase letters having the same value as lowercase
	if cp >= '0' && cp <= '9' {
		return int(cp - '0'), nil
	} else if cp >= 'a' && cp <= 'z' {
		return int(cp - 87), nil
	} else if cp >= 'A' && cp <= 'Z' {
		return int(cp - 55), nil
	}
	return 0, fmt.Errorf("Expected 0-9, A-Z, a-z")
}

func Base62ToNumber(cp rune) (n int, err error) {
	// with uppercase letters having a higher value than lowercase
	if cp >= '0' && cp <= '9' {
		return int(cp - '0'), nil
	} else if cp >= 'a' && cp <= 'z' {
		return int(cp - 87), nil
	} else if cp >= 'A' && cp <= 'Z' {
		return int(cp - 29), nil
	}
	return 0, fmt.Errorf("Expected 0-9, A-Z, a-z")
}

func ReverseSlice(cpSlc []rune) []rune {
	newSlc := make([]rune, len(cpSlc))
	for i, j := 0, len(cpSlc)-1; i < len(cpSlc); i++ {
		newSlc[i] = cpSlc[j]
		j--
	}
	return newSlc
}
