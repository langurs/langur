// langur/decimal/langur_test.go

package decimal

import (
	"testing"
)

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

func TestScientificNotation(t *testing.T) {
	tests := []struct {
		d1     string
		result string
	}{
		{"0", "0e+00"},
		{"123.4567", "1.234567e+02"},
		{"1234567", "1.234567e+06"},
		{"12.34567", "1.234567e+01"},
		{"1.234567", "1.234567e+00"},
		{"0.1234567", "1.234567e-01"},
		{"0.01234567", "1.234567e-02"},
		{"0.00000001234567", "1.234567e-08"},
		{"-100", "-1e+02"},
	}

	for _, tt := range tests {
		d1 := RequireFromString(tt.d1)
		result := d1.ScientificNotation(false, false, true, false, true, true, 0, 2)

		if result != tt.result {
			t.Errorf("Scientific Notation Value Failed: expected=%s, received=%s", tt.result, result)
		}
	}
}
