// langur/object/range_index.go

package object

import (
	"fmt"
)

// returnOtherObjType: unused on range
func (left *Range) Index(index Object, returnOtherObjType bool) (result Object, err error) {
	result, err = left.index(index, returnOtherObjType)
	if err != nil {
		return left, fmt.Errorf("Index out of range")
	}
	return
}

func (left *Range) index(index Object, returnOtherObjType bool) (result Object, err error) {
	switch idx := index.(type) {
	case nil:
		return left, nil

	case *Number:
		n, ok := left.IndexNativeInt(idx)
		if !ok {
			return left, fmt.Errorf("Range index not an integer or is out of range")
		}
		switch n {
		case 0: // 1 in langur
			return left.Start, nil
		case 1: // 2 in langur
			return left.End, nil
		default:
			return left, fmt.Errorf("Range index out of range")
		}

	case *List:
		arr := &List{}
		for _, v := range idx.Elements {
			elements, err := left.index(v, returnOtherObjType)
			if err != nil {
				return left, err
			}
			switch e := elements.(type) {
			case *List:
				arr.Elements = append(arr.Elements, e.Elements...)
			default:
				arr.Elements = append(arr.Elements, e)
			}
		}
		return arr, nil

	default:
		// invalid index type
		return left, fmt.Errorf("Invalid index type for range")
	}
}

func (left *Range) IndexValid(index Object) bool {
	switch idx := index.(type) {
	case *Number:
		_, ok := left.IndexNativeInt(idx)
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
func (left *Range) SetIndex(index, setTo Object) (Object, error) {
	n, ok := left.IndexNativeInt(index)
	if !ok {
		return left, fmt.Errorf("Cannot set range index value from invalid index (not an integer or out of range)")
	}
	left = left.CopyRefs().(*Range)

	switch n {
	case 0: // 1 in langur
		if !RangeValid(setTo, left.End) {
			return left, fmt.Errorf("Cannot set start of range to type incompatible with end of range type")
		}
		left.Start = setTo

	case 1: // 2 in langur
		if !RangeValid(left.Start, setTo) {
			return left, fmt.Errorf("Cannot set end of range to type incompatible with start of range type")
		}
		left.End = setTo

	default:
		return left, fmt.Errorf("Cannot set range index value from invalid index")
	}

	return left, nil
}

// receives 1-based integer index
// returns 0-based integer index
func (left *Range) IndexNativeInt(index Object) (idx int, ok bool) {
	idx, ok = NumberToInt(index)
	if !ok {
		return
	}

	switch idx {
	case 1, 2:
		// good
		// convert 1-based index to 0-based (native) and return
		return idx - 1, true
	default:
		// bad
		// NOTE: intentionally not doing negative indices on range
		return 0, false
	}
}
