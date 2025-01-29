// langur/compiler/compiler.go

package compiler

import (
	"fmt"
	"langur/ast"
	"langur/bytecode"
	"langur/modes"
	"langur/object"
	"langur/opcode"
	"langur/symbol"
)

func bug(fnName, s string) {
	panic("Compiler Bug: " + s)
}

func (c *Compiler) makeErr(node ast.Node, err string) error {
	tok := node.TokenInfo()
	if node == nil {
		return fmt.Errorf("%s", err)
	}
	return fmt.Errorf("[%s] %s", tok.Where.String(), err)
}

func (c *Compiler) makeWarning(node ast.Node, err string) error {
	return c.makeErr(node, "warning: "+err)
}

type Compiler struct {
	InsPackage opcode.InsPackage

	constants   []object.Object
	symbolTable *symbol.SymbolTable
	lateIDs     []string
	lateIDsUsed []string
	lastNode    ast.Node
	Modes       *modes.CompileModes

	// compile once and reuse...
	noValueIns opcode.Instructions

	// should be 0 at the end...
	breakStmtCount       int
	nextStmtCount        int
	fallthroughStmtCount int

	loopVarStack []ast.Node

	functionLevel int

	moduleDeclaredImpureEffects bool
	impureEffects               bool
}

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

func New(m *modes.CompileModes) (compiler *Compiler, err error) {
	defer func() {
		if panik := recover(); panik != nil {
			err = object.PanicToError(panik)
		}
	}()

	compiler = &Compiler{
		InsPackage: opcode.InsPackage{},
		constants:  []object.Object{},
		lateIDs:    late,
	}
	if m == nil {
		compiler.Modes = modes.NewCompileModes()
	} else {
		compiler.Modes = m.Copy()
	}
	compiler.symbolTable = symbol.NewSymbolTable(nil, compiler.Modes)

	compiler.noValueIns, err = compiler.compileNode(ast.NoValue)
	return
}

func NewWithState(s *symbol.SymbolTable, constants []object.Object, m *modes.CompileModes) (
	compiler *Compiler, err error) {
	compiler, err = New(m)
	compiler.symbolTable = s
	compiler.constants = constants
	return
}

func (c *Compiler) constantIns(obj object.Object) opcode.Instructions {
	// We do this so often that it seems best to make a small function for it.
	return opcode.Make(opcode.OpConstant, c.addConstant(obj))
}

func (c *Compiler) Compile(node *ast.Program, doAllBindings bool) (err error) {
	defer func() {
		if panik := recover(); panik != nil {
			err = object.PanicToError(panik)
		}
	}()

	var ins opcode.Instructions

	ins, err = c.generateBindings(early, c.lateIDs, node.VarNamesUsed, doAllBindings)
	if err != nil {
		return
	}
	c.InsPackage.Instructions = append(c.InsPackage.Instructions, ins...)

	ins, err = c.compileProgram(node, true)
	c.InsPackage.Instructions = append(c.InsPackage.Instructions, ins...)

	if err == nil {
		err = c.checkStatementCounts()
	}

	return
}

// helps with the REPL not to try to set early/late bindings every time
// also for running compiler tests, so we don't get extra opcodes
func (c *Compiler) CompileAnother(node *ast.Program) (err error) {
	defer func() {
		if panik := recover(); panik != nil {
			err = object.PanicToError(panik)
		}
	}()

	c.InsPackage.Instructions, err = c.compileProgram(node, true)

	if err == nil {
		err = c.checkStatementCounts()
	}

	return
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

func (c *Compiler) compileVmMode(node *ast.ModeNode) (ins opcode.Instructions, err error) {
	if c.symbolTable.Outer != nil {
		err = c.makeErr(node, "Current implementation can only set modes in global context")
		return
		// The idea is that modes will have scope like variables. ...
		// ... Therefore, if set, they have to be reset when exiting scope.
	}
	ins, err = c.compileNode(node.Setting)
	if err != nil {
		return
	}
	code, ok := modes.ModeNames[node.Name]
	if !ok {
		err = c.makeErr(node, fmt.Sprintf("Unknown mode setting %s", node.Name))
		return
	}
	ins = append(ins, opcode.Make(opcode.OpMode, code)...)
	return
}

func (c *Compiler) compileNodeWithPopIfExprStmt(node ast.Node) (ins opcode.Instructions, err error) {
	_, isExprStmt := node.(*ast.ExpressionStatementNode)
	ins, err = c.compileNode(node)
	if isExprStmt {
		ins = append(ins, opcode.Make(opcode.OpPop)...)
	}
	return
}

func (c *Compiler) compileNode(node ast.Node) (ins opcode.Instructions, err error) {
	switch node := node.(type) {
	case *ast.BlockNode:
		ins, err = c.compileBlock(node, true)

	case *ast.ExpressionStatementNode:
		ins, err = c.compileNode(node.Expression)

	case *ast.PrefixExpressionNode:
		ins, err = c.compilePrefixExpression(node)

	case *ast.InfixExpressionNode:
		ins, err = c.compileInfixExpression(node)

	case *ast.IfNode:
		ins, err = c.compileIfExpression(node)

	case *ast.ForNode:
		ins, err = c.compileFor(node)

	case *ast.BreakNode:
		ins, err = c.compileBreak(node)
	case *ast.NextNode:
		ins, err = c.compileNext(node)
	case *ast.FallThroughNode:
		ins, err = c.compileFallthrough(node)

	case *ast.LineDeclarationNode:
		ins, err = c.compileDeclarationAndAssignments(node)

	case *ast.AssignmentNode:
		ins, err = c.compileAssignment(node)

	case *ast.IdentNode:
		ins, err = c.compileIdentifierNode(node)

	case *ast.StringNode:
		ins, err = c.compileStringNode(node)

	case *ast.RegexNode:
		ins, err = c.compileRegexNode(node)

	case *ast.DateTimeNode:
		ins, err = c.compileDateTimeNode(node)

	case *ast.DurationNode:
		ins, err = c.compileDurationNode(node)

	case *ast.NumberNode:
		if node.Imaginary {
			// stand-alone imaginary number compiled to complex
			ins, err = c.compileComplexNumber(nil, node, false)
		} else {
			ins, err = c.compileNumberNode(node)
		}

	case *ast.BooleanNode:
		ins, err = c.compileBooleanNode(node)
	case *ast.NullNode:
		ins, err = c.compileNullNode(node)

	case *ast.ListNode:
		ins, err = c.compileListNode(node)
	case *ast.HashNode:
		ins, err = c.compileHashNode(node)
	case *ast.IndexNode:
		ins, err = c.compileIndexNode(node)

	case *ast.FunctionNode:
		ins, err = c.compileFunctionNode(node)

	case *ast.ReturnNode:
		ins, err = c.compileReturnNode(node)

	case *ast.CallNode:
		ins, err = c.compileCallNode(node)

	case *ast.SelfNode:
		ins, err = c.compileSelfRef(node)

	case *ast.TryCatchNode:
		ins, err = c.compileTryCatch(node)

	case *ast.ThrowNode:
		ins, err = c.compileThrow(node)

	case *ast.ModeNode:
		ins, err = c.compileVmMode(node)

	case *ast.NoneNode:
		ins, err = c.compileNoneNode(node)

	case nil:
		err = c.makeErr(c.lastNode, "Nil node")
		// bug("compileNode", "Nil node")
		return

	default:
		//bug(fmt.Sprintf("Node type %T not accounted for", node))
		exprNode, ok := c.lastNode.(*ast.ExpressionStatementNode)
		if ok {
			err = c.makeErr(node, fmt.Sprintf("Node type %T not accounted for in this context (possible parsing error or compiler incomplete; last node type %T)", node, exprNode.Expression))
			// bug("compileNode", fmt.Sprintf("Node type %T not accounted for (parsing error or compiler incomplete; last node type %T)", node, exprNode.Expression))
		} else {
			err = c.makeErr(node, fmt.Sprintf("Node type %T not accounted for in this context (possible parsing error or compiler incomplete; last node type %T)", node, c.lastNode))
			// bug("compileNode", fmt.Sprintf("Node type %T not accounted for (parsing error or compiler incomplete; last node type %T)", node, c.lastNode))
		}
		return
	}

	c.lastNode = node
	return
}
