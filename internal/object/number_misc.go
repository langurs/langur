// langur/object/number_misc.go

package object

import (
	"fmt"
	dec "langur/decimal"
	"langur/modes"
	"math"
)

func (l *Number) Tangent() *Number {
	return numberFromDecimal(l.ToDecimal().Tan())
}

func (l *Number) ArcTangent() *Number {
	return numberFromDecimal(l.ToDecimal().Atan())
}

func (l *Number) Sine() *Number {
	return numberFromDecimal(l.ToDecimal().Sin())
}

func (l *Number) Cosine() *Number {
	return numberFromDecimal(l.ToDecimal().Cos())
}

func makeDecimalSliceFromNumberSlice(nSlc ...*Number) []decType {
	nums := make([]decType, len(nSlc))
	for i := range nSlc {
		nums[i] = nSlc[i].ToDecimal()
	}
	return nums
}

func Mid(n ...*Number) *Number {
	return numberFromDecimal(dec.Mid(makeDecimalSliceFromNumberSlice(n...)...))
}

func Mean(n ...*Number) *Number {
	return numberFromDecimal(dec.Mean(makeDecimalSliceFromNumberSlice(n...)...))
}

func (n *Number) Floor() *Number {
	return numberFromDecimal(n.ToDecimal().Floor())
}

func (n *Number) Ceiling() *Number {
	return numberFromDecimal(n.ToDecimal().Ceil())
}

func (n *Number) RoundByMode(
	max int,
	addTrailingZeroes, trimTrailingZeroes bool,
	mode modes.RoundingMode) (*Number, error) {

	if max > math.MaxInt32 || max < math.MinInt32 {
		return Zero, fmt.Errorf("Number of digits to round to is too high")
	}
	_, ok := modes.RoundHashModeNames[mode]
	if !ok {
		return Zero, fmt.Errorf("Invalid Rounding Mode (use " + modes.RoundHashName + " hash)")
	}
	rMode := modes.LangurRoundingModeToDecimalRoundingMode(mode)
	return numberFromDecimal(n.ToDecimal().RoundByMode(int32(max), addTrailingZeroes, trimTrailingZeroes, rMode)), nil
}

func (n *Number) Truncate(max int, addTrailingZeroes, trimTrailingZeroes bool) (*Number, error) {
	if max > math.MaxInt32 || max < math.MinInt32 {
		return Zero, fmt.Errorf("Number of digits to truncate to is too high")
	}
	return numberFromDecimal(n.ToDecimal().TruncateWithZeroes(int32(max), addTrailingZeroes, trimTrailingZeroes)), nil
}

// greatest common divisor
func Gcd(a, b *Number) (*Number, error) {
	return numberFromDecimal(dec.Gcd(a.ToDecimal(), b.ToDecimal())), nil
}

// least common multiple
func Lcm(a, b *Number) (*Number, error) {
	return numberFromDecimal(dec.Lcm(a.ToDecimal(), b.ToDecimal())), nil
}

func (n *Number) Simplify() Object {
	// remove trailing zeros
	if n.usingIntOptimization {
		return n
	}
	return numberFromDecimal(n.decimal.Simplify())
}

func (n *Number) IsPositive() bool {
	if n.usingIntOptimization {
		return n.integer >= 0
	}
	return n.decimal.IsPositive()
}

func (n *Number) IsNegative() bool {
	if n.usingIntOptimization {
		return n.integer < 0
	}
	return n.decimal.IsNegative()
}

func (n *Number) IsZero() bool {
	if n.usingIntOptimization {
		return n.integer == 0
	}
	return n.decimal.IsZero()
}

func (n *Number) IsInteger() bool {
	return n.usingIntOptimization || n.decimal.IsInteger()
}
