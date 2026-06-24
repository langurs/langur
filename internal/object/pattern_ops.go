// langur/object/pattern_ops.go

package object

func (r *Pattern) Forward(o2 Object) Object {
	result, err := PatternMatchingOrError(r, o2)
	if err != nil {
		// return NewError(ERR_GENERAL, "->", err.Error())
		return nil
	}
	return result
}
