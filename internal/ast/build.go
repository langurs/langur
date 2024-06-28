// langur/ast/build.go

package ast

import (
	"fmt"
	"langur/str"
	"langur/token"
)

func ListToStatements(nodes []Node) {
	for i := range nodes {
		switch nodes[i].(type) {
		case Statement:
			// matches Statement interface; do nothing
		default:
			// wrap into expression statement node
			nodes[i] = &ExpressionStatementNode{
				Token: nodes[i].TokenInfo(), Expression: nodes[i]}
		}
	}
}

func ListToBlock(nodes []Node) *BlockNode {
	blk := &BlockNode{}
	if len(nodes) != 0 {
		blk.Token = nodes[0].TokenInfo()
	}
	blk.Statements = nodes
	return blk
}

func BlockOrAsBlock(node Node) *BlockNode {
	blk, ok := node.(*BlockNode)
	if ok {
		return blk
	}
	return &BlockNode{Token: node.TokenInfo(), Statements: []Node{node}}
}

func AssignmentsToDeclarations(assignments []Node, mutable bool) (
	declarations []Node, err error) {

	for _, each := range assignments {
		var assign Node
		assign, err = AssignmentToDeclaration(each, mutable)
		if err != nil {
			break
		}
		declarations = append(declarations,
			&ExpressionStatementNode{Token: assign.TokenInfo(), Expression: assign})
	}
	return
}

func AssignmentToDeclaration(assignment Node, mutable bool) (decl Node, err error) {
	assign, ok := assignment.(*AssignmentNode)
	if !ok {
		err = fmt.Errorf("Cannot convert non-assignment to declaration")
		return assign, err
	}
	return &LineDeclarationNode{
		Token: assign.Token, Assignment: assign, Mutable: mutable}, nil
}

func FlattenDeclaration(decl *LineDeclarationNode) (declarations []*LineDeclarationNode, err error) {
	assign, ok := decl.Assignment.(*AssignmentNode)
	if !ok {
		err = fmt.Errorf("Expected assignment in declaration")
		return
	}

	if len(assign.Identifiers) == len(assign.Values) {
		for i := range assign.Identifiers {
			declarations = append(declarations,
				MakeDeclarationAssignmentExpression(
					assign.Identifiers[i], assign.Values[i],
					assign.SystemAssignment, decl.Mutable).(*LineDeclarationNode))
		}

	} else {
		// nothing we can flatten
		declarations = []*LineDeclarationNode{decl}
	}

	return
}

func MakeDeclarationAssignmentStatement(
	identifierNode, valueNode Node, systemAssignment, mutable bool) Node {

	return &ExpressionStatementNode{
		Token:      identifierNode.TokenInfo(),
		Expression: MakeDeclarationAssignmentExpression(identifierNode, valueNode, systemAssignment, mutable),
	}
	// wrapped it in a statement, so it may be popped off the stack by the VM ...
	// ... (assignment usually an expression, leaving it's value behind)
}

func MakeDeclarationAssignmentExpression(
	identifierNode, valueNode Node, systemAssignment, mutable bool) Node {

	return &LineDeclarationNode{
		Token:      identifierNode.TokenInfo(),
		Mutable:    mutable,
		Assignment: MakeAssignmentExpression(identifierNode, valueNode, systemAssignment),
	}
}

func MakeAssignmentStatement(
	identifierNode, valueNode Node, systemAssignment bool) Node {

	return &ExpressionStatementNode{
		Token:      identifierNode.TokenInfo(),
		Expression: MakeAssignmentExpression(identifierNode, valueNode, systemAssignment),
	}
}

func MakeAssignmentExpression(
	identifierNode, valueNode Node, systemAssignment bool) Node {

	var values []Node
	if valueNode != nil {
		values = []Node{valueNode}
	}

	return &AssignmentNode{
		Token:            identifierNode.TokenInfo(),
		SystemAssignment: systemAssignment,
		Identifiers:      []Node{identifierNode},
		Values:           values,
	}
}

func FixTryCatchInNodeSlice(nodes []Node, asExprStatement bool) (
	[]Node, error) {

	var expr *ExpressionStatementNode

	newNodes := nodes
	for i := 0; i < len(newNodes); i++ {
		// look for catch...
		tryCatch, ok := newNodes[i].(*TryCatchNode)
		if !ok {
			expr, ok = newNodes[i].(*ExpressionStatementNode)
			if ok {
				// try/catch node embedded in expression statement node
				tryCatch, ok = expr.Expression.(*TryCatchNode)
			}
		}
		if !ok || ok && tryCatch.Try != nil {
			// is not an incomplete TryCatch node
			continue
		}

		if i == 0 {
			return nil, fmt.Errorf("Catch cannot be first in a block of statements")
		}

		// move all preceding nodes within the statement slice into Try as a Block
		precedingNodes := newNodes[:i]

		if blk, ok := precedingNodes[0].(*BlockNode); ok && len(precedingNodes) == 1 {
			tryCatch.Try = blk

		} else {
			tryCatch.Try = &BlockNode{
				Token:      newNodes[i].TokenInfo(),
				Statements: precedingNodes}
		}

		if asExprStatement {
			// NOTE: wraps entire try/catch in an expression statement node, not the try and catch individually
			newNodes[i] = &ExpressionStatementNode{
				Token: newNodes[i].TokenInfo(), Expression: tryCatch}

		} else {
			newNodes[i] = tryCatch
		}
		newNodes = newNodes[i:]

		// start over at second node
		i = 0
	}
	return newNodes, nil
}

func MakeFunctionFromOperator(op token.Token, left, right Node) (
	fn *FunctionNode, err error) {

	if !token.IsInfixOp(op.Type) {
		err = fmt.Errorf("Expected infix operator")
		return
	}

	params := []Node{}

	if left == nil {
		left = NewVariableNode(op, "_x", true)
		params = append(params, left)
	}
	if right == nil {
		right = NewVariableNode(op, "_y", true)
		params = append(params, right)
	}

	fn = &FunctionNode{
		Token:      op,
		Parameters: params,
		Body: &BlockNode{
			Token: op,
			Statements: []Node{
				&InfixExpressionNode{
					Token:    op,
					Left:     left,
					Operator: op,
					Right:    right,
				},
			},
		},
	}
	return
}

// func MakeFoldingFunctionFromOperator(op token.Token, useParameterExpansion bool) (
// 	fn *FunctionNode, err error) {

// 	if !token.IsInfixOp(op.Type) {
// 		err = fmt.Errorf("Expected infix operator")
// 		return
// 	}

// 	left := NewVariableNode(op, "_x", true)
// 	right := NewVariableNode(op, "_y", true)
// 	action := &FunctionNode{
// 		Token:      op,
// 		Parameters: []Node{left, right},
// 		Body: &BlockNode{
// 			Statements: []Node{
// 				&InfixExpressionNode{
// 					Token:    op,
// 					Left:     left,
// 					Operator: op,
// 					Right:    right,
// 				},
// 			},
// 		},
// 	}

// 	foldx := NewVariableNode(op, "_foldx", true)
// 	params := []Node{foldx}
// 	if useParameterExpansion {
// 		params = []Node{
// 			&ExpansionNode{
//				Token: op,
// 				Continuation: foldx,
// 				Limits: &InfixExpressionNode{
//					Token: op,
// 					// expects at least 2 values
// 					Left:     MakeNumberFromString(op, "2"),
// 					Operator: node.TokenInfo().NewTokenCopyPosInfo(token.RANGE, "(..)"),
// 					// unlimited top end
// 					Right: MakeNumberFromString(op, "-1"),
// 				},
// 			},
// 		}
// 	}

// 	fn = &FunctionNode{
// 		Token:      op,
// 		Parameters: params,
// 		Body: &BlockNode{
// 			Statements: []Node{
//				Token:    op,
// 				&CallNode{
// 					Function: NewBuiltInNode(op, "_fold", true),
// 					Args:     []Node{action, foldx},
// 				},
// 			},
// 		},
// 	}
// 	return
// }

func MakeDecouplingAssignment(
	assign *AssignmentNode,
	tempCompositeResultVarNode Node,
	setResultsNodes, setNonResultsNodes []Node,
	expansionMin, expansionMax int) (Node, error) {

	minValues := len(assign.Identifiers)
	switch expansionMax {
	case 0: // good; no changes
	case -1:
		switch expansionMin {
		case 1: // good; no changes
		case 0:
			minValues--
		default:
			return nil, fmt.Errorf("Expansion minimum expected 0 or 1 on decoupling assignment")
		}
	default:
		return nil, fmt.Errorf("Expansion maximum expected 0 or blank (infinity) on decoupling assignment")
	}

	return &BlockNode{
		Token:    assign.Token,
		HasScope: true,
		Statements: []Node{
			// first evaluate and assign result to a temporary system variable
			MakeDeclarationAssignmentStatement(tempCompositeResultVarNode, assign.Values[0], true, false),

			&IfNode{
				Token: assign.Token,
				TestsAndActions: []TestDo{
					// if len(_Decouple_) < len(node.Identifiers) { ... } else { ... }
					{
						// check if enough elements available
						Test: &InfixExpressionNode{
							Token: assign.Token,
							Left: &CallNode{
								Token:    assign.Token,
								Function: NewBuiltInNode(assign.Token, "_len", true),
								Args:     []Node{tempCompositeResultVarNode},
							},
							Operator: assign.Token.NewTokenCopyPosInfo(token.LESS_THAN, "(<)"),
							Right:    MakeNumberFromInt(assign.Token, minValues),
						},
						// not enough elements
						Do: &BlockNode{Token: assign.Token,
							Statements: append(setNonResultsNodes,
								&BooleanNode{Token: assign.Token, Value: false})},
						// return false for failure
					},
					// } else {
					{
						Test: nil, // no test on else
						Do: &BlockNode{Token: assign.Token,
							Statements: append(setResultsNodes,
								&BooleanNode{Token: assign.Token, Value: true})},
						// return true for success after setting each variable
					},
				},
			},
		},
	}, nil
}

func MakeAssignmentIndexValueStatement(assignTo Node, left Node, start int, system bool,
	expansionMin, expansionMax int) (Node, error) {

	var index, alt Node
	index = MakeNumberFromInt(assignTo.TokenInfo(), start)

	// for expansion on decoupling assignment
	switch expansionMax {
	case 0: // good; do nothing with expansion
	case -1:
		switch expansionMin {
		case 0:
			// empty list returned instead of error
			alt = &ListNode{Token: assignTo.TokenInfo()}
			fallthrough
		case 1:
			index = &InfixExpressionNode{
				Token:    assignTo.TokenInfo(),
				Operator: assignTo.TokenInfo().NewTokenCopyPosInfo(token.RANGE, "(..)"),
				Left:     index,
				Right: &CallNode{
					Token:    assignTo.TokenInfo(),
					Function: NewBuiltInNode(assignTo.TokenInfo(), "_len", true),
					Args:     []Node{left},
				},
			}
		default:
			return nil, fmt.Errorf("Expansion minimum expected 0 or 1 on decoupling assignment")
		}
	default:
		return nil, fmt.Errorf("Expansion maximum expected 0 or -1 on decoupling assignment")
	}

	value := &IndexNode{
		Token:     left.TokenInfo(),
		Left:      left,
		Index:     index,
		Alternate: alt,
	}

	return MakeAssignmentStatement(assignTo, value, system), nil
}

func MakeNumberFromInt(tok token.Token, i int) *NumberNode {
	return &NumberNode{Token: tok, Base: 10, Value: str.IntToStr(i, 10)}
}

func MakeNumberFromString(tok token.Token, s string) *NumberNode {
	return &NumberNode{Token: tok, Base: 10, Value: s}
}
