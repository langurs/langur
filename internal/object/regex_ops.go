// langur/object/regex_ops.go

package object

func (r *Regex) Forward(o2 Object) Object {
	result, err := RegexMatchingOrError(r, o2)
	if err != nil {
		// return NewError(ERR_GENERAL, "->", err.Error())
		return nil
	}
	return result
}
