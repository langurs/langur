// langur/decimal/langur_zeroes.go

// methods added to shopspring/decimal for langur
// These are here to treat trailing zeroes differently.
// see also langur/decimal/langur.go

package decimal

import (
	"langur/modes"
	"strings"
)

func (d Decimal) StringWithTrailingZeros() string {
	return d.string(false)
}

func (d Decimal) Simplify() Decimal {
	// remove trailing zeroes
	d2, _ := NewFromString(d.string(true))
	return d2
}

func (d Decimal) Same(d2 Decimal) bool {
	// Same different than Equal
	// 1 == 1.0 but they are not the same.
	// same meaning the representation is the same
	return d.string(false) == d2.string(false)
}

func (d Decimal) DivTruncate(d2 Decimal, scale int32) Decimal {
	return d.Div(d2).Truncate(scale)
}

func (d Decimal) DivFloor(d2 Decimal) Decimal {
	return d.Div(d2).Floor()
}

func (d Decimal) TruncateWithZeroes(places int32, trimTrailingZeroes bool) Decimal {
	// The decimal package treats zeroes differently for truncation than for rounding, ...
	// ... whether trailing zeroes or using a negative for places to work on the integer.

	d.ensureInitialized()
	if places >= 0 && -places > d.exp {
		d = d.rescale(-places)
	}

	parts := strings.Split(d.StringWithTrailingZeros(), ".")

	if places == 0 {
		d, _ = NewFromString(parts[0])
		return d

	} else if places < 0 {
		sc := int(-places)
		// truncate on integer portion
		L := len(parts[0])
		if L <= sc {
			return Zero
		}

		d, _ = NewFromString(parts[0][:L-sc] + strings.Repeat("0", sc))
		return d

	} else {
		// scale > 0
		switch len(parts) {
		case 2:
			if len(parts[1]) < int(places) {
				parts[0] += "." + parts[1] + strings.Repeat("0", int(places)-len(parts[1]))
				d, _ = NewFromString(parts[0])
			}

		case 1:
			parts[0] += "." + strings.Repeat("0", int(places))
			d, _ = NewFromString(parts[0])
		}
	}

	if trimTrailingZeroes {
		return d.Simplify()
	}
	return d
}

func (d Decimal) DivWithMinMaxScale(d2 Decimal) Decimal {
	// minimum scale determined by first number; This might change.
	minScale := decimalScale(d)
	div := d.Div(d2)
	return div.RescaleMin(minScale, true)
}

func decimalScale(d Decimal) int {
	ns := d.string(false)
	idx := strings.IndexByte(ns, '.')
	if idx == -1 {
		return 0
	}
	return len(ns) - idx - 1
}

func (d Decimal) RescaleMin(minScale int, withDivMax bool) Decimal {
	// max scale handled by DivisionPrecision
	// check min scale and remove extra zeroes at same time
	ds := d.string(true) // true to remove extra zeroes that the decimal library adds

	// 0.13 bug fix
	if withDivMax && minScale > DivisionPrecision {
		minScale = DivisionPrecision
	}

	if minScale > 0 {
		parts := strings.Split(ds, ".")
		if len(parts) == 1 {
			ds = parts[0] + "." + strings.Repeat("0", minScale)
		} else {
			divScale := len(parts[1])
			if divScale < minScale {
				ds = parts[0] + "." + parts[1] + strings.Repeat("0", minScale-divScale)
			}
		}
	}

	var err error
	d, err = NewFromString(ds)
	if err != nil {
		decThrow(err.Error())
	}
	return d
}

// RoundByMode()
// here both for trailing zeroes and for using a mode setting
func (d Decimal) RoundByMode(places int32, trimTrailingZeroes bool, mode int) Decimal {
	var d2 Decimal
	switch mode {
	case modes.Round_halfAwayFromZero:
		d2 = d.Round(places)
	case modes.Round_halfEven:
		d2 = d.RoundBank(places)
	default:
		decThrow("Invalid Rounding Mode")
		return Zero
	}
	if trimTrailingZeroes {
		return d2.Simplify()
	}
	return d2
}
