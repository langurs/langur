// langur/compiler/identifiers.go

package compiler

import (
	"fmt"
	"langur/ast"
	"langur/opcode"
	"langur/symbol"
	"langur/vm/process"
)

func (c *Compiler) compileIdentifierNode(node *ast.IdentNode) (ins opcode.Instructions, err error) {
	if node.Type != nil {
		return nil, c.makeErr(node, "This revision of langur not able to accept explicit variable type")
	}

	bi := process.GetBuiltInByName(node.Name)
	if bi == nil {
		// not a built-in; must be a variable
		return c.resolveAndGetInstructions(node, node.Name)
	}

	if process.GetBuiltInImpurityStatus(node.Name) {
		c.addToImpureEffectsList(node.Name)
	}

	ins = c.constantIns(bi)
	return
}

func (c *Compiler) addToImpureEffectsList(s string) {
	c.symbolTable.AddImpureEffects(s)
	c.impureEffects = true
}

func (c *Compiler) makeOpGetInstructions(node ast.Node, sym symbol.Symbol, level int) (
	ins opcode.Instructions, err error) {

	switch sym.Scope {
	case symbol.GlobalScope:
		return opcode.MakeWithErrTest(opcode.OpGetGlobal, sym.Index)

	case symbol.LocalScope:
		if level == 0 {
			return opcode.MakeWithErrTest(opcode.OpGetLocal, sym.Index)
		}
		return opcode.MakeWithErrTest(opcode.OpGetNonLocal, sym.Index, level)

	case symbol.FreeScope:
		return opcode.MakeWithErrTest(opcode.OpGetFree, sym.Index)

	case symbol.SelfScope:
		return c.compileSelfRef(node)
	}
	err = c.makeErr(node, fmt.Sprintf("Attempt to create OpGet instructions on %s for scope %s not accounted for", sym.Name, sym.Scope))
	bug("makeOpGetInstructions", err.Error())
	return nil, err
}

func (c *Compiler) resolveAndGetInstructions(node ast.Node, name string) (ins opcode.Instructions, err error) {
	sym, level, ok := c.symbolTable.Resolve(name)
	if !ok {
		err = c.makeErr(node, fmt.Sprintf("Could not resolve variable in scope: %s", name))
		return
	}
	return c.makeOpGetInstructions(node, sym, level)
}

func (c *Compiler) pushVariableScope() {
	c.symbolTable = symbol.NewSymbolTable(c.symbolTable, c.Modes)
}
func (c *Compiler) pushNonScope() {
	// a placeholder table for a VM frame with no scope
	c.symbolTable = symbol.NewSymbolTable(c.symbolTable, c.Modes)
	c.symbolTable.IsNonScope = true
}
func (c *Compiler) pushVariableScopeWithTable(st *symbol.SymbolTable) {
	c.symbolTable = symbol.ReuseSymbolTable(c.symbolTable, st)
}
func (c *Compiler) popVariableScope() {
	c.symbolTable = c.symbolTable.Outer
}
