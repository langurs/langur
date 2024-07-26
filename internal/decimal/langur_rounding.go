// langur/decimal/langur_rounding.go

package decimal

import (
	"strings"
)

type RoundingMode int

const (
	RoundingMode_Default RoundingMode = iota // might otherwise be called "half away from zero"
	RoundingMode_Bank                        // might otherwise be called "half even"
	RoundingMode_HalfUp
	RoundingMode_HalfDown
	RoundingMode_Ceiling
	RoundingMode_Floor
)

// custom decimal rounding function to...
// add trailing zeroes (if more precision than original decimal)
// trim trailing zeroes
// using a mode setting
func (d Decimal) RoundByMode(
	places int32, addTrailingZeroes, trimTrailingZeroes bool,
	mode RoundingMode) Decimal {

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
		case RoundingMode_Default:
			d = d.Round(places)
		case RoundingMode_Bank:
			d = d.RoundBank(places)
		case RoundingMode_HalfUp:
			d = d.RoundUp(places)
		case RoundingMode_HalfDown:
			d = d.RoundDown(places)
		case RoundingMode_Ceiling:
			d = d.RoundCeil(places)
		case RoundingMode_Floor:
			d = d.RoundFloor(places)
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
