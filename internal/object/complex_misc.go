// langur/object/complex_misc.go

package object

func (c *Complex) Simplify() Object {
	// if no imaginary number, return real number only
	if c.imaginary.IsZero() {
		return c.real
	}
	// if contains imaginary number, return complex number
	return c
}
