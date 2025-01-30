// langur/object/number_ops.go

package object

import (
	"fmt"
	"langur/native"
)

func (left *Number) Negate() Object {
	if left.usingIntOptimization {
		n, ok := native.NegateInt64(left.integer)
		if ok {
			return NumberFromInt64(n)
		}
	}
	return numberFromDecimal(left.ToDecimal().Neg())
}

func (left *Number) Abs() Object {
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

func (left *Number) Append(o2 Object) Object {
	var s *String

	if o2 == NONE {
		return left.AppendToNone()
	}

	cp1, err := left.ToRune()

	switch right := o2.(type) {
	case *Number:
		if err == nil {
			// generate string, concatenating 2 code points
			var cp2 rune
			cp2, err = right.ToRune()
			if err == nil {
				s, err = NewStringFromParts(cp1, cp2)
			}
		}

	case *String:
		if err == nil {
			// append code point and string
			s, err = NewStringFromParts(cp1, right.String())
		}

	case *Range:
		if err == nil {
			// append code point and range of code points
			var rSlc []rune
			rSlc, err = right.toRuneSlice()
			if err == nil {
				s, err = NewStringFromParts(cp1, rSlc)
			}
		}

	default:
		return nil
	}

	if err != nil {
		return NewError(ERR_GENERAL, "Append", fmt.Sprintf("failed to append to number: %s", err.Error()))
	}

	return s
}

func (l *Number) AppendToNone() Object {
	n, err := l.ToRune()
	if err != nil {
		return NewError(ERR_GENERAL, "Append", "number not an integer usable for code point")

	} else if n < 0 {
		return NewError(ERR_GENERAL, "Append", "number not an integer usable for code point")

	}
	s, err := NewStringFromParts(n)
	if err == nil {
		return s
	}
	return NewError(ERR_GENERAL, "Append", err.Error())
}

func (left *Number) Add(o2 Object) Object {
	switch right := o2.(type) {
	case *Number:
		if left.usingIntOptimization && right.usingIntOptimization {
			n, ok := native.AddInt64(left.integer, right.integer)
			if ok {
				return NumberFromInt64(n)
			}
		}
		// failing that for overflow or non-integers
		return numberFromDecimal(left.ToDecimal().Add(right.ToDecimal()))

	case *Complex:
		// simply reverse the operands in this case (for addition won't affect result)
		return right.Add(left)
	}

	return nil
}

func (left *Number) Subtract(o2 Object) Object {
	switch right := o2.(type) {
	case *Number:
		if left.usingIntOptimization && right.usingIntOptimization {
			n, ok := native.SubInt64(left.integer, right.integer)
			if ok {
				return NumberFromInt64(n)
			}
		}
		// failing that for overflow or non-integers
		return numberFromDecimal(left.ToDecimal().Sub(right.ToDecimal()))
		
	case *Complex:
		return NewComplex(left, Zero).Subtract(right)
	}

	return nil
}

func (left *Number) Multiply(o2 Object) Object {
	switch right := o2.(type) {
	case *Number:
		if left.usingIntOptimization && right.usingIntOptimization {
			n, ok := native.MultiplyInt64(left.integer, right.integer)
			if ok {
				return NumberFromInt64(n)
			}
		}
		return numberFromDecimal(left.ToDecimal().Mul(right.ToDecimal()))

	case *Complex:
		return right.Multiply(left)
	case *String:
		return right.Multiply(left)
	case *List:
		return right.Multiply(left)

	case *Boolean:
		if right.Value {
			return left
		}
		return Zero
	}

	return nil
}

func (left *Number) Divide(o2 Object) Object {
	switch right := o2.(type) {
	case *Number:
		if left.usingIntOptimization && right.usingIntOptimization {
			n, ok := native.DivideInt64(left.integer, right.integer)
			if ok {
				return NumberFromInt64(n)
			}
		}
		return numberFromDecimal(left.ToDecimal().DivWithMinMaxScale(right.ToDecimal()))
	
	case *Complex:
		return NewComplex(left, Zero).Divide(right)
	}

	return nil
}

func (left *Number) DivideTruncate(o2 Object) Object {
	switch right := o2.(type) {
	case *Number:
		if left.usingIntOptimization && right.usingIntOptimization {
			n, ok := native.DivideTruncateInt64(left.integer, right.integer)
			if ok {
				return NumberFromInt64(n)
			}
		}
		return numberFromDecimal(left.ToDecimal().DivTruncate(right.ToDecimal(), 0))
	}

	return nil
}

func (left *Number) DivideFloor(o2 Object) Object {
	switch right := o2.(type) {
	case *Number:
		if left.usingIntOptimization && right.usingIntOptimization {
			n, ok := native.DivideFloorInt64(left.integer, right.integer)
			if ok {
				return NumberFromInt64(n)
			}
		}
		return numberFromDecimal(left.ToDecimal().DivFloor(right.ToDecimal()))
	}

	return nil
}

func (left *Number) DivisibleBy(o2 Object) (bool, bool) {
	// "divisible" meaning evenly divisible
	switch right := o2.(type) {
	case *Number:
		n := left.Remainder(right)
		if n != nil {
			return n.(*Number).IsZero(), true
		}
	}

	return false, false
}

func (left *Number) Remainder(o2 Object) Object {
	switch right := o2.(type) {
	case *Number:
		if left.usingIntOptimization && right.usingIntOptimization {
			if right.integer != 0 {
				n, ok := native.RemainderInt64(left.integer, right.integer)
				if ok {
					return NumberFromInt64(n)
				}
			}
		}
		return numberFromDecimal(left.ToDecimal().Mod(right.ToDecimal()))
	}

	return nil
}

func (left *Number) Modulus(o2 Object) Object {
	switch right := o2.(type) {
	case *Number:
		if left.usingIntOptimization && right.usingIntOptimization {
			n, ok := native.ModulusInt64(left.integer, right.integer)
			if ok {
				return NumberFromInt64(n)
			}
		}
		return numberFromDecimal(left.ToDecimal().TrueMod(right.ToDecimal()))
	}

	return nil
}

func (left *Number) Power(o2 Object) Object {
	switch right := o2.(type) {
	case *Number:
		return numberFromDecimal(left.ToDecimal().Pow(right.ToDecimal()))
	}

	return nil
}

func (left *Number) Root(o2 Object) Object {
	switch right := o2.(type) {
	case *Number:
		return numberFromDecimal(left.ToDecimal().Root(right.ToDecimal()))
	}

	return nil
}
