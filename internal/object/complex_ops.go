// langur/object/complex_ops.go

package object

func (left *Complex) Negate() Object {
	return NewComplex(left.real.Negate().(*Number), left.imaginary.Negate().(*Number))
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

// func (left *Complex) Multiply(o2 Object) Object {
// 	switch right := o2.(type) {}

// 	return nil
// }

// func (left *Complex) Divide(o2 Object) Object {
// 	switch right := o2.(type) {}

// 	return nil
// }

// func (left *Complex) Power(o2 Object) Object {
// 	switch right := o2.(type) {
// 	}

// 	return nil
// }
