// langur/test/test.go

package test

// for tests moved outside of packages to prevent import cycle

import (
	"fmt"
	"langur/ast"
	"langur/lexer"
	"langur/object"
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


func testNumberObject(expected *object.Number, actual object.Object) error {
	result, ok := actual.(*object.Number)
	if !ok {
		return fmt.Errorf("object not a *Number, received=%T (%+v)", actual, actual)
	}
	if !expected.Same(actual) {
		return fmt.Errorf("object value wrong\nexpected=%s\nreceived=%s", expected.String(), result.String())
	}
	return nil
}

func testString(expected string, actual object.Object) error {
	switch actual.(type) {
	case *object.String:
		if actual.String() != expected {
			return fmt.Errorf("String value expected=%q, received=%q", expected, actual.String())
		}
		return nil
	case *object.Number:
		if actual.String() != expected {
			return fmt.Errorf("Number value expected=%q, received=%q", expected, actual.String())
		}
		return nil
	default:
		return fmt.Errorf("Object type %T (%+v) not expected", actual, actual)
	}
}
