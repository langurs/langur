// langur/ast/compiler.go

package ast

import (
	"langur/token"
	"fmt"
	"langur/object"
	"langur/modes"
	"langur/symbol"
	"langur/opcode"
	"langur/bytecode"
)

// to pass from compiler to VM
func (c *Compiler) ByteCode() *bytecode.ByteCode {
	return &bytecode.ByteCode{
		StartCode: &object.CompiledCode{
			InsPackage:         c.InsPackage,
			LocalBindingsCount: c.symbolTable.DefinitionCount,
		},

		Constants: c.constants,
		Late:      c.lateIDsUsed,
	}
}

type Compiler struct {
	InsPackage opcode.InsPackage

	constants     []object.Object
	symbolTable   *symbol.SymbolTable
	lateIDs       []string
	lateIDsUsed   []string
	Modes         *modes.CompileModes
	doAllBindings bool

	// compile once and reuse...
	noValueIns opcode.InsPackage

	// should be 0 at the end...
	breakStmtCount       int
	nextStmtCount        int
	fallthroughStmtCount int

	loopVarStack []Node

	functionLevel int

	moduleDeclaredImpureEffects bool
	impureEffects               bool
}

func NewCompiler(m *modes.CompileModes, doAllBindings bool) (compiler *Compiler, err error) {
	defer func() {
		if p := recover(); p != nil {
			err = object.PanicToError(p)
		}
	}()

	compiler = &Compiler{
		InsPackage:    opcode.InsPackage{},
		constants:     []object.Object{},
		lateIDs:       late,
		doAllBindings: doAllBindings,
	}
	if m == nil {
		compiler.Modes = modes.NewCompileModes()
	} else {
		compiler.Modes = m.Copy()
	}
	compiler.symbolTable = symbol.NewSymbolTable(nil, compiler.Modes)

	compiler.noValueIns, err = NoValue.Compile(compiler)
	return
}

func NewCompilerWithState(
	s *symbol.SymbolTable, constants []object.Object, m *modes.CompileModes) (
	compiler *Compiler, err error) {

	compiler, err = NewCompiler(m, false)
	compiler.symbolTable = s
	compiler.constants = constants
	return
}

func (c *Compiler) makeErr(node Node, err string) error {
	tok := node.TokenInfo()
	if node == nil {
		return fmt.Errorf("%s", err)
	}
	return fmt.Errorf("[%s] %s", tok.Where.String(), err)
}

func (c *Compiler) makeWarning(node Node, err string) error {
	return c.makeErr(node, "warning: "+err)
}

func (c *Compiler) constantIns(obj object.Object) opcode.InsPackage {
	// We do this so often that it seems best to make a small function for it.
	return opcode.MakePkg(token.Token{}, opcode.OpConstant, c.addConstant(obj))
}

func (c *Compiler) addConstant(obj object.Object) int {
	// add constants once
	idx := constantsSliceIndex(obj, c.constants)
	if idx == -1 {
		idx = len(c.constants)
		c.constants = append(c.constants, obj)
	}
	return idx
}

func constantsSliceIndex(obj object.Object, objSlc []object.Object) int {
	for i := range objSlc {
		if sameConstant(obj, objSlc[i]) {
			return i
		}
	}
	return -1
}

func sameConstant(obj1, obj2 object.Object) bool {
	// First, compare pointers....
	if obj1 == obj2 {
		// That was easy.
		return true

	} else if obj1.Type() == object.COMPILED_CODE_OBJ && obj2.Type() == object.COMPILED_CODE_OBJ {
		// Note that sometimes compiled code constants are modified by the compiler to fix jumps ...
		// ... AFTER being added, so we'll not call them equal even if they have the same bytecodes.
		// TODO: maybe check compiled code for placeholders and compare if not present
		return false
	}

	return object.Equal(obj1, obj2)
}

func (c *Compiler) compileNodeWithPopIfExprStmt(node Node) (ins opcode.InsPackage, err error) {
	_, isExprStmt := node.(*ExpressionStatementNode)
	ins, err = node.Compile(c)
	if isExprStmt {
		ins = ins.Append(opcode.MakePkg(node.TokenInfo(), opcode.OpPop))
	}
	return
}

func (c *Compiler) wrapInstructions(pkg opcode.InsPackage) int {
	// NOTE: Call this before c.popVariableScope().
	compiled := &object.CompiledCode{
		InsPackage:         pkg,
		LocalBindingsCount: c.symbolTable.DefinitionCount,
	}
	return c.addConstant(compiled)
}
func (c *Compiler) wrapInstructionsWithExecute(pkg opcode.InsPackage, tok token.Token) opcode.InsPackage {
	// NOTE: Call this before c.popVariableScope().
	index := c.wrapInstructions(pkg)
	return opcode.MakePkg(tok, opcode.OpExecute, index)
}
