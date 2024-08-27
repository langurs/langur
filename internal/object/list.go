// langur/object/list.go

package object

import (
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
