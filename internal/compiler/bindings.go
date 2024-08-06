// langur/compiler/bindings.go

package compiler

import (
	"langur/ast"
	"langur/bytecode"
	"langur/modes"
	"langur/opcode"
	"langur/str"
	"langur/token"
)

type binding struct {
	name  string
	value ast.Node
}

// early bindings (system variables with values known at compile-time)
var early = []binding{
	// langur revision number
	{"_rev", &ast.StringNode{Values: []string{bytecode.LangurRev}}},

	{modes.RoundHashName, &ast.HashNode{Pairs: []ast.KeyValuePair{
		makeRoundModeKeyValuePair(modes.RoundHalfAwayFromZero),
		makeRoundModeKeyValuePair(modes.RoundHalfEven),
	}}},
}

func makeRoundModeKeyValuePair(mode modes.RoundingMode) ast.KeyValuePair {
	return ast.KeyValuePair{Key: &ast.StringNode{Values: []string{modes.RoundHashModeNames[mode]}},
		Value: &ast.NumberNode{Value: str.IntToStr(int(mode), 10)}}
}

// must coordinate late-binding ID's with the VM, but the order is automatically coordinated
var late = []string{"_env", "_args", "_file"}

func (c *Compiler) generateBindings(
	early []binding, late []string, varNamesParsed []string, doAllBindings bool) (

	ins opcode.Instructions, err error) {

	var temp opcode.Instructions

	// add early bindings
	for _, v := range early {
		// only add if used
		if doAllBindings || str.IsInSlice(v.name, varNamesParsed) {
			temp, err = c.compileNodeWithPopIfExprStmt(
				ast.MakeDeclarationAssignmentStatement(ast.NewVariableNode(token.Token{}, v.name, true), v.value, true, false),
			)

			if err != nil {
				return
			}
			ins = append(ins, temp...)
		}
	}

	// add late bindings
	// The last shall be first and the first shall be last (add in reverse)....
	for i := len(late) - 1; i >= 0; i-- {
		if doAllBindings || str.IsInSlice(late[i], varNamesParsed) {
			c.lateIDsUsed = append([]string{late[i]}, c.lateIDsUsed...)

			temp, err = c.compileNodeWithPopIfExprStmt(
				ast.MakeDeclarationAssignmentStatement(
					ast.NewVariableNode(token.Token{}, late[i], true), nil, true, false),
			)

			if err != nil {
				return
			}
			ins = append(ins, temp...)
		}
	}
	return
}
