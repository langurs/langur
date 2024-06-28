// langur/decimal/langur_test.go

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

func TestTrueModulus(t *testing.T) {
	tests := []struct {
		d1     string
		d2     string
		result string
	}{
		{"1.1", "1", "0.1"},
		{"10", "3", "1"},
		{"9", "7", "2"},
		{"132", "789", "132"},
		{"789", "31", "14"},
		{"500", "234.54", "30.92"},

		// ensure it is a modulus operation, not remainder
		{"-5", "3", "1"}, // -2 if remainder; 1 if modulo
		{"5", "-3", "2"}, // 2 in either case

		{"-7", "-3", "2"}, // -1 if remainder; 2 if modulo
		{"-7", "3", "2"},  // -1 if remainder; 2 if modulo
	}

	for _, tt := range tests {
		d1 := RequireFromString(tt.d1)
		d2 := RequireFromString(tt.d2)
		d3 := d1.TrueMod(d2)
		result := RequireFromString(tt.result)
		if !d3.Same(result) {
			t.Errorf("True Modulus Failed: expected=%s, received=%s", tt.result, d3.String())
		}
	}
}

func TestRoots(t *testing.T) {
	tests := []struct {
		d1     string
		d2     string
		result string
	}{
		{"4", "2", "2"},
		{"27", "3", "3"},
		{"256", "2", "16"},
		{"-27", "3", "-3"},
		{"387420489", "9", "9"},
		{"35831808", "7", "12"},
		{"823543", "7", "7"},
		{"256", "-4", "0.25"},

		{"0.25", "2", "0.5"},
		{"0.0625", "-4", "2"},

		// broad range with 34 digit precision...
		// is SLOOOOWWW...
		{"12341253466456", "2345", "1.012937546028772479694875775932214"},
	}

	DivisionPrecision = 33

	for _, tt := range tests {
		d1 := RequireFromString(tt.d1)
		d2 := RequireFromString(tt.d2)
		d3 := d1.Root(d2)

		result := RequireFromString(tt.result)
		if !d3.Same(result) {
			t.Errorf("Roots Failed: expected=%s, received=%s", tt.result, d3.String())
		}
	}
}

func TestMid(t *testing.T) {
	tests := []struct {
		d1     string
		d2     string
		result string
	}{
		{"4", "2", "3"},
		{"27", "3", "15"},
		{"-4.234", "456", "225.883"},
	}

	for _, tt := range tests {
		d1 := RequireFromString(tt.d1)
		d2 := RequireFromString(tt.d2)
		d3 := Mid(d1, d2)
		result := RequireFromString(tt.result)
		if !d3.Same(result) {
			t.Errorf("Mid Failed: expected=%s, received=%s", tt.result, d3.String())
		}
	}
}

func TestMean(t *testing.T) {
	tests := []struct {
		d1     string
		d2     string
		result string
	}{
		{"4", "2", "3"},
		{"27", "3", "15"},
		{"-4.234", "456", "225.883"},
	}

	for _, tt := range tests {
		d1 := RequireFromString(tt.d1)
		d2 := RequireFromString(tt.d2)
		d3 := Mean(d1, d2)
		result := RequireFromString(tt.result)
		if !d3.Same(result) {
			t.Errorf("Mean Failed: expected=%s, received=%s", tt.result, d3.String())
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

func TestScientificNotation(t *testing.T) {
	tests := []struct {
		d1     string
		result string
	}{
		{"0", "0.0e+00"},
		{"123.4567", "1.234567e+02"},
		{"1234567", "1.234567e+06"},
		{"12.34567", "1.234567e+01"},
		{"1.234567", "1.234567e+00"},
		{"0.1234567", "1.234567e-01"},
		{"0.01234567", "1.234567e-02"},
		{"0.00000001234567", "1.234567e-08"},
	}

	for _, tt := range tests {
		d1 := RequireFromString(tt.d1)
		result := d1.ScientificNotation(false, false, true, false, 0, 2)

		if result != tt.result {
			t.Errorf("Scientific Notation Value Failed: expected=%s, received=%s", tt.result, result)
		}
	}
}
