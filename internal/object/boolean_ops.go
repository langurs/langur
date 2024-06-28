// langur/object/boolean_ops.go

package object

func (left *Boolean) Multiply(o2 Object) Object {
	right, ok := o2.(IMultiply)
	if ok {
		// if no case for Boolean multiplication, will return nil
		return right.Multiply(left)
	}
	return nil
}
