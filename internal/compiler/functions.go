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
		return 0, makeErr(node, "Expected integers only")
	}
}

func (c *Compiler) compileFunctionNode(node *ast.FunctionNode) (ins opcode.Instructions, err error) {
	if node.ReturnType != nil {
		return nil, makeErr(node, "This revision of langur not able to compile explicit return type")
	}

	c.pushVariableScope() // pop scope in deferred function below
	c.symbolTable.IsFunction = true
	c.functionLevel++

	sig := &object.Signature{Name: node.Name}

	sig.ImpureEffects = node.Impure // self-declared impure; to be tested later...
	defer func() {
		c.popVariableScope()
		c.functionLevel--

		// Impurity is transitive.
		if sig.ImpureEffects {
			c.addToImpuritiesList(node.Name)
		}
	}()

	var body opcode.Instructions

	switch sig.Name {
	case "_main":
		if len(node.Parameters) != 0 {
			err = makeErr(node, "Function _main() cannot have parameters")
			return
		}

	case "":
		// no name ... no defining self

	default:
		c.symbolTable.DefineSelf(sig.Name)
	}

	err = c.compileFunctionNodeParameters(node, sig)
	if err != nil {
		return
	}

	if node.Body != nil {
		body, err = c.compileNode(node.Body, false)
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
	sig.ImpureEffects = sig.ImpureEffects || c.symbolTable.Impurities != nil

	if sig.ImpureEffects && !node.Impure {
		if node.Name == "" {
			err = makeErr(node, "Anonymous impure function not declared as impure; use a * to declare impurity, such as fn*() { }")
		} else {
			err = makeErr(node, fmt.Sprintf("Impure function (%s) not declared as impure; use a * to declare impurity, such as fn*() { }", str.ReformatInput(node.Name)))
		}
		return
	}

	compiledFn := &object.CompiledCode{
		FnSignature:        sig,
		Instructions:       body,
		LocalBindingsCount: localsCount,
	}
	fnIndex := c.addConstant(compiledFn)

	ins = nil
	if len(freeSymbols) > 0 {
		// a closure
		ins, err = c.instructionsForSymbols(node, freeSymbols)
		if err != nil {
			return
		}
		ins = append(ins, opcode.Make(opcode.OpClosure, fnIndex, len(freeSymbols))...)

	} else {
		// not a closure
		ins = opcode.Make(opcode.OpConstant, fnIndex)
	}

	return
}

func (c *Compiler) instructionsForSymbols(node ast.Node, symbols []symbol.Symbol) (ins opcode.Instructions, err error) {
	var temp opcode.Instructions
	for _, sym := range symbols {
		// add opcodes to push "free" values onto the stack so they can be picked up when the VM hits OpClosure
		temp, err = c.makeOpGetInstructions(node, sym, 0)
		if err != nil {
			return
		}
		ins = append(ins, temp...)
	}
	return
}

func (c *Compiler) compileFunctionNodeParameters(node *ast.FunctionNode, sig *object.Signature) (err error) {
	var params []object.Parameter

	isOptionalParameterNode := func(node ast.Node) bool {
		switch p := node.(type) {
		case *ast.LineDeclarationNode:
			switch p.Assignment.(type) {
			case *ast.AssignmentNode:
				return true
			}

		case *ast.AssignmentNode:
			return true
		}
		return false
	}

	opt := false
	for i, p := range node.Parameters {
		var param object.Parameter

		lastPositional := false
		if len(node.Parameters) == 1 && !isOptionalParameterNode(node.Parameters[0]) {
			lastPositional = true

		} else if !opt {
			if i < len(node.Parameters)-1 {
				if isOptionalParameterNode(node.Parameters[i+1]) {
					opt = true
					lastPositional = true
				}

			} else if i == len(node.Parameters)-1 {
				lastPositional = true
			}
		}

		param, sig.ParamExpansionMin, sig.ParamExpansionMax, err =
			c.compileParameter(p, i+1, 0, lastPositional)

		if err != nil {
			return
		}
		params = append(params, param)
	}

	opt = false
	for _, p := range params {
		if p.DefaultValue == nil && len(p.DefaultValueInstructions) == 0 {
			if opt {
				// a positional parameter declared after optional parameters
				err = makeErr(node, "Cannot declare positional parameter after optional parameter")
				return
			}
			sig.ParamPositional = append(sig.ParamPositional, p)

		} else {
			// optional parameter
			opt = true
			sig.ParamByName = append(sig.ParamByName, p)
		}
	}

	return
}

func (c *Compiler) compileParameter(node ast.Node, pnum, level int, lastPositional bool) (
	param object.Parameter, paramExpansionMin, paramExpansionMax int, err error) {

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
			param, err = c.assessOptionalParameter(assign)
			if err != nil {
				err = makeErr(node, err.Error())
				return
			}
			system = assign.SystemAssignment

		default:
			err = makeErr(node, fmt.Sprintf("Parameter %d invalid", pnum))
			return
		}
		param.Mutable = p.Mutable

	case *ast.AssignmentNode:
		param, err = c.assessOptionalParameter(p)
		if err != nil {
			err = makeErr(node, err.Error())
			return
		}
		system = p.SystemAssignment

	case *ast.ExpansionNode:
		if level > 0 {
			err = makeErr(node, "Invalid parameter expansion node")
			return
		}
		if !lastPositional {
			err = makeErr(node, "Parameter expansion only allowed on last positional parameter")
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
					// [0..]
					paramExpansionMax = -1
				} else {
					paramExpansionMax, err = decodeInt(lim.Right)
				}

			default:
				err = makeErr(node, "Invalid expression for limits on parameter expansion")
			}

		default:
			err = makeErr(node, fmt.Sprintf("Invalid limit type on parameter expansion (%T)", lim))
		}
		if err == nil &&
			(paramExpansionMin < 0 || paramExpansionMax < -1 || paramExpansionMax == 0 ||
				(paramExpansionMin > paramExpansionMax && paramExpansionMax != -1)) {

			err = makeErr(node, "Invalid limits on parameter expansion")
		}
		if err != nil {
			return
		}

		param, _, _, err = c.compileParameter(p.Continuation, pnum, level+1, lastPositional)
		return

	default:
		err = makeErr(node, fmt.Sprintf("Parameter %d not a variable", pnum))
		return
	}

	// An external name (for an optional parameter) may shadow a keyword ...
	// since the context makes the meaning clear, ...
	// but an internal name (used within the function) may not.

	_, err = c.symbolTable.DefineVariable(param.InternalName, param.Mutable, system)
	if err != nil {
		err = makeErr(node, fmt.Sprintf("Parameter %d definition error: %s", pnum, err.Error()))
	}

	return
}

func (c *Compiler) assessOptionalParameter(assign *ast.AssignmentNode) (param object.Parameter, err error) {
	if len(assign.Identifiers) != 1 || len(assign.Values) != 1 {
		err = makeErr(assign, "Expected 1 identifier and 1 value for optional parameter assignment")
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
			err = makeErr(assign, "Expected identifier or identifier/alias for optional parameter")
		}

	default:
		err = makeErr(assign, "Expected identifier or identifier/alias for optional parameter")
		return
	}

	defaultIns, err := c.compileNode(assign.Values[0], true)
	if err != nil {
		err = makeErr(assign, fmt.Sprintf("Failure to compile default value for parameter %s: %s", str.ReformatInput(param.InternalName), err.Error()))
		return
	}
	param.DefaultValueInstructions = defaultIns

	return
}

func (c *Compiler) compileSelfRef(node ast.Node) (opcode.Instructions, error) {
	if c.symbolTable.Outer == nil {
		return nil, makeErr(node, "Cannot use self token in global scope")
	}
	return opcode.MakeWithErrTest(opcode.OpGetSelf)
}

func (c *Compiler) compileReturnNode(node *ast.ReturnNode) (ins opcode.Instructions, err error) {
	if c.functionLevel == 0 {
		err = makeErr(node, "Cannot use return outside of function")
		return
	}
	ins, err = c.compileNode(node.ReturnValue, true)
	if err != nil {
		return
	}
	ins = append(ins, opcode.Make(opcode.OpReturnValue)...)
	return
}

func (c *Compiler) compileCallNode(node *ast.CallNode) (ins opcode.Instructions, err error) {
	hasExpansion := false
	if len(node.Args) > 0 {
		switch post := node.Args[len(node.Args)-1].(type) {
		case *ast.PostfixExpressionNode:
			if post.Operator.Type == token.EXPANSION {
				node.Args[len(node.Args)-1] = post.Left
				hasExpansion = true
			}
		}
	}

	// Compiling the function first ...
	// ... but we add it to the instructions after the arguments.
	var fn opcode.Instructions
	fn, err = c.compileNode(node.Function, true)
	if err != nil {
		return
	}

	var bslc []byte
	for _, arg := range node.Args {
		bslc, err = c.compileNode(arg, true)
		if err != nil {
			break
		}
		ins = append(ins, bslc...)
	}

	// NOTE: putting function to call onto stack after arguments and will be popped first
	ins = append(ins, fn...)

	if hasExpansion {
		ins = append(ins, opcode.Make(opcode.OpCallWithExpansion, len(node.Args))...)
	} else {
		ins = append(ins, opcode.Make(opcode.OpCall, len(node.Args))...)
	}
	return
}
