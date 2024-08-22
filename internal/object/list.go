// langur/object/list.go

package object

import (
	"fmt"
	"langur/common"
	"strings"
)

type List struct {
	Elements []Object
}

var EmptyList = &List{}

func (left *List) HasImpureEffects() bool {
	return SliceHasImpureEffects(left.Elements...)
}

func (left *List) Copy() Object {
	return &List{Elements: CopySlice(left.Elements)}
}

func (l *List) Equal(left2 Object) bool {
	r, ok := left2.(*List)
	if !ok {
		return false
	}
	if len(l.Elements) != len(r.Elements) {
		return false
	}
	for i := range l.Elements {
		if !Equal(l.Elements[i], r.Elements[i]) {
			return false
		}
	}
	return true
}

func (left *List) CopyRefs() Object {
	return &List{Elements: CopyRefSlice(left.Elements)}
}

func (left *List) Type() ObjectType {
	return LIST_OBJ
}
func (left *List) TypeString() string {
	return common.ListTypeName
}

func (left *List) IsTruthy() bool {
	return len(left.Elements) != 0
}

func (left *List) String() string {
	elements := []string{}

	for _, e := range left.Elements {
		elements = append(elements, ComposedOrRegularString(e))
	}

	return "[" + strings.Join(elements, ", ") + "]"
}

func (left *List) ReplString() string {
	elements := []string{}

	for _, e := range left.Elements {
		elements = append(elements, e.ReplString())
	}

	return common.ListTypeName + " [" + strings.Join(elements, ", ") + "]"
}

func ListFromInt64Slice(ints []int64) *List {
	arr := &List{Elements: make([]Object, len(ints))}
	for i := range ints {
		arr.Elements[i] = NumberFromInt64(ints[i])
	}
	return arr
}

func ListFromRuneSlice(rSlc []rune) *List {
	arr := &List{Elements: make([]Object, len(rSlc))}
	for i := range rSlc {
		arr.Elements[i] = NumberFromInt(int(rSlc[i]))
	}
	return arr
}

// builds new list without keys we want to remove
func (left *List) RemoveIndices(indices Object) (*List, error) {
	intIdx, err := makeNativeIntIndexMap(left, indices)
	if err != nil {
		return nil, err
	}

	elements := []Object{}
	for i := range left.Elements {
		if !intIdx[i] {
			elements = append(elements, left.Elements[i])
		}
	}

	return &List{Elements: elements}, nil
}

func makeNativeIntIndexMap(obj, index Object) (indexmap map[int]bool, err error) {
	indexmap = make(map[int]bool)

	resolve := func(n int) (int, bool) {
		iii, ok := obj.(IIndexNativeInt)
		if ok {
			n, ok = iii.indexNativeInt(n)
			if ok {
				// resolved and valid index; add to map
				indexmap[n] = true
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
			resolve(n)

		case *Range:
			start, ok := NumberToInt(idx.Start)
			if !ok {
				return fmt.Errorf("Start of range not an integer")
			}
			start, ok = resolve(start)
			if !ok {
				return fmt.Errorf("Start of range not resolvable")
			}
			end, ok := NumberToInt(idx.End)
			if !ok {
				return fmt.Errorf("End of range not an integer")
			}
			end, ok = resolve(end)
			if !ok {
				return fmt.Errorf("End of range not resolvable")
			}
			if end < start {
				start, end = end, start
			}
			for n := start + 1; n < end+1; n++ {
				resolve(n)
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
