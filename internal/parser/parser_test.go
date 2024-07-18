// langur/parser/parser_test.go

package parser

import (
	"fmt"
	"langur/ast"
	"langur/lexer"
	"langur/str"
	"langur/token"
	"strings"
	"testing"
)

// TODO: universal testing mechanism for the parser, ...
// ... like the testing mechanisms used for the compiler and VM

func TestListLiterals(t *testing.T) {
	input := "[1, 5 + 2, 7 * 7]"

	l, err := lexer.New(input, "test", nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	p := New(l, nil)
	var program *ast.Program
	program, err = p.ParseProgram()
	if err != nil {
		t.Errorf(err.Error())
	}
	checkParseErrors(t, p, input)

	stmt, ok := program.Statements[0].(*ast.ExpressionStatementNode)
	list, ok := stmt.Expression.(*ast.ListNode)
	if !ok {
		t.Fatalf("exp not an ast.ListLiteral, received=%T", stmt.Expression)
	}

	if len(list.Elements) != 3 {
		t.Fatalf("len(list.Elements != 3, received=%d", len(list.Elements))
	}

	testNumberLiteral(t, list.Elements[0], str.IntToStr(1, 10))
	testInfixExpression(t, list.Elements[1], 5, token.PLUS, 2)
	testInfixExpression(t, list.Elements[2], 7, token.ASTERISK, 7)
}

func TestParsingIndexExpressions(t *testing.T) {
	input := "yoList[1 + 1]"

	l, err := lexer.New(input, "test", nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	p := New(l, nil)
	var program *ast.Program
	program, err = p.ParseProgram()
	if err != nil {
		t.Errorf(err.Error())
	}
	checkParseErrors(t, p, input)

	if len(program.Statements) == 0 {
		t.Fatal("len(program.Statements) == 0")
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatementNode)
	if !ok {
		t.Fatalf("exp not *ast.IndexNode, received=%T", stmt)
	}

	indexExp, ok := stmt.Expression.(*ast.IndexNode)
	if !testVariable(t, indexExp.Left, "yoList") {
		return
	}

	if !testInfixExpression(t, indexExp.Index, 1, token.PLUS, 1) {
		return
	}
}

func TestStringExpression(t *testing.T) {
	expect := "yo don't you know"
	input := "\"" + expect + "\""

	l, err := lexer.New(input, "test", nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	p := New(l, nil)
	var program *ast.Program
	program, err = p.ParseProgram()
	if err != nil {
		t.Errorf(err.Error())
	}
	checkParseErrors(t, p, input)

	if len(program.Statements) != 1 {
		t.Fatalf("program statements != 1, received=%d", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatementNode)
	if !ok {
		t.Fatalf("program.Statements[0] not an *ast.ExpressionStatementNode, received=%T", program.Statements[0])
	}

	str, ok := stmt.Expression.(*ast.StringNode)
	if !ok {
		t.Fatalf("expression not an *ast.StringNode, received=%T", stmt.Expression)
	}
	// arr, ok := stmt.Expression.(*ast.ListNode)
	// if !ok {
	// 	t.Fatalf("expression not an *ast.ListNode, received=%T", stmt.Expression)
	// }
	// str, ok := arr.Elements[0].(*ast.StringNode)
	// if !ok {
	// 	t.Fatalf("arr.Elements[0] expression not an *ast.StringNode, received=%T", arr.Elements[0])
	// }

	if str.Values[0] != expect {
		t.Fatalf("string not \"%s\", received=\"%s\"", expect, str.Values[0])
	}

	if str.TokenRepresentation() != input {
		t.Fatalf("string token literal not %s, received=%s", input, str.TokenRepresentation())
	}
}

func TestVariableExpression(t *testing.T) {
	input := "einstein"

	l, err := lexer.New(input, "test", nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	p := New(l, nil)
	var program *ast.Program
	program, err = p.ParseProgram()
	if err != nil {
		t.Errorf(err.Error())
	}
	checkParseErrors(t, p, input)

	if len(program.Statements) != 1 {
		t.Fatalf("program statements != 1, received=%d", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatementNode)
	if !ok {
		t.Fatalf("program.Statements[0] not an *ast.ExpressionStatementNode, received=%T", program.Statements[0])
	}

	ident, ok := stmt.Expression.(*ast.IdentNode)
	if !ok {
		t.Fatalf("expression not an *ast.IdentNode, received=%T", stmt.Expression)
	}

	testVariable(t, ident, "einstein")
}

func TestNumberLiteralExpression(t *testing.T) {
	input := "7.7"

	l, err := lexer.New(input, "test", nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	p := New(l, nil)
	var program *ast.Program
	program, err = p.ParseProgram()
	if err != nil {
		t.Errorf(err.Error())
	}
	checkParseErrors(t, p, input)

	if len(program.Statements) != 1 {
		t.Fatalf("program statements != 1, received=%d", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatementNode)
	if !ok {
		t.Fatalf("program.Statements[0] not an *ast.ExpressionStatementNode, received=%T", program.Statements[0])
	}

	literal, ok := stmt.Expression.(*ast.NumberNode)
	if !ok {
		t.Fatalf("expression not an *ast.NumberNode, received=%T", stmt.Expression)
	}

	//	if literal.Value.String() != "7.7" {
	//		t.Errorf("literal.Value not 7.7, received=%s", literal.Value)
	//	}

	if literal.TokenRepresentation() != "7.7" {
		t.Errorf("literal.TokenRepresentation not 7.7, received=%s", literal.TokenRepresentation())
	}
}

func TestReturnStatements(t *testing.T) {
	tests := []struct {
		input         string
		expectedValue interface{}
	}{
		{"return 5", 5},
		{"return true;", true},
		{"return foobar;", "foobar"},
	}

	for _, tt := range tests {
		l, err := lexer.New(tt.input, "test", nil)
		if err != nil {
			t.Fatal(err.Error())
		}
		p := New(l, nil)
		var program *ast.Program
		program, err = p.ParseProgram()
		if err != nil {
			t.Errorf(err.Error())
		}
		checkParseErrors(t, p, tt.input)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements does not contain 1 statements. got=%d",
				len(program.Statements))
		}

		stmt := program.Statements[0]
		returnStmt, ok := stmt.(*ast.ReturnNode)
		if !ok {
			t.Fatalf("stmt not *ast.ReturnNode. got=%T", stmt)
		}
		//		if returnStmt.TokenRepresentation() != "return " {
		//			t.Fatalf("returnStmt.TokenRepresentation not 'return', got %q",
		//				returnStmt.TokenRepresentation())
		//		}
		if testLiteralExpression(t, returnStmt.ReturnValue, tt.expectedValue) {
			return
		}
	}
}

func TestDeclarations(t *testing.T) {
	tests := []struct {
		input         string
		expectedIdent string
		expectedValue interface{}
		expectMutable bool
	}{
		{"val x = 123", "x", 123, false},
		{"val y = true", "y", true, false},
		{"val yoyo = y;", "yoyo", "y", false},
		{"var x = 123", "x", 123, true},
	}

	for _, tt := range tests {
		l, err := lexer.New(tt.input, "test", nil)
		if err != nil {
			t.Fatal(err.Error())
		}
		p := New(l, nil)
		var program *ast.Program
		program, err = p.ParseProgram()
		if err != nil {
			t.Errorf(err.Error())
		}
		checkParseErrors(t, p, tt.input)

		if program == nil {
			t.Fatalf("ParseProgram() returned nil")
		}
		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements does not contain 1 statement. received=%d",
				len(program.Statements))
		}

		stmt := program.Statements[0].(*ast.ExpressionStatementNode).Expression

		decl, ok := stmt.(*ast.LineDeclarationNode)
		if !ok {
			t.Fatalf("stmt not *ast.LineDeclarationNode, received=%T", stmt)
		}
		if !testLiteralExpression(t, decl.Assignment.(*ast.AssignmentNode).Values[0], tt.expectedValue) {
			return
		}

		if !testAssignmentStatement(t, decl.Assignment, tt.expectedIdent) {
			return
		}

		if decl.Mutable != tt.expectMutable {
			t.Errorf("Expected mutable=%t, receieved=%t", tt.expectMutable, decl.Mutable)
		}
	}
}

func TestParsingPrefixExpressions(t *testing.T) {
	prefixTests := []struct {
		input           string
		operatorTokType token.Type
		value           interface{}
	}{
		{"not 7", token.NOT, 7},
		{"not foodbar;", token.NOT, "foodbar"},
		{"-foodbar;", token.MINUS, "foodbar"},
		{"not true;", token.NOT, true},
		{"not false;", token.NOT, false},
	}

	for _, tt := range prefixTests {
		l, err := lexer.New(tt.input, "test", nil)
		if err != nil {
			t.Fatal(err.Error())
		}
		p := New(l, nil)
		var program *ast.Program
		program, err = p.ParseProgram()
		if err != nil {
			t.Errorf(err.Error())
		}
		checkParseErrors(t, p, tt.input)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements does not contain %d statements. got=%d\n",
				1, len(program.Statements))
		}

		stmt, ok := program.Statements[0].(*ast.ExpressionStatementNode)
		if !ok {
			t.Fatalf("program.Statements[0] is not ast.ExpressionStatementNode. got=%T",
				program.Statements[0])
		}

		exp, ok := stmt.Expression.(*ast.PrefixExpressionNode)
		if !ok {
			t.Fatalf("stmt is not ast.PrefixExpressionNode. got=%T", stmt.Expression)
		}
		if exp.Operator.Type != tt.operatorTokType {
			t.Fatalf("exp.Operator is not '%s'. got=%s",
				token.TypeDescription(tt.operatorTokType), token.TypeDescription(exp.Operator.Type))
		}
		if !testLiteralExpression(t, exp.Right, tt.value) {
			return
		}
	}
}

func TestParsingInfixExpressions(t *testing.T) {
	infixTests := []struct {
		input      string
		leftValue  interface{}
		operator   token.Type
		rightValue interface{}
	}{
		{"7 + 70;", 7, token.PLUS, 70},
		{"7 - 70;", 7, token.MINUS, 70},
		{"7 * 70;", 7, token.ASTERISK, 70},
		{"7 / 70;", 7, token.SLASH, 70},
		{"7 > 70;", 7, token.GREATER_THAN, 70},
		{"7 < 70;", 7, token.LESS_THAN, 70},
		{"7 == 70;", 7, token.EQUAL, 70},
		{"7 != 70;", 7, token.NOT_EQUAL, 70},
		{"fooey + nofoobars;", "fooey", token.PLUS, "nofoobars"},
		{"fooey - nofoobars;", "fooey", token.MINUS, "nofoobars"},
		{"fooey * nofoobars;", "fooey", token.ASTERISK, "nofoobars"},
		{"fooey / nofoobars;", "fooey", token.SLASH, "nofoobars"},
		{"fooey > nofoobars;", "fooey", token.GREATER_THAN, "nofoobars"},
		{"fooey < nofoobars;", "fooey", token.LESS_THAN, "nofoobars"},
		{"fooey == nofoobars;", "fooey", token.EQUAL, "nofoobars"},
		{"fooey != nofoobars;", "fooey", token.NOT_EQUAL, "nofoobars"},
		{"true == true;", true, token.EQUAL, true},
		{"true != false;", true, token.NOT_EQUAL, false},
		{"false == false;", false, token.EQUAL, false},
	}

	for _, tt := range infixTests {
		l, err := lexer.New(tt.input, "test", nil)
		if err != nil {
			t.Fatal(err.Error())
		}
		p := New(l, nil)
		var program *ast.Program
		program, err = p.ParseProgram()
		if err != nil {
			t.Errorf(err.Error())
		}
		checkParseErrors(t, p, tt.input)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements does not contain %d statements. got=%d\n",
				1, len(program.Statements))
		}

		stmt, ok := program.Statements[0].(*ast.ExpressionStatementNode)
		if !ok {
			t.Fatalf("program.Statements[0] is not ast.ExpressionStatementNode. got=%T",
				program.Statements[0])
		}

		if !testInfixExpression(t, stmt.Expression, tt.leftValue,
			tt.operator, tt.rightValue) {
			return
		}
	}
}

func TestHashLiterals(t *testing.T) {
	input := `{"one": 1, "two": 2, "three": 3}`

	l, err := lexer.New(input, "test", nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	p := New(l, nil)
	var program *ast.Program
	program, err = p.ParseProgram()
	if err != nil {
		t.Errorf(err.Error())
	}
	checkParseErrors(t, p, input)

	stmt := program.Statements[0].(*ast.ExpressionStatementNode)
	hash, ok := stmt.Expression.(*ast.HashNode)
	if !ok {
		t.Fatalf("exp not *ast.HashNode, received=%T", stmt.Expression)
	}

	if len(hash.Pairs) != 3 {
		t.Errorf("hash.Pairs has wrong length, received=%d", len(hash.Pairs))
	}

	expected := map[string]int64{
		"one":   1,
		"two":   2,
		"three": 3,
	}

	for _, kv := range hash.Pairs {
		literal, ok := kv.Key.(*ast.StringNode)
		if !ok {
			t.Errorf("key not an *ast.StringNode, received=%T", kv.Key)
		}

		expectedValue := expected[literal.Values[0]]

		testNumberLiteral(t, kv.Value, str.Int64ToStr(expectedValue, 10))
	}
}

func TestParsingHashLiteralsWithExpressions(t *testing.T) {
	input := `{"one": 0 + 1, "two": 9 - 7, "three": 21 / 7}`

	l, err := lexer.New(input, "test", nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	p := New(l, nil)
	var program *ast.Program
	program, err = p.ParseProgram()
	if err != nil {
		t.Errorf(err.Error())
	}
	checkParseErrors(t, p, input)

	stmt := program.Statements[0].(*ast.ExpressionStatementNode)
	hash, ok := stmt.Expression.(*ast.HashNode)
	if !ok {
		t.Fatalf("exp not *ast.HashNode, received=%T", stmt.Expression)
	}

	if len(hash.Pairs) != 3 {
		t.Errorf("hash.Pairs has wrong length, received=%d", len(hash.Pairs))
	}

	tests := map[string]func(ast.Node){
		"one": func(e ast.Node) {
			testInfixExpression(t, e, 0, token.PLUS, 1)
		},
		"two": func(e ast.Node) {
			testInfixExpression(t, e, 9, token.MINUS, 7)
		},
		"three": func(e ast.Node) {
			testInfixExpression(t, e, 21, token.SLASH, 7)
		},
	}

	for _, kv := range hash.Pairs {
		literal, ok := kv.Key.(*ast.StringNode)
		if !ok {
			t.Errorf("key not an *ast.StringNode, received=%T", kv.Key)
			continue
		}

		testFn, ok := tests[literal.Values[0]]
		if !ok {
			t.Errorf("test function for key %q not found", literal.String())
			continue
		}

		testFn(kv.Value)
	}
}

func TestOperatorPrecedenceAndAssociativityParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			"-a * b;",
			"((-a) * b)",
		},
		{
			"not -a;",
			"(not (-a))",
		},
		{
			"a + b + c;",
			"((a + b) + c)",
		},
		{
			"a + b rem c;",
			"(a + (b rem c))",
		},
		{
			"a + b - c;",
			"((a + b) - c)",
		},
		{
			"a * b * c;",
			"((a * b) * c)",
		},
		{
			"a * b / c;",
			"((a * b) / c)",
		},
		{
			"a + b / c;",
			"(a + (b / c))",
		},
		{
			`a + b * c + 
				d / e - f`,
			"(((a + (b * c)) + (d / e)) - f)",
		},
		// {
		// 	`3 + 4
		// 	-5 * 5`,
		// 	"(3 + 4); ((-5) * 5)",
		// },
		{
			`3 + 4
			-5 * 5`,
			"(3 + 4); (-5 * 5)",
		},
		{
			"5 > 4 == 3 < 4;",
			"((5 > 4) == (3 < 4))",
		},
		{
			"5 < 4 != 3 > 4;",
			"((5 < 4) != (3 > 4))",
		},
		{
			"3 + 4 * 5 == 3 * 1 + 4 * 5;",
			"((3 + (4 * 5)) == ((3 * 1) + (4 * 5)))",
		},

		{
			"true;",
			"true",
		},
		{
			"false;",
			"false",
		},
		{
			"3 > 5 == false;",
			"((3 > 5) == false)",
		},
		{
			"3 < 5 == true;",
			"((3 < 5) == true)",
		},

		{
			"1 + (2 + 3) + 4;",
			"((1 + (2 + 3)) + 4)",
		},
		{
			"(5 + 5) * 2;",
			"((5 + 5) * 2)",
		},
		{
			"2 / (5 + 5);",
			"(2 / (5 + 5))",
		},
		{
			"(5 + 5) * 2 * (5 + 5);",
			"(((5 + 5) * 2) * (5 + 5))",
		},
		{
			"-(5 + 5);",
			"(-(5 + 5))",
		},

		// exponent and root to be right associative
		// exponent with higher precedence than negation a change as of 0.12
		{
			"-7 ^ 2",
			"(-(7 ^ 2))",
		},
		// {
		// 	"(-7) ^ 2",
		// 	"((-7) ^ 2)",
		// },
		{
			"(-7) ^ 2",
			"(-7 ^ 2)",
		},

		// negative number at start of open argument list
		{
			"split -3, qs'abc'",
			`split(-3, "abc")`,
		},

		{
			"7 ^ 2 ^ 3",
			"(7 ^ (2 ^ 3))",
		},
		{
			"7 * 2 ^ 3",
			"(7 * (2 ^ 3))",
		},
		{
			"7 + 2 ^ 3 ^/ 3",
			"(7 + (2 ^ (3 ^/ 3)))",
		},

		{
			"not true == false;",
			"((not true) == false)",
		},
		{
			"not (true == true);",
			"(not (true == true))",
		},

		{
			"not 5 > 3",
			"(not (5 > 3))",
		},
		{
			"not 5 == true",
			"((not 5) == true)",
		},

		{
			"a + add(b * c) + d;",
			"((a + add((b * c))) + d)",
		},
		{
			"add(a, b, 1, 2 * 3, 4 + 5, add(6, 7 * 8));",
			"add(a, b, 1, (2 * 3), (4 + 5), add(6, (7 * 8)))",
		},
		{
			"add(a + b + c * d / f + g);",
			"add((((a + b) + ((c * d) / f)) + g))",
		},
		{
			"a * [1, 2, 3, 4][b * c] * d",
			"((a * ([1, 2, 3, 4][(b * c)])) * d)",
		},
		{
			"add(a * b[2], b[1], 2 * [1, 2][1])",
			"add((a * (b[2])), (b[1]), (2 * ([1, 2][1])))",
		},
	}

	for _, tt := range tests {
		//	for i, tt := range tests {
		//		fmt.Printf("%d: %s\n", i, tt.input)

		l, err := lexer.New(tt.input, "test", nil)
		if err != nil {
			t.Fatal(err.Error())
		}
		p := New(l, nil)
		var program *ast.Program
		program, err = p.ParseProgram()
		if err != nil {
			t.Errorf(err.Error())
		}

		checkParseErrors(t, p, tt.input)

		actual := program.TokenRepresentation()
		if actual != tt.expected {
			t.Errorf("expected=%q, received=%q", tt.expected, actual)
		}
	}
}

func TestParsingIfExpression(t *testing.T) {
	input := "if (x < y) { x }"

	l, err := lexer.New(input, "test", nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	p := New(l, nil)
	var program *ast.Program
	program, err = p.ParseProgram()
	if err != nil {
		t.Errorf(err.Error())
	}
	checkParseErrors(t, p, input)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Body does not contain 1 statement, received=%d", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatementNode)
	if !ok {
		t.Fatalf("program.Statements[0] not *ast.ExpressionStatementNode, received=%T", program.Statements[0])
	}

	exp, ok := stmt.Expression.(*ast.IfNode)
	if !ok {
		t.Fatalf("stmt.Expression not *ast.IfNode, received=%T", stmt.Expression)
	}

	if strings.Replace(stmt.Expression.TokenRepresentation(), "(scoped)", "", -1) != input {
		t.Errorf("stmt.Expression.TokenRepresentation() is different than input\n%s\n%s", stmt.Expression.TokenRepresentation(), input)
	}

	if !testInfixExpression(t, exp.TestsAndActions[0].Test, "x", token.LESS_THAN, "y") {
		return
	}

	if len(exp.TestsAndActions[0].Do.(*ast.BlockNode).Statements) != 1 {
		t.Errorf("consequence not 1 statement, received=%d", len(exp.TestsAndActions[0].Do.(*ast.BlockNode).Statements))
	}

	consequence, ok := exp.TestsAndActions[0].Do.(*ast.BlockNode).Statements[0].(*ast.ExpressionStatementNode)
	if !ok {
		t.Fatalf("Statements[0] not *ast.ExpressionStatementNode, received=%T", exp.TestsAndActions[0].Do.(*ast.BlockNode).Statements[0])
	}

	if !testVariable(t, consequence.Expression, "x") {
		return
	}

	if len(exp.TestsAndActions) > 1 {
		t.Fatal("Parsed Too Many Sections for If/Else")
	}
}

func TestParsingIfElseExpression(t *testing.T) {
	input := `if (x < y) { x } else { y }`

	l, err := lexer.New(input, "test", nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	p := New(l, nil)
	var program *ast.Program
	program, err = p.ParseProgram()
	if err != nil {
		t.Errorf(err.Error())
	}

	checkParseErrors(t, p, input)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Body does not contain %d statements. got=%d\n",
			1, len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatementNode)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ExpressionStatementNode. got=%T",
			program.Statements[0])
	}

	exp, ok := stmt.Expression.(*ast.IfNode)
	if !ok {
		t.Fatalf("stmt.Expression is not ast.IfNode. got=%T", stmt.Expression)
	}

	if strings.Replace(stmt.Expression.TokenRepresentation(), "(scoped)", "", -1) != input {
		t.Errorf("stmt.Expression.TokenRepresentation() is different than input\n%s\n%s", stmt.Expression.TokenRepresentation(), input)
	}

	if !testInfixExpression(t, exp.TestsAndActions[0].Test, "x", token.LESS_THAN, "y") {
		return
	}

	if len(exp.TestsAndActions[0].Do.(*ast.BlockNode).Statements) != 1 {
		t.Errorf("consequence is not 1 statements. got=%d\n",
			len(exp.TestsAndActions[0].Do.(*ast.BlockNode).Statements))
	}

	consequence, ok := exp.TestsAndActions[0].Do.(*ast.BlockNode).Statements[0].(*ast.ExpressionStatementNode)
	if !ok {
		t.Fatalf("Statements[0] is not ast.ExpressionStatementNode. got=%T",
			exp.TestsAndActions[0].Do.(*ast.BlockNode).Statements[0])
	}

	if !testVariable(t, consequence.Expression, "x") {
		return
	}

	if len(exp.TestsAndActions) < 2 {
		t.Fatal("Failed to Parse ELSE section")
	}
	if len(exp.TestsAndActions) > 2 {
		t.Fatal("Parsed Too Many Sections for If/Else")
	}

	if exp.TestsAndActions[1].Test != nil {
		t.Errorf("exp.TestsAndActions[1].Test not nil, received=%+v", exp.TestsAndActions[1].Test)
	}

	if len(exp.TestsAndActions[1].Do.(*ast.BlockNode).Statements) != 1 {
		t.Errorf("exp.TestsAndActions[1].Do.(*ast.BlockNode).Statements does not contain 1 statements. got=%d\n",
			len(exp.TestsAndActions[1].Do.(*ast.BlockNode).Statements))
	}

	alternative, ok := exp.TestsAndActions[1].Do.(*ast.BlockNode).Statements[0].(*ast.ExpressionStatementNode)
	if !ok {
		t.Fatalf("exp.TestsAndActions[1].Do.(*ast.BlockNode).Statements[0] is not ast.ExpressionStatementNode. got=%T",
			exp.TestsAndActions[1].Do.(*ast.BlockNode).Statements[0])
	}

	if !testVariable(t, alternative.Expression, "y") {
		return
	}
}

func TestParsingElseIfExpression(t *testing.T) {
	input := "if (10 < 5) { 1 } else if (5 > 10) { 2 } else if (1 == 2) { x } else { 4 }"

	l, err := lexer.New(input, "test", nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	p := New(l, nil)
	var program *ast.Program
	program, err = p.ParseProgram()
	if err != nil {
		t.Errorf(err.Error())
	}
	checkParseErrors(t, p, input)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Body does not contain %d statements. received=%d\n",
			1, len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatementNode)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ExpressionStatementNode. received=%T",
			program.Statements[0])
	}

	exp, ok := stmt.Expression.(*ast.IfNode)
	if !ok {
		t.Fatalf("stmt.Expression is not ast.IfNode. received=%T", stmt.Expression)
	}

	if strings.Replace(stmt.Expression.TokenRepresentation(), "(scoped)", "", -1) != input {
		t.Errorf("stmt.Expression.TokenRepresentation() is different than input\n%s\n%s", stmt.Expression.TokenRepresentation(), input)
	}

	if len(exp.TestsAndActions) != 4 {
		t.Errorf("len(exp.TestsAndActions) != 4, received=%d\n", len(exp.TestsAndActions))
	}

	if !testInfixExpression(t, exp.TestsAndActions[0].Test, 10, token.LESS_THAN, 5) {
		return
	}
	if !testInfixExpression(t, exp.TestsAndActions[1].Test, 5, token.GREATER_THAN, 10) {
		return
	}
	if !testInfixExpression(t, exp.TestsAndActions[2].Test, 1, token.EQUAL, 2) {
		return
	}
	if exp.TestsAndActions[3].Test != nil {
		t.Errorf("4th test not nil")
	}

	exp1, ok := exp.TestsAndActions[0].Do.(*ast.BlockNode).Statements[0].(*ast.ExpressionStatementNode)
	if !ok {
		t.Fatalf("exp.TestsAndActions[0].Do not an *ast.ExpressionStatementNode, received=%T", exp1)
	}
	if exp1.Expression.TokenRepresentation() != "1" {
		t.Errorf("Expression 1 not \"1\", received=%q", exp1.Expression.TokenRepresentation())
	}

	exp2, ok := exp.TestsAndActions[1].Do.(*ast.BlockNode).Statements[0].(*ast.ExpressionStatementNode)
	if !ok {
		t.Fatalf("exp.TestsAndActions[1].Do not an *ast.ExpressionStatementNode, received=%T", exp1)
	}
	if exp2.Expression.TokenRepresentation() != "2" {
		t.Errorf("Expression 2 not \"2\", received=%q", exp1.Expression.TokenRepresentation())
	}

	exp3, ok := exp.TestsAndActions[2].Do.(*ast.BlockNode).Statements[0].(*ast.ExpressionStatementNode)
	if !ok {
		t.Fatalf("exp.TestsAndActions[2].Do not an *ast.ExpressionStatementNode, received=%T", exp1)
	}
	if exp3.Expression.TokenRepresentation() != "x" {
		t.Errorf("Expression 3 not \"x\", received=%q", exp1.Expression.TokenRepresentation())
	}

	exp4, ok := exp.TestsAndActions[3].Do.(*ast.BlockNode).Statements[0].(*ast.ExpressionStatementNode)
	if !ok {
		t.Fatalf("exp.TestsAndActions[3].Do not an *ast.ExpressionStatementNode, received=%T", exp1)
	}
	if exp4.Expression.TokenRepresentation() != "4" {
		t.Errorf("Expression 4 not \"4\", received=%q", exp1.Expression.TokenRepresentation())
	}
}

func TestParsingCallExpression(t *testing.T) {
	input := `add(
				1, 2 * 3, 
				4 + 5)`

	l, err := lexer.New(input, "test", nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	p := New(l, nil)
	var program *ast.Program
	program, err = p.ParseProgram()
	if err != nil {
		t.Errorf(err.Error())
	}
	checkParseErrors(t, p, input)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements != 1; received=%d", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatementNode)
	if !ok {
		t.Fatalf("program.Statements[0] not an *ast.ExpressionStatment; received=%T", program.Statements[0])
	}

	exp, ok := stmt.Expression.(*ast.CallNode)
	if !ok {
		t.Fatalf("stmt.Expression not an *ast.CallNode; received=%T", stmt.Expression)
	}

	if !testVariable(t, exp.Function, "add") {
		return
	}

	if len(exp.Args) != 3 {
		t.Fatalf("wrong number of args; received=%d", len(exp.Args))
	}

	testLiteralExpression(t, exp.Args[0], 1)
	testInfixExpression(t, exp.Args[1], 2, token.ASTERISK, 3)
	testInfixExpression(t, exp.Args[2], 4, token.PLUS, 5)
}

func TestParsingCallExpressionParameters(t *testing.T) {
	tests := []struct {
		input         string
		expectedIdent string
		expectedArgs  []string
	}{
		{
			input:         "add();",
			expectedIdent: "add",
			expectedArgs:  []string{},
		},
		{
			input:         "add(1);",
			expectedIdent: "add",
			expectedArgs:  []string{"1"},
		},
		{
			input:         "add(1, 2 * 3, 4 + 5);",
			expectedIdent: "add",
			expectedArgs:  []string{"1", "(2 * 3)", "(4 + 5)"},
		},
	}

	for _, tt := range tests {
		l, err := lexer.New(tt.input, "test", nil)
		if err != nil {
			t.Fatal(err.Error())
		}
		p := New(l, nil)
		var program *ast.Program
		program, err = p.ParseProgram()
		if err != nil {
			t.Errorf(err.Error())
		}
		checkParseErrors(t, p, tt.input)

		stmt := program.Statements[0].(*ast.ExpressionStatementNode)
		exp, ok := stmt.Expression.(*ast.CallNode)
		if !ok {
			t.Fatalf("stmt.Expression is not ast.CallNode. got=%T",
				stmt.Expression)
		}

		if !testVariable(t, exp.Function, tt.expectedIdent) {
			return
		}

		if len(exp.Args) != len(tt.expectedArgs) {
			t.Fatalf("wrong number of arguments. want=%d, got=%d",
				len(tt.expectedArgs), len(exp.Args))
		}

		for i, arg := range tt.expectedArgs {
			if exp.Args[i].TokenRepresentation() != arg {
				t.Errorf("argument %d wrong. want=%q, got=%q", i,
					arg, exp.Args[i].TokenRepresentation())
			}
		}
	}
}

func TestParsingFunctionLiteral(t *testing.T) {
	input := `fn(x, y) {
				 x + y
				}`

	l, err := lexer.New(input, "test", nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	p := New(l, nil)
	var program *ast.Program
	program, err = p.ParseProgram()
	if err != nil {
		t.Errorf(err.Error())
	}
	checkParseErrors(t, p, input)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Body does not contain 1 statement; received=%d", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatementNode)
	if !ok {
		t.Fatalf("program.Statements[0] not an *ast.ExpressionStatementNode; received=%T", program.Statements[0])
	}

	function, ok := stmt.Expression.(*ast.FunctionNode)
	if !ok {
		t.Fatalf("stmt.Expression not an *ast.FunctionNode; received=%T", stmt.Expression)
	}

	if len(function.Parameters) != 2 {
		t.Fatalf("parameter count not 2; received=%d", len(function.Parameters))
	}

	testLiteralExpression(t, function.Parameters[0], "x")
	testLiteralExpression(t, function.Parameters[1], "y")

	if len(function.Body.(*ast.BlockNode).Statements) != 1 {
		t.Fatalf("function.Body.(*ast.BlockNode).Statements != 1; received=%d", len(function.Body.(*ast.BlockNode).Statements))
	}

	bodyStmt, ok := function.Body.(*ast.BlockNode).Statements[0].(*ast.ExpressionStatementNode)
	if !ok {
		t.Fatalf("function.Body.(*ast.BlockNode).Statements[0] not an *ast.ExpressionStatementNode; received=%T", function.Body.(*ast.BlockNode).Statements[0])
	}

	testInfixExpression(t, bodyStmt.Expression, "x", token.PLUS, "y")
}

func TestParsingFunctionLiteralParameters(t *testing.T) {
	tests := []struct {
		input          string
		expectedParams []string
	}{
		{input: "fn{}", expectedParams: []string{}},
		{input: "fn(x) {}", expectedParams: []string{"x"}},
		{input: "fn(x, y, z) {}", expectedParams: []string{"x", "y", "z"}},
		{input: "fn(a, b) {a < b}", expectedParams: []string{"a", "b"}},

		{input: "fn x: {}", expectedParams: []string{"x"}},
		{input: "fn x, y, z: {}", expectedParams: []string{"x", "y", "z"}},
		{input: "fn a, b: a < b", expectedParams: []string{"a", "b"}},
	}

	for _, tt := range tests {
		l, err := lexer.New(tt.input, "test", nil)
		if err != nil {
			t.Fatal(err.Error())
		}
		p := New(l, nil)
		var program *ast.Program
		program, err = p.ParseProgram()
		if err != nil {
			t.Errorf(err.Error())
		}
		checkParseErrors(t, p, tt.input)

		stmt := program.Statements[0].(*ast.ExpressionStatementNode)
		function := stmt.Expression.(*ast.FunctionNode)

		if len(function.Parameters) != len(tt.expectedParams) {
			t.Errorf("parameter len not as expected; expected=%d; received=%d", len(tt.expectedParams), len(function.Parameters))

		} else {
			for i, ident := range tt.expectedParams {
				testLiteralExpression(t, function.Parameters[i], ident)
			}
		}
	}
}

func TestFunctionDefinitionParsing(t *testing.T) {
	input := `val yofunction = fn() { }`

	l, err := lexer.New(input, "test", nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	p := New(l, nil)
	var program *ast.Program
	program, err = p.ParseProgram()
	if err != nil {
		t.Errorf(err.Error())
	}
	checkParseErrors(t, p, input)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Body does not contain %d statements. got=%d\n",
			1, len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatementNode)
	if !ok {
		t.Fatalf("program.Statements[0] is not *ast.ExpressionStatementNode. got=%T",
			program.Statements[0])
	}

	decl, ok := stmt.Expression.(*ast.LineDeclarationNode)
	if !ok {
		t.Fatalf("program.Statements[0].(*ast.ExpressionStatementNode) not *ast.LineDeclarationNode. got=%T",
			stmt.Expression)
	}

	assign, ok := decl.Assignment.(*ast.AssignmentNode)
	if !ok {
		t.Fatalf("decl.Assignment not *ast.AssignmentNode, received=%T", decl.Assignment)
	}

	function, ok := assign.Values[0].(*ast.FunctionNode)
	if !ok {
		t.Fatalf("assign.Values[0] is not *ast.FunctionNode. got=%T",
			assign.Values[0])
	}

	if function.Name != "yofunction" {
		t.Fatalf("function literal name wrong. want 'yofunction', got=%q\n",
			function.Name)
	}
}

func TestParsingTryCatch(t *testing.T) {
	// There's no such thing as an explicit (semantic) "try" for this.
	// You catch on preceding statements in a block and they are automatically wrapped into a try.
	input := `
		val x = 123 / 0
		catch[err] { if err["cat"] == "MATH" { 890 } else { 456 } }
		`

	l, err := lexer.New(input, "test", nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	p := New(l, nil)
	//	prog, err := p.ParseProgram()
	_, err = p.ParseProgram()
	if err != nil {
		t.Errorf(err.Error())
	}

	checkParseErrors(t, p, input)

	// TODO: more testing
}

func checkParseErrors(t *testing.T, p *Parser, input string) {
	t.Helper()

	errors := p.Errs
	if len(errors) == 0 {
		return
	}

	t.Errorf("(%s)\nParser has %d errors.", input, len(errors))
	for _, msg := range errors {
		t.Errorf("Parser error: %q", msg)
	}
	t.FailNow()
}

func testAssignmentStatement(t *testing.T, s ast.Node, name string) bool {
	assignExpr, ok := s.(*ast.AssignmentNode)
	if !ok {
		t.Errorf("s not *ast.AssignmentNode, received=%T", s)
		return false
	}

	if assignExpr.Identifiers[0].(*ast.IdentNode).Name != name {
		t.Errorf("assignExpr.Identifiers[0].Name not '%s', received='%s'",
			name, assignExpr.Identifiers[0].(*ast.IdentNode).Name)
		return false
	}

	if assignExpr.Identifiers[0].TokenRepresentation() != name {
		t.Errorf("s.Identifiers[0].TokenRepresentation() not %q, received %q", name,
			assignExpr.Identifiers[0].TokenRepresentation())
		return false
	}

	return true
}

func testLiteralExpression(t *testing.T, exp ast.Node, expected interface{}) bool {
	switch v := expected.(type) {
	case bool:
		return testBooleanLiteral(t, exp, v)
	case int:
		return testNumberLiteral(t, exp, str.IntToStr(v, 10))
	case int64:
		return testNumberLiteral(t, exp, str.Int64ToStr(v, 10))
	case string:
		return testVariable(t, exp, v)
	}
	t.Errorf("Type of exp not handled, received=%T", exp)
	return false
}

func testBooleanLiteral(t *testing.T, exp ast.Node, value bool) bool {
	bo, ok := exp.(*ast.BooleanNode)
	if !ok {
		t.Errorf("exp not *ast.BooleanNode, received=%T", exp)
		return false
	}

	if bo.Value != value {
		t.Errorf("Boolean value not %t", value)
		return false
	}

	if bo.TokenRepresentation() != fmt.Sprintf("%t", value) {
		t.Errorf("Boolean TokenRepresentation() != %t, received=%s", value, bo.TokenRepresentation())
		return false
	}

	return true
}

func testNumberLiteral(t *testing.T, nl ast.Node, value string) bool {
	n, ok := nl.(*ast.NumberNode)
	if !ok {
		t.Errorf("fl not *ast.NumberNode, received=%T", nl)
		return false
	}

	if n.TokenRepresentation() != value {
		t.Errorf("f.TokenRepresentation not %s, received=%s", value, n.TokenRepresentation())
		return false
	}

	return true
}

func testVariable(t *testing.T, exp ast.Node, name string) bool {
	t.Helper()

	ident, ok := exp.(*ast.IdentNode)
	if !ok {
		t.Errorf("exp not *ast.IdentNode, received=%T", exp)
		return false
	}

	if ident.Name != name {
		t.Errorf("ident.Name not %s, received=%s", name, ident.Name)
		return false
	}

	if ident.TokenRepresentation() != name {
		t.Errorf("ident.TokenRepresentation() not %s, received=%s", name, ident.TokenRepresentation())
		return false
	}

	return true
}

func testInfixExpression(
	t *testing.T, exp ast.Node,
	left interface{}, operatorTokType token.Type, right interface{},
) bool {

	opExp, ok := exp.(*ast.InfixExpressionNode)
	if !ok {
		t.Errorf("exp not an *ast.InfixExpressionNode, received=%T(%s)", exp, exp)
		return false
	}

	if !testLiteralExpression(t, opExp.Left, left) {
		return false
	}

	if opExp.Operator.Type != operatorTokType {
		t.Errorf("exp.Operator not %s, received=%q",
			token.TypeDescription(operatorTokType),
			token.TypeDescription(opExp.Operator.Type))
		return false
	}

	if !testLiteralExpression(t, opExp.Right, right) {
		return false
	}

	return true
}
