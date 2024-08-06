// langur/compiler/functions.go

package compiler

import (
	"fmt"
	"langur/ast"
	"langur/object"
	"langur/opcode"
	"langur/str"
	"langur/symbol"
	"langur/token"
)

func decodeInt(node ast.Node) (int, error) {
	switch n := node.(type) {
	case *ast.NumberNode:
		return n.DecodeInt()
	default:
		return 0, fmt.Errorf("Expected integers only")
	}
}

func (c *Compiler) compileFunctionNode(node *ast.FunctionNode) (ins opcode.Instructions, err error) {
	if node.ReturnType != nil {
		return nil, c.makeErr(node, "This revision of langur not able to compile explicit return type")
	}

	c.pushVariableScope() // pop scope in deferred function below
	c.symbolTable.IsFunction = true
	c.functionLevel++

	sig := &object.Signature{Name: node.Name}

	sig.ImpureEffects = node.ImpureEffects // self-declared impure; to be tested later...
	defer func() {
		c.popVariableScope()
		c.functionLevel--

		// Impurity is transitive.
		if sig.ImpureEffects {
			c.addToImpureEffectsList(node.Name)
		}
	}()

	var body opcode.Instructions

	switch sig.Name {
	case "_main":
		if len(node.PositionalParameters) != 0 || len(node.ByNameParameters) != 0 {
			err = c.makeErr(node, "Function _main() cannot have parameters")
			return
		}

	case "":
		// no name ... no defining self

	default:
		c.symbolTable.DefineSelf(sig.Name)
	}

	// compile parameters before function body so that each is added to the symbol table
	var defaultInsTotal opcode.Instructions
	var defaultCount int
	defaultInsTotal, defaultCount, err = c.compileFunctionNodeParameters(node, sig)
	if err != nil {
		return
	}

	if node.Body != nil {
		body, err = c.compileNode(node.Body)
		if err != nil {
			return
		}
	}

	if len(body) == 0 {
		// no body; return no value
		body = append(c.noValueIns, opcode.Make(opcode.OpReturnValue)...)

	} else if !ast.EndsWithDefiniteJump(node.Body.(*ast.BlockNode).Statements) {
		// append return if doesn't already end with return
		body = append(body, opcode.Make(opcode.OpReturnValue)...)
	}

	freeSymbols := c.symbolTable.FreeSymbols
	localsCount := c.symbolTable.DefinitionCount

	// may be self-declared or proven impure
	sig.ImpureEffects = sig.ImpureEffects || c.symbolTable.ImpureEffects != nil

	if sig.ImpureEffects && !node.ImpureEffects {
		if node.Name == "" {
			err = c.makeErr(node, "Anonymous impure function not declared as impure; use a * to declare impurity, such as fn*() { }")
		} else {
			err = c.makeErr(node, fmt.Sprintf("Impure function (%s) not declared as impure; use a * to declare impurity, such as fn*() { }", str.ReformatInput(node.Name)))
		}
		return
	}

	compiledFn := &object.CompiledCode{
		FnSignature:        sig,
		InsPackage:         opcode.InsPackage{Instructions: body},
		LocalBindingsCount: localsCount,
	}
	fnIndex := c.addConstant(compiledFn)

	ins = nil
	if len(freeSymbols) != 0 || defaultCount != 0 {
		// a closure or has optional parameter defaults that are to be determined at run-time
		ins, err = c.instructionsForSymbols(node, freeSymbols)
		if err != nil {
			return
		}
		ins = append(ins, defaultInsTotal...)
		ins = append(ins, opcode.Make(opcode.OpFunction, fnIndex, len(freeSymbols), defaultCount)...)

	} else {
		// not a closure and has all optional parameter defaults already determined
		ins = opcode.Make(opcode.OpConstant, fnIndex)
	}

	return
}

func (c *Compiler) instructionsForSymbols(node ast.Node, symbols []symbol.Symbol) (ins opcode.Instructions, err error) {
	var temp opcode.Instructions
	for _, sym := range symbols {
		// add opcodes to push "free" values onto the stack so they can be picked up when the VM hits OpFunction
		temp, err = c.makeOpGetInstructions(node, sym, 0)
		if err != nil {
			return
		}
		ins = append(ins, temp...)
	}
	return
}

func (c *Compiler) compileFunctionNodeParameters(
	node *ast.FunctionNode, sig *object.Signature) (
	defaultInsTotal opcode.Instructions, defaultCount int, err error) {

	// to set optional parameter defaults that are determined at run-time ...
	// ... that may include variables (not closure "free" variables)
	previousFreezeDefineFree := c.symbolTable.FreezeDefineFree
	c.symbolTable.FreezeDefineFree = true
	defer func() {
		c.symbolTable.FreezeDefineFree = previousFreezeDefineFree
	}()

	var param object.Parameter

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

	// External names are not registered in a symbol table.
	// check for duplicates to prevent confusion and chaos
	var externalNames []string

	for i, p := range node.ByNameParameters {
		var defaultIns opcode.Instructions
		param, defaultIns, _, _, err = c.compileParameter(p, i+1, false)
		if err != nil {
			return
		}

		sig.ParamByName = append(sig.ParamByName, param)

		if len(defaultIns) != 0 {
			name := c.constantIns(object.NewString(param.ExternalName))
			defaultInsTotal = append(defaultInsTotal, name...)
			defaultInsTotal = append(defaultInsTotal, defaultIns...)
			defaultCount++
		}

		// check for duplicate external names (not registered in symbol tables)
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

func (c *Compiler) compileParameter(node ast.Node, pnum int, lastPositional bool) (
	param object.Parameter, defaultIns opcode.Instructions,
	paramExpansionMin, paramExpansionMax int, err error) {

	system := false
	param.Mutable = false

	switch p := node.(type) {
	case *ast.IdentNode:
		param.InternalName = p.Name
		system = p.System

	case *ast.LineDeclarationNode:
		// to use var token to make parameter mutable
		switch assign := p.Assignment.(type) {
		case *ast.IdentNode:
			param.InternalName = assign.Name
			system = assign.System

		case *ast.AssignmentNode:
			param, defaultIns, err = c.assessParameterByName(assign)
			if err != nil {
				err = c.makeErr(node, err.Error())
				return
			}
			system = assign.SystemAssignment

		default:
			err = c.makeErr(node, fmt.Sprintf("Parameter %d invalid", pnum))
			return
		}
		param.Mutable = p.Mutable

	case *ast.AssignmentNode:
		param, defaultIns, err = c.assessParameterByName(p)
		if err != nil {
			err = c.makeErr(node, err.Error())
			return
		}
		system = p.SystemAssignment

	case *ast.ExpansionNode:
		if !lastPositional {
			err = c.makeErr(node, "Parameter expansion only allowed on last positional parameter")
			return
		}

		switch lim := p.Limits.(type) {
		case nil:
			// no limits given
			paramExpansionMax = -1
			paramExpansionMin = 0

		case *ast.NumberNode:
			paramExpansionMax, err = decodeInt(lim)
			paramExpansionMin = paramExpansionMax

		case *ast.InfixExpressionNode:
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
				err = c.makeErr(node, "Invalid expression for limits on parameter expansion")
			}

		default:
			err = c.makeErr(node, fmt.Sprintf("Invalid limit type on parameter expansion (%T)", lim))
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
		case *ast.IdentNode:
			param.InternalName = continuation.Name
			system = continuation.System
		default:
			err = c.makeErr(node, "Invalid parameter expansion; expected variable name only")
			return
		}

	default:
		err = c.makeErr(node, fmt.Sprintf("Parameter %d invalid", pnum))
		return
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

func (c *Compiler) assessParameterByName(assign *ast.AssignmentNode) (
	param object.Parameter, defaultIns opcode.Instructions, err error) {

	// optional by name or required by name?
	requiredByName := assign.Values == nil

	if len(assign.Identifiers) != 1 ||
		(!requiredByName && len(assign.Values) != 1) {

		err = c.makeErr(assign, "Expected 1 identifier and 1 value for parameter by name assignment")
		return
	}

	param = object.Parameter{}

	switch expr := assign.Identifiers[0].(type) {
	case *ast.IdentNode:
		param.InternalName, param.ExternalName = expr.Name, expr.Name

	case *ast.InfixExpressionNode:
		if expr.Operator.Type == token.AS {
			param.InternalName = expr.Left.(*ast.IdentNode).Name
			param.ExternalName = expr.Right.(*ast.IdentNode).Name
		} else {
			err = c.makeErr(assign, "Expected identifier or identifier/alias for parameter by name")
		}

	default:
		err = c.makeErr(assign, "Expected identifier or identifier/alias for parameter by name")
		return
	}

	if !requiredByName {
		// attempt to build default value now (if possible)
		defaultIns, param.DefaultValue, err = c.compileOrEvaluateNode(assign.Values[0])
		if err != nil {
			err = c.makeErr(assign, fmt.Sprintf("Failure to compile default value for optional parameter %s: %s", str.ReformatInput(param.InternalName), err.Error()))
			return
		}
		if param.DefaultValue == nil {
			// default failed to evaluate at compile-time
			// set to no value for now to indicate an optional parameter, not a "required by name" parameter
			// instructions to be evaluated at run-time
			param.DefaultValue = object.NONE
		}
	}

	return
}

func (c *Compiler) compileSelfRef(node ast.Node) (opcode.Instructions, error) {
	if c.symbolTable.Outer == nil {
		return nil, c.makeErr(node, "Cannot use self token in global scope")
	}
	return opcode.MakeWithErrTest(opcode.OpGetSelf)
}

func (c *Compiler) compileReturnNode(node *ast.ReturnNode) (ins opcode.Instructions, err error) {
	if c.functionLevel == 0 {
		err = c.makeErr(node, "Cannot use return outside of function")
		return
	}
	ins, err = c.compileNode(node.ReturnValue)
	if err != nil {
		return
	}
	ins = append(ins, opcode.Make(opcode.OpReturnValue)...)
	return
}

func (c *Compiler) compileCallNode(node *ast.CallNode) (ins opcode.Instructions, err error) {
	hasExpansion := false

	// Compiling the function first ...
	// ... but we add it to the instructions after the arguments.
	var fn opcode.Instructions
	fn, err = c.compileNode(node.Function)
	if err != nil {
		return
	}

	var bslc []byte

	for _, arg := range node.PositionalArgs {
		if hasExpansion {
			// already set hasExpansion and have another positional argument
			err = c.makeErr(arg, fmt.Sprintf("Argument expansion only possible on last positional argument"))
			return
		}

		switch post := arg.(type) {
		case *ast.PostfixExpressionNode:
			if post.Operator.Type == token.EXPANSION {
				arg = post.Left
				hasExpansion = true
			}
		}

		bslc, err = c.compileNode(arg)
		if err != nil {
			return
		}

		ins = append(ins, bslc...)
	}

	var externalNames []string

	for _, arg := range node.ByNameArgs {
		externalName := ""

		if assign, ok := arg.(*ast.AssignmentNode); ok {
			externalName = assign.Identifiers[0].TokenRepresentation()
			name := &ast.StringNode{
				Token: assign.Token, Values: []string{externalName}}
			bslc, err = c.compileNode(name)
			if err != nil {
				return
			}

			// compiling to name/value object (internally used for argument by name)
			var value opcode.Instructions
			value, err = c.compileNode(assign.Values[0])
			if err != nil {
				return
			}
			bslc = append(bslc, value...)
			bslc = append(bslc, opcode.Make(opcode.OpNameValue)...)

			ins = append(ins, bslc...)

			// check for duplicate external (argument) names
			if externalName != "" {
				if str.IsInSlice(externalName, externalNames) {
					err = c.makeErr(arg, fmt.Sprintf("Duplicate of argument by name (%s)", str.ReformatInput(externalName)))
					return
				}
				externalNames = append(externalNames, externalName)
			}

		} else {
			// not an assignment node
			err = c.makeErr(arg, fmt.Sprintf("Expected assignment node for argument by name (%s)", str.ReformatInput(externalName)))
			return
		}
	}

	// NOTE: putting function to call onto stack after arguments and will be popped first
	ins = append(ins, fn...)

	if hasExpansion {
		ins = append(ins, opcode.Make(opcode.OpCallWithExpansion, len(node.PositionalArgs), len(node.ByNameArgs))...)
	} else {
		ins = append(ins, opcode.Make(opcode.OpCall, len(node.PositionalArgs), len(node.ByNameArgs))...)
	}
	return
}
