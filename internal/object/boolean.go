// langur/object/boolean.go
// Boolean and null

package object

import (
	"langur/common"
)

// CONSTANT OBJECTS

var (
	// to reference unchanging value Objects rather than creating new ones
	// also, using these allows comparing pointers for equality, that is value == TRUE
	TRUE  = &Boolean{Value: true}
	FALSE = &Boolean{Value: false}
	NULL  = &Null{}

	// return NONE when there is no value to return
	NONE = NULL
)

func NativeBoolToObject(input bool) *Boolean {
	if input {
		return TRUE
	}
	return FALSE
}

// BOOLEAN

type Boolean struct {
	Value bool
}

func (b *Boolean) Copy() Object {
	// don't make a new Boolean object
	// see notes above about Boolean and null
	return b
}

// fulfilling the Object interface; not necessarily to be called, as this will be tested natively
func (l *Boolean) Equal(b2 Object) bool {
	r, ok := b2.(*Boolean)
	if !ok {
		return false
	}
	return l == r
}

func (b *Boolean) String() string {
	if b.Value {
		return common.TrueTokenLiteral
	} else {
		return common.FalseTokenLiteral
	}
}

func (b *Boolean) ReplString() string {
	if b.Value {
		return common.BooleanTypeName + " " + common.TrueTokenLiteral
	} else {
		return common.BooleanTypeName + " " + common.FalseTokenLiteral
	}
}

func (b *Boolean) Type() ObjectType {
	return BOOLEAN_OBJ
}
func (b *Boolean) TypeString() string {
	return common.BooleanTypeName
}

func (b *Boolean) IsTruthy() bool {
	return b.Value
}

// NULL

type Null struct{}

func (n *Null) Copy() Object {
	// don't make a new Null object
	// see notes above about Boolean and null
	return n
}

// fulfilling the Object interface; not necessarily to be called, as this will be tested natively
func (l *Null) Equal(n2 Object) bool {
	_, ok := n2.(*Null)
	if !ok {
		return false
	}
	return true
}

func (n *Null) String() string {
	return common.NullTokenLiteral
}
func (n *Null) ReplString() string {
	return common.NullTypeName
}

func (n *Null) Type() ObjectType {
	return NULL_OBJ
}
func (n *Null) TypeString() string {
	return common.NullTypeName
}

func (n *Null) IsTruthy() bool {
	return false
}
