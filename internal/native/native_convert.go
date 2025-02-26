// native_convert.go

package native

import (
	"math"
)

const IntSize = 32 << (^uint(0) >> 63)

func ToInt(from int64) (i int, ok bool) {
	if from > math.MaxInt || from < math.MinInt {
		return 0, false
	}
	return int(from), true
}

func ToIntNegated(from int64) (i int, ok bool) {
	from, ok = NegateInt64(from)
	if !ok {
		return
	}
	if from > math.MaxInt || from < math.MinInt {
		return 0, false
	}
	return int(from), true
}
