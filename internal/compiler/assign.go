// langur/compiler/assign.go

package compiler

import (
	"fmt"
	"langur/ast"
	"langur/opcode"
	"langur/symbol"
)

const implicitDecouplingExpansionMin = 0 // 0 or 1

func (c *Compiler) makeOpSetInstructions(node ast.Node, sym symbol.Symbol, level int) (
	ins opcode.Instructions, err error) {

	if sym.Scope == symbol.GlobalScope {
		ins, err = opcode.MakeWithErrTest(opcode.OpSetGlobal, sym.Index)

	} else if sym.Scope == symbol.LocalScope {
		if level == 0 {
			ins, err = opcode.MakeWithErrTest(opcode.OpSetLocal, sym.Index)
		} else {
			ins, err = opcode.MakeWithErrTest(opcode.OpSetNonLocal, sym.Index, level)
		}

	} else {
		err = c.makeErr(node, fmt.Sprintf("Attempt to create OpSet instructions on %s for scope %s", sym.Name, sym.Scope))
	}
	return
}

func (c *Compiler) makeOpSetIndexInstructions(node ast.Node, sym symbol.Symbol, level int, index ast.Node) (
	ins opcode.Instructions, err error) {

	var temp opcode.Instructions

	ins, err = c.compileNode(index)
	if err != nil {
		return
	}

	if sym.Scope == symbol.GlobalScope {
		temp, err = opcode.MakeWithErrTest(opcode.OpSetGlobalIndexedValue, sym.Index)
		if err != nil {
			err = c.makeErr(node, err.Error())
			return
		}
		ins = append(ins, temp...)

	} else if sym.Scope == symbol.LocalScope {
		if level == 0 {
			temp, err = opcode.MakeWithErrTest(opcode.OpSetLocalIndexedValue, sym.Index)
			if err != nil {
				err = c.makeErr(node, err.Error())
				return
			}
			ins = append(ins, temp...)
		} else {
			temp, err = opcode.MakeWithErrTest(opcode.OpSetNonLocalIndexedValue, sym.Index, level)
			if err != nil {
				err = c.makeErr(node, err.Error())
				return
			}
			ins = append(ins, temp...)
		}

	} else {
		err = c.makeErr(node, fmt.Sprintf("Attempt to create OpSet Indexed instructions on %s for scope %s", sym.Name, sym.Scope))
	}
	return
}

func (c *Compiler) compileDeclarationAndAssignments(
	decl *ast.LineDeclarationNode) (
	ins opcode.Instructions, err error) {

	assign, ok := decl.Assignment.(*ast.AssignmentNode)
	if !ok {
		// parser failed
		bug("compileDeclarationAndAssignments", "Expected *ast.AssignmentNode in *ast.LineDeclarationNode")
		err = c.makeErr(assign, "Expected assignment in declaration")
		return
	}

	if assign.Values == nil || len(assign.Values) == len(assign.Identifiers) {
		// Compile values first (must be on the stack), then the setting instructions.
		var temp opcode.Instructions
		// push values in reverse order
		for i := len(assign.Values) - 1; i > -1; i-- {
			temp, err = c.compileNode(assign.Values[i])
			if err != nil {
				return
			}
			ins = append(ins, temp...)
		}

		for i, id := range assign.Identifiers {
			variable, ok := id.(*ast.IdentNode)
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
			ins = append(ins, temp...)

			// pop all but the last one
			if i < len(assign.Identifiers)-1 {
				ins = append(ins, opcode.Make(opcode.OpPop)...)
			}
		}

	} else if len(assign.Values) == 1 {
		ins, err = c.compileDecouplingDeclarationAssignment(assign, decl.Mutable)

	} else {
		// parser should have caught this...
		bug("compileDeclarationAndAssignments", "Identifier/value count mismatch in Declaration Assignment")
	}

	return
}

func (c *Compiler) compileAssignment(node *ast.AssignmentNode) (ins opcode.Instructions, err error) {
	// not a declaration assignment

	if len(node.Values) == len(node.Identifiers) {
		// push values in reverse order
		var temp opcode.Instructions
		for i := len(node.Values) - 1; i > -1; i-- {
			temp, err = c.compileNode(node.Values[i])
			if err != nil {
				return
			}
			ins = append(ins, temp...)
		}

		var variable *ast.IdentNode
		var index ast.Node
		for i, id := range node.Identifiers {
			switch n := id.(type) {
			case *ast.IdentNode:
				variable = n
				index = nil

			case *ast.IndexNode:
				// x[1] = ...
				variable = n.Left.(*ast.IdentNode)
				index = n.Index

			default:
				err = c.makeErr(node, fmt.Sprintf("Invalid type for assignment identifier: %T", n))
				return
			}

			name := variable.Name
			sym, cnt, ok := c.symbolTable.Resolve(name)
			if !ok {
				err = c.makeErr(node, fmt.Sprintf("Unable to resolve variable %s for assignment", name))
				return
			}

			if !sym.Mutable && !node.SystemAssignment {
				if variable.System {
					err = c.makeErr(node, fmt.Sprintf("Cannot assign to system variable %s (not mutable by user)", name))
					return
				} else {
					err = c.makeErr(node, fmt.Sprintf("Cannot assign to immutable %s (use var instead of val to make mutable declaration)", name))
					return
				}
			}

			if index == nil {
				temp, err = c.makeOpSetInstructions(node, sym, cnt)
				if err != nil {
					return
				}
				ins = append(ins, temp...)

			} else {
				temp, err = c.makeOpSetIndexInstructions(node, sym, cnt, index)
				if err != nil {
					return
				}
				ins = append(ins, temp...)
			}

			// pop all but the last one
			if i < len(node.Identifiers)-1 {
				ins = append(ins, opcode.Make(opcode.OpPop)...)
			}
		}

	} else if len(node.Values) == 1 {
		ins, err = c.compileDecouplingAssignment(node)

	} else {
		bug("compileAssignment", "Identifier/value count mismatch in Assignment")
	}

	return
}

// see also compileDecouplingAssignment()
func (c *Compiler) compileDecouplingDeclarationAssignment(
	node *ast.AssignmentNode, mutable bool) (
	ins opcode.Instructions, err error) {

	if len(node.Values) != 1 {
		// parser should have caught this...
		bug("compileDecouplingDeclarationAssignment", "Attempt to set declaration assignment decoupling when len(node.Values) != 1")
		return
	}

	var expansionMin, expansionMax int
	var temp ast.Node

	tempCompositeResultVarNode := ast.NewVariableNode(node.Token, "_Decouple_", true)

	setResultsNodes := []ast.Node{}
	setNonResultsNodes := []ast.Node{}
	for i, id := range node.Identifiers {
		switch id := id.(type) {
		case *ast.NoneNode:
			// skip index number
			continue

		case *ast.IdentNode:
			_, err = c.symbolTable.DefineVariable(id.Name, mutable, id.System)
			if err != nil {
				err = c.makeErr(node, err.Error())
				return
			}

			temp, err = ast.MakeAssignmentIndexValueStatement(id, tempCompositeResultVarNode, i+1, true, 0, 0)
			if err != nil {
				return
			}

			setResultsNodes = append(setResultsNodes, temp)
			setNonResultsNodes = append(setNonResultsNodes, ast.MakeAssignmentStatement(id, ast.NoValue, true))

		case *ast.ExpansionNode:
			if i < len(node.Identifiers)-1 {
				err = c.makeErr(node, "Expansion possible on last variable of decoupling assignment only")
				return
			}
			variable, ok := id.Continuation.(*ast.IdentNode)
			if !ok {
				err = c.makeErr(node, "Expansion expected variable on decoupling assignment")
				return
			}

			_, err = c.symbolTable.DefineVariable(variable.Name, mutable, variable.System)
			if err != nil {
				err = c.makeErr(node, err.Error())
				return
			}

			expansionMin, expansionMax = implicitDecouplingExpansionMin, -1

			temp, err = ast.MakeAssignmentIndexValueStatement(variable, tempCompositeResultVarNode, i+1, true, expansionMin, expansionMax)
			if err != nil {
				return
			}

			setResultsNodes = append(setResultsNodes, temp)
			setNonResultsNodes = append(setNonResultsNodes, ast.MakeAssignmentStatement(variable, ast.NoValue, true))

		default:
			err = c.makeErr(node, fmt.Sprintf("Invalid type for declaration assignment identifier: %T", id))
			return
		}
	}

	temp, err = ast.MakeDecouplingAssignment(node, tempCompositeResultVarNode,
		setResultsNodes, setNonResultsNodes, expansionMin, expansionMax)

	if err == nil {
		ins, err = c.compileNode(temp)
	}
	return
}

// see also compileDecouplingDeclarationAssignment()
func (c *Compiler) compileDecouplingAssignment(node *ast.AssignmentNode) (
	ins opcode.Instructions, err error) {
	// not a declaration assignment

	if len(node.Values) != 1 {
		bug("compileDecouplingAssignment", "Attempt to set assignment decoupling when len(node.Values) != 1")
		return
	}

	var expansionMin, expansionMax int
	var temp ast.Node

	tempCompositeResultVarNode := ast.NewVariableNode(node.Token, "_Decouple_", true)

	setResultsNodes := []ast.Node{}
	for i, id := range node.Identifiers {
		switch id := id.(type) {
		case *ast.NoneNode:
			// skip index number
			continue

		case *ast.IdentNode, *ast.IndexNode:
			temp, err = ast.MakeAssignmentIndexValueStatement(id, tempCompositeResultVarNode, i+1, true, 0, 0)
			if err != nil {
				return
			}
			setResultsNodes = append(setResultsNodes, temp)

		case *ast.ExpansionNode:
			if i < len(node.Identifiers)-1 {
				err = c.makeErr(node, "Expansion possible on last variable of decoupling assignment only")
				return
			}
			expansionMin, expansionMax = implicitDecouplingExpansionMin, -1

			switch variable := id.Continuation.(type) {
			case *ast.IdentNode, *ast.IndexNode:
				temp, err = ast.MakeAssignmentIndexValueStatement(variable, tempCompositeResultVarNode, i+1, true, expansionMin, expansionMax)
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

	temp, err = ast.MakeDecouplingAssignment(node, tempCompositeResultVarNode,
		setResultsNodes, nil, expansionMin, expansionMax)

	if err == nil {
		ins, err = c.compileNode(temp)
	}
	return
}
