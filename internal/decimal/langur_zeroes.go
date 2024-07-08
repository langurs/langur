// langur/decimal/langur_zeroes.go

// methods added to shopspring/decimal for langur
// These are here to treat trailing zeroes differently.
// see also langur/decimal/langur.go

package decimal

import (
	"langur/modes"
	"strings"
)

const (
	RoundingMode_HalfAwayFromZero = modes.Round_halfAwayFromZero
	RoundingMode_HalfEven         = modes.Round_halfEven
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

func (d Decimal) DivTruncate(d2 Decimal, places int32) Decimal {
	return d.Div(d2).Truncate(places)
}

func (d Decimal) DivFloor(d2 Decimal) Decimal {
	return d.Div(d2).Floor()
}

// custom decimal truncate function to...
// 1. add trailing zeroes
// 2. trim trailing zeroes
// 3. use a negative for places, to truncate on the integer
func (d Decimal) TruncateWithZeroes(
	places int32, addTrailingZeroes, trimTrailingZeroes bool) Decimal {

	d.ensureInitialized()
	if places >= 0 && -places > d.exp {
		d = d.rescale(-places)
	}

	parts := strings.Split(d.StringWithTrailingZeros(), ".")

	if places == 0 {
		// integer only
		d, _ = NewFromString(parts[0])
		return d

	} else if places < 0 {
		// truncate on integer portion
		sc := int(-places)
		L := len(parts[0])
		if L <= sc {
			return Zero
		}
		d, _ = NewFromString(parts[0][:L-sc] + strings.Repeat("0", sc))
		return d

	} else {
		// places > 0
		p2 := ""
		if len(parts) == 2 {
			p2 = parts[1]
		}
		if addTrailingZeroes && len(p2) < int(places) {
			p2 += strings.Repeat("0", int(places)-len(p2))
		}
		if trimTrailingZeroes {
			p2 = strings.TrimRight(p2, "0")
		}

		if p2 == "" {
			d, _ = NewFromString(parts[0])
		} else {
			d, _ = NewFromString(parts[0] + "." + p2)
		}
		return d
	}
}

// custom decimal rounding function to...
// add trailing zeroes
// trim trailing zeroes
// using a mode setting
func (d Decimal) RoundByMode(places int32, addTrailingZeroes, trimTrailingZeroes bool, mode int) Decimal {
	var d2 Decimal
	switch mode {
	case RoundingMode_HalfAwayFromZero:
		d2 = d.Round(places)
	case RoundingMode_HalfEven:
		d2 = d.RoundBank(places)
	default:
		decThrow("Invalid Rounding Mode")
		return Zero
	}
	if !addTrailingZeroes && decimalScale(d2) > decimalScale(d) {
		// The decimal functions used above may add trailing zeroes.
		// Here we trim just the added zeroes and no more.
		parts := strings.Split(d2.string(true), ".")
		d2, _ = NewFromString(parts[0] + "." + parts[1][:decimalScale(d)])
	}
	if trimTrailingZeroes {
		return d2.Simplify()
	}
	return d2
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
