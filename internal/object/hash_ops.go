// langur/object/hash_ops.go

package object

func (left *Hash) Multiply(o2 Object) Object {
	switch right := o2.(type) {
	case *Boolean:
		if right.Value {
			return left
		}
		return EmptyHash

	default:
		return nil
	}
}

func (left *Hash) Append(o2 Object) Object {
	if o2 == NONE {
		// append none; return original object
		return left
	}

	switch right := o2.(type) {
	case *Hash:
		return left.AppendWithOverWrite(right)
	}

	return nil
}

func (l *Hash) AppendToNone() Object {
	return l
}

func (l *Hash) Contains(value Object) (bool, bool) {
	for _, kv := range l.Pairs {
		if kv.Value.Equal(value) {
			return true, true
		}
	}
	return false, true
}
