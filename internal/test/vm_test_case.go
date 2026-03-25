// langur/test/vm_test_case.go

package test

import (
	"langur/trace"
	"fmt"
	"langur/ast"
	"langur/object"
	"langur/vm"
	"testing"
	"time"
)

type vmTestCase struct {
	input        string
	expected     interface{}
	expectedType object.ObjectType
}

func runVmTests(t *testing.T, tests []vmTestCase, printTestFirst, testPrintSpeed bool) {
	t.Helper()

	for i, tt := range tests {
		result := oneResult(t, i, tt.input, printTestFirst, testPrintSpeed)
		testExpectedObject(t, i, tt.input, tt.expected, tt.expectedType, result)
	}
}

func oneResult(t *testing.T, testno int, input string, printTestFirst, testPrintSpeed bool) object.Object {
	if printTestFirst {
		fmt.Printf("Test %d: %q\n", testno, input)
	}

	var where *trace.Where
	printCodeLocationTrace := true

	program := parse(t, input)

	comp, err := ast.NewCompiler(nil, false)
	if err != nil {
		t.Fatalf("Test %d: (%q) compiler error on New: %s", testno, input, err)
	}
	
	_, err = program.Compile(comp)
	if err != nil {
		t.Fatalf("Test %d: (%q)\ncompiler error: %s", testno, input, err)
	}

	machine := vm.New(comp.ByteCode(), nil)

	var start, end int64

	if testPrintSpeed {
		start = time.Now().UnixNano()
	}

	err, where = machine.Run()
	if err != nil {
		tr := ""
		if printCodeLocationTrace && where != nil {
			tr = "\n" + where.Trace(input)
		}
		
		t.Fatalf("Test %d: (%q)\nvm error: %s%s", testno, input, err, tr)
	}

	if testPrintSpeed {
		end = time.Now().UnixNano()
		fmt.Printf("Test %d: VM Test Time in Microseconds (Nanoseconds): %d (%d)\n", testno, (end-start)/1000, end-start)
	}

	return machine.LastValue()
}

func testExpectedObject(
	t *testing.T,
	testno int,
	input string,
	expected interface{},
	expectedType object.ObjectType,
	actual object.Object,
) {
	t.Helper()

	switch expectedType {
	case object.NUMBER_OBJ:
		switch expected.(type) {
		case string:
			// might be a Go string to represent a langur decimal floating point number
			num, err := object.NumberFromString(expected.(string))
			if err != nil {
				t.Errorf("Test %d: (%q) testNumberObject failed: %s", testno, input, err)
			}
			err = testNumberObject(num, actual)
			if err != nil {
				t.Errorf("Test %d: (%q) testNumberObject failed: %s", testno, input, err)
			}

		case int:
			err := testNumberObject(object.NumberFromInt(expected.(int)), actual)
			if err != nil {
				t.Errorf("Test %d: (%q) testNumberObject failed: %s", testno, input, err)
			}

		default:
			t.Fatalf("Test %d: (%q) testExpectedObject failed for object.NUMBER_OBJ: no case for %T", testno, input, expected)
		}

	case object.COMPLEX_OBJ:
		err := testComplexObject(expected.(string), actual)
		if err != nil {
			t.Errorf("Test %d: (%q) testComplexObject failed: %s", testno, input, err)
		}

	case object.BOOLEAN_OBJ:
		err := testBooleanObject(expected.(bool), actual)
		if err != nil {
			t.Errorf("Test %d: (%q) testBooleanObject failed: %s", testno, input, err)
		}

	case object.NULL_OBJ:
		err := testNullObject(actual)
		if err != nil {
			t.Errorf("Test %d: (%q) testNullObject failed: %s", testno, input, err)
		}

	case object.STRING_OBJ:
		err := testStringObject(expected.(string), actual)
		if err != nil {
			t.Errorf("Test %d: (%q) testStringObject failed: %s", testno, input, err)
		}

	case object.REGEX_OBJ:
		err := testRegexObject(expected.(string), actual)
		if err != nil {
			t.Errorf("Test %d: (%q) testRegexObject failed: %s", testno, input, err)
		}

	case object.DATETIME_OBJ:
		err := testDateTimeObject(expected.(string), actual)
		if err != nil {
			t.Errorf("Test %d: (%q) testDateTimeObject failed: %s", testno, input, err)
		}

	case object.DURATION_OBJ:
		err := testDurationObject(expected.(string), actual)
		if err != nil {
			t.Errorf("Test %d: (%q) testDurationObject failed: %s", testno, input, err)
		}

	case object.RANGE_OBJ:
		err := testRangeObject(expected.([]int64), actual)
		if err != nil {
			t.Errorf("Test %d: (%q) testRangeObject failed: %s", testno, input, err)
		}

	case object.LIST_OBJ:
		intList, ok := expected.([]int)
		if ok {
			err := testListOfIntObject(intList, actual)
			if err != nil {
				t.Errorf("Test %d: (%q) testListOfIntObject failed: %s", testno, input, err)
			}

		} else if strArr, ok := expected.([]string); ok {
			err := testListOfStringObject(strArr, actual)
			if err != nil {
				t.Errorf("Test %d: (%q) testListOfStringObject failed: %s", testno, input, err)
			}

		} else if intList, ok := expected.([][]int); ok {
			err := testListOfListOfIntegerObject(intList, actual)
			if err != nil {
				t.Errorf("Test %d: (%q) testListOfListOfIntegerObject failed: %s", testno, input, err)
			}

		} else if rngList, ok := expected.([][]int64); ok {
			err := testListOfRangeObject(rngList, actual)
			if err != nil {
				t.Errorf("Test %d: (%q) testListOfRangeObject failed: %s", testno, input, err)
			}

		} else if rngList2, ok := expected.([][][]int64); ok {
			err := testListOfListsOfRangeObject(rngList2, actual)
			if err != nil {
				t.Errorf("Test %d: (%q) testListOfListsOfRangeObject failed: %s", testno, input, err)
			}

		} else if rngList, ok := expected.([][]string); ok {
			err := testListOfListOfStringsObject(rngList, actual)
			if err != nil {
				t.Errorf("Test %d: (%q) testListOfListOfStringsObject failed: %s", testno, input, err)
			}

		} else {
			t.Errorf("Test %d: (%q) no test method found for expected %T", testno, input, expected)
		}

	case object.HASH_OBJ:
		err := testHashObject(expected.([][]object.Object), actual)
		if err != nil {
			t.Errorf("Test %d: (%q) testHashObject failed: %s", testno, input, err)
		}

	default:
		t.Errorf("Test %d: (%q) unknown for expected type, received=%v", testno, input, expectedType)
	}
}

func testObject(expected, actual object.Object) error {
	var err error
	switch e := expected.(type) {
	case *object.List:
		err = testListObject(e, actual)
		if err != nil {
			return fmt.Errorf("testListObject failed: %s", err)
		}

	case *object.Number:
		err = testNumberObject(e, actual)
		if err != nil {
			return fmt.Errorf("testNumberObject failed: %s", err)
		}

	case *object.String:
		err = testStringObject(expected.String(), actual)
		if err != nil {
			return fmt.Errorf("testStringObject failed: %s", err)
		}

	case *object.Boolean:
		err = testBooleanObject(expected.(*object.Boolean).Value, actual)
		if err != nil {
			return fmt.Errorf("testBooleanObject failed: %s", err)
		}

	default:
		return fmt.Errorf("No test for %T by testObject", expected)
	}
	return nil
}

func testComplexObject(expected string, actual object.Object) error {
	result, ok := actual.(*object.Complex)
	if !ok {
		return fmt.Errorf("object not a Complex, received=%T (%+v)", actual, actual)
	}

	if result.String() != expected {
		return fmt.Errorf("object value wrong\nexpected=%q\nreceived=%q", expected, result.String())
	}

	return nil
}

func testBooleanObject(expected bool, actual object.Object) error {
	result, ok := actual.(*object.Boolean)
	if !ok {
		return fmt.Errorf("object not a Boolean, received=%T (%+v)", actual, actual)
	}

	if result.Value != expected {
		return fmt.Errorf("object value wrong, expected=%t, received=%t", expected, result.Value)
	}

	return nil
}

func testNullObject(actual object.Object) error {
	if _, ok := actual.(*object.Null); !ok {
		return fmt.Errorf("object not a Null, received=%T (%+v)", actual, actual)
	}
	return nil
}

func testStringObject(expected string, actual object.Object) error {
	result, ok := actual.(*object.String)
	if !ok {
		return fmt.Errorf("object not a string, received=%T (%+v)", actual, actual)
	}

	if result.String() != expected {
		return fmt.Errorf("object value wrong\nexpected=%q\nreceived=%q", expected, result.String())
	}

	return nil
}

func testRegexObject(expected string, actual object.Object) error {
	// accounting for the extra text put into regex text...
	expected = "(?-smiUx:" + expected + ")"
	
	result, ok := actual.(*object.Regex)
	if !ok {
		return fmt.Errorf("object not a regex, received=%T (%+v)", actual, actual)
	}

	if result.String() != expected {
		return fmt.Errorf("object value wrong\nexpected=%q\nreceived=%q", expected, result.String())
	}

	return nil
}

func testDateTimeObject(expected string, actual object.Object) error {
	result, ok := actual.(*object.DateTime)
	if !ok {
		return fmt.Errorf("object not a date-time, received=%T (%+v)", actual, actual)
	}

	rs := result.String()
	if rs != expected {
		return fmt.Errorf("object value wrong\nexpected=%q\nreceived=%q", expected, rs)
	}

	return nil
}

func testDurationObject(expected string, actual object.Object) error {
	result, ok := actual.(*object.Duration)
	if !ok {
		return fmt.Errorf("object not a duration, received=%T (%+v)", actual, actual)
	}

	rs := result.String()
	if rs != expected {
		return fmt.Errorf("object value wrong\nexpected=%q\nreceived=%q", expected, rs)
	}

	return nil
}

func testRangeObject(expected []int64, actual object.Object) error {
	return testRangeOfIntObject(expected[0], expected[1], actual)
}

func testRangeOfIntObject(expectLeft, expectRight int64, actual object.Object) error {
	r, ok := actual.(*object.Range)
	if !ok {
		return fmt.Errorf("object not a range, received=%T (%+v)", actual, actual)
	}

	err := testNumberObject(object.NumberFromInt64(expectLeft), r.Start)
	if err != nil {
		return fmt.Errorf("Start of range test failed: %s", err)
	}

	err = testNumberObject(object.NumberFromInt64(expectRight), r.End)
	if err != nil {
		return fmt.Errorf("End of range test failed: %s", err)
	}

	return nil
}

func testListObject(expected *object.List, actual object.Object) error {
	list, ok := actual.(*object.List)
	if !ok {
		return fmt.Errorf("object not a list, received=%T (%+v)", actual, actual)
	}
	if len(list.Elements) != len(expected.Elements) {
		return fmt.Errorf("object list length received=%d, expected=%d", len(list.Elements), len(expected.Elements))
	}
	for i, e := range expected.Elements {
		err := testObject(e, list.Elements[i])
		if err != nil {
			return fmt.Errorf("testObject failed: %s", err)
		}
	}
	return nil
}

func testListOfIntObject(expected []int, actual object.Object) error {
	list, ok := actual.(*object.List)
	if !ok {
		return fmt.Errorf("object not a list, received=%T (%+v)", actual, actual)
	}
	if len(list.Elements) != len(expected) {
		return fmt.Errorf("object list length received=%d, expected=%d", len(list.Elements), len(expected))
	}
	for i, e := range expected {
		err := testNumberObject(object.NumberFromInt(e), list.Elements[i])
		if err != nil {
			return fmt.Errorf("testNumberObject failed: %s", err)
		}
	}
	return nil
}

func testListOfRangeObject(expected [][]int64, actual object.Object) error {
	list, ok := actual.(*object.List)
	if !ok {
		return fmt.Errorf("object not a list, received=%T (%+v)", actual, actual)
	}
	if len(list.Elements) != len(expected) {
		return fmt.Errorf("object list length received=%d, expected=%d", len(list.Elements), len(expected))
	}

	for i, e := range expected {
		err := testRangeObject(e, list.Elements[i])
		if err != nil {
			return fmt.Errorf("testRangeObject failed: %s", err)
		}
	}

	return nil
}

func testListOfListsOfRangeObject(expected [][][]int64, actual object.Object) error {
	list, ok := actual.(*object.List)
	if !ok {
		return fmt.Errorf("object not a list, received=%T (%+v)", actual, actual)
	}
	if len(list.Elements) != len(expected) {
		return fmt.Errorf("object list length received=%d, expected=%d", len(list.Elements), len(expected))
	}

	for i, e := range expected {
		err := testListOfRangeObject(e, list.Elements[i])
		if err != nil {
			return fmt.Errorf("testListOfRangeObject failed: %s", err)
		}
	}

	return nil
}

func testListOfListOfIntegerObject(expected [][]int, actual object.Object) error {
	list, ok := actual.(*object.List)
	if !ok {
		return fmt.Errorf("object not a list, received=%T (%+v)", actual, actual)
	}
	if len(list.Elements) != len(expected) {
		return fmt.Errorf("object list length received=%d, expected=%d", len(list.Elements), len(expected))
	}

	for i, e := range expected {
		err := testListOfIntObject(e, list.Elements[i])
		if err != nil {
			return fmt.Errorf("testListOfIntObject failed: %s", err)
		}
	}

	return nil
}

func testListOfListOfStringsObject(expected [][]string, actual object.Object) error {
	list, ok := actual.(*object.List)
	if !ok {
		return fmt.Errorf("object not a list, received=%T (%+v)", actual, actual)
	}
	if len(list.Elements) != len(expected) {
		return fmt.Errorf("object list length received=%d, expected=%d", len(list.Elements), len(expected))
	}

	for i, e := range expected {
		err := testListOfStringObject(e, list.Elements[i])
		if err != nil {
			return fmt.Errorf("testListOfStringObject failed: %s", err)
		}
	}

	return nil
}

func testListOfStringObject(expected []string, actual object.Object) error {
	list, ok := actual.(*object.List)
	if !ok {
		return fmt.Errorf("object not a list, received=%T (%+v)", actual, actual)
	}
	if len(list.Elements) != len(expected) {
		return fmt.Errorf("object list length received=%d, expected=%d", len(list.Elements), len(expected))
	}

	for i, e := range expected {
		err := testStringObject(e, list.Elements[i])
		if err != nil {
			return fmt.Errorf("testStringObject failed: %s", err)
		}
	}

	return nil
}

func testListOfHashesObject(expected [][][]object.Object, actual object.Object) error {
	list, ok := actual.(*object.List)
	if !ok {
		return fmt.Errorf("object not a list, received=%T (%+v)", actual, actual)
	}
	if len(list.Elements) != len(expected) {
		return fmt.Errorf("object list length received=%d, expected=%d", len(list.Elements), len(expected))
	}

	for i, e := range expected {
		err := testHashObject(e, list.Elements[i])
		if err != nil {
			return fmt.Errorf("testHashObject failed: %s", err)
		}
	}

	return nil
}

func testHashObject(expected [][]object.Object, actual object.Object) error {
	hash, ok := actual.(*object.Hash)
	if !ok {
		return fmt.Errorf("object not a hash, received=%T (%+v)", actual, actual)
	}
	if len(hash.Pairs) != len(expected) {
		return fmt.Errorf("object hash length received=%d, expected=%d", len(hash.Pairs), len(expected))
	}

	for _, kv := range expected {
		expectKey, expectValue := kv[0], kv[1]

		value, err := hash.GetValue(expectKey)
		if err != nil {
			if expectKey.Type() == object.NUMBER_OBJ {
				return fmt.Errorf("No pair for key %s in Pairs", expectKey.String())
			}
			return fmt.Errorf("No pair for key %q in Pairs", expectKey.String())
		}

		err = testObject(expectValue, value)
		if err != nil {
			if expectKey.Type() == object.NUMBER_OBJ {
				return fmt.Errorf("Error for key %s in Pairs: %s", expectKey.String(), err)
			}
			return fmt.Errorf("Error for key %q in Pairs: %s", expectKey.String(), err)
		}
	}

	return nil
}
