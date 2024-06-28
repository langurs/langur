// langur/object/number_misc.go

package object

import (
	"fmt"
	dec "langur/decimal"
	"langur/modes"
	"langur/native"
	"math"
)

func (left *Number) Negate() *Number {
	if left.usingIntOptimization {
		n, ok := native.NegateInt64(left.integer)
		if ok {
			return NumberFromInt64(n)
		}
	}
	return numberFromDecimal(left.ToDecimal().Neg())
}

func (left *Number) Abs() *Number {
	if left.usingIntOptimization {
		if left.integer < 0 {
			n, ok := native.NegateInt64(left.integer)
			if ok {
				return NumberFromInt64(n)
			}
			// failed; use decimal
			left = left.UseDecimal()

		} else {
			return left
		}
	}

	if left.decimal.IsNegative() {
		return numberFromDecimal(left.decimal.Neg())
	}
	return left
}

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

func (n *Number) Round(max int) (*Number, error) {
	if max > math.MaxInt32 || max < math.MinInt32 {
		return Zero, fmt.Errorf("Number of digits to round to is too high")
	}
	return numberFromDecimal(n.ToDecimal().RoundWithZeroes(int32(max))), nil
}

func (n *Number) RoundBy(max, mode int) (*Number, error) {
	if max > math.MaxInt32 || max < math.MinInt32 {
		return Zero, fmt.Errorf("Number of digits to round to is too high")
	}
	_, ok := modes.RoundHashModeNames[mode]
	if !ok {
		return Zero, fmt.Errorf("Invalid Rounding Mode (use " + modes.RoundHashName + " hash)")
	}
	return numberFromDecimal(n.ToDecimal().RoundByWithZeroes(int32(max), mode)), nil
}

func (n *Number) Truncate(max int) (*Number, error) {
	if max > math.MaxInt32 || max < math.MinInt32 {
		return Zero, fmt.Errorf("Number of digits to truncate to is too high")
	}
	return numberFromDecimal(n.ToDecimal().TruncateWithZeroes(int32(max))), nil
}

func Gcd(a, b *Number) (*Number, error) {
	// greatest common divisor
	// use the Euclidian method for fast calculation with large numbers
	for !a.IsZero() {
		newA := b.Remainder(a)
		if newA == nil {
			return b, fmt.Errorf("failure determining remainder in calculating GCD")
		}
		a, b = newA.(*Number), a
	}
	return b, nil
}

func Lcm(a, b *Number) (*Number, error) {
	// least common multiple
	if gt, _ := b.GreaterThan(a); gt {
		a, b = b, a
	}
	gcd, err := Gcd(a, b)
	if err != nil {
		return a, err
	}
	result := a.Multiply(b)
	if result == nil {
		return a, fmt.Errorf("failure multiplying in calculating LCM")
	}
	result = result.(*Number).DivideTruncate(gcd)
	if result == nil {
		return nil, fmt.Errorf("failure division in calculating LCM")
	}
	return result.(*Number), nil
}

func (n *Number) Simplify() *Number {
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
