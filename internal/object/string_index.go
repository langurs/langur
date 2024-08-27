// langur/object/string_index.go

package object

import (
	"fmt"
)

// returnOtherObjType: string instead of code point(s)
func (left *String) Index(index Object, returnOtherObjType bool) (result Object, err error) {
	result, err = left.index(index, returnOtherObjType)
	if err != nil {
		return left, fmt.Errorf("Index out of range")
	}
	return
}

func (left *String) index(index Object, returnOtherObjType bool) (
	result Object, err error) {

	switch idx := index.(type) {
	case *Number:
		n, ok := left.IndexNativeInt(idx)
		if !ok {
			return left, fmt.Errorf("String index not an integer or out of range")
		}
		if returnOtherObjType {
			s, err := NewStringFromParts(left.RuneSlc()[n])
			return s, err
		}
		return NumberFromRune(left.RuneSlc()[n]), nil

	case nil:
		// implicit range; may be called from built-in functions
		if returnOtherObjType {
			return left, nil
		}
		return ListFromRuneSlice(left.RuneSlc()), nil

	case *Range, *List:
		intIdx, err := makeNativeIntIndexSlice(left, index)
		if err != nil {
			return nil, err
		}

		orig := left.RuneSlc()
		rSlc := []rune{}
		for _, n := range intIdx {
			rSlc = append(rSlc, orig[n])
		}

		if returnOtherObjType {
			s, err := NewStringFromParts(rSlc)
			return s, err
		}
		return ListFromRuneSlice(rSlc), nil

	default:
		// invalid index type
		return left, fmt.Errorf("Invalid index type for string (%s)", idx.TypeString())
	}
}

func (left *String) IndexValid(index Object) bool {
	switch idx := index.(type) {
	case *Number:
		_, ok := left.IndexNativeInt(idx)
		return ok

	case *Range:
		_, ok := left.IndexNativeInt(idx.Start)
		if ok {
			_, ok = left.IndexNativeInt(idx.End)
		}
		return ok

	case *List:
		for _, v := range idx.Elements {
			if !left.IndexValid(v) {
				return false
			}
		}
		return true

	default:
		return false
	}
}

// assumes mutability of an object (checked elsewhere)
func (left *String) SetIndex(index, setTo Object) (Object, error) {
	n, ok := left.IndexNativeInt(index)
	if !ok {
		return left, fmt.Errorf("Cannot set string index value from invalid index (not an integer)")
	}

	var piece string
	switch to := setTo.(type) {
	case *Number:
		n2, err := to.ToRune()
		if err != nil {
			return left, fmt.Errorf("Cannot set string index value from a non-integer (not a code point)")
		}
		piece = string(n2)

	case *String:
		piece = to.String()
		if piece == "" {
			return left, fmt.Errorf("Cannot set string index value from empty string")
		}
		// one code point; leaves out others if there are more
		piece = string([]rune(piece)[0])

	default:
		return left, fmt.Errorf("Cannot set index of string from type %s", left.TypeString())
	}

	cpSlc := left.RuneSlc()
	return NewStringFromParts(cpSlc[:n], piece, cpSlc[n+1:])
}

func (left *String) IndexNativeInt(index Object) (idx int, ok bool) {
	idx, ok = NumberToInt(index)
	if !ok {
		return
	}
	return left.indexNativeInt(idx)
}

// receives 1-based integer index (langur)
// returns 0-based integer index (Go)
// if applicable, flips negative index to positive
// ok true/false depending on validity
func (left *String) indexNativeInt(index int) (idx int, ok bool) {
	idx = index

	cpCnt := left.LenCP()
	if idx < 0 {
		// flip negative index
		idx = cpCnt + idx + 1

	} else if idx > cpCnt {
		return 0, false
	}

	if idx < 1 {
		return 0, false
	}

	// All is well.
	// convert 1-based (langur) index to 0-based (native Go) and return
	return idx - 1, true
}

func (s *String) RemoveIndices(indices Object) (*String, error) {
	// build new string without indices we want to remove
	cpSlc := []rune{}

	intIdx, err := makeNativeIntIndexSlice(s, indices)
	if err != nil {
		return nil, err
	}

	for i, cp := range s.RuneSlc() {
		if !intInSlice(i, intIdx) {
			cpSlc = append(cpSlc, cp)
		}
	}

	return NewStringFromParts(cpSlc)
}
