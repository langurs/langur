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

	isImpure := node.Impure // self-declared impure; to be tested later...
	defer func() {
		c.popVariableScope()
		c.functionLevel--

		// Impurity is transitive.
		if isImpure {
			c.addToImpuritiesList(node.Name)
		}
	}()

	var body opcode.Instructions

	switch node.Name {
	case "_main":
		if len(node.Parameters) != 0 {
			err = makeErr(node, "Function _main() cannot have parameters")
			return
		}

	case "":
		// no defining self

	default:
		c.symbolTable.DefineSelf(node.Name)
	}

	var paramExpansionMin, paramExpansionMax int
	paramExpansionMin, paramExpansionMax, err = c.compileFunctionNodeParameters(node)
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
	isImpure = isImpure || c.symbolTable.Impurities != nil

	if isImpure && !node.Impure {
		if node.Name == "" {
			err = makeErr(node, "Anonymous impure function not declared as impure; use a * to declare impurity, such as fn*() { }")
		} else {
			err = makeErr(node, fmt.Sprintf("Impure function (%s) not declared as impure; use a * to declare impurity, such as fn*() { }", str.ReformatInput(node.Name)))
		}
		return
	}

	compiledFn := &object.CompiledCode{
		Name:               node.Name,
		IsFunction:         true,
		Instructions:       body,
		LocalBindingsCount: localsCount,
		ParamMin:           len(node.Parameters),
		ParamMax:           len(node.Parameters),
		ParamExpansionMin:  paramExpansionMin,
		ParamExpansionMax:  paramExpansionMax,
		ImpureEffects:      isImpure,
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

func (c *Compiler) compileFunctionNodeParameters(node *ast.FunctionNode) (
	paramExpansionMin, paramExpansionMax int, err error) {

	for i, p := range node.Parameters {
		_, paramExpansionMin, paramExpansionMax, err = c.compileParameter(p, i+1, 0, i == len(node.Parameters)-1)
		if err != nil {
			return
		}
	}

	return
}

func (c *Compiler) compileParameter(node ast.Node, pnum, level int, last bool) (
	mutable bool, paramExpansionMin, paramExpansionMax int, err error) {

	var name string
	system := false
	mutable = false

	switch p := node.(type) {
	case *ast.IdentNode:
		name = p.Name
		system = p.System

	case *ast.LineDeclarationNode:
		// to use var token to make parameter mutable
		switch assign := p.Assignment.(type) {
		case *ast.AssignmentNode:
			// no optional parameters option
			err = makeErr(node, fmt.Sprintf("Parameter %d contains assignment", pnum))
			return

		case *ast.IdentNode:
			name = assign.Name
			system = assign.System

		default:
			err = makeErr(node, fmt.Sprintf("Parameter %d invalid", pnum))
			return
		}
		mutable = p.Mutable

	case *ast.AssignmentNode:
		// no optional parameters option
		err = makeErr(node, fmt.Sprintf("Parameter %d contains assignment", pnum))
		return

	case *ast.ExpansionNode:
		// mutable, paramExpansionMin, paramExpansionMax, err =
		// 	c.compileParameterExpansion(node, p, pnum, level, last)

		if level > 0 {
			err = makeErr(node, "Invalid parameter expansion node")
			return
		}
		if !last {
			err = makeErr(node, "Parameter expansion only allowed on last parameter")
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

		mutable, _, _, err = c.compileParameter(p.Continuation, pnum, level+1, last)
		return

	default:
		err = makeErr(node, fmt.Sprintf("Parameter %d not a variable", pnum))
		return
	}

	_, err = c.symbolTable.DefineVariable(name, mutable, system)
	if err != nil {
		err = makeErr(node, fmt.Sprintf("Parameter %d definition error: %s", pnum, err.Error()))
	}

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
