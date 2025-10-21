// langur/object/complex_misc.go

package object

func (c *Complex) Simplify() Object {
	return NewComplex(c.real.Simplify(), c.imaginary.Simplify())
}
