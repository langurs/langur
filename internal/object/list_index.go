// langur/object/list_index.go

package object

import (
	"fmt"
)

// returnOtherObjType: unused on list
func (left *List) Index(index Object, returnOtherObjType bool) (result Object, err error) {
	result, err, _ = left.index(index, returnOtherObjType)
	if err != nil {
		return left, fmt.Errorf("Index out of range")
	}
	return
}

func (left *List) index(index Object, returnOtherObjType bool) (result Object, err error, isPoly bool) {
	switch idx := index.(type) {
	case *Number:
		n, ok := left.IndexNativeInt(idx)
		if !ok {
			return left, fmt.Errorf("List index not an integer, or out of range for native integer type"), false
		}
		return left.Elements[n], nil, false

	case *Range:
		start, ok := left.IndexNativeInt(idx.Start)
		if !ok {
			return left, fmt.Errorf("List start of range index not an integer, or out of range for native integer type"), true
		}
		end, ok := left.IndexNativeInt(idx.End)
		if !ok {
			return left, fmt.Errorf("List end of range index not an integer, or out of range for native integer type"), true
		}

		// build a new list
		var elements []Object
		if start > end {
			// high to low range
			elements = make([]Object, 0, start-end+1)
			for i := start; i >= end; i-- {
				elements = append(elements, left.Elements[i])
			}

		} else {
			// low to high range
			elements = make([]Object, 0, end-start+1)
			for _, v := range left.Elements[start : end+1] {
				elements = append(elements, v)
			}
		}

		return &List{Elements: elements}, nil, true

	case *List:
		arr := &List{}
		for _, v := range idx.Elements {
			e, err, poly := left.index(v, returnOtherObjType)
			if err != nil {
				return left, err, poly
			}
			if poly {
				for _, e2 := range e.(*List).Elements {
					arr.Elements = append(arr.Elements, e2)
				}
			} else {
				arr.Elements = append(arr.Elements, e)
			}
		}
		return arr, nil, true

	default:
		// invalid index type
		return left, fmt.Errorf("Invalid index type for list"), false
	}
}

func (left *List) IndexValid(index Object) bool {
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
		// invalid index type
		return false
	}
}

// assumes mutability of an object (checked elsewhere)
func (left *List) SetIndex(index, setTo Object) (Object, error) {
	n, ok := left.IndexNativeInt(index)
	if !ok {
		return left, fmt.Errorf("Cannot set list value from invalid index (not an integer or out of range)")
	}

	// Since we don't know how many references there are to the list object we're changing, ...
	// ... we make a new one.
	left = left.CopyRefs().(*List)

	left.Elements[n] = setTo
	return left, nil
}

func (left *List) IndexNativeInt(index Object) (idx int, ok bool) {
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
func (left *List) indexNativeInt(index int) (idx int, ok bool) {
	idx = index

	if idx < 0 {
		// flip negative index
		idx = len(left.Elements) + idx + 1

	} else if idx > len(left.Elements) {
		return 0, false
	}

	if idx < 1 {
		return 0, false
	}

	// All is well.
	// convert 1-based index to 0-based (native) and return
	return idx - 1, true
}
