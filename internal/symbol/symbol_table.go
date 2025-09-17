// langur/symbol/symbol_table.go

package symbol

import (
	"fmt"
	"langur/modes"
	"langur/object"
	"langur/str"
	"langur/token"
	"langur/vm/process"
)

func bug(fnName, s string) {
	panic("Symbol Table Bug: " + s)
}

type symbolScope string

const (
	GlobalScope symbolScope = "GLOBAL"
	LocalScope  symbolScope = "LOCAL"
	FreeScope   symbolScope = "FREE"
	SelfScope   symbolScope = "SELF"
)

// Note: Built-ins don't have scope and don't need a symbol table.

type Symbol struct {
	Name    string
	Type    object.ObjectType
	Scope   symbolScope
	Index   int
	Mutable bool
}

type SymbolTable struct {
	Outer           *SymbolTable
	store           map[string]Symbol
	DefinitionCount int
	IsNonScope      bool // a placeholder for a non-scoped VM frame

	FreeSymbols      []Symbol // free symbols for closures
	FreezeDefineFree bool     // when setting optional parameter defaults
	IsFunction       bool

	Modes         *modes.CompileModes
	ImpureEffects []string
}

func NewSymbolTable(
	outer *SymbolTable,
	modes *modes.CompileModes) *SymbolTable {

	return &SymbolTable{
		store:       make(map[string]Symbol),
		Outer:       outer,
		FreeSymbols: []Symbol{},
		Modes:       modes,
	}
}

// for pushing variable scope with an existing symbol table
func ReuseSymbolTable(outer, reuse *SymbolTable) *SymbolTable {
	reuse.Outer = outer
	return reuse
}

func (st *SymbolTable) DefineVariable(name string, mutable, system bool) (sym Symbol, err error) {
	if system {
		return st.DefineSystemVariable(name, mutable)
	}
	return st.DefineUserVariable(name, mutable)
}

func (st *SymbolTable) DefineUserVariable(name string, mutable bool) (sym Symbol, err error) {
	if name[0] == '_' {
		err = fmt.Errorf("User-defined variable names cannot start with underscore")
		return
	}
	return st.defineSymbol(name, mutable, 0)
}

func (st *SymbolTable) DefineSystemVariable(name string, mutable bool) (sym Symbol, err error) {
	if name[0] != '_' {
		bug("DefineSystemVariable", fmt.Sprintf("System variable name %q invalid", name))
		err = fmt.Errorf("System variable names start with underscore")
		return
	}
	return st.defineSymbol(name, mutable, 0)
}

// to check if identifier name allowed for declaration
func isNonShadowedWord(name string) bool {
	_, isKeyword := token.Keywords[name]
	if isKeyword {
		return true
	}
	bi := process.GetBuiltInByName(name)
	if bi != nil {
		return true
	}
	return false
}

func (st *SymbolTable) defineSymbol(name string, mutable bool, stype object.ObjectType) (Symbol, error) {
	if st.IsNonScope {
		// safe to do this blindly (without checking for null), ...
		// ... as the root table will never be a non-scope table
		return st.Outer.defineSymbol(name, mutable, stype)
	}

	// first, check if it is already defined in this scope
	sym, ok := st.store[name]
	if ok {
		return sym, fmt.Errorf("Cannot declare variable (%s) already declared within scope", name)
	}

	// check if a non-shadowed word already defined
	if isNonShadowedWord(name) {
		return sym, fmt.Errorf("Cannot declare variable as shadow to referent or keyword (%s)", name)
	}

	sym = Symbol{
		Name:    name,
		Type:    stype,
		Index:   st.DefinitionCount,
		Mutable: mutable,
	}

	if st.Outer == nil {
		sym.Scope = GlobalScope
	} else {
		sym.Scope = LocalScope
	}

	st.store[name] = sym
	st.DefinitionCount++

	return sym, nil
}

func (st *SymbolTable) defineRootSymbol(name string, mutable bool) (Symbol, error) {
	if st.Outer == nil {
		return st.defineSymbol(name, mutable, 0)
	}
	return st.Outer.defineRootSymbol(name, mutable)
}

func (st *SymbolTable) defineFree(original Symbol) Symbol {
	st.FreeSymbols = append(st.FreeSymbols, original)

	sym := Symbol{
		Name:    original.Name,
		Type:    original.Type.Copy(),
		Index:   len(st.FreeSymbols) - 1,
		Scope:   FreeScope,
		Mutable: false,
	}

	st.store[original.Name] = sym
	return sym
}

func (st *SymbolTable) DefineSelf(name string) Symbol {
	sym := Symbol{
		Name:    name,
		Type:    0,
		Index:   0,
		Scope:   SelfScope,
		Mutable: false,
	}
	st.store[name] = sym
	return sym
}

func (st *SymbolTable) Resolve(name string) (sym Symbol, level int, ok bool) {
	sym, level, ok = st.resolveSymbol(name, 0)
	return
}

func (st *SymbolTable) resolveSymbol(name string, fromLevel int) (
	sym Symbol, level int, ok bool) {

	level = fromLevel

	sym, ok = st.store[name]

	if !ok && st.Outer != nil {
		// not found in current symbol table; check outer symbol table
		sym, level, ok = st.Outer.resolveSymbol(name, fromLevel+1)

		if ok && st.IsFunction && !st.FreezeDefineFree {
			// resolves from beyond function border
			// define a "free" variable for this scope
			sym = st.defineFree(sym)
		}
	}

	return
}

func (st *SymbolTable) AddImpureEffects(s string) {
	// add to impurities list at the function level
	if st.IsFunction {
		if !str.IsInSlice(s, st.ImpureEffects) {
			st.ImpureEffects = append(st.ImpureEffects, s)
		}
	} else {
		if st.Outer != nil {
			st.Outer.AddImpureEffects(s)
		}
	}
}
