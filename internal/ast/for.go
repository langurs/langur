// langur/ast/for.go

package ast

import (
	"fmt"
	"langur/token"
)

// compiler does not see ForInOfNode (only ForNode)
func ConvertForInOfNodeToForNode(node *ForInOfNode) (Node, error) {
	if node.Of {
		switch node.Over.(type) {
		case *ListNode, *StringNode, *NumberNode:
			// optimization for times when we know it isn't necessary to check for a hash

		case *HashNode:
			return buildForOfOverHash(node)

		default:
			// might have a hash after evaluation
			return buildForOfOverPotentialHash(node)
		}
	}
	return convertForInOfNodeToForNode(node)
}

func convertForInOfNodeToForNode(node *ForInOfNode) (Node, error) {
	fornode := &ForNode{
		Token:         node.Token,
		LoopValueInit: node.LoopValueInit, // the loop expression value
		Body:          node.Body,
	}

	invalidIndex := NoValue

	var loopCtlVarList []*IdentNode

	switch node.Var.(type) {
	case nil:
		// variable-free for of loop; needs an internal control variable
		loopCtlVarList = append(loopCtlVarList, NewVariableNode(node.Token, "_LoopCtl_", true))

	case *IdentNode:
		loopCtlVarList = append(loopCtlVarList, node.Var.(*IdentNode))

	default:
		return fornode, fmt.Errorf("Invalid loop control variable in for in/of loop")
	}

	for _, v := range loopCtlVarList {
		if v.Name[0] == '_' && !v.System {
			return fornode, fmt.Errorf("Cannot use variable name starting with underscore as user-declared loop control variable")
		}
	}

	loopOverVar := NewVariableNode(node.Token, "_LoopOver_", true)
	// lets us do 3 things...
	// 1. pre-check the value type
	// 2. protect the loop over variable
	// 3. change the loop over value
	loopOverDecl := MakeDeclarationAssignmentStatement(loopOverVar, node.Over, true, false)

	mayNeedConversionOfLoopOver := true
	switch node.Over.(type) {
	case *ListNode, *StringNode:
		mayNeedConversionOfLoopOver = false
	}

	loopOverConversion := MakeAssignmentStatement(
		loopOverVar,
		&CallNode{
			Token:          node.Token,
			Function:       NewBuiltInNode(node.Token, "_values", true),
			PositionalArgs: []Node{loopOverVar}},
		true,
	)

	incrementVar := NewVariableNode(node.Token, "_LoopInc_", true)
	if node.Of {
		incrementVar = loopCtlVarList[0]
	}
	incrementDecl := MakeDeclarationAssignmentStatement(
		incrementVar, MakeNumberFromString(node.Token, "1"), true, false)

	// calculate limit once
	limitVar := NewVariableNode(node.Token, "_LoopLimit_", true)
	limit := &CallNode{
		Token:          node.Token,
		Function:       NewBuiltInNode(node.Token, "_len", true),
		PositionalArgs: []Node{loopOverVar},
	}
	if node.Of {
		limit = &CallNode{
			Token:          node.Token,
			Function:       NewBuiltInNode(node.Token, "_limit", true),
			PositionalArgs: []Node{loopOverVar},
		}
	}
	limitDecl := MakeDeclarationAssignmentStatement(limitVar, limit, true, false)

	// INIT SECTION
	if node.Of {
		// for ... of ...
		fornode.Init = append(fornode.Init, loopOverDecl)
		fornode.Init = append(fornode.Init, limitDecl)
		fornode.Init = append(fornode.Init, incrementDecl)

	} else {
		// for ... in ...
		// note: conversion of the loop-over variable ...
		// ... must occur before the increment initialization
		if mayNeedConversionOfLoopOver {
			fornode.Init = append(fornode.Init, loopOverDecl)
			fornode.Init = append(fornode.Init, loopOverConversion)
			fornode.Init = append(fornode.Init, limitDecl)

			for i, v := range loopCtlVarList {
				fornode.Init = append(fornode.Init, MakeDeclarationAssignmentStatement(
					v,
					&IndexNode{
						Token:     loopOverVar.Token,
						Left:      loopOverVar,
						Index:     MakeNumberFromInt(loopOverVar.Token, i+1),
						Alternate: invalidIndex,
					},
					true, false))
			}
			fornode.Init = append(fornode.Init, incrementDecl)

		} else {
			fornode.Init = append(fornode.Init, loopOverDecl)
			fornode.Init = append(fornode.Init, limitDecl)

			for i, v := range loopCtlVarList {
				fornode.Init = append(fornode.Init, MakeDeclarationAssignmentStatement(
					v,
					&IndexNode{
						Token:     loopOverVar.Token,
						Left:      loopOverVar,
						Index:     MakeNumberFromInt(loopOverVar.Token, i+1),
						Alternate: invalidIndex,
					},
					true, false))
			}
			fornode.Init = append(fornode.Init, incrementDecl)
		}
	}

	// TEST SECTION
	fornode.Test = &InfixExpressionNode{
		Token:    incrementVar.Token,
		Left:     incrementVar,
		Operator: incrementVar.Token.NewTokenCopyPosInfo(token.LT_OR_EQUAL, "(<=)"),
		Right:    limitVar,
	}

	// INCREMENT SECTION
	// increment variable
	fornode.Increment = []Node{
		MakeAssignmentStatement(
			incrementVar,
			&InfixExpressionNode{
				Token:    incrementVar.Token,
				Left:     incrementVar,
				Operator: incrementVar.Token.NewTokenCopyPosInfo(token.PLUS, "(+)"),
				Right:    MakeNumberFromInt(incrementVar.Token, len(loopCtlVarList)),
			},
			true,
		),
	}
	if !node.Of {
		// indexed values (must be set after increment variable)
		for i, v := range loopCtlVarList {
			var index Node = incrementVar
			if i > 0 {
				index = &InfixExpressionNode{
					Token:    incrementVar.Token,
					Left:     incrementVar,
					Operator: incrementVar.Token.NewTokenCopyPosInfo(token.PLUS, "(+)"),
					Right:    MakeNumberFromInt(incrementVar.Token, i),
				}
			}
			fornode.Increment = append(fornode.Increment,
				MakeAssignmentStatement(
					v,
					&IndexNode{
						Token:     loopOverVar.Token,
						Left:      loopOverVar,
						Index:     index,
						Alternate: invalidIndex,
					},
					true,
				),
			)
		}
	}

	return fornode, nil
}

func buildForOfOverPotentialHash(node *ForInOfNode) (Node, error) {
	var fornode Node
	var err error

	// evaluate loop over value once
	// 1. checked to see if it is a hash
	// 2. used within a selected for loop afterwards
	loopOverVar := NewVariableNode(node.Token, "_LoopOverPreEval_", true)
	loopOverDecl := MakeDeclarationAssignmentStatement(loopOverVar, node.Over, true, false)
	node.Over = loopOverVar

	// make a regular for node
	fornode, err = convertForInOfNodeToForNode(node)
	if err != nil {
		return nil, err
	}

	// make a for of loop over hash keys (a for in loop really)
	var hashKeysForNode Node
	hashKeysForNode, err = buildForOfOverHash(node)
	if err != nil {
		return nil, err
	}

	// make an if node to choose which for loop to run
	ifnode := &IfNode{
		Token: node.Token,
		TestsAndActions: []TestDo{
			{ // if is hash
				Test: &CallNode{
					Token:          node.Token,
					Function:       NewBuiltInNode(node.Token, "_is_hash", true),
					PositionalArgs: []Node{node.Over},
				},
				Do: &BlockNode{Token: node.Token, Statements: []Node{hashKeysForNode}},
			},
			{ // else
				Do: &BlockNode{Token: node.Token, Statements: []Node{fornode}},
			},
		},
	}

	// enclose in a scoped block, first declaring and setting the loop over pre-evaluation
	blockNode := &BlockNode{
		HasScope: true,
		Statements: []Node{
			loopOverDecl,
			ifnode,
		},
	}

	return blockNode, nil
}

func buildForOfOverHash(node *ForInOfNode) (Node, error) {
	// generate a for in loop over hash keys
	var err error
	inKeys := node.Copy().(*ForInOfNode)
	inKeys.Of = false
	inKeys.Over = &CallNode{
		Token:          node.Token,
		Function:       NewBuiltInNode(node.Token, "_keys", true),
		PositionalArgs: []Node{inKeys.Over},
	}
	var hashKeysForNode Node
	hashKeysForNode, err = convertForInOfNodeToForNode(inKeys)
	if err != nil {
		return nil, err
	}
	return hashKeysForNode, nil
}
