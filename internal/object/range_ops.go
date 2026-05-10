// langur/object/range_ops.go

package object

import (
	"fmt"
)

func (l *Range) Append(o2 Object) Object {
	var rSlc2 []rune
	var s *String

	rSlc, err := l.toRuneSlice()

	switch r := o2.(type) {
	case *String:
		if err == nil {
			s, err = NewStringFromParts(rSlc, r.String())
		}

	case *Range:
		if err == nil {
			rSlc2, err = r.toRuneSlice()
			if err == nil {
				s, err = NewStringFromParts(rSlc, rSlc2)
			}
		}

	case *Number:
		if err == nil {
			var cp2 rune
			cp2, err = r.ToRune()
			if err == nil {
				s, err = NewStringFromParts(rSlc, cp2)
			}
		}

	default:
		return nil
	}

	if err != nil {
		return NewError(ERR_GENERAL, "Append", fmt.Sprintf("failure to append to range: %s", err.Error()))
	}

	return s
}

func (l *Range) AppendToNone() Object {
	r, err := l.ToString()
	if err != nil {
		return NewError(ERR_GENERAL, "Append", fmt.Sprintf("failure to convert range to string: %s", err.Error()))
	}
	return r
}

func (r *Range) Contains(value Object) (bool, bool) {
	if r.Start.Type() == value.Type() {
		var high, low Object
		if test, _ := GreaterOrEqual(r.End, r.Start); test {
			high, low = r.End, r.Start
		} else {
			low, high = r.End, r.Start
		}
		test, _ := GreaterOrEqual(high, value)
		if test {
			test, _ = GreaterOrEqual(value, low)
		}
		return test, true
	}
	return false, false
}
