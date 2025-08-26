// langur/ast/compiler_identifiers.go

package ast

import (
	"fmt"
	"langur/opcode"
	"langur/symbol"
)

func (c *Compiler) addToImpureEffectsList(s string) {
	c.symbolTable.AddImpureEffects(s)
	c.impureEffects = true
}

func (c *Compiler) makeOpGetInstructions(node Node, sym symbol.Symbol, level int) (
	pkg opcode.InsPackage, err error) {

	switch sym.Scope {
	case symbol.GlobalScope:
		return opcode.MakePkgWithErrTest(node.TokenInfo(), opcode.OpGetGlobal, sym.Index)

	case symbol.LocalScope:
		if level == 0 {
			return opcode.MakePkgWithErrTest(node.TokenInfo(), opcode.OpGetLocal, sym.Index)
		}
		return opcode.MakePkgWithErrTest(node.TokenInfo(), opcode.OpGetNonLocal, sym.Index, level)

	case symbol.FreeScope:
		return opcode.MakePkgWithErrTest(node.TokenInfo(), opcode.OpGetFree, sym.Index)

	case symbol.SelfScope:
		return c.compileSelfRef(node)

		// return node.Compile(c)
	}
	err = c.makeErr(node, fmt.Sprintf("Attempt to create OpGet instructions on %s for scope %s not accounted for", sym.Name, sym.Scope))
	bug("makeOpGetInstructions", err.Error())
	return
}

func (c *Compiler) compileSelfRef(node Node) (pkg opcode.InsPackage, err error) {
	if c.symbolTable.Outer == nil {
		err = c.makeErr(node, "Cannot use self token in global scope")
		return
	}
	return opcode.MakePkgWithErrTest(node.TokenInfo(), opcode.OpGetSelf)
}

func (c *Compiler) resolveAndGetInstructions(node Node, name string) (ins opcode.InsPackage, err error) {
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
