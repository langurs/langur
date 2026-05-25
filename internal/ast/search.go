// langur/ast/search.go

package ast

import (
	"reflect"
)

type searchCriteria struct{
	OfTypes		[]Node
	DontSearch	[]Node
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

func nodeIsOfType(node Node, ofTypes []Node) bool {
	for _, n := range ofTypes {
		if reflect.TypeOf(node) == reflect.TypeOf(n) {
			return true
		}
	}
	return false
}

func CopyBeforeAssignment(val Node) bool {
	// copy value before assignment?
	// inefficient to always copy

	// FIXME: a temporary patch; inefficient
	return true

	// return val.Search(copyBeforeAssignment_searchCritera)
}
var copyBeforeAssignment_searchCritera = &searchCriteria{
	OfTypes: []Node{&IdentNode{}},
	DontSearch: []Node{
		&LineDeclarationNode{}, &AssignmentNode{}, &ModeNode{},
		&CallNode{}, &FunctionNode{},
		&StringNode{}, &RegexNode{}, &DateTimeNode{}, &DurationNode{},
	},
}

// for testing whether to wrap into scope
func NodeContainsFirstScopeLevelDeclaration(node Node) bool {
	return node.Search(firstScopeLevelDeclaration_searchCriteria)
}
var firstScopeLevelDeclaration_searchCriteria = &searchCriteria{
	OfTypes: []Node{&LineDeclarationNode{}},
	DontSearch: []Node{
		&FunctionNode{},
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
