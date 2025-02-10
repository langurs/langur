// langur/object/complex_ops.go

package object

func (left *Complex) Negate() Object {
	return NewComplex(left.real.Negate().(*Number), left.imaginary.Negate().(*Number))
}

func (left *Complex) Abs() Object {
	// (real^2 + imaginary^2) ^/ 2
	return left.real.Multiply(left.real).(*Number).
		Add(left.imaginary.Multiply(left.imaginary)).(*Number).
			Root(Two)
}

func (left *Complex) Add(o2 Object) Object {
	switch right := o2.(type) {
	case *Complex:
		return NewComplex(
			left.real.Add(right.real).(*Number), 
			left.imaginary.Add(right.imaginary).(*Number),
		)

	case *Number:
		return NewComplex(
			left.real.Add(right).(*Number), 
			left.imaginary,
		)
	}

	return nil
}

func (left *Complex) Subtract(o2 Object) Object {
	switch right := o2.(type) {
	case *Complex:
		return NewComplex(
			left.real.Subtract(right.real).(*Number), 
			left.imaginary.Subtract(right.imaginary).(*Number),
		)

	case *Number:
		return NewComplex(
			left.real.Subtract(right).(*Number), 
			left.imaginary,
		)
	}

	return nil
}

func (left *Complex) Multiply(o2 Object) Object {
	switch right := o2.(type) {
	case *Complex:
		// https://www.mathsisfun.com/algebra/complex-number-multiply.html
		// simplified from the FOIL method for quicker/simpler application
		// ... that is, since i^2 == -1 ...
		// (a+bi)(c+di) = (ac−bd) + (ad+bc)i

		// real: (ac−bd)
		real := left.real.Multiply(right.real).(*Number).
			Subtract(left.imaginary.Multiply(right.imaginary)).(*Number)

		// imaginary: (ad+bc)i
		imaginary := left.real.Multiply(right.imaginary).(*Number).
			Add(left.imaginary.Multiply(right.real)).(*Number)

		return NewComplex(real, imaginary)		
		
	case *Number:
		// convert number to complex and call self
		return left.Multiply(NewComplex(right, Zero))
	}

	return nil
}

func (left *Complex) Divide(o2 Object) Object {
	switch right := o2.(type) {
	case *Complex:
		// https://www.cuemath.com/numbers/division-of-complex-numbers/
		// ((ac+bd) / (c^2+d^2)) + ((bc−ad) / (c^2+d^2))i

		// denominator: c^2+d^2
		denominator := right.real.Multiply(right.real).(*Number).
			Add(right.imaginary.Multiply(right.imaginary)).(*Number)

		// real: ((ac+bd) / (c^2+d^2))
		real := left.real.Multiply(right.real).(*Number).				// ac
			Add(left.imaginary.Multiply(right.imaginary)).(*Number).	// +bd
				Divide(denominator).(*Number)

		// imaginary: ((bc−ad) / (c^2+d^2))i
		imaginary := left.imaginary.Multiply(right.real).(*Number).	// bc
			Subtract(left.real.Multiply(right.imaginary)).(*Number).	// -ad
				Divide(denominator).(*Number)

		return NewComplex(real, imaginary)

	case *Number:
		// convert number to complex and call self
		return left.Divide(NewComplex(right, Zero))
	}

	return nil
}

func (left *Complex) Power(o2 Object) Object {
	switch right := o2.(type) {
	case *Number:
		power, err := right.ToInt()
		if err != nil {
			return NewError(ERR_MATH, "^", "Cannot calculate non-integer power on complex")
		}
		inverse := power < 0
		if inverse {
			power = -power
		}

		result := NewComplex(One, Zero)
		for i := 0; i < power; i++ {
			result = result.Multiply(left).(*Complex)
		}

		if inverse {
			result = NewComplex(One, Zero).Divide(result).(*Complex)
		}
		return result

	// case *Complex:
		// need a log function?
		// return Exp(power * Log(a))		
	}

	return nil
}
