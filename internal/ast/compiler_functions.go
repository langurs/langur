// langur/ast/compiler_functions.go

package ast

import (
	"fmt"
	"langur/common"
	"langur/opcode"
	"langur/object"
	"langur/str"
	"langur/token"
	"langur/symbol"
)

func decodeInt(node Node) (int, error) {
	switch n := node.(type) {
	case *NumberNode:
		return n.DecodeInt()
	default:
		return 0, fmt.Errorf("Expected integers only")
	}
}

func (c *Compiler) compileFunctionNodeParameters(
	node *FunctionNode, sig *object.Signature) (
	defaultInsTotal opcode.InsPackage, defaultCount int, err error) {

	// to set optional parameter defaults that are determined at run-time ...
	// ... that may include variables (not closure "free" variables)
	previousFreezeDefineFree := c.symbolTable.FreezeDefineFree
	c.symbolTable.FreezeDefineFree = true
	defer func() {
		c.symbolTable.FreezeDefineFree = previousFreezeDefineFree
	}()

	var param object.Parameter

	// POSITIONAL PARAMETERS
	for i, p := range node.PositionalParameters {
		lastPositional := i == len(node.PositionalParameters)-1

		var expansionMin, expansionMax int
		param, _, expansionMin, expansionMax, err =
			c.compileParameter(p, i+1, lastPositional)

		if err != nil {
			return
		}

		if lastPositional {
			sig.ParamExpansionMin = expansionMin
			sig.ParamExpansionMax = expansionMax
		}

		sig.ParamPositional = append(sig.ParamPositional, param)
	}

	// CHECK MAX COUNTS
	cnt := len(node.PositionalParameters) - len(node.ByNameParameters)
	maxExpansionMax := common.ArgCountMax - cnt + 1

	if sig.ParamExpansionMax == -1 {
		sig.ParamExpansionMax = maxExpansionMax
	}

	if sig.ParamExpansionMax < 0 ||
		sig.ParamExpansionMax > maxExpansionMax ||
		cnt > common.ArgCountMax {

		err = c.makeErr(node, fmt.Sprintf("Max parameter/argument count (%d) exceeded", common.ArgCountMax))
		return
	}

	// PARAMETERS BY NAME
	var externalNames []string

	for i, p := range node.ByNameParameters {
		var defaultIns opcode.InsPackage
		param, defaultIns, _, _, err = c.compileParameter(p, i+1, false)
		if err != nil {
			return
		}

		sig.ParamByName = append(sig.ParamByName, param)

		if len(defaultIns.Instructions) != 0 {
			name := c.constantIns(object.NewString(param.ExternalName))
			defaultInsTotal = defaultInsTotal.Append(name)
			defaultInsTotal = defaultInsTotal.Append(defaultIns)
			defaultCount++
		}

		// External names are not registered in a symbol table.
		// Therefore, we check for duplicates to prevent confusion and chaos.
		if param.ExternalName != "" {
			if str.IsInSlice(param.ExternalName, externalNames) {
				err = c.makeErr(node, fmt.Sprintf("Duplicate external name declared (%s) for parameters by name", str.ReformatInput(param.ExternalName)))
				return
			}
			externalNames = append(externalNames, param.ExternalName)
		}
	}

	return
}

func (c *Compiler) compileParameter(node Node, pnum int, lastPositional bool) (
	param object.Parameter, defaultIns opcode.InsPackage,
	paramExpansionMin, paramExpansionMax int, err error) {

	system := false
	param.Mutable = false

	switch p := node.(type) {
	case *IdentNode:
		param.InternalName = p.Name
		system = p.System

	case *LineDeclarationNode:
		// to use var token to make parameter mutable
		switch assign := p.Assignment.(type) {
		case *IdentNode:
			param.InternalName = assign.Name
			system = assign.System

		case *AssignmentNode:
			param, defaultIns, err = c.assessParameterByName(assign)
			if err != nil {
				err = c.makeErr(node, err.Error())
				return
			}
			system = assign.SystemAssignment

		default:
			err = c.makeErr(p, fmt.Sprintf("Parameter %d invalid", pnum))
			return
		}
		param.Mutable = p.Mutable

	case *InfixExpressionNode:
		// required parameter by name
		param = object.Parameter{Required: true}

		if p.Operator.Type == token.AS {
			param.InternalName = p.Left.(*IdentNode).Name
			param.ExternalName = p.Right.(*IdentNode).Name
		} else {
			err = c.makeErr(p, "Expected identifier or identifier/alias for parameter by name")
		}

	case *AssignmentNode:
		param, defaultIns, err = c.assessParameterByName(p)
		if err != nil {
			err = c.makeErr(node, err.Error())
			return
		}
		system = p.SystemAssignment

	case *ExpansionNode:
		if !lastPositional {
			err = c.makeErr(p, "Parameter expansion only allowed on last positional parameter")
			return
		}

		switch lim := p.Limits.(type) {
		case nil:
			// no limits given
			paramExpansionMax = -1
			paramExpansionMin = 0

		case *NumberNode:
			paramExpansionMax, err = decodeInt(lim)
			paramExpansionMin = paramExpansionMax

		case *InfixExpressionNode:
			switch lim.Operator.Type {
			case token.RANGE:
				paramExpansionMin, err = decodeInt(lim.Left)
				if err != nil {
					return
				}
				if lim.Right == nil {
					// ...[0..] x
					paramExpansionMax = -1
				} else {
					paramExpansionMax, err = decodeInt(lim.Right)
				}

			default:
				err = c.makeErr(p, "Invalid expression for limits on parameter expansion")
			}

		default:
			err = c.makeErr(p, fmt.Sprintf("Invalid limit type on parameter expansion (%T)", lim))
		}
		if err == nil &&
			(paramExpansionMin < 0 || paramExpansionMax < -1 || paramExpansionMax == 0 ||
				(paramExpansionMin > paramExpansionMax && paramExpansionMax != -1)) {

			err = c.makeErr(node, "Invalid limits on parameter expansion")
		}
		if err != nil {
			return
		}

		switch continuation := p.Continuation.(type) {
		case *IdentNode:
			param.InternalName = continuation.Name
			system = continuation.System
		default:
			err = c.makeErr(p, "Invalid parameter expansion; expected variable name only")
			return
		}

	default:
		err = c.makeErr(p, fmt.Sprintf("Parameter %d invalid", pnum))
		return
	}

	// ADD EXPLICIT TYPE TO PARAMETER
	// NOTE: This version of langur cannot compile explicit type with parameter expansion.
	if _, exp := node.(*ExpansionNode); !exp {
		err = addParameterType(&param, node)
		if err != nil {
			err = c.makeErr(node, err.Error())
			return
		}
	}
	
	if param.Type != 0 {
		if defaultIns.Instructions != nil {
			err = c.makeErr(node, "Cannot compile explicit parameter type with default value not known at compile time")
			return
		}
		if defaultIns.Instructions == nil && param.DefaultValue != nil {
			if param.DefaultValue.Type() != param.Type {
				err = c.makeErr(node, fmt.Sprintf("Parameter type does not match default value type"))
				return
			}
		}
		if param.Mutable {
			err = c.makeErr(node, "This version of langur cannot use a mutable parameter with an explicit type")
			return
		}
	}

	// DEFINE IN SYMBOL TABLE
	// An external name (for an optional parameter) may shadow a keyword ...
	// since the context makes the meaning clear, ...
	// but an internal name (used within a compiled function) may not.

	_, err = c.symbolTable.DefineVariable(param.InternalName, param.Mutable, system)
	if err != nil {
		err = c.makeErr(node, fmt.Sprintf("Parameter %d definition error: %s", pnum, err.Error()))
	}

	return
}

func addParameterType(param *object.Parameter, pnode Node) error {
	var ok bool
	for {
		switch p := pnode.(type) {
		case nil:
			return nil
			
		case *IdentNode:
			if p.Type == nil {
				// no type specified
				return nil
			}

			tname := p.Type.TokenRepresentation()
			param.Type, ok = object.TypeNameToType[tname]
			if !ok {
				return fmt.Errorf("Cannot compile node %q as a langur type", tname)
			}
			return nil
			
		case *InfixExpressionNode:
			// alias with *as* keyword
			pnode = p.Left
			
		case *LineDeclarationNode:
			pnode = p.Assignment
			
		case *AssignmentNode:
			pnode = p.Identifiers[0]
			
		case *ExpansionNode:
			// pnode = p.Continuation
			
			return fmt.Errorf("This version of langur cannot compile explicit type with parameter expansion")
			
		default:
			return fmt.Errorf("Failed to compile type for parameter")
		}
	}
}

func (c *Compiler) assessParameterByName(assign *AssignmentNode) (
	param object.Parameter, defaultIns opcode.InsPackage, err error) {

	if len(assign.Identifiers) != 1 {
		err = c.makeErr(assign, "Expected 1 identifier and 1 value for parameter by name assignment")
		return
	}

	param = object.Parameter{}

	switch expr := assign.Identifiers[0].(type) {
	case *IdentNode:
		param.InternalName, param.ExternalName = expr.Name, expr.Name

	case *InfixExpressionNode:
		if expr.Operator.Type == token.AS {
			param.InternalName = expr.Left.(*IdentNode).Name
			param.ExternalName = expr.Right.(*IdentNode).Name
		} else {
			err = c.makeErr(assign, "Expected identifier or identifier/alias for parameter by name")
			return
		}

	default:
		err = c.makeErr(assign, "Expected identifier or identifier/alias for parameter by name")
		return
	}

	// attempt to build default value now (if possible)
	param.DefaultValue = assign.Values[0].Evaluate()
	if param.DefaultValue == nil {
		defaultIns, err = assign.Values[0].Compile(c)
		if err != nil {
			err = c.makeErr(assign, fmt.Sprintf("Failure to compile default value for optional parameter %s: %s", str.ReformatInput(param.InternalName), err.Error()))
			return
		}
	}
	if param.DefaultValue == nil {
		// default did not evaluate at compile-time
		// set to no value for now to indicate an optional parameter (not a "required by name" parameter)
		// instructions to be evaluated at run-time
		param.DefaultValue = object.NONE
	}

	return
}

func (c *Compiler) instructionsForSymbols(node Node, symbols []symbol.Symbol) (pkg opcode.InsPackage, err error) {
	var temp opcode.InsPackage
	for _, sym := range symbols {
		// add opcodes to push "free" values onto the stack so they can be picked up when the VM hits OpFunction
		temp, err = c.makeOpGetInstructions(node, sym, 0)
		if err != nil {
			return
		}
		pkg = pkg.Append(temp)
	}
	return
}
