// langur/ast/search.go

package ast

import (
	"langur/common"
	"reflect"
)

func nodeMatching(node, parent Node, mc *matchCriteria) bool {
	if mc != nil {
		for _, m := range mc.Types {
			if reflect.TypeOf(node) == reflect.TypeOf(m) {
				// check special cases
				switch n := node.(type) {
				case *BlockNode:
					if mc.Match_BlockNode_HasScope {
						return n.HasScope == m.(*BlockNode).HasScope
					}

				case *ModeNode:
					if mc.Match_ModeNode_NotGlobal {
						switch p := parent.(type) {
						case *Program:
							return false
						case *FunctionNode:
							if p.Name == common.MainFnName {
								return false
							}
						}
					}
				}

				// no special case matches; return true
				return true
			}
		}
	}
	return false
}

type matchCriteria struct{
	Types []Node
	Match_BlockNode_HasScope bool
	Match_ModeNode_NotGlobal bool
}

type searchCriteria struct{
	OfTypes    *matchCriteria
	DontSearch *matchCriteria
}

func (sc *searchCriteria) searchNodes(parent Node, checkNodes ...Node) (found bool) {
	return sc.searchNodeSlice(parent, checkNodes)
}
func (sc *searchCriteria) searchNodeSlice(parent Node, checkNodes []Node) (found bool) {
	for _, n := range checkNodes {
		if n != nil {
			if found = n.Search(parent, sc); found {
				return
			}
		}
	}
	return
}

// func CopyBeforeAssignment(val Node) bool {
// 	// copy value before assignment?
// 	// inefficient to always copy
// 	return val.Search(nil, copyBeforeAssignment_searchCritera)
// }
// var copyBeforeAssignment_searchCritera = &searchCriteria{
// 	OfTypes: &matchCriteria{Types: []Node{&IdentNode{}}},
// 	DontSearch: &matchCriteria{
// 		Types: []Node{
// 			&DeclarationNode{}, &AssignmentNode{}, &ModeNode{},
// 			&CallNode{}, &FunctionNode{},
// 			&StringNode{}, &RegexNode{}, &DateTimeNode{}, &DurationNode{},
// 		},
// 	},
// }

// for testing whether to wrap into scope
func NodeContainsFirstScopeLevelDeclaration(node Node) bool {
	return node.Search(nil, firstScopeLevelDeclaration_searchCriteria)
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

var modeNotGlobal_searchCriteria = &searchCriteria{
	OfTypes: &matchCriteria{
		Types: []Node{&ModeNode{}},
		Match_ModeNode_NotGlobal: true,
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
