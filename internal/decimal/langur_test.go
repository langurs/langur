// langur/decimal/langur_test.go

package decimal

import (
	"fmt"
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
		result := d1.ScientificNotation(false, false, true, false, true, true, 0, 2, RoundingMode_Default, '.')

		if result != tt.result {
			t.Errorf("Scientific Notation Value Failed: expected=%s, received=%s", tt.result, result)
		}
	}
}

func TestToFraction(t *testing.T) {
	tests := []struct {
		from     string
		d1 string
		d2 string
	}{
		{"0", "0", "1"},
		{"0.0", "0", "1"},

		{"0.1", "1", "10"},
		{"-0.1", "-1", "10"},
		{"0.12", "3", "25"},	// with simplified fraction, 3/25 instead of 12/100
		{"-0.12", "-3", "25"},

		{"1", "1", "1"},
		{"-1", "-1", "1"},
		{"123", "123", "1"},
		{"-123", "-123", "1"},
		{"1.1", "11", "10"},
		{"-1.1", "-11", "10"},
		{"123.4", "617", "5"},
		{"-123.4", "-617", "5"},
		{"123.45", "2469", "20"},
		{"-123.45", "-2469", "20"},
		{"0.123", "123", "1000"},
		{"-0.123", "-123", "1000"},
		{"1.456789", "1456789", "1000000"},
		{"-1.456789", "-1456789", "1000000"},
	}

	for _, tt := range tests {
		d1, d2 := RequireFromString(tt.from).ToFraction()

		if d1.String() != tt.d1 || d2.String() != tt.d2 {
			t.Errorf("ToFraction() failed: expected=%s and %s, received=%s and %s",
				tt.d1, tt.d2, d1, d2)
		}
	}
}

func TestPow2(t *testing.T) {
	tests := []struct {
		base string
		exp string
		result string
	}{
		{"2", "3", "8"},
		{"2", "3.5", "11.313708498984760390413509793677585"},
		{"1.5", "34", "970739.7373664756887592375278472900390625"},
		{"1.5", "34.33", "1109718.748389653202836265514985705560694"},
	}

	for _, tt := range tests {
		base := RequireFromString(tt.base)
		exp := RequireFromString(tt.exp)
		result := base.Pow2(exp)

		if result.String() != tt.result {
			t.Errorf("Pow2() failed: expected=%s, received=%s",
				tt.result, result)
		}
	}
}

func TestGcd(t *testing.T) {
	tests := []struct {
		d1 string
		d2 string
		gcd string
	}{
		{"2", "3", "1"},
		{"35", "10", "5"},
		{"12", "4", "4"},
		{"11", "3", "1"},
	}

	for _, tt := range tests {
		d1 := RequireFromString(tt.d1)
		d2 := RequireFromString(tt.d2)
		gcd := Gcd(d1, d2)

		if gcd.String() != tt.gcd {
			t.Errorf("Gcd() failed: expected=%s, received=%s",
				tt.gcd, gcd)
		}
	}
}

func TestLcm(t *testing.T) {
	tests := []struct {
		d1 string
		d2 string
		lcm string
	}{
		{"2", "1", "2"},
		{"3", "9", "9"},
		{"3", "7", "21"},
	}

	for _, tt := range tests {
		d1 := RequireFromString(tt.d1)
		d2 := RequireFromString(tt.d2)
		lcm := Lcm(d1, d2)

		if lcm.String() != tt.lcm {
			t.Errorf("Lcm() failed: expected=%s, received=%s",
				tt.lcm, lcm)
		}
	}
}
