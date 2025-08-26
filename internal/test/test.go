// langur/test/test.go

package test

// for tests moved outside of packages to prevent import cycle

import (
	"langur/ast"
	"langur/lexer"
	"langur/parser"
	"testing"
)

func parse(t *testing.T, input string) *ast.Program {
	l, err := lexer.New(input, "test", nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	p := parser.New(l, nil)
	var program *ast.Program
	program, err = p.ParseProgram()
	if err != nil {
		t.Errorf(err.Error())
	}
	checkParseErrors(t, p, input)
	return program
}

func checkParseErrors(t *testing.T, p *parser.Parser, input string) {
	t.Helper()

	errors := p.Errs
	if len(errors) == 0 {
		return
	}

	t.Errorf("(%q)\nParser has %d errors.", input, len(errors))
	for _, msg := range errors {
		t.Errorf("Parser error: %q", msg)
	}
	t.FailNow()
}
