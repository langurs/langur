// /native/native_int64.go

package native

import (
	"math"
)

func AddInt64(x, y int64) (int64, bool) {
	// check for overflow
	if x >= 0 {
		if y > math.MaxInt64-x {
			return 0, false
		}
	} else {
		if y < math.MinInt64-x {
			return 0, false
		}
	}
	return x + y, true
}

func SubInt64(x, y int64) (int64, bool) {
	// check for overflow
	if x >= 0 {
		if y < -(math.MaxInt64 - x) {
			return 0, false
		}
	} else {
		if y > -(math.MinInt64 - x) {
			return 0, false
		}
	}
	return x - y, true
}

func AbsInt64(x int64) (int64, bool) {
	if x == math.MinInt64 {
		return 0, false
	} else if x < 0 {
		return -x, true
	}
	return x, true
}

func NegateInt64(x int64) (int64, bool) {
	if x == math.MinInt64 {
		return 0, false
	}
	return -x, true
}

func MultiplyInt64(x, y int64) (int64, bool) {
	switch x {
	case 0:
		return 0, true
	case 1:
		return y, true
	}
	switch y {
	case 0:
		return 0, true
	case 1:
		return x, true
	}

	// working based on the fact that multiplication can be done with addition
	var base, times, product int64
	var negatives int
	var ok bool

	// to work only with positive numbers
	if x < 0 {
		negatives++
		x, ok = NegateInt64(x)
		if !ok {
			return 0, false
		}
	}
	if y < 0 {
		negatives++
		y, ok = NegateInt64(y)
		if !ok {
			return 0, false
		}
	}

	// do the fewest calculations, using the smaller number as times
	// It would work either way, but if you have 50 x 10000 would you rather do 50 calculations or 10000?
	if x > y {
		base, times = x, y
	} else {
		base, times = y, x
	}

	var i int64
	for i = 0; i < times; i++ {
		// product inititalized to 0 by go
		product, ok = AddInt64(product, base)
		if !ok {
			return 0, false
		}
	}
	if negatives == 1 {
		product, ok = NegateInt64(product)
		if !ok {
			return 0, false
		}
	}

	return product, true
}

func DivideInt64(x, y int64) (int64, bool) {
	// clear integer result with no remainder
	if y == 0 {
		return 0, false
	}
	if x%y == 0 {
		return x / y, true
	}
	return 0, false
}

func DivideTruncateInt64(x, y int64) (int64, bool) {
	if y == 0 {
		return 0, false
	}
	return x / y, true
}

func DivideFloorInt64(x, y int64) (int64, bool) {
	if y == 0 {
		return 0, false
	}
	// starting with truncating division
	result, ok := x/y, true
	if x%y != 0 {
		if y > 0 && x%y < 0 {
			result, ok = SubInt64(result, 1)

		} else if y < 0 {
			x, ok = NegateInt64(x)
			if ok {
				y, ok = NegateInt64(y)
				if ok {
					if x%y < 0 {
						result, ok = SubInt64(result, 1)
					}
				}
			}
		}
	}
	return result, ok
}

func RemainderInt64(x, y int64) (int64, bool) {
	if y == 0 {
		return 0, false
	}
	return x % y, true
}

func ModulusInt64(x, y int64) (int64, bool) {
	result, ok := RemainderInt64(x, y)
	if ok {
		if result < 0 {
			add, ok2 := AbsInt64(y)
			if !ok2 {
				return result, false
			}
			result, ok = AddInt64(result, add)
		}
	}
	return result, ok
}
