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
		// ... that is, since i squared == -1 ...
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

// func (left *Complex) Divide(o2 Object) Object {
// 	switch right := o2.(type) {}

// 	return nil
// }

// func (left *Complex) Power(o2 Object) Object {
// 	switch right := o2.(type) {
// 	}

// 	return nil
// }
