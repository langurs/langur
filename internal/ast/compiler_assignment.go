// langur/ast/compiler_assignment.go

package ast

import (
	"fmt"
	"langur/opcode"
	"langur/symbol"
	"langur/token"
)

const implicitDecouplingExpansionMin = 0 // set to 0; must be 0 or 1

func (c *Compiler) makeOpSetInstructions(node Node, sym symbol.Symbol, level int) (
	pkg opcode.InsPackage, err error) {

	if sym.Scope == symbol.GlobalScope {
		pkg, err = opcode.MakePkgWithErrTest(node.TokenInfo(), opcode.OpSetGlobal, sym.Index)

	} else if sym.Scope == symbol.LocalScope {
		if level == 0 {
			pkg, err = opcode.MakePkgWithErrTest(node.TokenInfo(), opcode.OpSetLocal, sym.Index)
		} else {
			pkg, err = opcode.MakePkgWithErrTest(node.TokenInfo(), opcode.OpSetNonLocal, sym.Index, level)
		}

	} else {
		err = c.makeErr(node, fmt.Sprintf("Attempt to create OpSet instructions on %s for scope %s", sym.Name, sym.Scope))
	}
	return
}

func (c *Compiler) makeOpSetDefineInstructions(node Node) (
	pkg opcode.InsPackage, err error) {

	var temp opcode.InsPackage

	switch n := node.(type) {
	case Definable:
		pkg, err = n.CompileDefine(c)
		if err != nil {
			return
		}

		temp, err = opcode.MakePkgWithErrTest(node.TokenInfo(), opcode.OpDefine)
		if err != nil {
			return
		}
		pkg = pkg.Append(temp)

	default:
		err = c.makeErr(n, fmt.Sprintf("Attempt to create OpDefine instructions on non-definable (%T)", n))
	}

	return
}

// called by LineDeclarationNode.Compile()
func (c *Compiler) compileDeclarationAndAssignments(
	decl *LineDeclarationNode) (
	pkg opcode.InsPackage, err error) {

	assign, ok := decl.Assignment.(*AssignmentNode)
	if !ok {
		// parser failed
		bug("compileDeclarationAndAssignments", "Expected *AssignmentNode in *LineDeclarationNode")
		err = c.makeErr(assign, "Expected assignment in declaration")
		return
	}

	if decl.Public {
		// not ready to compile public declarations (future use)
		err = c.makeErr(assign, "Cannot compile public declaration (future use)")
		return
	}

	if assign.Values == nil || len(assign.Values) == len(assign.Identifiers) {
		// Compile values first (must be on the stack), then the setting instructions.
		var temp opcode.InsPackage
		// push values in reverse order
		for i := len(assign.Values) - 1; i > -1; i-- {
			temp, err = assign.Values[i].Compile(c)
			if err != nil {
				return
			}
			pkg = pkg.Append(temp)
		}

		for i, id := range assign.Identifiers {
			variable, ok := id.(*IdentNode)
			if !ok {
				bug("compileDeclarationAndAssignments", fmt.Sprintf("Wrong node for variable in Declaration Assignment node: %T", id))
			}

			var sym symbol.Symbol
			sym, err = c.symbolTable.DefineVariable(variable.Name, decl.Mutable, variable.System)
			if err != nil {
				err = c.makeErr(assign, err.Error())
				return
			}

			temp, err = c.makeOpSetInstructions(assign, sym, 0)
			if err != nil {
				err = c.makeErr(assign, err.Error())
				return
			}
			pkg = pkg.Append(temp)

			// pop all but the last one
			if i < len(assign.Identifiers)-1 {
				pkg = pkg.Append(opcode.MakePkg(id.TokenInfo(), opcode.OpPop))
			}
		}

	} else if len(assign.Values) == 1 {
		pkg, err = c.compileDecouplingDeclarationAssignment(assign, decl.Mutable)

	} else {
		// parser should have caught this...
		bug("compileDeclarationAndAssignments", "Identifier/value count mismatch in Declaration Assignment")
	}

	return
}

// see also compileDecouplingAssignment()
func (c *Compiler) compileDecouplingDeclarationAssignment(
	node *AssignmentNode, mutable bool) (
	pkg opcode.InsPackage, err error) {

	if len(node.Values) != 1 {
		// parser should have caught this...
		bug("compileDecouplingDeclarationAssignment", "Attempt to set declaration assignment decoupling when len(node.Values) != 1")
		return
	}

	var expansionMin, expansionMax int
	var temp Node

	tempCompositeResultVarNode := NewVariableNode(node.Token, "_Decouple_", true)

	setResultsNodes := []Node{}
	setNonResultsNodes := []Node{}
	for i, id := range node.Identifiers {
		switch id := id.(type) {
		case *NoneNode:
			// skip index number
			continue

		case *IdentNode:
			_, err = c.symbolTable.DefineVariable(id.Name, mutable, id.System)
			if err != nil {
				err = c.makeErr(node, err.Error())
				return
			}

			temp, err = MakeAssignmentIndexValueStatement(id, tempCompositeResultVarNode, i+1, true, 0, 0)
			if err != nil {
				return
			}

			setResultsNodes = append(setResultsNodes, temp)
			setNonResultsNodes = append(setNonResultsNodes, MakeAssignmentStatement(id, NoValue, true))

		case *ExpansionNode:
			if i < len(node.Identifiers)-1 {
				err = c.makeErr(node, "Expansion possible on last variable of decoupling assignment only")
				return
			}
			variable, ok := id.Continuation.(*IdentNode)
			if !ok {
				err = c.makeErr(node, "Expansion expected variable on decoupling assignment")
				return
			}

			_, err = c.symbolTable.DefineVariable(variable.Name, mutable, variable.System)
			if err != nil {
				err = c.makeErr(node, err.Error())
				return
			}

			if id.Limits != nil {
				err = c.makeErr(node, "Expansion limits not expected on decoupling assignment")
				return
			}
			expansionMin, expansionMax = implicitDecouplingExpansionMin, -1

			temp, err = MakeAssignmentIndexValueStatement(variable, tempCompositeResultVarNode, i+1, true, expansionMin, expansionMax)
			if err != nil {
				return
			}

			setResultsNodes = append(setResultsNodes, temp)
			setNonResultsNodes = append(setNonResultsNodes, MakeAssignmentStatement(variable, NoValue, true))

		default:
			err = c.makeErr(node, fmt.Sprintf("Invalid type for declaration assignment identifier: %T", id))
			return
		}
	}

	temp, err = MakeDecouplingAssignment(node, tempCompositeResultVarNode,
		setResultsNodes, setNonResultsNodes, expansionMin, expansionMax)

	if err == nil {
		pkg, err = temp.Compile(c)
	}
	return
}

// NOTE: mutability checked elsewhere; not checked in getVarAndDefinable
func (c *Compiler) getVarAndDefinable(node Node, expansionOk, altOk bool) (
	variable *IdentNode, definable Node, err error) {

	for i := 1; ; i++ {
		switch n := node.(type) {
		case *NoneNode:
			if i > 1 {
				err = c.makeErr(n, "Invalid use of none in assignment")
			}
			// skip
			return

		case *IdentNode:
			variable = n
			return

		case *IndexNode:
			if n.Alternate != nil && (
				i > 1 || !altOk) {
				err = c.makeErr(n, "Invalid use of alternate for index in assignment")
				return
			}

			if definable == nil {
				definable = n
			}

			switch left := n.Left.(type) {
			case *IdentNode:
				// x[1] = ...
				variable = left
				return

			default:
				if n.Alternate != nil {
					err = c.makeErr(n, "Invalid use of alternate for index in assignment")
					return
				}
				node = left
			}

		// FUTURE: dot notation
		// case *DotNode:

		case *ExpansionNode:
			if i > 1 || !expansionOk {
				err = c.makeErr(n, "Invalid use of expansion in assignment")
			}
			if n.Limits != nil {
				err = c.makeErr(node, "Expansion limits not expected on decoupling assignment")
				return
			}
			node = n.Continuation

		default:
			err = c.makeErr(n, fmt.Sprintf("Invalid type for assignment identifier: %T", n))
			return
		}
	}
}

func (c *Compiler) checkVarForAssignment(node *AssignmentNode, variable *IdentNode) (
	sym symbol.Symbol, cnt int, err error) {

	name := variable.Name
	var ok bool
	sym, cnt, ok = c.symbolTable.Resolve(name)
	if !ok {
		err = c.makeErr(node, fmt.Sprintf("Unable to resolve variable %s for assignment", name))
		return
	}

	if !sym.Mutable && !node.SystemAssignment {
		if variable.System {
			err = c.makeErr(node, fmt.Sprintf("Cannot assign to system variable %s (not mutable by user)", name))
		} else {
			err = c.makeErr(node, fmt.Sprintf("Cannot assign to immutable %s (use var instead of val to make mutable declaration)", name))
		}
	}

	return
}

// called by AssignmentNode.Compile()
// for things already declared; not for declaration assignment
func (c *Compiler) compileAssignment(node *AssignmentNode) (pkg opcode.InsPackage, err error) {
	if len(node.Values) == len(node.Identifiers) {
		// push values onto the stack in reverse order
		var temp opcode.InsPackage
		for i := len(node.Values) - 1; i > -1; i-- {
			temp, err = node.Values[i].Compile(c)
			if err != nil {
				return
			}
			pkg = pkg.Append(temp)
		}

		altOk := token.IsComboOp(node.Token) && len(node.Identifiers) == 1

		var variable *IdentNode
		var definable Node
		var sym symbol.Symbol
		var cnt int
		for i, id := range node.Identifiers {
			variable, definable, err = c.getVarAndDefinable(id, false, altOk)
			if err != nil {
				return
			}

			if variable != nil {
				sym, cnt, err = c.checkVarForAssignment(node, variable)
				if err != nil {
					return
				}
			}

			if definable == nil {
				if variable != nil {
					temp, err = c.makeOpSetInstructions(node, sym, cnt)
					if err != nil {
						return
					}
					pkg = pkg.Append(temp)
				}

			} else {
				if variable == nil {
					err = c.makeErr(id, "Invalid use of none in assignment")
					return
				}
				temp, err = c.makeOpSetDefineInstructions(definable)
				if err != nil {
					return
				}
				pkg = pkg.Append(temp)
			}

			// pop all but the last one
			if i < len(node.Identifiers)-1 {
				pkg = pkg.Append(opcode.MakePkg(id.TokenInfo(), opcode.OpPop))
			}
		}

	} else if len(node.Values) == 1 {
		pkg, err = c.compileDecouplingAssignment(node)

	} else {
		err = c.makeErr(node, "Identifier/value count mismatch in Assignment")
	}

	return
}

// see also compileDecouplingDeclarationAssignment()
// not a declaration decoupling assignment
func (c *Compiler) compileDecouplingAssignment(node *AssignmentNode) (
	pkg opcode.InsPackage, err error) {

	if len(node.Values) != 1 {
		bug("compileDecouplingAssignment", "Attempt to set assignment decoupling when len(node.Values) != 1")
		return
	}

	var expansionMin, expansionMax int
	var temp Node

	tempCompositeResultVarNode := NewVariableNode(node.Token, "_Decouple_", true)

	setResultsNodes := []Node{}
	for i, id := range node.Identifiers {
		switch id := id.(type) {
		case *NoneNode:
			// skip index number
			continue

		case *IdentNode, *IndexNode:
			temp, err = MakeAssignmentIndexValueStatement(id, tempCompositeResultVarNode, i+1, true, 0, 0)
			if err != nil {
				return
			}
			setResultsNodes = append(setResultsNodes, temp)

		case *ExpansionNode:
			if i < len(node.Identifiers)-1 {
				err = c.makeErr(node, "Expansion possible on last variable of decoupling assignment only")
				return
			}

			if id.Limits != nil {
				err = c.makeErr(node, "Expansion limits not expected on decoupling assignment")
				return
			}
			expansionMin, expansionMax = implicitDecouplingExpansionMin, -1

			switch variable := id.Continuation.(type) {
			case *IdentNode, *IndexNode:
				temp, err = MakeAssignmentIndexValueStatement(variable, tempCompositeResultVarNode, i+1, true, expansionMin, expansionMax)
				if err != nil {
					return
				}
				setResultsNodes = append(setResultsNodes, temp)

			default:
				err = c.makeErr(node, "Expansion expected variable on decoupling assignment")
				return
			}

		default:
			bug("compileDecouplingAssignment", fmt.Sprintf("Invalid type for assignment identifier: %T", id))
		}
	}

	temp, err = MakeDecouplingAssignment(node, tempCompositeResultVarNode,
		setResultsNodes, nil, expansionMin, expansionMax)

	if err == nil {
		pkg, err = temp.Compile(c)
	}
	return
}
