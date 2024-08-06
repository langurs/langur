// langur/compiler/strings.go

package compiler

import (
	"fmt"
	"langur/ast"
	"langur/object"
	"langur/opcode"
	"langur/regex"
)

func (c *Compiler) compileDateTimeNode(node *ast.DateTimeNode) (ins opcode.Instructions, err error) {
	patternNode, ok := node.Pattern.(*ast.StringNode)
	if !ok {
		return nil, c.makeErr(node, fmt.Sprintf("Expected String Node within DateTime Node"))
	}
	s := patternNode.Values[0]

	if len(patternNode.Interpolations) == 0 {
		if !object.IsValidDateTimeString(s, true) {
			err = c.makeErr(node, "Invalid date-time literal string")
			return
		}

		dt, ok := node.Evaluate()
		if ok {
			ins = c.constantIns(dt)
			return
		}
	}

	// built at run-time (either contains interpolations or is a "now" date-time)
	ins, err = c.compileString(patternNode, regex.NONE)
	if err != nil {
		return
	}
	ins = append(ins, opcode.Make(opcode.OpDateTime)...)

	return
}

func (c *Compiler) compileDurationNode(node *ast.DurationNode) (ins opcode.Instructions, err error) {
	dur, ok := node.Evaluate()
	if ok {
		ins = c.constantIns(dur)
		return
	}

	patternNode, ok := node.Pattern.(*ast.StringNode)
	if !ok {
		return nil, c.makeErr(node, fmt.Sprintf("Expected String Node within Duration Node"))
	}

	// built at run-time (contains interpolations)
	ins, err = c.compileString(patternNode, regex.NONE)
	if err != nil {
		return
	}
	ins = append(ins, opcode.Make(opcode.OpDuration)...)

	return
}

func (c *Compiler) compileRegexNode(node *ast.RegexNode) (ins opcode.Instructions, err error) {
	patternNode, ok := node.Pattern.(*ast.StringNode)
	if !ok {
		return nil, c.makeErr(node, fmt.Sprintf("Expected String Node within Regex Node"))
	}

	var code int
	if node.RegexType == regex.RE2 {
		code = opcode.OC_Regex_Re2

	} else {
		bug("compileRegexNode", "Unknown regex type")
		return nil, c.makeErr(node, fmt.Sprintf("Unknown regex type"))
	}

	if len(patternNode.Interpolations) == 0 {
		// optimize by compiling a regex pattern now, rather than having the VM compile it
		var re object.Object

		re, err = object.NewRegex(patternNode.Values[0], node.RegexType)
		if err != nil {
			return
		}
		ins = c.constantIns(re)

	} else {
		ins, err = c.compileString(patternNode, node.RegexType)
		if err != nil {
			return
		}
		ins = append(ins, opcode.Make(opcode.OpRegex, code)...)
	}

	return
}

func (c *Compiler) compileStringNode(node *ast.StringNode) (ins opcode.Instructions, err error) {
	return c.compileString(node, regex.NONE)
}

func (c *Compiler) compileString(
	node *ast.StringNode, regexType regex.RegexType) (
	ins opcode.Instructions, err error) {

	if len(node.Interpolations) != len(node.Values)-1 {
		bug("compileString", "string value/interpolation node mismatch")
	}

	if len(node.Values) == 1 {
		// plain string (no interpolation)
		str := object.NewString(node.Values[0])
		ins = c.constantIns(str)

	} else {
		// interpolation
		count := 0
		for i := range node.Values {
			// add string constant
			if node.Values[i] != "" {
				str := object.NewString(node.Values[i])
				ins = append(ins, c.constantIns(str)...)
				count++
			}

			if i < len(node.Values)-1 {
				// not the last string section; add interpolation value
				interp, ok := node.Interpolations[i].(*ast.InterpolatedNode)
				if !ok {
					bug("compileStringNode", fmt.Sprintf("Expected interpolation node for value %d", i))
					err = c.makeErr(interp, fmt.Sprintf("Expected interpolation node for value %d", i))
					return
				}

				if regexType != regex.NONE {
					// interpolating regex into regex?
					// check that regex types match
					re, ok := interp.Value.(*ast.RegexNode)
					if ok && re.RegexType != regexType {
						err = c.makeErr(interp, fmt.Sprintf("Interpolated regex type value (%s) does not match regex literal type (%s)", re.RegexType.String(), regexType.String()))
						return
					}
				}

				interpolation, err := c.compileNode(interp.Value)
				if err != nil {
					return ins, err
				}
				ins = append(ins, interpolation...)
				count++

				mods, err := c.compileInterpolationModifiers(node, interp.Modifiers, regexType)
				if err != nil {
					return ins, err
				}
				ins = append(ins, mods...)
			}
		}
		ins = append(ins, opcode.Make(opcode.OpString, count)...)
	}

	return
}
