// langur/ast/search.go

package ast

import (
	"reflect"
)

func IsSimple(node Node) bool {
	switch n := node.(type) {
	case *StringNode:
		return len(n.Interpolations) == 0

	case *RegexNode:
		return IsSimple(n.Pattern)

	case *DateTimeNode:
		return IsSimple(n.Pattern)

	case *DurationNode:
		return IsSimple(n.Pattern)

	case *NumberNode, *BooleanNode, *NullNode, *NoneNode,
		*FallThroughNode, *NextNode,
		*SelfNode, *IdentNode:

		return true

	case *BreakNode:
		return n.Value == nil
	}
	return false
}

func EndsWithDefiniteJump(nodes []Node) bool {
	if len(nodes) > 0 {
		switch n := nodes[len(nodes)-1].(type) {
		case *FallThroughNode, *BreakNode, *NextNode, *ThrowNode, *ReturnNode:
			return true
		case *BlockNode:
			return EndsWithDefiniteJump(n.Statements)
		}
	}
	return false
}

// for testing whether to wrap into scope
func NodeContainsFirstScopeLevelDeclaration(node Node) bool {
	return NodeContainsFirstScopeLevelSomething(node, 0,
		[]Node{&LineDeclarationNode{Token: node.TokenInfo()}, &ModeNode{Token: node.TokenInfo()}})
}

func NodeContainsFirstScopeLevelSomething(node Node, level int, ofTypes []Node) bool {
	if level > 1 {
		return false
	}

	for _, n := range ofTypes {
		if reflect.TypeOf(node) == reflect.TypeOf(n) {
			return true
		}
	}

	switch n := node.(type) {
	// not going into a function node here...

	case *Program:
		for _, v := range n.Statements {
			if NodeContainsFirstScopeLevelSomething(v, level, ofTypes) {
				return true
			}
		}

	case *LineDeclarationNode:
		return NodeContainsFirstScopeLevelSomething(n.Assignment, level+1, ofTypes)

	case *AssignmentNode:
		for _, v := range n.Values {
			if NodeContainsFirstScopeLevelSomething(v, level, ofTypes) {
				return true
			}
		}

	case *ModeNode:
		return NodeContainsFirstScopeLevelSomething(n.Setting, level, ofTypes)

	case *CallNode:
		for _, arg := range n.PositionalArgs {
			if NodeContainsFirstScopeLevelSomething(arg, level, ofTypes) {
				return true
			}
		}
		for _, arg := range n.ByNameArgs {
			if NodeContainsFirstScopeLevelSomething(arg, level, ofTypes) {
				return true
			}
		}

	case *StringNode:
		for _, v := range n.Interpolations {
			if NodeContainsFirstScopeLevelSomething(v, level, ofTypes) {
				return true
			}
		}

	case *InterpolatedNode:
		return NodeContainsFirstScopeLevelSomething(n.Value, level, ofTypes)
	case *RegexNode:
		return NodeContainsFirstScopeLevelSomething(n.Pattern, level, ofTypes)
	case *DateTimeNode:
		return NodeContainsFirstScopeLevelSomething(n.Pattern, level, ofTypes)
	case *DurationNode:
		return NodeContainsFirstScopeLevelSomething(n.Pattern, level, ofTypes)

	case *ListNode:
		for _, e := range n.Elements {
			if NodeContainsFirstScopeLevelSomething(e, level, ofTypes) {
				return true
			}
		}

	case *HashNode:
		for _, kv := range n.Pairs {
			if NodeContainsFirstScopeLevelSomething(kv.Key, level, ofTypes) {
				return true
			}
			if NodeContainsFirstScopeLevelSomething(kv.Value, level, ofTypes) {
				return true
			}
		}

	case *IndexNode:
		return NodeContainsFirstScopeLevelSomething(n.Left, level, ofTypes) ||
			NodeContainsFirstScopeLevelSomething(n.Index, level, ofTypes) ||
			NodeContainsFirstScopeLevelSomething(n.Alternate, level, ofTypes)

	case *BlockNode:
		add := 0
		if n.HasScope {
			add = 1
		}
		for _, stmt := range n.Statements {
			if NodeContainsFirstScopeLevelSomething(stmt, level+add, ofTypes) {
				return true
			}
		}

	case *ExpressionStatementNode:
		return NodeContainsFirstScopeLevelSomething(n.Expression, level, ofTypes)

	case *InfixExpressionNode:
		return NodeContainsFirstScopeLevelSomething(n.Left, level, ofTypes) ||
			NodeContainsFirstScopeLevelSomething(n.Right, level, ofTypes)

	case *PrefixExpressionNode:
		return NodeContainsFirstScopeLevelSomething(n.Right, level, ofTypes)

	case *PostfixExpressionNode:
		return NodeContainsFirstScopeLevelSomething(n.Left, level, ofTypes)

	case *ForNode:
		return NodeContainsFirstScopeLevelSomething(n.Body, level+1, ofTypes)

	case *ForInOfNode:
		return NodeContainsFirstScopeLevelSomething(n.Body, level+1, ofTypes)

	case *IfNode:
		for _, ta := range n.TestsAndActions {
			if NodeContainsFirstScopeLevelSomething(ta.Test, level+1, ofTypes) ||
				NodeContainsFirstScopeLevelSomething(ta.Do, level+1, ofTypes) {
				return true
			}
		}

	case *SwitchNode:
		for _, e := range n.Expressions {
			if NodeContainsFirstScopeLevelSomething(e.Expr, level+1, ofTypes) {
				return true
			}
		}
		for _, ca := range n.CasesAndActions {
			if NodeContainsFirstScopeLevelSomething(ca.Do, level+1, ofTypes) {
				return true
			}
			for _, cond := range ca.MatchConditions {
				if NodeContainsFirstScopeLevelSomething(cond, level+1, ofTypes) {
					return true
				}
			}
			for _, cond := range ca.OtherConditions {
				if NodeContainsFirstScopeLevelSomething(cond, level+1, ofTypes) {
					return true
				}
			}
		}

	case *TryCatchNode:
		return NodeContainsFirstScopeLevelSomething(n.Try, level, ofTypes) ||
			NodeContainsFirstScopeLevelSomething(n.Catch, level+1, ofTypes) ||
			NodeContainsFirstScopeLevelSomething(n.Else, level+1, ofTypes)

	case *ThrowNode:
		return NodeContainsFirstScopeLevelSomething(n.Exception, level, ofTypes)

	case *ReturnNode:
		return NodeContainsFirstScopeLevelSomething(n.ReturnValue, level, ofTypes)

	case *ExpansionNode:
		return NodeContainsFirstScopeLevelSomething(n.Limits, level, ofTypes) ||
			NodeContainsFirstScopeLevelSomething(n.Continuation, level, ofTypes)
	}

	return false
}
