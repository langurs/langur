// langur/ast/switch.go

package ast

import (
	"fmt"
	"langur/token"
)

// Since a switch expression is a glorified if/else, ...
// ... there's no need for the compiler to deal with it ...
// ... (except for fallthrough and for catch switch exceptions (when no default given)).
func ConvertSwitchNodeToIfNode(sw *SwitchNode, defaultCompOp token.Token) (*IfNode, error) {
	var testDo TestDo

	ifnode := &IfNode{Token: sw.TokenInfo(), IsSwitchExpr: true, CatchException: sw.CatchException}

	buildCondition := func(
		variable, condition Node,
		op token.Token, variableAsLeftOperandOnOriginalExpr bool) (
		test Node, err error) {

		test = condition.Copy()

		condInfix, condIsInfix := condition.(*InfixExpressionNode)

		if variable == nil {
			// extra tests; not compared with a variable
			if condIsInfix &&
				(condInfix.Left == nil || condInfix.Right == nil) {
				err = fmt.Errorf("Cannot use incomplete expression without comparison to expression in switch case condition")
			}

		} else {
			// within count of variables; compare with one variable
			if condIsInfix && condInfix.Left == nil {
				// incomplete infix, such as case > 10: ...
				// finish infix expression variable on left
				test.(*InfixExpressionNode).Left = variable

			} else if condIsInfix && condInfix.Right == nil {
				// incomplete infix, such as case 10 >: ...
				// finish infix expression variable on right
				test.(*InfixExpressionNode).Right = variable

			} else if variableAsLeftOperandOnOriginalExpr {
				test = &InfixExpressionNode{
					Token: test.TokenInfo(),
					Left:  variable, Operator: op, Right: condition}

			} else {
				test = &InfixExpressionNode{
					Token: test.TokenInfo(),
					Left:  condition, Operator: op, Right: variable}
			}
		}

		return
	}

	addTest := func(td *TestDo, test Node, logicalOp token.Token) {
		if td.Test == nil {
			td.Test = test
		} else {
			td.Test = &InfixExpressionNode{
				Token: td.Test.TokenInfo(),
				Left:  td.Test, Operator: logicalOp, Right: test}
		}
	}

	const noTests = "Expected condition test for switch case"

	if len(sw.CasesAndActions) == 0 {
		return nil, fmt.Errorf("Empty switch expression not allowed")
	}

	for i, cd := range sw.CasesAndActions {
		last := i == len(sw.CasesAndActions)-1

		testDo.Test = nil
		testDo.Do = cd.Do

		// TEST CONDITIONS
		if cd.MatchConditions == nil && cd.OtherConditions == nil {
			// last condition a "default"
			if !last {
				return nil, fmt.Errorf(noTests)
			}

		} else {
			// not the "default"

			if len(cd.MatchConditions) == 1 {
				// 1 condition and possibly multiple expressions
				// all variables tested against same condition
				if IsNoOp(cd.MatchConditions[0]) {
					return nil, fmt.Errorf(noTests)
				}

				for i := range sw.Expressions {
					test, err := buildCondition(
						sw.Expressions[i].Expr, cd.MatchConditions[0],
						sw.Expressions[i].Op, sw.Expressions[i].LeftOperand)

					if err != nil {
						return nil, err
					}

					addTest(&testDo, test, cd.MatchLogicalOp)
				}

			} else {
				// none or multiple conditions
				// 1 or more expressions
				for i, condition := range cd.MatchConditions {
					if !IsNoOp(condition) {
						j := i
						if j >= len(sw.Expressions) {
							// more conditions than expressions; use last expression
							j = len(sw.Expressions) - 1
						}

						test, err := buildCondition(sw.Expressions[j].Expr, condition,
							sw.Expressions[j].Op, sw.Expressions[j].LeftOperand)

						if err != nil {
							return nil, err
						}

						addTest(&testDo, test, cd.MatchLogicalOp)
					}
				}
			} // variable and condition count

			// check other conditions (not tied to any test expressions)
			for _, condition := range cd.OtherConditions {
				if IsNoOp(condition) {
					return nil, fmt.Errorf("Invalid use of no-op in switch case condition")
				}

				test, err := buildCondition(nil, condition, defaultCompOp, false)
				if err != nil {
					return nil, err
				}

				addTest(&testDo, test, cd.OtherLogicalOp)
			}

			if testDo.Test == nil {
				// all used no op; not allowed (that's "default")
				return nil, fmt.Errorf(noTests)
			}
		} // case or default

		// ACTION
		if len(testDo.Do.(*BlockNode).Statements) == 0 {
			// no case body statements; is implicit fallthrough
			testDo.Do.(*BlockNode).Statements = []Node{&FallThroughNode{Token: testDo.Do.TokenInfo()}}
		}

		ifnode.TestsAndActions = append(ifnode.TestsAndActions, testDo)
	}

	return ifnode, nil
}

func ExtractDeclarationsAndAssignmentsForSwitchTests(sw *SwitchNode) (nodes []Node, err error) {
	var declarationsAndAssignments []Node

	var extract func(node *Node)
	extract = func(node *Node) {
		switch n := (*node).(type) {
		case *LineDeclarationNode:
			declarationsAndAssignments = append(declarationsAndAssignments, n)
			if len(n.Assignment.(*AssignmentNode).Identifiers) == 1 {
				*node = n.Assignment.(*AssignmentNode).Identifiers[0]
			} else {
				err = fmt.Errorf("Cannot use multi-variable declaration in switch expression test value")
			}

		case *AssignmentNode:
			declarationsAndAssignments = append(declarationsAndAssignments, n)
			if len(n.Identifiers) == 1 {
				*node = n.Identifiers[0]
			} else {
				err = fmt.Errorf("Cannot use multi-variable assignment in switch expression test value")
			}

		case *ExpressionStatementNode:
			extract(&n.Expression)

		case *InfixExpressionNode:
			extract(&n.Left)
			extract(&n.Right)

		case *PrefixExpressionNode:
			extract(&n.Right)

		default:
			err = fmt.Errorf("Error determining switch expression test value")
		}
	}

	for i := range sw.Expressions {
		extract(&sw.Expressions[i].Expr)
	}

	return declarationsAndAssignments, nil
}

func ShouldConvertSwitchTestToSystemAssignment(node Node) bool {
	if IsSimple(node) {
		return false
	}
	switch n := node.(type) {
	case *LineDeclarationNode, *AssignmentNode:
		return false
	case *ExpressionStatementNode:
		return ShouldConvertSwitchTestToSystemAssignment(n.Expression)
	}
	return true
}
