// langur/decimal/langur_zeroes_test.go

package decimal

import (
	"testing"
)

func TestAsEqual(t *testing.T) {
	tests := []struct {
		d1, d2      string
		shouldMatch bool
	}{
		{"1", "1", true},
		{"1", "1.0", true},
		{"1.23", "1.23", true},
		{"1.23", "1.2300", true},
		{"0.23", "1.23", false},
	}

	for _, tt := range tests {
		d1 := RequireFromString(tt.d1)
		d2 := RequireFromString(tt.d2)

		if tt.shouldMatch {
			if !d1.Equal(d2) {
				t.Errorf("Expected %s and %s to compare as \"Equal\"", tt.d1, tt.d2)
			}
		} else {
			if d1.Equal(d2) {
				t.Errorf("Expected %s and %s NOT to compare as \"Equal\"", tt.d1, tt.d2)
			}
		}
	}
}

// "same" different than "equal"
// 1 == 1.0 but not the same
func TestAsSame(t *testing.T) {
	tests := []struct {
		d1, d2      string
		shouldMatch bool
	}{
		{"1", "1", true},
		{"1", "1.0", false},
		{"1.23", "1.23", true},
		{"1.2300", "1.2300", true},
		{"1.23", "1.2300", false},
		{"0.23", "1.23", false},
	}

	for _, tt := range tests {
		d1 := RequireFromString(tt.d1)
		d2 := RequireFromString(tt.d2)

		if tt.shouldMatch {
			if !d1.Same(d2) {
				t.Errorf("Expected %s and %s to compare as \"Same\"", tt.d1, tt.d2)
			}
		} else {
			if d1.Same(d2) {
				t.Errorf("Expected %s and %s NOT to compare as \"Same\"", tt.d1, tt.d2)
			}
		}
	}
}

func TestSimplify(t *testing.T) {
	tests := []struct {
		d1     string
		result string
	}{
		{"0", "0"},
		{"0.1", "0.1"},
		{"0.10000", "0.1"},
		{"0.1234567890000", "0.123456789"},
		{"1.0", "1"},
		{"123.000", "123"},
		{"1230.00", "1230"},
		{"123000", "123000"},
		{"123.234000", "123.234"},
	}

	for _, tt := range tests {
		d1 := RequireFromString(tt.d1)
		result := d1.Simplify()

		if result.string(false) != tt.result {
			t.Errorf("Simplify Value Failed: expected=%s, received=%s", tt.result, result.String())
		}
	}
}

func TestTruncatedDivision(t *testing.T) {
	tests := []struct {
		d1     string
		d2     string
		scale  int32
		result string
	}{
		{"7", "2", 0, "3"},
		{"16", "3", 0, "5"},
		{"1.1", "1.0", 0, "1"},
		{"132", "789", 0, "0"},
		{"234.54", "500", 0, "0"},

		// truncated, not floor division
		{"-3", "2", 0, "-1"},
		{"40", "-3", 0, "-13"},
		{"-6", "2", 0, "-3"}, // -3 in either case
	}

	for _, tt := range tests {
		d1 := RequireFromString(tt.d1)
		d2 := RequireFromString(tt.d2)
		d3 := d1.DivTruncate(d2, tt.scale)
		result := RequireFromString(tt.result)
		if !d3.Same(result) {
			t.Errorf("Truncated Division Failed: expected=%s, received=%s", tt.result, d3.String())
		}
	}
}

func TestFloorDivision(t *testing.T) {
	tests := []struct {
		d1     string
		d2     string
		scale  int32
		result string
	}{
		{"7", "2", 0, "3"},
		{"16", "3", 0, "5"},
		{"1.1", "1.0", 0, "1"},
		{"132", "789", 0, "0"},
		{"234.54", "500", 0, "0"},

		// floor division, not truncated division
		{"-3", "2", 0, "-2"},
		{"40", "-3", 0, "-14"},
		{"-6", "2", 0, "-3"}, // -3 in either case
	}

	for _, tt := range tests {
		d1 := RequireFromString(tt.d1)
		d2 := RequireFromString(tt.d2)
		d3 := d1.DivFloor(d2)
		result := RequireFromString(tt.result)
		if !d3.Same(result) {
			t.Errorf("Floor Division Failed: expected=%s, received=%s", tt.result, d3.String())
		}
	}
}

func TestRoundByMode(t *testing.T) {
	tests := []struct {
		d1     string
		max    int32
		add    bool
		trim   bool
		result string
	}{
		{"1", 0, true, true, "1"},
		{"1", 1, true, true, "1"},
		{"1.0", 1, true, true, "1"},
		{"1.0", 2, true, true, "1"},
		{"1.00", 1, true, true, "1"},
		{"1.00", 2, true, true, "1"},

		{"3.10", 2, true, true, "3.1"},
		{"3.100", 2, true, true, "3.1"},
		{"3.1001", 2, true, true, "3.1"},

		{"1", 0, true, false, "1"},
		{"1", 1, true, false, "1.0"},
		{"1.0", 1, true, false, "1.0"},
		{"1.0", 2, true, false, "1.00"},
		{"1.00", 1, true, false, "1.0"},
		{"1.00", 2, true, false, "1.00"},

		{"3.10", 2, true, false, "3.10"},
		{"3.100", 2, true, false, "3.10"},
		{"3.1001", 2, true, false, "3.10"},

		{"3.10", -1, true, false, "0"},
		{"311", -1, true, false, "310"},
		{"311", -2, true, false, "300"},
		{"311", -3, true, false, "0"},
		{"311", -4, true, false, "0"},
	}

	for _, tt := range tests {
		d1 := RequireFromString(tt.d1)
		result := d1.RoundByMode(tt.max, tt.add, tt.trim, RoundingMode_Default).StringWithTrailingZeros()

		if result != tt.result {
			t.Errorf("RoundByMode(%d, %t) Value Failed: expected=%s, received=%s", tt.max, tt.trim, tt.result, result)
		}
	}
}

func TestTruncateWithZeroes(t *testing.T) {
	tests := []struct {
		d1     string
		max    int32
		add    bool
		trim   bool
		result string
	}{
		{"1", 0, true, true, "1"},
		{"1", 1, true, true, "1"},
		{"1.0", 1, true, true, "1"},
		{"1.0", 2, true, true, "1"},
		{"1.00", 1, true, true, "1"},
		{"1.00", 2, true, true, "1"},

		{"3.10", 2, true, true, "3.1"},
		{"3.100", 2, true, true, "3.1"},
		{"3.1001", 2, true, true, "3.1"},

		{"1", 0, true, false, "1"},
		{"1", 1, true, false, "1.0"},
		{"1.0", 1, true, false, "1.0"},
		{"1.0", 2, true, false, "1.00"},
		{"1.00", 1, true, false, "1.0"},
		{"1.00", 2, true, false, "1.00"},

		{"3.10", 2, true, false, "3.10"},
		{"3.100", 2, true, false, "3.10"},
		{"3.1001", 2, true, false, "3.10"},

		{"3.10", -1, true, false, "0"},
		{"311", -1, true, false, "310"},
		{"311", -2, true, false, "300"},
		{"311", -3, true, false, "0"},
		{"311", -4, true, false, "0"},
	}

	for _, tt := range tests {
		d1 := RequireFromString(tt.d1)
		result := d1.TruncateWithZeroes(tt.max, tt.add, tt.trim).StringWithTrailingZeros()

		if result != tt.result {
			t.Errorf("TruncateWithZeroes Value Failed: expected=%s, received=%s", tt.result, result)
		}
	}
}
