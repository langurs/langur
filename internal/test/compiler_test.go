// langur/test/compiler_test.go

package test

import (
	"fmt"
	"langur/ast"
	"langur/object"
	"langur/opcode"
	"langur/vm/process"
	"testing"
)

type compilerTestCase struct {
	input                string
	expectedConstants    []interface{}
	expectedInstructions []opcode.Instructions
}

func runCompilerTests(
	t *testing.T,
	tests []compilerTestCase,
	printTestFirst bool) {

	t.Helper()

	for _, tt := range tests {
		if printTestFirst {
			fmt.Printf("Test: %q\n", tt.input)
		}

		program := parse(t, tt.input)

		compiler, err := ast.NewCompiler(nil, false)
		if err != nil {
			t.Fatalf("compiler error on New: %s", err)
		}
		_, err = program.CompileAnother(compiler)
		if err != nil {
			t.Fatalf("compiler error: %s", err)
		}

		byteCode := compiler.ByteCode()

		err = testInstructions(tt.expectedInstructions, byteCode.StartCode.InsPackage.Instructions)
		if err != nil {
			t.Fatalf("(%q) testInstructions failed: %s", tt.input, err)
		}

		err = testConstants(t, tt.expectedConstants, byteCode.Constants)
		if err != nil {
			t.Fatalf("(%q) testConstants failed: %s", tt.input, err)
		}
	}
}

func testInstructions(
	expected []opcode.Instructions,
	actual opcode.Instructions,
) error {
	appended := concatInstructions(expected)

	if len(actual) != len(appended) {
		return fmt.Errorf("instruction length wrong\nexpected=%q\nreceived=%q", appended, actual)
	}

	for i, ins := range appended {
		if actual[i] != ins {
			return fmt.Errorf("instruction wrong at %d\nexpected=%q\nreceived=%q",
				i, appended, actual)
		}
	}

	return nil
}

func concatInstructions(s []opcode.Instructions) opcode.Instructions {
	out := opcode.Instructions{}

	for _, ins := range s {
		out = append(out, ins...)
	}

	return out
}

func testConstants(
	t *testing.T,
	expected []interface{},
	actual []object.Object,
) error {
	if len(expected) != len(actual) {
		return fmt.Errorf("wrong constants count; expected=%d, received=%d", len(expected), len(actual))
	}

	for i, constant := range expected {
		switch constant := constant.(type) {
		case int:
			err := testNumberObject(object.NumberFromInt(constant), actual[i])
			if err != nil {
				return fmt.Errorf("Constant %d: testIntegerObject failed: %s", i, err)
			}

		case string:
			err := testString(constant, actual[i])
			if err != nil {
				return fmt.Errorf("Constant %d: testStringObject failed: %s", i, err)
			}

		case []opcode.Instructions:
			fn, ok := actual[i].(*object.CompiledCode)
			if !ok {
				return fmt.Errorf("Constant %d not *object.CompiledCode, received %T", i, actual[i])
			}
			err := testInstructions(constant, fn.InsPackage.Instructions)
			if err != nil {
				return fmt.Errorf("Constant %d testInstructions failed: %s", i, err)
			}
		}
	}

	return nil
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

func TestCompilerOpCodeSequencing(t *testing.T) {
	tests := []compilerTestCase{
		{
			// testing OpPop...
			input:             "1; 2",
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpPop),
				opcode.Make(opcode.OpConstant, 1),
				opcode.Make(opcode.OpPop),
			},
		},

		{
			input:             "1 + 2",
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpConstant, 1),
				opcode.Make(opcode.OpAdd),
				opcode.Make(opcode.OpPop),
			},
		},

		{
			input:             "1 - 2",
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpConstant, 1),
				opcode.Make(opcode.OpSubtract),
				opcode.Make(opcode.OpPop),
			},
		},

		{
			input:             "1 * 2",
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpConstant, 1),
				opcode.Make(opcode.OpMultiply),
				opcode.Make(opcode.OpPop),
			},
		},

		{
			input:             "1 / 2",
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpConstant, 1),
				opcode.Make(opcode.OpDivide),
				opcode.Make(opcode.OpPop),
			},
		},

		{
			input:             "1 rem 2",
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpConstant, 1),
				opcode.Make(opcode.OpRemainder),
				opcode.Make(opcode.OpPop),
			},
		},

		{
			input:             `1 \ 2`,
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpConstant, 1),
				opcode.Make(opcode.OpTruncateDivide),
				opcode.Make(opcode.OpPop),
			},
		},

		{
			input:             "(1 < 2) == true",
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpConstant, 1),
				opcode.Make(opcode.OpLessThan, 0, 0),
				opcode.Make(opcode.OpTrue),
				opcode.Make(opcode.OpEqual, 0, 0),
				opcode.Make(opcode.OpPop),
			},
		},

		{
			input:             "-1",
			expectedConstants: []interface{}{-1},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpPop),
			},
		},
		{
			// prefix minus test
			input:             "-(1)",
			expectedConstants: []interface{}{-1},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpPop),
			},
		},
		{
			input:             "-(-(1))",
			expectedConstants: []interface{}{1},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpPop),
			},
		},

		{
			// repeated constant added once
			input:             "7; 7",
			expectedConstants: []interface{}{7},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpPop),
				opcode.Make(opcode.OpConstant, 0), // same constant number (0)
				opcode.Make(opcode.OpPop),
			},
		},
	}

	runCompilerTests(t, tests, false)
}

func TestCompilerBooleanExpressions(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "true",
			expectedConstants: []interface{}{},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpTrue),
				opcode.Make(opcode.OpPop),
			},
		},
		{
			input:             "false",
			expectedConstants: []interface{}{},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpFalse),
				opcode.Make(opcode.OpPop),
			},
		},

		{
			input:             "1 > 2",
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpConstant, 1),
				opcode.Make(opcode.OpGreaterThan, 0, 0),
				opcode.Make(opcode.OpPop),
			},
		},
		{
			input:             "1 < 2",
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpConstant, 1),
				opcode.Make(opcode.OpLessThan, 0, 0),
				opcode.Make(opcode.OpPop),
			},
		},
		{
			input:             "1 >= 2",
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpConstant, 1),
				opcode.Make(opcode.OpGreaterThanOrEqual, 0, 0),
				opcode.Make(opcode.OpPop),
			},
		},
		{
			input:             "1 <= 2",
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpConstant, 1),
				opcode.Make(opcode.OpLessThanOrEqual, 0, 0),
				opcode.Make(opcode.OpPop),
			},
		},

		{
			input:             "1 == 2",
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpConstant, 1),
				opcode.Make(opcode.OpEqual, 0, 0),
				opcode.Make(opcode.OpPop),
			},
		},
		{
			input:             "1 != 2",
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpConstant, 1),
				opcode.Make(opcode.OpNotEqual, 0, 0),
				opcode.Make(opcode.OpPop),
			},
		},

		{
			input:             "true == false",
			expectedConstants: []interface{}{},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpTrue),
				opcode.Make(opcode.OpFalse),
				opcode.Make(opcode.OpEqual, 0, 0),
				opcode.Make(opcode.OpPop),
			},
		},
		{
			input:             "true != false",
			expectedConstants: []interface{}{},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpTrue),
				opcode.Make(opcode.OpFalse),
				opcode.Make(opcode.OpNotEqual, 0, 0),
				opcode.Make(opcode.OpPop),
			},
		},
		{
			input:             "true and 1 != false",
			expectedConstants: []interface{}{1},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpTrue),
				opcode.Make(opcode.OpLogicalAnd, 0, 12), // short-circuit: test first side of and; jump 12 if result found
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpFalse),
				opcode.Make(opcode.OpNotEqual, 0, 0),
				opcode.Make(opcode.OpLogicalAnd, 0, 0),
				opcode.Make(opcode.OpPop),
			},
		},

		{
			// prefix negation
			input:             "not true",
			expectedConstants: []interface{}{},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpTrue),
				opcode.Make(opcode.OpLogicalNegation, 0),
				opcode.Make(opcode.OpPop),
			},
		},
	}

	runCompilerTests(t, tests, false)
}

func TestCompilerConditionals(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "if true {10}; 777;",
			expectedConstants: []interface{}{10, 777},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpTrue),
				opcode.Make(opcode.OpJumpIfNotTruthy, 8),
				opcode.Make(opcode.OpConstant, 0), // 3
				opcode.Make(opcode.OpJump, 1),     // 3
				opcode.Make(opcode.OpNull),
				opcode.Make(opcode.OpPop),
				opcode.Make(opcode.OpConstant, 1),
				opcode.Make(opcode.OpPop),
			},
		},

		{
			input:             "if true {10} else {20}",
			expectedConstants: []interface{}{10, 20},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpTrue),
				opcode.Make(opcode.OpJumpIfNotTruthy, 8),
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpJump, 3),
				opcode.Make(opcode.OpConstant, 1),
				opcode.Make(opcode.OpPop),
			},
		},

		{
			input:             "if true {10} else {20}; 777;",
			expectedConstants: []interface{}{10, 20, 777},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpTrue),
				opcode.Make(opcode.OpJumpIfNotTruthy, 8),
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpJump, 3),
				opcode.Make(opcode.OpConstant, 1),
				opcode.Make(opcode.OpPop),
				opcode.Make(opcode.OpConstant, 2),
				opcode.Make(opcode.OpPop),
			},
		},

		{
			input:             "if true {10} else if false {20}",
			expectedConstants: []interface{}{10, 20},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpTrue),
				opcode.Make(opcode.OpJumpIfNotTruthy, 8),
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpJump, 15),
				opcode.Make(opcode.OpFalse),
				opcode.Make(opcode.OpJumpIfNotTruthy, 8),
				opcode.Make(opcode.OpConstant, 1),
				opcode.Make(opcode.OpJump, 1),
				opcode.Make(opcode.OpNull),
				opcode.Make(opcode.OpPop),
			},
		},

		{
			input:             "if true {10} else if false {20} else {30}",
			expectedConstants: []interface{}{10, 20, 30},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpTrue),
				opcode.Make(opcode.OpJumpIfNotTruthy, 8),
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpJump, 17),
				opcode.Make(opcode.OpFalse),
				opcode.Make(opcode.OpJumpIfNotTruthy, 8),
				opcode.Make(opcode.OpConstant, 1),
				opcode.Make(opcode.OpJump, 3),
				opcode.Make(opcode.OpConstant, 2),
				opcode.Make(opcode.OpPop),
			},
		},

		{
			input:             "val x = 123; val y = 123; if x == 120 and y == 120 { 456 } else { 890 }",
			expectedConstants: []interface{}{123, 120, 456, 890},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0), // 123
				opcode.Make(opcode.OpSetGlobal, 0),
				opcode.Make(opcode.OpPop),
				opcode.Make(opcode.OpConstant, 0), // 123
				opcode.Make(opcode.OpSetGlobal, 1),
				opcode.Make(opcode.OpPop),
				opcode.Make(opcode.OpGetGlobal, 0), // x
				opcode.Make(opcode.OpConstant, 1),  // 120
				opcode.Make(opcode.OpEqual, 0, 0),
				opcode.Make(opcode.OpLogicalAnd, 0, 14), // short-circuit: jump 14 if answer known
				opcode.Make(opcode.OpGetGlobal, 1),      // y
				opcode.Make(opcode.OpConstant, 1),       // 120
				opcode.Make(opcode.OpEqual, 0, 0),
				opcode.Make(opcode.OpLogicalAnd, 0, 0), // no short-circuit now (jump 0)
				opcode.Make(opcode.OpJumpIfNotTruthy, 8),
				opcode.Make(opcode.OpConstant, 2), // 456
				opcode.Make(opcode.OpJump, 3),
				opcode.Make(opcode.OpConstant, 3), // 890
				opcode.Make(opcode.OpPop),
			},
		},
	}

	runCompilerTests(t, tests, false)
}

func TestCompilerSwitchExpressions(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "val x = true; switch x { case true: 10 }; 777",
			expectedConstants: []interface{}{10, 777},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpTrue),
				opcode.Make(opcode.OpSetGlobal, 0),
				opcode.Make(opcode.OpPop),
				opcode.Make(opcode.OpGetGlobal, 0), // beginning of switch
				opcode.Make(opcode.OpTrue),         // case true
				opcode.Make(opcode.OpEqual, 0, 0),
				opcode.Make(opcode.OpJumpIfNotTruthy, 8),
				opcode.Make(opcode.OpConstant, 0), // 10
				opcode.Make(opcode.OpJump, 1),
				opcode.Make(opcode.OpNull), // no "default"
				opcode.Make(opcode.OpPop),
				opcode.Make(opcode.OpConstant, 1), // 777
				opcode.Make(opcode.OpPop),
			},
		},

		// with implicit fallthrough ("case alternate")
		{input: `val x = 123
			switch x { case 100:	# matches 100 or 123, result 2
					case 123: 2;
					case 200: 3 }`,
			expectedConstants: []interface{}{123, 100, 200, 2, 3},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),  // 123
				opcode.Make(opcode.OpSetGlobal, 0), // x
				opcode.Make(opcode.OpPop),

				opcode.Make(opcode.OpGetGlobal, 0),       // x
				opcode.Make(opcode.OpConstant, 1),        // 100
				opcode.Make(opcode.OpEqual, 0, 0),        // if x == 100
				opcode.Make(opcode.OpJumpIfNotTruthy, 5), // x != 100, do next test

				// implicit fallthrough
				opcode.Make(opcode.OpJump, 15), // x == 100, skip next test to next action

				opcode.Make(opcode.OpGetGlobal, 0), // x
				opcode.Make(opcode.OpConstant, 0),  // 123
				opcode.Make(opcode.OpEqual, 0, 0),  // if x == 123
				opcode.Make(opcode.OpJumpIfNotTruthy, 8),

				opcode.Make(opcode.OpConstant, 3), // 2
				opcode.Make(opcode.OpJump, 24),

				opcode.Make(opcode.OpGetGlobal, 0),
				opcode.Make(opcode.OpConstant, 2), // 200
				opcode.Make(opcode.OpEqual, 0, 0),
				opcode.Make(opcode.OpJumpIfNotTruthy, 8),

				opcode.Make(opcode.OpConstant, 4), // 3
				opcode.Make(opcode.OpJump, 1),

				opcode.Make(opcode.OpNull), // no matches
				opcode.Make(opcode.OpPop),
			},
		},
	}

	runCompilerTests(t, tests, false)
}

func TestCompilerTryCatch(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: `val x = 123 / 0
					catch { if _err["cat"] == "MATH" { 890 } else { 456 } }
					`,

			expectedConstants: []interface{}{
				123, 0,
				[]opcode.Instructions{ // try frame
					opcode.Make(opcode.OpConstant, 0),  // 123
					opcode.Make(opcode.OpConstant, 1),  // 0
					opcode.Make(opcode.OpDivide),       // 123 / 0
					opcode.Make(opcode.OpSetGlobal, 0), // x = 123 / 0
					// no pop here
				},
				"cat", "MATH", 890, 456,
				[]opcode.Instructions{ // catch frame
					opcode.Make(opcode.OpSetLocal, 0), // catch _err {
					opcode.Make(opcode.OpPop),
					opcode.Make(opcode.OpGetLocal, 0),
					opcode.Make(opcode.OpConstant, 3),
					opcode.Make(opcode.OpIndex, 0),
					opcode.Make(opcode.OpConstant, 4),
					opcode.Make(opcode.OpEqual, 0, 0),
					opcode.Make(opcode.OpJumpIfNotTruthy, 8),
					opcode.Make(opcode.OpConstant, 5),
					opcode.Make(opcode.OpJump, 3),
					opcode.Make(opcode.OpConstant, 6),
					// no pop here
				},
			},

			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpTryCatch, 2, 7, 0),
				opcode.Make(opcode.OpPop),
			},
		},
	}

	runCompilerTests(t, tests, false)
}

func TestCompilerTryCatchElse(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: `val x = 123 / 0
					catch {
						if _err["cat"] == "MATH" { 890 } else { 456 }
					} else {
						159
					}
					`,

			expectedConstants: []interface{}{
				123, 0,
				[]opcode.Instructions{ // try frame
					opcode.Make(opcode.OpConstant, 0),  // 123
					opcode.Make(opcode.OpConstant, 1),  // 0
					opcode.Make(opcode.OpDivide),       // 123 / 0
					opcode.Make(opcode.OpSetGlobal, 0), // x = 123 / 0
					// no pop here
				},
				"cat", "MATH", 890, 456,
				[]opcode.Instructions{ // catch frame
					opcode.Make(opcode.OpSetLocal, 0), // catch _err {
					opcode.Make(opcode.OpPop),
					opcode.Make(opcode.OpGetLocal, 0),
					opcode.Make(opcode.OpConstant, 3),
					opcode.Make(opcode.OpIndex, 0),
					opcode.Make(opcode.OpConstant, 4),
					opcode.Make(opcode.OpEqual, 0, 0),
					opcode.Make(opcode.OpJumpIfNotTruthy, 8),
					opcode.Make(opcode.OpConstant, 5),
					opcode.Make(opcode.OpJump, 3),
					opcode.Make(opcode.OpConstant, 6),
					// no pop here
				},
				159,
				[]opcode.Instructions{ // else frame
					opcode.Make(opcode.OpConstant, 8),
				},
			},

			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpTryCatch, 2, 7, 9),
				opcode.Make(opcode.OpPop),
			},
		},
	}

	runCompilerTests(t, tests, false)
}

func TestCompilerGlobalDeclarationStatements(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "val one = 1; val two = 2;",
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpSetGlobal, 0),
				opcode.Make(opcode.OpPop),
				opcode.Make(opcode.OpConstant, 1),
				opcode.Make(opcode.OpSetGlobal, 1),
				opcode.Make(opcode.OpPop),
			},
		},

		{
			input:             "val one = 1; one;",
			expectedConstants: []interface{}{1},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpSetGlobal, 0),
				opcode.Make(opcode.OpPop),
				opcode.Make(opcode.OpGetGlobal, 0),
				opcode.Make(opcode.OpPop),
			},
		},

		{
			input:             "val one = 1; val two = one; two;",
			expectedConstants: []interface{}{1},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpSetGlobal, 0),
				opcode.Make(opcode.OpPop),
				opcode.Make(opcode.OpGetGlobal, 0),
				opcode.Make(opcode.OpSetGlobal, 1),
				opcode.Make(opcode.OpPop),
				opcode.Make(opcode.OpGetGlobal, 1),
				opcode.Make(opcode.OpPop),
			},
		},
	}

	runCompilerTests(t, tests, false)
}

func TestCompilerStringExpressions(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             `"langur"`,
			expectedConstants: []interface{}{"langur"},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpPop),
			},
		},

		{
			input:             `"lan" ~ "gur"`,
			expectedConstants: []interface{}{"lan", "gur"},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpConstant, 1),
				opcode.Make(opcode.OpAppend, 0),
				opcode.Make(opcode.OpPop),
			},
		},
	}

	runCompilerTests(t, tests, false)
}

func TestCompilerStringInterpolation(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             `val x = 123; "abc {{x}} joe"`,
			expectedConstants: []interface{}{123, "abc ", " joe"},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),  // 123
				opcode.Make(opcode.OpSetGlobal, 0), // x
				opcode.Make(opcode.OpPop),

				opcode.Make(opcode.OpConstant, 1),  // "abc "
				opcode.Make(opcode.OpGetGlobal, 0), // x, 123
				opcode.Make(opcode.OpConstant, 2),  // " joe"
				opcode.Make(opcode.OpString, 3),
				opcode.Make(opcode.OpPop),
			},
		},

		{
			input:             `val x = 123; "{{x}} "`,
			expectedConstants: []interface{}{123, " "},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),  // 123
				opcode.Make(opcode.OpSetGlobal, 0), // x
				opcode.Make(opcode.OpPop),

				opcode.Make(opcode.OpGetGlobal, 0), // x
				opcode.Make(opcode.OpConstant, 1),  // " "
				opcode.Make(opcode.OpString, 2),
				opcode.Make(opcode.OpPop),
			},
		},

		{
			input:             `val x, y = 123, 4567; "{{x}}{{y}}"`,
			expectedConstants: []interface{}{4567, 123},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),  // 4567
				opcode.Make(opcode.OpConstant, 1),  // 123
				opcode.Make(opcode.OpSetGlobal, 0), // x
				opcode.Make(opcode.OpPop),
				opcode.Make(opcode.OpSetGlobal, 1), // y
				opcode.Make(opcode.OpPop),

				opcode.Make(opcode.OpGetGlobal, 0), // x
				opcode.Make(opcode.OpGetGlobal, 1), // y
				opcode.Make(opcode.OpString, 2),
				opcode.Make(opcode.OpPop),
			},
		},

		{
			input:             `val x = "yes"; "answer: {{x}}"`,
			expectedConstants: []interface{}{"yes", "answer: "},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),  // "yes"
				opcode.Make(opcode.OpSetGlobal, 0), // val x = ...
				opcode.Make(opcode.OpPop),          // ;
				opcode.Make(opcode.OpConstant, 1),  // "answer: "
				opcode.Make(opcode.OpGetGlobal, 0), // x
				opcode.Make(opcode.OpString, 2),
				opcode.Make(opcode.OpPop),
			},
		},

		{
			input:             `val x = 7; val y = 42; "x: {{x}}, y: {{y}}, x + y: {{x + y}} si"`,
			expectedConstants: []interface{}{7, 42, "x: ", ", y: ", ", x + y: ", " si"},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),  // 7
				opcode.Make(opcode.OpSetGlobal, 0), // val x = ...
				opcode.Make(opcode.OpPop),
				opcode.Make(opcode.OpConstant, 1),  // 42
				opcode.Make(opcode.OpSetGlobal, 1), // val y = ...
				opcode.Make(opcode.OpPop),
				opcode.Make(opcode.OpConstant, 2),
				opcode.Make(opcode.OpGetGlobal, 0),
				opcode.Make(opcode.OpConstant, 3),
				opcode.Make(opcode.OpGetGlobal, 1),
				opcode.Make(opcode.OpConstant, 4),
				opcode.Make(opcode.OpGetGlobal, 0),
				opcode.Make(opcode.OpGetGlobal, 1),
				opcode.Make(opcode.OpAdd),
				opcode.Make(opcode.OpConstant, 5),
				opcode.Make(opcode.OpString, 7),
				opcode.Make(opcode.OpPop),
			},
		},
	}

	runCompilerTests(t, tests, false)
}

func TestCompilerListLiterals(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: "[]",
			expectedConstants: []interface{}{
				object.EmptyList,
			},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0), // empty list as constant
				opcode.Make(opcode.OpPop),
			},
		},
		{
			input:             "[1, 2, 3]",
			expectedConstants: []interface{}{1, 2, 3},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpConstant, 1),
				opcode.Make(opcode.OpConstant, 2),
				opcode.Make(opcode.OpList, 3),
				opcode.Make(opcode.OpPop),
			},
		},
		{
			input:             "[1 + 2, 3 * 4, 5 - 6, 7]",
			expectedConstants: []interface{}{1, 2, 3, 4, 5, 6, 7},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpConstant, 1),
				opcode.Make(opcode.OpAdd),
				opcode.Make(opcode.OpConstant, 2),
				opcode.Make(opcode.OpConstant, 3),
				opcode.Make(opcode.OpMultiply),
				opcode.Make(opcode.OpConstant, 4),
				opcode.Make(opcode.OpConstant, 5),
				opcode.Make(opcode.OpSubtract),
				opcode.Make(opcode.OpConstant, 6),
				opcode.Make(opcode.OpList, 4),
				opcode.Make(opcode.OpPop),
			},
		},
		{
			input:             "[[1, 2, 3], 2, 3]",
			expectedConstants: []interface{}{1, 2, 3},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpConstant, 1),
				opcode.Make(opcode.OpConstant, 2),
				opcode.Make(opcode.OpList, 3),
				opcode.Make(opcode.OpConstant, 1),
				opcode.Make(opcode.OpConstant, 2),
				opcode.Make(opcode.OpList, 3),
				opcode.Make(opcode.OpPop),
			},
		},
		{
			input:             "[[1, 2, 3]]",
			expectedConstants: []interface{}{1, 2, 3},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpConstant, 1),
				opcode.Make(opcode.OpConstant, 2),
				opcode.Make(opcode.OpList, 3),
				opcode.Make(opcode.OpList, 1),
				opcode.Make(opcode.OpPop),
			},
		},
		{
			input:             "[1, [1, 2, 4], 2]",
			expectedConstants: []interface{}{1, 2, 4},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpConstant, 1),
				opcode.Make(opcode.OpConstant, 2),
				opcode.Make(opcode.OpList, 3),
				opcode.Make(opcode.OpConstant, 1),
				opcode.Make(opcode.OpList, 3),
				opcode.Make(opcode.OpPop),
			},
		},
	}

	runCompilerTests(t, tests, false)
}

func TestCompilerHashLiterals(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "{1: 2, 3: 4, 5: 7}",
			expectedConstants: []interface{}{1, 2, 3, 4, 5, 7},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpConstant, 1),
				opcode.Make(opcode.OpConstant, 2),
				opcode.Make(opcode.OpConstant, 3),
				opcode.Make(opcode.OpConstant, 4),
				opcode.Make(opcode.OpConstant, 5),
				opcode.Make(opcode.OpHash, 6),
				opcode.Make(opcode.OpPop),
			},
		},
		{
			input:             "{1: 2 + 3, 4: 5 + 7}",
			expectedConstants: []interface{}{1, 2, 3, 4, 5, 7},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpConstant, 1),
				opcode.Make(opcode.OpConstant, 2),
				opcode.Make(opcode.OpAdd),
				opcode.Make(opcode.OpConstant, 3),
				opcode.Make(opcode.OpConstant, 4),
				opcode.Make(opcode.OpConstant, 5),
				opcode.Make(opcode.OpAdd),
				opcode.Make(opcode.OpHash, 4),
				opcode.Make(opcode.OpPop),
			},
		},
	}

	runCompilerTests(t, tests, false)
}

func TestCompilerIndexExpressions(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "[7, 14, 21][4 - 1]",
			expectedConstants: []interface{}{7, 14, 21, 4, 1},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpConstant, 1),
				opcode.Make(opcode.OpConstant, 2),
				opcode.Make(opcode.OpList, 3),
				opcode.Make(opcode.OpConstant, 3),
				opcode.Make(opcode.OpConstant, 4),
				opcode.Make(opcode.OpSubtract),
				opcode.Make(opcode.OpIndex, 0),
				opcode.Make(opcode.OpPop),
			},
		},
		{
			input:             "{1: 123, 2: 789}[4 - 2]",
			expectedConstants: []interface{}{1, 123, 2, 789, 4},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0), // 1
				opcode.Make(opcode.OpConstant, 1), // 123
				opcode.Make(opcode.OpConstant, 2), // 2
				opcode.Make(opcode.OpConstant, 3), // 789
				opcode.Make(opcode.OpHash, 4),
				opcode.Make(opcode.OpConstant, 4), // 4
				opcode.Make(opcode.OpConstant, 2), // 2 (added earlier)
				opcode.Make(opcode.OpSubtract),
				opcode.Make(opcode.OpIndex, 0),
				opcode.Make(opcode.OpPop),
			},
		},

		// alternate values for invalid indices
		{
			input:             "[7, 14, 21][4 - 1; 123]",
			expectedConstants: []interface{}{7, 14, 21, 4, 1, 123},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpConstant, 1),
				opcode.Make(opcode.OpConstant, 2),
				opcode.Make(opcode.OpList, 3),
				opcode.Make(opcode.OpConstant, 3),
				opcode.Make(opcode.OpConstant, 4),
				opcode.Make(opcode.OpSubtract),
				opcode.Make(opcode.OpIndex, 3),
				opcode.Make(opcode.OpConstant, 5), // 123
				opcode.Make(opcode.OpPop),
			},
		},
		{
			input:             "{1: 123, 2: 789}[4 - 2; 123]",
			expectedConstants: []interface{}{1, 123, 2, 789, 4},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0), // 1
				opcode.Make(opcode.OpConstant, 1), // 123
				opcode.Make(opcode.OpConstant, 2), // 2
				opcode.Make(opcode.OpConstant, 3), // 789
				opcode.Make(opcode.OpHash, 4),
				opcode.Make(opcode.OpConstant, 4), // 4
				opcode.Make(opcode.OpConstant, 2), // 2 (added earlier)
				opcode.Make(opcode.OpSubtract),
				opcode.Make(opcode.OpIndex, 3),
				opcode.Make(opcode.OpConstant, 1), // 123
				opcode.Make(opcode.OpPop),
			},
		},

		// index alternates with short-circuiting...
		{
			input:             "[1, 2, 3][7; 3 * 4]",
			expectedConstants: []interface{}{1, 2, 3, 7, 4},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0), // 1
				opcode.Make(opcode.OpConstant, 1), // 2
				opcode.Make(opcode.OpConstant, 2), // 3
				opcode.Make(opcode.OpList, 3),
				opcode.Make(opcode.OpConstant, 3), // 7
				opcode.Make(opcode.OpIndex, 7),
				opcode.Make(opcode.OpConstant, 2), // 3
				opcode.Make(opcode.OpConstant, 4), // 4
				opcode.Make(opcode.OpMultiply),    // 3 * 4
				opcode.Make(opcode.OpPop),
			},
		},
	}

	runCompilerTests(t, tests, false)
}

func TestCompilerRangeExpressions(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "1 .. 3",
			expectedConstants: []interface{}{1, 3},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpConstant, 1),
				opcode.Make(opcode.OpRange),
				opcode.Make(opcode.OpPop),
			},
		},

		{
			input:             "1 + 2 .. 7",
			expectedConstants: []interface{}{1, 2, 7},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpConstant, 1),
				opcode.Make(opcode.OpAdd),
				opcode.Make(opcode.OpConstant, 2),
				opcode.Make(opcode.OpRange),
				opcode.Make(opcode.OpPop),
			},
		},

		{
			input:             "1 .. 4 * 7",
			expectedConstants: []interface{}{1, 4, 7},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpConstant, 1),
				opcode.Make(opcode.OpConstant, 2),
				opcode.Make(opcode.OpMultiply),
				opcode.Make(opcode.OpRange),
				opcode.Make(opcode.OpPop),
			},
		},
	}

	runCompilerTests(t, tests, false)
}

func TestCompilerFunctionDefinitions(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: "fn() { return 7 * 3 }",
			expectedConstants: []interface{}{
				7, 3,
				[]opcode.Instructions{
					opcode.Make(opcode.OpConstant, 0),
					opcode.Make(opcode.OpConstant, 1),
					opcode.Make(opcode.OpMultiply),
					opcode.Make(opcode.OpReturnValue),
				},
			},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 2),
				opcode.Make(opcode.OpPop),
			},
		},

		{
			input: "fn() { 7 * 3 }",
			expectedConstants: []interface{}{
				7, 3,
				[]opcode.Instructions{
					opcode.Make(opcode.OpConstant, 0),
					opcode.Make(opcode.OpConstant, 1),
					opcode.Make(opcode.OpMultiply),
					opcode.Make(opcode.OpReturnValue),
				},
			},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 2),
				opcode.Make(opcode.OpPop),
			},
		},

		{
			input: "fn() { 7; 3 }",
			expectedConstants: []interface{}{
				7, 3,
				[]opcode.Instructions{
					opcode.Make(opcode.OpConstant, 0),
					opcode.Make(opcode.OpPop),
					opcode.Make(opcode.OpConstant, 1),
					opcode.Make(opcode.OpReturnValue),
				},
			},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 2),
				opcode.Make(opcode.OpPop),
			},
		},

		{
			input: "fn() {}",
			expectedConstants: []interface{}{
				[]opcode.Instructions{
					opcode.Make(opcode.OpNull),
					opcode.Make(opcode.OpReturnValue),
				},
			},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpPop),
			},
		},
	}

	runCompilerTests(t, tests, false)
}

func TestCompilerFunctionCalls(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: "fn() { 21 }()",
			expectedConstants: []interface{}{
				21,
				[]opcode.Instructions{
					opcode.Make(opcode.OpConstant, 0), // 21
					opcode.Make(opcode.OpReturnValue),
				},
			},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 1), // compiled function literal
				opcode.Make(opcode.OpCall, 0, 0),
				opcode.Make(opcode.OpPop),
			},
		},

		{
			input: "val noArg = fn() { 21 } \n noArg()",
			expectedConstants: []interface{}{
				21,
				[]opcode.Instructions{
					opcode.Make(opcode.OpConstant, 0), // 21
					opcode.Make(opcode.OpReturnValue),
				},
			},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 1), // compiled function literal
				opcode.Make(opcode.OpSetGlobal, 0),
				opcode.Make(opcode.OpPop),
				opcode.Make(opcode.OpGetGlobal, 0),
				opcode.Make(opcode.OpCall, 0, 0),
				opcode.Make(opcode.OpPop),
			},
		},

		{
			input: `
				val oneArg = fn(a) { a };
				oneArg(21);
				`,
			expectedConstants: []interface{}{
				[]opcode.Instructions{
					opcode.Make(opcode.OpGetLocal, 0),
					opcode.Make(opcode.OpReturnValue),
				},
				21,
			},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpSetGlobal, 0),
				opcode.Make(opcode.OpPop),
				opcode.Make(opcode.OpConstant, 1),
				opcode.Make(opcode.OpGetGlobal, 0),
				opcode.Make(opcode.OpCall, 1, 0),
				opcode.Make(opcode.OpPop),
			},
		},

		{
			input: `
				val manyArg = fn(a, b, c) { a; b; c + c };
				manyArg(21, 28, 35);
				`,
			expectedConstants: []interface{}{
				[]opcode.Instructions{
					opcode.Make(opcode.OpGetLocal, 0),
					opcode.Make(opcode.OpPop),
					opcode.Make(opcode.OpGetLocal, 1),
					opcode.Make(opcode.OpPop),
					opcode.Make(opcode.OpGetLocal, 2),
					opcode.Make(opcode.OpGetLocal, 2),
					opcode.Make(opcode.OpAdd),
					opcode.Make(opcode.OpReturnValue),
				},
				21,
				28,
				35,
			},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpSetGlobal, 0),
				opcode.Make(opcode.OpPop),
				opcode.Make(opcode.OpConstant, 1),
				opcode.Make(opcode.OpConstant, 2),
				opcode.Make(opcode.OpConstant, 3),
				opcode.Make(opcode.OpGetGlobal, 0),
				opcode.Make(opcode.OpCall, 3, 0),
				opcode.Make(opcode.OpPop),
			},
		},
	}

	runCompilerTests(t, tests, false)
}

func TestCompilerAssignmentStatementScopes(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: `
			val num = 55
			fn() { num }
			`,
			expectedConstants: []interface{}{
				55,
				[]opcode.Instructions{
					opcode.Make(opcode.OpGetFree, 0),
					opcode.Make(opcode.OpReturnValue),
				},
			},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpSetGlobal, 0),
				opcode.Make(opcode.OpPop),
				opcode.Make(opcode.OpGetGlobal, 0),
				opcode.Make(opcode.OpFunction, 1, 1, 0),
				opcode.Make(opcode.OpPop),
			},
		},

		{
			input: `
			fn() {
				val num = 55
				num
			}
			`,
			expectedConstants: []interface{}{
				55,
				[]opcode.Instructions{
					opcode.Make(opcode.OpConstant, 0),
					opcode.Make(opcode.OpSetLocal, 0),
					opcode.Make(opcode.OpPop),
					opcode.Make(opcode.OpGetLocal, 0),
					opcode.Make(opcode.OpReturnValue),
				},
			},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 1),
				opcode.Make(opcode.OpPop),
			},
		},

		{
			input: `
			fn() {
				val a = 55
				val b = 77
				a + b
			}
			`,
			expectedConstants: []interface{}{
				55,
				77,
				[]opcode.Instructions{
					opcode.Make(opcode.OpConstant, 0),
					opcode.Make(opcode.OpSetLocal, 0),
					opcode.Make(opcode.OpPop),
					opcode.Make(opcode.OpConstant, 1),
					opcode.Make(opcode.OpSetLocal, 1),
					opcode.Make(opcode.OpPop),
					opcode.Make(opcode.OpGetLocal, 0),
					opcode.Make(opcode.OpGetLocal, 1),
					opcode.Make(opcode.OpAdd),
					opcode.Make(opcode.OpReturnValue),
				},
			},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 2),
				opcode.Make(opcode.OpPop),
			},
		},

		{
			input: `val x = 7; { val x = 14 }`,
			expectedConstants: []interface{}{
				7, 14,
				[]opcode.Instructions{ // { val x = 14 }
					opcode.Make(opcode.OpConstant, 1), // 14
					opcode.Make(opcode.OpSetLocal, 0),
				},
			},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0), // 7
				opcode.Make(opcode.OpSetGlobal, 0),
				opcode.Make(opcode.OpPop),        // ; {
				opcode.Make(opcode.OpExecute, 2), // { val x = 14 }
				opcode.Make(opcode.OpPop),
			},
		},

		{ // scope on each section of if expression
			input: `if x = 14 { x }`,
			expectedConstants: []interface{}{
				14,
				[]opcode.Instructions{
					opcode.Make(opcode.OpConstant, 0),
					opcode.Make(opcode.OpSetLocal, 0),
					opcode.Make(opcode.OpJumpRelayIfNotTruthy, 5, 1),
					opcode.Make(opcode.OpGetLocal, 0),
				},
			},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpExecute, 1),
				opcode.Make(opcode.OpJump, 1),
				opcode.Make(opcode.OpNull),
				opcode.Make(opcode.OpPop),
			},
		},
		{ // only block within scoped
			input: `if true { val x = 21 }`,
			expectedConstants: []interface{}{
				21,
				[]opcode.Instructions{ // { val x = 21 }
					opcode.Make(opcode.OpConstant, 0),
					opcode.Make(opcode.OpSetLocal, 0),
				},
			},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpTrue),
				opcode.Make(opcode.OpJumpIfNotTruthy, 8),
				opcode.Make(opcode.OpExecute, 1),
				opcode.Make(opcode.OpJump, 1),
				opcode.Make(opcode.OpNull),
				opcode.Make(opcode.OpPop),
			},
		},

		{ // decoupling assignment within if expression test
			input: `if x, y = [] { x }`,
			expectedConstants: []interface{}{
				object.EmptyList,
				process.GetBuiltInByName("_len"), 2, 1,
				[]opcode.Instructions{ // decoupling scope (uses internal temporary variable)
					opcode.Make(opcode.OpConstant, 0), // []
					opcode.Make(opcode.OpSetLocal, 0), // val _Decouple_ = []
					opcode.Make(opcode.OpPop),
					opcode.Make(opcode.OpGetLocal, 0),    // _Decouple_
					opcode.Make(opcode.OpConstant, 1),    // _len
					opcode.Make(opcode.OpCall, 1, 0),     // len(_Decouple_)
					opcode.Make(opcode.OpConstant, 2),    // 2
					opcode.Make(opcode.OpLessThan, 0, 0), // len(_Decouple_) < 2
					opcode.Make(opcode.OpJumpIfNotTruthy, 16),
					opcode.Make(opcode.OpNull),
					opcode.Make(opcode.OpSetNonLocal, 0, 1), // x = null
					opcode.Make(opcode.OpPop),
					opcode.Make(opcode.OpNull),
					opcode.Make(opcode.OpSetNonLocal, 1, 1), // y = null
					opcode.Make(opcode.OpPop),
					opcode.Make(opcode.OpFalse), // decoupling failure
					opcode.Make(opcode.OpJump, 25),
					opcode.Make(opcode.OpGetLocal, 0),       // _Decouple_
					opcode.Make(opcode.OpConstant, 3),       // 1
					opcode.Make(opcode.OpIndex, 0),          // _Decouple_[1]
					opcode.Make(opcode.OpSetNonLocal, 0, 1), // x = _Decouple_[1]
					opcode.Make(opcode.OpPop),
					opcode.Make(opcode.OpGetLocal, 0),       // _Decouple_
					opcode.Make(opcode.OpConstant, 2),       // 2
					opcode.Make(opcode.OpIndex, 0),          // _Decouple_[2]
					opcode.Make(opcode.OpSetNonLocal, 1, 1), // y = _Decouple_[2]
					opcode.Make(opcode.OpPop),
					opcode.Make(opcode.OpTrue), // decoupling success
				},
				[]opcode.Instructions{ // if test / action scope
					opcode.Make(opcode.OpExecute, 4),                 // decoupling scope
					opcode.Make(opcode.OpJumpRelayIfNotTruthy, 5, 1), // jump to "else" result
					opcode.Make(opcode.OpGetLocal, 0),                // x
				},
			},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpExecute, 5), // if test / action scope
				opcode.Make(opcode.OpJump, 1),    // jump over "else" (out of if/else)
				opcode.Make(opcode.OpNull),       // else null
				opcode.Make(opcode.OpPop),
			},
		},
	}

	runCompilerTests(t, tests, false)
}

func TestCompilerForLoop(t *testing.T) {
	tests := []compilerTestCase{
		// 3 part for loop
		{
			input: "for x = 1; x < 7; x = x + 1 { x }",
			expectedConstants: []interface{}{
				1, 7,
				[]opcode.Instructions{
					// init
					opcode.Make(opcode.OpConstant, 0), // 1
					opcode.Make(opcode.OpSetLocal, 0), // x = 1
					opcode.Make(opcode.OpPop),
					opcode.Make(opcode.OpNull),
					opcode.Make(opcode.OpSetLocal, 1), // _for = null
					opcode.Make(opcode.OpPop),

					// test
					opcode.Make(opcode.OpGetLocal, 0),    // x
					opcode.Make(opcode.OpConstant, 1),    // 7
					opcode.Make(opcode.OpLessThan, 0, 0), // x < 7
					opcode.Make(opcode.OpJumpIfNotTruthy, 17),

					// body
					opcode.Make(opcode.OpGetLocal, 0), // { x }
					opcode.Make(opcode.OpPop),

					// increment
					opcode.Make(opcode.OpGetLocal, 0), // x
					opcode.Make(opcode.OpConstant, 0), // 1
					opcode.Make(opcode.OpAdd),
					opcode.Make(opcode.OpSetLocal, 0), // x = x + 1
					opcode.Make(opcode.OpPop),

					opcode.Make(opcode.OpJumpBack, 26),

					opcode.Make(opcode.OpGetLocal, 1), // _for
				},
			},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpExecute, 2),
				opcode.Make(opcode.OpPop),
			},
		},

		// for ... in
		{
			input: "for x in [1, 2] { x }",
			expectedConstants: []interface{}{
				1, 2, process.GetBuiltInByName("_len"),
				[]opcode.Instructions{
					// init
					opcode.Make(opcode.OpConstant, 0), // 1
					opcode.Make(opcode.OpConstant, 1), // 2
					opcode.Make(opcode.OpList, 2),     // [1, 2]

					opcode.Make(opcode.OpSetLocal, 0), // _LoopOver_ = [1, 2]
					opcode.Make(opcode.OpPop),

					opcode.Make(opcode.OpGetLocal, 0), // _LoopOver_
					opcode.Make(opcode.OpConstant, 2), // _len(
					opcode.Make(opcode.OpCall, 1, 0),  // _len(_LoopOver_)
					opcode.Make(opcode.OpSetLocal, 1), // _LoopLimit_ = len(_LoopOver_)
					opcode.Make(opcode.OpPop),

					opcode.Make(opcode.OpGetLocal, 0), // _LoopOver_
					opcode.Make(opcode.OpConstant, 0), // 1
					opcode.Make(opcode.OpIndex, 1),
					opcode.Make(opcode.OpNull),
					opcode.Make(opcode.OpSetLocal, 2), // x = _LoopOver_[1; null]
					opcode.Make(opcode.OpPop),

					opcode.Make(opcode.OpConstant, 0), // 1
					opcode.Make(opcode.OpSetLocal, 3), // _LoopInc_ = 1
					opcode.Make(opcode.OpPop),

					opcode.Make(opcode.OpNull),
					opcode.Make(opcode.OpSetLocal, 4), // _for = null
					opcode.Make(opcode.OpPop),

					// test
					opcode.Make(opcode.OpGetLocal, 3), // _LoopInc_
					opcode.Make(opcode.OpGetLocal, 1), // _LoopLimit_
					opcode.Make(opcode.OpLessThanOrEqual, 0, 0),
					opcode.Make(opcode.OpJumpIfNotTruthy, 28),

					// body
					opcode.Make(opcode.OpGetLocal, 2), // { x }
					opcode.Make(opcode.OpPop),

					// increment
					opcode.Make(opcode.OpGetLocal, 3), // _LoopInc_
					opcode.Make(opcode.OpConstant, 0), // 1
					opcode.Make(opcode.OpAdd),
					opcode.Make(opcode.OpSetLocal, 3), // _LoopInc_ += 1
					opcode.Make(opcode.OpPop),

					opcode.Make(opcode.OpGetLocal, 0), // _LoopOver_
					opcode.Make(opcode.OpGetLocal, 3), // _LoopInc_
					opcode.Make(opcode.OpIndex, 1),
					opcode.Make(opcode.OpNull),
					opcode.Make(opcode.OpSetLocal, 2), // x = _LoopOver_[_LoopInc_; null]
					opcode.Make(opcode.OpPop),

					opcode.Make(opcode.OpJumpBack, 36),

					opcode.Make(opcode.OpGetLocal, 4), // _for
				},
			},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpExecute, 3),
				opcode.Make(opcode.OpPop),
			},
		},
	}

	runCompilerTests(t, tests, false)
}

func TestCompilerNextBreak(t *testing.T) {
	tests := []compilerTestCase{
		// break
		{
			input: "for x = 1; x < 7; x = x + 1 { if x > 3 { break } }",
			expectedConstants: []interface{}{
				1, 7, 3,
				[]opcode.Instructions{
					// init
					opcode.Make(opcode.OpConstant, 0), // 1
					opcode.Make(opcode.OpSetLocal, 0), // x = 1
					opcode.Make(opcode.OpPop),

					opcode.Make(opcode.OpNull),
					opcode.Make(opcode.OpSetLocal, 1), // _for = null
					opcode.Make(opcode.OpPop),

					// test
					opcode.Make(opcode.OpGetLocal, 0),    // x
					opcode.Make(opcode.OpConstant, 1),    // 7
					opcode.Make(opcode.OpLessThan, 0, 0), // x < 7
					opcode.Make(opcode.OpJumpIfNotTruthy, 37),

					// body
					opcode.Make(opcode.OpGetLocal, 0),       // x
					opcode.Make(opcode.OpConstant, 2),       // 3
					opcode.Make(opcode.OpGreaterThan, 0, 0), // x > 3
					opcode.Make(opcode.OpJumpIfNotTruthy, 7),
					opcode.Make(opcode.OpGetLocal, 1), // _for
					opcode.Make(opcode.OpJump, 16),    // break; jump past increment and JumpBack
					// jump to end would go here if the compiler doesn't know better (as it's extraneous)
					opcode.Make(opcode.OpNull),
					opcode.Make(opcode.OpPop),

					// increment
					opcode.Make(opcode.OpGetLocal, 0), // x
					opcode.Make(opcode.OpConstant, 0), // 1
					opcode.Make(opcode.OpAdd),
					opcode.Make(opcode.OpSetLocal, 0), // x = x + 1
					opcode.Make(opcode.OpPop),

					opcode.Make(opcode.OpJumpBack, 46),

					opcode.Make(opcode.OpGetLocal, 1), // _for

				},
			},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpExecute, 3),
				opcode.Make(opcode.OpPop),
			},
		},

		// next
		{
			input: "for x = 1; x < 7; x = x + 1 { next; 123 }",
			expectedConstants: []interface{}{
				1, 7, 123,
				[]opcode.Instructions{
					// init
					opcode.Make(opcode.OpConstant, 0), // 1
					opcode.Make(opcode.OpSetLocal, 0), // x = 1
					opcode.Make(opcode.OpPop),

					opcode.Make(opcode.OpNull),
					opcode.Make(opcode.OpSetLocal, 1), // _for = null
					opcode.Make(opcode.OpPop),

					// test
					opcode.Make(opcode.OpGetLocal, 0),         // x
					opcode.Make(opcode.OpConstant, 1),         // 7
					opcode.Make(opcode.OpLessThan, 0, 0),      // x < 7
					opcode.Make(opcode.OpJumpIfNotTruthy, 24), // jump out

					// body
					opcode.Make(opcode.OpNull),        // FIXME: unnecessary, but works for now (in case of OpJumpRelay expecting a value on the stack)
					opcode.Make(opcode.OpJump, 4),     // next; jump to increment
					opcode.Make(opcode.OpConstant, 2), // 123
					opcode.Make(opcode.OpPop),

					// increment
					opcode.Make(opcode.OpGetLocal, 0), // x
					opcode.Make(opcode.OpConstant, 0), // 1
					opcode.Make(opcode.OpAdd),
					opcode.Make(opcode.OpSetLocal, 0), // x = x + 1
					opcode.Make(opcode.OpPop),

					opcode.Make(opcode.OpJumpBack, 33), // jump back to test

					opcode.Make(opcode.OpGetLocal, 1), // _for
				},
			},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpExecute, 3),
				opcode.Make(opcode.OpPop),
			},
		},

		// next embedded in a deeper scope (in another frame)
		{
			input: `for x = 1; x < 3; x += 1 {
						  {  var y = 7
						     if x > 2 {
							    next
					 		 }
						  }
					  }`,
			expectedConstants: []interface{}{
				1, 3, 7, 2,
				[]opcode.Instructions{ // { var y = ... }
					opcode.Make(opcode.OpConstant, 2), // 7
					opcode.Make(opcode.OpSetLocal, 0), // y = 7
					opcode.Make(opcode.OpPop),
					opcode.Make(opcode.OpGetNonLocal, 0, 1), // x
					opcode.Make(opcode.OpConstant, 3),       // 2
					opcode.Make(opcode.OpGreaterThan, 0, 0), // x > 2
					opcode.Make(opcode.OpJumpIfNotTruthy, 6),
					opcode.Make(opcode.OpNull), // FIXME: unnecessary, but works for now (in case of OpJumpRelay expecting a value on the stack)
					opcode.Make(opcode.OpJumpRelay, 1, 1),
					// jump to end would go here if the compiler doesn't know better (as it's extraneous)
					opcode.Make(opcode.OpNull),
				},
				[]opcode.Instructions{
					// init
					opcode.Make(opcode.OpConstant, 0), // 1
					opcode.Make(opcode.OpSetLocal, 0), // x = 1
					opcode.Make(opcode.OpPop),

					opcode.Make(opcode.OpNull),
					opcode.Make(opcode.OpSetLocal, 1), // _for = null
					opcode.Make(opcode.OpPop),

					// test
					opcode.Make(opcode.OpGetLocal, 0),
					opcode.Make(opcode.OpConstant, 1), // 3
					opcode.Make(opcode.OpLessThan, 0, 0),
					opcode.Make(opcode.OpJumpIfNotTruthy, 18),

					// body
					opcode.Make(opcode.OpExecute, 4), // { ... }
					opcode.Make(opcode.OpPop),

					// increment
					opcode.Make(opcode.OpGetLocal, 0), // x
					opcode.Make(opcode.OpConstant, 0), // 1
					opcode.Make(opcode.OpAdd),
					opcode.Make(opcode.OpSetLocal, 0),
					opcode.Make(opcode.OpPop),
					opcode.Make(opcode.OpJumpBack, 27),

					opcode.Make(opcode.OpGetLocal, 1), // _for
				},
			},
			expectedInstructions: []opcode.Instructions{
				// scope of the for loop
				opcode.Make(opcode.OpExecute, 5),
				opcode.Make(opcode.OpPop),
			},
		},

		{
			input: `
			  var sum = 0
			  for x = 1; x <= 10; x += 1 {
				  {  var y = 7
				     if x > 2 {
					    next
			 		 }
				  }
				  for i = 0; i < 10; i += 1 {
					 sum += 7
				  }
				  var a = 123
			  }
			  sum`,
			expectedConstants: []interface{}{
				0, 1, 10, 7, 2,
				[]opcode.Instructions{ // { var y = 7 ... }
					opcode.Make(opcode.OpConstant, 3), // 7
					opcode.Make(opcode.OpSetLocal, 0),
					opcode.Make(opcode.OpPop),

					opcode.Make(opcode.OpGetNonLocal, 0, 1),
					opcode.Make(opcode.OpConstant, 4),
					opcode.Make(opcode.OpGreaterThan, 0, 0),
					opcode.Make(opcode.OpJumpIfNotTruthy, 6),
					opcode.Make(opcode.OpNull), // FIXME: unnecessary, but works for now (in case of OpJumpRelay expecting a value on the stack)
					opcode.Make(opcode.OpJumpRelay, 11, 1),
					// jump to end would go here if the compiler doesn't know better (as it's extraneous)
					opcode.Make(opcode.OpNull),
				},
				[]opcode.Instructions{
					// init
					opcode.Make(opcode.OpConstant, 0),
					opcode.Make(opcode.OpSetLocal, 0), // i = 0
					opcode.Make(opcode.OpPop),

					opcode.Make(opcode.OpNull),
					opcode.Make(opcode.OpSetLocal, 1), // _for = null
					opcode.Make(opcode.OpPop),

					// test
					opcode.Make(opcode.OpGetLocal, 0), // i
					opcode.Make(opcode.OpConstant, 2),
					opcode.Make(opcode.OpLessThan, 0, 0), // i < 10
					opcode.Make(opcode.OpJumpIfNotTruthy, 25),

					// body
					opcode.Make(opcode.OpGetGlobal, 0), // sum
					opcode.Make(opcode.OpConstant, 3),  // 7
					opcode.Make(opcode.OpAdd),
					opcode.Make(opcode.OpSetGlobal, 0), // sum += 7
					opcode.Make(opcode.OpPop),

					// increment
					opcode.Make(opcode.OpGetLocal, 0), // i
					opcode.Make(opcode.OpConstant, 1),
					opcode.Make(opcode.OpAdd),
					opcode.Make(opcode.OpSetLocal, 0), // i += 1
					opcode.Make(opcode.OpPop),

					opcode.Make(opcode.OpJumpBack, 34),

					opcode.Make(opcode.OpGetLocal, 1), // _for
				},
				123,
				[]opcode.Instructions{
					// init
					opcode.Make(opcode.OpConstant, 1),
					opcode.Make(opcode.OpSetLocal, 0), // x = 1
					opcode.Make(opcode.OpPop),

					opcode.Make(opcode.OpNull),
					opcode.Make(opcode.OpSetLocal, 1), // _for = null
					opcode.Make(opcode.OpPop),

					// test
					opcode.Make(opcode.OpGetLocal, 0), // x
					opcode.Make(opcode.OpConstant, 2),
					opcode.Make(opcode.OpLessThanOrEqual, 0, 0), // x < 10
					opcode.Make(opcode.OpJumpIfNotTruthy, 28),

					// body
					opcode.Make(opcode.OpExecute, 5), // where the OpJumpFromLevel occurs
					opcode.Make(opcode.OpPop),
					opcode.Make(opcode.OpExecute, 6),
					opcode.Make(opcode.OpPop),

					opcode.Make(opcode.OpConstant, 7),
					opcode.Make(opcode.OpSetLocal, 2), // val a = 123
					opcode.Make(opcode.OpPop),

					// increment
					opcode.Make(opcode.OpGetLocal, 0), // x
					opcode.Make(opcode.OpConstant, 1),
					opcode.Make(opcode.OpAdd),
					opcode.Make(opcode.OpSetLocal, 0), // x += 1
					opcode.Make(opcode.OpPop),

					opcode.Make(opcode.OpJumpBack, 37),

					opcode.Make(opcode.OpGetLocal, 1), // _for
				},
			},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),  // 0
				opcode.Make(opcode.OpSetGlobal, 0), // var sum = 0
				opcode.Make(opcode.OpPop),
				opcode.Make(opcode.OpExecute, 8),
				opcode.Make(opcode.OpPop),
				opcode.Make(opcode.OpGetGlobal, 0), // sum
				opcode.Make(opcode.OpPop),
			},
		},
	}

	runCompilerTests(t, tests, false)
}

func TestCompilerBuiltIns(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "len([])",
			expectedConstants: []interface{}{object.EmptyList, process.GetBuiltInByName("len")},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 1),
				opcode.Make(opcode.OpConstant, 0), // len
				opcode.Make(opcode.OpCall, 1, 0),
				opcode.Make(opcode.OpPop),
			},
		},
		{
			input: "fn() { len([]) }",
			expectedConstants: []interface{}{
				object.EmptyList,
				process.GetBuiltInByName("len"),
				[]opcode.Instructions{
					opcode.Make(opcode.OpConstant, 1),
					opcode.Make(opcode.OpConstant, 0), // len
					opcode.Make(opcode.OpCall, 1, 0),
					opcode.Make(opcode.OpReturnValue),
				},
			},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 2),
				opcode.Make(opcode.OpPop),
			},
		},
		{
			input: "fn() { len([1]) }",
			expectedConstants: []interface{}{
				process.GetBuiltInByName("len"), 1,
				[]opcode.Instructions{
					opcode.Make(opcode.OpConstant, 1), // 1
					opcode.Make(opcode.OpList, 1),
					opcode.Make(opcode.OpConstant, 0), // len
					opcode.Make(opcode.OpCall, 1, 0),
					opcode.Make(opcode.OpReturnValue),
				},
			},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 2),
				opcode.Make(opcode.OpPop),
			},
		},
		{
			input:             "len([7])",
			expectedConstants: []interface{}{process.GetBuiltInByName("len"), 7},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 1), // 7
				opcode.Make(opcode.OpList, 1),
				opcode.Make(opcode.OpConstant, 0), // len
				opcode.Make(opcode.OpCall, 1, 0),
				opcode.Make(opcode.OpPop),
			},
		},
	}

	runCompilerTests(t, tests, false)
}

func TestCompilerCallingCompiledFromBuiltIns(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: `
				val x = fn(y) { y > 123 }
				map(x, [4, 7, 212])`,
			expectedConstants: []interface{}{
				123,
				[]opcode.Instructions{
					opcode.Make(opcode.OpGetLocal, 0),
					opcode.Make(opcode.OpConstant, 0),
					opcode.Make(opcode.OpGreaterThan, 0, 0),
					opcode.Make(opcode.OpReturnValue),
				},
				process.GetBuiltInByName("map"), 4, 7, 212,
			},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 1),
				opcode.Make(opcode.OpSetGlobal, 0),
				opcode.Make(opcode.OpPop),
				opcode.Make(opcode.OpGetGlobal, 0),
				opcode.Make(opcode.OpConstant, 3),
				opcode.Make(opcode.OpConstant, 4),
				opcode.Make(opcode.OpConstant, 5),
				opcode.Make(opcode.OpList, 3),
				opcode.Make(opcode.OpConstant, 2), // map
				opcode.Make(opcode.OpCall, 2, 0),
				opcode.Make(opcode.OpPop),
			},
		},
	}

	runCompilerTests(t, tests, false)
}

func TestCompilerClosures(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: `
						fn(a) {
							fn(b) {
								a + b
							}
						}`,
			expectedConstants: []interface{}{
				[]opcode.Instructions{ // fn(b) ...
					opcode.Make(opcode.OpGetFree, 0),  // a
					opcode.Make(opcode.OpGetLocal, 0), // b
					opcode.Make(opcode.OpAdd),
					opcode.Make(opcode.OpReturnValue),
				},
				[]opcode.Instructions{ // fn(a) ...
					opcode.Make(opcode.OpGetLocal, 0), // fn(b)
					opcode.Make(opcode.OpFunction, 0, 1, 0),
					opcode.Make(opcode.OpReturnValue),
				},
			},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 1),
				opcode.Make(opcode.OpPop),
			},
		},

		{ // deeply nested closures
			input: `
				fn(a) {
					fn(b) {
						fn(c) {
							a + b + c
						}
					}
				}`,
			expectedConstants: []interface{}{
				[]opcode.Instructions{ // fn(c) ...
					opcode.Make(opcode.OpGetFree, 0), // a
					opcode.Make(opcode.OpGetFree, 1), // b
					opcode.Make(opcode.OpAdd),
					opcode.Make(opcode.OpGetLocal, 0), // c
					opcode.Make(opcode.OpAdd),
					opcode.Make(opcode.OpReturnValue),
				},
				[]opcode.Instructions{ // fn(b) ....
					opcode.Make(opcode.OpGetFree, 0),        // a
					opcode.Make(opcode.OpGetLocal, 0),       // b
					opcode.Make(opcode.OpFunction, 0, 2, 0), // fn(c) ...
					opcode.Make(opcode.OpReturnValue),
				},
				[]opcode.Instructions{ // fn(a) ...
					opcode.Make(opcode.OpGetLocal, 0),       // a
					opcode.Make(opcode.OpFunction, 1, 1, 0), // fn(b) ...
					opcode.Make(opcode.OpReturnValue),
				},
			},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 2), // fn(a) ...
				opcode.Make(opcode.OpPop),
			},
		},

		{
			input: `
			val global = 55

			fn() {
				val a = 66

				fn() {
					val b = 77

					fn() {
						val c = 88
						global + a + b + c
					}
				}
			}
			`,
			expectedConstants: []interface{}{
				55,
				66,
				77,
				88,
				[]opcode.Instructions{
					opcode.Make(opcode.OpConstant, 3), // 88
					opcode.Make(opcode.OpSetLocal, 0), // val c = 88
					opcode.Make(opcode.OpPop),
					opcode.Make(opcode.OpGetFree, 0), // global
					opcode.Make(opcode.OpGetFree, 1), // a
					opcode.Make(opcode.OpAdd),
					opcode.Make(opcode.OpGetFree, 2), // b

					opcode.Make(opcode.OpAdd),
					opcode.Make(opcode.OpGetLocal, 0), // c
					opcode.Make(opcode.OpAdd),
					opcode.Make(opcode.OpReturnValue),
				},
				[]opcode.Instructions{
					opcode.Make(opcode.OpConstant, 2), // 77
					opcode.Make(opcode.OpSetLocal, 0), // val b = 77
					opcode.Make(opcode.OpPop),
					opcode.Make(opcode.OpGetFree, 0),
					opcode.Make(opcode.OpGetFree, 1),
					opcode.Make(opcode.OpGetLocal, 0),
					opcode.Make(opcode.OpFunction, 4, 3, 0),

					opcode.Make(opcode.OpReturnValue),
				},
				[]opcode.Instructions{
					opcode.Make(opcode.OpConstant, 1), // 66
					opcode.Make(opcode.OpSetLocal, 0), // val a = 66
					opcode.Make(opcode.OpPop),
					opcode.Make(opcode.OpGetFree, 0),
					opcode.Make(opcode.OpGetLocal, 0),
					opcode.Make(opcode.OpFunction, 5, 2, 0),

					opcode.Make(opcode.OpReturnValue),
				},
			},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),  // 55
				opcode.Make(opcode.OpSetGlobal, 0), // val global = 55
				opcode.Make(opcode.OpPop),
				opcode.Make(opcode.OpGetGlobal, 0), // global
				opcode.Make(opcode.OpFunction, 6, 1, 0),

				opcode.Make(opcode.OpPop),
			},
		},

		{ // changing the value used in building a closure's "free" variable...
			input: `var x = 7
			val chk = fn() { x ^ 2 }

			# change x before testing chk()
			x = 21

			chk()
			`,
			expectedConstants: []interface{}{
				7, 2,
				[]opcode.Instructions{
					opcode.Make(opcode.OpGetFree, 0),  // x
					opcode.Make(opcode.OpConstant, 1), // 2
					opcode.Make(opcode.OpPower),       // x ^ 2
					opcode.Make(opcode.OpReturnValue),
				},
				21,
			},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),  // 7
				opcode.Make(opcode.OpSetGlobal, 0), // var x = ...
				opcode.Make(opcode.OpPop),
				opcode.Make(opcode.OpGetGlobal, 0),
				opcode.Make(opcode.OpFunction, 2, 1, 0), // fn() { x ^ 2 }
				opcode.Make(opcode.OpSetGlobal, 1),      // val chk = ...

				opcode.Make(opcode.OpPop),
				opcode.Make(opcode.OpConstant, 3),  // 21
				opcode.Make(opcode.OpSetGlobal, 0), // x = 21
				opcode.Make(opcode.OpPop),
				opcode.Make(opcode.OpGetGlobal, 1),
				opcode.Make(opcode.OpCall, 0, 0),
				opcode.Make(opcode.OpPop),
			},
		},
	}

	runCompilerTests(t, tests, false)
}

func TestCompilerRecursion(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: `
				val factorial = fn(x) { if x > 1 { x * factorial(x - 1) } else { 1 } }
				factorial(7)
				`,
			expectedConstants: []interface{}{
				1,
				[]opcode.Instructions{
					opcode.Make(opcode.OpGetLocal, 0),       // x
					opcode.Make(opcode.OpConstant, 0),       // 1
					opcode.Make(opcode.OpGreaterThan, 0, 0), // x > 1
					opcode.Make(opcode.OpJumpIfNotTruthy, 18),
					opcode.Make(opcode.OpGetLocal, 0), // x
					opcode.Make(opcode.OpGetLocal, 0), // x
					opcode.Make(opcode.OpConstant, 0), // 1
					opcode.Make(opcode.OpSubtract),    // x - 1
					opcode.Make(opcode.OpGetSelf),     // factorial
					opcode.Make(opcode.OpCall, 1, 0),  // factorial(x - 1)
					opcode.Make(opcode.OpMultiply),    // x * factorial(x - 1)
					opcode.Make(opcode.OpJump, 3),
					opcode.Make(opcode.OpConstant, 0), // 1
					opcode.Make(opcode.OpReturnValue),
				},
				7,
			},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 1),
				opcode.Make(opcode.OpSetGlobal, 0),
				opcode.Make(opcode.OpPop),
				opcode.Make(opcode.OpConstant, 2),
				opcode.Make(opcode.OpGetGlobal, 0),
				opcode.Make(opcode.OpCall, 1, 0),
				opcode.Make(opcode.OpPop),
			},
		},

		{
			input: `
					# This is just a test. We would use the built-in map() function instead.
					val mapping = fn(f, arr) {
						# test recursive closure declared inside other function
					    val iter = fn(remaining, accumulated) {
					        if len(remaining) == 0 {
					            accumulated
					        } else {
					            iter(less(remaining, of=1), more(accumulated, f(remaining[1])))
					       }
					    }
					    iter(arr, [])
					}
					val doubled = mapping(fn{*2}, [1, 2, 3, 4])
					`,

			expectedConstants: []interface{}{
				process.GetBuiltInByName("len"),
				0,
				process.GetBuiltInByName("less"),
				object.NewString("of"),
				1,
				process.GetBuiltInByName("more"),
				[]opcode.Instructions{ // iter function
					opcode.Make(opcode.OpGetLocal, 0), // remaining
					opcode.Make(opcode.OpConstant, 0), // len(...
					opcode.Make(opcode.OpCall, 1, 0),  // len(remaining)
					opcode.Make(opcode.OpConstant, 1), // 0
					opcode.Make(opcode.OpEqual, 0, 0), // len(remaining) == 0
					opcode.Make(opcode.OpJumpIfNotTruthy, 7),
					opcode.Make(opcode.OpGetLocal, 1), // accumulated
					opcode.Make(opcode.OpJump, 40),
					opcode.Make(opcode.OpGetLocal, 0), // remaining
					opcode.Make(opcode.OpConstant, 3), // of=
					opcode.Make(opcode.OpConstant, 4), // 1
					opcode.Make(opcode.OpNameValue),   // of=1
					opcode.Make(opcode.OpConstant, 2), // less(
					opcode.Make(opcode.OpCall, 1, 1),  // less(remaining, of=1)
					opcode.Make(opcode.OpGetLocal, 1), // accumulated
					opcode.Make(opcode.OpGetLocal, 0), // remaining
					opcode.Make(opcode.OpConstant, 4), // 1
					opcode.Make(opcode.OpIndex, 0),    // remaining[1]
					opcode.Make(opcode.OpGetFree, 0),  // fn
					opcode.Make(opcode.OpCall, 1, 0),  // fn(first(remaining))
					opcode.Make(opcode.OpConstant, 5), // more(
					opcode.Make(opcode.OpCall, 2, 0),  // more(accumulated, fn(first(remaining)))
					opcode.Make(opcode.OpGetSelf),     // iter
					opcode.Make(opcode.OpCall, 2, 0),  // iter(rest(remaining), more(accumulated, fn(first(remaining))))
					opcode.Make(opcode.OpReturnValue),
				},
				object.EmptyList,
				[]opcode.Instructions{ // mapping function
					opcode.Make(opcode.OpGetLocal, 0),
					opcode.Make(opcode.OpFunction, 6, 1, 0),
					opcode.Make(opcode.OpSetLocal, 2), // iter =
					opcode.Make(opcode.OpPop),
					opcode.Make(opcode.OpGetLocal, 1), // arr
					opcode.Make(opcode.OpConstant, 7), // empty list
					opcode.Make(opcode.OpGetLocal, 2), // iter
					opcode.Make(opcode.OpCall, 2, 0),  // iter(...)
					opcode.Make(opcode.OpReturnValue),
				},
				2,
				[]opcode.Instructions{ // fn{*2}
					opcode.Make(opcode.OpGetLocal, 0), // x
					opcode.Make(opcode.OpConstant, 9), // 2
					opcode.Make(opcode.OpMultiply),    // x * 2
					opcode.Make(opcode.OpReturnValue),
				},
				3, 4,
			},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 8),  // ... = fn(f, arr) { ...
				opcode.Make(opcode.OpSetGlobal, 0), // mapping = ...
				opcode.Make(opcode.OpPop),
				opcode.Make(opcode.OpConstant, 10), // fn{*2}
				opcode.Make(opcode.OpConstant, 4),  // 1
				opcode.Make(opcode.OpConstant, 9),  // 2
				opcode.Make(opcode.OpConstant, 11), // 3
				opcode.Make(opcode.OpConstant, 12), // 4
				opcode.Make(opcode.OpList, 4),      // [1, 2, 3, 4]
				opcode.Make(opcode.OpGetGlobal, 0), // mapping(...
				opcode.Make(opcode.OpCall, 2, 0),   // mapping(fn{*2}, [1, 2, 3, 4])
				opcode.Make(opcode.OpSetGlobal, 1), // doubled = ...
				opcode.Make(opcode.OpPop),
			},
		},
	}

	runCompilerTests(t, tests, false)
}

// func TestCompiledInfixOpsWithPreevaluation(t *testing.T) {
// 	tests := []compilerTestCase{
// 		{
// 			input: `1 + 1`,
// 			expectedConstants: []interface{}{
// 				2,
// 			},
// 			expectedInstructions: []opcode.Instructions{
// 				opcode.Make(opcode.OpConstant, 0),
// 				opcode.Make(opcode.OpPop),
// 			},
// 		},
// 	}

// 	runCompilerTests(t, tests, false)
// }

// func TestCompilerSomething(t *testing.T) {
// 	tests := []compilerTestCase{
// 		{
// 			input: ``,
// 			expectedConstants: []interface{}{
				
// 			},
// 			expectedInstructions: []opcode.Instructions{
				
// 			},
// 		},
// 	}

// 	runCompilerTests(t, tests, false)
// }
