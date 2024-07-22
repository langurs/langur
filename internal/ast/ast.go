// langur/ast/ast.go

package ast

import (
	"langur/object"
	"langur/token"
)

func bug(fnName, s string) {
	panic(s)
}

var NoValue = &NullNode{}

var ExecuteMain = &CallNode{Function: &IdentNode{Name: "_main", System: true}}

// NOTE: For error reporting, we attach tokens to the nodes (besides operators).
// Besides the literal, tokens contain line numbers, etc.
type Node interface {
	TokenInfo() token.Token
	TokenRepresentation() string
	String() string
	Copy() Node
}

type PreBuilder interface {
	PreBuild() (object.Object, bool)
}

// simply marks it as a statement node (no action)
type Statement interface {
	Node
	statementNode()
}

// simply marks it as an expression node (no action)
type Expression interface {
	Node
	expressionNode()
}

func IsExpressionNode(node Node) bool {
	_, ok := node.(Expression)
	return ok
}

func copyOrNil(node Node) Node {
	if node == nil {
		return nil
	}
	return node.Copy()
}

func copyNodeSlice(nodes []Node) []Node {
	copiedNodes := []Node{}
	for _, n := range nodes {
		copiedNodes = append(copiedNodes, copyOrNil(n))
	}
	return copiedNodes
}

// NOTE: Prevent exceptions in printing node representations in the REPL by using these functions.
func tokenRepOrNil(node Node) string {
	if node == nil {
		return "<nil>"
	}
	return node.TokenRepresentation()
}
func stringOrNil(node Node) string {
	if node == nil {
		return "<nil>"
	}
	return node.String()
}

func operatorTokenString(tok token.Token) string {
	return tok.Literal + "(" + token.TypeDescription(tok.Type) + ")"
}

func Pop(nodeStack []Node) []Node {
	return nodeStack[:len(nodeStack)-1]
}
