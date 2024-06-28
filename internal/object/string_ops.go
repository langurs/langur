// langur/object/string_ops.go

package object

import (
	"fmt"
	"strings"
)

func (s *String) HashKey() Object {
	return s
}

func (l *String) Append(o2 Object) Object {
	if o2 == NONE {
		// append none; return original string object
		return l
	}

	var err error
	var s *String

	switch r := o2.(type) {
	case *String:
		return NewString(l.String() + r.String())

	case *Number:
		var cp rune
		cp, err = r.ToRune()
		if err == nil {
			s, err = NewStringFromParts(l.String(), cp)
		}

	case *Range:
		var rSlc []rune
		rSlc, err = r.toRuneSlice()
		if err == nil {
			s, err = NewStringFromParts(l.String(), rSlc)
		}

	default:
		return nil
	}

	if err != nil {
		return NewError(ERR_GENERAL, "Append", fmt.Sprintf("failure to append to string: %s", err.Error()))
	}

	return s
}

func (l *String) AppendToNone() Object {
	return l
}

func (l *String) Multiply(o2 Object) Object {
	switch r := o2.(type) {
	case *Number:
		n, err := r.ToInt()
		if err != nil {
			return NewError(ERR_GENERAL, "Multipy", "could to convert to int for string multiplication")
		}
		// negative number same as 0
		if n < 1 {
			return ZLS
		}
		return NewString(strings.Repeat(l.String(), n))

	case *Boolean:
		if r.Value {
			return l
		}
		return ZLS

	default:
		return nil
	}
}

func (s *String) Contains(value Object) (bool, bool) {
	var sub string
	switch v := value.(type) {
	case *Number:
		r, ok := NumberToRune(v)
		if !ok {
			return false, false
		}
		sub = string(r)

	case *String:
		sub = v.String()

	default:
		return false, false
	}

	index, err := StringIndex(sub, s.String())
	if err != nil || index == NONE {
		return false, true
	}

	return true, true
}
