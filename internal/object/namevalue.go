// langur/object/namevalue.go

package object

import (
	"fmt"
	"langur/common"
	"strings"
)

// INTERNAL ONLY; for optional argument
type NameValue struct {
	Name  string
	Value Object
}

func (nv *NameValue) HasImpureEffects() bool {
	d, ok := nv.Value.(IDefilableEffects)
	if ok {
		return d.HasImpureEffects()
	}
	return false
}

func NewNameValue(name, value Object) (*NameValue, error) {
	var use string
	switch n := name.(type) {
	case *String:
		use = n.value
	default:
		return nil, fmt.Errorf("Expected string for name of name/value")
	}
	return &NameValue{Name: use, Value: value}, nil
}

func (nv *NameValue) Copy() Object {
	return &NameValue{
		Name:  nv.Name,
		Value: nv.Value.Copy(),
	}
}

func (nv *NameValue) Equal(o2 Object) bool {
	switch nv2 := o2.(type) {
	case *NameValue:
		return nv.Name == nv2.Name &&
			nv.Equal(nv2)
	}
	return false
}

func (nv *NameValue) Type() ObjectType {
	return NAME_VALUE
}
func (nv *NameValue) TypeString() string {
	return common.NameValueTypeName
}

func (nv *NameValue) IsTruthy() bool {
	return nv.Value.IsTruthy()
}

func (nv *NameValue) String() string {
	// langur string; should not happen
	return INTERNAL_OBJECT_ONLY
}

func (nv *NameValue) ReplString() string {
	var sb strings.Builder

	sb.WriteString("NameValue(")
	sb.WriteString(nv.Name)
	sb.WriteString(" = ")
	sb.WriteString(nv.Value.ReplString())
	sb.WriteRune(')')

	return sb.String()
}
