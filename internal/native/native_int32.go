// /native/native_int32.go

package native

import (
	"math"
	"unicode"
)

func ToInt32(from int64) (int32, bool) {
	if from >= math.MinInt32 && from <= math.MaxInt32 {
		return int32(from), true
	}
	return 0, false
}

func ToRune(from int64) (rune, bool) {
	if from >= 0 && from <= unicode.MaxRune {
		return rune(from), true
	}
	return 0, false
}
