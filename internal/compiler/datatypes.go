// langur/compiler/datatypes.go

package compiler

import (
	"langur/ast"
	"langur/object"
	"langur/opcode"
	"langur/token"
)

func (c *Compiler) compileBooleanNode(node *ast.BooleanNode) (ins opcode.Instructions, err error) {
	if node.Value {
		ins = opcode.Make(opcode.OpTrue)
	} else {
		ins = opcode.Make(opcode.OpFalse)
	}
	return
}
func (c *Compiler) compileNullNode(node *ast.NullNode) (ins opcode.Instructions, err error) {
	ins = opcode.Make(opcode.OpNull)
	return
}
func (c *Compiler) compileNoneNode(node *ast.NoneNode) (ins opcode.Instructions, err error) {
	if node.Token.Literal == "_" {
		// must be interpreted by context
		err = c.makeErr(c.lastNode, "Underscore no-op literal not dealt with in this context")
		return
	}
	// no-op by keyword...
	ins = c.constantIns(object.NONE)
	return
}

func (c *Compiler) compileNumberNode(node *ast.NumberNode) (ins opcode.Instructions, err error) {
	if c.Modes.WarnOnIntegerLiteralsStartingWithZero {
		if node.Token.Type == token.INT && len(node.Value) > 1 &&
			node.Token.Code2 == token.CODE_DEFAULT && node.Value[0] == '0' {
			err = c.makeWarning(node, "Integer literal starting with zero (might be confused for a base 8 number (as used in other languages)")
			return
		}
	}

	var number *object.Number
	number, err = object.NumberFromStringBase(node.Value, node.Base)
	ins = c.constantIns(number)
	return
}

func (c *Compiler) compileListNode(node *ast.ListNode) (ins opcode.Instructions, err error) {
	if len(node.Elements) == 0 {
		// no elements; return empty list constant
		ins = c.constantIns(object.EmptyList)
		return
	}

	var b []byte
	for _, e := range node.Elements {
		if ast.IsNoOp(e) {
			b = c.constantIns(object.NONE)

		} else {
			b, err = c.compileNode(e)
			if err != nil {
				return
			}
		}
		ins = append(ins, b...)
	}
	ins = append(ins, opcode.Make(opcode.OpList, len(node.Elements))...)
	return
}

func (c *Compiler) compileHashNode(node *ast.HashNode) (ins opcode.Instructions, err error) {
	if len(node.Pairs) == 0 {
		// no entries; return empty hash constant
		ins = c.constantIns(object.EmptyHash)
		return
	}

	var b []byte
	for _, kv := range node.Pairs {
		b, err = c.compileNode(kv.Key)
		if err != nil {
			return
		}
		ins = append(ins, b...)

		b, err = c.compileNode(kv.Value)
		if err != nil {
			return
		}
		ins = append(ins, b...)
	}
	ins = append(ins, opcode.Make(opcode.OpHash, len(node.Pairs)*2)...)
	return
}
