// langur/decimal/langur_zeroes.go

// methods added to shopspring/decimal for langur
// These are here to treat trailing zeroes differently.

package decimal

import (
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
