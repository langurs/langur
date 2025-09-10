// langur/object/complex_index.go

package object

import (
	"fmt"
)

func (left *Complex) IndexCount() int {
	return 2
}

func (left *Complex) IndexKeys() *List {
	return &List{Elements: []Object{NumberFromInt(1), NumberFromInt(2)}}
}

// returnOtherObjType: unused on complex
func (left *Complex) Index(index Object, returnOtherObjType bool) (result Object, err error) {
	result, err = left.index(index, returnOtherObjType)
	if err != nil {
		return left, fmt.Errorf("Index out of range")
	}
	return
}

func (left *Complex) index(index Object, returnOtherObjType bool) (result Object, err error) {
	if returnOtherObjType {
		return left, fmt.Errorf("No alternate return type for Complex")
	}
	
	switch idx := index.(type) {
	case nil:
		return left, nil

	case *Number:
		n, ok := left.IndexNativeInt(idx)
		if !ok {
			return left, fmt.Errorf("Complex index not an integer or is out of range")
		}
		switch n {
		case 0: // 1 in langur
			return left.real, nil
		case 1: // 2 in langur
			return left.imaginary, nil
		default:
			return left, fmt.Errorf("Complex index out of range")
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

func (left *Complex) IndexValid(index Object) bool {
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

func (left *Complex) IndexNativeInt(index Object) (idx int, ok bool) {
	idx, ok = NumberToInt(index)
	if !ok {
		return
	}
	return left.indexNativeInt(idx)
}

// receives 1-based integer index (langur)
// returns 0-based integer index (Go)
// ok true/false depending on validity
func (left *Complex) indexNativeInt(index int) (idx int, ok bool) {
	idx = index

	switch idx {
	case 1, 2:
		// good
		// convert 1-based index to 0-based (native) and return
		return idx - 1, true
	default:
		// bad
		// NOTE: intentionally not doing negative indices on complex
		return 0, false
	}
}
