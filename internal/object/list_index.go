// langur/object/list_index.go

package object

import (
	"fmt"
)

// returnOtherObjType: unused on list
func (left *List) Index(index Object, returnOtherObjType bool) (result Object, err error) {
	result, err = left.index(index, returnOtherObjType)
	if err != nil {
		return left, fmt.Errorf("Index out of range")
	}
	return
}

func (left *List) index(index Object, returnOtherObjType bool) (
	result Object, err error) {

	switch idx := index.(type) {
	case *Number:
		n, ok := left.IndexNativeInt(idx)
		if !ok {
			return left, fmt.Errorf("List index not an integer, or out of range for native integer type")
		}
		return left.Elements[n], nil

	case *Range, *List:
		intIdx, err := makeNativeIntIndexSlice(left, index)
		if err != nil {
			return nil, err
		}
		list := &List{}
		for _, n := range intIdx {
			list.Elements = append(list.Elements, left.Elements[n])
		}
		return list, nil

	default:
		// invalid index type
		return left, fmt.Errorf("Invalid index type for list")
	}
}

func (left *List) IndexInverse(index Object, returnOtherObjType bool) (
	result Object, err error) {

	switch index.(type) {
	case *Range, *List, *Number:
		intIdx, err := makeNativeIntIndexSlice(left, index)
		if err != nil {
			return nil, err
		}
		list := &List{}
		for n := range left.Elements {
			if !intInSlice(n, intIdx) {
				list.Elements = append(list.Elements, left.Elements[n])
			}
		}
		return list, nil

	default:
		// invalid index type
		return left, fmt.Errorf("Invalid index type for list")
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

func intInSlice(i int, iSlc []int) bool {
	for _, n := range iSlc {
		if i == n {
			return true
		}
	}
	return false
}

func makeNativeIntIndexSlice(obj, index Object) (iSlc []int, err error) {
	resolve := func(n int, add bool) (int, bool) {
		iii, ok := obj.(IIndexNativeInt)
		if ok {
			n, ok = iii.indexNativeInt(n)
			if ok {
				// resolved and valid index; add to slice
				if add {
					iSlc = append(iSlc, n)
				}
				return n, true
			}
		}
		return 0, false
	}

	var recursive func(Object) error
	recursive = func(index Object) error {
		switch idx := index.(type) {
		case *Number:
			n, ok := NumberToInt(idx)
			if !ok {
				return fmt.Errorf("Number not an integer")
			}
			if _, ok := resolve(n, true); !ok {
				return fmt.Errorf("Could not resolve integer index")
			}

		case *Range:
			// resolve start and end first; could be negative indices
			start, ok := NumberToInt(idx.Start)
			if !ok {
				return fmt.Errorf("Start of range not an integer")
			}
			start, ok = resolve(start, false)
			if !ok {
				return fmt.Errorf("Start of range not resolvable")
			}
			end, ok := NumberToInt(idx.End)
			if !ok {
				return fmt.Errorf("End of range not an integer")
			}
			end, ok = resolve(end, false)
			if !ok {
				return fmt.Errorf("End of range not resolvable")
			}
			start++
			end++

			if end < start {
				for n := start; n > end-1; n-- {
					resolve(n, true)
				}

			} else {
				for n := start; n < end+1; n++ {
					resolve(n, true)
				}
			}

		case *List:
			for _, item := range idx.Elements {
				err := recursive(item)
				if err != nil {
					return err
				}
			}

		default:
			return fmt.Errorf("Expected integer, range of integers, or list of such")
		}

		return nil
	}

	err = recursive(index)
	return
}
