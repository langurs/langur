// langur/object/list_ops.go

package object

func (left *List) Append(o2 Object) Object {
	if o2 == NONE {
		// append none; return original object
		return left
	}

	switch right := o2.(type) {
	case *List:
		newElements := make([]Object, len(left.Elements)+len(right.Elements))
		copy(newElements, left.Elements)
		copy(newElements[len(left.Elements):], right.Elements)
		return &List{Elements: newElements}
	}

	return nil
}

func (l *List) AppendToNone() Object {
	return l
}

func (left *List) Multiply(o2 Object) Object {
	switch right := o2.(type) {
	case *Number:
		n, err := right.ToInt()
		if err != nil {
			return NewError(ERR_GENERAL, "Multiply", "failure to convert number to integer for list multiplication")
		}
		L := len(left.Elements)

		// negative number same as 0
		if n < 1 || L == 0 {
			return &List{}
		} else if n == 1 {
			return left
		}
		arr := &List{Elements: make([]Object, L*n)}

		for i := 0; i < n; i++ {
			copy(arr.Elements[i*L:], left.Elements)
		}

		return arr

	case *Boolean:
		if right.Value {
			return left
		}
		return EmptyList()
	}

	return nil
}

func (l *List) Contains(value Object) (bool, bool) {
	for _, v := range l.Elements {
		if v.Equal(value) {
			return true, true
		}
	}
	return false, true
}
