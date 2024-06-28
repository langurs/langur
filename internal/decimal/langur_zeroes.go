// langur/decimal/langur_zeroes.go

// methods added to shopspring/decimal for langur
// These are here to treat trailing zeroes differently.
// see also langur/decimal/langur.go

package decimal

import (
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

// if scale positive, includes trailing zeroes in truncated number
func (d Decimal) TruncateWithZeroes(scale int32) Decimal {
	padZeroes := true
	if scale < 0 {
		padZeroes = false
		scale = -scale
	}
	if !padZeroes && scale > int32(decimalScale(d)) {
		return d
	}
	return d.TruncateAndPad(int32(scale))
}

func (d Decimal) TruncateAndPad(scale int32) Decimal {
	d.ensureInitialized()
	if scale >= 0 && -scale > d.exp {
		d = d.rescale(-scale)
	}

	// FIXME(davis): ? do this in a rescale ?
	ds := strings.Split(d.StringWithTrailingZeros(), ".")

	if len(ds) == 2 && len(ds[1]) < int(scale) {
		ds[0] += "." + ds[1] + strings.Repeat("0", int(scale)-len(ds[1]))
		d, _ = NewFromString(ds[0])

	} else if len(ds) == 1 && scale > 0 {
		ds[0] += "." + strings.Repeat("0", int(scale))
		d, _ = NewFromString(ds[0])
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

// if max positive, includes trailing zeroes in rounded number
func (d Decimal) RoundWithZeroes(max int32) Decimal {
	padZeroes := true
	if max < 0 {
		padZeroes = false
		max = -max
	}
	if !padZeroes && max > int32(decimalScale(d)) {
		return d
	}
	return d.RoundByMode(max)
}

func (d Decimal) RoundByWithZeroes(max int32, mode int) Decimal {
	padZeroes := true
	if max < 0 {
		padZeroes = false
		max = -max
	}
	if !padZeroes && max > int32(decimalScale(d)) {
		return d
	}
	return d.RoundBy(max, mode)
}
