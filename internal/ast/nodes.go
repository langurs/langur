// langur/ast/nodes.go

package ast

import (
	"bytes"
	"fmt"
	"langur/common"
	"langur/object"
	"langur/regex"
	"langur/str"
	"langur/token"
	"strings"
)

// NOTE: If adding node types, you may need to add them to functions in ast/search.go.

// THE BASE OF THE TREE
type Program struct {
	Token        token.Token
	Statements   []Node
	VarNamesUsed []string
}

func (p *Program) Copy() Node {
	return &Program{Token: p.Token.Copy(), Statements: CopyNodeSlice(p.Statements)}
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
}

func (d *LineDeclarationNode) expressionNode() {}

func (d *LineDeclarationNode) Copy() Node {
	return &LineDeclarationNode{
		Token:      d.Token.Copy(),
		Assignment: copyOrNil(d.Assignment),
		Mutable:    d.Mutable,
	}
}

func (d *LineDeclarationNode) TokenRepresentation() string {
	var out bytes.Buffer

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
	return &ExpressionStatementNode{Token: es.Token.Copy(), Expression: copyOrNil(es.Expression)}
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

// UNCOMPILED FUNCTIONS
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

func NodeToTypeCode(node Node) int {
	switch id := node.(type) {
	case *IdentNode:
		code, ok := object.TypeNameToType[id.Name]
		if ok {
			return int(code)
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

func (i *IdentNode) TokenRepresentation() string {
	var out bytes.Buffer

	out.WriteString(i.Name)

	if i.Type != nil {
		out.WriteByte(' ')
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
		out.WriteByte(' ')
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

func (b *BooleanNode) Evaluate() (object.Object, bool) {
	return object.NativeBoolToObject(b.Value), true
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

func (n *NullNode) Evaluate() (object.Object, bool) {
	return object.NativeBoolToObject(false), true
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

// NONE, THAT IS _ (UNDERSCORE) SOMEASTERISK
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

func (s *StringNode) Evaluate() (object.Object, bool) {
	if len(s.Interpolations) == 0 {
		return object.NewString(s.Values[0]), true
	}
	return nil, false
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

func (r *RegexNode) Evaluate() (object.Object, bool) {
	var re *object.Regex

	patternNode, ok := r.Pattern.(*StringNode)
	if ok {
		ok = len(patternNode.Interpolations) == 0
		if ok {
			re2, err := object.NewRegex(patternNode.Values[0], r.RegexType)
			ok = err != nil
			if ok {
				re = re2.(*object.Regex)
			}
		}
	}

	return re, ok
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

func (d *DateTimeNode) Evaluate() (object.Object, bool) {
	var dt *object.DateTime

	patternNode, ok := d.Pattern.(*StringNode)
	if ok {
		ok = len(patternNode.Interpolations) == 0
		if ok {
			s := patternNode.Values[0]
			ok = !object.StringForNowDateTime(s, true)
			if ok {
				// only build if not a "now" datetime, which would have to be determined at run-time
				var err error
				dt, err = object.NewDateTimeFromLiteralString(s, false)
				ok = err == nil
			}
		}
	}

	return dt, ok
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

func (d *DurationNode) Evaluate() (object.Object, bool) {
	var dur *object.Duration

	patternNode, ok := d.Pattern.(*StringNode)
	if ok {
		ok = len(patternNode.Interpolations) == 0
		if ok {
			var err error
			dur, err = object.NewDurationFromString(patternNode.Values[0])
			ok = err == nil
		}
	}
	return dur, ok
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
}

func (n *NumberNode) expressionNode() {}

func (n *NumberNode) Copy() Node {
	return &NumberNode{Token: n.Token.Copy(), Value: n.Value, Base: n.Base}
}

func (n *NumberNode) Evaluate() (object.Object, bool) {
	number, err := object.NumberFromStringBase(n.Value, n.Base)
	ok := err == nil
	return number, ok
}

func (n *NumberNode) TokenRepresentation() string {
	if n.Base == 10 || n.Base == 0 {
		return n.Value
	}
	return str.NumberWithBasePrefix(n.Value, n.Base)
}
func (n *NumberNode) String() string {
	var out bytes.Buffer

	out.WriteString("Number ")
	if n.Base == 10 || n.Base == 0 {
		out.WriteString(n.Value)

		v, err := str.StrToInt64(n.Value, 10)
		if err == nil {
			out.WriteString(" (")
			out.WriteString(str.NumberWithBasePrefix(str.Int64ToStr(v, 16), 16))
			out.WriteString(")")
		}

	} else {
		out.WriteString(str.NumberWithBasePrefix(n.Value, n.Base))
	}

	return out.String()
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
		Token: i.Token.Copy(),
		Left:  copyOrNil(i.Left), Index: copyOrNil(i.Index), Alternate: copyOrNil(i.Alternate)}
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

// func (ie *InfixExpressionNode) Evaluate() (object.Object, bool) {
// 	left, ok := TryEvaluate(ie.Left)
// 	if ok {
// 		right, ok := TryEvaluate(ie.Right)
// 		if ok {
// 			obj, err := object.BinaryOperation(ie.Operator, left, right, 0)
// 			ok = err == nil
// 			return obj, ok
// 		}
// 	}
// 	return nil, false
// }

func (ie *InfixExpressionNode) TokenRepresentation() string {
	return "(" + tokenRepOrNil(ie.Left) + " " + ie.Operator.Literal + " " + tokenRepOrNil(ie.Right) + ")"
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

func (t *ThrowNode) TokenRepresentation() string { return "throw " + tokenRepOrNil(t.Exception) }

func (t *ThrowNode) String() string {
	return "Throw " + stringOrNil(t.Exception)
}

func (t *ThrowNode) TokenInfo() token.Token {
	return t.Token
}
