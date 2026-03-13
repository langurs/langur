// langur/symbol/symbol_table_test.go

package symbol

import (
	"langur/modes"
	"testing"
)

func TestDefine(t *testing.T) {
	expected := map[string]Symbol{
		"a": Symbol{Name: "a", Scope: GlobalScope, Index: 0},
		"b": Symbol{Name: "b", Scope: GlobalScope, Index: 1},
		"c": Symbol{Name: "c", Scope: LocalScope, Index: 0},
		"d": Symbol{Name: "d", Scope: LocalScope, Index: 1},
		"e": Symbol{Name: "e", Scope: LocalScope, Index: 0},
		"f": Symbol{Name: "f", Scope: LocalScope, Index: 1},
	}

	modes := modes.NewCompileModes()

	global := NewSymbolTable(nil, modes)

	a, err := global.DefineVariable("a", false, false)
	if err != nil {
		t.Errorf("Error defining symbol a: %s", err)
	}
	if a != expected["a"] {
		t.Errorf("expected a=%+v, received=%+v", expected["a"], a)
	}

	b, err := global.DefineVariable("b", false, false)
	if err != nil {
		t.Errorf("Error defining symbol b: %s", err)
	}
	if b != expected["b"] {
		t.Errorf("expected b=%+v, received=%+v", expected["b"], b)
	}

	firstLocal := NewSymbolTable(global, modes)
	firstLocal.IsFunction = true

	c, err := firstLocal.DefineVariable("c", false, false)
	if err != nil {
		t.Errorf("Error defining symbol c: %s", err)
	}
	if c != expected["c"] {
		t.Errorf("expected c=%+v, received=%+v", expected["c"], c)
	}

	d, err := firstLocal.DefineVariable("d", false, false)
	if err != nil {
		t.Errorf("Error defining symbol d: %s", err)
	}
	if d != expected["d"] {
		t.Errorf("expected d=%+v, received=%+v", expected["d"], d)
	}

	secondLocal := NewSymbolTable(firstLocal, modes)
	secondLocal.IsFunction = true

	e, err := secondLocal.DefineVariable("e", false, false)
	if err != nil {
		t.Errorf("Error defining symbol e: %s", err)
	}
	if e != expected["e"] {
		t.Errorf("expected e=%+v, received=%+v", expected["e"], e)
	}

	f, err := secondLocal.DefineVariable("f", false, false)
	if err != nil {
		t.Errorf("Error defining symbol f: %s", err)
	}
	if f != expected["f"] {
		t.Errorf("expected f=%+v, received=%+v", expected["f"], f)
	}
}

func TestResolveGlobal(t *testing.T) {
	modes := modes.NewCompileModes()
	global := NewSymbolTable(nil, modes)
	global.DefineVariable("a", false, false)
	global.DefineVariable("b", false, false)

	expected := []Symbol{
		Symbol{Name: "a", Scope: GlobalScope, Index: 0},
		Symbol{Name: "b", Scope: GlobalScope, Index: 1},
	}

	for _, sym := range expected {
		result, _, ok := global.Resolve(sym.Name)
		if !ok {
			t.Errorf("name %s not resolved", sym.Name)
			continue
		}
		if result != sym {
			t.Errorf("expected %s to resolve to %+v, received=%+v", sym.Name, sym, result)
		}
	}
}

func TestResolveLocal(t *testing.T) {
	modes := modes.NewCompileModes()
	global := NewSymbolTable(nil, modes)
	global.DefineVariable("a", false, false)
	global.DefineVariable("b", false, false)

	local := NewSymbolTable(global, modes)
	local.IsFunction = true
	local.DefineVariable("c", false, false)
	local.DefineVariable("d", false, false)

	expected := []Symbol{
		Symbol{Name: "a", Scope: FreeScope, Index: 0},
		Symbol{Name: "b", Scope: FreeScope, Index: 1},
		Symbol{Name: "c", Scope: LocalScope, Index: 0},
		Symbol{Name: "d", Scope: LocalScope, Index: 1},
	}

	for _, expectedSymbol := range expected {
		resultSymbol, _, ok := local.Resolve(expectedSymbol.Name)
		if !ok {
			t.Errorf("Name %s not resolvable", expectedSymbol.Name)
		}
		if resultSymbol != expectedSymbol {
			t.Errorf("Expeceted %s to resolve to %+v, received=%+v",
				expectedSymbol.Name, expectedSymbol, resultSymbol)
		}
	}
}

func TestResolveNestedLocal(t *testing.T) {
	modes := modes.NewCompileModes()

	global := NewSymbolTable(nil, modes)
	global.DefineVariable("a", false, false)
	global.DefineVariable("b", false, false)

	firstLocal := NewSymbolTable(global, modes)
	firstLocal.IsFunction = true
	firstLocal.DefineVariable("c", false, false)
	firstLocal.DefineVariable("d", false, false)

	secondLocal := NewSymbolTable(firstLocal, modes)
	secondLocal.IsFunction = true
	secondLocal.DefineVariable("e", false, false)
	secondLocal.DefineVariable("f", false, false)

	thirdLocal := NewSymbolTable(secondLocal, modes)
	thirdLocal.IsFunction = true
	thirdLocal.DefineVariable("a", false, false)
	thirdLocal.DefineVariable("g", false, false)

	tests := []struct {
		table           *SymbolTable
		expectedSymbols []Symbol
	}{
		{
			firstLocal,
			[]Symbol{
				Symbol{Name: "a", Scope: FreeScope, Index: 0},
				Symbol{Name: "b", Scope: FreeScope, Index: 1},

				Symbol{Name: "c", Scope: LocalScope, Index: 0},
				Symbol{Name: "d", Scope: LocalScope, Index: 1},
			},
		},
		{
			secondLocal,
			[]Symbol{
				Symbol{Name: "a", Scope: FreeScope, Index: 0},
				Symbol{Name: "b", Scope: FreeScope, Index: 1},
				Symbol{Name: "e", Scope: LocalScope, Index: 0},
				Symbol{Name: "f", Scope: LocalScope, Index: 1},
			},
		},
		{
			thirdLocal,
			[]Symbol{
				Symbol{Name: "a", Scope: LocalScope, Index: 0},
				Symbol{Name: "g", Scope: LocalScope, Index: 1},
				Symbol{Name: "e", Scope: FreeScope, Index: 0},
				Symbol{Name: "b", Scope: FreeScope, Index: 1},
			},
		},
	}

	for _, tt := range tests {
		for _, sym := range tt.expectedSymbols {
			result, _, ok := tt.table.Resolve(sym.Name)
			if !ok {
				t.Errorf("Name %s not resolvable", sym.Name)
				continue
			}
			if result != sym {
				t.Errorf("Expected %s to resolve to %+v, received=%+v",
					sym.Name, sym, result)
			}
		}
	}
}

func TestResolveFree(t *testing.T) {
	modes := modes.NewCompileModes()

	global := NewSymbolTable(nil, modes)
	global.DefineVariable("a", false, false)
	global.DefineVariable("b", false, false)

	firstLocal := NewSymbolTable(global, modes)
	firstLocal.IsFunction = true
	firstLocal.DefineVariable("c", false, false)
	firstLocal.DefineVariable("d", false, false)

	secondLocal := NewSymbolTable(firstLocal, modes)
	secondLocal.IsFunction = true
	secondLocal.DefineVariable("e", false, false)
	secondLocal.DefineVariable("f", false, false)

	tests := []struct {
		table           *SymbolTable
		expectedSymbols []Symbol
	}{
		{
			firstLocal,
			[]Symbol{
				Symbol{Name: "a", Scope: FreeScope, Index: 0},
				Symbol{Name: "b", Scope: FreeScope, Index: 1},
				Symbol{Name: "c", Scope: LocalScope, Index: 0},
				Symbol{Name: "d", Scope: LocalScope, Index: 1},
			},
		},
		{
			secondLocal,
			[]Symbol{
				Symbol{Name: "a", Scope: FreeScope, Index: 0},
				Symbol{Name: "b", Scope: FreeScope, Index: 1},
				Symbol{Name: "c", Scope: FreeScope, Index: 2},
				Symbol{Name: "d", Scope: FreeScope, Index: 3},
				Symbol{Name: "e", Scope: LocalScope, Index: 0},
				Symbol{Name: "f", Scope: LocalScope, Index: 1},
			},
		},
	}

	for _, tt := range tests {
		for _, sym := range tt.expectedSymbols {
			result, _, ok := tt.table.Resolve(sym.Name)
			if !ok {
				t.Errorf("name %s not resolvable", sym.Name)
				continue
			}
			if result != sym {
				t.Errorf("expected %s to resolve to %+v, got=%+v",
					sym.Name, sym, result)
			}
		}
	}
}

func TestResolveUnresolvableFree(t *testing.T) {
	modes := modes.NewCompileModes()

	global := NewSymbolTable(nil, modes)
	global.DefineVariable("a", false, false)

	firstLocal := NewSymbolTable(global, modes)
	firstLocal.IsFunction = true
	firstLocal.DefineVariable("c", false, false)

	secondLocal := NewSymbolTable(firstLocal, modes)
	secondLocal.IsFunction = true
	secondLocal.DefineVariable("e", false, false)
	secondLocal.DefineVariable("f", false, false)

	expected := []Symbol{
		Symbol{Name: "a", Scope: FreeScope, Index: 0},
		Symbol{Name: "c", Scope: FreeScope, Index: 1},
		Symbol{Name: "e", Scope: LocalScope, Index: 0},
		Symbol{Name: "f", Scope: LocalScope, Index: 1},
	}

	for _, sym := range expected {
		result, _, ok := secondLocal.Resolve(sym.Name)
		if !ok {
			t.Errorf("name %s not resolvable", sym.Name)
			continue
		}
		if result != sym {
			t.Errorf("expected %s to resolve to %+v, got=%+v",
				sym.Name, sym, result)
		}
	}

	expectedUnresolvable := []string{
		"b",
		"d",
	}

	for _, name := range expectedUnresolvable {
		_, _, ok := secondLocal.Resolve(name)
		if ok {
			t.Errorf("name %s resolved, but was expected not to", name)
		}
	}
}

func TestDefineAndResolveSelf(t *testing.T) {
	expected := Symbol{Name: "a", Scope: SelfScope, Index: 0}

	modes := modes.NewCompileModes()

	global := NewSymbolTable(nil, modes)
	global.DefineSelf("a")

	result, _, ok := global.Resolve(expected.Name)
	if !ok {
		t.Fatalf("self name %s not resolvable", expected.Name)
	}

	if result != expected {
		t.Errorf("expected %s to resolve to %+v, got=%+v",
			expected.Name, expected, result)
	}
}
