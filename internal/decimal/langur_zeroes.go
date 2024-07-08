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
	d, _ = NewFromString(d.string(true))
	return d
}

func (d Decimal) Same(d2 Decimal) bool {
	// Same different than Equal
	// 1 == 1.0 but they are not the same.
	// same meaning the precision is also the same
	return d.string(false) == d2.string(false)
}

func (d Decimal) DivTruncate(d2 Decimal, places int32) Decimal {
	return d.Div(d2).Truncate(places)
}

func (d Decimal) DivFloor(d2 Decimal) Decimal {
	return d.Div(d2).Floor()
}

// custom decimal truncate function to...
// 1. add trailing zeroes (if more precision than original decimal)
// 2. trim trailing zeroes
// 3. use a negative for places, to truncate on the integer
func (d Decimal) TruncateWithZeroes(
	places int32, addTrailingZeroes, trimTrailingZeroes bool) Decimal {

	d.ensureInitialized()
	if places >= 0 && -places > d.exp {
		d = d.rescale(-places)
	}

	parts := d.StringParts()

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
		if addTrailingZeroes && len(parts[1]) < int(places) {
			parts[1] += strings.Repeat("0", int(places)-len(parts[1]))
		}
		if trimTrailingZeroes {
			parts[1] = strings.TrimRight(parts[1], "0")
		}
		return NewFromParts(parts)
	}
}

func (d Decimal) StringParts() [2]string {
	parts := strings.Split(d.string(false), ".")
	if len(parts) == 1 {
		parts = append(parts, "")
	}
	return [2]string(parts)
}

func NewFromParts(parts [2]string) Decimal {
	d, err := NewFromString(parts[0] + "." + parts[1])
	if err != nil {
		decThrow(err.Error())
	}
	return d
}

// custom decimal rounding function to...
// add trailing zeroes (if more precision than original decimal)
// trim trailing zeroes
// using a mode setting
func (d Decimal) RoundByMode(
	places int32, addTrailingZeroes, trimTrailingZeroes bool,
	mode int) Decimal {

	originalScale := decimalScale(d)

	if originalScale < int(places) {
		// nothing to round
		// add zeroes?
		if addTrailingZeroes && !trimTrailingZeroes {
			parts := d.StringParts()
			parts[1] += strings.Repeat("0", int(places)-originalScale)
			d = NewFromParts(parts)
		}

	} else {
		switch mode {
		case modes.Round_halfAwayFromZero:
			d = d.Round(places)
		case modes.Round_halfEven:
			d = d.RoundBank(places)
		default:
			decThrow("Invalid Rounding Mode")
			return Zero
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
	ds := d.string(true)

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
