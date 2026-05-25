// langur/ast/search.go

package ast

import (
	"reflect"
)

func nodeMatching(node Node, mc *matchCriteria) bool {
	for _, m := range mc.Types {
		if reflect.TypeOf(node) == reflect.TypeOf(m) {
			switch n := node.(type) {
			case *BlockNode:
				if mc.Match_BlockNode_HasScope {
					return n.HasScope == m.(*BlockNode).HasScope
				}
			}

			return true
		}
	}
	return false
}

type matchCriteria struct{
	Types []Node
	Match_BlockNode_HasScope bool
}

type searchCriteria struct{
	OfTypes    *matchCriteria
	DontSearch *matchCriteria
}

func (sc *searchCriteria) searchNodes(checkNodes ...Node) (found bool) {
	return sc.searchNodeSlice(checkNodes)
}
func (sc *searchCriteria) searchNodeSlice(checkNodes []Node) (found bool) {
	for _, n := range checkNodes {
		if n != nil {
			if found = n.Search(sc); found {
				return
			}
		}
	}
	return
}

func CopyBeforeAssignment(val Node) bool {
	// copy value before assignment?
	// inefficient to always copy

	// FIXME: a temporary patch; inefficient
	return true

	// return val.Search(copyBeforeAssignment_searchCritera)
}
var copyBeforeAssignment_searchCritera = &searchCriteria{
	OfTypes: &matchCriteria{Types: []Node{&IdentNode{}}},
	DontSearch: &matchCriteria{
		Types: []Node{
			&DeclarationNode{}, &AssignmentNode{}, &ModeNode{},
			&CallNode{}, &FunctionNode{},
			&StringNode{}, &RegexNode{}, &DateTimeNode{}, &DurationNode{},
		},
	},
}

// for testing whether to wrap into scope
func NodeContainsFirstScopeLevelDeclaration(node Node) bool {
	return node.Search(firstScopeLevelDeclaration_searchCriteria)
}
var firstScopeLevelDeclaration_searchCriteria = &searchCriteria{
	OfTypes: &matchCriteria{Types: []Node{&DeclarationNode{}}},
	DontSearch: &matchCriteria{
		Types: []Node{
			&FunctionNode{}, &BlockNode{HasScope: true},
			&IfNode{}, &SwitchNode{}, &ForNode{}, &ForInOfNode{},
		},
		Match_BlockNode_HasScope: true,
	},
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
