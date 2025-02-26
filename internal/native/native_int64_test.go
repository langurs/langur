// native_int64_test.go

package native

import (
	"fmt"
	"math"
	"testing"
)

type int64test struct {
	x      int64
	y      int64
	result int64
	ok     bool
}

func TestAddInt64(t *testing.T) {
	tests := []int64test{
		{1, 0, 1, true},
		{1, 1, 2, true},
		{0, 0, 0, true},
		{0, 1, 1, true},
		{-1, 0, -1, true},
		{-1, -1, -2, true},
		{0, 0, 0, true},
		{0, -1, -1, true},
		{1, -1, 0, true},
		{-1, 1, 0, true},

		{5, 5, 10, true},
		{500, 5000, 5500, true},

		{math.MaxInt64, 5, 0, false},
		{math.MaxInt64, 1, 0, false},
		{math.MaxInt64, 0, math.MaxInt64, true},
		{0, math.MaxInt64, math.MaxInt64, true},
		{0, math.MinInt64, math.MinInt64, true},

		{math.MaxInt64, math.MinInt64, -1, true},
		{math.MinInt64, math.MaxInt64, -1, true},
		{math.MinInt64, math.MinInt64, 0, false},

		{math.MaxInt64 - 1000, 10000, 0, false},
		{math.MinInt64 + 1000, -10000, 0, false},

		{math.MaxInt64 - 1, 1, math.MaxInt64, true},
		{math.MinInt64 + 1, -1, math.MinInt64, true},
	}

	for _, test := range tests {
		result, ok := AddInt64(test.x, test.y)
		if !ok {
			if ok != test.ok {
				t.Errorf("test %s x %s failed: okay status %t, expected %t",
					mmtextInt64(test.x), mmtextInt64(test.y), ok, test.ok)
			}

		} else if result != test.result {
			t.Errorf("test %s + %s result failed, expected=%s, receieved=%s",
				mmtextInt64(test.x), mmtextInt64(test.y), mmtextInt64(test.result), mmtextInt64(result))
		}
	}
}

func mmtextInt64(n int64) string {
	if n == math.MaxInt64 {
		return "MAX_INT64"
	}
	if n == math.MinInt64 {
		return "MIN_INT64"
	}
	return fmt.Sprintf("%d", n)
}

func TestSubInt64(t *testing.T) {
	tests := []int64test{
		{1, 0, 1, true},
		{1, 1, 0, true},
		{0, 0, 0, true},
		{0, 1, -1, true},
		{-1, 0, -1, true},
		{-1, -1, 0, true},
		{0, 0, 0, true},
		{0, -1, 1, true},
		{1, -1, 2, true},
		{-1, 1, -2, true},

		{5, 5, 0, true},
		{500, 5000, -4500, true},

		{math.MinInt64, 1, 0, false},
		{math.MinInt64, 2, 0, false},
		{math.MinInt64 + 1, 2, 0, false},
		{math.MinInt64 + 2, 2, math.MinInt64, true},
		{math.MinInt64, math.MaxInt64, 0, false},

		{math.MinInt64, -1, math.MinInt64 + 1, true},
		{0, math.MinInt64 + 1, -(math.MinInt64 + 1), true},

		{math.MaxInt64, 1, math.MaxInt64 - 1, true},
		{math.MaxInt64, -1, 0, false},

		{math.MinInt64 + 1, 1, math.MinInt64, true},
		{math.MinInt64 + 2, 2, math.MinInt64, true},
		{math.MinInt64 + 3, 4, 0, false},

		{0, math.MaxInt64, -math.MaxInt64, false},
	}

	for _, test := range tests {
		result, ok := SubInt64(test.x, test.y)
		if !ok {
			if ok != test.ok {
				t.Errorf("test %s x %s failed: okay status %t, expected %t",
					mmtextInt64(test.x), mmtextInt64(test.y), ok, test.ok)
			}

		} else if result != test.result {
			t.Errorf("test %s - %s result failed, expected=%s, receieved=%s",
				mmtextInt64(test.x), mmtextInt64(test.y), mmtextInt64(test.result), mmtextInt64(result))
		}
	}
}

func TestMultiplyInt64(t *testing.T) {
	tests := []int64test{
		{1, 0, 0, true},
		{1, 1, 1, true},
		{0, 0, 0, true},
		{0, 1, 0, true},
		{-1, 0, 0, true},
		{-1, -1, 1, true},
		{0, 0, 0, true},
		{0, -1, 0, true},
		{1, -1, -1, true},
		{-1, 1, -1, true},

		{5, 5, 25, true},
		{500, 5000, 2500000, true},

		// slow ones...
		{45612378945, 89456453, 4080321633311582085, true},
		{45612378945, 5000, 228061894725000, true},
		{45612378945, 456789456453, 0, false}, // beyond int64

		{math.MinInt64, 1, math.MinInt64, true},

		{math.MinInt64 + 1, 1, math.MinInt64 + 1, true},

		{math.MinInt64, 2, 0, false},
		{math.MinInt64 + 1, 2, 0, false},
		{math.MinInt64 + 2, 2, 0, false},
		{math.MinInt64, math.MaxInt64, 0, false},

		{math.MaxInt64, 1, math.MaxInt64, true},
		{1, math.MaxInt64, math.MaxInt64, true},
		{math.MaxInt64, 2, math.MaxInt64, false},
		{2, math.MaxInt64, math.MaxInt64, false},

		{math.MaxInt64 / 2, 2, math.MaxInt64 - 1, true},
		{math.MaxInt64 / 4, 4, math.MaxInt64 - 3, true},

		{math.MinInt64 / 2, 2, 0, false},
		{(math.MinInt64 + 2) / 2, 2, math.MinInt64 + 2, true},
	}

	for _, test := range tests {
		result, ok := MultiplyInt64(test.x, test.y)
		if !ok {
			if ok != test.ok {
				t.Errorf("test %s x %s failed: okay status %t, expected %t",
					mmtextInt64(test.x), mmtextInt64(test.y), ok, test.ok)
			}

		} else if result != test.result {
			t.Errorf("test %s x %s result failed, expected=%s, receieved=%s",
				mmtextInt64(test.x), mmtextInt64(test.y), mmtextInt64(test.result), mmtextInt64(result))
		}
	}
}

func TestDivideInt64(t *testing.T) {
	tests := []int64test{
		{1, 1, 1, true},
		{1, 0, 0, false},
		{-1, 1, -1, true},
		{1, -1, -1, true},
		{0, 1, 0, true},

		{144, 12, 12, true},
		{-144, 12, -12, true},
		{144, 11, 0, false},

		{3, 1, 3, true},
		{3, -1, -3, true},
		{3, -3, -1, true},

		{3, 2, 0, false},
		{3, 4, 0, false},
	}

	for _, test := range tests {
		result, ok := DivideInt64(test.x, test.y)
		if !ok {
			if ok != test.ok {
				t.Errorf("test %s / %s failed: okay status %t, expected %t",
					mmtextInt64(test.x), mmtextInt64(test.y), ok, test.ok)
			}

		} else if result != test.result {
			t.Errorf("test %s / %s result failed, expected=%s, receieved=%s",
				mmtextInt64(test.x), mmtextInt64(test.y), mmtextInt64(test.result), mmtextInt64(result))
		}
	}
}

func TestModulusInt64(t *testing.T) {
	tests := []int64test{
		{-21, 4, 3, true},
		{-5, 3, 1, true},
		{5, 3, 2, true},
		{5, 2, 1, true},
	}

	for _, test := range tests {
		result, ok := ModulusInt64(test.x, test.y)
		if !ok {
			if ok != test.ok {
				t.Errorf("test %s mod %s failed: okay status %t, expected %t",
					mmtextInt64(test.x), mmtextInt64(test.y), ok, test.ok)
			}

		} else if result != test.result {
			t.Errorf("test %s mod %s result failed, expected=%s, receieved=%s",
				mmtextInt64(test.x), mmtextInt64(test.y), mmtextInt64(test.result), mmtextInt64(result))
		}
	}
}

func TestFloorDivisionInt64(t *testing.T) {
	tests := []int64test{
		{1, 0, 0, false},
		{1, 2, 0, true},
		{-1, 2, -1, true},
		{1, -2, -1, true},
		{5, 2, 2, true},
		{5, 2, 2, true},
		{10, 2, 5, true},
		{5, 2, 2, true},
		{-5, 2, -3, true},
		{-10, 2, -5, true},
		{5, 2, 2, true},
		{-40, 3, -14, true},
		{40, -3, -14, true},
		{-41, 3, -14, true},
	}

	for _, test := range tests {
		result, ok := DivideFloorInt64(test.x, test.y)
		if !ok {
			if ok != test.ok {
				t.Errorf("test %s // %s failed: okay status %t, expected %t",
					mmtextInt64(test.x), mmtextInt64(test.y), ok, test.ok)
			}

		} else if result != test.result {
			t.Errorf("test %s // %s result failed, expected=%s, receieved=%s",
				mmtextInt64(test.x), mmtextInt64(test.y), mmtextInt64(test.result), mmtextInt64(result))
		}
	}
}
