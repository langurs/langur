// langur/ast/compiler_datatypes.go

package ast

import (
	"fmt"
	"langur/object"
	"langur/opcode"
	"langur/token"
	"langur/regex"
)

func (c *Compiler) compileNumberObject(node *NumberNode) (number *object.Number, err error) {
	if c.Modes.WarnOnIntegerLiteralsStartingWithZero {
		if node.Token.Type == token.INT && len(node.Value) > 1 &&
			node.Token.Code2 == token.CODE_DEFAULT && node.Value[0] == '0' {
			err = c.makeWarning(node, "Integer literal starting with zero (might be confused for a base 8 number (as used in other languages)")
			return
		}
	}

	if node.Imaginary {
		err = c.makeErr(node, "Misplaced imaginary number (not part of a complex number)")
		return
	}

	number, err = object.NumberFromStringBase(node.Value, node.Base)
	return
}

func (c *Compiler) compileComplexNumber(
	real Node, imaginary Node, conjugate bool) (pkg opcode.InsPackage, err error) {

	var r, i *object.Number

	switch real := real.(type) {
	case nil:
		r = object.Zero
		
	case *NumberNode:
		if real.Imaginary {
			err = c.makeErr(real, "Expected real number for real part of complex number")
			return
		}
		r, err = c.compileNumberObject(real)
		
	default:
		err = c.makeErr(real, "Expected number for real part of complex number")
	}
	if err != nil {
		return
	}
	
	switch imaginary := imaginary.(type) {
	case nil:
		i = object.Zero

	case *NumberNode:
		if !imaginary.Imaginary {
			err = c.makeErr(real, "Expected imaginary number for imaginary part of complex number")
			return
		}
		imaginary.Imaginary = false // remove flag here so it doesn't throw an error
		i, err = c.compileNumberObject(imaginary)
		
	default:
		err = c.makeErr(real, "Expected number for imaginary part of complex number")
	}

	if conjugate {
		i = i.Negate().(*object.Number)
	}

	pkg = c.constantIns(object.NewComplex(r, i))
	return
}

// check for complex number such as 1 + 1i
func (c *Compiler) checkForComplexNumber(node *InfixExpressionNode, op opcode.OpCode) (pkg opcode.InsPackage, err error) {
	numeric := func(node Node) int {
		switch node := node.(type) {
		case *NumberNode:
			if node.Imaginary {
				return 2
			}
			return 1
		}
		return 0
	}
	if (op == opcode.OpAdd || op == opcode.OpSubtract) && 
		numeric(node.Left) == 1 &&
		numeric(node.Right) == 2 {
		
		pkg, err = c.compileComplexNumber(node.Left, node.Right, op == opcode.OpSubtract)
	}
	return
}

func (c *Compiler) compileString(
	node *StringNode, regexType regex.RegexType) (
	pkg opcode.InsPackage, err error) {

	if len(node.Interpolations) != len(node.Values)-1 {
		err = c.makeErr(node, "string value/interpolation node mismatch")
		bug("compileString", err.Error())
		return
	}

	if len(node.Values) == 1 {
		// plain string (no interpolation)
		str := object.NewString(node.Values[0])
		pkg = c.constantIns(str)

	} else {
		// interpolation
		count := 0
		for i := range node.Values {
			// add string constant
			if node.Values[i] != "" {
				str := object.NewString(node.Values[i])
				pkg = pkg.Append(c.constantIns(str))
				count++
			}

			if i < len(node.Values)-1 {
				// not the last string section; add interpolation value
				interp, ok := node.Interpolations[i].(*InterpolatedNode)
				if !ok {
					err = c.makeErr(interp, fmt.Sprintf("Expected interpolation node for value %d", i))
					bug("compileStringNode", err.Error())
					return
				}

				if regexType != regex.NONE {
					// interpolating regex into regex?
					// check that regex types match
					re, ok := interp.Value.(*RegexNode)
					if ok && re.RegexType != regexType {
						err = c.makeErr(interp, fmt.Sprintf("Interpolated regex type value (%s) does not match regex literal type (%s)", re.RegexType.String(), regexType.String()))
						return
					}
				}

				interpolation, err := interp.Value.Compile(c)
				if err != nil {
					return pkg, err
				}
				pkg = pkg.Append(interpolation)
				count++

				mods, err := c.compileInterpolationModifiers(node, interp.Modifiers, regexType)
				if err != nil {
					return pkg, err
				}
				pkg = pkg.Append(mods)
			}
		}
		pkg = pkg.Append(opcode.MakePkg(node.Token, opcode.OpString, count))
	}

	return
}
