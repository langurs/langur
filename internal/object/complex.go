// langur/object/complex.go

package object

import (
	"strings"
	"langur/common"
)

func NewComplex(real, imaginary *Number) *Complex {
	return &Complex{
		real: real,
		imaginary: imaginary,
	}
}

type Complex struct {
	real *Number
	imaginary *Number
}

func (c *Complex) Type() ObjectType {
	return COMPLEX_OBJ
}
func (c *Complex) TypeString() string {
	return common.ComplexTypeName
}

func (c *Complex) Copy() Object {
	return &Complex{
		real: c.real.Copy().(*Number), 
		imaginary: c.imaginary.Copy().(*Number),
	}
}

func (l *Complex) Equal(n2 Object) bool {
	switch right := n2.(type) {
	case *Complex:
		return l.real.Equal(right.real) && l.imaginary.Equal(right.imaginary)
	}
	return false
}

func (l *Complex) Same(n2 Object) bool {
	switch right := n2.(type) {
	case *Complex:
		return l.real.Same(right.real) && l.imaginary.Same(right.imaginary)
	}
	return false
}

func (c *Complex) IsTruthy() bool {
	return c.real.IsTruthy() && c.imaginary.IsTruthy()
}

func (c *Complex) String() string {
	var sb strings.Builder
	
	sb.WriteString(c.real.String())
	
	if !c.imaginary.IsNegative() {
		sb.WriteRune('+')
	}
	
	sb.WriteString(c.imaginary.String())
	sb.WriteRune('i')
	
	return sb.String()
}

func (c *Complex) ReplString() string {
	return common.ComplexTypeName + " " + c.String()
}
