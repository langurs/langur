// langur/ast/nodes.go

package ast

import (
	"bytes"
	"fmt"
	"langur/common"
	"langur/object"
	"langur/regex"
	"langur/str"
	"langur/symbol"
	"langur/token"
	"langur/modes"
	"langur/opcode"
	"langur/vm/process"
	"strings"
)

// NOTE: If adding node types, you may need to add them to functions in ast/search.go.

func cannotDirectlyCompile(s string) (opcode.InsPackage, error) {
	return opcode.InsPackage{}, fmt.Errorf("Cannot directly compile node of type %s", s)
}

// THE BASE OF THE TREE
type Program struct {
	Token        token.Token
	Statements   []Node
	VarNamesUsed []string
}

func (p *Program) Copy() Node {
	return &Program{Token: p.Token.Copy(), Statements: CopyNodeSlice(p.Statements)}
}

func (p *Program) Evaluate() object.Object {
	return nil
}

func (node *Program) Compile(c *Compiler) (pkg opcode.InsPackage, err error) {
	defer func() {
		if p := recover(); p != nil {
			err = object.PanicToError(p)
		}
	}()

	var temp opcode.InsPackage

	temp, err = c.generateBindings(early, c.lateIDs, node.VarNamesUsed, c.doAllBindings)
	if err != nil {
		return
	}
	c.InsPackage = c.InsPackage.Append(temp)

	temp, err = c.compileProgram(node, true)
	c.InsPackage = c.InsPackage.Append(temp)

	if err == nil {
		err = c.checkStatementCounts()
	}

	return
}

// helps with the REPL not to try to set early/late bindings every time
// also for running compiler tests, so we don't get extra opcodes
func (node *Program) CompileAnother(c *Compiler) (pkg opcode.InsPackage, err error) {
	defer func() {
		if p := recover(); p != nil {
			err = object.PanicToError(p)
		}
	}()

	c.InsPackage, err = c.compileProgram(node, true)

	if err == nil {
		err = c.checkStatementCounts()
	}

	return	
}

func (p *Program) String() string {
	var out bytes.Buffer

	for i, s := range p.Statements {
		out.WriteString(fmt.Sprintf("%d: ", i+1) + stringOrNil(s) + "\n")
	}

	return out.String()
}

func (p *Program) TokenRepresentation() string {
	var out bytes.Buffer

	for i, s := range p.Statements {
		out.WriteString(tokenRepOrNil(s))
		if i < len(p.Statements)-1 {
			out.WriteString("; ")
		}
	}

	return out.String()
}

func (p *Program) TokenInfo() token.Token {
	return p.Token
}

// MODULE STATEMENT
type ModuleNode struct {
	Token         token.Token
	Name          string
	ImpureEffects bool
}

func (m *ModuleNode) statementNode() {}

func (m *ModuleNode) Copy() Node {
	return m
}

func (m *ModuleNode) Evaluate() object.Object {
	return nil
}

func (m *ModuleNode) Compile(c *Compiler) (opcode.InsPackage, error) {
	return cannotDirectlyCompile("ModuleNode")
}

func (m *ModuleNode) TokenRepresentation() string {
	var sb strings.Builder

	sb.WriteString("module")
	if m.ImpureEffects {
		sb.WriteRune('*')
	}

	if m.Name != "" {
		sb.WriteRune(' ')
		sb.WriteString(m.Name)
	}

	return sb.String()
}

func (m *ModuleNode) String() string {
	var sb strings.Builder

	if m.ImpureEffects {
		sb.WriteString("Impure ")
	}
	sb.WriteString("Module")

	if m.Name != "" {
		sb.WriteRune(' ')
		sb.WriteString(m.Name)
	}

	return sb.String()
}

func (m *ModuleNode) TokenInfo() token.Token {
	return m.Token
}

// IMPORT STATEMENT
type ImportAs struct {
	Import string
	As     string // zero-length string if not applicable
}

func (ia *ImportAs) Copy() *ImportAs {
	return &ImportAs{Import: ia.Import, As: ia.As}
}

type ImportNode struct {
	Token   token.Token
	Modules []ImportAs
}

func (i *ImportNode) statementNode() {}

func (i *ImportNode) Copy() Node {
	var modules []ImportAs
	if i.Modules != nil {
		modules = make([]ImportAs, len(i.Modules))
		for x := range i.Modules {
			modules = append(modules, *(i.Modules[x].Copy()))
		}
	}
	return &ImportNode{
		Token:   i.Token.Copy(),
		Modules: modules,
	}
}

func (i *ImportNode) Evaluate() object.Object {
	return nil
}

func (i *ImportNode) Compile(c *Compiler) (opcode.InsPackage, error) {
	return cannotDirectlyCompile("ImportNode")
}

func (i *ImportNode) TokenRepresentation() string {
	var out bytes.Buffer
	out.WriteString("import ")

	if i.Modules == nil {
		out.WriteString("INVALID")

	} else {
		for x := range i.Modules {
			if x != 0 {
				out.WriteString(", ")
			}

			out.WriteString(i.Modules[x].Import)

			if i.Modules[x].As != "" {
				out.WriteString(" as ")
				out.WriteString(i.Modules[x].As)
			}
		}
	}

	return out.String()
}

func (i *ImportNode) String() string {
	var out bytes.Buffer
	out.WriteString("Import ")

	if i.Modules == nil {
		out.WriteString("INVALID")

	} else {
		for x := range i.Modules {
			if x != 0 {
				out.WriteString(", ")
			}

			out.WriteString(i.Modules[x].Import)

			if i.Modules[x].As != "" {
				out.WriteString(" As ")
				out.WriteString(i.Modules[x].As)
			}
		}
	}

	return out.String()
}

func (i *ImportNode) TokenInfo() token.Token {
	return i.Token
}

// RETURN STATEMENT
type ReturnNode struct {
	Token       token.Token
	ReturnValue Node
}

func (r *ReturnNode) statementNode() {}

func (r *ReturnNode) Copy() Node {
	return &ReturnNode{Token: r.Token.Copy(), ReturnValue: copyOrNil(r.ReturnValue)}
}

func (r *ReturnNode) Evaluate() object.Object {
	return nil
}

func (r *ReturnNode) Compile(c *Compiler) (pkg opcode.InsPackage, err error) {
	if c.functionLevel == 0 {
		err = c.makeErr(r, "Cannot use return outside of function")
		return
	}
	pkg, err = r.ReturnValue.Compile(c)
	if err != nil {
		return
	}
	pkg = pkg.Append(opcode.MakePkg(r.Token, opcode.OpReturnValue))
	return	
}

func (r *ReturnNode) TokenRepresentation() string {
	return "return " + tokenRepOrNil(r.ReturnValue)
}

func (r *ReturnNode) String() string {
	var out bytes.Buffer

	out.WriteString("Return ")

	if r.ReturnValue != nil {
		out.WriteString(r.ReturnValue.String())
	}

	return out.String()
}

func (r *ReturnNode) TokenInfo() token.Token {
	return r.Token
}

// LINE DECLARATION EXPRESSION
type LineDeclarationNode struct {
	Token      token.Token
	Assignment Node // assignment or variable
	Mutable    bool
	Public     bool
}

func (d *LineDeclarationNode) expressionNode() {}

func (d *LineDeclarationNode) Copy() Node {
	return &LineDeclarationNode{
		Token:      d.Token.Copy(),
		Assignment: copyOrNil(d.Assignment),
		Mutable:    d.Mutable,
		Public:     d.Public,
	}
}

func (d *LineDeclarationNode) Evaluate() object.Object {
	return nil
}

func (d *LineDeclarationNode) Compile(c *Compiler) (pkg opcode.InsPackage, err error) {
	return c.compileDeclarationAndAssignments(d)
}

func (d *LineDeclarationNode) TokenRepresentation() string {
	var out bytes.Buffer

	if d.Public {
		out.WriteString("public ")
	}

	if d.Mutable {
		out.WriteString("var ")
	} else {
		out.WriteString("val ")
	}
	out.WriteString(tokenRepOrNil(d.Assignment))

	return out.String()
}
func (d *LineDeclarationNode) String() string {
	var out bytes.Buffer

	if d.Public {
		out.WriteString("Public ")
	}

	if d.Mutable {
		out.WriteString("Var ")
	} else {
		out.WriteString("Val ")
	}
	out.WriteString(stringOrNil(d.Assignment))

	return out.String()
}

func (d *LineDeclarationNode) TokenInfo() token.Token {
	return d.Token
}

// ASSIGNMENT EXPRESSION
type AssignmentNode struct {
	Token            token.Token
	Identifiers      []Node
	Values           []Node
	SystemAssignment bool
}

func (a *AssignmentNode) expressionNode() {}

func (a *AssignmentNode) Copy() Node {
	return &AssignmentNode{
		Token:            a.Token.Copy(),
		Identifiers:      CopyNodeSlice(a.Identifiers),
		Values:           CopyNodeSlice(a.Values),
		SystemAssignment: a.SystemAssignment,
	}
}

func (a *AssignmentNode) Evaluate() object.Object {
	return nil
}

func (a *AssignmentNode) Compile(c *Compiler) (pkg opcode.InsPackage, err error) {
	return c.compileAssignment(a)
}

func (a *AssignmentNode) TokenRepresentation() string {
	var out bytes.Buffer

	if len(a.Identifiers) == 1 {
		out.WriteString(tokenRepOrNil(a.Identifiers[0]))
	} else {
		for i, id := range a.Identifiers {
			out.WriteString(tokenRepOrNil(id))
			if i < len(a.Identifiers)-1 {
				out.WriteString(", ")
			}
		}
	}

	if a.SystemAssignment {
		out.WriteString(" (sys)= ")
	} else {
		out.WriteString(" = ")
	}

	if len(a.Values) == 1 {
		out.WriteString(tokenRepOrNil(a.Values[0]))
	} else {
		for i, val := range a.Values {
			out.WriteString(tokenRepOrNil(val))
			if i < len(a.Values)-1 {
				out.WriteString(", ")
			}
		}
	}

	return out.String()
}
func (a *AssignmentNode) String() string {
	var out bytes.Buffer

	if a.SystemAssignment {
		out.WriteString("SysAssign ")
	} else {
		out.WriteString("Assign ")
	}

	if len(a.Identifiers) == 1 {
		out.WriteString(stringOrNil(a.Identifiers[0]))
	} else {
		for i, id := range a.Identifiers {
			out.WriteString(stringOrNil(id))
			if i < len(a.Identifiers)-1 {
				out.WriteString(", ")
			}
		}
	}

	out.WriteString(" = ")

	if len(a.Values) == 1 {
		out.WriteString(stringOrNil(a.Values[0]))
	} else {
		for i, val := range a.Values {
			out.WriteString(stringOrNil(val))
			if i < len(a.Values)-1 {
				out.WriteString(", ")
			}
		}
	}

	return out.String()
}

func (a *AssignmentNode) TokenInfo() token.Token {
	return a.Token
}

// EXPRESSION STATEMENT
type ExpressionStatementNode struct {
	Token      token.Token
	Expression Node
}

func (es *ExpressionStatementNode) statementNode() {}

func (es *ExpressionStatementNode) Copy() Node {
	return &ExpressionStatementNode{
		Token: es.Token.Copy(), Expression: copyOrNil(es.Expression)}
}

func (es *ExpressionStatementNode) Evaluate() object.Object {
	return es.Evaluate()
}

func (es *ExpressionStatementNode) Compile(c *Compiler) (pkg opcode.InsPackage, err error) {
	return es.Expression.Compile(c)
}

func (es *ExpressionStatementNode) TokenRepresentation() string {
	return tokenRepOrNil(es.Expression)
}

func (es *ExpressionStatementNode) String() string {
	var out bytes.Buffer

	out.WriteString("ExprStmt (")
	if es.Expression != nil {
		out.WriteString(stringOrNil(es.Expression))
	}
	out.WriteString(")")

	return out.String()
}

func (es *ExpressionStatementNode) TokenInfo() token.Token {
	return es.Token
}

// FUNCTION CALLS
type CallNode struct {
	Token          token.Token
	Function       Node // Identifier or Function Literal
	PositionalArgs []Node
	ByNameArgs     []Node
}

func (fc *CallNode) expressionNode() {}

func (fc *CallNode) Copy() Node {
	return &CallNode{
		Token:          fc.Token.Copy(),
		Function:       copyOrNil(fc.Function),
		PositionalArgs: CopyNodeSlice(fc.PositionalArgs),
		ByNameArgs:     CopyNodeSlice(fc.ByNameArgs),
	}
}

func (fc *CallNode) Evaluate() object.Object {
	return nil
}

func (fc *CallNode) Compile(c *Compiler) (pkg opcode.InsPackage, err error) {
	hasExpansion := false

	// Compiling the function first ...
	// ... but we add it to the instructions after the arguments.
	var fn opcode.InsPackage
	fn, err = fc.Function.Compile(c)
	if err != nil {
		return
	}

	var bslc opcode.InsPackage

	for _, arg := range fc.PositionalArgs {
		if hasExpansion {
			// already set hasExpansion and have another positional argument
			err = c.makeErr(arg, fmt.Sprintf("Argument expansion only possible on last positional argument"))
			return
		}

		switch post := arg.(type) {
		case *PostfixExpressionNode:
			if post.Operator.Type == token.EXPANSION {
				arg = post.Left
				hasExpansion = true
			}
		}

		bslc, err = arg.Compile(c)
		if err != nil {
			return
		}

		pkg = pkg.Append(bslc)
	}

	var externalNames []string

	for _, arg := range fc.ByNameArgs {
		externalName := ""

		if assign, ok := arg.(*AssignmentNode); ok {
			externalName = assign.Identifiers[0].TokenRepresentation()
			name := &StringNode{
				Token: assign.Token, Values: []string{externalName}}
			bslc, err = name.Compile(c)
			if err != nil {
				return
			}

			// compiling to name/value object (internally used for argument by name)
			var value opcode.InsPackage
			value, err = assign.Values[0].Compile(c)
			if err != nil {
				return
			}
			pkg = pkg.Append(bslc.Append(value).Append(opcode.MakePkg(assign.Token, opcode.OpNameValue)))

			// check for duplicate external (argument) names
			if externalName != "" {
				if str.IsInSlice(externalName, externalNames) {
					err = c.makeErr(arg, fmt.Sprintf("Duplicate of argument by name (%s)", str.ReformatInput(externalName)))
					return
				}
				externalNames = append(externalNames, externalName)
			}

		} else {
			// not an assignment node
			err = c.makeErr(arg, fmt.Sprintf("Expected assignment node for argument by name (%s)", str.ReformatInput(externalName)))
			return
		}
	}

	// NOTE: putting function to call onto stack after arguments and will be popped first
	pkg = pkg.Append(fn)

	op := opcode.OpCall
	if hasExpansion {
		op = opcode.OpCallWithExpansion
	}
	pkg = pkg.Append(opcode.MakePkg(fc.Token, op, len(fc.PositionalArgs), len(fc.ByNameArgs)))

	return
}

func (fc *CallNode) TokenRepresentation() string {
	var out bytes.Buffer

	out.WriteString(tokenRepOrNil(fc.Function) + "(")

	for i, a := range fc.PositionalArgs {
		out.WriteString(tokenRepOrNil(a))
		if i < len(fc.PositionalArgs)-1 {
			out.WriteString(", ")
		}
	}
	if len(fc.PositionalArgs) != 0 && len(fc.ByNameArgs) != 0 {
		out.WriteString(", ")
	}
	for i, a := range fc.ByNameArgs {
		out.WriteString(tokenRepOrNil(a))
		if i < len(fc.ByNameArgs)-1 {
			out.WriteString(", ")
		}
	}

	out.WriteString(")")

	return out.String()
}

func (fc *CallNode) String() string {
	var out bytes.Buffer

	out.WriteString("Call ")
	out.WriteString(stringOrNil(fc.Function) + "(")

	for i, a := range fc.PositionalArgs {
		out.WriteString(stringOrNil(a))
		if i < len(fc.PositionalArgs)-1 {
			out.WriteString(", ")
		}
	}
	if len(fc.PositionalArgs) != 0 && len(fc.ByNameArgs) != 0 {
		out.WriteString(", ")
	}
	for i, a := range fc.ByNameArgs {
		out.WriteString(stringOrNil(a))
		if i < len(fc.ByNameArgs)-1 {
			out.WriteString(", ")
		}
	}

	out.WriteString(")")

	return out.String()
}

func (fc *CallNode) TokenInfo() token.Token {
	return fc.Token
}

// UNCOMPILED/USER-DEFINED FUNCTIONS
type FunctionNode struct {
	Token                token.Token
	ReturnType           Node // nil for no explicit return type
	Name                 string
	PositionalParameters []Node
	ByNameParameters     []Node
	Body                 Node
	ImpureEffects        bool
}

func (f *FunctionNode) expressionNode() {}

func (f *FunctionNode) Copy() Node {
	return &FunctionNode{
		Token:                f.Token.Copy(),
		ReturnType:           copyOrNil(f.ReturnType),
		Name:                 f.Name,
		PositionalParameters: CopyNodeSlice(f.PositionalParameters),
		ByNameParameters:     CopyNodeSlice(f.ByNameParameters),
		Body:                 copyOrNil(f.Body),
		ImpureEffects:        f.ImpureEffects,
	}
}

func (f *FunctionNode) Evaluate() object.Object {
	return nil
}

func (f *FunctionNode) Compile(c *Compiler) (pkg opcode.InsPackage, err error) {
	return c.compileFunctionNode(f)
}

func (f *FunctionNode) TokenRepresentation() string {
	params := []string{}
	for _, p := range f.PositionalParameters {
		params = append(params, tokenRepOrNil(p))
	}
	for _, p := range f.ByNameParameters {
		params = append(params, tokenRepOrNil(p))
	}

	var sb strings.Builder
	sb.WriteString(common.FunctionTokenLiteral)
	if f.ImpureEffects {
		sb.WriteRune('*')
	}
	sb.WriteString("(")
	if f.Name != "" {
		sb.WriteString("named " + f.Name + ")(")
	}
	sb.WriteString(strings.Join(params, ", ") + ") ")

	if f.ReturnType != nil {
		sb.WriteString(f.ReturnType.TokenRepresentation())
		sb.WriteByte(' ')
	}

	sb.WriteString(tokenRepOrNil(f.Body))

	return sb.String()
}

func (f *FunctionNode) String() string {
	positional := []string{}
	for _, p := range f.PositionalParameters {
		positional = append(positional, stringOrNil(p))
	}
	byname := []string{}
	for _, p := range f.ByNameParameters {
		byname = append(byname, stringOrNil(p))
	}

	var sb strings.Builder
	sb.WriteString(common.FuntionTypeName)
	if f.ImpureEffects {
		sb.WriteRune('*')
	}
	sb.WriteRune(' ')
	if f.Name != "" {
		sb.WriteString(f.Name + " ")
	}

	if len(positional) != 0 {
		sb.WriteString("Positional: ")
	}
	sb.WriteString(strings.Join(positional, ", "))
	if len(positional) != 0 && len(byname) != 0 {
		sb.WriteString(", ")
	}
	if len(byname) != 0 {
		sb.WriteString("ByName: ")
	}
	sb.WriteString(strings.Join(byname, ", "))

	sb.WriteString(") ")

	if f.ReturnType != nil {
		sb.WriteString(f.ReturnType.String())
		sb.WriteByte(' ')
	}

	sb.WriteString(stringOrNil(f.Body))

	return sb.String()
}

func (f *FunctionNode) TokenInfo() token.Token {
	return f.Token
}

// PARAMETER EXPANSION NODE
type ExpansionNode struct {
	Token        token.Token
	Limits       Node
	Continuation Node
}

func (pe *ExpansionNode) expressionNode() {}

func (pe *ExpansionNode) Copy() Node {
	return &ExpansionNode{
		Token:        pe.Token.Copy(),
		Limits:       copyOrNil(pe.Limits),
		Continuation: copyOrNil(pe.Continuation),
	}
}

func (pe *ExpansionNode) Evaluate() object.Object {
	return nil
}

func (pe *ExpansionNode) Compile(c *Compiler) (opcode.InsPackage, error) {
	return cannotDirectlyCompile("ExpansionNode")
}

func (pe *ExpansionNode) TokenRepresentation() string {
	var out bytes.Buffer

	out.WriteString("...")
	if pe.Limits == nil {
		out.WriteByte(' ')
	} else {
		out.WriteByte('[')
		out.WriteString(pe.Limits.TokenRepresentation())
		out.WriteString("] ")
	}
	out.WriteString(tokenRepOrNil(pe.Continuation))

	return out.String()
}
func (pe *ExpansionNode) String() string {
	var out bytes.Buffer

	out.WriteString("...")
	if pe.Limits == nil {
		out.WriteByte(' ')
	} else {
		out.WriteString("[Limits ")
		out.WriteString(pe.Limits.String())
		out.WriteString("] ")
	}
	out.WriteString(stringOrNil(pe.Continuation))

	return out.String()
}

func (pe *ExpansionNode) TokenInfo() token.Token {
	return pe.Token
}

// FUNCTION SELF NODE
type SelfNode struct {
	Token token.Token
}

func (s *SelfNode) expressionNode() {}

func (s *SelfNode) Copy() Node {
	return &SelfNode{Token: s.Token.Copy()}
}

func (s *SelfNode) Evaluate() object.Object {
	return nil
}

func (s *SelfNode) Compile(c *Compiler) (pkg opcode.InsPackage, err error) {
	return c.compileSelfRef(s)
}

func (s *SelfNode) TokenRepresentation() string {
	return "self"
}
func (s *SelfNode) String() string {
	return "Self"
}

func (s *SelfNode) TokenInfo() token.Token {
	return s.Token
}

// IDENTIFIER
// using NewVariableNode(...) instead of &IdentNode{...} for variables
func NewVariableNode(tok token.Token, name string, sys bool) *IdentNode {
	return &IdentNode{Token: tok, Name: name, System: sys}
}

// using NewBuiltInNode(...) instead of &IdentNode{...} for built-in functions
func NewBuiltInNode(tok token.Token, name string, sys bool) *IdentNode {
	return &IdentNode{Token: tok, Name: name, System: sys}
}

// returns langur Object type code as defined in object package, or 0 if not found
func NodeToLangurType(node Node) object.ObjectType {
	switch id := node.(type) {
	case *IdentNode:
		switch id.Name {
		case common.FunctionTokenLiteral:
			// ... is fn
			return object.COMPILED_CODE_OBJ
		}
		
		code, ok := object.TypeNameToType[id.Name]
		if ok {
			return code
		}
	}
	return 0
}

type IdentNode struct {
	Token  token.Token
	Type   Node // type or subtype; nil for no explicit type
	Name   string
	System bool
}

func (i *IdentNode) expressionNode() {}

func (i *IdentNode) Copy() Node {
	return &IdentNode{
		Token:  i.Token.Copy(),
		Type:   copyOrNil(i.Type),
		Name:   i.Name,
		System: i.System,
	}
}

func (i *IdentNode) Evaluate() object.Object {
	return nil
}

func (i *IdentNode) Compile(c *Compiler) (pkg opcode.InsPackage, err error) {
	if i.Type != nil {
		err = c.makeErr(i, "This version of langur not able to accept explicit variable type")
		return
	}

	bi := process.GetBuiltInByName(i.Name)
	if bi == nil {
		// not a built-in; must be a variable
		return c.resolveAndGetInstructions(i, i.Name)
	}

	if process.GetBuiltInImpurityStatus(i.Name) {
		c.addToImpureEffectsList(i.Name)
	}

	pkg = c.constantIns(bi)
	return
}

func (i *IdentNode) TokenRepresentation() string {
	var out bytes.Buffer

	out.WriteString(i.Name)

	if i.Type != nil {
		out.WriteRune(' ')
		out.WriteString(i.Type.TokenRepresentation())
	}

	return out.String()
}
func (i *IdentNode) String() string {
	var out bytes.Buffer

	if i.System {
		out.WriteString("System ")
	}
	out.WriteString("Ident ")
	out.WriteString(i.Name)

	if i.Type != nil {
		out.WriteRune(' ')
		out.WriteString(i.Type.String())
	}

	return out.String()
}

func (i *IdentNode) TokenInfo() token.Token {
	return i.Token
}

// MODE NODE
type ModeNode struct {
	Token   token.Token
	Name    string
	Setting Node
}

func (m *ModeNode) statementNode() {}

func (m *ModeNode) Copy() Node {
	return &ModeNode{
		Token:   m.Token.Copy(),
		Name:    m.Name,
		Setting: copyOrNil(m.Setting),
	}
}

func (m *ModeNode) Evaluate() object.Object {
	return nil
}

func (m *ModeNode) Compile(c *Compiler) (pkg opcode.InsPackage, err error) {
	if c.symbolTable.Outer != nil {
		err = c.makeErr(m, "Current implementation can only set modes in global context")
		return
		// The idea is that modes will have scope like variables. ...
		// ... Therefore, if set, they have to be reset when exiting scope.
	}
	pkg, err = m.Setting.Compile(c)
	if err != nil {
		return
	}
	code, ok := modes.ModeNames[m.Name]
	if !ok {
		err = c.makeErr(m, fmt.Sprintf("Unknown mode setting %s", m.Name))
		return
	}
	pkg = pkg.Append(opcode.MakePkg(m.Token, opcode.OpMode, code))
	return
}

func (m *ModeNode) TokenRepresentation() string {
	var out bytes.Buffer

	out.WriteString("mode ")
	out.WriteString(m.Name)
	out.WriteString(" = ")
	out.WriteString(tokenRepOrNil(m.Setting))

	return out.String()
}
func (m *ModeNode) String() string {
	var out bytes.Buffer

	out.WriteString("Mode ")
	out.WriteString(m.Name)
	out.WriteString(" = ")
	out.WriteString(stringOrNil(m.Setting))

	return out.String()
}

func (m *ModeNode) TokenInfo() token.Token {
	return m.Token
}

// FOR NODE
type ForNode struct {
	Token         token.Token
	LoopValueInit Node
	Init          []Node
	Test          Node
	Increment     []Node
	Body          Node
}

func (n *ForNode) expressionNode() {}

func (n *ForNode) Copy() Node {
	return &ForNode{
		Token:         n.Token.Copy(),
		LoopValueInit: copyOrNil(n.LoopValueInit),
		Init:          CopyNodeSlice(n.Init),
		Test:          copyOrNil(n.Test),
		Increment:     CopyNodeSlice(n.Increment),
		Body:          copyOrNil(n.Body),
	}
}

func (n *ForNode) Evaluate() object.Object {
	return nil
}

func (n *ForNode) Compile(c *Compiler) (pkg opcode.InsPackage, err error) {
	var init, test, body, increment opcode.InsPackage

	c.pushVariableScope()
	defer func() {
		pkg = c.wrapInstructionsWithExecute(pkg, n.Token)
		c.popVariableScope()
	}()

	// The sections are...
	// 1. init
	// 2. test
	//	(conditionally jump out)
	// 3. body
	// 4. increment
	//	(jump back to test)
	// ... (as of 0.7+ ...)
	// 5. for loop value

	// Prior to 0.7, we would use init = c.noValueIns here to make sure something is on the stack.
	// Now, we set the for loop value a different way.

	for _, each := range n.Init {
		var i opcode.InsPackage
		i, err = c.compileNodeWithPopIfExprStmt(each)
		if err != nil {
			return
		}
		init = init.Append(i)
	}

	var loopValueInit opcode.InsPackage
	loopValueInit, err = c.compileNodeWithPopIfExprStmt(n.LoopValueInit)
	if err != nil {
		return
	}
	init = init.Append(loopValueInit)

	loopValueVar := n.LoopValueInit.(*ExpressionStatementNode).Expression.(*LineDeclarationNode).Assignment.(*AssignmentNode).Identifiers[0]
	// for setting break value when not specified as something else
	c.loopVarStack = append(c.loopVarStack, loopValueVar)
	defer func() {
		c.loopVarStack = c.loopVarStack[:len(c.loopVarStack)-1]
	}()

	if n.Test != nil {
		test, err = n.Test.Compile(c)
		if err != nil {
			return
		}
	}

	body, err = c.compileNodeWithPopIfExprStmt(n.Body)
	if err != nil {
		return
	}

	for _, each := range n.Increment {
		var i opcode.InsPackage
		i, err = c.compileNodeWithPopIfExprStmt(each)
		if err != nil {
			return
		}
		increment = increment.Append(i)
	}

	// fix jumps for next and break, replacing placeholders
	body.Instructions = c.fixJumps(body.Instructions, false, opcode.OC_PlaceHolder_Next, &c.nextStmtCount, len(body.Instructions), 0, 0)
	body.Instructions = c.fixJumps(body.Instructions, false, opcode.OC_PlaceHolder_Break, &c.breakStmtCount, len(body.Instructions)+len(increment.Instructions)+opcode.OP_JUMP_LEN, 0, 0)

	jumptok := token.Token{}
	if len(test.Instructions) > 0 {
		test = test.Append(
			opcode.MakePkg(n.Test.TokenInfo(), opcode.OpJumpIfNotTruthy,
				len(body.Instructions)+len(increment.Instructions)+opcode.OP_JUMP_LEN))
				
		jumptok = n.Test.TokenInfo()
	}
	// after increment, jump back to start of test section (or body if there is no test)
	jumpback := opcode.MakePkg(jumptok, opcode.OpJumpBack, len(test.Instructions)+len(body.Instructions)+len(increment.Instructions))

	pkg = pkg.Append(init).Append(test).Append(body).Append(increment).Append(jumpback)

	// append loop value to very end; vm will push onto stack before exiting frame
	var loopValue opcode.InsPackage
	loopValue, err = c.loopVarStack[len(c.loopVarStack)-1].Compile(c)
	if err != nil {
		return
	}
	pkg = pkg.Append(loopValue)

	return	
}

func (n *ForNode) TokenRepresentation() string {
	var out bytes.Buffer

	out.WriteString(n.Token.Literal)
	out.WriteRune('[')
	out.WriteString(tokenRepOrNil(n.LoopValueInit))
	out.WriteString("] ")

	for i := range n.Init {
		out.WriteString(tokenRepOrNil(n.Init[i]))
		if i < len(n.Init)-1 {
			out.WriteString(", ")
		}
	}
	out.WriteString("; ")

	if n.Test != nil {
		out.WriteString(tokenRepOrNil(n.Test))
	}
	out.WriteString("; ")

	for i := range n.Increment {
		out.WriteString(tokenRepOrNil(n.Increment[i]))
		if i < len(n.Increment)-1 {
			out.WriteString(", ")
		}
	}
	out.WriteString(" ")
	out.WriteString(tokenRepOrNil(n.Body))

	return out.String()
}
func (n *ForNode) String() string {
	var out bytes.Buffer

	if n.Token.Literal == "for" {
		out.WriteString("For")
	} else {
		out.WriteString("While")
	}
	out.WriteRune('[')
	out.WriteString(stringOrNil(n.LoopValueInit))
	out.WriteString("] ")

	if len(n.Init) > 0 {
		out.WriteString("Init ")
		for i := range n.Init {
			out.WriteString(stringOrNil(n.Init[i]))
			if i < len(n.Init)-1 {
				out.WriteString(", ")
			}
		}
	}
	out.WriteString("; ")

	if n.Test != nil {
		out.WriteString("Test ")
		out.WriteString(stringOrNil(n.Test))
	}
	out.WriteString("; ")

	if len(n.Increment) > 0 {
		out.WriteString("Increment ")
		for i := range n.Increment {
			out.WriteString(stringOrNil(n.Increment[i]))
			if i < len(n.Increment)-1 {
				out.WriteString(", ")
			}
		}
	}
	out.WriteString(" ")
	out.WriteString(stringOrNil(n.Body))

	return out.String()
}

func (n *ForNode) TokenInfo() token.Token {
	return n.Token
}

// FOR IN/OF NODE
type ForInOfNode struct {
	Token         token.Token
	LoopValueInit Node
	Var           Node
	Of            bool
	Over          Node
	Body          Node
}

func (n *ForInOfNode) expressionNode() {}

func (n *ForInOfNode) Copy() Node {
	return &ForInOfNode{
		Token:         n.Token.Copy(),
		LoopValueInit: copyOrNil(n.LoopValueInit),
		Var:           copyOrNil(n.Var),
		Of:            n.Of,
		Over:          copyOrNil(n.Over),
		Body:          copyOrNil(n.Body),
	}
}

func (n *ForInOfNode) Evaluate() object.Object {
	return nil
}

func (n *ForInOfNode) Compile(c *Compiler) (opcode.InsPackage, error) {
	return cannotDirectlyCompile("ForInOfNode")
}

func (n *ForInOfNode) TokenRepresentation() string {
	var out bytes.Buffer

	out.WriteString("for")
	out.WriteRune('[')
	out.WriteString(tokenRepOrNil(n.LoopValueInit))
	out.WriteString("] ")

	out.WriteString(tokenRepOrNil(n.Var))
	if n.Of {
		out.WriteString(" of ")
	} else {
		out.WriteString(" in ")
	}
	out.WriteString(tokenRepOrNil(n.Over))
	out.WriteString(" ")
	out.WriteString(tokenRepOrNil(n.Body))

	return out.String()
}
func (n *ForInOfNode) String() string {
	var out bytes.Buffer

	out.WriteString("For")
	out.WriteRune('[')
	out.WriteString(stringOrNil(n.LoopValueInit))
	out.WriteString("] ")

	out.WriteString(stringOrNil(n.Var))
	if n.Of {
		out.WriteString(" Of ")
	} else {
		out.WriteString(" In ")
	}
	out.WriteString(stringOrNil(n.Over))
	out.WriteString(" ")
	out.WriteString(stringOrNil(n.Body))

	return out.String()
}

func (n *ForInOfNode) TokenInfo() token.Token {
	return n.Token
}

// BREAK
type BreakNode struct {
	Token token.Token
	Value Node
}

func (n *BreakNode) statementNode() {}

func (n *BreakNode) Copy() Node {
	return &BreakNode{
		Token: n.Token.Copy(),
		Value: copyOrNil(n.Value),
	}
}

func (n *BreakNode) Evaluate() object.Object {
	return nil
}

func (n *BreakNode) Compile(c *Compiler) (pkg opcode.InsPackage, err error) {
	c.breakStmtCount++

	if len(c.loopVarStack) < 1 {
		err = c.makeErr(n, "Break declared outside of loop")
		return
	}

	if n.Value == nil {
		// break with current for loop value
		pkg, err = c.loopVarStack[len(c.loopVarStack)-1].Compile(c)

	} else {
		// break with specified value
		// FIXME: ? redundancy when embedded in scope (no need to set variable)
		pkg, err = MakeAssignmentExpression(c.loopVarStack[len(c.loopVarStack)-1], n.Value, false).Compile(c)
	}

	pkg = pkg.Append(opcode.MakePkg(n.Token, opcode.OpJumpPlaceHolder, opcode.OC_PlaceHolder_Break))
	return
}

func (n *BreakNode) TokenRepresentation() string {
	var out bytes.Buffer

	out.WriteString("break")
	if n.Value != nil {
		out.WriteString(" ")
		out.WriteString(n.Value.TokenRepresentation())
	}

	return out.String()
}
func (n *BreakNode) String() string {
	var out bytes.Buffer

	out.WriteString("Break")
	if n.Value != nil {
		out.WriteString(" ")
		out.WriteString(n.Value.String())
	}

	return out.String()
}
func (n *BreakNode) TokenInfo() token.Token {
	return n.Token
}

// NEXT
type NextNode struct {
	Token token.Token
}

func (n *NextNode) statementNode() {}

func (n *NextNode) Copy() Node {
	return &NextNode{Token: n.Token.Copy()}
}

func (n *NextNode) Evaluate() object.Object {
	return nil
}

func (n *NextNode) Compile(c *Compiler) (pkg opcode.InsPackage, err error) {
	c.nextStmtCount++
	pkg = opcode.MakePkg(n.Token, opcode.OpJumpPlaceHolder, opcode.OC_PlaceHolder_Next)

	// FIXME: ? redundancy for OpJumpRelay; don't need to push a value here
	pkg = c.noValueIns.Append(pkg)

	return
}

func (n *NextNode) TokenRepresentation() string {
	return "next"
}
func (n *NextNode) String() string {
	return "Next"
}
func (n *NextNode) TokenInfo() token.Token {
	return n.Token
}

// BOOLEAN
type BooleanNode struct {
	Token token.Token
	Value bool
}

func (b *BooleanNode) expressionNode() {}

func (b *BooleanNode) Copy() Node {
	return &BooleanNode{Token: b.Token.Copy(), Value: b.Value}
}

func (b *BooleanNode) Evaluate() object.Object {
	return object.NativeBoolToObject(b.Value)
}

func (b *BooleanNode) Compile(c *Compiler) (pkg opcode.InsPackage, err error) {
	if b.Value {
		pkg = opcode.MakePkg(b.Token, opcode.OpTrue)
	} else {
		pkg = opcode.MakePkg(b.Token, opcode.OpFalse)
	}
	return
}

func (b *BooleanNode) TokenRepresentation() string {
	return fmt.Sprintf("%t", b.Value)
}
func (b *BooleanNode) String() string {
	return "Boolean " + fmt.Sprintf("%t", b.Value)
}

func (b *BooleanNode) TokenInfo() token.Token {
	return b.Token
}

// NULL
type NullNode struct {
	Token token.Token
}

func (n *NullNode) expressionNode() {}

func (n *NullNode) Copy() Node {
	return &NullNode{Token: n.Token.Copy()}
}

func (n *NullNode) Evaluate() object.Object {
	return object.NativeBoolToObject(false)
}

func (n *NullNode) Compile(c *Compiler) (pkg opcode.InsPackage, err error) {
	pkg = opcode.MakePkg(n.Token, opcode.OpNull)
	return
}

func (n *NullNode) TokenRepresentation() string {
	return "null"
}
func (n *NullNode) String() string {
	return "Null"
}

func (n *NullNode) TokenInfo() token.Token {
	return n.Token
}

// NONE, THAT IS _ (UNDERSCORE)
func IsNoOp(node Node) bool {
	switch node.(type) {
	case *NoneNode:
		return token.IsNoneBySymbol(node.TokenInfo())
	}
	return false
}

type NoneNode struct {
	Token token.Token
}

func (no *NoneNode) expressionNode() {}

func (no *NoneNode) Copy() Node {
	return &NoneNode{Token: no.Token.Copy()}
}

func (no *NoneNode) Evaluate() object.Object {
	return nil
}

func (no *NoneNode) Compile(c *Compiler) (pkg opcode.InsPackage, err error) {
	if no.Token.Literal == "_" {
		// must be interpreted by context
		err = c.makeErr(no, "Underscore no-op literal not dealt with in this context")
		return
	}
	// no-op by keyword...
	pkg = c.constantIns(object.NONE)
	return
}

func (no *NoneNode) TokenRepresentation() string {
	return "_"
}

func (no *NoneNode) String() string {
	return "None"
}

func (no *NoneNode) TokenInfo() token.Token {
	return no.Token
}

// STRING
type StringNode struct {
	Token          token.Token
	Values         []string
	Interpolations []Node
}

func (s *StringNode) expressionNode() {}

func (s *StringNode) Copy() Node {
	return &StringNode{
		Token:          s.Token.Copy(),
		Values:         s.Values,
		Interpolations: CopyNodeSlice(s.Interpolations),
	}
}

func (s *StringNode) Evaluate() object.Object {
	if len(s.Interpolations) == 0 {
		return object.NewString(s.Values[0])
	}
	return nil
}

func (s *StringNode) Compile(c *Compiler) (opcode.InsPackage, error) {
	return c.compileString(s, regex.NONE)
}

func (s *StringNode) TokenRepresentation() string {
	if len(s.Interpolations) == 0 {
		return fmt.Sprintf("%q", s.Values[0])
	}
	var out bytes.Buffer

	out.WriteRune('$')
	out.WriteString(s.TokenRepWithoutDollarToken())

	return out.String()
}
func (s *StringNode) TokenRepWithoutDollarToken() string {
	if len(s.Interpolations) == 0 {
		return fmt.Sprintf("%q", s.Values[0])
	}
	var out bytes.Buffer

	out.WriteRune('"')
	for i := range s.Values {
		out.WriteString(fmt.Sprintf(str.EscapeGo(s.Values[i])))
		if i > len(s.Interpolations)-2 {
			out.WriteString("\\{<missing interpolation>}")
		} else if i < len(s.Values)-1 {
			out.WriteString("\\{")
			out.WriteString(tokenRepOrNil(s.Interpolations[i]))
			out.WriteString("}")
		}
	}
	out.WriteRune('"')

	return out.String()
}
func (s *StringNode) String() string {
	if len(s.Interpolations) == 0 {
		return fmt.Sprintf(`String %q`, s.Values[0])
	}
	var out bytes.Buffer

	out.WriteString("String ")
	for i := range s.Values {
		out.WriteString(fmt.Sprintf("Value %d: %q", i+1, s.Values[i]))
		if i > len(s.Interpolations)-2 {
			out.WriteString(" {<Missing Interpolation>} ")
		} else if i < len(s.Values)-1 {
			out.WriteString(fmt.Sprintf(", Interpolation %d: ", i+1))
			out.WriteString(stringOrNil(s.Interpolations[i]))
			out.WriteString(", ")
		}
	}

	return out.String()
}

func (s *StringNode) TokenInfo() token.Token {
	return s.Token
}

// INTERPOLATED NODE
type InterpolatedNode struct {
	Token     token.Token
	Value     Node
	Modifiers []string
}

func (i *InterpolatedNode) expressionNode() {}

func (i *InterpolatedNode) Copy() Node {
	return &InterpolatedNode{
		Token:     i.Token.Copy(),
		Value:     copyOrNil(i.Value),
		Modifiers: str.CopySlice(i.Modifiers),
	}
}

func (i *InterpolatedNode) Evaluate() object.Object {
	return nil
}

func (i *InterpolatedNode) Compile(c *Compiler) (opcode.InsPackage, error) {
	return cannotDirectlyCompile("InterpolatedNode")
}

func (i *InterpolatedNode) TokenRepresentation() string {
	var out bytes.Buffer

	out.WriteString(tokenRepOrNil(i.Value))

	for _, m := range i.Modifiers {
		out.WriteRune(':')
		out.WriteString(m)
	}

	return out.String()
}
func (i *InterpolatedNode) String() string {
	var out bytes.Buffer

	out.WriteString(stringOrNil(i.Value))

	for _, m := range i.Modifiers {
		out.WriteString(" : ")
		out.WriteString(m)
	}

	return out.String()
}

func (i *InterpolatedNode) TokenInfo() token.Token {
	return i.Token
}

// REGEX LITERAL
type RegexNode struct {
	Token     token.Token
	Pattern   Node
	RegexType regex.RegexType
}

func (r *RegexNode) expressionNode() {}

func (r *RegexNode) Copy() Node {
	return &RegexNode{
		Token:     r.Token.Copy(),
		Pattern:   copyOrNil(r.Pattern),
		RegexType: r.RegexType,
	}
}

func (r *RegexNode) Evaluate() object.Object {
	patternNode, ok := r.Pattern.(*StringNode)
	if ok {
		if len(patternNode.Interpolations) == 0 {
			reggie, err := object.NewRegex(patternNode.Values[0], r.RegexType)
			if err != nil {
				return reggie.(*object.Regex)
			}
		}
	}

	return nil
}

func (node *RegexNode) Compile(c *Compiler) (pkg opcode.InsPackage, err error) {
	patternNode, ok := node.Pattern.(*StringNode)
	if !ok {
		err = c.makeErr(node, fmt.Sprintf("Expected String Node within Regex Node"))
		return
	}

	var code int
	if node.RegexType == regex.RE2 {
		code = opcode.OC_Regex_Re2

	} else {
		bug("RegexNode.Compile", "Unknown regex type")
		err = c.makeErr(node, fmt.Sprintf("Unknown regex type"))
		return
	}

	if len(patternNode.Interpolations) == 0 {
		// optimize by compiling a regex pattern now, rather than having the VM compile it
		var re object.Object

		re, err = object.NewRegex(patternNode.Values[0], node.RegexType)
		if err != nil {
			return
		}
		pkg = c.constantIns(re)

	} else {
		pkg, err = c.compileString(patternNode, node.RegexType)
		if err != nil {
			return
		}
		pkg = pkg.Append(opcode.MakePkg(node.Token, opcode.OpRegex, code))
	}

	return
}

func (r *RegexNode) TokenRepresentation() string {
	var out bytes.Buffer

	if len(r.Pattern.(*StringNode).Interpolations) > 0 {
		out.WriteRune('$')
	}

	out.WriteString(r.RegexType.LiteralString())
	out.WriteByte('/')

	out.WriteString(r.Pattern.(*StringNode).TokenRepWithoutDollarToken())

	out.WriteByte('/')

	return out.String()
}
func (r *RegexNode) String() string {
	var out bytes.Buffer

	out.WriteString("Regex(" + r.RegexType.String() + ") ")
	out.WriteString(stringOrNil(r.Pattern))

	return out.String()
}

func (r *RegexNode) TokenInfo() token.Token {
	return r.Token
}

// DATE-TIME
type DateTimeNode struct {
	Token   token.Token
	Pattern Node // pattern string, to be interpreted later
}

func (dt *DateTimeNode) expressionNode() {}

func (dt *DateTimeNode) Copy() Node {
	return &DateTimeNode{
		Token:   dt.Token.Copy(),
		Pattern: dt.Pattern,
	}
}

func (d *DateTimeNode) Evaluate() object.Object {
	patternNode, ok := d.Pattern.(*StringNode)
	if ok {
		if len(patternNode.Interpolations) == 0 {
			s := patternNode.Values[0]
			if !object.StringForNowDateTime(s, true) {
				// only build if not a "now" datetime, which would have to be determined at run-time
				dt, err := object.NewDateTimeFromLiteralString(s, false)
				if err == nil {
					return dt
				}
			}
		}
	}

	return nil
}

func (node *DateTimeNode) Compile(c *Compiler) (pkg opcode.InsPackage, err error) {
	dt := node.Evaluate()
	if dt != nil {
		pkg = c.constantIns(dt)
		return
	}

	patternNode, ok := node.Pattern.(*StringNode)
	if !ok {
		err = c.makeErr(node, fmt.Sprintf("Expected String Node within DateTime Node"))
		return
	}
	s := patternNode.Values[0]

	if len(patternNode.Interpolations) == 0 {
		if !object.IsValidDateTimeString(s, true) {
			err = c.makeErr(node, "Invalid date-time literal string")
			return
		}
	}

	// built at run-time (either contains interpolations or is a "now" date-time)
	code, _, _ := opcode.TokenCodeToOcCode(node.Token.Code)

	pkg, err = c.compileString(patternNode, regex.NONE)
	if err != nil {
		return
	}
	pkg = pkg.Append(opcode.MakePkg(node.Token, opcode.OpDateTime, code))

	return
}

func (dt *DateTimeNode) TokenRepresentation() string {
	return common.DateTimeTokenLiteral + "/" + tokenRepOrNil(dt.Pattern) + "/"
}

func (dt *DateTimeNode) String() string {
	return "DateTime " + stringOrNil(dt.Pattern)
}

func (dt *DateTimeNode) TokenInfo() token.Token {
	return dt.Token
}

// DURATION
type DurationNode struct {
	Token   token.Token
	Pattern Node // pattern string, to be interpreted later
}

func (d *DurationNode) expressionNode() {}

func (d *DurationNode) Copy() Node {
	return &DurationNode{
		Token:   d.Token.Copy(),
		Pattern: d.Pattern,
	}
}

func (d *DurationNode) Evaluate() object.Object {
	patternNode, ok := d.Pattern.(*StringNode)
	if ok {
		if len(patternNode.Interpolations) == 0 {
			dur, err := object.NewDurationFromString(patternNode.Values[0])
			if err == nil {
				return dur
			}
		}
	}

	return nil
}

func (node *DurationNode) Compile(c *Compiler) (pkg opcode.InsPackage, err error) {
	dur := node.Evaluate()
	if dur != nil {
		pkg = c.constantIns(dur)
		return
	}

	patternNode, ok := node.Pattern.(*StringNode)
	if !ok {
		err = c.makeErr(node, fmt.Sprintf("Expected String Node within Duration Node"))
		return
	}

	// built at run-time (contains interpolations)
	pkg, err = c.compileString(patternNode, regex.NONE)
	if err != nil {
		return
	}
	pkg = pkg.Append(opcode.MakePkg(node.Token, opcode.OpDuration))

	return
}

func (d *DurationNode) TokenRepresentation() string {
	return common.DurationTokenLiteral + "/" + tokenRepOrNil(d.Pattern) + "/"
}

func (d *DurationNode) String() string {
	return "Duration " + stringOrNil(d.Pattern)
}

func (d *DurationNode) TokenInfo() token.Token {
	return d.Token
}

// NUMBER
type NumberNode struct {
	Token token.Token
	Value string
	Base  int
	Imaginary bool
}

func (n *NumberNode) expressionNode() {}

func (n *NumberNode) Copy() Node {
	return &NumberNode{
		Token: n.Token.Copy(), 
		Value: n.Value, 
		Base: n.Base, 
		Imaginary: n.Imaginary,
	}
}

func (n *NumberNode) Evaluate() object.Object {
	if !n.Imaginary {
		number, err := object.NumberFromStringBase(n.Value, n.Base)
		if err == nil {
			return number
		}
	}
	return nil
}

func (node *NumberNode) Compile(c *Compiler) (pkg opcode.InsPackage, err error) { 
	if node.Imaginary {
		// stand-alone imaginary number compiled to complex
		pkg, err = c.compileComplexNumber(nil, node, false)
		return
	}

	var number *object.Number
	number, err = c.compileNumberObject(node)
	if err == nil {
		pkg = c.constantIns(number)
	}

	return
}

func (n *NumberNode) TokenRepresentation() string {
	var sb strings.Builder
	
	if n.Base == 10 || n.Base == 0 {
		sb.WriteString(n.Value)
	} else {
		sb.WriteString(str.NumberWithBasePrefix(n.Value, n.Base))
	}
	
	if n.Imaginary {
		sb.WriteByte('i')
	}
	
	return sb.String()
}
func (n *NumberNode) String() string {
	var sb strings.Builder

	sb.WriteString("Number ")
	if n.Base == 10 || n.Base == 0 {
		sb.WriteString(n.Value)

		v, err := str.StrToInt64(n.Value, 10)
		if err == nil {
			sb.WriteString(" (")
			sb.WriteString(str.NumberWithBasePrefix(str.Int64ToStr(v, 16), 16))
			sb.WriteString(")")
		}

	} else {
		sb.WriteString(str.NumberWithBasePrefix(n.Value, n.Base))
	}

	if n.Imaginary {
		sb.WriteRune('i')
	}

	return sb.String()
}

func (n *NumberNode) TokenInfo() token.Token {
	return n.Token
}

func (n *NumberNode) DecodeInt() (i int, err error) {
	base := n.Base
	if base == 0 {
		base = 10
	}
	i, err = str.StrToInt(n.Value, base)
	return
}

// LIST NODE
type ListNode struct {
	Token    token.Token
	Elements []Node
}

func (a *ListNode) expressionNode() {}

func (a *ListNode) Copy() Node {
	return &ListNode{
		Token:    a.Token.Copy(),
		Elements: CopyNodeSlice(a.Elements),
	}
}

func (a *ListNode) Evaluate() object.Object {
	return nil
}

func (node *ListNode) Compile(c *Compiler) (pkg opcode.InsPackage, err error) {
	if len(node.Elements) == 0 {
		// no elements; return empty list constant
		pkg = c.constantIns(object.EmptyList)
		return
	}

	var b opcode.InsPackage
	for _, e := range node.Elements {
		if IsNoOp(e) {
			b = c.constantIns(object.NONE)

		} else {
			b, err = e.Compile(c)
			if err != nil {
				return
			}
		}
		pkg = pkg.Append(b)
	}
	pkg = pkg.Append(opcode.MakePkg(node.Token, opcode.OpList, len(node.Elements)))
	return
}

func (a *ListNode) TokenRepresentation() string {
	elements := []string{}

	for _, el := range a.Elements {
		elements = append(elements, tokenRepOrNil(el))
	}

	return "[" + strings.Join(elements, ", ") + "]"
}
func (a *ListNode) String() string {
	elements := []string{}

	for _, el := range a.Elements {
		elements = append(elements, stringOrNil(el))
	}

	return "List [" + strings.Join(elements, ", ") + "]"
}

func (a *ListNode) TokenInfo() token.Token {
	return a.Token
}

// INDEX NODE
type IndexNode struct {
	Token     token.Token
	Left      Node
	Index     Node
	Alternate Node
}

func (i *IndexNode) expressionNode() {}

func (i *IndexNode) Copy() Node {
	return &IndexNode{
		Token:     i.Token.Copy(),
		Left:      copyOrNil(i.Left),
		Index:     copyOrNil(i.Index),
		Alternate: copyOrNil(i.Alternate)}
}

func (i *IndexNode) Evaluate() object.Object {
	return nil
}

func (node *IndexNode) Compile(c *Compiler) (pkg opcode.InsPackage, err error) {
	var b opcode.InsPackage

	// Get "left" node
	b, err = node.Left.Compile(c)
	if err != nil {
		return
	}
	pkg = b

	// Get the index
	b, err = node.Index.Compile(c)
	if err != nil {
		return
	}
	pkg = pkg.Append(b)

	if node.Alternate == nil {
		pkg = pkg.Append(opcode.MakePkg(node.Token, opcode.OpIndex, 0))

	} else {
		// alternate for an invalid index
		var alt opcode.InsPackage
		alt, err = node.Alternate.Compile(c)
		if err != nil {
			return
		}
		pkg = pkg.Append(opcode.MakePkg(node.Token, opcode.OpIndex, len(alt.Instructions)))
		pkg = pkg.Append(alt)
	}

	return
}

func (i *IndexNode) TokenRepresentation() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(tokenRepOrNil(i.Left))
	out.WriteString("[")
	out.WriteString(tokenRepOrNil(i.Index))

	if i.Alternate != nil {
		out.WriteString("; " + tokenRepOrNil(i.Alternate))
	}
	out.WriteString("])")

	return out.String()
}
func (i *IndexNode) String() string {
	var out bytes.Buffer

	out.WriteString("Index (")
	out.WriteString(stringOrNil(i.Left))
	out.WriteString("[")
	out.WriteString(stringOrNil(i.Index))

	if i.Alternate != nil {
		out.WriteString("; " + stringOrNil(i.Alternate))
	}
	out.WriteString("])")

	return out.String()
}

func (i *IndexNode) TokenInfo() token.Token {
	return i.Token
}

// HASH NODE
type KeyValuePair struct {
	Key   Node
	Value Node
}

type HashNode struct {
	Token token.Token
	Pairs []KeyValuePair
}

func (d *HashNode) expressionNode() {}

func (d *HashNode) Copy() Node {
	pairs := []KeyValuePair{}
	for _, kv := range d.Pairs {
		pairs = append(pairs, KeyValuePair{copyOrNil(kv.Key), copyOrNil(kv.Value)})
	}
	return &HashNode{Token: d.Token.Copy(), Pairs: pairs}
}

func (d *HashNode) Evaluate() object.Object {
	return nil
}

func (node *HashNode) Compile(c *Compiler) (pkg opcode.InsPackage, err error) {
	if len(node.Pairs) == 0 {
		// no entries; return empty hash constant
		pkg = c.constantIns(object.EmptyHash)
		return
	}

	var b opcode.InsPackage
	for _, kv := range node.Pairs {
		b, err = kv.Key.Compile(c)
		if err != nil {
			return
		}
		pkg = pkg.Append(b)

		b, err = kv.Value.Compile(c)
		if err != nil {
			return
		}
		pkg = pkg.Append(b)
	}
	pkg = pkg.Append(opcode.MakePkg(node.Token, opcode.OpHash, len(node.Pairs)*2))
	return
}

func (d *HashNode) TokenRepresentation() string {
	if len(d.Pairs) == 0 {
		return "{:}"
	}

	pairs := []string{}
	for _, kv := range d.Pairs {
		pairs = append(pairs, tokenRepOrNil(kv.Key)+": "+tokenRepOrNil(kv.Value))
	}
	return "{" + strings.Join(pairs, ", ") + "}"
}
func (d *HashNode) String() string {
	if len(d.Pairs) == 0 {
		return "Hash {:}"
	}

	pairs := []string{}
	for _, kv := range d.Pairs {
		pairs = append(pairs, stringOrNil(kv.Key)+": "+stringOrNil(kv.Value))
	}
	return "Hash {" + strings.Join(pairs, ", ") + "}"
}

func (d *HashNode) TokenInfo() token.Token {
	return d.Token
}

// PREFIX EXPRESSION
type PrefixExpressionNode struct {
	Token    token.Token
	Operator token.Token
	Right    Node
}

func (pe *PrefixExpressionNode) expressionNode() {}

func (pe *PrefixExpressionNode) Copy() Node {
	return &PrefixExpressionNode{
		Token:    pe.Token.Copy(),
		Operator: pe.Operator,
		Right:    copyOrNil(pe.Right),
	}
}

func (pe *PrefixExpressionNode) Evaluate() object.Object {
	return nil
}

func (node *PrefixExpressionNode) Compile(c *Compiler) (pkg opcode.InsPackage, err error) {
	var b opcode.InsPackage
	b, err = node.Right.Compile(c)
	if err != nil {
		return
	}

	code, _, _ := opcode.TokenCodeToOcCode(node.Operator.Code)

	switch node.Operator.Type {
	case token.NOT:
		pkg = b.Append(opcode.MakePkg(node.Token, opcode.OpLogicalNegation, code))
	case token.MINUS:
		pkg = b.Append(opcode.MakePkg(node.Token, opcode.OpNumericNegation))
	default:
		err = c.makeErr(node, fmt.Sprintf("Unknown prefix operator %s", token.TypeDescription(node.Operator.Type)))
	}

	return
}

func (pe *PrefixExpressionNode) TokenRepresentation() string {
	var out bytes.Buffer

	out.WriteRune('(')
	out.WriteString(pe.Operator.Literal)
	if pe.Operator.Type == token.NOT {
		out.WriteRune(' ')
	}
	out.WriteString(tokenRepOrNil(pe.Right))
	out.WriteRune(')')

	return out.String()
}
func (pe *PrefixExpressionNode) String() string {
	var out bytes.Buffer

	out.WriteString("Prefix (")
	out.WriteString(operatorTokenString(pe.Operator))
	if pe.Operator.Type == token.NOT {
		out.WriteRune(' ')
	}
	out.WriteString(stringOrNil(pe.Right))
	out.WriteRune(')')

	return out.String()
}

func (pe *PrefixExpressionNode) TokenInfo() token.Token {
	return pe.Token
}

// POSTFIX EXPRESSION
type PostfixExpressionNode struct {
	Token    token.Token
	Left     Node
	Operator token.Token
}

func (pe *PostfixExpressionNode) expressionNode() {}

func (pe *PostfixExpressionNode) Copy() Node {
	return &PostfixExpressionNode{
		Token:    pe.Token.Copy(),
		Left:     copyOrNil(pe.Left),
		Operator: pe.Operator,
	}
}

func (pe *PostfixExpressionNode) Evaluate() object.Object {
	return nil
}

func (pe *PostfixExpressionNode) Compile(c *Compiler) (opcode.InsPackage, error) {
	// FIXME:
	return opcode.InsPackage{}, nil
}

func (pe *PostfixExpressionNode) TokenRepresentation() string {
	var out bytes.Buffer

	out.WriteRune('(')
	out.WriteString(tokenRepOrNil(pe.Left))
	out.WriteString(pe.Operator.Literal)
	out.WriteRune(')')

	return out.String()
}
func (pe *PostfixExpressionNode) String() string {
	var out bytes.Buffer

	out.WriteString("Postfix (")
	out.WriteString(stringOrNil(pe.Left))
	out.WriteString(operatorTokenString(pe.Operator))
	out.WriteString(")")

	return out.String()
}

func (pe *PostfixExpressionNode) TokenInfo() token.Token {
	return pe.Token
}

// INFIX EXPRESSION
type InfixExpressionNode struct {
	Token    token.Token
	Left     Node
	Operator token.Token
	Right    Node
}

func (ie *InfixExpressionNode) expressionNode() {}

func (ie *InfixExpressionNode) Copy() Node {
	return &InfixExpressionNode{
		Token: ie.Token.Copy(),
		Left:  copyOrNil(ie.Left), Operator: ie.Operator, Right: copyOrNil(ie.Right)}
}

func (ie *InfixExpressionNode) Evaluate() object.Object {
	return nil
}

// TODO: untested....
// NOTE: Operations that depend on modes should not be evaluated at compile-time.
// func (ie *InfixExpressionNode) Evaluate() object.Object {
// 	left := ie.Left.Evaluate()
// 	if left != nil {
// 		op, code, isDatabaseOperation, _, err := ie.examineOpToken(ie.Operator)
// 		if err != nil {
// 			return nil
// 		}

// 		right := ie.Right.Evaluate()

// 		if isDatabaseOperation && (
// 			left == object.NULL || right == object.NULL) {
// 			return object.NULL
// 		}

// 		if right != nil {
// 			// evaluated left and right; can we do operation at compile time?
// 			// may depend on a run-time (VM) mode...
// 			if object.OpNotReadyAtCompileTime(op, left, right) {
// 				return nil
// 			}

// 			// try for a result...
// 			result, err := object.InfixOperation(op, left, right, code)
// 			if err == nil {
// 				return result
// 			}
// 		}
// 	}
// 	return nil
// }

// separated so could be used for compiler or evaluation (when possible)
func (node *InfixExpressionNode) examineOpToken(opTok token.Token) (
	op opcode.OpCode, code int, dbOperation, negated bool, err error) {

	code, dbOperation, _ = opcode.TokenCodeToOcCode(opTok.Code)

	op, negated, ok := opcode.InfixTokenToOpCode(opTok)
	if !ok {
		err = fmt.Errorf("no infix token to opcode conversion for %s", token.TypeDescription(node.Operator.Type))
		return
	}
	if negated {
		code |= opcode.OC_Negated_Op
	}

	return
}

func (node *InfixExpressionNode) Compile(c *Compiler) (pkg opcode.InsPackage, err error) {
	var left, right opcode.InsPackage

	op, code, isDatabaseOperation, negated, err2 := node.examineOpToken(node.Operator)
	if err2 != nil {
		err = c.makeErr(node, err2.Error())
		return
	}

	if !negated && node.Operator.Code == 0 {
		pkg, err = c.checkForComplexNumber(node, op)
		if pkg.Instructions != nil || err != nil {
			return
		}
	}

	left, err = node.Left.Compile(c)
	if err != nil {
		return
	}

	rightTypeCode := byte(NodeToLangurType(node.Right))
	rightIsType := rightTypeCode != 0

	if !rightIsType || node.Operator.Type != token.IS {
		right, err = node.Right.Compile(c)
		if err != nil {
			return
		}
	}

	plain := func() (pkg opcode.InsPackage, err error) {
		pkg = left.Append(right)
		pkg = pkg.Append(opcode.MakePkg(node.Token, op))
		if negated {
			pkg = pkg.Append(opcode.MakePkg(node.Token, opcode.OpLogicalNegation, 0))
		}
		return
	}

	plainWithCode := func() (pkg opcode.InsPackage, err error) {
		pkg = left.Append(right)
		pkg = pkg.Append(opcode.MakePkg(node.Token, op, code))
		return
	}

	nonShortCircuiting := func() (pkg opcode.InsPackage, err error) {
		pkg = left.Append(right)
		pkg = pkg.Append(opcode.MakePkg(node.Token, op, code, 0))
		return
	}

	shortCircuiting := func() (pkg opcode.InsPackage, err error) {
		evalWithRight := opcode.MakePkg(node.Token, op, code, 0)

		// len(right)+len(evalWithRight) == opcodes to jump if left gives the answer
		pkg = left.Append(opcode.MakePkg(node.Token, op, code, len(right.Instructions)+len(evalWithRight.Instructions)))

		// if we didn't short-circuit, must evaluate here...
		pkg = pkg.Append(right)
		pkg = pkg.Append(evalWithRight)
		return
	}

	either := func() (pkg opcode.InsPackage, err error) {
		// either: for operations that could have short-circuiting
		// but only when used as "database" (null propagating) operators
		if isDatabaseOperation {
			return shortCircuiting()
		}
		return nonShortCircuiting()
	}

	withCodeAndTypeCode := func() (pkg opcode.InsPackage, err error) {
		tcode := 0 // 0 indicates requirement for right operand
		pkg = left

		if rightIsType {
			tcode = int(rightTypeCode)
		} else {
			pkg = pkg.Append(right)
		}

		pkg = pkg.Append(opcode.MakePkg(node.Token, op, code, tcode))
		return
	}

	switch op {
	case opcode.OpAppend, opcode.OpIn, opcode.OpOf:
		return plainWithCode()

	case opcode.OpIs:
		return withCodeAndTypeCode()

	case opcode.OpRange,
		opcode.OpAdd, opcode.OpSubtract,
		opcode.OpMultiply, opcode.OpDivide,
		opcode.OpTruncateDivide, opcode.OpFloorDivide,
		opcode.OpRemainder, opcode.OpModulus,
		opcode.OpPower, opcode.OpRoot,
		opcode.OpForward:

		return plain()

	case opcode.OpLogicalAnd, opcode.OpLogicalNAnd,
		opcode.OpLogicalOr, opcode.OpLogicalNOr:

		return shortCircuiting()

	case opcode.OpEqual, opcode.OpNotEqual,
		opcode.OpGreaterThan, opcode.OpGreaterThanOrEqual,
		opcode.OpLessThan, opcode.OpLessThanOrEqual,
		opcode.OpDivisibleBy, opcode.OpNotDivisibleBy,

		opcode.OpLogicalXor, opcode.OpLogicalNXor:

		return either()

	default:
		err = c.makeErr(node, fmt.Sprintf("unknown operator (%s)", token.TypeDescription(node.Operator.Type)))
	}

	return
}

func (ie *InfixExpressionNode) TokenRepresentation() string {
	space := " "
	if ie.Operator.Type == token.DOT {
		space = ""
	}
	return "(" + tokenRepOrNil(ie.Left) + space + ie.Operator.Literal + space + tokenRepOrNil(ie.Right) + ")"
}
func (ie *InfixExpressionNode) String() string {
	return "Infix ( " + stringOrNil(ie.Left) + " " + operatorTokenString(ie.Operator) + " " + stringOrNil(ie.Right) + ")"
}

func (ie *InfixExpressionNode) TokenInfo() token.Token {
	return ie.Token
}

// BLOCK
type BlockNode struct {
	Token      token.Token
	Statements []Node
	HasScope   bool
}

func (b *BlockNode) expressionNode() {}

func (b *BlockNode) Copy() Node {
	return &BlockNode{
		Token:      b.Token.Copy(),
		Statements: CopyNodeSlice(b.Statements),
		HasScope:   b.HasScope,
	}
}

func (b *BlockNode) Evaluate() object.Object {
	return nil
}

func (node *BlockNode) Compile(c *Compiler) (pkg opcode.InsPackage, err error) {
	noValueIfEmpty := true

	var ins1 opcode.InsPackage

	if node.HasScope {
		// only wrap expressions containing declarations (as an efficiency improvement)
		if NodeContainsFirstScopeLevelDeclaration(node) {
			defer func() {
				pkg = c.wrapInstructionsWithExecute(pkg, node.Token)
				c.popVariableScope()
			}()
			c.pushVariableScope()
		}
	}

	if noValueIfEmpty && len(node.Statements) == 0 {
		pkg = c.noValueIns

	} else {
		for i, s := range node.Statements {
			if i < len(node.Statements)-1 {
				ins1, err = c.compileNodeWithPopIfExprStmt(s)
			} else {
				// last node in Block; not to pop on last node of Block
				ins1, err = s.Compile(c)
			}
			pkg = pkg.Append(ins1)

			if err != nil {
				return
			}
		}
	}
	return
}

func (b *BlockNode) TokenRepresentation() string {
	var out bytes.Buffer

	if b.HasScope {
		out.WriteString("{(scoped) ")
	} else {
		out.WriteString("{ ")
	}
	for i, s := range b.Statements {
		out.WriteString(tokenRepOrNil(s))
		if i < len(b.Statements)-1 {
			out.WriteString("; ")
		}
	}
	out.WriteString(" }")

	return out.String()
}
func (b *BlockNode) TokenRepresentationWithFallThrough() string {
	var out bytes.Buffer

	if b.HasScope {
		out.WriteString("{(scoped) ")
	} else {
		out.WriteString("{ ")
	}
	for i, s := range b.Statements {
		out.WriteString(tokenRepOrNil(s))
		if i < len(b.Statements)-1 {
			out.WriteString("; ")
		}
	}
	out.WriteString(" ; fallthrough }")

	return out.String()
}

func (b *BlockNode) String() string {
	var out bytes.Buffer

	if b.HasScope {
		out.WriteString("ScopeBlock { ")
	} else {
		out.WriteString("Block { ")
	}
	for i, s := range b.Statements {
		out.WriteString(stringOrNil(s))
		if i < len(b.Statements)-1 {
			out.WriteString("; ")
		}
	}
	out.WriteString(" }")

	return out.String()
}

func (b *BlockNode) TokenInfo() token.Token {
	return b.Token
}

// IF ... ELSEIF ... ELSE
type TestDo struct {
	Test Node
	Do   Node
}

func (td *TestDo) Copy() *TestDo {
	return &TestDo{Test: copyOrNil(td.Test), Do: copyOrNil(td.Do)}
}

type IfNode struct {
	Token           token.Token
	TestsAndActions []TestDo
	IsSwitchExpr    bool
}

func (i *IfNode) expressionNode() {}

func (i *IfNode) Copy() Node {
	taSlc := []TestDo{}
	for _, td := range i.TestsAndActions {
		taSlc = append(taSlc, *(td.Copy()))
	}

	return &IfNode{
		Token:           i.Token.Copy(),
		TestsAndActions: taSlc,
		IsSwitchExpr:    i.IsSwitchExpr,
	}
}

func (i *IfNode) Evaluate() object.Object {
	return nil
}

func (node *IfNode) Compile(c *Compiler) (pkg opcode.InsPackage, err error) {
	if node.TestsAndActions[len(node.TestsAndActions)-1].Test != nil {
		// no else/default section; add implicit else/default section returning null
		node.TestsAndActions = append(node.TestsAndActions,
			TestDo{Test: nil, Do: &BlockNode{Statements: []Node{NoValue}}})
	}

	type compiled struct {
		pkg opcode.InsPackage
		st  *symbol.SymbolTable
	}
	compiledTests := make([]compiled, len(node.TestsAndActions))
	compiledActions := make([]compiled, len(node.TestsAndActions))
	compiledTA := make([]compiled, len(node.TestsAndActions))

	/*
		opCodes (except for the great complication of scope frames)...
		test
		jump to next test if not truthy
		action
		jump to end
		...
		(rinse, repeat)
	*/

	/*
		Each section of if/else gets it's own scope. This compiles each to one of the following.
		1. no scope wrapping (no declarations in test or action)
		2. wrap scope over test and action (declarations in test and possibly in action)
		3. wrap scope over action only (declarations in action, but none in test)
	*/

	jumpToEndOpCodeLen := func(i int) int {
		// not adding a jump to the end if we already end with a fallthrough, break, or next
		if EndsWithDefiniteJump(node.TestsAndActions[i].Do.(*BlockNode).Statements) {
			return 0
		} else {
			return opcode.OP_JUMP_LEN
		}
	}

	// Compile tests first.
	for i, ta := range node.TestsAndActions {
		lastOne := i == len(node.TestsAndActions)-1

		compiledTests[i].st = nil

		if ta.Test == nil {
			if !lastOne {
				// Houston, we have a bug.
				if node.IsSwitchExpr {
					bug("compileIfExpression", "Default not last part of switch expression")
					err = c.makeErr(node, "Default not last part of switch expression")
				} else {
					bug("compileIfExpression", "Else not last part of if/else expression")
					err = c.makeErr(node, "Else not last part of if/else expression")
				}
				return
			}

		} else {
			if NodeContainsFirstScopeLevelDeclaration(ta.Test) {
				// push and pop and save symbol table; wrap test/action together later
				c.pushVariableScope()
				compiledTests[i].pkg, err = ta.Test.Compile(c)
				compiledTests[i].st = c.symbolTable // save table for re-use
				c.popVariableScope()
			} else {
				// no scope on test
				compiledTests[i].pkg, err = ta.Test.Compile(c)
			}
			if err != nil {
				return
			}
		}
	}

	// Now compile the actions.
	for i, ta := range node.TestsAndActions {
		compiledActions[i].st = nil

		// push scope?
		if compiledTests[i].st != nil {
			// declarations in the test section and possibly in the action
			// using saved symbol table
			c.pushVariableScopeWithTable(compiledTests[i].st)
			compiledActions[i].st = c.symbolTable

		} else if NodeContainsFirstScopeLevelDeclaration(ta.Do) {
			// declarations in the action, but not the test section
			// using new symbol table
			c.pushVariableScope()
			compiledActions[i].st = c.symbolTable
		}

		compiledActions[i].pkg, err = ta.Do.Compile(c)
		if err != nil {
			return
		}

		if compiledActions[i].st != nil {
			if compiledTests[i].st == nil {
				// wrap only action into scope, not the test
				compiledActions[i].pkg = c.wrapInstructionsWithExecute(compiledActions[i].pkg, ta.Do.TokenInfo())
			}
			c.popVariableScope()
		}

		// set conditional jump over action
		if len(compiledTests[i].pkg.Instructions) > 0 {
			// not "else" or "default"
			if compiledTests[i].st == nil {
				compiledTests[i].pkg = compiledTests[i].pkg.Append(
					opcode.MakePkg(ta.Do.TokenInfo(), opcode.OpJumpIfNotTruthy, len(compiledActions[i].pkg.Instructions)+jumpToEndOpCodeLen(i)))

			} else {
				// going to have to add an OpJumpRelayIfNotTruthy b/c of test being buried in scope
				compiledTests[i].pkg = compiledTests[i].pkg.Append(
					opcode.MakePkg(ta.Do.TokenInfo(), opcode.OpJumpPlaceHolder, opcode.OC_PlaceHolder_IfElse_TestFailed))
			}
		}
	}

	// put it together
	for i := range node.TestsAndActions {
		lastOne := i == len(node.TestsAndActions)-1

		if node.IsSwitchExpr {
			// fix fallthrough on switch expressions only

			if compiledTests[i].st != nil {
				// If we allowed declarations within case statements, they would have to be included...
				// ...in scope wrapping, making it impossible to set a jump for fallthrough.
				err = c.makeErr(node, "Cannot use declarations in case statement of switch expression")
				return
			}

			// not looking for fallthrough in default section
			if !lastOne {
				compiledActions[i].pkg.Instructions = c.fixJumps(
					compiledActions[i].pkg.Instructions, false,
					opcode.OC_PlaceHolder_Fallthrough, &c.fallthroughStmtCount,

					len(compiledActions[i].pkg.Instructions)+ // over current action
						jumpToEndOpCodeLen(i)+ // over jump to end
						len(compiledTests[i+1].pkg.Instructions), // over next test

					0, 0)
			}
		}

		// put together test and action
		compiledTA[i].pkg = compiledTests[i].pkg.Append(compiledActions[i].pkg)
		compiledTA[i].st = compiledActions[i].st

		if compiledTests[i].st != nil {
			// wrap test and action into scope together
			c.pushVariableScopeWithTable(compiledTests[i].st)
			compiledTA[i].pkg = c.wrapInstructionsWithExecute(compiledTA[i].pkg, node.TestsAndActions[i].Test.TokenInfo())
			c.popVariableScope()

			compiledTA[i].pkg.Instructions = c.fixJumps(
				compiledTA[i].pkg.Instructions, true,
				opcode.OC_PlaceHolder_IfElse_TestFailed, nil,
				len(compiledTA[i].pkg.Instructions), jumpToEndOpCodeLen(i), 0)
		}

		// jump to end
		if !lastOne {
			if jumpToEndOpCodeLen(i) > 0 {
				compiledTA[i].pkg = compiledTA[i].pkg.Append(
					opcode.MakePkg(node.TestsAndActions[i].Test.TokenInfo(), opcode.OpJumpPlaceHolder, opcode.OC_PlaceHolder_IfElse_Exit))
			}
		}

		pkg = pkg.Append(compiledTA[i].pkg)
	}

	pkg.Instructions = c.fixJumps(pkg.Instructions, false, opcode.OC_PlaceHolder_IfElse_Exit, nil, len(pkg.Instructions), 0, 0)
	return
}

func (i *IfNode) TokenRepresentation() string {
	var out bytes.Buffer

	for ta := range i.TestsAndActions {
		if ta > 0 {
			out.WriteString(" else ")
		}
		if i.TestsAndActions[ta].Test != nil {
			if ta == 0 && i.IsSwitchExpr {
				out.WriteString("if(switch) ")
			} else {
				out.WriteString("if ")
			}
			out.WriteString(tokenRepOrNil(i.TestsAndActions[ta].Test))
			out.WriteString(" ")
		}
		out.WriteString(tokenRepOrNil(i.TestsAndActions[ta].Do))
	}

	return out.String()
}

func (i *IfNode) String() string {
	var out bytes.Buffer

	for ta := range i.TestsAndActions {
		if ta > 0 {
			out.WriteString(" Else ")
		}
		if i.TestsAndActions[ta].Test != nil {
			if ta == 0 && i.IsSwitchExpr {
				out.WriteString("If(Switch) ")
			} else {
				out.WriteString("If ")
			}
			out.WriteString(stringOrNil(i.TestsAndActions[ta].Test))
			out.WriteString(" ")
		}
		out.WriteString(stringOrNil(i.TestsAndActions[ta].Do))
	}

	return out.String()
}

func (i *IfNode) TokenInfo() token.Token {
	return i.Token
}

// SWITCH ... CASE ... DEFAULT
type CaseDo struct {
	MatchConditions []Node
	OtherConditions []Node
	Do              Node
	MatchLogicalOp  token.Token
	OtherLogicalOp  token.Token
}

func (cd *CaseDo) Copy() *CaseDo {
	return &CaseDo{MatchConditions: CopyNodeSlice(cd.MatchConditions),
		OtherConditions: CopyNodeSlice(cd.OtherConditions),
		Do:              copyOrNil(cd.Do), MatchLogicalOp: cd.MatchLogicalOp}
}

type PartialExpr struct {
	Expr        Node
	Op          token.Token
	LeftOperand bool
}

func (pe *PartialExpr) Copy() *PartialExpr {
	return &PartialExpr{
		Expr:        copyOrNil(pe.Expr),
		Op:          pe.Op,
		LeftOperand: pe.LeftOperand,
	}
}

type SwitchNode struct {
	Token            token.Token
	Expressions      []PartialExpr
	CasesAndActions  []CaseDo
	DefaultLogicalOp token.Token
}

func (g *SwitchNode) expressionNode() {}

func (g *SwitchNode) Copy() Node {
	cdSlc := []CaseDo{}
	for _, cd := range g.CasesAndActions {
		cdSlc = append(cdSlc, *(cd.Copy()))
	}
	eSlc := []PartialExpr{}
	for _, pe := range g.Expressions {
		eSlc = append(eSlc, *(pe.Copy()))
	}
	return &SwitchNode{
		Token:            g.Token.Copy(),
		CasesAndActions:  cdSlc,
		Expressions:      eSlc,
		DefaultLogicalOp: g.DefaultLogicalOp,
	}
}

func (g *SwitchNode) Evaluate() object.Object {
	return nil
}

func (g *SwitchNode) Compile(c *Compiler) (opcode.InsPackage, error) {
	return cannotDirectlyCompile("SwitchNode")
}

func (g *SwitchNode) TokenRepresentation() string {
	var out bytes.Buffer

	if g.DefaultLogicalOp.Type == token.OR {
		out.WriteString("switch ")
	} else {
		out.WriteString("switch[" + g.DefaultLogicalOp.Literal + "] ")
	}

	for i, v := range g.Expressions {
		if v.Op.Type == token.EQUAL {
			out.WriteString(tokenRepOrNil(v.Expr))
		} else {
			if v.LeftOperand {
				out.WriteString(tokenRepOrNil(v.Expr) + operatorTokenString(v.Op))
			} else {
				out.WriteString(operatorTokenString(v.Op) + " " + tokenRepOrNil(v.Expr))
			}
		}
		if i < len(g.Expressions)-1 {
			out.WriteString(", ")
		}
	}

	out.WriteString(" { ")

	for i, cd := range g.CasesAndActions {
		if cd.MatchConditions == nil && cd.OtherConditions == nil {
			out.WriteString("default")
		} else {
			out.WriteString("case ")
		}

		if cd.MatchConditions != nil {
			if cd.MatchLogicalOp.Type != token.AND {
				out.WriteString(cd.MatchLogicalOp.Literal + " ")
			}

			for i, c := range cd.MatchConditions {
				out.WriteString(tokenRepOrNil(c))
				if i < len(cd.MatchConditions)-1 {
					out.WriteString(", ")
				}
			}
		}
		if cd.OtherConditions != nil {
			if cd.OtherLogicalOp.Type != cd.MatchLogicalOp.Type {
				out.WriteString(cd.OtherLogicalOp.Literal + " ")
			}

			for i, c := range cd.OtherConditions {
				out.WriteString(tokenRepOrNil(c))
				if i < len(cd.OtherConditions)-1 {
					out.WriteString(", ")
				}
			}
		}

		out.WriteString(": ")
		out.WriteString(tokenRepOrNil(cd.Do))

		if i < len(g.CasesAndActions)-1 {
			out.WriteString("; ")
		}
	}

	out.WriteString(" } ")

	return out.String()
}

func (g *SwitchNode) String() string {
	var out bytes.Buffer

	if g.DefaultLogicalOp.Type == token.OR {
		out.WriteString("Switch ")
	} else {
		out.WriteString("Switch[" + g.DefaultLogicalOp.Literal + "] ")
	}

	for i, v := range g.Expressions {
		if v.Op.Type == token.EQUAL {
			out.WriteString(stringOrNil(v.Expr))
		} else {
			if v.LeftOperand {
				out.WriteString(stringOrNil(v.Expr) + " OP(" + operatorTokenString(v.Op) + ")")
			} else {
				out.WriteString("OP(" + operatorTokenString(v.Op) + ") " + stringOrNil(v.Expr))
			}
		}
		if i < len(g.Expressions)-1 {
			out.WriteString(", ")
		}
	}

	out.WriteString(" { ")

	for i, cd := range g.CasesAndActions {
		if cd.MatchConditions == nil && cd.OtherConditions == nil {
			out.WriteString("Default")
		} else {
			out.WriteString("Case ")
		}

		if cd.MatchConditions != nil {
			if cd.MatchLogicalOp.Type != token.AND {
				out.WriteString("(Op " + operatorTokenString(cd.MatchLogicalOp) + ") ")
			}

			for i, c := range cd.MatchConditions {
				out.WriteString(stringOrNil(c))
				if i < len(cd.MatchConditions)-1 {
					out.WriteString(", ")
				}
			}
		}
		if cd.OtherConditions != nil {
			if cd.OtherLogicalOp.Type != cd.MatchLogicalOp.Type {
				out.WriteString("(Op " + operatorTokenString(cd.OtherLogicalOp) + ") ")
			}

			for i, c := range cd.OtherConditions {
				out.WriteString(stringOrNil(c))
				if i < len(cd.OtherConditions)-1 {
					out.WriteString(", ")
				}
			}
		}

		out.WriteString(": ")
		out.WriteString(stringOrNil(cd.Do))

		if i < len(g.CasesAndActions)-1 {
			out.WriteString("; ")
		}
	}

	out.WriteString(" } ")

	return out.String()
}

func (g *SwitchNode) TokenInfo() token.Token {
	return g.Token
}

// FALLTHROUGH
// for switch expression

type FallThroughNode struct {
	Token token.Token
}

func (f *FallThroughNode) statementNode() {}

func (f *FallThroughNode) Copy() Node {
	return &FallThroughNode{Token: f.Token.Copy()}
}

func (f *FallThroughNode) Evaluate() object.Object {
	return nil
}

func (node *FallThroughNode) Compile(c *Compiler) (pkg opcode.InsPackage, err error) {
	c.fallthroughStmtCount++
	pkg = opcode.MakePkg(node.Token, opcode.OpJumpPlaceHolder, opcode.OC_PlaceHolder_Fallthrough)
	return
}

func (f *FallThroughNode) TokenRepresentation() string {
	return "fallthrough"
}
func (f *FallThroughNode) String() string {
	return "FallThrough"
}

func (f *FallThroughNode) TokenInfo() token.Token {
	return f.Token
}

// TRY / CATCH NODE
type TryCatchNode struct {
	Token        token.Token
	Try          Node
	ExceptionVar Node
	Catch        Node
	Else         Node
}

func (t *TryCatchNode) expressionNode() {}

func (t *TryCatchNode) Copy() Node {
	return &TryCatchNode{
		Token:        t.Token.Copy(),
		Try:          copyOrNil(t.Try),
		ExceptionVar: copyOrNil(t.ExceptionVar),
		Catch:        copyOrNil(t.Catch),
		Else:         copyOrNil(t.Else),
	}
}

func (t *TryCatchNode) Evaluate() object.Object {
	return nil
}

func (node *TryCatchNode) Compile(c *Compiler) (pkg opcode.InsPackage, err error) {
	var try, catch, tcelse opcode.InsPackage

	// The try frame doesn't have scope, but catch and else frames do.
	c.pushNonScope()
	try, err = node.Try.Compile(c)
	c.popVariableScope()
	if err != nil {
		return
	}
	tryIndex := c.addConstant(&object.CompiledCode{InsPackage: try})

	// push scope for the catch frame, including the exception variable
	c.pushVariableScope()
	defer c.popVariableScope()

	var setException opcode.InsPackage
	if node.ExceptionVar != nil {
		setException, err = c.compileNodeWithPopIfExprStmt(
			MakeDeclarationAssignmentStatement(node.ExceptionVar, nil, true, false),
		)

		if err != nil {
			return
		}
	}

	catch, err = node.Catch.Compile(c)
	if err != nil {
		return
	}
	if node.ExceptionVar != nil {
		catch = setException.Append(catch)
	}
	catchIndex := c.wrapInstructions(catch)

	elseIndex := 0
	if node.Else != nil {
		// pop scope from catch; else with different scope
		c.popVariableScope()
		c.pushVariableScope()

		tcelse, err = node.Else.Compile(c)
		elseIndex = c.wrapInstructions(tcelse)
		if elseIndex == 0 {
			bug("compileTryCatch", "elseIndex 0 (0 used as indicator for no else section)")
		}
	}

	pkg = opcode.MakePkg(node.Token, opcode.OpTryCatch, tryIndex, catchIndex, elseIndex)
	return
}

func (t *TryCatchNode) TokenRepresentation() string {
	var out bytes.Buffer

	out.WriteString("(try) ")
	out.WriteString(tokenRepOrNil(t.Try))
	out.WriteString(" catch ")

	if t.ExceptionVar != nil {
		out.WriteString(t.ExceptionVar.TokenRepresentation() + " ")
	}
	out.WriteString(tokenRepOrNil(t.Catch))

	if t.Else != nil {
		out.WriteString(" else ")
		out.WriteString(t.Else.TokenRepresentation())
	}

	return out.String()
}
func (t *TryCatchNode) String() string {
	var out bytes.Buffer

	out.WriteString("Try ")
	if t.Try == nil {
		out.WriteString("(TBD)")
	} else {
		out.WriteString(stringOrNil(t.Try))
	}
	out.WriteString(" Catch ")

	if t.ExceptionVar != nil {
		out.WriteString(t.ExceptionVar.String() + " ")
	}
	out.WriteString(stringOrNil(t.Catch))

	if t.Else != nil {
		out.WriteString(" Else ")
		out.WriteString(t.Else.String())
	}

	return out.String()
}

func (t *TryCatchNode) TokenInfo() token.Token {
	return t.Token
}

// THROW STATEMENT NODE
type ThrowNode struct {
	Token     token.Token
	Exception Node
}

func (t *ThrowNode) statementNode() {}

func (t *ThrowNode) Copy() Node {
	return &ThrowNode{Token: t.Token.Copy()}
}

func (t *ThrowNode) Evaluate() object.Object {
	return nil
}

func (node *ThrowNode) Compile(c *Compiler) (pkg opcode.InsPackage, err error) {
	pkg, err = node.Exception.Compile(c)
	if err != nil {
		return
	}
	pkg = pkg.Append(opcode.MakePkg(node.Token, opcode.OpThrow))
	return
}

func (t *ThrowNode) TokenRepresentation() string { return "throw " + tokenRepOrNil(t.Exception) }

func (t *ThrowNode) String() string {
	return "Throw " + stringOrNil(t.Exception)
}

func (t *ThrowNode) TokenInfo() token.Token {
	return t.Token
}
