// /langur/cpoint/utf8.go

package cpoint

import (
	"fmt"
	"unicode/utf8"
)

func BytesToString(fnName string, bSlc []byte) (string, error) {
	// TODO: check for surrogate pairs and fix

	if utf8.Valid(bSlc) {
		return string(bSlc), nil
	}
	return "", fmt.Errorf("Invalid UTF-8 byte sequence")
}

func Decode(sp *string, codeUnitPosition int) (cp rune, width int, err error) {
	if codeUnitPosition >= len(*sp) {
		cp, width, err = 0, 0, fmt.Errorf(endOfStringIndicator)
	} else {
		cp, width = utf8.DecodeRuneInString((*sp)[codeUnitPosition:])
		if cp == utf8.RuneError {
			err = fmt.Errorf("Error decoding code point (invalid UTF-8 encoding)")

		} else if IsSurrogate(cp) {
			err = fmt.Errorf("Surrogate code point (invalid UTF-8 encoding)")

		} else if cp > MAX {
			err = fmt.Errorf("Code point out of range (invalid UTF-8 encoding)")
		}
	}
	return
}

// startCodeUnit, codeUnitPosition: 0-based code unit position
// codePointPosition: 1-based code point position
func CodePointPosFromCodeUnitPos(sp *string, codeUnitPosition int) (
	codePointPosition int, err error) {

	if codeUnitPosition == 0 && len(*sp) > 0 {
		return 1, nil
	}

	cpCnt := 1
	for i := 0; i < codeUnitPosition; {
		_, width, err := Decode(sp, i)
		if err != nil {
			return 0, err
		}
		i += width
		cpCnt++

		if i == codeUnitPosition {
			return cpCnt, nil
		}
	}
	return 0, fmt.Errorf("Code Point Position not found")
}
