// langur/ast/search.go

package ast

import (
	"reflect"
)

func nodeIsOfType(node Node, ofTypes []Node) bool {
	for _, n := range ofTypes {
		if reflect.TypeOf(node) == reflect.TypeOf(n) {
			return true
		}
	}
	return false
}

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

func CopyBeforeAssignment(val Node) bool {
	// copy value before assignment?
	// inefficient to always copy

	// FIXME: a temporary patch; inefficient
	return true

	// ofTypes := []Node{&IdentNode{}}
	// dontSearch := []Node{
	// 	&LineDeclarationNode{}, &ModeNode{}, &CallNode{}, &AssignmentNode{},
	// 	&StringNode{}, &RegexNode{}, &DateTimeNode{}, &DurationNode{},
	// }

	// return val.Search(ofTypes, dontSearch)
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

// a convenience function to list nodes to search
func searchNodes(ofTypes, dontSearch []Node, checkNodes ...Node) (found bool) {
	return searchNodeSlice(ofTypes, dontSearch, checkNodes)
}

func searchNodeSlice(ofTypes, dontSearch []Node, checkNodes []Node) (found bool) {
	for _, n := range checkNodes {
		if n != nil {
			if found = n.Search(ofTypes, dontSearch); found {
				return
			}
		}
	}
	return
}

// for testing whether to wrap into scope
func NodeContainsFirstScopeLevelDeclaration(node Node) bool {
	found, atlevel := NodeSearch(node, 0,
		[]Node{&LineDeclarationNode{}, &ModeNode{}}, nil)

	if !found || atlevel > 1 {
		return false
	}
	return true
}

func NodeSearch(node Node, level int, ofTypes, stopAt []Node) (found bool, atlevel int) {
	if nodeIsOfType(node, ofTypes) {
		return true, level
	}
	if nodeIsOfType(node, stopAt) {
		return false, level
	}

	switch n := node.(type) {
	// not going into a function node here...

	case *Program:
		for _, v := range n.Statements {
			if found, atlevel = NodeSearch(v, level, ofTypes, stopAt); found {
				return
			}
		}

	case *LineDeclarationNode:
		return NodeSearch(n.Assignment, level+1, ofTypes, stopAt)

	case *AssignmentNode:
		for _, v := range n.Values {
			if found, atlevel = NodeSearch(v, level, ofTypes, stopAt); found {
				return
			}
		}

	case *ModeNode:
		return NodeSearch(n.Setting, level, ofTypes, stopAt)

	case *CallNode:
		for _, arg := range n.PositionalArgs {
			if found, atlevel = NodeSearch(arg, level, ofTypes, stopAt); found {
				return
			}
		}
		for _, arg := range n.ByNameArgs {
			if found, atlevel = NodeSearch(arg, level, ofTypes, stopAt); found {
				return
			}
		}

	case *StringNode:
		for _, v := range n.Interpolations {
			if found, atlevel = NodeSearch(v, level, ofTypes, stopAt); found {
				return
			}
		}

	case *InterpolatedNode:
		found, atlevel = NodeSearch(n.Value, level, ofTypes, stopAt)
	case *RegexNode:
		found, atlevel = NodeSearch(n.Pattern, level, ofTypes, stopAt)
	case *DateTimeNode:
		found, atlevel = NodeSearch(n.Pattern, level, ofTypes, stopAt)
	case *DurationNode:
		found, atlevel = NodeSearch(n.Pattern, level, ofTypes, stopAt)

	case *ListNode:
		for _, e := range n.Elements {
			if found, atlevel = NodeSearch(e, level, ofTypes, stopAt); found {
				return
			}
		}

	case *HashNode:
		for _, kv := range n.Pairs {
			if found, atlevel = NodeSearch(kv.Key, level, ofTypes, stopAt); found {
				return
			}
			if found, atlevel = NodeSearch(kv.Value, level, ofTypes, stopAt); found {
				return
			}
		}

	case *IndexNode:
		if found, atlevel = NodeSearch(n.Left, level, ofTypes, stopAt); found {
			return
		}
		if found, atlevel = NodeSearch(n.Index, level, ofTypes, stopAt); found {
			return
		}
		found, atlevel = NodeSearch(n.Alternate, level, ofTypes, stopAt)

	case *BlockNode:
		add := 0
		if n.HasScope {
			add = 1
		}
		for _, stmt := range n.Statements {
			if found, atlevel = NodeSearch(stmt, level+add, ofTypes, stopAt); found {
				return
			}
		}

	case *ExpressionStatementNode:
		found, atlevel = NodeSearch(n.Expression, level, ofTypes, stopAt)

	case *InfixExpressionNode:
		if found, atlevel = NodeSearch(n.Left, level, ofTypes, stopAt); found {
			return
		}
		found, atlevel = NodeSearch(n.Right, level, ofTypes, stopAt)

	case *PrefixExpressionNode:
		found, atlevel = NodeSearch(n.Right, level, ofTypes, stopAt)

	case *PostfixExpressionNode:
		found, atlevel = NodeSearch(n.Left, level, ofTypes, stopAt)

	case *ForNode:
		found, atlevel = NodeSearch(n.Body, level+1, ofTypes, stopAt)

	case *ForInOfNode:
		found, atlevel = NodeSearch(n.Body, level+1, ofTypes, stopAt)

	case *IfNode:
		for _, ta := range n.TestsAndActions {
			if found, atlevel = NodeSearch(ta.Test, level+1, ofTypes, stopAt); found {
				return true, level
			}
			if found, atlevel = NodeSearch(ta.Do, level+1, ofTypes, stopAt); found {
				return true, level
			}
		}

	case *SwitchNode:
		for _, e := range n.Expressions {
			if found, atlevel = NodeSearch(e.Expr, level+1, ofTypes, stopAt); found {
				return true, level
			}
		}
		for _, ca := range n.CasesAndActions {
			if found, atlevel = NodeSearch(ca.Do, level+1, ofTypes, stopAt); found {
				return true, level
			}
			for _, cond := range ca.MatchConditions {
				if found, atlevel = NodeSearch(cond, level+1, ofTypes, stopAt); found {
					return true, level
				}
			}
			for _, cond := range ca.OtherConditions {
				if found, atlevel = NodeSearch(cond, level+1, ofTypes, stopAt); found {
					return true, level
				}
			}
		}

	case *TryCatchNode:
		if found, atlevel = NodeSearch(n.Try, level, ofTypes, stopAt); found {
			return
		}
		if found, atlevel = NodeSearch(n.Catch, level+1, ofTypes, stopAt); found {
			return
		}
		found, atlevel = NodeSearch(n.Else, level+1, ofTypes, stopAt)

	case *ThrowNode:
		found, atlevel = NodeSearch(n.Exception, level, ofTypes, stopAt)

	case *ReturnNode:
		found, atlevel = NodeSearch(n.ReturnValue, level, ofTypes, stopAt)

	case *ExpansionNode:
		if found, atlevel = NodeSearch(n.Limits, level, ofTypes, stopAt); found {
			return
		}
		found, atlevel = NodeSearch(n.Continuation, level, ofTypes, stopAt)
		
	default:
		// a bug
	}

	return
}
