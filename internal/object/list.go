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

func (left *List) RemoveIndices(indices Object) (*List, error) {
	// build new list without keys we want to remove
	elements := []Object{}
	for i := range left.Elements {
		excludeThis, err := intInObject(i+1, indices)
		if err != nil {
			return left, err
		}
		if !excludeThis {
			elements = append(elements, left.Elements[i])
		}
	}

	return &List{Elements: elements}, nil
}

func intInObject(i int, obj Object) (bool, error) {
	switch o := obj.(type) {
	case *Number:
		n, ok := NumberToInt(o)
		if !ok {
			return false, fmt.Errorf("Number not an integer")
		}
		return i == n, nil

	case *Range:
		start, ok := NumberToInt(o.Start)
		if !ok {
			return false, fmt.Errorf("Start of range not an integer")
		}
		end, ok := NumberToInt(o.End)
		if !ok {
			return false, fmt.Errorf("End of range not an integer")
		}
		if start > end {
			return i >= end && i <= start, nil
		}
		return i >= start && i <= end, nil

	case *List:
		for _, e := range o.Elements {
			in, err := intInObject(i, e)
			if err != nil {
				return false, err
			}
			if in {
				return true, nil
			}
		}
		return false, nil

	default:
		return false, fmt.Errorf("Expected integer, range of integers, or list of such")
	}
}
