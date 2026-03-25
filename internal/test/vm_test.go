// langur/test/vm_test.go

package test

import (
	"fmt"
	"langur/ast"
	"langur/common"
	"langur/object"
	"langur/system"
	"langur/vm"
	"runtime"
	"strings"
	"testing"
)

func TestEqualAndSame(t *testing.T) {
	tests := []vmTestCase{
		{"1 == 1.0", true, object.BOOLEAN_OBJ},
		{"[1] == [1.0]", true, object.BOOLEAN_OBJ},
		{"[1, 2, 3] == [1.0, 2.0, 3.0000]", true, object.BOOLEAN_OBJ},
		{"[1, 2, 3] == [3.0000, 1.0, 2.0]", false, object.BOOLEAN_OBJ},

		{"1.0 == 1.0", true, object.BOOLEAN_OBJ},
		{"[1.0] == [1.0]", true, object.BOOLEAN_OBJ},
	}

	runVmTests(t, tests, false, false)
}

func TestNumberLiterals(t *testing.T) {
	tests := []vmTestCase{
		{"2", 2, object.NUMBER_OBJ},
		{"16xff", 255, object.NUMBER_OBJ},
		{"-2x0011_0000_0011_1011", -12347, object.NUMBER_OBJ},

		{"2i", "0+2i", object.COMPLEX_OBJ},
		{"-2i", "0-2i", object.COMPLEX_OBJ},
		{"2 + 2i", "2+2i", object.COMPLEX_OBJ},
		{"2 - 2i", "2-2i", object.COMPLEX_OBJ},

		{"simplify(2i)", "0+2i", object.COMPLEX_OBJ},
		{"simplify(0i)", "0+0i", object.COMPLEX_OBJ},
		{"simplify(3+11i)", "3+11i", object.COMPLEX_OBJ},
		{"simplify(3+11.100i)", "3+11.1i", object.COMPLEX_OBJ},
		{"simplify(3+0i)", "3+0i", object.COMPLEX_OBJ},
		{"simplify(3.0+11.0i)", "3+11i", object.COMPLEX_OBJ},
		{"simplify(3.1560+11.0120i)", "3.156+11.012i", object.COMPLEX_OBJ},
		{"simplify(3.1560+0i)", "3.156+0i", object.COMPLEX_OBJ},

		// with continuation
		{`6._
			_2831853071_7958647692_
			_5286766559_0057683943`,
			"6.2831853071795864769252867665590057683943",
			object.NUMBER_OBJ},

		{`6._
			_2831853071_7958647692_  # with comment
			_5286766559_0057683943`,
			"6.2831853071795864769252867665590057683943",
			object.NUMBER_OBJ},

		{`2x0011_0000_
		  _0011_1011`, 12347, object.NUMBER_OBJ},
	}

	runVmTests(t, tests, false, false)
}

func TestMath(t *testing.T) {
	tests := []vmTestCase{
		{"1", "1", object.NUMBER_OBJ},
		{"2", "2", object.NUMBER_OBJ},

		{"-(1)", "-1", object.NUMBER_OBJ},
		{"-(-1)", "1", object.NUMBER_OBJ},

		{"abs(1)", "1", object.NUMBER_OBJ},
		{"abs(-1)", "1", object.NUMBER_OBJ},

		{"1 + 2", "3", object.NUMBER_OBJ},
		{"1 + 2.2", "3.2", object.NUMBER_OBJ},
		{"3.2 + 3.8", "7.0", object.NUMBER_OBJ},

		{"1 - 2", "-1", object.NUMBER_OBJ},

		// multiply
		{"2 * 2", "4", object.NUMBER_OBJ},
		{"2 * -1", "-2", object.NUMBER_OBJ},
		{"-2 * -1", "2", object.NUMBER_OBJ},

		// power
		{`2 ^ 0`, "1", object.NUMBER_OBJ},
		{`2 ^ 1`, "2", object.NUMBER_OBJ},
		{`2 ^ 2`, "4", object.NUMBER_OBJ},
		{`2 ^ 3`, "8", object.NUMBER_OBJ},

		{`2 ^ -3`, "0.125", object.NUMBER_OBJ},
		{`2 ^ -1`, "0.5", object.NUMBER_OBJ},

		// FIXME: unsatisfactory results from fractional exponents

		// {`2 ^ 3.5`, "11.313708498984760390413509793677585", object.NUMBER_OBJ},
		// {`2 ^ -3.5`, "0.088388347648318440550105545263106", object.NUMBER_OBJ},

		{`-2 ^ 1`, "-2", object.NUMBER_OBJ},
		// {`-2 ^ -3.5`, "-0.088388347648318440550105545263106", object.NUMBER_OBJ},

		// root
		{`16 ^/ 2`, "4", object.NUMBER_OBJ},
		{`144 ^/ 2`, "12", object.NUMBER_OBJ},
		{`8 ^/ 3`, "2", object.NUMBER_OBJ},

		// divide
		{"5 / 2", "2.5", object.NUMBER_OBJ},

		{`5 \ 2`, "2", object.NUMBER_OBJ},
		{`2 \ 3`, "0", object.NUMBER_OBJ},

		// remainder
		{"5 rem 2", "1", object.NUMBER_OBJ},
		{"5.2 rem 2", "1.2", object.NUMBER_OBJ},

		{"-5 rem 3", "-2", object.NUMBER_OBJ},
		{"-7 rem 3", "-1", object.NUMBER_OBJ},
		{"-21 rem 4", "-1", object.NUMBER_OBJ},

		// modulus
		{"-5 mod 3", "1", object.NUMBER_OBJ},
		{"-7 mod 3", "2", object.NUMBER_OBJ},

		{"5 mod 2", "1", object.NUMBER_OBJ},
		{"5.2 mod 2", "1.2", object.NUMBER_OBJ},
		{"-21 mod 4", "3", object.NUMBER_OBJ},

		// integer division (truncated)
		{`1 \ 2`, "0", object.NUMBER_OBJ},
		{`-1 \ 2`, "0", object.NUMBER_OBJ},
		{`5 \ 2`, "2", object.NUMBER_OBJ},
		{`10 \ 2`, "5", object.NUMBER_OBJ},
		{`-5 \ 2`, "-2", object.NUMBER_OBJ},
		{`-10 \ 2`, "-5", object.NUMBER_OBJ},
		{`-40 \ 3`, "-13", object.NUMBER_OBJ},

		// floor division
		{`1 // 2`, "0", object.NUMBER_OBJ},
		{`-1 // 2`, "-1", object.NUMBER_OBJ},
		{`1 // -2`, "-1", object.NUMBER_OBJ},
		{`5 // 2`, "2", object.NUMBER_OBJ},
		{`10 // 2`, "5", object.NUMBER_OBJ},
		{`-5 // 2`, "-3", object.NUMBER_OBJ},
		{`-10 // 2`, "-5", object.NUMBER_OBJ},
		{`10 // -3`, "-4", object.NUMBER_OBJ},
		{`-10 // -3`, "3", object.NUMBER_OBJ},
		{`-40 // 3`, "-14", object.NUMBER_OBJ},
		{`40 // -3`, "-14", object.NUMBER_OBJ},
		{`-41 // 3`, "-14", object.NUMBER_OBJ},

		{"50 / 2 * 2 + 7", "57", object.NUMBER_OBJ},

		{"-7", "-7", object.NUMBER_OBJ},
		{"10 + -7", "3", object.NUMBER_OBJ},

		{"-7.0", "-7.0", object.NUMBER_OBJ},
		{"-7.7", "-7.7", object.NUMBER_OBJ},
		{"10.3215 + -7", "3.3215", object.NUMBER_OBJ},

		{"1.23e-2 + 1.23", "1.2423", object.NUMBER_OBJ},
		{"1.23e+2 + 1.23", "124.23", object.NUMBER_OBJ},
	}

	runVmTests(t, tests, false, false)
}

func TestMoreMath(t *testing.T) {
	tests := []vmTestCase{
		// exponents right associative
		{"2 ^ 3 ^ 4", "2417851639229258349412352", object.NUMBER_OBJ},
		{"2 ^ (1.5 * 2) ^ 4", "2417851639229258349412352", object.NUMBER_OBJ},

		// exponents relative to the negative sign
		{"-5^2", "-25", object.NUMBER_OBJ},   // changed result as of 0.12
		{"-(5)^2", "-25", object.NUMBER_OBJ}, // changed result as of 0.12
		{"(-5)^2", "25", object.NUMBER_OBJ},
		{"-(5^2)", "-25", object.NUMBER_OBJ},
		{"-5^3", "-125", object.NUMBER_OBJ},
		{"-(5)^3", "-125", object.NUMBER_OBJ},
		{"(-5)^3", "-125", object.NUMBER_OBJ},
		{"-(5^3)", "-125", object.NUMBER_OBJ},
	}

	runVmTests(t, tests, false, false)
}

func TestMathBeyondIntOptimization(t *testing.T) {
	tests := []vmTestCase{
		{"9223372036854775807 + 1", "9223372036854775808", object.NUMBER_OBJ},
		{"-9223372036854775808 - 1", "-9223372036854775809", object.NUMBER_OBJ},
		{"9223372036854775807 * 2", "18446744073709551614", object.NUMBER_OBJ},
	}

	runVmTests(t, tests, false, false)
}

func TestComplexMath(t *testing.T) {
	tests := []vmTestCase{
		{"-(1+1i)", "-1-1i", object.COMPLEX_OBJ},
		{"-(-1-1i)", "1+1i", object.COMPLEX_OBJ},

		{"abs(1+1i)", "1.414213562373095048801688724209699", object.NUMBER_OBJ},
		{"abs(1.5+3i)", "3.354101966249684544613760503096915", object.NUMBER_OBJ},
		{"abs(1+2i)", "2.236067977499789696409173668731277", object.NUMBER_OBJ},

		{"(2+2i) + (5+13.2i)", "7+15.2i", object.COMPLEX_OBJ},
		{"(2+2i) + (5-13.2i)", "7-11.2i", object.COMPLEX_OBJ},
		{"(2+2i) + 5", "7+2i", object.COMPLEX_OBJ},
		{"(2+2i) + 5i", "2+7i", object.COMPLEX_OBJ},
		{"5 + (2+2i)", "7+2i", object.COMPLEX_OBJ},
		{"5i + (2+2i)", "2+7i", object.COMPLEX_OBJ},
		
		{"(2+2i) - (5+13.2i)", "-3-11.2i", object.COMPLEX_OBJ},
		{"(2+2i) - (5-13.2i)", "-3+15.2i", object.COMPLEX_OBJ},
		{"(2+2i) - 5", "-3+2i", object.COMPLEX_OBJ},
		{"(2+2i) - 5i", "2-3i", object.COMPLEX_OBJ},
		{"5 - (2+2i)", "3-2i", object.COMPLEX_OBJ},
		{"5i - (2+2i)", "-2+3i", object.COMPLEX_OBJ},

		{"(1+1i) * (3.141592653589793+1.2i)", "1.941592653589793+4.341592653589793i", object.COMPLEX_OBJ},
		{"(3+2i) * 2", "6+4i", object.COMPLEX_OBJ},
		{"2 * (3+2i)", "6+4i", object.COMPLEX_OBJ},
		{"2 * (3-2i)", "6-4i", object.COMPLEX_OBJ},
		{"1 * (3-2i)", "3-2i", object.COMPLEX_OBJ},
		{"0 * (3-2i)", "0+0i", object.COMPLEX_OBJ},

		{"(1.5+3i) / (1.5+1.5i)", "1.50+0.50i", object.COMPLEX_OBJ},
		{"(5+3i) / (4-3i)", "0.44+1.08i", object.COMPLEX_OBJ},
		{"1 / (1+1i)", "0.5-0.5i", object.COMPLEX_OBJ},
		{"1 / (1+2i)", "0.2-0.4i", object.COMPLEX_OBJ},
		{"1 / (4-3i)", "0.16+0.12i", object.COMPLEX_OBJ},
		{"2 / (4-3i)", "0.32+0.24i", object.COMPLEX_OBJ},
		{"3 / (4-3i)", "0.48+0.36i", object.COMPLEX_OBJ},
		{"(4-3i) / 1", "4-3i", object.COMPLEX_OBJ},
		{"(4-3i) / 2", "2-1.5i", object.COMPLEX_OBJ},
		{"(4-3i) / 3", "1.333333333333333333333333333333333-1i", object.COMPLEX_OBJ},
		{"(4-3i) / (2+0i)", "2-1.5i", object.COMPLEX_OBJ},

		{"(4-3i) ^ 0", "1+0i", object.COMPLEX_OBJ},

		{"(4-3i) ^ 1", "4-3i", object.COMPLEX_OBJ},
		{"1 / (4-3i)", "0.16+0.12i", object.COMPLEX_OBJ},
		{"(4-3i) ^ -1", "0.16+0.12i", object.COMPLEX_OBJ},

		{"(4-3i) * (4-3i)", "7-24i", object.COMPLEX_OBJ},
		{"(4-3i) ^ 2", "7-24i", object.COMPLEX_OBJ},
		{"1 / (7-24i)", "0.0112+0.0384i", object.COMPLEX_OBJ},
		{"(4-3i) ^ -2", "0.0112+0.0384i", object.COMPLEX_OBJ},

		{"(4-3i) * (4-3i) * (4-3i)", "-44-117i", object.COMPLEX_OBJ},
		{"(4-3i) ^ 3", "-44-117i", object.COMPLEX_OBJ},
	}

	runVmTests(t, tests, false, false)
}

func TestDivisibleByOperator(t *testing.T) {
	tests := []vmTestCase{
		{"5 div 2", false, object.BOOLEAN_OBJ},
		{"5 div 2.5", true, object.BOOLEAN_OBJ},
		{"5 div 2.4", false, object.BOOLEAN_OBJ},
		{"15 div 3", true, object.BOOLEAN_OBJ},
		{"7 div 1", true, object.BOOLEAN_OBJ},
		{"not 7 div 1", false, object.BOOLEAN_OBJ},

		{"5 ndiv 2", true, object.BOOLEAN_OBJ},
		{"5 ndiv 2.5", false, object.BOOLEAN_OBJ},
		{"5 ndiv 2.4", true, object.BOOLEAN_OBJ},
		{"15 ndiv 3", false, object.BOOLEAN_OBJ},
		{"7 ndiv 1", false, object.BOOLEAN_OBJ},
		{"not 7 ndiv 1", true, object.BOOLEAN_OBJ},
	}

	runVmTests(t, tests, false, false)
}

func TestVmModes(t *testing.T) {
	tests := []vmTestCase{
		{`mode divMaxScale = 33; 2 / 3`, "0.666666666666666666666666666666667", object.NUMBER_OBJ},
		{`mode divMaxScale = 7; 2 / 3`, "0.6666667", object.NUMBER_OBJ},
		{`mode divMaxScale = 7
		  val x = 2 / 3
		  mode divMaxScale = 2
		  val y = 2 / 3
		  string([x, y])`, "[0.6666667, 0.67]", object.STRING_OBJ},

		{`mode consoleText = false`, false, object.BOOLEAN_OBJ},
		{`mode consoleText = true`, true, object.BOOLEAN_OBJ},
	}

	runVmTests(t, tests, false, false)
}

func TestCombinationOperators(t *testing.T) {
	tests := []vmTestCase{
		{"var x = 12; x += 3; x", "15", object.NUMBER_OBJ},
		{"var x = 12; x -= 3; x", "9", object.NUMBER_OBJ},
		{"var x = 12; x *= 3; x", "36", object.NUMBER_OBJ},
		{"var x = 12; x /= 3; x", "4", object.NUMBER_OBJ},
		{`var x = 12; x \= 3; x`, "4", object.NUMBER_OBJ},
		{"var x = 12; x rem= 3; x", "0", object.NUMBER_OBJ},
		{"var x = 12; x mod= 3; x", "0", object.NUMBER_OBJ},

		{`var s = "Z"; s ~= "oetic"; s`, "Zoetic", object.STRING_OBJ},

		{"var x = true; x and= false; x", false, object.BOOLEAN_OBJ},
		{"var x = true; x and= true; x", true, object.BOOLEAN_OBJ},
		{"var x = true; x and= null; x", false, object.BOOLEAN_OBJ},
		{"var x = true; x and?= null; x", nil, object.NULL_OBJ},

		{"var x = false; x or= false; x", false, object.BOOLEAN_OBJ},
		{"var x = false; x or= true; x", true, object.BOOLEAN_OBJ},
		{"var x = false; x or= null; x", false, object.BOOLEAN_OBJ},
		{"var x = false; x or?= null; x", nil, object.NULL_OBJ},

		{"var x = true; x nand= false; x", true, object.BOOLEAN_OBJ},
		{"var x = true; x nand= true; x", false, object.BOOLEAN_OBJ},
		{"var x = true; x nand= null; x", true, object.BOOLEAN_OBJ},
		{"var x = true; x nand?= null; x", nil, object.NULL_OBJ},

		{"var x = false; x nor= false; x", true, object.BOOLEAN_OBJ},
		{"var x = false; x nor= true; x", false, object.BOOLEAN_OBJ},
		{"var x = false; x nor= null; x", true, object.BOOLEAN_OBJ},
		{"var x = false; x nor?= null; x", nil, object.NULL_OBJ},

		{"var x = false; x xor= false; x", false, object.BOOLEAN_OBJ},
		{"var x = false; x xor= true; x", true, object.BOOLEAN_OBJ},
		{"var x = false; x xor= null; x", false, object.BOOLEAN_OBJ},
		{"var x = false; x xor?= null; x", nil, object.NULL_OBJ},
		{"var x = false; x xor?= true; x", true, object.BOOLEAN_OBJ},

		{"var x = false; x nxor= false; x", true, object.BOOLEAN_OBJ},
		{"var x = false; x nxor= true; x", false, object.BOOLEAN_OBJ},
		{"var x = false; x nxor= null; x", true, object.BOOLEAN_OBJ},
		{"var x = false; x nxor?= null; x", nil, object.NULL_OBJ},
		{"var x = false; x nxor?= true; x", false, object.BOOLEAN_OBJ},

		{"var x = 3; x ..= 7", []int64{3, 7}, object.RANGE_OBJ},

		// used with indexing
		{`var l = [15, 7, 9]; l[2] += 21; l[1] *= 3; l`, []int{45, 28, 9}, object.LIST_OBJ},
		{`var l = [true, false, null]; l[2] and?= null; l[2]`, nil, object.NULL_OBJ},
		{`var l = [true, false, null]; l[2] and= null; l[2]`, false, object.BOOLEAN_OBJ},
		{`var l = [true, false, null]; l[1] or= null; l[1]`, true, object.BOOLEAN_OBJ},

		/*
			With a combination operator, it is as though the right side is wrapped in parentheses.

			sum *= 3 + 4
			... should be the same as ...
			sum = sum * (3 + 4)
			... not ...
			sum = sum * 3 + 4

			We check for that here.
			6 * 7 == 42
		*/
		{`var sum = 6; sum *= 3 + 4; sum`, "42", object.NUMBER_OBJ},
	}

	runVmTests(t, tests, false, false)
}

func TestBooleanExpressionsAndComparisons(t *testing.T) {
	tests := []vmTestCase{
		{"true", true, object.BOOLEAN_OBJ},
		{"false", false, object.BOOLEAN_OBJ},

		{"true == true", true, object.BOOLEAN_OBJ},
		{"true == false", false, object.BOOLEAN_OBJ},
		{"true != true", false, object.BOOLEAN_OBJ},
		{"true != false", true, object.BOOLEAN_OBJ},

		{"1 > 2", false, object.BOOLEAN_OBJ},
		{"1 >= 2", false, object.BOOLEAN_OBJ},
		{"1 > 1", false, object.BOOLEAN_OBJ},
		{"1 >= 1", true, object.BOOLEAN_OBJ},

		{"1 < 2", true, object.BOOLEAN_OBJ},
		{"1 <= 2", true, object.BOOLEAN_OBJ},
		{"1 < 1", false, object.BOOLEAN_OBJ},
		{"1 <= 1", true, object.BOOLEAN_OBJ},

		{"1 == 1", true, object.BOOLEAN_OBJ},
		{"1 != 1", false, object.BOOLEAN_OBJ},
		{"1 == 2", false, object.BOOLEAN_OBJ},
		{"1 != 2", true, object.BOOLEAN_OBJ},

		{"1 == null", false, object.BOOLEAN_OBJ},
		{"1 ==? null", nil, object.NULL_OBJ},
		{"null == null", true, object.BOOLEAN_OBJ},
		{"null ==? null", nil, object.NULL_OBJ},

		{`"HELLO" == "hello"`, false, object.BOOLEAN_OBJ},
		{`"hello" == "hello"`, true, object.BOOLEAN_OBJ},

		{"{1: null} == {1: null}", true, object.BOOLEAN_OBJ},
		{"{1: null} ==? {1: null}", true, object.BOOLEAN_OBJ},

		{"[null] == [null]", true, object.BOOLEAN_OBJ},
		{"[null] ==? [null]", true, object.BOOLEAN_OBJ},

		{"(1 < 2) == true", true, object.BOOLEAN_OBJ},
		{"(1 <= 2) == true", true, object.BOOLEAN_OBJ},
		{"(1 < 1) == true", false, object.BOOLEAN_OBJ},
		{"(1 <= 1) != true", false, object.BOOLEAN_OBJ},

		{"(1 > 2) == true", false, object.BOOLEAN_OBJ},
		{"(1 >= 2) == true", false, object.BOOLEAN_OBJ},
		{"(1 > 1) == true", false, object.BOOLEAN_OBJ},
		{"(1 >= 1) != true", false, object.BOOLEAN_OBJ},

		{"not true", false, object.BOOLEAN_OBJ},
		{"not false", true, object.BOOLEAN_OBJ},
		{"not not false", false, object.BOOLEAN_OBJ},
		{"not not null", false, object.BOOLEAN_OBJ},
		{"not? not? null", nil, object.NULL_OBJ},
		{"not not true", true, object.BOOLEAN_OBJ},
		{"not 5", false, object.BOOLEAN_OBJ},
		{"not not 5", true, object.BOOLEAN_OBJ},

		{"not (1 > 3)", true, object.BOOLEAN_OBJ},
		{"not(1 > 3)", true, object.BOOLEAN_OBJ},
		// precedence of not below that of greater than (parentheses not required)
		{"not 1 > 3", true, object.BOOLEAN_OBJ},

		// null and negation of null
		{"null", nil, object.NULL_OBJ},
		{"not? null", nil, object.NULL_OBJ},
		{"not? (if false {5})", nil, object.NULL_OBJ},
		{"not null", true, object.BOOLEAN_OBJ},
		{"not (if false {5})", true, object.BOOLEAN_OBJ},

		{"1 < 3 and 14 > 7", true, object.BOOLEAN_OBJ},
		{"1 < 3 or 14 > 7", true, object.BOOLEAN_OBJ},
		{"1 > 3 or 14 > 7", true, object.BOOLEAN_OBJ},
		{"1 > 3 or 14 < 7", false, object.BOOLEAN_OBJ},
		{"1 > 3 xor 14 > 7", true, object.BOOLEAN_OBJ},
		{"1 > 3 xor 14 < 7", false, object.BOOLEAN_OBJ},
	}

	runVmTests(t, tests, false, false)
}

func TestForwardOperator(t *testing.T) {
	tests := []vmTestCase{
		{`"abc" -> re/a/`, true, object.BOOLEAN_OBJ},
		{`"abc" -> re/A/`, false, object.BOOLEAN_OBJ},

		{`123.345 -> re/123/`, true, object.BOOLEAN_OBJ},
		{`123.345 -> re/13/`, false, object.BOOLEAN_OBJ},

		{`val x = fn(a) { a is number }
		  123.456 -> x`, true, object.BOOLEAN_OBJ},
		{`val x = fn(a) { a is number }
		  "asdf" -> x`, false, object.BOOLEAN_OBJ},

		{`"abc" -> len`, 3, object.NUMBER_OBJ},
		{`val z = "abc"; val f = fn(x) { len(x) * 7 }; z -> f`, 21, object.NUMBER_OBJ},

		{`if "abc" -> re/a/ {
			1
		} else { 2 }`, 1, object.NUMBER_OBJ},
		{`if "abc" -> re/A/ {
			1
		} else { 2 }`, 2, object.NUMBER_OBJ},

		{`"123" -> len -> re/^3$/`, true, object.BOOLEAN_OBJ},
		{`"123" -> number -> string`, "123", object.STRING_OBJ},
	}

	runVmTests(t, tests, false, false)
}

func TestInOperator(t *testing.T) {
	tests := []vmTestCase{
		{"1 in []", false, object.BOOLEAN_OBJ},
		{"1 in [2, 3, 4]", false, object.BOOLEAN_OBJ},
		{"1 in [1, 2, 3, 4]", true, object.BOOLEAN_OBJ},
		{"3 in [1, 2, 3, 4]", true, object.BOOLEAN_OBJ},
		{"5 in [1, 2, 3, 4]", false, object.BOOLEAN_OBJ},

		{`"abc" in fw/joe know abc/`, true, object.BOOLEAN_OBJ},
		{`"abc" in fw/joe know/`, false, object.BOOLEAN_OBJ},

		{"1 in {:}", false, object.BOOLEAN_OBJ},
		{"1 in {1:2}", false, object.BOOLEAN_OBJ},
		{"1 in {2:1}", true, object.BOOLEAN_OBJ},
		{"2 in {1:2}", true, object.BOOLEAN_OBJ},

		{`"abc" in {:}`, false, object.BOOLEAN_OBJ},
		{`"abc" in {"abc" : 123}`, false, object.BOOLEAN_OBJ},
		{`"abc" in {123 : "abc"}`, true, object.BOOLEAN_OBJ},
		{`"abc" in {123 : "abcd"}`, false, object.BOOLEAN_OBJ},

		{`"abc" in ""`, false, object.BOOLEAN_OBJ},
		{`"abc" in "abcd"`, true, object.BOOLEAN_OBJ},
		{`"bc" in "abcd"`, true, object.BOOLEAN_OBJ},
		{`"bd" in "abcd"`, false, object.BOOLEAN_OBJ},

		{`97 in ""`, false, object.BOOLEAN_OBJ},
		{`97 in "abc"`, true, object.BOOLEAN_OBJ},
		{`100 in "abc"`, false, object.BOOLEAN_OBJ},

		// not in
		{"1 not in []", true, object.BOOLEAN_OBJ},
		{"1 not in [2, 3, 4]", true, object.BOOLEAN_OBJ},
		{"1 not in [1, 2, 3, 4]", false, object.BOOLEAN_OBJ},
		{"3 not in [1, 2, 3, 4]", false, object.BOOLEAN_OBJ},
		{"5 not in [1, 2, 3, 4]", true, object.BOOLEAN_OBJ},

		{`"abc" not in fw/joe know abc/`, false, object.BOOLEAN_OBJ},
		{`"abc" not in fw/joe know/`, true, object.BOOLEAN_OBJ},

		{"1 not in {:}", true, object.BOOLEAN_OBJ},
		{"1 not in {1:2}", true, object.BOOLEAN_OBJ},
		{"1 not in {2:1}", false, object.BOOLEAN_OBJ},
		{"2 not in {1:2}", false, object.BOOLEAN_OBJ},

		{`"abc" not in {:}`, true, object.BOOLEAN_OBJ},
		{`"abc" not in {"abc" : 123}`, true, object.BOOLEAN_OBJ},
		{`"abc" not in {123 : "abc"}`, false, object.BOOLEAN_OBJ},
		{`"abc" not in {123 : "abcd"}`, true, object.BOOLEAN_OBJ},

		{`"abc" not in ""`, true, object.BOOLEAN_OBJ},
		{`"abc" not in "abcd"`, false, object.BOOLEAN_OBJ},
		{`"bc" not in "abcd"`, false, object.BOOLEAN_OBJ},
		{`"bd" not in "abcd"`, true, object.BOOLEAN_OBJ},

		{`97 not in ""`, true, object.BOOLEAN_OBJ},
		{`97 not in "abc"`, false, object.BOOLEAN_OBJ},
		{`100 not in "abc"`, true, object.BOOLEAN_OBJ},

		{`var x, y = 0, [4, 5, 6, 7]
		  while x not in y { x += 1 }
		  x`, 4, object.NUMBER_OBJ},
	}

	runVmTests(t, tests, false, false)
}

func TestOfOperator(t *testing.T) {
	tests := []vmTestCase{
		{"1 of []", false, object.BOOLEAN_OBJ},
		{"1 of [2, 3, 4]", true, object.BOOLEAN_OBJ},
		{"1 of [1, 2, 3, 4]", true, object.BOOLEAN_OBJ},
		{"3 of [1, 2, 3, 4]", true, object.BOOLEAN_OBJ},
		{"5 of [1, 2, 3, 4]", false, object.BOOLEAN_OBJ},

		{"-1 of []", false, object.BOOLEAN_OBJ},
		{"-1 of [2, 3, 4]", true, object.BOOLEAN_OBJ},
		{"-1 of [1, 2, 3, 4]", true, object.BOOLEAN_OBJ},
		{"-3 of [1, 2, 3, 4]", true, object.BOOLEAN_OBJ},
		{"-5 of [1, 2, 3, 4]", false, object.BOOLEAN_OBJ},

		{"1 of {:}", false, object.BOOLEAN_OBJ},
		{"1 of {1:2}", true, object.BOOLEAN_OBJ},
		{"1 of {2:1}", false, object.BOOLEAN_OBJ},
		{"2 of {1:2}", false, object.BOOLEAN_OBJ},

		{`"abc" of {:}`, false, object.BOOLEAN_OBJ},
		{`"abc" of {"abc" : 123}`, true, object.BOOLEAN_OBJ},
		{`"abc" of {123 : "abc"}`, false, object.BOOLEAN_OBJ},
		{`"abc" of {"abcd": 123}`, false, object.BOOLEAN_OBJ},

		{`1 of ""`, false, object.BOOLEAN_OBJ},
		{`97 of "abc"`, false, object.BOOLEAN_OBJ},
		{`1 of "abc"`, true, object.BOOLEAN_OBJ},
		{`3 of "abc"`, true, object.BOOLEAN_OBJ},
		{`4 of "abc"`, false, object.BOOLEAN_OBJ},

		{`-1 of ""`, false, object.BOOLEAN_OBJ},
		{`-97 of "abc"`, false, object.BOOLEAN_OBJ},
		{`-1 of "abc"`, true, object.BOOLEAN_OBJ},
		{`-3 of "abc"`, true, object.BOOLEAN_OBJ},
		{`-4 of "abc"`, false, object.BOOLEAN_OBJ},

		// not of
		{"1 not of []", true, object.BOOLEAN_OBJ},
		{"1 not of [2, 3, 4]", false, object.BOOLEAN_OBJ},
		{"1 not of [1, 2, 3, 4]", false, object.BOOLEAN_OBJ},
		{"3 not of [1, 2, 3, 4]", false, object.BOOLEAN_OBJ},
		{"5 not of [1, 2, 3, 4]", true, object.BOOLEAN_OBJ},

		{"-1 not of []", true, object.BOOLEAN_OBJ},
		{"-1 not of [2, 3, 4]", false, object.BOOLEAN_OBJ},
		{"-1 not of [1, 2, 3, 4]", false, object.BOOLEAN_OBJ},
		{"-3 not of [1, 2, 3, 4]", false, object.BOOLEAN_OBJ},
		{"-5 not of [1, 2, 3, 4]", true, object.BOOLEAN_OBJ},

		{"1 not of {:}", true, object.BOOLEAN_OBJ},
		{"1 not of {1:2}", false, object.BOOLEAN_OBJ},
		{"1 not of {2:1}", true, object.BOOLEAN_OBJ},
		{"2 not of {1:2}", true, object.BOOLEAN_OBJ},

		{`"abc" not of {:}`, true, object.BOOLEAN_OBJ},
		{`"abc" not of {"abc" : 123}`, false, object.BOOLEAN_OBJ},
		{`"abc" not of {123 : "abc"}`, true, object.BOOLEAN_OBJ},
		{`"abc" not of {"abcd": 123}`, true, object.BOOLEAN_OBJ},

		{`1 not of ""`, true, object.BOOLEAN_OBJ},
		{`97 not of "abc"`, true, object.BOOLEAN_OBJ},
		{`1 not of "abc"`, false, object.BOOLEAN_OBJ},
		{`3 not of "abc"`, false, object.BOOLEAN_OBJ},
		{`4 not of "abc"`, true, object.BOOLEAN_OBJ},

		{`-1 not of ""`, true, object.BOOLEAN_OBJ},
		{`-97 not of "abc"`, true, object.BOOLEAN_OBJ},
		{`-1 not of "abc"`, false, object.BOOLEAN_OBJ},
		{`-3 not of "abc"`, false, object.BOOLEAN_OBJ},
		{`-4 not of "abc"`, true, object.BOOLEAN_OBJ},

		{`var x, y = 4, []
		  while x not of y { y ~= [7] }
		  len y`, 4, object.NUMBER_OBJ},
	}

	runVmTests(t, tests, false, false)
}

func TestNonDBLogicalOperators(t *testing.T) {
	tests := []vmTestCase{
		// non-db without null...
		{"false and false", false, object.BOOLEAN_OBJ},
		{"false and true", false, object.BOOLEAN_OBJ},
		{"true and false", false, object.BOOLEAN_OBJ},
		{"true and true", true, object.BOOLEAN_OBJ},

		{"false or false", false, object.BOOLEAN_OBJ},
		{"false or true", true, object.BOOLEAN_OBJ},
		{"true or false", true, object.BOOLEAN_OBJ},
		{"true or true", true, object.BOOLEAN_OBJ},

		{"false nand false", true, object.BOOLEAN_OBJ},
		{"false nand true", true, object.BOOLEAN_OBJ},
		{"true nand false", true, object.BOOLEAN_OBJ},
		{"true nand true", false, object.BOOLEAN_OBJ},

		{"false nor false", true, object.BOOLEAN_OBJ},
		{"false nor true", false, object.BOOLEAN_OBJ},
		{"true nor false", false, object.BOOLEAN_OBJ},
		{"true nor true", false, object.BOOLEAN_OBJ},

		{"false xor false", false, object.BOOLEAN_OBJ},
		{"false xor true", true, object.BOOLEAN_OBJ},
		{"true xor false", true, object.BOOLEAN_OBJ},
		{"true xor true", false, object.BOOLEAN_OBJ},

		{"false nxor false", true, object.BOOLEAN_OBJ},
		{"false nxor true", false, object.BOOLEAN_OBJ},
		{"true nxor false", false, object.BOOLEAN_OBJ},
		{"true nxor true", true, object.BOOLEAN_OBJ},

		// non-db with null...
		{"null and null", false, object.BOOLEAN_OBJ},
		{"null and true", false, object.BOOLEAN_OBJ},
		{"true and null", false, object.BOOLEAN_OBJ},
		{"null and false", false, object.BOOLEAN_OBJ},
		{"false and null", false, object.BOOLEAN_OBJ},

		{"null or null", false, object.BOOLEAN_OBJ},
		{"null or true", true, object.BOOLEAN_OBJ},
		{"true or null", true, object.BOOLEAN_OBJ},
		{"null or false", false, object.BOOLEAN_OBJ},
		{"false or null", false, object.BOOLEAN_OBJ},

		{"null nand null", true, object.BOOLEAN_OBJ},
		{"null nand true", true, object.BOOLEAN_OBJ},
		{"true nand null", true, object.BOOLEAN_OBJ},
		{"null nand false", true, object.BOOLEAN_OBJ},
		{"false nand null", true, object.BOOLEAN_OBJ},

		{"null nor null", true, object.BOOLEAN_OBJ},
		{"null nor true", false, object.BOOLEAN_OBJ},
		{"true nor null", false, object.BOOLEAN_OBJ},
		{"null nor false", true, object.BOOLEAN_OBJ},
		{"false nor null", true, object.BOOLEAN_OBJ},

		{"null xor null", false, object.BOOLEAN_OBJ},
		{"null xor true", true, object.BOOLEAN_OBJ},
		{"true xor null", true, object.BOOLEAN_OBJ},
		{"null xor false", false, object.BOOLEAN_OBJ},
		{"false xor null", false, object.BOOLEAN_OBJ},

		{"null nxor null", true, object.BOOLEAN_OBJ},
		{"null nxor true", false, object.BOOLEAN_OBJ},
		{"true nxor null", false, object.BOOLEAN_OBJ},
		{"null nxor false", true, object.BOOLEAN_OBJ},
		{"false nxor null", true, object.BOOLEAN_OBJ},
	}
	runVmTests(t, tests, false, false)
}

func TestDBLogicalOperators(t *testing.T) {
	tests := []vmTestCase{
		// db without null...
		{"false and? false", false, object.BOOLEAN_OBJ},
		{"false and? true", false, object.BOOLEAN_OBJ},
		{"true and? false", false, object.BOOLEAN_OBJ},
		{"true and? true", true, object.BOOLEAN_OBJ},

		{"false or? false", false, object.BOOLEAN_OBJ},
		{"false or? true", true, object.BOOLEAN_OBJ},
		{"true or? false", true, object.BOOLEAN_OBJ},
		{"true or? true", true, object.BOOLEAN_OBJ},

		{"false xor? false", false, object.BOOLEAN_OBJ},
		{"false xor? true", true, object.BOOLEAN_OBJ},
		{"true xor? false", true, object.BOOLEAN_OBJ},
		{"true xor? true", false, object.BOOLEAN_OBJ},

		{"false nxor? false", true, object.BOOLEAN_OBJ},
		{"false nxor? true", false, object.BOOLEAN_OBJ},
		{"true nxor? false", false, object.BOOLEAN_OBJ},
		{"true nxor? true", true, object.BOOLEAN_OBJ},

		{"false nand? false", true, object.BOOLEAN_OBJ},
		{"false nand? true", true, object.BOOLEAN_OBJ},
		{"true nand? false", true, object.BOOLEAN_OBJ},
		{"true nand? true", false, object.BOOLEAN_OBJ},

		{"false nor? false", true, object.BOOLEAN_OBJ},
		{"false nor? true", false, object.BOOLEAN_OBJ},
		{"true nor? false", false, object.BOOLEAN_OBJ},
		{"true nor? true", false, object.BOOLEAN_OBJ},

		// db with null...
		{"null and? null", nil, object.NULL_OBJ},
		{"null and? true", nil, object.NULL_OBJ},
		{"true and? null", nil, object.NULL_OBJ},
		{"null and? false", nil, object.NULL_OBJ},
		{"false and? null", nil, object.NULL_OBJ},

		{"null or? null", nil, object.NULL_OBJ},
		{"null or? true", nil, object.NULL_OBJ},
		{"true or? null", nil, object.NULL_OBJ},
		{"null or? false", nil, object.NULL_OBJ},
		{"false or? null", nil, object.NULL_OBJ},

		{"null xor? null", nil, object.NULL_OBJ},
		{"null xor? true", nil, object.NULL_OBJ},
		{"true xor? null", nil, object.NULL_OBJ},
		{"null xor? false", nil, object.NULL_OBJ},
		{"false xor? null", nil, object.NULL_OBJ},

		{"null nxor? null", nil, object.NULL_OBJ},
		{"null nxor? true", nil, object.NULL_OBJ},
		{"true nxor? null", nil, object.NULL_OBJ},
		{"null nxor? false", nil, object.NULL_OBJ},
		{"false nxor? null", nil, object.NULL_OBJ},

		{"null nand? null", nil, object.NULL_OBJ},
		{"null nand? true", nil, object.NULL_OBJ},
		{"true nand? null", nil, object.NULL_OBJ},
		{"null nand? false", nil, object.NULL_OBJ},
		{"false nand? null", nil, object.NULL_OBJ},

		{"null nor? null", nil, object.NULL_OBJ},
		{"null nor? true", nil, object.NULL_OBJ},
		{"true nor? null", nil, object.NULL_OBJ},
		{"null nor? false", nil, object.NULL_OBJ},
		{"false nor? null", nil, object.NULL_OBJ},
	}
	runVmTests(t, tests, false, false)
}

func TestDBComparisonOperators(t *testing.T) {
	tests := []vmTestCase{
		{"0 <? 1", true, object.BOOLEAN_OBJ},
		{"0 <? null", nil, object.NULL_OBJ},
		{"null <? 1", nil, object.NULL_OBJ},

		{"0 >? 1", false, object.BOOLEAN_OBJ},
		{"0 >? null", nil, object.NULL_OBJ},
		{"null >? 1", nil, object.NULL_OBJ},

		{"0 <=? 1", true, object.BOOLEAN_OBJ},
		{"0 <=? null", nil, object.NULL_OBJ},
		{"null <=? 1", nil, object.NULL_OBJ},

		{"0 >=? 1", false, object.BOOLEAN_OBJ},
		{"0 >=? null", nil, object.NULL_OBJ},
		{"null >=? 1", nil, object.NULL_OBJ},

		{"0 !=? 1", true, object.BOOLEAN_OBJ},
		{"0 != null", true, object.BOOLEAN_OBJ},
		{"null != 0", true, object.BOOLEAN_OBJ},
		{"0 !=? null", nil, object.NULL_OBJ},
		{"null !=? 1", nil, object.NULL_OBJ},
		{"null != null", false, object.BOOLEAN_OBJ},
		{"null !=? null", nil, object.NULL_OBJ},

		{"0 ==? 1", false, object.BOOLEAN_OBJ},
		{"0 ==? null", nil, object.NULL_OBJ},
		{"0 == null", false, object.BOOLEAN_OBJ},
		{"null == 0", false, object.BOOLEAN_OBJ},
		{"null ==? 1", nil, object.NULL_OBJ},
		{"null == null", true, object.BOOLEAN_OBJ},
		{"null ==? null", nil, object.NULL_OBJ},

		{"21 div 7", true, object.BOOLEAN_OBJ},
		{"21 div? null", nil, object.NULL_OBJ},
		{"null div? 7", nil, object.NULL_OBJ},
		{"21 ndiv 7", false, object.BOOLEAN_OBJ},
		{"21 ndiv? null", nil, object.NULL_OBJ},
		{"null ndiv? 7", nil, object.NULL_OBJ},
	}
	runVmTests(t, tests, false, false)
}

func TestOtherComparisons(t *testing.T) {
	tests := []vmTestCase{
		{`"abc" == "abc"`, true, object.BOOLEAN_OBJ},
		{`"abc" == "a" ~ "bc"`, true, object.BOOLEAN_OBJ},
		{`"abc" != "abc"`, false, object.BOOLEAN_OBJ},
		{`"abc" != "a" ~ "bc"`, false, object.BOOLEAN_OBJ},
		{`"def" == "abc"`, false, object.BOOLEAN_OBJ},
		{`"def" == "a" ~ "bc"`, false, object.BOOLEAN_OBJ},
		{`"def" != "abc"`, true, object.BOOLEAN_OBJ},
		{`"def" != "a" ~ "bc"`, true, object.BOOLEAN_OBJ},

		{`"ABC" > "abc"`, false, object.BOOLEAN_OBJ},
		{`"ABC" > "a" ~ "bc"`, false, object.BOOLEAN_OBJ},
		{`"ABC" >= "abc"`, false, object.BOOLEAN_OBJ},
		{`"ABC" >= "a" ~ "bc"`, false, object.BOOLEAN_OBJ},
		{`"abc" > "ABC"`, true, object.BOOLEAN_OBJ},
		{`"abc" > "A" ~ "bc"`, true, object.BOOLEAN_OBJ},
		{`"abc" >= "ABC"`, true, object.BOOLEAN_OBJ},
		{`"abc" >= "A" ~ "bc"`, true, object.BOOLEAN_OBJ},

		{`"ABC" < "abc"`, true, object.BOOLEAN_OBJ},
		{`"ABC" < "a" ~ "bc"`, true, object.BOOLEAN_OBJ},
		{`"ABC" <= "abc"`, true, object.BOOLEAN_OBJ},
		{`"ABC" <= "a" ~ "bc"`, true, object.BOOLEAN_OBJ},
		{`"abc" < "ABC"`, false, object.BOOLEAN_OBJ},
		{`"abc" < "A" ~ "bc"`, false, object.BOOLEAN_OBJ},
		{`"abc" <= "ABC"`, false, object.BOOLEAN_OBJ},
		{`"abc" <= "A" ~ "bc"`, false, object.BOOLEAN_OBJ},

		{`[1, 2, 3] == [1, 2, 3]`, true, object.BOOLEAN_OBJ},
		{`[1, 2, 3] == [1, 2] ~ [3]`, true, object.BOOLEAN_OBJ},
		{`[1, 2, 3] != [1, 2, 3]`, false, object.BOOLEAN_OBJ},
		{`[1, 2, 3] != [1, 2] ~ [3]`, false, object.BOOLEAN_OBJ},
		{`[2, 3, 1] == [1, 2, 3]`, false, object.BOOLEAN_OBJ},
		{`[2, 3, 1] == [1, 2] ~ [3]`, false, object.BOOLEAN_OBJ},
		{`[2, 3, 1] != [1, 2, 3]`, true, object.BOOLEAN_OBJ},
		{`[2, 3, 1] != [1, 2] ~ [3]`, true, object.BOOLEAN_OBJ},

		{`{1: 2, 3: 4} == {1: 2, 3: 4}`, true, object.BOOLEAN_OBJ},
		{`{1: 2, 3: 4} == {1: 2} ~ {3: 4}`, true, object.BOOLEAN_OBJ},
		{`{1: 2, 3: 4} != {1: 2, 3: 4}`, false, object.BOOLEAN_OBJ},
		{`{1: 2, 3: 4} != {1: 2} ~ {3: 4}`, false, object.BOOLEAN_OBJ},
		{`{3: 4, 1: 2} == {1: 2, 3: 4}`, true, object.BOOLEAN_OBJ},
		{`{3: 4, 1: 2} == {1: 2} ~ {3: 4}`, true, object.BOOLEAN_OBJ},
		{`{3: 4, 1: 2} != {1: 2, 3: 4}`, false, object.BOOLEAN_OBJ},
		{`{3: 4, 1: 2} != {1: 2} ~ {3: 4}`, false, object.BOOLEAN_OBJ},

		{`{1: 7, 3: 4} == {1: 2, 3: 4}`, false, object.BOOLEAN_OBJ},
		{`{1: 7, 3: 4} == {1: 2} ~ {3: 4}`, false, object.BOOLEAN_OBJ},
		{`{1: 7, 3: 4} != {1: 2, 3: 4}`, true, object.BOOLEAN_OBJ},
		{`{1: 7, 3: 4} != {1: 2} ~ {3: 4}`, true, object.BOOLEAN_OBJ},
		{`{3: 7, 1: 2} == {1: 2, 3: 4}`, false, object.BOOLEAN_OBJ},
		{`{3: 7, 1: 2} == {1: 2} ~ {3: 4}`, false, object.BOOLEAN_OBJ},
		{`{3: 7, 1: 2} != {1: 2, 3: 4}`, true, object.BOOLEAN_OBJ},
		{`{3: 7, 1: 2} != {1: 2} ~ {3: 4}`, true, object.BOOLEAN_OBJ},

		{`1..7 == 1..7`, true, object.BOOLEAN_OBJ},
		{`1..7 == 1..21/3`, true, object.BOOLEAN_OBJ},
		{`3..7 == 1..7`, false, object.BOOLEAN_OBJ},
		{`3..7 == 1..21/3`, false, object.BOOLEAN_OBJ},
		{`1..7 != 1..7`, false, object.BOOLEAN_OBJ},
		{`1..7 != 1..21/3`, false, object.BOOLEAN_OBJ},
		{`3..7 != 1..7`, true, object.BOOLEAN_OBJ},
		{`3..7 != 1..21/3`, true, object.BOOLEAN_OBJ},
	}

	runVmTests(t, tests, false, false)
}

func TestTruthiness(t *testing.T) {
	tests := []vmTestCase{
		{"not not true", true, object.BOOLEAN_OBJ},
		{"not not false", false, object.BOOLEAN_OBJ},

		{"not not null", false, object.BOOLEAN_OBJ},

		{"not not []", false, object.BOOLEAN_OBJ},
		{"not not [1]", true, object.BOOLEAN_OBJ},

		{"not not {:}", false, object.BOOLEAN_OBJ},
		{"not not {1: 3}", true, object.BOOLEAN_OBJ},

		{"not not 0", false, object.BOOLEAN_OBJ},
		{"not not 1", true, object.BOOLEAN_OBJ},
		{"not not -1", true, object.BOOLEAN_OBJ},

		{"not not 0+0i", false, object.BOOLEAN_OBJ},
		{"not not 1+0i", true, object.BOOLEAN_OBJ},
		{"not not 0+1i", true, object.BOOLEAN_OBJ},
		{"not not 1+1i", true, object.BOOLEAN_OBJ},

		{`not not zls`, false, object.BOOLEAN_OBJ},
		{`not not "A"`, true, object.BOOLEAN_OBJ},

		{"not not re//", false, object.BOOLEAN_OBJ},
		{"not not re/1/", true, object.BOOLEAN_OBJ},

		{"not not 3 .. 1", false, object.BOOLEAN_OBJ},
		{"not not 1 .. 3", true, object.BOOLEAN_OBJ},
		{"not not 3 .. 3", true, object.BOOLEAN_OBJ},

		// proleptic Gregorian as false
		{"not not dt/1581-10-15/", false, object.BOOLEAN_OBJ},

		// ... test with UTC
		{"not not dt/1582-10-14 23:59:59Z/", false, object.BOOLEAN_OBJ},
		{"not not dt/1582-10-14 23:59:59.999999999Z/", false, object.BOOLEAN_OBJ},
		{"not not dt/1582-10-14 00:00:00Z/", false, object.BOOLEAN_OBJ},
		{"not not dt/1582-10-15 00:00:00Z/", true, object.BOOLEAN_OBJ},
		{"not not dt/1582-10-16 00:00:00Z/", true, object.BOOLEAN_OBJ},
		{"not not dt/1582-09-14 00:00:00Z/", false, object.BOOLEAN_OBJ},
		{"not not dt/1582-09-15 00:00:00Z/", false, object.BOOLEAN_OBJ},
		{"not not dt/1582-09-16 00:00:00Z/", false, object.BOOLEAN_OBJ},
		{"not not dt/1582-11-14 00:00:00Z/", true, object.BOOLEAN_OBJ},
		{"not not dt/1582-11-15 00:00:00Z/", true, object.BOOLEAN_OBJ},
		{"not not dt/1582-11-16 00:00:00Z/", true, object.BOOLEAN_OBJ},

		// ... regardless of time zone used
		{"not not dt/1582-10-14 23:59:59-05:00/", false, object.BOOLEAN_OBJ},
		{"not not dt/1582-10-14 23:59:59.999999999-05:00/", false, object.BOOLEAN_OBJ},
		{"not not dt/1582-10-14 00:00:00-05:00/", false, object.BOOLEAN_OBJ},
		{"not not dt/1582-10-15 00:00:00-05:00/", true, object.BOOLEAN_OBJ},
		{"not not dt/1582-10-16 00:00:00-05:00/", true, object.BOOLEAN_OBJ},
		{"not not dt/1582-09-14 00:00:00-05:00/", false, object.BOOLEAN_OBJ},
		{"not not dt/1582-09-15 00:00:00-05:00/", false, object.BOOLEAN_OBJ},
		{"not not dt/1582-09-16 00:00:00-05:00/", false, object.BOOLEAN_OBJ},
		{"not not dt/1582-11-14 00:00:00-05:00/", true, object.BOOLEAN_OBJ},
		{"not not dt/1582-11-15 00:00:00-05:00/", true, object.BOOLEAN_OBJ},
		{"not not dt/1582-11-16 00:00:00-05:00/", true, object.BOOLEAN_OBJ},

		{"not not dr/0D/", false, object.BOOLEAN_OBJ},
		{"not not dr/1D/", true, object.BOOLEAN_OBJ},

		{`not not(fn() {})`, true, object.BOOLEAN_OBJ},
		{`not not(fn*() {})`, false, object.BOOLEAN_OBJ},
		{`not not(len)`, true, object.BOOLEAN_OBJ},
		{`not not(cd)`, false, object.BOOLEAN_OBJ},
	}

	runVmTests(t, tests, false, false)
}

func TestIfExpressions(t *testing.T) {
	tests := []vmTestCase{
		{"if false {7}", nil, object.NULL_OBJ},
		{"if 1 > 2 {7}", nil, object.NULL_OBJ},
		{"if true {7}", "7", object.NUMBER_OBJ},
		{"if true {7} else {14}", "7", object.NUMBER_OBJ},
		{"if false {7} else {14}", "14", object.NUMBER_OBJ},
		{"if 10 {7}", "7", object.NUMBER_OBJ},
		{"if 1 < 2 {7}", "7", object.NUMBER_OBJ},
		{"if 1 < 2 {7} else {14}", "7", object.NUMBER_OBJ},
		{"if 1 > 2 {7} else {14}", "14", object.NUMBER_OBJ},
		{"if 1 > 2 {7} else if 5 == 7 {14} else if false {21} else {28}", "28", object.NUMBER_OBJ},
		{"if 1 > 2 {7} else if 5 <= 7 {14} else if false {21} else {28}", "14", object.NUMBER_OBJ},
		{"if true { } else if false { 2 } else if true { 3 }", nil, object.NULL_OBJ},

		{`if 1 > 2 {
			7
		  } else if 5 <= 7 {
		    14
		  } else if false {
		    21
		  } else {
		    28
		  }`,
			"14", object.NUMBER_OBJ,
		},
	}

	runVmTests(t, tests, false, false)
}

func TestShortenedFormIfExpressions(t *testing.T) {
	tests := []vmTestCase{
		{"if(false: 7)", nil, object.NULL_OBJ},
		{"if(1 > 2: 7)", nil, object.NULL_OBJ},
		{"if(true: 7)", "7", object.NUMBER_OBJ},
		{"if(true: 7; 14)", "7", object.NUMBER_OBJ},
		{"if(false: 7; 14)", "14", object.NUMBER_OBJ},
		{"if(10: 7)", "7", object.NUMBER_OBJ},
		{"if(1 < 2: 7)", "7", object.NUMBER_OBJ},
		{"if(1 < 2: 7; 14)", "7", object.NUMBER_OBJ},
		{"if(1 > 2: 7; 14)", "14", object.NUMBER_OBJ},
		{"if(1 > 2: 7; 5 == 7: 14; false: 21; 28)", "28", object.NUMBER_OBJ},
		{"if(1 > 2: 7; 5 <= 7: 14; false: 21; 28)", "14", object.NUMBER_OBJ},
		{"if(true: null; false: 2; true: 3)", nil, object.NULL_OBJ},
		{"if(false: null; false: 2; true: 3)", "3", object.NUMBER_OBJ},

		{`if(1 > 2: 7
		     5 <= 7: 14
		     false: 21
		     28)`,
			"14", object.NUMBER_OBJ,
		},
	}

	runVmTests(t, tests, false, false)
}

func TestShortCircuitingOperations(t *testing.T) {
	tests := []vmTestCase{
		{`val x = "abc"; len(x) >= 7 and x[7] == 'o'`, false, object.BOOLEAN_OBJ},       // sc
		{`val x = "abcd yoyo"; len(x) >= 7 and x[7] == 'o'`, true, object.BOOLEAN_OBJ},  // not sc
		{`val x = "abcd yoyo"; len(x) >= 7 and x[7] == 'y'`, false, object.BOOLEAN_OBJ}, // not sc

		// scope blocks in expression context
		{`var y = 7; true or { y = 14 }; y`, "7", object.NUMBER_OBJ},        // sc
		{`var y = 7; true and { y = 14 }; y`, "14", object.NUMBER_OBJ},      // not sc
		{"var x = 7; [1, 2, 3][3; { x = 14 }]; x", "7", object.NUMBER_OBJ},  // sc
		{"var x = 7; [1, 2, 3][4; { x = 14 }]; x", "14", object.NUMBER_OBJ}, // not sc

		{`null and? true`, nil, object.NULL_OBJ},     // sc
		{`null or? true`, nil, object.NULL_OBJ},      // sc
		{`null and true`, false, object.BOOLEAN_OBJ}, // not sc
		{`null or true`, true, object.BOOLEAN_OBJ},   // not sc

		{`null >? 7`, nil, object.NULL_OBJ},   // sc
		{`7 >? null`, nil, object.NULL_OBJ},   // not sc
		{`7 >? 7`, false, object.BOOLEAN_OBJ}, // not sc
		{`14 >? 7`, true, object.BOOLEAN_OBJ}, // not sc

		// On db operations, you can only short-circuit if the left is null.
		{`true and? null`, nil, object.NULL_OBJ},  // not sc
		{`false and? null`, nil, object.NULL_OBJ}, // not sc
		{`true or? null`, nil, object.NULL_OBJ},   // not sc
		{`false or? null`, nil, object.NULL_OBJ},  // not sc

		{`true and? true`, true, object.BOOLEAN_OBJ},    // not sc
		{`false and? true`, false, object.BOOLEAN_OBJ},  // not sc
		{`true and? false`, false, object.BOOLEAN_OBJ},  // not sc
		{`false and? false`, false, object.BOOLEAN_OBJ}, // not sc
		{`true or? true`, true, object.BOOLEAN_OBJ},     // not sc
		{`false or? true`, true, object.BOOLEAN_OBJ},    // not sc
		{`true or? false`, true, object.BOOLEAN_OBJ},    // not sc
		{`false or? false`, false, object.BOOLEAN_OBJ},  // not sc
	}

	runVmTests(t, tests, false, false)
}

func TestConditionalDeclaration(t *testing.T) {
	var conditionalDeclNoResult interface{} = nil
	conditionalDeclNoResultType := object.NULL_OBJ

	tests := []vmTestCase{
		// declarations in the try block
		{`123 / 0
		  val yo = 789
		  catch { yo }`,
			conditionalDeclNoResult, conditionalDeclNoResultType},

		{`123 / 1
		  val yo = 789
		  catch { yo }`,
			"789", object.NUMBER_OBJ},
	}

	runVmTests(t, tests, false, false)
}

func TestSwitchExpressions(t *testing.T) {
	tests := []vmTestCase{
		// match single value
		{`val x = 123; val y = 159
			switch x { 
				case 120: 1
				case y: 2
				case 123: 3
				default: 4
			}`,
			"3", object.NUMBER_OBJ,
		},
		// default with no match
		{`val x = 123; val y = 159
			switch x { 
				case 120: 1
				case y: 2
				case 159: 3
				default: 4
			}`,
			"4", object.NUMBER_OBJ,
		},
		// no default and no match
		{`val x = 123; val y = 159
			switch x { 
				case 120: 1
				case y: 2
				case 159: 3
			}`,
			nil, object.NULL_OBJ,
		},

		// both match 1 condition
		{`val x = 123; val y = 123;
			switch[and] x, y {
				case 123: 1
				case _, x: 2
				case 170: 3
				default: 4
			}`,
			"1", object.NUMBER_OBJ,
		},
		// 1 match and 1 no-op
		{`val x = 123; val y = 123;
			switch x, y {
				case 123, _: 1
				case _, x: 2
				case 123: 3
				default: 4 
			}`,
			"1", object.NUMBER_OBJ,
		},
		// x... == y
		{`val x = 123; val y = 123;
			switch x, y {
				case 120, _: 1
				case _, x: 2
				case 123: 3
				default: 4
			}`,
			"2", object.NUMBER_OBJ,
		},
		// both match 3rd condition
		{`val x = 123; val y = 123;
			switch x, y {
				case 120, _: 1
				case _, 159: 2
				case[and] 123: 3
				default: 4
			}`,
			"3", object.NUMBER_OBJ,
		},
		// no match
		{`val x = 123; val y = 123;
			switch x, y {
				case 120, _: 1
				case _, 159: 2
				case[and] 120: 3
				default: 4
			}`,
			"4", object.NUMBER_OBJ,
		},
		// each match own condition
		{`val x = 123; val y = 159;
			switch[and] x, y {
				case 120, _: 1
				case 123: 2
				case 123, 159: 3
				default: 4
			}`,
			"3", object.NUMBER_OBJ,
		},
		// not both matching same condition
		{`val x = 123; val y = 159;
			switch x, y {
				case 120, _: 1
				case _, x: 2
				case[and] 123: 3
				default: 4
			}`,
			"4", object.NUMBER_OBJ,
		},
		{`val x = 123; val y = 159;
			switch x, y {
				case 120, _: 1
				case _, x: 2
				case[and] 159: 3
				default: 4
			}`,
			"4", object.NUMBER_OBJ,
		},

		// 1 match and 1 no-op
		{`val x = 123; val y = 159;
			switch x, y {
				case 120, _: 1
				case _, x: 2
				case 123, _: 3
				default: 4
			}`,
			"3", object.NUMBER_OBJ,
		},
		{`val x = 123; val y = 159;
			switch x, y {
				case 120, _: 1
				case _, x: 2
				case _, 159: 3
				default: 4
			}`,
			"3", object.NUMBER_OBJ,
		},

		// with regex literal as condition
		{`val x = 123; val y = 159
			switch x { 
				case 120: 1
				case y: 2
				case -> RE/^1\d\d/: 3		# matches here
				default: 4
			}`,
			"3", object.NUMBER_OBJ,
		},
		// with regex literal as variable
		{`val x = 123; val y = 159
			switch -> RE/\d0$/ { 
				case 123: 1
				case y: 2
				case x: 3
				case 40: 4				# matches here
				default: 5
			}`,
			"4", object.NUMBER_OBJ,
		},
		// with regex literal as both, without using the matching operator
		{`val x = 123; val y = 159
			switch RE/\d0$/ { 
				case 123: 1
				case y: 2
				case RE/\d0$/: 3			# matches here (regex == regex)
				case 40: 4				
				default: 5
			}`,
			"3", object.NUMBER_OBJ,
		},

		// with function as condition
		{`val x = 123; val y = 159
		  val f = fn(x) {x -> RE/^1\d\d/}
			switch x { 
				case 120: 1
				case y: 2
				case -> f: 3		# matches here
				default: 4
			}`,
			3, object.NUMBER_OBJ,
		},

		{`val x = 123; val y = 159
		  val f = fn(x) {x -> RE/^2\d\d/}
			switch x { 
				case 120: 1
				case y: 2
				case -> f: 3
				default: 4
			}`,
			4, object.NUMBER_OBJ,
		},

		// more than 2...
		{`val x = 123; val y = 159; val z = 753
			switch[and] x, y, z {
				case _, x, _: "nada"
				case 123, 180, 753: "?"
				case 123: ""		# doesn't match this case b/c a single test with no no-ops matches all variables
				case 123, 159, 753: "zip"
				case 123, 159, 789: "code"
				default: "nope"
			}`,
			"zip", object.STRING_OBJ,
		},
		{`val x = 123; val y = 159; val z = 753
			switch[and] x, y, z {
				case _, x, _: "nada"
				case 123, 180, 753: "?"
				case 123, 159, 789: "code"
				case 123, _: "something"	# matches here
				case 123, 159, 753: "zip"	# would match this, but the previous case already matched
				default: "nope"
			}`,
			"something", object.STRING_OBJ,
		},

		// one expression with multiple conditions
		{`switch[and] > 7 {
			 case 7, 9: 10
			 case 9, 8: 11
			 case 12: 14
		  }`,
			"11", object.NUMBER_OBJ,
		},
		{`switch > 7 {
			 case 7, 9: 10
			 case 9, 8: 11
			 case 12: 14
		  }`,
			"10", object.NUMBER_OBJ,
		},
		{`switch[and] -> re/a/ {
			 case 7, 9: 10
			 case "z", "b": 11
			 case "abc", "a": 12
			 case 123: 14
		  }`,
			"12", object.NUMBER_OBJ,
		},
		{`switch[and] -> re/a/ {
			 case 7, 9: 10
			 case "z", "b": 11
			 case "abc": 12
			 case 123: 14
		  }`,
			"12", object.NUMBER_OBJ,
		},

		// individual alternative operators (default ==) with nil left
		{`val x = 123; val y = 159
					switch x, y {
						case[and] > 120: 1		# match; both over 120
						case _, x: 2
						case _, 159: 3
						default: 4
					}`,
			"1", object.NUMBER_OBJ,
		},
		{`val x = 123; val y = 159
					switch x, y {
						case[and] > 120, 123: 1
						case _, x: 2
						case _, 159: 3		# match
						default: 4
					}`,
			"3", object.NUMBER_OBJ,
		},
		{`val x = 123; val y = 159
					switch x, y {
						case[and] 120, <= 123: 1
						case _, x: 2
						case _, >= 159: 3		# match
						default: 4
					}`,
			"3", object.NUMBER_OBJ,
		},
		{`val x = 123; val y = 159
					switch[and] x, y {
						case > 120, < 123: 1
						case >= 123, <= x: 2
						case > 150, <= 159: 3
						case > 50, <= 159: 4	# match
						case > 500, <= 159: 5
						default: 6
					}`,
			"4", object.NUMBER_OBJ,
		},

		// individual alternative operators (default ==) with nil right
		{`val x = 123; val y = 159
					switch[and] x, y {
						case 120 <, 123 >  : 1
						case 123 <=, x >= : 2
						case 150 <, 159 >= : 3
						case 50 <, 159 >=  : 4	# match
						case 500 <, 159 >= : 5
						default: 6
					}`,
			"4", object.NUMBER_OBJ,
		},
		{`val x = 123; val y = 159
					switch[and] x, y {
						case 123, 123: 1
						case 5 <, 159 == : 5
						default: 6
					}`,
			"5", object.NUMBER_OBJ,
		},

		// different expression default operator (default ==)
		{`val x = 123; val y = 159
			switch[and] != x, y != {
				case 123: 1
				case 123, 127: 2
				case _, 159: 3
				default: 4
			}`,
			"4", object.NUMBER_OBJ,
		},
		{`val x = 123; val y = 159
			switch[and] x !=, y != {
				case 123: 1
				case 123, 159: 2
				case _, 123: 3
				default: 4
			}`,
			"3", object.NUMBER_OBJ,
		},
		{`val x = 123; val y = 159
			switch[and] x !=, y != {
				case 7: 1
				case 123, 159: 2
				case _, 123: 3
				default: 4
			}`,
			"1", object.NUMBER_OBJ,
		},
		{`val x = 123; val y = 159
			switch[and] x <, y < {
				case 7: 1
				case 123, 159: 2
				case 159, 170: 3
				default: 4
			}`,
			"3", object.NUMBER_OBJ,
		},
		{`val x = 123; val y = 159
			switch[and] x <, y < {
				case 7: 1
				case == 123, == 159: 2
				case 159, 170: 3
				default: 4
			}`,
			"2", object.NUMBER_OBJ,
		},

		{`val x = 123; val y = 159
			switch[and] < x, y < {
				case 7: 1
				case 123, == 159: 2
				case 159, 170: 3
				default: 4
			}`,
			"4", object.NUMBER_OBJ,
		},
		{`val x = 123; val y = 159
			switch[and] < x, y < {
				case 7: 1
				case 17, == 159: 2
				case 159, 170: 3
				default: 4
			}`,
			"2", object.NUMBER_OBJ,
		},
		{`val x = 123; val y = 159
			switch[and] < x, y < {
				case 7: 1
				case 17, 170: 2
				case 159, 170: 3
				default: 4
			}`,
			"2", object.NUMBER_OBJ,
		},

		{`val x = 123; val y = 159
			switch[and] x >=, != y {
				case 700: 1
				case == 123, == 159: 2
				case 159, 170: 3
				default: 4
			}`,
			"2", object.NUMBER_OBJ,
		},
		{`val x = 123; val y = 159
			switch[and] x >=, != y {
				case 123: 1
				case == 123, == 159: 2
				case 159, 170: 3
				default: 4
			}`,
			"1", object.NUMBER_OBJ,
		},
		{`val x = 123; val y = 159
			switch[and] x >=, != y {
				case 123, 160: 1
				case == 123, == 159: 2
				case 159, 170: 3
				default: 4
			}`,
			"1", object.NUMBER_OBJ,
		},
		{`val x = 123; val y = 159
			switch[and] x >=, != y {
				case 700: 1
				case 123, 159: 2
				case < 159, 170: 3
				default: 4
			}`,
			"3", object.NUMBER_OBJ,
		},
		{`val x = 123; val y = 159
			switch[and] x >=, != y {
				case 700: 1
				case 123, 159: 2
				case 159, 170: 3
				default: 4
			}`,
			"4", object.NUMBER_OBJ,
		},

		// an odd test, to be sure it's working right in context
		{`switch 65 ~ {
				case 0: 1  # non-zero-length string ... true
				default: 4
			}`,
			1, object.NUMBER_OBJ,
		},

		// alternate case logical operator (default "and" between single case conditions)
		{`val x = 123; val y = 159
			switch[and] x, y {
				case[or] 7: 1
				case[or] 123, 789: 2		# match here
				case 159, 170: 3
				default: 4
			}`,
			"2", object.NUMBER_OBJ,
		},
		{`val x = true; val y = false; val z = null
			switch[and] x, y {
				case[or] true: 1			# match here
				case[xor] true, true: 2		
				case null: 3
				default: 4
			}`,
			"1", object.NUMBER_OBJ,
		},
		{`val x = true; val y = false; val z = null
			switch[and] x, y {
				case null: 1
				case[xor] true, true: 2		# match here
				case null: 3
				default: 4
			}`,
			"2", object.NUMBER_OBJ,
		},

		// explicit fallthrough
		{`val x = 123
			switch x {
				case > 100: 1
					fallthrough
				case 100: 2
			}`,
			"2", object.NUMBER_OBJ,
		},
		{`val x = 123
			switch x {
				case > 100: 1
					fallthrough
				case 100: 2
				case 200: 3
			}`,
			"2", object.NUMBER_OBJ,
		},
		{`val x = 123
			switch x {
				case > 100: 1
				case 123: 2
					fallthrough
				case 200: 3
			}`,
			"1", object.NUMBER_OBJ,
		},
		{`val x = 123
			switch x {
				case 100: 1
				case 123: 2
					fallthrough
				case 200: 3
			}`,
			"3", object.NUMBER_OBJ,
		},
		{`val x = 123
			switch x {
				case 100: 1
				case 123: 2
					fallthrough
				default: 3
			}`,
			"3", object.NUMBER_OBJ,
		},
		{`switch 123 {
		      case 123: fallthrough
			  case 100: var x = 7
			  default: 10
		  }`,
			"7", object.NUMBER_OBJ,
		},

		// fallthrough from anywhere in switch block (unique to langur, perhaps)
		{`val x = 123
			switch x {
				case 100: 1
				case 123: 2
					fallthrough
					777
				default: 3
			}`,
			"3", object.NUMBER_OBJ,
		},
		{`val x = 123
			switch x {
				case 100: 1
				case <= 123:
					if 0 < 1 { fallthrough } else { 777 }
				case 7: 3
			}`,
			"3", object.NUMBER_OBJ,
		},

		// implicit fallthrough (or "case alternates")
		{`val x = 123
			switch x {
					case 100:
					case 123: 2;
					case 200: 3 }`,
			"2", object.NUMBER_OBJ,
		},
		{`val x = 123
			switch x {
					case 123:
					case 150: 2
					case 200: 3 }`,
			"2", object.NUMBER_OBJ,
		},
		{`val x = 123
			switch x {
					case 100:
					case 101:
					case 102:
					case 123: 2
					case 103:
					case 200: 3 }`,
			"2", object.NUMBER_OBJ,
		},
		{`val x = 100
			switch x {
					case 100:
					case 101:
					case 102:
					case 123: 2
					case 103:
					case 200: 3 }`,
			"2", object.NUMBER_OBJ,
		},
		{`val x = 101
			switch x {
					case 100:
					case 101:
					case 102:
					case 123: 2
					case 103:
					case 200: 3 }`,
			"2", object.NUMBER_OBJ,
		},
		{`val x = 102
			switch x {
					case 100:
					case 101:
					case 102:
					case 123: 2
					case 103:
					case 200: 3 }`,
			"2", object.NUMBER_OBJ,
		},
		{`val x = 123
			switch x {
					case 100:
					case 101:
					case 102:
					case 123: 2
					case 103:
					case 200: 3 }`,
			"2", object.NUMBER_OBJ,
		},
		{`val x = 123
			switch x {
					case 100:
					case 101:
					case 102:
					case 103: 2
					case 123:
					case 200: 3 }`,
			"3", object.NUMBER_OBJ,
		},
		{`val x = 123
			switch x {
				case 100: 1
				case <= 123:
					if 0 < 1 { fallthrough } else { 777 }
				case 1:
				case 2:
				case 3:
				case 4:
				case 5: 30
				case 7: 40
			}`,
			"30", object.NUMBER_OBJ,
		},

		// last and least, non-variable switch (a semantic convenience, except for fallthrough)
		{`val x = 123
			switch {
				case x == 100: 1
				case x >= 123: 2
					fallthrough
				case false: 3
			}`,
			"3", object.NUMBER_OBJ,
		},
		{`val x = 123; val y = 159
			switch {
				case x == 100: 1
				case x >= y: 2
				case y >= x: 3
				default: 4
			}`,
			"3", object.NUMBER_OBJ,
		},
	}

	runVmTests(t, tests, false, false)
}

func TestSwitchWithIsAndIsNotOperators(t *testing.T) {
	tests := []vmTestCase{
		{`val x = 123
			switch x is {
				case string: 1
				case number: 2
				default: 3
			}`,
			2, object.NUMBER_OBJ,
		},
		{`val x = 123
			switch x {
				case is string: 1
				case is number: 2
				default: 3
			}`,
			2, object.NUMBER_OBJ,
		},

		{`val x = 123
			switch x is not {
				case string: 1
				case number: 2
				default: 3
			}`,
			1, object.NUMBER_OBJ,
		},
		{`val x = 123
			switch x {
				case is not string: 1
				case is not number: 2
				default: 3
			}`,
			1, object.NUMBER_OBJ,
		},

		{`val x = fn{+}
			switch x is {
				case fn: 1
				case number: 2
				default: 3
			}`,
			1, object.NUMBER_OBJ,
		},
		{`val x = fn{+}
			switch x is {
				case number: 1
				case fn: 2
				default: 3
			}`,
			2, object.NUMBER_OBJ,
		},
		{`val x = fn{+}
			switch x {
				case is fn: 1
				case is not number: 2
				default: 3
			}`,
			1, object.NUMBER_OBJ,
		},
		{`val x = fn{+}
			switch x {
				case is not fn: 1
				case is not number: 2
				default: 3
			}`,
			2, object.NUMBER_OBJ,
		},
		{`val x = fn{+}
			switch x {
				case is not fn: 1
				case is number: 2
				default: 3
			}`,
			3, object.NUMBER_OBJ,
		},
	}

	runVmTests(t, tests, false, false)
}

func TestSwitchWithAlternateExpressions(t *testing.T) {
	tests := []vmTestCase{
		// extra conditions (after semicolon within case condition list)
		{`val x = 123; val y = 159
			switch[and] x, y {
				case 123 ; 50 > 1 : 1
				case _, x: 2
				case 170: 3
				default: 4
			}`,
			"4", object.NUMBER_OBJ,
		},
		{`val x = 123; val y = 159
			switch x, y {
				case ; x > y: 1
				case ; x < y: 2		# match
				case _, 159: 3
				default: 4
			}`,
			"2", object.NUMBER_OBJ,
		},
		{`val x = 123; val y = 159
			switch x, y {
				case ; x < y: 1		# match
				case ; x > y: 2
				case _, 159: 3
				default: 4
			}`,
			"1", object.NUMBER_OBJ,
		},
		{`val x = 123; val y = 159
			switch x, y {
				case _, _; x < y: 1		# match
				case ; x > y: 2
				case _, 159: 3
				default: 4
			}`,
			"1", object.NUMBER_OBJ,
		},
		{`val x = 123; val y = 159
			switch[and] x, y {
				case 10, 20; x < y: 1
				case 123, 159; x == y: 4	# no match
				case 123, 159; x > y: 2	# no match
				case 123, 159; x < y: 3	# match
				case 123, 159; x >= y: 5
				case ; x > y: 6
				case _, 159: 7
			}`,
			"3", object.NUMBER_OBJ,
		},

		// different operator on other conditions
		{`switch[and] 200, 400 { case 200, 300; false: 1; default: 2 }`,
			"2", object.NUMBER_OBJ,
		},
		{`switch 200, 400 { case 200, 300; true: 1; default: 2 }`,
			"1", object.NUMBER_OBJ,
		},

		{`val x = 123; val y = 159; val z = false
			switch[and] x, y {
				case 123; x < y: 1
				case 123, 170; or z: 4
				case 100, 159; or x > y: 2
				case 159, 123; or x < y: 3	# match with or
				case 123, 159; x >= y: 5
				case ; x > y: 6
				case _, 159: 7
			}`,
			"3", object.NUMBER_OBJ,
		},
	}

	runVmTests(t, tests, false, false)
}

func TestShortenedFormSwitchExpressions(t *testing.T) {
	// not as flexibe as full form switch expressions
	tests := []vmTestCase{
		{`val x = 123
		  switch(x; > 100: 1)`, "1", object.NUMBER_OBJ},

		{`val x = 99
		  switch(x; < 100: 1; 100: 2; > 100: 3)`, "1", object.NUMBER_OBJ},
		{`val x = 100
		  switch(x; < 100: 1; 100: 2; > 100: 3)`, "2", object.NUMBER_OBJ},
		{`val x = 101
		  switch(x; < 100: 1; 100: 2; > 100: 3)`, "3", object.NUMBER_OBJ},

		{`val x = 101
		  switch(x; < 100: 1; 100: 2; > 100: 3; 4)`, "3", object.NUMBER_OBJ},
		{`val x = 101
		  switch(x; < 100: 1; 100: 2; 110: 3; 4)`, "4", object.NUMBER_OBJ},

		{`val x = 101
		  switch(x
			< 100: 1
			100: 2
			> 100: 3
			4)`,
			"3", object.NUMBER_OBJ},

		{`val x = 101; val y = 99
		  switch(x, y; < 99: 1; > 101: 2; 100: 3; 4)`, "4", object.NUMBER_OBJ},

		{`val x = 101; val y = 99
		  switch(x, y; < 100, > 100: 1; > 100, < 100: 2; 100: 3; 4)`, "2", object.NUMBER_OBJ},
	}

	runVmTests(t, tests, false, false)
}

func TestGlobalAssignments(t *testing.T) {
	tests := []vmTestCase{
		{"val one = 1; one;", "1", object.NUMBER_OBJ},
		{"val one = 1; val two = 2; one + two;", "3", object.NUMBER_OBJ},
		{"val one = 1; val two = one + one; one + two;", "3", object.NUMBER_OBJ},
	}

	runVmTests(t, tests, false, false)
}

func TestIndexListAssignment(t *testing.T) {
	tests := []vmTestCase{
		// set global indexed list value
		{`var l = [7, 8, 9]
		      { var x = 1
				l[2] = 14
			}
			l
		  `,
			[]int{7, 14, 9}, object.LIST_OBJ},

		// ... with negative index
		{`var l = [7, 8, 9]
		      { var x = 1
				l[-2] = 14
				l[-1] = 21
			}
			l
		  `,
			[]int{7, 14, 21}, object.LIST_OBJ},

		// set local indexed list value
		{`{ var x = 1
	       { 
				var l = [7, 8, 9]
				l[2] = 14
				l
			}
		  }`,
			[]int{7, 14, 9}, object.LIST_OBJ},

		// set non-local indexed list value
		{`{ var l = [7, 8, 9]
	       { var x = 1
				l[2] = 14
			}
			l
		  }`,
			[]int{7, 14, 9}, object.LIST_OBJ},

		// copied list references properly?
		{
			`val x = [[1, 2], [3, 4]]
			 var y = x
			 y[1] = [6, 7]
			 x
			`,
			[][]int{{1, 2}, {3, 4}},
			object.LIST_OBJ,
		},

		// FIXME: set a value with multiple indices
		// {`var x = [[2, 4], [7, 8]]
		//   x[2][1] = 14
		//   x[2][1] + 7`,
		// 	21, object.NUMBER_OBJ,
		// },
	}

	runVmTests(t, tests, false, false)
}

func TestIndexHashAssignment(t *testing.T) {
	tests := []vmTestCase{
		// set global indexed hash value for existing key
		{`var h = {1: 7, 2: 0, -1: 21, 1.1: 42, -1.1: 35}
		      { var x = 1
				h[2] = 14
			  }
			[h[1], h[2], h[-1], h[-1.1], h[1.1]]
		  `,
			[]int{7, 14, 21, 35, 42}, object.LIST_OBJ},

		{`var h = {1: 7, 2: 0, -1: 21, 1.10: 42, -1.1: 35}
		      { var x = 1
				h[-1.1] = 14
			  }
			[h[1], h[2], h[-1], h[-1.1], h[1.1]]
		  `,
			[]int{7, 0, 21, 14, 42}, object.LIST_OBJ},

		// set global indexed hash value for non-existing key
		{`var h = {1: 7}
		      { var x = 1
				h[2] = 14
			  }
			[h[1], h[2]]
		  `,
			[]int{7, 14}, object.LIST_OBJ},

		{`var h = {1: 7}
		      { var x = 1
				h[-2.4] = 14
			  }
			[h[1], h[-2.4]]
		  `,
			[]int{7, 14}, object.LIST_OBJ},

		// set local indexed hash value for existing key
		{`var x = 1
		      { var h = {1: 7, 2: 0}
				h[2] = 14
				[h[1], h[2]]
			  }
		  `,
			[]int{7, 14}, object.LIST_OBJ},

		// set local indexed hash value for non-existing key
		{`var x = 1
		      { var h = {1: 7}
				h[2] = 14
				[h[1], h[2]]
			  }
		  `,
			[]int{7, 14}, object.LIST_OBJ},

		// set non-local indexed hash value for existing key
		{`{ var h = {1: 7, 2: 0}
		      { var x = 1
				h[2] = 14
			  }
			[h[1], h[2]]
		  }`,
			[]int{7, 14}, object.LIST_OBJ},

		// set non-local indexed hash value for non-existing key
		{`{ var h = {1: 7}
		      { var x = 1
				h[2] = 14
			  }
			[h[1], h[2]]
		  }`,
			[]int{7, 14}, object.LIST_OBJ},

		// copied hash references properly?
		{`{ val h1 = {1: 7, 2: 0}
			var h2 = h1
			h2[1] = 14
			h1[1]
		  }`,
			"7", object.NUMBER_OBJ,
		},
	}

	runVmTests(t, tests, false, false)
}

func TestIndexStringAssignment(t *testing.T) {
	tests := []vmTestCase{
		{
			`var x = "abcd"
			 x[1] = "Z"
			 x`,
			"Zbcd", object.STRING_OBJ,
		},

		// ... with negative index
		{
			`var x = "abcd"
			 x[-1] = "Z"
			 x`,
			"abcZ", object.STRING_OBJ,
		},

		{
			`var x = "abcd"
			 x[3] = "\u05d0"
			 x`,
			"ab\u05d0d", object.STRING_OBJ,
		},
		{
			`var x = "a\u05d0cd"
			 x[2] = "Z"
			 x`,
			"aZcd", object.STRING_OBJ,
		},

		// clobbers one code point only
		{
			`var x = "abcd"
			 x[2] = "zzz"
			 x`,
			"azcd", object.STRING_OBJ,
		},

		// set from a code point number
		{
			`var x = "abcd"
			 x[4] = 67
			 x`,
			"abcC", object.STRING_OBJ,
		},
	}

	runVmTests(t, tests, false, false)
}

func TestMultiVariableDeclarationAssignment(t *testing.T) {
	tests := []vmTestCase{
		{
			`var x, y = 7, 14
			 x`,
			"7", object.NUMBER_OBJ,
		},
		{
			`val x, y = 7, 14
			 y`,
			"14", object.NUMBER_OBJ,
		},
		{
			`val x, y = 7, 14
			 x + y`,
			"21", object.NUMBER_OBJ,
		},
		{
			`var x, y
			 x`,
			nil, object.NULL_OBJ,
		},
	}

	runVmTests(t, tests, false, false)
}

func TestMultiVariableAssignment(t *testing.T) {
	// separate from declaration
	tests := []vmTestCase{
		{
			`var x; var y
			 x, y = 7, 14
			 x`,
			"7", object.NUMBER_OBJ,
		},
		{ // with parentheses (backwards compatible)
			`var x; var y
			 x, y = 7, 14
			 x`,
			"7", object.NUMBER_OBJ,
		},

		// swap values
		{
			`var x = 7; var y = 14
			 x, y = y, x
			 [x, y]`,
			[]int{14, 7}, object.LIST_OBJ,
		},
	}

	runVmTests(t, tests, false, false)
}

func TestDecouplingDeclarationAssignment(t *testing.T) {
	tests := []vmTestCase{
		{
			`var x, y = [7, 14]
			 x + y`,
			"21", object.NUMBER_OBJ,
		},
		{
			`val test = [7, 14]
			 var x, y = test
			 x`,
			"7", object.NUMBER_OBJ,
		},
		{
			`val x, y = [123, 456]`,
			true, object.BOOLEAN_OBJ,
		},

		// not enough things...
		{
			`var x, y = [123]`,
			false, object.BOOLEAN_OBJ,
		},
		{
			`val x, y = [123]`,
			false, object.BOOLEAN_OBJ,
		},
		{
			`val x, y = [123]
			 x`,
			nil, object.NULL_OBJ,
		},

		// with no-op
		{
			`var x, _, y = [7, 14, 21]
			 x + y`,
			"28", object.NUMBER_OBJ,
		},
		{
			`var _, x, y = [7, 14, 21]
			 x + y`,
			"35", object.NUMBER_OBJ,
		},
	}

	runVmTests(t, tests, false, false)
}

func TestDecouplingAssignment(t *testing.T) {
	// separate from declaration
	tests := []vmTestCase{
		{
			`var x; var y
			 x, y = [7, 14]
			 x`,
			"7", object.NUMBER_OBJ,
		},

		// swap values with decoupling
		{
			`var x = 7; var y = 14
			 x, y = [y, x]
			 x`,
			"14", object.NUMBER_OBJ,
		},

		// with no-op
		{
			`var x; var y
			x, _, y = [7, 14, 21]
			 x + y`,
			"28", object.NUMBER_OBJ,
		},

		// not enough things...
		{
			`var x, y
			x, y = [123]`,
			false, object.BOOLEAN_OBJ,
		},
		{
			`var x, y
			x, y = [123]
			 x`,
			nil, object.NULL_OBJ,
		},
	}

	runVmTests(t, tests, false, false)
}

func TestDecouplingWithExpansion(t *testing.T) {
	tests := []vmTestCase{
		// using 2 variables
		{
			`var x, y
			 x, y... = [7, 14, 21]
			 x`,
			7, object.NUMBER_OBJ,
		},
		{
			`var x, y
			 x, y... = [7, 14, 21]
			 y[1] + y[2]`,
			35, object.NUMBER_OBJ,
		},
		{
			`var x, y
			 x, y ... = [7, 14]
			 y[1]`,
			14, object.NUMBER_OBJ,
		},

		{
			`var x, y
			 x, y ... = [7, 14, 21]
			 len y`,
			2, object.NUMBER_OBJ,
		},
		{
			`var x, y
			 x, y ... = [7, 14]
			 len y`,
			1, object.NUMBER_OBJ,
		},
		{
			// 0 minimum by default
			`var x, y
			 x, y ... = [7]
			 len y`,
			0, object.NUMBER_OBJ,
		},
		{
			`var x, y
			 x, y ... = []
			 y`,
			nil, object.NULL_OBJ,
		},

		// using 2 variables with index on expansion variable
		{
			`var x
			 var y = [1, 2, 3]
			 x, y[2] ... = [7, 14, 21]
			 x`,
			7, object.NUMBER_OBJ,
		},
		{
			`var x
			 var y = [1, 2, 3]
			 x, y[2] ... = [7, 14, 21]
			 y[1] + y[3]`,
			4, object.NUMBER_OBJ,
		},
		{
			`var x
			 var y = [1, 2, 3]
			 x, y[2] ... = [7, 14, 21]
			 y[2][1] + y[2][2]`,
			35, object.NUMBER_OBJ,
		},

		// 2 variables in if test
		{
			`var x, y
			 if x, y ... = [7, 14, 21] { x } else { 0 }`,
			7, object.NUMBER_OBJ,
		},
		{
			`var x, y
			 if x, y ... = [7] { x } else { 0 }`,
			7, object.NUMBER_OBJ,
		},
		{
			`var x, y
			 if x, y ... = [] { x } else { 0 }`,
			0, object.NUMBER_OBJ,
		},
		{
			`var x, y
			 if x, y ... = [7, 14, 21] { y[1] + y[2] } else { 0 }`,
			35, object.NUMBER_OBJ,
		},
		{
			`var x, y
			 if x, y ... = [7, 14] { y[1] } else { 0 }`,
			14, object.NUMBER_OBJ,
		},
		{
			`var x, y
			 if x, y ... = [7, 14, 21] { len y } else { 0 }`,
			2, object.NUMBER_OBJ,
		},
		{
			`var x, y
			 if x, y ... = [7, 14] { len y } else { 0 }`,
			1, object.NUMBER_OBJ,
		},
		{
			`var x, y
			 if x, y ... = [7] { len y } else { -1 }`,
			0, object.NUMBER_OBJ,
		},
		{
			`var x, y
			 if x, y ... = [] { len y } else { -1 }`,
			-1, object.NUMBER_OBJ,
		},

		// using 3 variables
		{
			`var x, y, z
			 x, y, z ... = [7, 14, 21, 28, 35, 42]
			 x`,
			7, object.NUMBER_OBJ,
		},
		{
			`var x, y, z
			 x, y, z ... = [7, 14, 21, 28, 35, 42]
			 y`,
			14, object.NUMBER_OBJ,
		},
		{
			`var x, y, z
			 x, y, z ... = [7, 14, 21, 28, 35, 42]
			 z[1] + z[2] + z[3] + z[4]`,
			126, object.NUMBER_OBJ,
		},

		{
			`var x, y, z
			 x, y, z ... = [7, 14, 21, 28, 35, 42]
			 len z`,
			4, object.NUMBER_OBJ,
		},
		{
			`var x, y, z
			 x, y, z ... = [7, 14, 21, 28, 35]
			 len z`,
			3, object.NUMBER_OBJ,
		},
		{
			`var x, y, z
			 x, y, z ... = [7, 14, 21, 28]
			 len z`,
			2, object.NUMBER_OBJ,
		},
		{
			`var x, y, z
			 x, y, z ... = [7, 14, 21]
			 len z`,
			1, object.NUMBER_OBJ,
		},
		{
			`var x, y, z
			 x, y, z ... = [7, 14]
			 len z`,
			0, object.NUMBER_OBJ,
		},
		{
			`var x, y, z
			 x, y, z ... = [7]
			 z`,
			nil, object.NULL_OBJ,
		},

		// 3 variables in if test
		{
			`var x, y, z
			 if x, y, z ... = [7, 14, 21] { x } else { 0 }`,
			7, object.NUMBER_OBJ,
		},
		{
			`var x, y, z
			 if x, y, z ... = [7, 14, 21] { y } else { 0 }`,
			14, object.NUMBER_OBJ,
		},
		{
			`var x, y, z
			 if x, y, z ... = [] { y } else { -1 }`,
			-1, object.NUMBER_OBJ,
		},

		{
			`var x, y, z
			 if x, y, z ... = [7, 14, 21, 28] { len z } else { 0 }`,
			2, object.NUMBER_OBJ,
		},
		{
			`var x, y, z
			 if x, y, z ... = [7, 14, 21] { len z } else { 0 }`,
			1, object.NUMBER_OBJ,
		},
		{
			`var x, y, z
			 if x, y, z ... = [7, 14] { len z } else { 0 }`,
			0, object.NUMBER_OBJ,
		},
		{
			`var x, y, z
			 if x, y, z ... = [7] { x } else { -1 }`,
			-1, object.NUMBER_OBJ,
		},
	}

	runVmTests(t, tests, false, false)
}

func TestDecouplingDeclarationWithExpansion(t *testing.T) {
	tests := []vmTestCase{
		// using 2 variables
		{
			`val x, y ... = [7, 14, 21]
			 x`,
			7, object.NUMBER_OBJ,
		},
		{
			`val x, y ... = [7, 14, 21]
			 y[1] + y[2]`,
			35, object.NUMBER_OBJ,
		},
		{
			`val x, y ... = [7, 14]
			 y[1]`,
			14, object.NUMBER_OBJ,
		},

		{
			`val x, y ... = [7, 14, 21]
			 len y`,
			2, object.NUMBER_OBJ,
		},
		{
			`val x, y ... = [7, 14]
			 len y`,
			1, object.NUMBER_OBJ,
		},
		{
			// 0 minimum by default
			`val x, y ... = [7]
			 len y`,
			0, object.NUMBER_OBJ,
		},
		{
			`val x, y ... = []
			 y`,
			nil, object.NULL_OBJ,
		},

		// 2 variables in if test
		{
			`if val x, y ... = [7, 14, 21] { x } else { 0 }`,
			7, object.NUMBER_OBJ,
		},
		{
			`if val x, y ... = [7] { x } else { 0 }`,
			7, object.NUMBER_OBJ,
		},
		{
			`if val x, y ... = [] { x } else { 0 }`,
			0, object.NUMBER_OBJ,
		},
		{
			`if val x, y ... = [7, 14, 21] { y[1] + y[2] } else { 0 }`,
			35, object.NUMBER_OBJ,
		},
		{
			`if val x, y ... = [7, 14] { y[1] } else { 0 }`,
			14, object.NUMBER_OBJ,
		},
		{
			`if val x, y ... = [7, 14, 21] { len y } else { 0 }`,
			2, object.NUMBER_OBJ,
		},
		{
			`if val x, y ... = [7, 14] { len y } else { 0 }`,
			1, object.NUMBER_OBJ,
		},
		{
			`if val x, y ... = [7] { len y } else { -1 }`,
			0, object.NUMBER_OBJ,
		},
		{
			`if val x, y ... = [] { len y } else { -1 }`,
			-1, object.NUMBER_OBJ,
		},

		// using 3 variables
		{
			`var x, y, z ... = [7, 14, 21, 28, 35, 42]
			 x`,
			7, object.NUMBER_OBJ,
		}, // using 3 variables
		{
			`var x, y, z ... = [7, 14, 21, 28, 35, 42]
			 x`,
			7, object.NUMBER_OBJ,
		},
		{
			`val x, y, z ... = [7, 14, 21, 28, 35, 42]
			 y`,
			14, object.NUMBER_OBJ,
		},
		{
			`val x, y, z ... = [7, 14, 21, 28, 35, 42]
			 z[1] + z[2] + z[3] + z[4]`,
			126, object.NUMBER_OBJ,
		},

		{
			`var x, y, z ... = [7, 14, 21, 28, 35, 42]
			 len z`,
			4, object.NUMBER_OBJ,
		},
		{
			`var x, y, z ... = [7, 14, 21, 28, 35]
			 len z`,
			3, object.NUMBER_OBJ,
		},
		{
			`var x, y, z ... = [7, 14, 21, 28]
			 len z`,
			2, object.NUMBER_OBJ,
		},
		{
			`var x, y, z ... = [7, 14, 21]
			 len z`,
			1, object.NUMBER_OBJ,
		},
		{
			`var x, y, z ... = [7, 14]
			 len z`,
			0, object.NUMBER_OBJ,
		},
		{
			`var x, y, z ... = [7]
			 z`,
			nil, object.NULL_OBJ,
		},
		{
			`var x, y, z ... = [7, 14, 21, 28, 35, 42]
			 y`,
			14, object.NUMBER_OBJ,
		},
		{
			`var x, y, z ... = [7, 14, 21, 28, 35, 42]
			 z[1] + z[2] + z[3] + z[4]`,
			126, object.NUMBER_OBJ,
		},

		{
			`var x, y, z ... = [7, 14, 21, 28, 35, 42]
			 len z`,
			4, object.NUMBER_OBJ,
		},
		{
			`var x, y, z ... = [7, 14, 21, 28, 35]
			 len z`,
			3, object.NUMBER_OBJ,
		},
		{
			`var x, y, z ... = [7, 14, 21, 28]
			 len z`,
			2, object.NUMBER_OBJ,
		},
		{
			`var x, y, z ... = [7, 14, 21]
			 len z`,
			1, object.NUMBER_OBJ,
		},
		{
			`var x, y, z ... = [7, 14]
			 len z`,
			0, object.NUMBER_OBJ,
		},
		{
			`var x, y, z ... = [7]
			 z`,
			nil, object.NULL_OBJ,
		},

		// 3 variables in if test
		{
			`if val x, y, z ... = [7, 14, 21] { x } else { 0 }`,
			7, object.NUMBER_OBJ,
		},
		{
			`if val x, y, z ... = [7, 14, 21] { y } else { 0 }`,
			14, object.NUMBER_OBJ,
		},
		{
			`if val x, y, z ... = [] { y } else { -1 }`,
			-1, object.NUMBER_OBJ,
		},

		{
			`if val x, y, z ... = [7, 14, 21, 28] { len z } else { 0 }`,
			2, object.NUMBER_OBJ,
		},
		{
			`if val x, y, z ... = [7, 14, 21] { len z } else { 0 }`,
			1, object.NUMBER_OBJ,
		},
		{
			`if val x, y, z ... = [7, 14] { len z } else { 0 }`,
			0, object.NUMBER_OBJ,
		},
		{
			`if val x, y, z ... = [7] { x } else { -1 }`,
			-1, object.NUMBER_OBJ,
		},
	}

	runVmTests(t, tests, false, false)
}

func TestDecouplingAssignmentFromRegexInIfElse(t *testing.T) {
	tests := []vmTestCase{
		{
			`if val x, y = submatch("jklwer489werjk27,.dsjfkl56", by=RE/(\d+).+?(\d+)/) {
				[y, x]
			 }`,
			[]string{"27", "489"},
			object.LIST_OBJ,
		},

		{
			`val line = "jklwer489werjk27,.dsjfkl56"
			 if val x = submatch(line, by=RE/(\d+).+?(\d+)/) {
				x
			 }`,
			[]string{"489", "27"},
			object.LIST_OBJ,
		},

		{
			`val line = "jklwer489werjk27,.dsjfkl56"
			 if val _, x = submatch(line, by=RE/(\d+).+?(\d+)/) {
				x
			 }`,
			"27",
			object.STRING_OBJ,
		},

		// no match
		{
			`if val x, y = submatch("jklwer489werjk27,.dsjfkl56", by=RE/(asdf\d+).+?(vwewr\d+)/) {
				[y, x]
			 }`,
			nil,
			object.NULL_OBJ,
		},

		// no match on first test
		{
			`if val x, y = submatch("jklwer489werjk27,.dsjfkl56", by=RE/(asdf\d+).+?(vwewr\d+)/) {
				[y, x]
			 } else if var x = null {	# with scope on second test as well
				123; x
			 }`,
			nil,
			object.NULL_OBJ,
		},
		{
			`if val x, y = submatch("jklwer489werjk27,.dsjfkl56", by=RE/(asdf\d+).+?(vwewr\d+)/) {
				[y, x]
			 } else if false == true xor false {	# without scope on second test
				123; 456; 789
			 }`,
			nil,
			object.NULL_OBJ,
		},
		{
			`if val x, y = submatch("jklwer489werjk27,.dsjfkl56", by=RE/(asdf\d+).+?(vwewr\d+)/) {
				[y, x]
			 } else {			# with explicit else after scoped test and action
				123; 456; 789
			 }`,
			"789",
			object.NUMBER_OBJ,
		},
	}

	runVmTests(t, tests, false, false)
}

func TestAssignmentsInIfElse(t *testing.T) {
	tests := []vmTestCase{
		// Langur is not meant to be a purely functional language.
		// Obviously, these simple cases could be rewritten in a functional style, ...
		// ... but a lot of programming is not so simple. Maybe you want to set multiple variables ...
		// ... based on a switch condition.
		{"var x = 0; if false { x = 1 } else { x = 2 }; x", "2", object.NUMBER_OBJ},
		{"var x; if false { x = 1 } else { x = 2 }; x", "2", object.NUMBER_OBJ},
		{"var x; if false { x = 1 } else { x = 2 }", "2", object.NUMBER_OBJ},
		{"var x; if true { x = 1 } else { x = 2 }", "1", object.NUMBER_OBJ},
	}

	runVmTests(t, tests, false, false)
}

func TestStringExpressions(t *testing.T) {
	tests := []vmTestCase{
		{`"langur"`, "langur", object.STRING_OBJ},
		{"qs[langur]", "langur", object.STRING_OBJ},
		{"QS[langur]", "langur", object.STRING_OBJ},
		{`zls`, "", object.STRING_OBJ},
	}

	runVmTests(t, tests, false, false)
}

func TestStringBlockQuotes(t *testing.T) {
	tests := []vmTestCase{
		// blockquote
		{"QS:block ABC\nlangur\nABC", "langur", object.STRING_OBJ},
		{"QS:block ABC\nsomething\ngreat\nABC", "something\ngreat", object.STRING_OBJ},
		{"QS:block END\nnot the END\nEND", "not the END", object.STRING_OBJ},

		// using the lead modifier, which trims leading spaces on each line
		{`QS:lead:block END
			not the END
			asdf
			END`, "not the END\nasdf", object.STRING_OBJ},

		{`QS:lead:block END
			not the END

			asdf
			END`, "not the END\n\nasdf", object.STRING_OBJ},

		{`QS:lead:block END
			not the END

			asdf

			only the beginning
			END`, "not the END\n\nasdf\n\nonly the beginning", object.STRING_OBJ},

		// blockquote with interpolation
		{"QS:block ABC\n{{0.1+0}}\nABC", "0.1", object.STRING_OBJ},
		{"QS:block ABC\n{{0.1+0}} something\nABC", "0.1 something", object.STRING_OBJ},
		{"QS:block ABC\nlangur {{1.0+0}}\nABC", "langur 1.0", object.STRING_OBJ},
		{"QS:block ABC\nlangur {{1.0+0}}, you know?\nABC", "langur 1.0, you know?", object.STRING_OBJ},

		{`QS:lead:block END
			not the END

			{{7 * 7}}

			only the beginning
			END`, "not the END\n\n49\n\nonly the beginning", object.STRING_OBJ},

		// blockquote with other things following
		{`QS:block ABC
		     langur
		ABC
		  "okay"`, "okay", object.STRING_OBJ},

		// blockquote used in list
		{`val lead = [QS:lead:block ABC
		     42
		ABC,
			"7"]

		  lead[1] ~ lead[2]`, "427", object.STRING_OBJ},
	}

	runVmTests(t, tests, false, false)
}

func TestStringAndRegexModifiers(t *testing.T) {
	tests := []vmTestCase{
		{"QS:any(\uF8FF)", "\uF8FF", object.STRING_OBJ},
		{"matching(QS:any(\uF8FF), by=re:any(\uF8FF))", true, object.BOOLEAN_OBJ},
	}

	runVmTests(t, tests, false, false)
}

func TestStringAppend(t *testing.T) {
	tests := []vmTestCase{
		{"qs[lan] ~ qs(gurs)", "langurs", object.STRING_OBJ},
		{`qs[lan] ~ qs(gurs) ~ " ♥ bananas"`, "langurs ♥ bananas", object.STRING_OBJ},

		{`"A" ~ "B"`, "AB", object.STRING_OBJ},
		{`"A" ~ 99`, "Ac", object.STRING_OBJ},
		{`99 ~ "A"`, "cA", object.STRING_OBJ},

		{`99 ~ 65`, "cA", object.STRING_OBJ},
		{`97 ~ 98 ~ 99`, "abc", object.STRING_OBJ},

		{`"A" ~ 97 .. 101`, "Aabcde", object.STRING_OBJ},
		{`"A" ~ 101 .. 97`, "Aedcba", object.STRING_OBJ},

		// in the Latin extended section ...
		{`"" ~ 16xA1`, "¡", object.STRING_OBJ},
	}

	runVmTests(t, tests, false, false)
}

func TestBooleanMultiplication(t *testing.T) {
	tests := []vmTestCase{
		{`(1 > 15) * "A"`, "", object.STRING_OBJ},
		{`(1 < 15) * "A"`, "A", object.STRING_OBJ},

		// string
		{`false * "A"`, "", object.STRING_OBJ},
		{`true * "A"`, "A", object.STRING_OBJ},
		{`"abc" * false`, "", object.STRING_OBJ},
		{`"abc" * true`, "abc", object.STRING_OBJ},

		// number
		{`false * 42`, 0, object.NUMBER_OBJ},
		{`true * 42`, 42, object.NUMBER_OBJ},
		{`42 * false`, 0, object.NUMBER_OBJ},
		{`42 * true`, 42, object.NUMBER_OBJ},

		// duration
		{`false * dr/2d/`, "PT0S", object.DURATION_OBJ},
		{`true * dr/2d/`, "P2D", object.DURATION_OBJ},
		{`dr/2d/ * false`, "PT0S", object.DURATION_OBJ},
		{`dr/2d/ * true`, "P2D", object.DURATION_OBJ},

		// list
		{`false * [1, 2]`, []int{}, object.LIST_OBJ},
		{`true * [1, 2]`, []int{1, 2}, object.LIST_OBJ},
		{`[1, 2] * false`, []int{}, object.LIST_OBJ},
		{`[1, 2] * true`, []int{1, 2}, object.LIST_OBJ},

		// hash
		{
			`false * {1: 2, 2: 7}`, [][]object.Object{}, object.HASH_OBJ,
		},
		{
			`true * {1: 2, 2: 7}`,
			[][]object.Object{
				{object.NumberFromInt(1), object.NumberFromInt(2)},
				{object.NumberFromInt(2), object.NumberFromInt(7)},
			},
			object.HASH_OBJ,
		},
		{
			`{1: 2, 2: 7} * false`, [][]object.Object{}, object.HASH_OBJ,
		},
		{
			`{1: 2, 2: 7} * true`,
			[][]object.Object{
				{object.NumberFromInt(1), object.NumberFromInt(2)},
				{object.NumberFromInt(2), object.NumberFromInt(7)},
			},
			object.HASH_OBJ,
		},
	}

	runVmTests(t, tests, false, false)
}

func TestStringMultiplication(t *testing.T) {
	tests := []vmTestCase{
		{`"A" * 3`, "AAA", object.STRING_OBJ},
		{`"A" * 2`, "AA", object.STRING_OBJ},
		{`"A" * 1`, "A", object.STRING_OBJ},
		{`"A" * 0`, "", object.STRING_OBJ},

		{`"A" * -1`, "", object.STRING_OBJ},

		{`3 * "A"`, "AAA", object.STRING_OBJ},
		{`2 * "A"`, "AA", object.STRING_OBJ},
		{`1 * "A"`, "A", object.STRING_OBJ},
		{`0 * "A"`, "", object.STRING_OBJ},

		{`-1 * "A"`, "", object.STRING_OBJ},

		{`3 * "abc"`, "abcabcabc", object.STRING_OBJ},
	}

	runVmTests(t, tests, false, false)
}

func TestStringComparisons(t *testing.T) {
	tests := []vmTestCase{
		{`"resume\u0301" == "resume\u0301"`, true, object.BOOLEAN_OBJ},
		{`"resume\u0301" == "resum\u00E9"`, false, object.BOOLEAN_OBJ},
	}

	runVmTests(t, tests, false, false)
}

func TestStringNormalizations(t *testing.T) {
	tests := []vmTestCase{
		// TODO: more tests
		{`nfc("resume\u0301")`, "resum\u00E9", object.STRING_OBJ},
		{`nfd("resum\u00E9")`, "resume\u0301", object.STRING_OBJ},
		{`nfkc("\u212A")`, "\u004B", object.STRING_OBJ},
		{`nfkd("\u004B")`, "\u004B", object.STRING_OBJ},

		{`nfc("resume\u0301") == "resum\u00E9"`, true, object.BOOLEAN_OBJ},
		{`nfc("resume\u0301") == "resume\u0301"`, false, object.BOOLEAN_OBJ},

		{`nfkc("abc\u212Adef")`, "abc\u004Bdef", object.STRING_OBJ},
	}

	runVmTests(t, tests, false, false)
}

func TestStringInterpolations(t *testing.T) {
	tests := []vmTestCase{
		// non-interpolated
		{`val x = 123; qs:ni"{{x}}"`, "{{x}}", object.STRING_OBJ},

		{`val x = 123; "{{x}}"`, "123", object.STRING_OBJ},
		{`val x = 123; "{{x}}=123"`, "123=123", object.STRING_OBJ},

		{`val x = 123 + 12i; "{{x}}"`, "123+12i", object.STRING_OBJ},
		{`"{{123 + 12i}}"`, "123+12i", object.STRING_OBJ},

		{`val x = 123; "abc {{x}} yo"`, "abc 123 yo", object.STRING_OBJ},
		{`val x = 123; val y = 17; "abc {{x + y}} yo"`, "abc 140 yo", object.STRING_OBJ},

		{`val x = [1, 2, 3]; "abc {{len(x)}} yo"`, "abc 3 yo", object.STRING_OBJ},
		{`val x = [1, 2, 3]; "abc {{x}} yo"`, "abc [1, 2, 3] yo", object.STRING_OBJ},

		// test nesting marks within interpolation...
		{`val x = 123; val y = 17; "abc {{ (x + y) }} yo"`, "abc 140 yo", object.STRING_OBJ},
		{`val x = [7, 14, 21]; "abc {{x[2]}} yo"`, "abc 14 yo", object.STRING_OBJ},

		// nesting marks with interpolated string enclosing marks
		// The following looks confusing, but shows that the lexer is ...
		// ... reading tokens directly in the interpolated section, since ...
		// ... the }} inside the string inside the interpolation is ...
		// ... not taken to be a closing mark of any kind.
		{`"{{"}}"~zls}}"`, "}}", object.STRING_OBJ},
		{`"{{["}}"~zls]}}"`, "[\"}}\"]", object.STRING_OBJ},

		// with double quote, single quote, or forward slash within interpolation...
		{`val x = 123; "{{"1"~"2"}}"`, "12", object.STRING_OBJ},
		{`val x = 123; "{{' '+8}}"`, "40", object.STRING_OBJ},
		{`val x = 123; "{{1/2}}"`, "0.5", object.STRING_OBJ},

		// test newlines in allowed interpolations...
		{`val x = [7, 14, 21]
		  qs"meaning of life: {{
			fold(x, by=fn{+})
		   }}"
		`, "meaning of life: 42", object.STRING_OBJ},
	}

	runVmTests(t, tests, false, false)
}

func TestStringInterpolationModifiers(t *testing.T) {
	tests := []vmTestCase{
		// langur escapes from interpolated values
		{`val x = "\u000A"; "You know {{x:esc}}"`, "You know \\n", object.STRING_OBJ},
		{`val x = '\u000A'; "You know {{x:esc}}"`, "You know 10", object.STRING_OBJ},
		{`val x = "\u2028 !"; "I know {{x:esc}} noting"`, "I know \\u2028 ! noting", object.STRING_OBJ},
		{`val x = "\x0e\x7f"; "{{x:esc}}"`, "\\x0e\\x7f", object.STRING_OBJ},
		{`val x = "Z\x0e\u008f\u0090\u0091"; "{{x:esc}}"`, "Z\\x0e\\u008f\\u0090\\u0091", object.STRING_OBJ},

		{`val x = "\u000A"; "You know {{(x~""):esc}}"`, "You know \\n", object.STRING_OBJ},
		{`val x = '\u000A'; "You know {{(x~""):esc}}"`, "You know \\n", object.STRING_OBJ},
		{`val x = '\u000A'; "You know {{(x+0):esc}}"`, "You know 10", object.STRING_OBJ},
		{`val x = "\u2028 !"; "I know {{(x~"") : esc}} noting"`, "I know \\u2028 ! noting", object.STRING_OBJ},

		// code points
		{`val x = 97; "{{x:cp}}"`, "a", object.STRING_OBJ},
		{`val x = 16xA; "{{x:cp}}"`, "\n", object.STRING_OBJ},
		{`val x = 16x03B4; "{{x:cp}}"`, "δ", object.STRING_OBJ},
		{`val x = 16x1F355; "{{x:cp}}"`, "🍕", object.STRING_OBJ},

		{`val x = 97..100; "{{x:cp}}"`, "abcd", object.STRING_OBJ},
		{`val x = 63..58; "{{x:cp}}"`, "?>=<;:", object.STRING_OBJ},
		{`val x = [97..100, 63..58]; "{{x:cp}}"`, "abcd?>=<;:", object.STRING_OBJ},
		{`val x = [97..100, [63, 58..59]]; "{{x:cp}}"`, "abcd?:;", object.STRING_OBJ},

		// combinations
		{`val x = 1234567890; "{{x : L7 : -10(.)}}"`, "1234567...", object.STRING_OBJ},

		{`val x = 255; "{{x :x :2}}"`, "ff", object.STRING_OBJ},
		{`val x = 255; "{{x:X:5}}"`, "   FF", object.STRING_OBJ},
		{`val x = -255; "{{x:X:5}}"`, "  -FF", object.STRING_OBJ},
		{`val x = 14; "{{x:x02:2}}"`, "0e", object.STRING_OBJ},
		{`val x = 14; "{{x:X02:5}}"`, "   0E", object.STRING_OBJ},
		{`val x = 14; "{{x:X02:-5}}"`, "0E   ", object.STRING_OBJ},
		{`val x = 14; "{{x:x2:2}}"`, " e", object.STRING_OBJ},
		{`val x = 14; "{{x:X2:5}}"`, "    E", object.STRING_OBJ},
		{`val x = 14; "{{x:X2:-5}}"`, " E   ", object.STRING_OBJ},

		// custom formatting function
		{`val x = 255; val F = fn(s) { ucase s }; "{{x:x:fn F}}"`, "FF", object.STRING_OBJ},
		{`val x = "   sdf sdf  "; val T = fn(s) { trim s }; "{{x : fn T }}"`, "sdf sdf", object.STRING_OBJ},

		// type string
		{`val x = 255; "{{x:T}}"`, common.NumberTypeName, object.STRING_OBJ},
		{`val x = 255+1i; "{{x:T}}"`, common.ComplexTypeName, object.STRING_OBJ},
		{`val x = 1 .. 255; "{{x : T }}"`, common.RangeTypeName, object.STRING_OBJ},
		{`val x = fn{+}; "{{x:T}}"`, common.FuntionTypeName, object.STRING_OBJ},
		{`val x = len; "{{x:T}}"`, common.BuiltInTypeName, object.STRING_OBJ},
		{`val x = "255"; "{{x:T}}"`, common.StringTypeName, object.STRING_OBJ},
		{`val x = 255; "{{x:7:T}}"`, common.StringTypeName, object.STRING_OBJ},
		{`val x = re/255/; "{{x:T}}"`, common.RegexTypeName, object.STRING_OBJ},
		{`val x = null; "{{x:T}}"`, common.NullTypeName, object.STRING_OBJ},
		{`val x = true; "{{x:T}}"`, common.BooleanTypeName, object.STRING_OBJ},
		{`val x = []; "{{x:T}}"`, common.ListTypeName, object.STRING_OBJ},
		{`val x = {:}; "{{x:T}}"`, common.HashTypeName, object.STRING_OBJ},
		{`val x = dt//; "{{x:T}}"`, common.DateTimeTypeName, object.STRING_OBJ},
		{`val x = dr/10Y/; "{{x:T}}"`, common.DurationTypeName, object.STRING_OBJ},
	}

	runVmTests(t, tests, false, false)
}

func TestInterpolationModifierForCpAlignment(t *testing.T) {
	tests := []vmTestCase{
		// alignment with default ASCII space
		{`val x = 255; "{{x:3}}"`, "255", object.STRING_OBJ},
		{`val x = 255; "{{x:1}}"`, "255", object.STRING_OBJ},

		{`val x = 255; "{{x:5}}"`, "  255", object.STRING_OBJ},
		{`val x = 255; "{{x:-5}}"`, "255  ", object.STRING_OBJ},

		// alignment with alternate padding code point
		{`val x = 255; "{{x: 10(.)}}"`, ".......255", object.STRING_OBJ},
		{`val x = 255; "{{x:-10(.)}}"`, "255.......", object.STRING_OBJ},

		// double alignment
		{`val x = 255; "{{x:-10(.):17(*)}}"`, "*******255.......", object.STRING_OBJ},
		{`val x = 255; "{{x:10(.):17(*)}}"`, "*******.......255", object.STRING_OBJ},

		// alternate padding code point by hexadecimal number
		{`val x = 255; "{{x:10(2A)}}"`, "*******255", object.STRING_OBJ},
		{`val x = 255; "{{x:-10(2A)}}"`, "255*******", object.STRING_OBJ},
	}

	runVmTests(t, tests, false, false)
}

func TestInterpolationModifierForCpLimit(t *testing.T) {
	tests := []vmTestCase{
		// limit code points
		{`val x = ""; "{{x:L1}}"`, "", object.STRING_OBJ},
		{`val x = ""; "{{x:L2}}"`, "", object.STRING_OBJ},
		{`val x = 255; "{{x:L1}}"`, "2", object.STRING_OBJ},
		{`val x = 255; "{{x:L2}}"`, "25", object.STRING_OBJ},
		{`val x = 255; "{{x:L3}}"`, "255", object.STRING_OBJ},
		{`val x = 255; "{{x:L4}}"`, "255", object.STRING_OBJ},
		{`val x = 255; "{{x: L7 }}"`, "255", object.STRING_OBJ},

		{`val x = ""; "{{x:L-1}}"`, "", object.STRING_OBJ},
		{`val x = ""; "{{x:L-2}}"`, "", object.STRING_OBJ},
		{`val x = 255; "{{x:L-1}}"`, "5", object.STRING_OBJ},
		{`val x = 255; "{{x:L-2}}"`, "55", object.STRING_OBJ},
		{`val x = 255; "{{x:L-3}}"`, "255", object.STRING_OBJ},
		{`val x = 255; "{{x:L-4}}"`, "255", object.STRING_OBJ},
		{`val x = 255; "{{x:L-7}}"`, "255", object.STRING_OBJ},
		{`val x = 123456; "{{x:L-7}}"`, "123456", object.STRING_OBJ},
		{`val x = 1234567; "{{x:L-7}}"`, "1234567", object.STRING_OBJ},
		{`val x = 12345678; "{{x:L-7}}"`, "2345678", object.STRING_OBJ},

		// limit with internal overflow indicator
		{`val x = 1234567890; "{{x:L12(...)}}"`, "1234567890", object.STRING_OBJ},
		{`val x = 1234567890; "{{x:L11(...)}}"`, "1234567890", object.STRING_OBJ},
		{`val x = 1234567890; "{{x:L10(...)}}"`, "1234567890", object.STRING_OBJ},
		{`val x = 1234567890; "{{x:L9(...)}}"`, "123456...", object.STRING_OBJ},
		{`val x = 1234567890; "{{x:L8(...)}}"`, "12345...", object.STRING_OBJ},
		{`val x = 1234567890; "{{x:L7(...)}}"`, "1234...", object.STRING_OBJ},
		{`val x = 1234567890; "{{x:L6(...)}}"`, "123...", object.STRING_OBJ},
		{`val x = 1234567890; "{{x:L5(...)}}"`, "12...", object.STRING_OBJ},
		{`val x = 1234567890; "{{x:L4(...)}}"`, "1...", object.STRING_OBJ},
		{`val x = 1234567890; "{{x:L3(...)}}"`, "...", object.STRING_OBJ},
		{`val x = 1234567890; "{{x:L2(...)}}"`, "...", object.STRING_OBJ},
		{`val x = 1234567890; "{{x:L1(...)}}"`, "...", object.STRING_OBJ},

		{`val x = 1234567890; "{{x:L-12(...)}}"`, "1234567890", object.STRING_OBJ},
		{`val x = 1234567890; "{{x:L-11(...)}}"`, "1234567890", object.STRING_OBJ},
		{`val x = 1234567890; "{{x:L-10(...)}}"`, "1234567890", object.STRING_OBJ},
		{`val x = 1234567890; "{{x:L-9(...)}}"`, "...567890", object.STRING_OBJ},
		{`val x = 1234567890; "{{x:L-8(...)}}"`, "...67890", object.STRING_OBJ},
		{`val x = 1234567890; "{{x:L-7(...)}}"`, "...7890", object.STRING_OBJ},
		{`val x = 1234567890; "{{x:L-6(...)}}"`, "...890", object.STRING_OBJ},
		{`val x = 1234567890; "{{x:L-5(...)}}"`, "...90", object.STRING_OBJ},
		{`val x = 1234567890; "{{x:L-4(...)}}"`, "...0", object.STRING_OBJ},
		{`val x = 1234567890; "{{x:L-3(...)}}"`, "...", object.STRING_OBJ},
		{`val x = 1234567890; "{{x:L-2(...)}}"`, "...", object.STRING_OBJ},
		{`val x = 1234567890; "{{x:L-1(...)}}"`, "...", object.STRING_OBJ},

		{`val x = ""; "{{x:L3(...)}}"`, "", object.STRING_OBJ},
		{`val x = "1"; "{{x:L3(...)}}"`, "1", object.STRING_OBJ},
		{`val x = "12"; "{{x:L3(...)}}"`, "12", object.STRING_OBJ},
		{`val x = "123"; "{{x:L3(...)}}"`, "123", object.STRING_OBJ},
		{`val x = ""; "{{x:L2(...)}}"`, "", object.STRING_OBJ},
		{`val x = "1"; "{{x:L2(...)}}"`, "1", object.STRING_OBJ},
		{`val x = "12"; "{{x:L2(...)}}"`, "12", object.STRING_OBJ},
		{`val x = "123"; "{{x:L2(...)}}"`, "...", object.STRING_OBJ},
		{`val x = ""; "{{x:L1(...)}}"`, "", object.STRING_OBJ},
		{`val x = "1"; "{{x:L1(...)}}"`, "1", object.STRING_OBJ},
		{`val x = "12"; "{{x:L1(...)}}"`, "...", object.STRING_OBJ},
		{`val x = "123"; "{{x:L1(...)}}"`, "...", object.STRING_OBJ},

		{`val x = ""; "{{x:L-3(...)}}"`, "", object.STRING_OBJ},
		{`val x = "1"; "{{x:L-3(...)}}"`, "1", object.STRING_OBJ},
		{`val x = "12"; "{{x:L-3(...)}}"`, "12", object.STRING_OBJ},
		{`val x = "123"; "{{x:L-3(...)}}"`, "123", object.STRING_OBJ},
		{`val x = ""; "{{x:L-2(...)}}"`, "", object.STRING_OBJ},
		{`val x = "1"; "{{x:L-2(...)}}"`, "1", object.STRING_OBJ},
		{`val x = "12"; "{{x:L-2(...)}}"`, "12", object.STRING_OBJ},
		{`val x = "123"; "{{x:L-2(...)}}"`, "...", object.STRING_OBJ},
		{`val x = ""; "{{x:L-1(...)}}"`, "", object.STRING_OBJ},
		{`val x = "1"; "{{x:L-1(...)}}"`, "1", object.STRING_OBJ},
		{`val x = "12"; "{{x:L-1(...)}}"`, "...", object.STRING_OBJ},
		{`val x = "123"; "{{x:L-1(...)}}"`, "...", object.STRING_OBJ},

		// overflow indicator with escape sequences
		{`val x = 1234567890; QS"{{x:L-7(\)}}"`, "\\567890", object.STRING_OBJ},
		{`val x = 1234567890; "{{x: L-7(\\) }}"`, "\\567890", object.STRING_OBJ},
		{`val x = 1234567890; "{{x:L-7(\x2A)}}"`, "*567890", object.STRING_OBJ},
		{`val x = 1234567890; "{{x:L7(\x2A)}}"`, "123456*", object.STRING_OBJ},
		{`val x = 1234567890; "{{x:L7(\x2A!)}}"`, "12345*!", object.STRING_OBJ},
		{`val x = 255; "{{x:-10(\x2A)}}"`, "255*******", object.STRING_OBJ},
	}

	runVmTests(t, tests, false, false)
}

func TestInterpolationModifierForGraphemeLimit(t *testing.T) {
	tests := []vmTestCase{
		// limit by graphemes

		// farmer with pitchfork: U+1F9D1 U+200D U+1F33E
		// USA flag: U+1F1FA U+1F1F8
		// Finland flag: U+1F1EB U+1F1EE

		{`val farmer = "\U0001F9D1\u200D\U0001F33E"
		  val x = farmer * 7
		  "{{x:Lg1}}"`, "\U0001F9D1\u200D\U0001F33E", object.STRING_OBJ},

		{`val farmer = "\U0001F9D1\u200D\U0001F33E"
	      val flag1 = "\U0001F1FA\U0001F1F8"
		  val flag2 = "\U0001F1EB\U0001F1EE"
		  var x = farmer ~ flag1 ~ flag2
		  x *= 20  # now at 60 graphemes
		  "{{x:Lg1}}"`, "\U0001F9D1\u200D\U0001F33E", object.STRING_OBJ},

		{`val farmer = "\U0001F9D1\u200D\U0001F33E"
	      val flag1 = "\U0001F1FA\U0001F1F8"
		  val flag2 = "\U0001F1EB\U0001F1EE"
		  var x = farmer ~ flag1 ~ flag2
		  x *= 20  # now at 60 graphemes
		  "{{x:Lg-1}}"`, "\U0001F1EB\U0001F1EE", object.STRING_OBJ},

		{`val farmer = "\U0001F9D1\u200D\U0001F33E"
	      val flag1 = "\U0001F1FA\U0001F1F8"
		  val flag2 = "\U0001F1EB\U0001F1EE"
		  var x = farmer ~ flag1 ~ flag2
		  x *= 20  # now at 60 graphemes
		  "{{x:Lg4}}"`, "\U0001F9D1\u200D\U0001F33E\U0001F1FA\U0001F1F8\U0001F1EB\U0001F1EE\U0001F9D1\u200D\U0001F33E", object.STRING_OBJ},

		{`val farmer = "\U0001F9D1\u200D\U0001F33E"
	      val flag1 = "\U0001F1FA\U0001F1F8"
		  val flag2 = "\U0001F1EB\U0001F1EE"
		  var x = farmer ~ flag1 ~ flag2
		  x *= 20  # now at 60 graphemes
		  "{{x:Lg-4}}"`, "\U0001F1EB\U0001F1EE\U0001F9D1\u200D\U0001F33E\U0001F1FA\U0001F1F8\U0001F1EB\U0001F1EE", object.STRING_OBJ},

		// for simple code points, should come out same...
		{`val x = ""; "{{x:Lg1}}"`, "", object.STRING_OBJ},
		{`val x = ""; "{{x:Lg2}}"`, "", object.STRING_OBJ},
		{`val x = 255; "{{x:Lg1}}"`, "2", object.STRING_OBJ},
		{`val x = 255; "{{x:Lg2}}"`, "25", object.STRING_OBJ},
		{`val x = 255; "{{x:Lg-7}}"`, "255", object.STRING_OBJ},
		{`val x = 123456; "{{x:Lg-7}}"`, "123456", object.STRING_OBJ},
		{`val x = 1234567; "{{x:Lg-7}}"`, "1234567", object.STRING_OBJ},
		{`val x = 12345678; "{{x:Lg-7}}"`, "2345678", object.STRING_OBJ},
		{`val x = 1234567890; "{{x:Lg-10(...)}}"`, "1234567890", object.STRING_OBJ},
		{`val x = 1234567890; "{{x:Lg-9(...)}}"`, "...567890", object.STRING_OBJ},
		{`val x = 1234567890; "{{x:Lg-4(...)}}"`, "...0", object.STRING_OBJ},
		{`val x = 1234567890; "{{x:Lg-3(...)}}"`, "...", object.STRING_OBJ},
		{`val x = 1234567890; "{{x:Lg-2(...)}}"`, "...", object.STRING_OBJ},
	}

	runVmTests(t, tests, false, false)
}

func TestInterpolationModifierForRound(t *testing.T) {
	tests := []vmTestCase{
		// round to integer
		{`val x = 1.23456879; "{{x:r}}"`, "1", object.STRING_OBJ},
		{`val x = 1.23456879; "{{x:r0}}"`, "1", object.STRING_OBJ},
		{`val x = 123456879; "{{x:r}}"`, "123456879", object.STRING_OBJ},
		{`val x = 42.7; "{{x:r}}"`, "43", object.STRING_OBJ},
		{`val x = 42.7; "{{x:r:X}}"`, "2B", object.STRING_OBJ},
		{`val x = 42.7777; "{{x:r}}"`, "43", object.STRING_OBJ},
		{`val x = 42.3333; "{{x:r}}"`, "42", object.STRING_OBJ},
		{`val x = 42.3333; "{{x:r:X}}"`, "2A", object.STRING_OBJ},

		{`val x = -1.23456879; "{{x:r}}"`, "-1", object.STRING_OBJ},
		{`val x = -1.23456879; "{{x:r0}}"`, "-1", object.STRING_OBJ},
		{`val x = -123456879; "{{x:r}}"`, "-123456879", object.STRING_OBJ},
		{`val x = -42.7; "{{x:r}}"`, "-43", object.STRING_OBJ},
		{`val x = -42.7; "{{x:r:X}}"`, "-2B", object.STRING_OBJ},
		{`val x = -42.7777; "{{x:r}}"`, "-43", object.STRING_OBJ},
		{`val x = -42.3333; "{{x:r}}"`, "-42", object.STRING_OBJ},
		{`val x = -42.3333; "{{x:r:X}}"`, "-2A", object.STRING_OBJ},

		// round on integer
		{`val x = 42.3333; "{{x:r-1}}"`, "40", object.STRING_OBJ},
		{`val x = 42.3333; "{{x:r-1:X}}"`, "28", object.STRING_OBJ},
		{`val x = 42.3333; "{{x:r-2}}"`, "0", object.STRING_OBJ},
		{`val x = 45.3333; "{{x:r-1}}"`, "50", object.STRING_OBJ},
		{`val x = 45.3333; "{{x:r-1:X}}"`, "32", object.STRING_OBJ},

		// half away from zero
		{`mode rounding = _round'halfawayfrom0
		  val x = 1.5; "{{x:r}}"`, "2", object.STRING_OBJ},
		{`mode rounding = _round'halfawayfrom0
		  val x = 2.5; "{{x:r}}"`, "3", object.STRING_OBJ},
		{`mode rounding = _round'halfawayfrom0
		  val x = 2.25; "{{x:r1}}"`, "2.3", object.STRING_OBJ},

		// half away from zero with negative numbers
		{`mode rounding = _round'halfawayfrom0
		  val x = -1.4; "{{x:r}}"`, "-1", object.STRING_OBJ},
		{`mode rounding = _round'halfawayfrom0
		  val x = -1.5; "{{x:r}}"`, "-2", object.STRING_OBJ},
		{`mode rounding = _round'halfawayfrom0
		  val x = -1.6; "{{x:r}}"`, "-2", object.STRING_OBJ},
		{`mode rounding = _round'halfawayfrom0
		  val x = -2.4; "{{x:r}}"`, "-2", object.STRING_OBJ},
		{`mode rounding = _round'halfawayfrom0
		  val x = -2.5; "{{x:r}}"`, "-3", object.STRING_OBJ},
		{`mode rounding = _round'halfawayfrom0
		  val x = -2.6; "{{x:r}}"`, "-3", object.STRING_OBJ},
		{`mode rounding = _round'halfawayfrom0
		  val x = -2.24; "{{x:r1}}"`, "-2.2", object.STRING_OBJ},
		{`mode rounding = _round'halfawayfrom0
		  val x = -2.26; "{{x:r1}}"`, "-2.3", object.STRING_OBJ},
		{`mode rounding = _round'halfawayfrom0
		  val x = -2.25; "{{x:r1}}"`, "-2.3", object.STRING_OBJ},

		// half even
		{`mode rounding = _round'halfeven
		  val x = 1.5; "{{x:r}}"`, "2", object.STRING_OBJ},
		{`mode rounding = _round'halfeven
		  val x = 2.5; "{{x:r}}"`, "2", object.STRING_OBJ},
		{`mode rounding = _round'halfeven
		  val x = 2.25; "{{x:r1}}"`, "2.2", object.STRING_OBJ},

		{`mode rounding = _round'halfeven
		  val x = -1.5; "{{x:r}}"`, "-2", object.STRING_OBJ},
		{`mode rounding = _round'halfeven
		  val x = -2.5; "{{x:r}}"`, "-2", object.STRING_OBJ},
		{`mode rounding = _round'halfeven
		  val x = -3.5; "{{x:r}}"`, "-4", object.STRING_OBJ},
		{`mode rounding = _round'halfeven
		  val x = -2.25; "{{x:r1}}"`, "-2.2", object.STRING_OBJ},

		// round with padding zeroes (normal)
		{`val x = 1.23; "{{x: r4 }}"`, "1.2300", object.STRING_OBJ},
		{`val x = 1.234; "{{x:r4}}"`, "1.2340", object.STRING_OBJ},
		{`val x = 1.2345; "{{x:r4}}"`, "1.2345", object.STRING_OBJ},
		{`val x = 1.23456879; "{{x:r1}}"`, "1.2", object.STRING_OBJ},
		{`val x = 1.23456879; "{{x:r4}}"`, "1.2346", object.STRING_OBJ},
		{`val x = 123456879; "{{x:r1}}"`, "123456879.0", object.STRING_OBJ},
		{`val x = 1.2000000; "{{x:r4}}"`, "1.2000", object.STRING_OBJ},
		{`val x = 1.2000000; "{{x:r7}}"`, "1.2000000", object.STRING_OBJ},

		// trim padding zeroes
		{`val x = 1.23; "{{x:r4-}}"`, "1.23", object.STRING_OBJ},
		{`val x = 1.234; "{{x:r4-}}"`, "1.234", object.STRING_OBJ},
		{`val x = 1.2345; "{{x:r4-}}"`, "1.2345", object.STRING_OBJ},
		{`val x = 1.23456879; "{{x:r1-}}"`, "1.2", object.STRING_OBJ},
		{`val x = 1.23456879; "{{x:r4-}}"`, "1.2346", object.STRING_OBJ},
		{`val x = 123456879; "{{x:r1-}}"`, "123456879", object.STRING_OBJ},
		{`val x = 1.2000000; "{{x:r4-}}"`, "1.2", object.STRING_OBJ},
		{`val x = 1.2000000; "{{x:r7-}}"`, "1.2", object.STRING_OBJ},

		// don't trim or add zeroes
		{`val x = 1.2000000; "{{x:r4!}}"`, "1.2000", object.STRING_OBJ},
		{`val x = 1.2000000; "{{x:r7!}}"`, "1.2000000", object.STRING_OBJ},
		{`val x = 1.2000000; "{{x:r9!}}"`, "1.2000000", object.STRING_OBJ},

		{`val x = 123.1590; "{{x:r5!}}"`, "123.1590", object.STRING_OBJ},
		{`val x = 123.1590; "{{x:r5}}"`, "123.15900", object.STRING_OBJ},
		{`val x = 123.1590; "{{x:r5-}}"`, "123.159", object.STRING_OBJ},
	}

	runVmTests(t, tests, false, false)
}

func TestInterpolationModifierForTrunc(t *testing.T) {
	tests := []vmTestCase{
		// truncate to integer
		{`val x = 1.23456879; "{{x:t}}"`, "1", object.STRING_OBJ},
		{`val x = 1.23456879; "{{x:t0}}"`, "1", object.STRING_OBJ},
		{`val x = 123456879; "{{x:t}}"`, "123456879", object.STRING_OBJ},
		{`val x = 42.7; "{{x:t}}"`, "42", object.STRING_OBJ},
		{`val x = 42.7; "{{x:t:X}}"`, "2A", object.STRING_OBJ},
		{`val x = 42.7777; "{{x:t}}"`, "42", object.STRING_OBJ},
		{`val x = 42.3333; "{{x:t}}"`, "42", object.STRING_OBJ},
		{`val x = 42.3333; "{{x:t:X}}"`, "2A", object.STRING_OBJ},

		// truncate on integer
		{`val x = 42.3333; "{{x:t-1}}"`, "40", object.STRING_OBJ},
		{`val x = 42.3333; "{{x:t-1:X}}"`, "28", object.STRING_OBJ},
		{`val x = 42.3333; "{{x:t-2}}"`, "0", object.STRING_OBJ},

		// truncate with padding zeroes (normal)
		{`val x = 1.23; "{{x: t4 }}"`, "1.2300", object.STRING_OBJ},
		{`val x = 1.234; "{{x:t4}}"`, "1.2340", object.STRING_OBJ},
		{`val x = 1.2345; "{{x:t4}}"`, "1.2345", object.STRING_OBJ},
		{`val x = 1.23456879; "{{x:t1}}"`, "1.2", object.STRING_OBJ},
		{`val x = 1.23456879; "{{x:t4}}"`, "1.2345", object.STRING_OBJ},
		{`val x = 123456879; "{{x:t1}}"`, "123456879.0", object.STRING_OBJ},
		{`val x = 1.2000000; "{{x:t4}}"`, "1.2000", object.STRING_OBJ},
		{`val x = 1.2000000; "{{x:t7}}"`, "1.2000000", object.STRING_OBJ},

		// trim padding zeroes
		{`val x = 1.23; "{{x:t4-}}"`, "1.23", object.STRING_OBJ},
		{`val x = 1.234; "{{x:t4-}}"`, "1.234", object.STRING_OBJ},
		{`val x = 1.2345; "{{x:t4-}}"`, "1.2345", object.STRING_OBJ},
		{`val x = 1.23456879; "{{x:t1-}}"`, "1.2", object.STRING_OBJ},
		{`val x = 1.23456879; "{{x:t4-}}"`, "1.2345", object.STRING_OBJ},
		{`val x = 123456879; "{{x:t1-}}"`, "123456879", object.STRING_OBJ},
		{`val x = 1.2000000; "{{x:t4-}}"`, "1.2", object.STRING_OBJ},
		{`val x = 1.2000000; "{{x:t7-}}"`, "1.2", object.STRING_OBJ},

		// don't trim or add zeroes
		{`val x = 1.2000000; "{{x:t4!}}"`, "1.2000", object.STRING_OBJ},
		{`val x = 1.2000000; "{{x:t7!}}"`, "1.2000000", object.STRING_OBJ},
		{`val x = 1.2000000; "{{x:t9!}}"`, "1.2000000", object.STRING_OBJ},

		{`val x = 123.1590; "{{x:t5!}}"`, "123.1590", object.STRING_OBJ},
		{`val x = 123.1590; "{{x:t5}}"`, "123.15900", object.STRING_OBJ},
		{`val x = 123.1590; "{{x:t5-}}"`, "123.159", object.STRING_OBJ},
	}

	runVmTests(t, tests, false, false)
}

func TestInterpolationModifierForHexAndBaseConversion(t *testing.T) {
	tests := []vmTestCase{
		// hexadecimal
		{`val x = 255; "{{x:X}}"`, "FF", object.STRING_OBJ},
		{`val x = 255; "{{x:x}}"`, "ff", object.STRING_OBJ},
		{`val x = 10; "{{x:X}}"`, "A", object.STRING_OBJ},
		{`val x = 16x0D; "{{x:x}}"`, "d", object.STRING_OBJ},
		{`val x = 16x0A; "{{x:X02}}"`, "0A", object.STRING_OBJ},
		{`val x = 16x0D; "{{x:x02}}"`, "0d", object.STRING_OBJ},
		{`val x = 100; "{{x:X012}}"`, "000000000064", object.STRING_OBJ},

		{`val x = 16x0A; "{{x:X2}}"`, " A", object.STRING_OBJ},
		{`val x = 16x0D; "{{x:x2}}"`, " d", object.STRING_OBJ},
		{`val x = 100; "{{x:X12}}"`, "          64", object.STRING_OBJ},

		{`val x = 255.3; "{{x: X }}"`, "FF", object.STRING_OBJ},
		{`val x = 255.7; "{{x:X}}"`, "100", object.STRING_OBJ},

		{`val x = -255; "{{x:X}}"`, "-FF", object.STRING_OBJ},
		{`val x = -255; "{{x:x}}"`, "-ff", object.STRING_OBJ},
		{`val x = -10; "{{x:X}}"`, "-A", object.STRING_OBJ},
		{`val x = -16x0D; "{{x:x}}"`, "-d", object.STRING_OBJ},
		{`val x = -16x0A; "{{x:X2}}"`, "-A", object.STRING_OBJ},
		{`val x = -16x0A; "{{x:X3}}"`, " -A", object.STRING_OBJ},
		{`val x = -16x0A; "{{x:X03}}"`, "-0A", object.STRING_OBJ},
		{`val x = -16x0D; "{{x:x2}}"`, "-d", object.STRING_OBJ},
		{`val x = -100; "{{x:X12}}"`, "         -64", object.STRING_OBJ},
		{`val x = -100; "{{x:X012}}"`, "-00000000064", object.STRING_OBJ},

		{`val x = -255.3; "{{x:X}}"`, "-FF", object.STRING_OBJ},
		{`val x = -255.7; "{{x:X}}"`, "-100", object.STRING_OBJ},

		// hex with required sign
		{`val x = 255; "{{x:+x}}"`, "+ff", object.STRING_OBJ},
		{`val x = -255; "{{x:+x}}"`, "-ff", object.STRING_OBJ},

		{`val x = 16x0A; "{{x:+X3}}"`, " +A", object.STRING_OBJ},
		{`val x = 16x0A; "{{x:+X03}}"`, "+0A", object.STRING_OBJ},
		{`val x = 255.3; "{{x:+x}}"`, "+ff", object.STRING_OBJ},
		{`val x = 255.7; "{{x:+x}}"`, "+100", object.STRING_OBJ},
		{`val x = -255.3; "{{x:+x}}"`, "-ff", object.STRING_OBJ},
		{`val x = -255.7; "{{x:+x}}"`, "-100", object.STRING_OBJ},

		// to base ... (2 to 36)
		{`val x = 255; "{{x:16x}}"`, "ff", object.STRING_OBJ},
		{`val x = 255; "{{x:16X}}"`, "FF", object.STRING_OBJ},
		{`val x = 255; "{{x:8x}}"`, "377", object.STRING_OBJ},
		{`val x = 255; "{{x:8X}}"`, "377", object.STRING_OBJ},
		{`val x = 255; "{{x:2x}}"`, "11111111", object.STRING_OBJ},
		{`val x = 255; "{{x:4x}}"`, "3333", object.STRING_OBJ},
		{`val x = 255; "{{x:11x}}"`, "212", object.STRING_OBJ},

		{`val x = 255; "{{x:11x1}}"`, "212", object.STRING_OBJ},
		{`val x = 255; "{{x:11x11}}"`, "        212", object.STRING_OBJ},
		{`val x = 255; "{{x:11x011}}"`, "00000000212", object.STRING_OBJ},

		{`val x = 255.3; "{{x: 11x1 }}"`, "212", object.STRING_OBJ},
		{`val x = 255.7; "{{x:11x11}}"`, "        213", object.STRING_OBJ},
		{`val x = 255.7; "{{x:11x011}}"`, "00000000213", object.STRING_OBJ},

		{`val x = -255; "{{x:16x}}"`, "-ff", object.STRING_OBJ},
		{`val x = -255; "{{x:16X}}"`, "-FF", object.STRING_OBJ},
		{`val x = -255; "{{x:8x}}"`, "-377", object.STRING_OBJ},
		{`val x = -255; "{{x:8X}}"`, "-377", object.STRING_OBJ},
		{`val x = -255; "{{x:2x}}"`, "-11111111", object.STRING_OBJ},
		{`val x = -255; "{{x:4x}}"`, "-3333", object.STRING_OBJ},
		{`val x = -255; "{{x:11x}}"`, "-212", object.STRING_OBJ},

		{`val x = -255; "{{x:11x1}}"`, "-212", object.STRING_OBJ},
		{`val x = -255; "{{x:11x11}}"`, "       -212", object.STRING_OBJ},
		{`val x = -255; "{{x:11x01}}"`, "-212", object.STRING_OBJ},
		{`val x = -255; "{{x:11x011}}"`, "-0000000212", object.STRING_OBJ},

		// to base with required sign
		{`val x = 255; "{{x:+11x1}}"`, "+212", object.STRING_OBJ},
		{`val x = 255; "{{x:+11x5}}"`, " +212", object.STRING_OBJ},
		{`val x = 255; "{{x:+11x11}}"`, "       +212", object.STRING_OBJ},
		{`val x = -255; "{{x:+11x1}}"`, "-212", object.STRING_OBJ},
		{`val x = -255; "{{x:+11x11}}"`, "       -212", object.STRING_OBJ},

		{`val x = 255; "{{x:+11x05}}"`, "+0212", object.STRING_OBJ},
		{`val x = 255; "{{x:+11x011}}"`, "+0000000212", object.STRING_OBJ},
		{`val x = -255; "{{x:+11x01}}"`, "-212", object.STRING_OBJ},
		{`val x = -255; "{{x:+11x011}}"`, "-0000000212", object.STRING_OBJ},

		{`val x = 255.3; "{{x:+11x1}}"`, "+212", object.STRING_OBJ},
		{`val x = 255.7; "{{x:+11x11}}"`, "       +213", object.STRING_OBJ},
		{`val x = -255.3; "{{x:+11x1}}"`, "-212", object.STRING_OBJ},
		{`val x = -255.7; "{{x:+11x11}}"`, "       -213", object.STRING_OBJ},

		{`val x = 255.3; "{{x:+11x01}}"`, "+212", object.STRING_OBJ},
		{`val x = 255.7; "{{x:+11x011}}"`, "+0000000213", object.STRING_OBJ},
		{`val x = -255.3; "{{x:+11x01}}"`, "-212", object.STRING_OBJ},
		{`val x = -255.7; "{{x:+11x011}}"`, "-0000000213", object.STRING_OBJ},
	}

	runVmTests(t, tests, false, false)
}

func TestInterpolationModifierForFixedPoint(t *testing.T) {
	tests := []vmTestCase{
		{`val x = 255; "{{x:10x1.2}}"`, "255.00", object.STRING_OBJ},
		{`val x = 255; "{{x:10x7.2}}"`, "    255.00", object.STRING_OBJ},

		{`val x = -255; "{{x:+10x1.2}}"`, "-255.00", object.STRING_OBJ},
		{`val x = -255; "{{x:+10x7.2}}"`, "   -255.00", object.STRING_OBJ},
		{`val x = 255; "{{x:+10x1.2}}"`, "+255.00", object.STRING_OBJ},
		{`val x = 255; "{{x:+10x7.2}}"`, "   +255.00", object.STRING_OBJ},

		{`val x = 3.14159; "{{x:10x1.2}}"`, "3.14", object.STRING_OBJ},
		{`val x = 3.14159; "{{x:10x3.0}}"`, "  3", object.STRING_OBJ},

		{`val x = 255; "{{x:10x1.2}}"`, "255.00", object.STRING_OBJ},
		{`val x = 255; "{{x:10x07.2}}"`, "0000255.00", object.STRING_OBJ},

		{`val x = -255; "{{x: +10x1.2 }}"`, "-255.00", object.STRING_OBJ},
		{`val x = -255; "{{x:+10x07.2}}"`, "-000255.00", object.STRING_OBJ},
		{`val x = 255; "{{x:+10x1.2}}"`, "+255.00", object.STRING_OBJ},
		{`val x = 255; "{{x:+10x07.2}}"`, "+000255.00", object.STRING_OBJ},

		{`val x = 3.14159; "{{x:10x1.2}}"`, "3.14", object.STRING_OBJ},
		{`val x = 3.14159; "{{x:10x03.0}}"`, "003", object.STRING_OBJ},

		// trim trailing zeroes
		{`val x = 3.14159; "{{x:10x1.2-}}"`, "3.14", object.STRING_OBJ},
		{`val x = 3.10; "{{x:10x1.2-}}"`, "3.1", object.STRING_OBJ},
		{`val x = 3.100; "{{x:10x1.2-}}"`, "3.1", object.STRING_OBJ},
		{`val x = 3.1001; "{{x:10x1.2-}}"`, "3.1", object.STRING_OBJ},
		{`val x = 3.1; "{{x:10x1.2-}}"`, "3.1", object.STRING_OBJ},
		{`val x = 3.11; "{{x:10x1.2-}}"`, "3.11", object.STRING_OBJ},
		{`val x = 3.111; "{{x:10x1.2-}}"`, "3.11", object.STRING_OBJ},

		// add trailing zeroes?
		{`val x = 3.14159; "{{x:10x1.2!}}"`, "3.14", object.STRING_OBJ},
		{`val x = 3.1; "{{x:10x1.2!}}"`, "3.1", object.STRING_OBJ},
		{`val x = 3.1; "{{x:10x1.2}}"`, "3.10", object.STRING_OBJ},
		{`val x = 3; "{{x:10x1.2!}}"`, "3", object.STRING_OBJ},
		{`val x = 3; "{{x:10x1.2}}"`, "3.00", object.STRING_OBJ},

		// comma for decimal point
		{`val x = 3.14159; "{{x:10x1,2}}"`, "3,14", object.STRING_OBJ},
	}

	runVmTests(t, tests, false, false)
}

func TestInterpolationModifierForScientificNotation(t *testing.T) {
	tests := []vmTestCase{
		// every other set of tests here requiring a sign on the exponent

		// scientific notation without a scale
		{`val x = 123456879; "{{x:e}}"`, "1.23456879e8", object.STRING_OBJ},
		{`val x = 123456879; "{{x:E}}"`, "1.23456879E8", object.STRING_OBJ},
		{`val x = 1234.56879; "{{x:e}}"`, "1.23456879e3", object.STRING_OBJ},
		{`val x = 1.23456879; "{{x:e}}"`, "1.23456879e0", object.STRING_OBJ},
		{`val x = 0.123456879; "{{x:e}}"`, "1.23456879e-1", object.STRING_OBJ},
		{`val x = 0.0123456879; "{{x:e}}"`, "1.23456879e-2", object.STRING_OBJ},
		{`val x = 0.000000123456879; "{{x:e}}"`, "1.23456879e-7", object.STRING_OBJ},

		{`val x = 123456879; "{{x:e+}}"`, "1.23456879e+8", object.STRING_OBJ},
		{`val x = 123456879; "{{x:E+}}"`, "1.23456879E+8", object.STRING_OBJ},
		{`val x = 1234.56879; "{{x:e+}}"`, "1.23456879e+3", object.STRING_OBJ},
		{`val x = 1.23456879; "{{x:e+}}"`, "1.23456879e+0", object.STRING_OBJ},
		{`val x = 0.123456879; "{{x:e+}}"`, "1.23456879e-1", object.STRING_OBJ},
		{`val x = 0.0123456879; "{{x:e+}}"`, "1.23456879e-2", object.STRING_OBJ},
		{`val x = 0.000000123456879; "{{x:e+}}"`, "1.23456879e-7", object.STRING_OBJ},

		// scientific notation with a scale (rounding)
		{`val x = 1.23456879e+7000; "{{x:1.10-e}}"`, "1.23456879e7000", object.STRING_OBJ},
		{`val x = 1.23456879e-7000; "{{x:1.10-e}}"`, "1.23456879e-7000", object.STRING_OBJ},

		{`val x = 1.23456879e+7000; "{{x:1.10-e+}}"`, "1.23456879e+7000", object.STRING_OBJ},
		{`val x = 1.23456879e-7000; "{{x:1.10-e+}}"`, "1.23456879e-7000", object.STRING_OBJ},

		{`val x = 123456879; "{{x: 1.0e }}"`, "1e8", object.STRING_OBJ},
		{`val x = 173456879; "{{x:1.0e}}"`, "2e8", object.STRING_OBJ},
		{`val x = 123456879; "{{x:1.3e}}"`, "1.235e8", object.STRING_OBJ},
		{`val x = 123456879; "{{x:1.3E}}"`, "1.235E8", object.STRING_OBJ},
		{`val x = 1234.56879; "{{x:1.3e}}"`, "1.235e3", object.STRING_OBJ},
		{`val x = 1.23456879; "{{x:1.4e}}"`, "1.2346e0", object.STRING_OBJ},
		{`val x = 0.123456879; "{{x:1.4e}}"`, "1.2346e-1", object.STRING_OBJ},
		{`val x = 0.0123456879; "{{x:1.4e}}"`, "1.2346e-2", object.STRING_OBJ},
		{`val x = 0.000000123456879; "{{x:1.4e}}"`, "1.2346e-7", object.STRING_OBJ},

		{`val x = 123456879; "{{x:1.0e+}}"`, "1e+8", object.STRING_OBJ},
		{`val x = 173456879; "{{x:1.0e+}}"`, "2e+8", object.STRING_OBJ},
		{`val x = 123456879; "{{x:1.3e+}}"`, "1.235e+8", object.STRING_OBJ},
		{`val x = 123456879; "{{x:1.3E+}}"`, "1.235E+8", object.STRING_OBJ},
		{`val x = 1234.56879; "{{x:1.3e+}}"`, "1.235e+3", object.STRING_OBJ},
		{`val x = 1.23456879; "{{x:1.4e+}}"`, "1.2346e+0", object.STRING_OBJ},
		{`val x = 0.123456879; "{{x:1.4e+}}"`, "1.2346e-1", object.STRING_OBJ},
		{`val x = 0.0123456879; "{{x:1.4e+}}"`, "1.2346e-2", object.STRING_OBJ},
		{`val x = 0.000000123456879; "{{x:1.4e+}}"`, "1.2346e-7", object.STRING_OBJ},

		{`val x = 1234; "{{x:1.3e}}"`, "1.234e3", object.STRING_OBJ},
		{`val x = 10000; "{{x:1.3e}}"`, "1.000e4", object.STRING_OBJ},
		{`val x = 1234; "{{x:1.3-e}}"`, "1.234e3", object.STRING_OBJ},
		{`val x = 10000; "{{x:1.3-e}}"`, "1e4", object.STRING_OBJ},

		{`val x = 1234; "{{x:1.3e+}}"`, "1.234e+3", object.STRING_OBJ},
		{`val x = 10000; "{{x:1.3e+}}"`, "1.000e+4", object.STRING_OBJ},
		{`val x = 1234; "{{x:1.3-e+}}"`, "1.234e+3", object.STRING_OBJ},
		{`val x = 10000; "{{x:1.3-e+}}"`, "1e+4", object.STRING_OBJ},

		{`val x = -1234; "{{x:1.3e}}"`, "-1.234e3", object.STRING_OBJ},
		{`val x = -10000; "{{x:1.3e}}"`, "-1.000e4", object.STRING_OBJ},
		{`val x = -1234; "{{x:1.3-e}}"`, "-1.234e3", object.STRING_OBJ},
		{`val x = -10000; "{{x:1.3-e}}"`, "-1e4", object.STRING_OBJ},

		{`val x = -1234; "{{x:1.3e+}}"`, "-1.234e+3", object.STRING_OBJ},
		{`val x = -10000; "{{x:1.3e+}}"`, "-1.000e+4", object.STRING_OBJ},
		{`val x = -1234; "{{x:1.3-e+}}"`, "-1.234e+3", object.STRING_OBJ},
		{`val x = -10000; "{{x:1.3-e+}}"`, "-1e+4", object.STRING_OBJ},

		{`val x = 0.1; "{{x:1.4e}}"`, "1.0000e-1", object.STRING_OBJ},
		{`val x = 0.0000001; "{{x:1.4e}}"`, "1.0000e-7", object.STRING_OBJ},
		{`val x = 0.0000001; "{{x:1.4-e}}"`, "1e-7", object.STRING_OBJ},
		{`val x = 0.00000017; "{{x:1.4e}}"`, "1.7000e-7", object.STRING_OBJ},
		{`val x = 0.00000017; "{{x:1.4-e}}"`, "1.7e-7", object.STRING_OBJ},

		{`val x = 0.1; "{{x:1.4e+}}"`, "1.0000e-1", object.STRING_OBJ},
		{`val x = 0.0000001; "{{x:1.4e+}}"`, "1.0000e-7", object.STRING_OBJ},
		{`val x = 0.0000001; "{{x:1.4-e+}}"`, "1e-7", object.STRING_OBJ},
		{`val x = 0.00000017; "{{x:1.4e+}}"`, "1.7000e-7", object.STRING_OBJ},
		{`val x = 0.00000017; "{{x:1.4-e+}}"`, "1.7e-7", object.STRING_OBJ},

		// ... including scaling of the exponent
		{`val x = 1234; "{{x:1.3e2}}"`, "1.234e03", object.STRING_OBJ},
		{`val x = 10000; "{{x:1.3e2}}"`, "1.000e04", object.STRING_OBJ},
		{`val x = 1234; "{{x:1.3-e4}}"`, "1.234e0003", object.STRING_OBJ},
		{`val x = 10000; "{{x:1.3-e4}}"`, "1e0004", object.STRING_OBJ},

		{`val x = 1234; "{{x: 1.3e+2}}"`, "1.234e+03", object.STRING_OBJ},
		{`val x = 10000; "{{x:1.3e+2}}"`, "1.000e+04", object.STRING_OBJ},
		{`val x = 1234; "{{x:1.3-e+4}}"`, "1.234e+0003", object.STRING_OBJ},
		{`val x = 10000; "{{x:1.3-e+4}}"`, "1e+0004", object.STRING_OBJ},

		// scientific notation requiring a sign on first number
		{`val x = 1234; "{{x:+1.3e}}"`, "+1.234e3", object.STRING_OBJ},
		{`val x = 10000; "{{x:+1.3e}}"`, "+1.000e4", object.STRING_OBJ},
		{`val x = -1234; "{{x:+1.3e}}"`, "-1.234e3", object.STRING_OBJ},
		{`val x = -10000; "{{x:+1.3e}}"`, "-1.000e4", object.STRING_OBJ},

		{`val x = 1234; "{{x:+1.3e+}}"`, "+1.234e+3", object.STRING_OBJ},
		{`val x = 10000; "{{x:+1.3e+}}"`, "+1.000e+4", object.STRING_OBJ},
		{`val x = -1234; "{{x:+1.3e+}}"`, "-1.234e+3", object.STRING_OBJ},
		{`val x = -10000; "{{x:+1.3e+}}"`, "-1.000e+4", object.STRING_OBJ},

		// ... with required sign and scaling of exponent
		{`val x = 1234; "{{x:+1.3e2}}"`, "+1.234e03", object.STRING_OBJ},
		{`val x = 10000; "{{x:+1.3e2}}"`, "+1.000e04", object.STRING_OBJ},
		{`val x = 1234; "{{x:+1.3-e4}}"`, "+1.234e0003", object.STRING_OBJ},
		{`val x = 10000; "{{x:+1.3-e4}}"`, "+1e0004", object.STRING_OBJ},

		{`val x = 1234; "{{x:+1.3e+2}}"`, "+1.234e+03", object.STRING_OBJ},
		{`val x = 10000; "{{x:+1.3e+2}}"`, "+1.000e+04", object.STRING_OBJ},
		{`val x = 1234; "{{x:+1.3-e+4}}"`, "+1.234e+0003", object.STRING_OBJ},
		{`val x = 10000; "{{x:+1.3-e+4}}"`, "+1e+0004", object.STRING_OBJ},

		{`val x = -1234; "{{x:+1.3e2}}"`, "-1.234e03", object.STRING_OBJ},
		{`val x = -10000; "{{x:+1.3e2}}"`, "-1.000e04", object.STRING_OBJ},
		{`val x = -1234; "{{x:+1.3-e4}}"`, "-1.234e0003", object.STRING_OBJ},
		{`val x = -10000; "{{x:+1.3-e4}}"`, "-1e0004", object.STRING_OBJ},

		{`val x = -1234; "{{x:+1.3e+2}}"`, "-1.234e+03", object.STRING_OBJ},
		{`val x = -10000; "{{x:+1.3e+2}}"`, "-1.000e+04", object.STRING_OBJ},
		{`val x = -1234; "{{x:+1.3-e+4}}"`, "-1.234e+0003", object.STRING_OBJ},
		{`val x = -10000; "{{x:+1.3-e+4}}"`, "-1e+0004", object.STRING_OBJ},

		{`val x = -1000; "{{x:1.2-e+4}}"`, "-1e+0003", object.STRING_OBJ},
		{`val x = -100; "{{x:1.2-e+4}}"`, "-1e+0002", object.STRING_OBJ},
		{`val x = -10; "{{x:1.2-e+4}}"`, "-1e+0001", object.STRING_OBJ},
		{`val x = -1; "{{x:1.2-e+4}}"`, "-1e+0000", object.STRING_OBJ},

		// add zeroes to scale?
		{`val x = 1; "{{x:1.2!e+4}}"`, "1e+0000", object.STRING_OBJ},
		{`val x = 1; "{{x:1.2e+4}}"`, "1.00e+0000", object.STRING_OBJ},
		{`val x = 100; "{{x:1.2!e+4}}"`, "1.00e+0002", object.STRING_OBJ},
		{`val x = 1000; "{{x:1.2!e+4}}"`, "1.00e+0003", object.STRING_OBJ},
		{`val x = 10; "{{x:1.2!e+4}}"`, "1.0e+0001", object.STRING_OBJ},
		{`val x = 1.2; "{{x:1.3e}}"`, "1.200e0", object.STRING_OBJ},

		// comma for decimal point
		{`val x = 10; "{{x:1,2!e+4}}"`, "1,0e+0001", object.STRING_OBJ},
		{`val x = 1.2; "{{x:1,3e}}"`, "1,200e0", object.STRING_OBJ},
		{`val x = 1; "{{x:1,2!e+4}}"`, "1e+0000", object.STRING_OBJ},
	}

	runVmTests(t, tests, false, false)
}

func TestInterpolationModifierForDateTime(t *testing.T) {
	tests := []vmTestCase{
		// using format string directly
		{`val x = dt/2020-03-13 10:00/; "{{x:dt(2006)}}"`, "2020", object.STRING_OBJ},
		{`val x = dt/2020-03-13 10:00/; "{{x:dt(2006 in Jan)}}"`, "2020 in Mar", object.STRING_OBJ},
		{`val x = dt/2020-03-13 10:00/; "{{x:dt(03:04:05)}}"`, "10:00:00", object.STRING_OBJ},
		{`val x = dt/2020-03-13 10:00/; "{{x:dt(yo(Jan))}}"`, "yo(Mar)", object.STRING_OBJ},
		{`val x = dt/2020-03-13 10:00/; "{{x:dt(yo 03, 2006 \x28 )}}"`, "yo 10, 2020 ( ", object.STRING_OBJ},

		// using variable
		{`val format = "2006"
			val x = dt/2020-03-13 10:00/; "{{x:dt format}}"`, "2020", object.STRING_OBJ},
		{`val format = "2006 in Jan"
			val x = dt/2020-03-13 10:00/; "{{x:dt format}}"`, "2020 in Mar", object.STRING_OBJ},
		{`val format = "03:04:05"
			val x = dt/2020-03-13 10:00/; "{{x:dt format}}"`, "10:00:00", object.STRING_OBJ},
		{`val format = "yo(Jan)"
			val x = dt/2020-03-13 10:00/; "{{x:dt format}}"`, "yo(Mar)", object.STRING_OBJ},
		{`val format = "yo 03, 2006 ( "
			val x = dt/2020-03-13 10:00/; "{{x:dt format}}"`, "yo 10, 2020 ( ", object.STRING_OBJ},
	}

	runVmTests(t, tests, false, false)
}

func TestListLiterals(t *testing.T) {
	tests := []vmTestCase{
		{"[]", []int{}, object.LIST_OBJ},
		{"[1, 2, 3]", []int{1, 2, 3}, object.LIST_OBJ},
		{"[1 + 2, 3 * 4, 5 - 6, 7]", []int{3, 12, -1, 7}, object.LIST_OBJ},
		{`["A", "B", "Z"]`, []string{"A", "B", "Z"}, object.LIST_OBJ},

		// free word list, a semantic convenience
		{`fw/1 2 3 abc you know/ == ["1", "2", "3", "abc", "you", "know"]`, true, object.BOOLEAN_OBJ},
		{`fw/1 2 3 abc you\x20know/ == ["1", "2", "3", "abc", "you know"]`, true, object.BOOLEAN_OBJ},
		{`FW/1 2 3 abc you\x20know/ == ["1", "2", "3", "abc", "you\\x20know"]`, true, object.BOOLEAN_OBJ},

		// free word list as blockquote
		{`fw:block asdf
		1 2 3 abc
		you know
		asdf`, []string{"1", "2", "3", "abc", "you", "know"}, object.LIST_OBJ},

		// free word list as blockquote used in list
		{`val x = ["joe",
			fw:block asdf
				1 2 3 abc
				you know
			asdf,
			"jane"]
			x[2][5]`, "you", object.STRING_OBJ},

		// free word list with line returns
		{`val x = fw[
			1 2 3 abc
			you know]
		x == ["1", "2", "3", "abc", "you", "know"]`, true, object.BOOLEAN_OBJ},

		// append
		{"[1, 2] ~ [3]", []int{1, 2, 3}, object.LIST_OBJ},
		{"[1, 2] ~ [3, 4]", []int{1, 2, 3, 4}, object.LIST_OBJ},
	}

	runVmTests(t, tests, false, false)
}

func TestListMultiplication(t *testing.T) {
	tests := []vmTestCase{
		{"[] * -1", []int{}, object.LIST_OBJ},
		{"[] * 7", []int{}, object.LIST_OBJ},
		{"[1, 2] * -1", []int{}, object.LIST_OBJ},
		{"[1, 2] * 0", []int{}, object.LIST_OBJ},
		{"[1, 2] * 1", []int{1, 2}, object.LIST_OBJ},
		{"[1, 2] * 2", []int{1, 2, 1, 2}, object.LIST_OBJ},
		{"[1, 2] * 3", []int{1, 2, 1, 2, 1, 2}, object.LIST_OBJ},

		{"-1 * []", []int{}, object.LIST_OBJ},
		{"7 * []", []int{}, object.LIST_OBJ},
		{"-1 * [1, 2]", []int{}, object.LIST_OBJ},
		{"0 * [1, 2]", []int{}, object.LIST_OBJ},
		{"1 * [1, 2]", []int{1, 2}, object.LIST_OBJ},
		{"2 * [1, 2]", []int{1, 2, 1, 2}, object.LIST_OBJ},
		{"3 * [1, 2]", []int{1, 2, 1, 2, 1, 2}, object.LIST_OBJ},

		// check that values stick
		{`var x = [1, 2]
		  var y = x * 3
		  x[1] = 7
		  y`,
			[]int{1, 2, 1, 2, 1, 2}, object.LIST_OBJ,
		},
		{`var x = [[1], 2]
		  var y = x * 3
		  x[1] = 7
		  string(y)`,
			"[[1], 2, [1], 2, [1], 2]", object.STRING_OBJ,
		},
	}

	runVmTests(t, tests, false, false)
}

func TestHashLiterals(t *testing.T) {
	tests := []vmTestCase{
		{"{:}", [][]object.Object{}, object.HASH_OBJ},

		{"{1: 2, 2: 7}",
			[][]object.Object{
				{object.NumberFromInt(1), object.NumberFromInt(2)},
				{object.NumberFromInt(2), object.NumberFromInt(7)},
			},
			object.HASH_OBJ,
		},

		{"{1.0: 2, 2.0: 7}",
			[][]object.Object{
				{object.NumberFromInt(1), object.NumberFromInt(2)},
				{object.NumberFromInt(2), object.NumberFromInt(7)},
			},
			object.HASH_OBJ,
		},

		// 1.0 same as 1
		{"{1: 2, 2: 7} ~ {1.0: -7, 3.0: 14}",
			[][]object.Object{
				{object.NumberFromInt(1), object.NumberFromInt(-7)},
				{object.NumberFromInt(2), object.NumberFromInt(7)},
				{object.NumberFromInt(3), object.NumberFromInt(14)},
			},
			object.HASH_OBJ,
		},

		{`{1 + 1: 2 \ 2, 7: 7 * 2}`,
			[][]object.Object{
				{object.NumberFromInt(2), object.NumberFromInt(1)},
				{object.NumberFromInt(7), object.NumberFromInt(14)},
			},
			object.HASH_OBJ,
		},

		// hash function
		{"hash([1, 2, 3, 7])",
			[][]object.Object{
				{object.NumberFromInt(1), object.NumberFromInt(2)},
				{object.NumberFromInt(3), object.NumberFromInt(7)},
			},
			object.HASH_OBJ,
		},

		{"hash([1, 2], [3, 7])",
			[][]object.Object{
				{object.NumberFromInt(1), object.NumberFromInt(3)},
				{object.NumberFromInt(2), object.NumberFromInt(7)},
			},
			object.HASH_OBJ,
		},

		{"hash(1..3, 5..7)",
			[][]object.Object{
				{object.NumberFromInt(1), object.NumberFromInt(5)},
				{object.NumberFromInt(2), object.NumberFromInt(6)},
				{object.NumberFromInt(3), object.NumberFromInt(7)},
			},
			object.HASH_OBJ,
		},

		{"hash(1..4)",
			[][]object.Object{
				{object.NumberFromInt(1), object.NumberFromInt(2)},
				{object.NumberFromInt(3), object.NumberFromInt(4)},
			},
			object.HASH_OBJ,
		},

		{"hash({1: 123})",
			[][]object.Object{
				{object.NumberFromInt(1), object.NumberFromInt(123)},
			},
			object.HASH_OBJ,
		},
	}

	runVmTests(t, tests, false, false)
}

func TestAppend(t *testing.T) {
	tests := []vmTestCase{
		{`97 .. 99 ~ "go"`, "abcgo", object.STRING_OBJ},
		{`97 .. 99 ~ 101`, "abce", object.STRING_OBJ},
		{`97 .. 99 ~ 100 .. 99`, "abcdc", object.STRING_OBJ},

		{`97 ~ 99`, "ac", object.STRING_OBJ},
		{`97 ~ "joe"`, "ajoe", object.STRING_OBJ},
		{`97 ~ 100 .. 98`, "adcb", object.STRING_OBJ},

		{`"A" ~ 99`, "Ac", object.STRING_OBJ},
		{`"A" ~ "joe"`, "Ajoe", object.STRING_OBJ},
		{`"A" ~ 100 .. 98`, "Adcb", object.STRING_OBJ},
	}

	runVmTests(t, tests, false, false)
}

func TestAppendNull(t *testing.T) {
	// acts as no-op, sort of
	tests := []vmTestCase{
		{`"abc" ~ null`, "abc", object.STRING_OBJ},
		{`97 ~ null`, "a", object.STRING_OBJ},
		{`[1, 2, 3] ~ null`, []int{1, 2, 3}, object.LIST_OBJ},
		{`{1: 10, 2: 40} ~ null`,
			[][]object.Object{
				{object.NumberFromInt(1), object.NumberFromInt(10)},
				{object.NumberFromInt(2), object.NumberFromInt(40)},
			},
			object.HASH_OBJ,
		},
	}

	runVmTests(t, tests, false, false)
}

func TestAppendToNull(t *testing.T) {
	// returns right object for certain types
	tests := []vmTestCase{
		{`null ~ null`, nil, object.NULL_OBJ},
		{`null ~ []`, []int{}, object.LIST_OBJ},
		{`null ~ [1, 3]`, []int{1, 3}, object.LIST_OBJ},
		{`null ~ {1: 3, 2: 7}`,
			[][]object.Object{
				{object.NumberFromInt(1), object.NumberFromInt(3)},
				{object.NumberFromInt(2), object.NumberFromInt(7)},
			},
			object.HASH_OBJ,
		},
		{`null ~ {:}`,
			[][]object.Object{},
			object.HASH_OBJ,
		},
		{`null ~ ""`, "", object.STRING_OBJ},
		{`null ~ "omega"`, "omega", object.STRING_OBJ},
		{`null ~ 65`, "A", object.STRING_OBJ},
		{`null ~ 97 ~ 98 ~ 99`, "abc", object.STRING_OBJ},

		{`97 ~ null ~ 98 ~ 99`, "abc", object.STRING_OBJ},

		{`null ~ 97 .. 99`, "abc", object.STRING_OBJ},
	}

	runVmTests(t, tests, false, false)
}

func TestIndexExpressions(t *testing.T) {
	// 1-based indexing on lists and strings
	// ranges inclusive
	tests := []vmTestCase{
		// positive indices on lists
		{"[12, 14, 17][2]", "14", object.NUMBER_OBJ},
		{"[12, 14, 17][7; null]", nil, object.NULL_OBJ},
		{"[12, 14, 17][1 + 1; null]", "14", object.NUMBER_OBJ},
		{"[[3, 9, 21]][1][2]", "9", object.NUMBER_OBJ},
		{"[7, 14, 21, 28, 35, 42, 49][7]", "49", object.NUMBER_OBJ},
		{"[][1; null]", nil, object.NULL_OBJ},
		{"[7, 14, 21, 28, 35, 42, 49][3 .. 1]", []int{21, 14, 7}, object.LIST_OBJ},
		{"[7, 14, 21, 28, 35, 42, 49][1 .. 3]", []int{7, 14, 21}, object.LIST_OBJ},
		{"[7, 14, 21, 28, 35, 42, 49][2 .. 3]", []int{14, 21}, object.LIST_OBJ},
		{"[1, 2, 3][7; 3 * 4]", "12", object.NUMBER_OBJ},

		// negative indices on lists
		{"[12, 14, 17][-2]", "14", object.NUMBER_OBJ},
		{"[12, 14, 17][-7; null]", nil, object.NULL_OBJ},
		{"[12, 14, 17][1 - 2; null]", 17, object.NUMBER_OBJ},
		{"[12, 14, 17][1 - 22; null]", nil, object.NULL_OBJ},
		{"[[3, 9, 21]][-1][-1]", 21, object.NUMBER_OBJ},
		{"[7, 14, 21, 28, 35, 42, 49][-7]", 7, object.NUMBER_OBJ},
		{"[7, 14, 21, 28, 35, 42, 49][-1]", 49, object.NUMBER_OBJ},
		{"[7, 14, 21, 28, 35, 42, 49][-3 .. -1]", []int{35, 42, 49}, object.LIST_OBJ},
		{"[7, 14, 21, 28, 35, 42, 49][-1 .. -3]", []int{49, 42, 35}, object.LIST_OBJ},
		{"[7, 14, 21, 28, 35, 42, 49][-4 .. -2]", []int{28, 35, 42}, object.LIST_OBJ},
		{"[7, 14, 21, 28, 35, 42, 49][[-4 .. -2, -1, -7, 2]]", []int{28, 35, 42, 49, 7, 14}, object.LIST_OBJ},

		// indices on hashes
		{"{1: 2, 3: 4}[3; null]", "4", object.NUMBER_OBJ},
		{"{1: 2, 3: 4}[7; null]", nil, object.NULL_OBJ},
		{"{:}[1; null]", nil, object.NULL_OBJ},

		{`{"yo": 2, "joe": 4}["joe"]`, "4", object.NUMBER_OBJ},
		{`{"yo": 2, "joe": 4}["yo"]`, "2", object.NUMBER_OBJ},
		{`{"yo": 2, "joe": 4}["jo"; 7]`, "7", object.NUMBER_OBJ},
		{`{"yo": 2, "joe": 4}["joe"; 7]`, "4", object.NUMBER_OBJ},

		{"{1: 2, 3: 4}[[1, 3]; 0]", []int{2, 4}, object.LIST_OBJ},
		{"{1: 2, 3: 4}[[1, 4]; 0]", 0, object.NUMBER_OBJ},

		// indices on hashes using range
		{"{1: 34, 2: 44, 3: 5, 4: 17}[1..2; 0]", []int{34, 44}, object.LIST_OBJ},
		{"{1: 34, 2: 44, 3: 5, 4: 17}[2..3; 0]", []int{44, 5}, object.LIST_OBJ},
		{"{1: 34, 2: 44, 3: 5, 4: 17}[3..2; 0]", []int{5, 44}, object.LIST_OBJ},
		{"{1: 34, 2: 44, 3: 5, 4: 17}[3..7; 0]", 0, object.NUMBER_OBJ},

		{"{-1: 34, 0: 44, 1: 5, 2: 17}[1..2; 0]", []int{5, 17}, object.LIST_OBJ},
		{"{-1: 34, 0: 44, 1: 5, 2: 17}[-1..1; 0]", []int{34, 44, 5}, object.LIST_OBJ},
		{"{-1: 34, 0: 44, 1: 5, 2: 17}[-1..3; 0]", 0, object.NUMBER_OBJ},

		// short-hand string indexing (on hashes)
		{`val x = {"yo": 2, "joe": 4}; x'yo`, 2, object.NUMBER_OBJ},
		{`val x = {"yo": 2, "joe": 4}; x'joe`, 4, object.NUMBER_OBJ},
		{`val x = {"yo": 2, "joe": 4, "123": 7}; x'123`, 7, object.NUMBER_OBJ},

		// positive indices on strings
		{`"abc"[1]`, "97", object.NUMBER_OBJ},
		{`"abc"[3; null]`, "99", object.NUMBER_OBJ},
		{`"abc"[4; null]`, nil, object.NULL_OBJ},

		{`"abc yoyo"[5 .. 8]`, []int{121, 111, 121, 111}, object.LIST_OBJ},
		{`"abc yoyo"[8 .. 5]`, []int{111, 121, 111, 121}, object.LIST_OBJ},
		{`"abc yoyo"[5 .. 9; "nothing"]`, "nothing", object.STRING_OBJ},
		{`"abc"[[1, 3]]`, []int{97, 99}, object.LIST_OBJ},

		{`"abc"[1]`, "97", object.NUMBER_OBJ},
		{`"abc"[3; null]`, "99", object.NUMBER_OBJ},
		{`"abc"[4; null]`, nil, object.NULL_OBJ},

		// negative indices on strings
		{`"abc"[-1]`, "99", object.NUMBER_OBJ},
		{`"abc"[-3; null]`, "97", object.NUMBER_OBJ},
		{`"abc"[-4; null]`, nil, object.NULL_OBJ},

		{`"abc yoyo"[-5 .. -8]`, []int{32, 99, 98, 97}, object.LIST_OBJ},
		{`"abc yoyo"[-8 .. -5]`, []int{97, 98, 99, 32}, object.LIST_OBJ},
		{`"abc yoyo"[-5 .. -9; "nothing"]`, "nothing", object.STRING_OBJ},
		{`"abc"[[-1, 3]]`, []int{99, 99}, object.LIST_OBJ},

		{`"abc"[-1]`, "99", object.NUMBER_OBJ},
		{`"abc"[-3; null]`, "97", object.NUMBER_OBJ},
		{`"abc"[-4; null]`, nil, object.NULL_OBJ},

		{`('a'..'z')[1]`, 97, object.NUMBER_OBJ},
		{`('a'..'z')[2]`, 122, object.NUMBER_OBJ},
		{`('a'..'z')[1; null]`, 97, object.NUMBER_OBJ},
		{`('a'..'z')[7; null]`, nil, object.NULL_OBJ},

		{`(7+21i)[1]`, 7, object.NUMBER_OBJ},
		{`(7+21i)[2]`, 21, object.NUMBER_OBJ},
		{`(7+21i)[1; 4]`, 7, object.NUMBER_OBJ},
		{`(7+21i)[3; 4]`, 4, object.NUMBER_OBJ},

		{`val conjugate = fn(c) { complex(c[1], -c[2]) }
			conjugate(1+1i)`, "1-1i", object.COMPLEX_OBJ},

		// alternate index
		{`{1: 2, 3: 4}[3; "123"]`, "4", object.NUMBER_OBJ},
		{"[12, 14, 17][7; 21]", "21", object.NUMBER_OBJ},
		{`[][1; "abc"]`, "abc", object.STRING_OBJ},

		{"[7, 14, 21, 28, 35, 42, 49][1 .. 3; 77]", []int{7, 14, 21}, object.LIST_OBJ},
		{"[7, 14, 21, 28, 35, 42, 49][1 .. 30; 77]", "77", object.NUMBER_OBJ},

		{"{1: 2, 3: 4}[1; 7]", "2", object.NUMBER_OBJ},
		{"{1: 2, 3: 4}[10; 3 + 4]", "7", object.NUMBER_OBJ},
		{"{:}[1; 7]", "7", object.NUMBER_OBJ},

		{`"abc"[3; 123]`, "99", object.NUMBER_OBJ},
		{`"abc"[4; 123]`, "123", object.NUMBER_OBJ},
		{`cp2s("abc yoyo"[5 .. 8; "nothing"])`, "yoyo", object.STRING_OBJ},
		{`"abc yoyo"[5 .. 9; "nothing"]`, "nothing", object.STRING_OBJ},

		// alternate index on things not indexable
		{`null[3; 7]`, 7, object.NUMBER_OBJ},
		{`1234[3; 7]; catch { 10 }`, 10, object.NUMBER_OBJ},

		// alternate index with bad index types
		{`[][{:}; 7]`, 7, object.NUMBER_OBJ},
		{`[][dt//; 7]`, 7, object.NUMBER_OBJ},
	}

	runVmTests(t, tests, false, false)
}

func TestIndexingWithLists(t *testing.T) {
	tests := []vmTestCase{
		// indexing lists with lists
		{"[7, 14, 21, 28, 35, 42, 49][[1]]", []int{7}, object.LIST_OBJ},
		{"[7, 14, 21, 28, 35, 42, 49][[1..3]]", []int{7, 14, 21}, object.LIST_OBJ},
		{"[7, 14, 21, 28, 35, 42, 49][[2, 4, 6]]", []int{14, 28, 42}, object.LIST_OBJ},
		{"[7, 14, 21, 28, 35, 42, 49][[2, 4, 6..7]]", []int{14, 28, 42, 49}, object.LIST_OBJ},

		{"[7, 14, 21, 28, 35, 42, 49][[1]; 777]", []int{7}, object.LIST_OBJ},
		{"[7, 14, 21, 28, 35, 42, 49][[1..3]; 777]", []int{7, 14, 21}, object.LIST_OBJ},
		{"[7, 14, 21, 28, 35, 42, 49][[8]; 777]", "777", object.NUMBER_OBJ},
		{"[7, 14, 21, 28, 35, 42, 49][[1..8]; 777]", "777", object.NUMBER_OBJ},

		{"[7, 14, 21, 28, 35, 42, 49][[2, 4, [[1]], 6..7]]", []int{14, 28, 7, 42, 49}, object.LIST_OBJ},

		// indexing strings with lists
		{`"abc"[[1, 3]]`, []int{97, 99}, object.LIST_OBJ},
		{`"abcd yoyo"[[1, 3, 7..8]]`, []int{97, 99, 111, 121}, object.LIST_OBJ},

		{`"abc"[[1, 3]; "no!"]`, []int{97, 99}, object.LIST_OBJ},
		{`"abcd yoyo"[[1, 3, 8..9]; "no!"]`, []int{97, 99, 121, 111}, object.LIST_OBJ},
		{`"abc"[[1, 10]; "no!"]`, "no!", object.STRING_OBJ},
		{`"abcd yoyo"[[1, 3, 7..10]; "no!"]`, "no!", object.STRING_OBJ},

		// indexing hashes with lists
		{"{1: 2, 7: 3, 14: 4}[[1, 14]]", []int{2, 4}, object.LIST_OBJ},
		{"{1: 2, 7: 3, 14: 4}[[1, 14, 7]]", []int{2, 4, 3}, object.LIST_OBJ},

		{"{1: 2, 7: 3, 14: 4}[[1, 14]; []]", []int{2, 4}, object.LIST_OBJ},
		{"{1: 2, 7: 3, 14: 4}[[1, 14, 7]; []]", []int{2, 4, 3}, object.LIST_OBJ},
		{"{1: 2, 7: 3, 14: 4}[[1, 14, 21]; []]", []int{}, object.LIST_OBJ},
		{"{1: 2, 7: 3, 14: 4}[[1, 14, 7, 21]; []]", []int{}, object.LIST_OBJ},

		// indexing ranges with lists
		// necessary?
		{"(7..40)[[1, 2]]", []int{7, 40}, object.LIST_OBJ},
		{"(7..40)[[2, 1]]", []int{40, 7}, object.LIST_OBJ},
		{"(7..40)[[2, 1]; []]", []int{40, 7}, object.LIST_OBJ},
		{"(7..40)[[4, 1]; []]", []int{}, object.LIST_OBJ},

		// indexing complexes with lists
		// necessary?
		{"(7+40i)[[1, 2]]", []int{7, 40}, object.LIST_OBJ},
		{"(7+40i)[[2, 1]]", []int{40, 7}, object.LIST_OBJ},
		{"(7+40i)[[2, 1]; []]", []int{40, 7}, object.LIST_OBJ},
		{"(7+40i)[[4, 1]; []]", []int{}, object.LIST_OBJ},
	}

	runVmTests(t, tests, false, false)
}

func TestLoops(t *testing.T) {
	tests := []vmTestCase{
		// test only
		{`for ; 0 > 1; {}`, nil, object.NULL_OBJ},

		{`var sum = 0
		  for ; sum < 100; { sum += 7 }
		  sum`,
			"105", object.NUMBER_OBJ,
		},

		// 1 part
		// 0.8 uses a while keyword for a test only loop without semicolons instead of the for keyword.
		{`while 0 > 1 {}`, nil, object.NULL_OBJ},

		{`var sum = 0
		  while sum < 100 { sum = sum + 7 }
		  sum`,
			"105", object.NUMBER_OBJ,
		},

		// 3 part
		{`var thelast = 0
		  for x = 1; x <= 14; x += 1 { thelast = x }
		  thelast`,
			"14", object.NUMBER_OBJ,
		},

		{`# factorial
		  var answer = 1
		  for x = 1; x <= 7; x = x + 1 {
		  	answer *= x
		  }
		  answer`,
			"5040", object.NUMBER_OBJ,
		},

		// ... with multiple initializations and increments
		{`var sum = 0
		  for i, j = 1, 1; i < 10; i, j = i + 3, j + i + 3 { sum += j }
		  sum`,
			"18", object.NUMBER_OBJ,
		},
	}

	runVmTests(t, tests, false, false)
}

func TestForInLoops(t *testing.T) {
	tests := []vmTestCase{
		// over list
		{`var sum = 0
		  for x in [7, 3, 11] { sum += x }
		  sum`,
			"21", object.NUMBER_OBJ,
		},
		{`# factorial
		  var answer = 1
		  for x in series(1..7) {
		  	answer *= x
		  }
		  answer`,
			"5040", object.NUMBER_OBJ,
		},

		// over hash
		{`var sum = 0
		  for x in {1: 10, 2: 20, 3: 40} { sum += x }
		  sum`,
			"70", object.NUMBER_OBJ,
		},

		// over hash variable
		{`var sum = 0
		  var h = {1: 10, 2: 20, 3: 40}
		  for x in h { sum += x }
		  sum`,
			"70", object.NUMBER_OBJ,
		},

		// over string
		{`var new = ""		# 0; 1; just thrown in for good measure
		  for x in "abc" { 0; 1; new ~= ucase(cp2s(x)) ~ "~" }
		  new`,
			"A~B~C~", object.STRING_OBJ,
		},

		// over range
		{
			`var sum = 0
			for x in 1..4 {
				sum += x
			}
			sum`,
			"10", object.NUMBER_OBJ,
		},
		{
			`var sum = 0
			for x in 4..1 {
				sum += x
			}
			sum`,
			"10", object.NUMBER_OBJ,
		},

		// over range with giant numbers
		{
			`var sum = 0
			for x in 9223372036854775809 .. 9223372036854775810 {
				sum += x
			}
			sum`,
			"18446744073709551619", object.NUMBER_OBJ,
		},
		{
			`var sum = 0
			for x in 9223372036854775810 .. 9223372036854775809 {
				sum += x
			}
			sum`,
			"18446744073709551619", object.NUMBER_OBJ,
		},

		// over number (implicit range)
		{
			`var sum = 0
			for x in 4 {
				sum += x
			}
			sum`,
			"10", object.NUMBER_OBJ,
		},
		{
			`var sum = 0
			for x in -4 {
				sum += x
			}
			sum`,
			"-10", object.NUMBER_OBJ,
		},
	}

	runVmTests(t, tests, false, false)
}

func TestForOfLoops(t *testing.T) {
	tests := []vmTestCase{
		// over list
		{`# strange factorial (on index number)
		  var answer = 1
		  for x of [7, 6, 5, 4, 3, 14, 21] {
		  	answer *= x
		  }
		  answer`,
			"5040", object.NUMBER_OBJ,
		},

		// over list variable
		{`var sum = 0; var z = [7, 3, 11]
		  for x of z { sum = sum + z[x] }
		  sum`,
			"21", object.NUMBER_OBJ,
		},

		// over hash
		{`var sum = 0
		  for x of {1: 10, 2: 20, 3: 40} { sum += x }
		  sum`,
			"6", object.NUMBER_OBJ,
		},

		// over hash variable
		{`var sum = 0
		  var h = {1: 10, 2: 20, 3: 40}
		  for x of h { sum += x }
		  sum`,
			"6", object.NUMBER_OBJ,
		},

		// over string
		{`# strange factorial (on index number)
		  var answer = 1
		  for x of "abcdefg" {
		  	 answer *= x
		  }
		  answer`,
			"5040", object.NUMBER_OBJ,
		},

		// over range
		{`var sum = 0
		  for i of 1..7 { sum += 10 }
		  sum`,
			"70", object.NUMBER_OBJ,
		},

		// over range variable
		{`var rng = 1..7
		  var sum = 0
		  for i of rng { sum += 10 }
		  sum`,
			"70", object.NUMBER_OBJ,
		},

		// over descending range variable
		{`var rng = 7..1
		  var sum = 0
		  for i of rng { sum += 10 }
		  sum`,
			"0", object.NUMBER_OBJ,
		},

		// over range with giant numbers
		{
			`var sum = 0
			for x of 9223372036854775809 .. 9223372036854775810 {
				sum += x
			}
			sum`,
			"3", object.NUMBER_OBJ,
		},
		{
			`var sum = 0
			for x of 9223372036854775810 .. 9223372036854775809 {
				sum += x
			}
			sum`,
			"3", object.NUMBER_OBJ,
		},

		// over number variable (implicit range)
		{`var rng = 7
		  var sum = 0
		  for i of rng { sum += 10 }
		  sum`,
			"70", object.NUMBER_OBJ,
		},

		// over number (implicit range)
		{`var sum = 0
		  for i of 7 { sum += 10 }
		  sum`,
			"70", object.NUMBER_OBJ,
		},

		// over negative number
		{`var sum = 0
		  for i of -7 { sum += 10 }
		  sum`,
			"0", object.NUMBER_OBJ,
		},
	}

	runVmTests(t, tests, false, false)
}

func TestForOfLoopWithoutVariable(t *testing.T) {
	tests := []vmTestCase{
		// over range
		{`var sum = 0
		  for of 1..7 { sum += 10 }
		  sum`,
			"70", object.NUMBER_OBJ,
		},

		// over implicit range
		{`var sum = 0
		  for of 7 { sum += 10 }
		  sum`,
			"70", object.NUMBER_OBJ,
		},
		{`var sum = 0
		  for of 0 { sum += 10 }
		  sum`,
			"0", object.NUMBER_OBJ,
		},
		{`var sum = 0
		  for of 1 { sum += 10 }
		  sum`,
			"10", object.NUMBER_OBJ,
		},

		// over variable
		{`var sum = 0
		  var i = 1..7
		  for of i { sum += 10 }
		  sum`,
			"70", object.NUMBER_OBJ,
		},
		{`var sum = 0
		  var i = 7
		  for of i { sum += 10 }
		  sum`,
			"70", object.NUMBER_OBJ,
		},
		{`var sum = 0
		  var i = series(1..7)
		  for of i { sum += 10 }
		  sum`,
			"70", object.NUMBER_OBJ,
		},
		{`var sum = 0
		  var i = "1234567"
		  for of i { sum += 10 }
		  sum`,
			"70", object.NUMBER_OBJ,
		},
		{`var sum = 0
		  var i = -7
		  for of i { sum += 10 }
		  sum`,
			"0", object.NUMBER_OBJ,
		},
	}

	runVmTests(t, tests, false, false)
}

func TestForLoopMisc(t *testing.T) {
	tests := []vmTestCase{
		// embedded for loops
		{`var sum = 0
		  val a = [2, 3, 4]
		  val b = [7, 8, 9]
		
		  for x in a {
		     for y in b {
				sum += x + y
			 }
		  }
		  sum`,
			"99", object.NUMBER_OBJ,
		},

		// return out of for loop
		{`val test = fn(a) { for x of a {
		     if x > 1 {
				return x
			 }
		  }}
		  test([7, 14, 21, 28])`,
			"2", object.NUMBER_OBJ,
		},

		// return out of 2 for loops
		{`val a = [2, 3, 4]
		  val b = [7, 8, 9]
		
		  val test = fn() {
			 var sum = 0
			 for x in a {
		       for y in b {
				  sum += x + y
				  if sum > 60 {
					 return sum
				  }
			   }
		     }
		  }
		  test()`,
			"63", object.NUMBER_OBJ,
		},

		// catch in a for loop
		{`var alist = []
		  for x in series(-3 .. 3) {
			 var calc = 1 / x
			 catch { calc = 0 }
			 alist = more(alist, calc)
		  }
		  alist[4]
		  `,
			"0", object.NUMBER_OBJ,
		},

		// for loop that never runs
		{`var sum = 1
		 for i = 1; i < 1; i = i + 1 {
			sum += i
		 }
		 sum`,
			"1", object.NUMBER_OBJ,
		},
	}

	runVmTests(t, tests, false, false)
}

func TestForLoopBreak(t *testing.T) {
	tests := []vmTestCase{
		{`var sum = 0
		  for i = 0; i < 10; i = i + 1 {
		     if i > 4 {
				break
			 }
			 sum = sum + 7
		  }
		  sum`,
			"35", object.NUMBER_OBJ,
		},

		{`var sum = 0
		  for i = 0; i < 10; i = i + 1 {
		     if i > 4 : break
			 sum = sum + 7
		  }
		  sum`,
			"35", object.NUMBER_OBJ,
		},

		{`var sum = 0
		  for x in 1..3 {
			  for i = 0; i < 10; i = i + 1 {
			     if i > 4 : break
				 sum = sum + 7
			  }
		  }
		  sum`,
			"105", object.NUMBER_OBJ,
		},

		{`var sum = 0
		  for x in 1..3 {
			  if x > 2: break
			  for i = 0; i < 10; i = i + 1 {
			     if i > 4: break
				 sum = sum + 7
			  }
		  }
		  sum`,
			"70", object.NUMBER_OBJ,
		},

		{`var sum = 0
		  for x in 1..3 {
			  if x > 2: break
			  for i = 0; i < 10; i = i + 1 {
				 sum = sum + 7
			  }
		  }
		  sum`,
			"140", object.NUMBER_OBJ,
		},

		// break embedded in scope...
		{`var sum = 0
		  for x = 1; x <= 10; x += 1 {
			  {  var y = 7
			     if x > 2: break
			  }
			  sum += 70
		  }
		  sum`,
			"140", object.NUMBER_OBJ,
		},
		// break deeply embedded in scope...
		{`var sum = 0
		  for x in 1..10 {
			  {  var a = 123
				  {  var z = 0
					  {  var y = 7
					     if x > 2 {
						    break
				 		 }
					  }
				  }
			  }
			  for i = 0; i < 10; i += 1 {
				 sum += 7
			  }
		  }
		  sum`,
			"140", object.NUMBER_OBJ,
		},
		// break embedded in try
		{`var sum = -100
		  for x = 0; x < 10; x += 1 {
		      if x > 0: break
			  sum += 7
			  catch {}
		  }
		  sum`,
			"-93", object.NUMBER_OBJ,
		},
		// break embedded in catch
		{`var sum = 0
		  for x = -3; x < 10; x += 1 {
		      sum += 1 / (x rem 2)
			
				/*  x: x rem 2	sum
					-3: -1			-1
				*/
			  catch { break }
		  }
		  sum`,
			"-1", object.NUMBER_OBJ,
		},
		// break embedded in catch else
		{`var sum = 0
		  for x = -3; x < 10; x += 1 {
		      sum += 1 / (x rem 2)
			
				/*  x: x rem 2	sum
					-3: -1			-1
				*/
			  catch { 50 } else { break }
		  }
		  sum`,
			"-1", object.NUMBER_OBJ,
		},

		// break with value embedded in catch
		{`var sum = 0
		  for x = -3; x < 10; x += 1 {
		      sum += 1 / (x rem 2)
			  catch { break val=144 }
		  }
		  `,
			"144", object.NUMBER_OBJ,
		},
		// break with value embedded in catch else
		{`var sum = 0
		  for x = -3; x < 10; x += 1 {
		      sum += 1 / (x rem 2)
			  catch { 50 } else { break val=124 }
		  }
		  `,
			"124", object.NUMBER_OBJ,
		},
	}

	runVmTests(t, tests, false, false)
}

func TestForLoopNext(t *testing.T) {
	tests := []vmTestCase{
		{`var sum = 0
		  for i = 0; i < 10; i = i + 1 {
		     if i rem 2 != 0: next
			 sum = sum + 7
		  }
		  sum`,
			"35", object.NUMBER_OBJ,
		},

		{`var sum = 0
		  for x in 1..3 {
			  for i = 0; i < 10; i = i + 1 {
			     if i rem 2 != 0 {
					next
				 }
				 sum = sum + 7
			  }
		  }
		  sum`,
			"105", object.NUMBER_OBJ,
		},

		{`var sum = 0
		  for x in 1..3 {
			  if x > 2 {
				 next
	  		  }
			  for i = 0; i < 10; i = i + 1 {
			     if i rem 2 != 0: next
				 sum = sum + 7
			  }
		  }
		  sum`,
			"70", object.NUMBER_OBJ,
		},

		{`var sum = 0
		  for x in 1..3 {
			  if x > 2 {
				 next
	  		  }
			  for i = 0; i < 10; i = i + 1 {
				 sum = sum + 7
			  }
		  }
		  sum`,
			"140", object.NUMBER_OBJ,
		},

		// next embedded in scope...
		{`var sum = 0
		  for x = 1; x <= 10; x += 1 {
			  {  var y = 7
			     if x > 2 {
				    next
		 		 }
			  }
			  sum += 70
		  }
		  sum`,
			"140", object.NUMBER_OBJ,
		},
		// next deeply embedded in scope...
		{`var sum = 0
		  for x = 1; x <= 10; x += 1 {
			  {  var y = 7
				 { var z = 789
					 { var a = 1
						{ var b = 7
							{ var c
							     if x > 2 {
								    next
						 		 }
							}
						}
					 }
				 }
			  }
			  sum += 70
		  }
		  sum`,
			"140", object.NUMBER_OBJ,
		},
		// next embedded in try
		{`var sum = 0
		  for x = 0; x < 10; x += 1 {
		      if x < 3 { next }
			  sum += 7
			  catch {}
		  }
		  sum`,
			"49", object.NUMBER_OBJ,
		},
		// next embedded in catch
		{`var sum = 0
		  for x = -3; x < 10; x += 1 {
		      sum += 1 / (x rem 2)
			
				/*  x: x rem 2	sum
					-3: -1			-1
					-1: -1			-2
					 1: 1			-1
					 3: 1			0
					 5: 1			1
					 7: 1			2
					 9: 1			3
				*/
			  catch { next } else { 50 }
		  }
		  sum`,
			"3", object.NUMBER_OBJ,
		},
		// next embedded in catch else
		{`var sum = 0
		  for x = -3; x < 11; x += 1 {
			  1 / (x rem 2)
			  catch { 50 } else { next }
		      sum += 1
		  }
		  sum`,
			"7", object.NUMBER_OBJ,
		},
	}

	runVmTests(t, tests, false, false)
}

func TestForLoopBreakAndNext(t *testing.T) {
	tests := []vmTestCase{
		{`var sum = 0
		  for x in 1..3 {
			  if x > 2 {
				 break
			  }
			  for i = 0; i < 10; i = i + 1 {
			     if i > 4 {
					next
				 }
				 sum = sum + 7
			  }
		  }
		  sum`,
			"70", object.NUMBER_OBJ,
		},
		{`var sum = 0
		  for x in 1..3 {
			  if x > 2 {
				 next
			  }
			  for i = 0; i < 10; i = i + 1 {
			     if i > 4 {
					break
				 }
				 sum = sum + 7
			  }
		  }
		  sum`,
			"70", object.NUMBER_OBJ,
		},
	}

	runVmTests(t, tests, false, false)
}

func TestForLoopValue(t *testing.T) {
	tests := []vmTestCase{
		{`for of 10 {}`,
			nil,
			object.NULL_OBJ,
		},
		{`for of 10 { break }`,
			nil,
			object.NULL_OBJ,
		},
		{`for of 10 { 123; break }`,
			nil,
			object.NULL_OBJ,
		},

		{`for of 10 { break val=3.5 * 2 }`,
			"7.0",
			object.NUMBER_OBJ,
		},

		{`for i of 10 {
			 if i > 3 {
			     break val=i * 2
			 }
		  }`,
			"8",
			object.NUMBER_OBJ,
		},

		{`7 + for of 10 { break val=13 }`,
			"20",
			object.NUMBER_OBJ,
		},
		{`7 + for i of 10 { if val x = i > 3 { break val=i } }`,
			"11",
			object.NUMBER_OBJ,
		},
		{`7 + for i of 10 { if i > 3 { val x = i; break val=x } }`,
			"11",
			object.NUMBER_OBJ,
		},
		{`7 + for i of 10 {
			 1 / (i rem 2)
			 catch { } else {
				val x = 3
				if i > 7 {
					val y = 789
					break val=i  # i == 9
				}
			 }
		  }`,
			"16",
			object.NUMBER_OBJ,
		},

		// setting for loop value directly
		{`for i of 7 { _for = if _for { _for += i } else {i} }`,
			28,
			object.NUMBER_OBJ,
		},
		{`for i of 7 { _for = if(_for: if i rem 2 == 0 { _for += i } else {_for}; i) }`,
			13,
			object.NUMBER_OBJ,
		},
		{`for[f] i of 7 { f = if f {f += i} else {i} }`,
			28,
			object.NUMBER_OBJ,
		},
		{`for[f=0] i of 7 { f += i }`,
			28,
			object.NUMBER_OBJ,
		},
		{`for[=0] i of 7 { _for += i }`,
			28,
			object.NUMBER_OBJ,
		},
		{`for[f=0] i = 1; i < 8; i += 1 { f += i; if i > 3 { break } }`,
			10,
			object.NUMBER_OBJ,
		},
		{`for[f=0] i of 7 { f += i; break val=7 }`,
			7,
			object.NUMBER_OBJ,
		},

		{`var i = 0; while[=0] i < 7 { _while+=2; i+=1 }`,
			14,
			object.NUMBER_OBJ,
		},
		{`var i = 0; while[f=0] i < 7 { f+=2; i+=1 }`,
			14,
			object.NUMBER_OBJ,
		},

		// ... with nested for loops
		{`for[=0] i of 7 { _for += for[=0] j of 3 { _for += 7 } }`,
			147,
			object.NUMBER_OBJ,
		},
		{`for[=0] i of 7 { _for += for[f2=0] j of 3 { f2 += 7 } }`,
			147,
			object.NUMBER_OBJ,
		},
		{`for[f=0] i of 7 { f += for[=0] j of 3 { _for += 7 } }`,
			147,
			object.NUMBER_OBJ,
		},
		{`for[f=0] i of 7 { f += for[f2=0] j of 3 { f2 += 7 } }`,
			147,
			object.NUMBER_OBJ,
		},
		{`for[f=0] i of 7 { f += i; for j of 3 { f += 7 } }`,
			175,
			object.NUMBER_OBJ,
		},
	}

	runVmTests(t, tests, false, false)
}

func TestIntRangeExpressions(t *testing.T) {
	tests := []struct {
		input string
		start int64
		end   int64
	}{
		{"1 .. 3", 1, 3},
		{"1 + 1 .. 7", 2, 7},
		{"1 .. 3 * 7", 1, 21},
		{`17 \ 2 .. 3 * 7`, 8, 21},
	}

	for _, tt := range tests {
		program := parse(t, tt.input)

		comp, err := ast.NewCompiler(nil, false)
		if err != nil {
			t.Fatalf("(%q) compiler error on New: %s", tt.input, err)
		}

		_, err = program.Compile(comp)
		if err != nil {
			t.Fatalf("(%q) compiler error: %s", tt.input, err)
		}

		machine := vm.New(comp.ByteCode(), nil)
		err, _ = machine.Run()
		if err != nil {
			t.Fatalf("(%q) vm error: %s", tt.input, err)
		}

		stackElem := machine.LastValue()

		err = testRangeOfIntObject(tt.start, tt.end, stackElem)
		if err != nil {
			t.Errorf("(%q) Range Test Error: %s", tt.input, err)
		}
	}
}

func TestFunctionCallWithoutArgs(t *testing.T) {
	tests := []vmTestCase{
		{
			input:        "val yo = fn() { 7 * 7 }; \n yo();",
			expected:     "49",
			expectedType: object.NUMBER_OBJ,
		},

		{
			input: `val yo = fn() { 7 * 7 }
					val no = fn() { yo() + 10.14 }
					no()`,

			expected:     "59.14",
			expectedType: object.NUMBER_OBJ,
		},

		{
			input: `val yo = fn() { 7 * 7 }
					val no = fn() { yo() + 51 }
					val u = fn() { no() + yo() + no() }
					u()`,

			expected:     "249",
			expectedType: object.NUMBER_OBJ,
		},
	}

	runVmTests(t, tests, false, false)
}

func TestFunctionCallWithExplicitReturn(t *testing.T) {
	tests := []vmTestCase{
		{
			input:        "val earlyExit = fn() { return 50; 7; } \n earlyExit()",
			expected:     "50",
			expectedType: object.NUMBER_OBJ,
		},
		{
			input:        "val earlyExit = fn() { return 50; return 7; } \n earlyExit()",
			expected:     "50",
			expectedType: object.NUMBER_OBJ,
		},
		{
			input:        "val lateExit = fn() { 50; return 7; } \n lateExit()",
			expected:     "7",
			expectedType: object.NUMBER_OBJ,
		},
	}

	runVmTests(t, tests, false, false)
}

func TestFunctionCallWithoutReturnValue(t *testing.T) {
	tests := []vmTestCase{
		{
			input:        "val noReturn = fn() {}; noReturn()",
			expected:     nil,
			expectedType: object.NULL_OBJ,
		},
		{
			input: `val noReturn = fn() {}
					val noReturn2 = fn() { noReturn() }
					noReturn2()`,
			expected:     nil,
			expectedType: object.NULL_OBJ,
		},
	}

	runVmTests(t, tests, false, false)
}

func TestFirstClassFunctions(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `val f = fn() { 777 }
					val s = fn() { f }
					s()()`,
			expected:     "777",
			expectedType: object.NUMBER_OBJ,
		},
		{
			input: `
			val returnsOneReturner = fn() {
				val returnsOne = fn() { 7 }
				return returnsOne
			}
			returnsOneReturner()()
			`,
			expected:     "7",
			expectedType: object.NUMBER_OBJ,
		},
	}

	runVmTests(t, tests, false, false)
}

func TestCallingFunctionsWithBindings(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `
		val uno = fn() { val one = 1; one };
		uno();
		`,
			expected:     "1",
			expectedType: object.NUMBER_OBJ,
		},
		{
			input: `
		val oneAndTwo = fn() { val one = 1; val two = 2; one + two; };
		oneAndTwo();
		`,
			expected:     "3",
			expectedType: object.NUMBER_OBJ,
		},
		{
			input: `
		val oneAndTwo = fn() { val one = 1; val two = 2; one + two; };
		val threeAndFour = fn() { val three = 3; val four = 4; three + four; };
		oneAndTwo() + threeAndFour();
		`,
			expected:     "10",
			expectedType: object.NUMBER_OBJ,
		},
		{
			input: `
		val firstFoobar = fn() { val foobar = 50; foobar; };
		val secondFoobar = fn() { val foobar = 100; foobar; };
		firstFoobar() + secondFoobar();
		`,
			expected:     "150",
			expectedType: object.NUMBER_OBJ,
		},
		{
			input: `
		val globalSeed = 50;
		val minusOne = fn() {
			val num = 1;
			globalSeed - num;
		}
		val minusTwo = fn() {
			val num = 2;
			globalSeed - num;
		}
		minusOne() + minusTwo();
		`,
			expected:     "97",
			expectedType: object.NUMBER_OBJ,
		},
	}

	runVmTests(t, tests, false, false)
}

func TestCallingFunctionsWithArgumentsAndBindings(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `
		val identity = fn(a) { a }
		identity(4)
		`,
			expected:     "4",
			expectedType: object.NUMBER_OBJ,
		},
		{
			input: `
		val sum = fn(a, b) { a + b }
		sum(1, 2)
		`,
			expected:     "3",
			expectedType: object.NUMBER_OBJ,
		},
		{
			input: `
		val sum = fn(a, b) {
			val c = a + b
			c
		}
		sum(1, 2)
		`,
			expected:     "3",
			expectedType: object.NUMBER_OBJ,
		},
		{
			input: `
		val sum = fn(a, b) {
			val c = a + b
			c
		}
		sum(1, 2) + sum(3, 4)`,
			expected:     "10",
			expectedType: object.NUMBER_OBJ,
		},
		{
			input: `
		val sum = fn(a, b) {
			val c = a + b
			c
		}
		val outer = fn() {
			sum(1, 2) + sum(3, 4)
		}
		outer()
		`,
			expected:     "10",
			expectedType: object.NUMBER_OBJ,
		},
		{
			input: `
		val globalNum = 10

		val sum = fn(a, b) {
			val c = a + b
			c + globalNum
		}

		val outer = fn() {
			sum(1, 2) + sum(3, 4) + globalNum
		}

		outer() + globalNum
		`,
			expected:     "50",
			expectedType: object.NUMBER_OBJ,
		},
	}

	runVmTests(t, tests, false, false)
}

func TestOptionalParameters(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `
		val mult = fn(a, b=12) { a * b }
		mult(4)
		`,
			expected:     48,
			expectedType: object.NUMBER_OBJ,
		},
		{
			input: `
		val mult = fn(a, b=12) { a * b }
		mult(4, b=7)
		`,
			expected:     28,
			expectedType: object.NUMBER_OBJ,
		},

		{
			input: `
		val mult = fn(a, b=12, c=10) { a * b + c }
		mult(4)
		`,
			expected:     58,
			expectedType: object.NUMBER_OBJ,
		},
		{
			input: `
		val mult = fn(a, b=12, c=10) { a * b + c }
		mult(4, c=2.5)
		`,
			expected:     "50.5",
			expectedType: object.NUMBER_OBJ,
		},
		{
			input: `
		val mult = fn(a, b=12, c=10) { a * b + c }
		mult(4, c=2.5, b=10)
		`,
			expected:     "42.5",
			expectedType: object.NUMBER_OBJ,
		},
		{
			input: `
		val mult = fn(a, b=12, c=10) { a * b + c }
		mult(4, b=10, c=2.5)
		`,
			expected:     "42.5",
			expectedType: object.NUMBER_OBJ,
		},

		// with external name different than internal name (b as d)
		{
			input: `
		val mult = fn(a, b as d=12, c=4) { a * b + c }	# internally use b
		mult(4, d=10)									# called with d=...
		`,
			expected:     44,
			expectedType: object.NUMBER_OBJ,
		},
		{
			input: `
		val mult = fn(a, b as d=12, c=4) { a * b + c } # internally use b
		mult(4)
		`,
			expected:     52,
			expectedType: object.NUMBER_OBJ,
		},
		{ // ... and external name same as a keyword
			input: `
		val mult = fn(a, b as for=12, c=4) { a * b + c }	# internally use b
		mult(4, for=10)										# called with for=...
		`,
			expected:     44,
			expectedType: object.NUMBER_OBJ,
		},
		{ // ... and external name same as a keyword
			input: `
		val mult = fn(a, b as break=12, c=4) { a * b + c }	# internally use b
		mult(4, break=10)									# called with break=...
		`,
			expected:     44,
			expectedType: object.NUMBER_OBJ,
		},

		// with parameter mutability
		// {
		// 	input: `
		// val mult = fn(a, var b as break=12, c=4) { b += 1 ; a * b + c }	# internally use b
		// mult(4, break=10)									# called with break=...
		// `,
		// 	expected:     48,
		// 	expectedType: object.NUMBER_OBJ,
		// },

		{ // c not just a simple number; must be calculated
			input: `
		val mult = fn(a, b=12, c=4/2) { a * b + c }
		mult(4, b=10)
		`,
			expected:     42,
			expectedType: object.NUMBER_OBJ,
		},
		{
			input: `
		val mult = fn(a, b=12, c=4/2, d=42+0) { a * b - c + d }
		mult(4)
		`,
			expected:     88,
			expectedType: object.NUMBER_OBJ,
		},

		// with a "free" variable in setting the optional parameter default
		// shouldn't treat as a free variable (closure) if only part of parameter default
		{
			input: `
		val x = 2
		val mult = fn(a, b=12, c=x) { a * b - c }
		mult(4)
		`,
			expected:     46,
			expectedType: object.NUMBER_OBJ,
		},
		{
			input: `
		val x = 2
		val mult = fn(a, b=12, c=x) { a * b - c + x }
		mult(4)
		`,
			expected:     48,
			expectedType: object.NUMBER_OBJ,
		},
		{
			input: `
		val x = 2
		val y = 3
		val mult = fn(a, b=12, c=x) { a * b - c + y }
		mult(4)
		`,
			expected:     49,
			expectedType: object.NUMBER_OBJ,
		},

		// combined with parameter expansion
		{
			input: `
		val add = fn(a..., b=12) {
			var sum = 0
			for i in a {
				sum += i
			}
			sum += b
		}
		add()
		`,
			expected:     12,
			expectedType: object.NUMBER_OBJ,
		},
		{
			input: `
		val add = fn(a..., b=12) {
			var sum = 0
			for i in a {
				sum += i
			}
			sum += b
		}
		add(4)
		`,
			expected:     16,
			expectedType: object.NUMBER_OBJ,
		},
		{
			input: `
		val add = fn(a..., b=12) {
			var sum = 0
			for i in a {
				sum += i
			}
			sum += b
		}
		add(4, 7)
		`,
			expected:     23,
			expectedType: object.NUMBER_OBJ,
		},
		{
			input: `
		val add = fn(a..., b=12) {
			var sum = 0
			for i in a {
				sum += i
			}
			sum += b
		}
		add(4, 7, b=100)
		`,
			expected:     111,
			expectedType: object.NUMBER_OBJ,
		},

		// parameter expansion, argument expansion, and optional parameters
		{
			input: `
		val add = fn(a..., b=12) {
			var sum = 0
			for i in a {
				sum += i
			}
			sum += b
		}
		add([4, 7]..., b=100)
		`,
			expected:     111,
			expectedType: object.NUMBER_OBJ,
		},

		// optional parameters and a closure (both using OpFunction)
		{
			input: `
		val pre = 123
		val add = fn(alist, b=12) {
			var sum = pre	# closure on value
			for i in alist {
				sum += i
			}
			sum += b
		}
		add([4, 7], b=100)
		`,
			expected:     234,
			expectedType: object.NUMBER_OBJ,
		},

		// optional parameter using variable name same as another parameter to set default
		// parameter a not resolved while setting default value for b
		{
			input: `
		val a = 100
		val add = fn(a, b=a) {
			a + b
		}
		add(11)
		`,
			expected:     111,
			expectedType: object.NUMBER_OBJ,
		},
	
		// ... inside another function (compiling correctly?)
		{
			input: `
		val a = 100
		val test = fn() {
			val add = fn(a, b=a) {
				a + b
			}
			add(11)
		}
		test()
		`,
			expected:     111,
			expectedType: object.NUMBER_OBJ,
		},
		{
			input: `
		val test = fn() {
			val a = 100
			val add = fn(a, b=a) {
				a + b
			}
			add(11)
		}
		test()
		`,
			expected:     111,
			expectedType: object.NUMBER_OBJ,
		},

		// another function within with default values pointing to undefined values
		{
			input: `
		val add = fn(a, b) {
			val second = fn(x=a, y=b) {
				x + y
			}
			second()
		}
		add(3, 4)
		`,
			expected:     7,
			expectedType: object.NUMBER_OBJ,
		},
		{
			input: `
		val add = fn(a, b) {
			val c, d = a, b
			val second = fn(x=c, y=d) {
				x + y
			}
			second()
		}
		add(3, 4)
		`,
			expected:     7,
			expectedType: object.NUMBER_OBJ,
		},
	}

	runVmTests(t, tests, false, false)
}

func TestParametersRequiredByName(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `
				val add = fn(a, b as b) { a + b }
				add(4, b=10)`,
			expected:     14,
			expectedType: object.NUMBER_OBJ,
		},
		{
			input: `
				val add = fn(a number, b as b number) { a + b }
				add(4, b=10)`,
			expected:     14,
			expectedType: object.NUMBER_OBJ,
		},
	}

	runVmTests(t, tests, false, false)
}

func TestCallingFunctionsWithWrongArgumentCountOrType(t *testing.T) {
	tests := []vmTestCase{
		{
			input:    `fn() { }(1);`,
			expected: `args: Positional argument/parameter count mismatch, expected=0, received=1 ()`,
		},
		{
			input:    `fn(a) { }();`,
			expected: `args: Positional argument/parameter count mismatch, expected=1, received=0 ()`,
		},
		{
			input:    `fn(a, b) { }(1);`,
			expected: `args: Positional argument/parameter count mismatch, expected=2, received=1 ()`,
		},

		// with argument expansion
		{
			input:    `fn(a, b) { }([1, 2, 3]...);`,
			expected: `args: Positional argument/parameter count mismatch, expected=2, received=3 ()`,
		},
		{
			input:    `fn(a, b) { }(7, [1, 2, 3]...);`,
			expected: `args: Positional argument/parameter count mismatch, expected=2, received=4 ()`,
		},
		{
			input:    `fn(a, b) { }(7, []...);`,
			expected: `args: Positional argument/parameter count mismatch, expected=2, received=1 ()`,
		},

		// with parameter expansion
		{
			input:    `fn(a, b ...[1..]) { }(1);`,
			expected: `args: Parameter expansion min (1) not met (0) ()`,
			//`args: argument/parameter count mismatch, expected=2..-1, received=1 (fn)`,
		},
		{
			input:    `fn(a, c, b...) { }(1);`,
			expected: fmt.Sprintf(`args: Positional argument/parameter count mismatch, expected=2..%d, received=1 ()`, common.ArgCountMax),
		},

		{
			input:    `fn(a, c, b...[1..3]) { }(1, 2, 3, 4, 5, 6, 7);`,
			expected: `args: Parameter expansion max (3) exceeded (5) ()`,
		},
		{
			input:    `fn(a, c, b...[1..3]) { }(1, 2, [3, 4, 5, 6]...);`,
			expected: `args: Parameter expansion max (3) exceeded (4) ()`,
		},

		// wrong arg type
		{
			input:    `fn(a string) { }(123);`,
			expected: `args: Argument 1 type (number) does not match parameter a type (string) ()`,
		},
		{
			input:    `val x = fn(a string) { }; x(123);`,
			expected: `args: Argument 1 type (number) does not match parameter a type (string) (x)`,
		},
		{
			input:    `fn(a string, b number) { }(123, "");`,
			expected: `args: Argument 1 type (number) does not match parameter a type (string) ()`,
		},
		{
			input:    `fn(a string, b number) { }("", "");`,
			expected: `args: Argument 2 type (string) does not match parameter b type (number) ()`,
		},

		{
			input:    `fn(a string, b string = "") { }("yes", b=3);`,
			expected: `args: Argument b type (number) does not match parameter b type (string) ()`,
		},
		{
			input:    `fn(a number, b string = "") { }(132, b=3);`,
			expected: `args: Argument b type (number) does not match parameter b type (string) ()`,
		},
		{
			input:    `fn(a number, b string = "") { }("", b=3);`,
			expected: `args: Argument 1 type (string) does not match parameter a type (number) ()`,
		},
	}

	for _, tt := range tests {
		program := parse(t, tt.input)

		comp, err := ast.NewCompiler(nil, false)
		if err != nil {
			t.Fatalf("(%s)\ncompiler error: %s", tt.input, err)
		}

		_, err = program.Compile(comp)
		if err != nil {
			t.Fatalf("(%s)\ncompiler error: %s", tt.input, err)
		}

		machine := vm.New(comp.ByteCode(), nil)
		err, _ = machine.Run()
		if err == nil {
			t.Fatalf("(%s)\nexpected VM error but resulted in none.", tt.input)
		}

		if err.Error() != tt.expected {
			t.Fatalf("(%s)\nwrong VM error: wanted=%q\nreceived=%q", tt.input, tt.expected, err)
		}
	}
}

func TestPurityOfFunctionsAndClosures(t *testing.T) {
	tests := []vmTestCase{
		{ // changing the function variable used to make the closure...
			`var makeAdder = fn(x) { fn(y) { x + y }}
			val adder = makeAdder(7)

			# change makeAdder before testing adder
			makeAdder = fn(x) { fn(y) { x + y + 777 }}

			adder(14)`,
			"21",
			object.NUMBER_OBJ,
		},

		{ // changing the value used in building a closure's "free" variable...
			`var x = 7
			val chk = fn() { x ^ 2 }

			# change x before testing chk()
			x = 21

			chk()
		`,
			"49",
			object.NUMBER_OBJ,
		},
	}

	runVmTests(t, tests, false, false)
}

func TestPurityOfClosuresWithRedefinitionAndSingleCompile(t *testing.T) {
	tests := []vmTestCase{
		{`
			var add1, add2
			for i of 7 {
			   val temp = fn() { i }
			   switch i {
			      case 3: add1 = temp
				  case 5: add2 = temp
	  		   }
			}
			add1() + add2()
		`,
			"8",
			object.NUMBER_OBJ},
	}

	runVmTests(t, tests, false, false)
}

func TestImpureFunctionDeclarations(t *testing.T) {
	tests := []vmTestCase{
		{`val x = fn*() { writeln("something") }
		  21`,
			21, object.NUMBER_OBJ,
		},

		// can pass impure value to impure function
		{`val x = fn*(a) { 42 }
		  x(cd)`,
			42, object.NUMBER_OBJ,
		},

		// cannot pass impure value to pure function
		{`val x = fn(a) { 42 }
		  x(cd)
		  catch: 30`,
			30, object.NUMBER_OBJ,
		},

		// cannot pass impure value to built-in function
		{`val x = fn*(a) { a * 2 }
		  map x, [1, 2]
		  21
		  catch: 30`,
			30, object.NUMBER_OBJ,
		},
		{`val x = fn*(a) { a * 2 }
		  val y = fn(a) { a * 40 }
		  map [x, y], [1, 2]
		  21
		  catch: 30`,
			30, object.NUMBER_OBJ,
		},

		// Impurity is transitive.
		{`val x = fn*(a) { val z = fn*(y) { y * 2 }; z(a) }
		  x(12)
		  catch: 30`,
			24, object.NUMBER_OBJ,
		},
	}

	runVmTests(t, tests, false, false)
}

func TestBuiltinFunctions(t *testing.T) {
	tests := []vmTestCase{
		// length of string in code points
		{`len("")`, "0", object.NUMBER_OBJ},
		{`len("four")`, "4", object.NUMBER_OBJ},

		// NOTE: This next line displays wrong in my IDE, b/c of L/R confusion but returns the correct result (3).
		{`len("שלם")`, "3", object.NUMBER_OBJ},

		// length of list
		{"len([])", "0", object.NUMBER_OBJ},
		{"len([7])", "1", object.NUMBER_OBJ},
		{"len([7, 14, 21])", "3", object.NUMBER_OBJ},

		// length of hash
		{"len({:})", "0", object.NUMBER_OBJ},
		{"len({7: 14})", "1", object.NUMBER_OBJ},
		{"len({7: 14, 21: 35, 42: 49})", "3", object.NUMBER_OBJ},

		// non-null
		{"nn([null], alt=789)", "789", object.NUMBER_OBJ},
		{"nn([null, 123], alt=789)", "123", object.NUMBER_OBJ},
		{"nn([null, null, true, 123], alt=false)", true, object.BOOLEAN_OBJ},

		// reverse
		{"reverse([16, 14, 16, 13, 12, 25, 36, 42, 29, 49])",
			[]int{49, 29, 42, 36, 25, 12, 13, 16, 14, 16}, object.LIST_OBJ,
		},
		{"reverse(reverse([16, 14, 16, 13, 12, 25, 36, 42, 29, 49]))",
			[]int{16, 14, 16, 13, 12, 25, 36, 42, 29, 49}, object.LIST_OBJ,
		},
		{`reverse("abcd")`, "dcba", object.STRING_OBJ},
		{`reverse("a\u0341bcd")`, "dcba\u0341", object.STRING_OBJ},

		// attempt to reverse keys and values of hash
		{"reverse({1: 2, 2: 7})",
			[][]object.Object{
				{object.NumberFromInt(2), object.NumberFromInt(1)},
				{object.NumberFromInt(7), object.NumberFromInt(2)},
			},
			object.HASH_OBJ,
		},

		{"reverse(7)", 7, object.NUMBER_OBJ},
		{"reverse(1234)", 4321, object.NUMBER_OBJ},
		{"reverse(12345)", 54321, object.NUMBER_OBJ},
		{"reverse(-12345)", -54321, object.NUMBER_OBJ},
		{"reverse(-1234)", -4321, object.NUMBER_OBJ},
		{"reverse(1.234)", "432.1", object.NUMBER_OBJ},
		{"reverse(-12.34)", "-43.21", object.NUMBER_OBJ},

		// rotate list with positive number
		// 1 implied if not specified
		{"rotate([16, 14, 16, 13, 12, 25, 36, 42, 29, 49])",
			[]int{14, 16, 13, 12, 25, 36, 42, 29, 49, 16}, object.LIST_OBJ,
		},
		{"rotate([16, 14, 16, 13, 12, 25, 36, 42, 29, 49], distance=2)",
			[]int{16, 13, 12, 25, 36, 42, 29, 49, 16, 14}, object.LIST_OBJ,
		},
		{"rotate([16, 14, 16, 13, 12, 25, 36, 42, 29, 49], distance=10)", // same: no rotation
			[]int{16, 14, 16, 13, 12, 25, 36, 42, 29, 49}, object.LIST_OBJ,
		},
		{"rotate([16, 14, 16, 13, 12, 25, 36, 42, 29, 49], distance=12)", // excessive: 12 becomes 2
			[]int{16, 13, 12, 25, 36, 42, 29, 49, 16, 14}, object.LIST_OBJ,
		},

		{"rotate([])",
			[]int{}, object.LIST_OBJ,
		},
		{"rotate([], distance=7)",
			[]int{}, object.LIST_OBJ,
		},
		{"rotate([21])",
			[]int{21}, object.LIST_OBJ,
		},
		{"rotate([21], distance=3)",
			[]int{21}, object.LIST_OBJ,
		},

		// rotate list with negative number
		{"rotate([16, 14, 16, 13, 12, 25, 36, 42, 29, 49], distance=-1)",
			[]int{49, 16, 14, 16, 13, 12, 25, 36, 42, 29}, object.LIST_OBJ,
		},
		{"rotate([16, 14, 16, 13, 12, 25, 36, 42, 29, 49], distance=-2)",
			[]int{29, 49, 16, 14, 16, 13, 12, 25, 36, 42}, object.LIST_OBJ,
		},
		{"rotate([16, 14, 16, 13, 12, 25, 36, 42, 29, 49], distance=-10)", // same: no rotation
			[]int{16, 14, 16, 13, 12, 25, 36, 42, 29, 49}, object.LIST_OBJ,
		},
		{"rotate([16, 14, 16, 13, 12, 25, 36, 42, 29, 49], distance=-12)", // excessive: -12 becomes -2
			[]int{29, 49, 16, 14, 16, 13, 12, 25, 36, 42}, object.LIST_OBJ,
		},

		{"rotate([], distance=-1)",
			[]int{}, object.LIST_OBJ,
		},
		{"rotate([], distance=-7)",
			[]int{}, object.LIST_OBJ,
		},
		{"rotate([21], distance=-1)",
			[]int{21}, object.LIST_OBJ,
		},
		{"rotate([21], distance=-3)",
			[]int{21}, object.LIST_OBJ,
		},

		// rotate number within range
		// number outside of range passed through
		{"rotate(10, distance=-12, range=3..14)", "10", object.NUMBER_OBJ},
		{"rotate(10, distance=-12, range=14..21)", "10", object.NUMBER_OBJ},
		{"rotate(2, distance=-12, range=1..21)", "14", object.NUMBER_OBJ},
		{"rotate(2, distance=-33, range=1..21)", "14", object.NUMBER_OBJ},
		{"rotate(20, distance=-1, range=1..21)", "21", object.NUMBER_OBJ},
		{"rotate(20, distance=-2, range=1..21)", "1", object.NUMBER_OBJ},

		{"rotate(2, distance=12, range=1..21)", "11", object.NUMBER_OBJ},
		{"rotate(2, distance=33, range=1..21)", "11", object.NUMBER_OBJ},
		{"rotate(2, distance=1, range=1..21)", "1", object.NUMBER_OBJ},
		{"rotate(2, distance=2, range=1..21)", "21", object.NUMBER_OBJ},

		// group
		{`group([100, 2, 7, 98, 78], by=fn{< 50})`, [][]int{{2, 7}, {100, 98, 78}}, object.LIST_OBJ},
		{`group(["a", "abc", "z", "zzz", "ab"], by=len)`, [][]string{{"a", "z"}, {"abc", "zzz"}, {"ab"}}, object.LIST_OBJ},

		{`group([100, 2, 7, 98, 78], by=2)`, [][]int{{100, 2}, {7, 98}, {78}}, object.LIST_OBJ},
		{`group([100, 2, 7, 98, 78], by=-2)`, [][]int{{100}, {2, 7}, {98, 78}}, object.LIST_OBJ},
		{`group([100, 2, 7, 98, 78], by=3)`, [][]int{{100, 2, 7}, {98, 78}}, object.LIST_OBJ},
		{`group([100, 2, 7, 98, 78], by=-3)`, [][]int{{100, 2}, {7, 98, 78}}, object.LIST_OBJ},

		{`group([1, 2, 3, 4, 5, 6, 7, 8, 9, 10], by=3)`,
			[][]int{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}, {10}}, object.LIST_OBJ},
		{`group([1, 2, 3, 4, 5, 6, 7, 8, 9, 10], by=-3)`,
			[][]int{{1}, {2, 3, 4}, {5, 6, 7}, {8, 9, 10}}, object.LIST_OBJ},
		{`group([0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10], by=-3)`,
			[][]int{{0, 1}, {2, 3, 4}, {5, 6, 7}, {8, 9, 10}}, object.LIST_OBJ},
		{`group([-1, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10], by=-3)`,
			[][]int{{-1, 0, 1}, {2, 3, 4}, {5, 6, 7}, {8, 9, 10}}, object.LIST_OBJ},

		{`group([], by=3)`,
			[][]int{}, object.LIST_OBJ},
		{`group([], by=-3)`,
			[][]int{}, object.LIST_OBJ},
		{`group([1], by=3)`,
			[][]int{{1}}, object.LIST_OBJ},
		{`group([1], by=-3)`,
			[][]int{{1}}, object.LIST_OBJ},
		{`group([1, 2], by=3)`,
			[][]int{{1, 2}}, object.LIST_OBJ},
		{`group([1, 2], by=-3)`,
			[][]int{{1, 2}}, object.LIST_OBJ},
		{`group([1, 2, 3], by=3)`,
			[][]int{{1, 2, 3}}, object.LIST_OBJ},
		{`group([1, 2, 3], by=-3)`,
			[][]int{{1, 2, 3}}, object.LIST_OBJ},
		{`group([1, 2, 3, 4], by=3)`,
			[][]int{{1, 2, 3}, {4}}, object.LIST_OBJ},
		{`group([1, 2, 3, 4], by=-3)`,
			[][]int{{1}, {2, 3, 4}}, object.LIST_OBJ},
		{`group([1, 2, 3, 4, 5], by=3)`,
			[][]int{{1, 2, 3}, {4, 5}}, object.LIST_OBJ},
		{`group([1, 2, 3, 4, 5], by=-3)`,
			[][]int{{1, 2}, {3, 4, 5}}, object.LIST_OBJ},
		{`group([1, 2, 3, 4, 5, 6], by=3)`,
			[][]int{{1, 2, 3}, {4, 5, 6}}, object.LIST_OBJ},
		{`group([1, 2, 3, 4, 5, 6], by=-3)`,
			[][]int{{1, 2, 3}, {4, 5, 6}}, object.LIST_OBJ},
		{`group([1, 2, 3, 4, 5, 6, 7], by=3)`,
			[][]int{{1, 2, 3}, {4, 5, 6}, {7}}, object.LIST_OBJ},
		{`group([1, 2, 3, 4, 5, 6, 7], by=-3)`,
			[][]int{{1}, {2, 3, 4}, {5, 6, 7}}, object.LIST_OBJ},

		{`group([], by=2)`,
			[][]int{}, object.LIST_OBJ},
		{`group([], by=-2)`,
			[][]int{}, object.LIST_OBJ},
		{`group([1], by=2)`,
			[][]int{{1}}, object.LIST_OBJ},
		{`group([1], by=-2)`,
			[][]int{{1}}, object.LIST_OBJ},
		{`group([1, 2], by=2)`,
			[][]int{{1, 2}}, object.LIST_OBJ},
		{`group([1, 2], by=-2)`,
			[][]int{{1, 2}}, object.LIST_OBJ},
		{`group([1, 2, 3], by=2)`,
			[][]int{{1, 2}, {3}}, object.LIST_OBJ},
		{`group([1, 2, 3], by=-2)`,
			[][]int{{1}, {2, 3}}, object.LIST_OBJ},

		{`group([], by=1)`,
			[][]int{}, object.LIST_OBJ},
		{`group([], by=-1)`,
			[][]int{}, object.LIST_OBJ},
		{`group([1], by=1)`,
			[][]int{{1}}, object.LIST_OBJ},
		{`group([1], by=-1)`,
			[][]int{{1}}, object.LIST_OBJ},
		{`group([1, 2], by=1)`,
			[][]int{{1}, {2}}, object.LIST_OBJ},
		{`group([1, 2], by=-1)`,
			[][]int{{1}, {2}}, object.LIST_OBJ},
		{`group([1, 2, 3], by=1)`,
			[][]int{{1}, {2}, {3}}, object.LIST_OBJ},
		{`group([1, 2, 3], by=-1)`,
			[][]int{{1}, {2}, {3}}, object.LIST_OBJ},

		// group with 1 argument (by truthiness)
		// true first, false second, even if one or both empty
		{`group([])`,
			[][]int{{}, {}}, object.LIST_OBJ},
		{`group([1, 2, 3])`,
			[][]int{{1, 2, 3}, {}}, object.LIST_OBJ},
		{`group([0, 1, 2, 3])`,
			[][]int{{1, 2, 3}, {0}}, object.LIST_OBJ},
		{`string(group([true, false, null, [], [7]]))`,
			"[[true, [7]], [false, null, []]]", object.STRING_OBJ},
		{`string(group([false, null, []]))`,
			"[[], [false, null, []]]", object.STRING_OBJ},
		{`string(group([false, null, [], 7]))`,
			"[[7], [false, null, []]]", object.STRING_OBJ},

		// group on string code points
		{`group(s2cp("abcdefghijkl"), by=fn{< 100})`,
			[][]int{{97, 98, 99}, {100, 101, 102, 103, 104, 105, 106, 107, 108}}, object.LIST_OBJ,
		},

		// groupby
		{`string(groupby([100, 2, 7, 98, 78], by=fn{< 50}))`,
			"[[true, [2, 7]], [false, [100, 98, 78]]]", object.STRING_OBJ},

		{`string(groupby(["don't you know", "nada", 123, "hey", "yo"], by=fn(x) { x -> re/yo/ }))`,
			`[[true, ["don't you know", "yo"]], [false, ["nada", 123, "hey"]]]`, object.STRING_OBJ},

		{`val test = fn(x) { if(x < 0: "neg"; x > 200: "over"; "good") }
			string(groupby([-123, 345, 89, 150, 1000, -3.4], by=test))`,
			`[["neg", [-123, -3.4]], ["over", [345, 1000]], ["good", [89, 150]]]`, object.STRING_OBJ},

		// groupby with 1 argument (by truthiness)
		{`string(groupby([true, false, null, [], [7]]))`,
			"[[true, [true, [7]]], [false, [false, null, []]]]", object.STRING_OBJ},

		// groupby on string code points
		{`string(groupby(s2cp("abcdefghijkl"), by=fn(i) { if(i < 100: 0; 1) }))`,
			`[[0, [97, 98, 99]], [1, [100, 101, 102, 103, 104, 105, 106, 107, 108]]]`, object.STRING_OBJ,
		},

		// groupbyH
		{`groupbyH([100, 2, 7, 98, 78], by=fn(i) { if(i < 50: "A"; "B") })`,
			[][]object.Object{
				{object.NewString("A"), &object.List{Elements: []object.Object{
					object.NumberFromInt(2), object.NumberFromInt(7),
				}}},
				{object.NewString("B"), &object.List{Elements: []object.Object{
					object.NumberFromInt(100), object.NumberFromInt(98), object.NumberFromInt(78),
				}}},
			},
			object.HASH_OBJ},

		{`groupbyH(["don't you know", "nada", 123, "hey", "yo"], by=fn(x) { if(x -> re/yo/: "true"; "false") })`,
			[][]object.Object{
				{object.NewString("true"), &object.List{Elements: []object.Object{
					object.NewString("don't you know"), object.NewString("yo"),
				}}},
				{object.NewString("false"), &object.List{Elements: []object.Object{
					object.NewString("nada"), object.NumberFromInt(123), object.NewString("hey"),
				}}},
			},
			object.HASH_OBJ},

		{`val test = fn(x) { if(x < 0: "neg"; x > 200: "over"; "good") }
			groupbyH([-123, 345, 89, 150, 1000, -34], by=test)`,
			[][]object.Object{
				{object.NewString("neg"), &object.List{Elements: []object.Object{
					object.NumberFromInt(-123), object.NumberFromInt(-34),
				}}},
				{object.NewString("good"), &object.List{Elements: []object.Object{
					object.NumberFromInt(89), object.NumberFromInt(150),
				}}},
				{object.NewString("over"), &object.List{Elements: []object.Object{
					object.NumberFromInt(345), object.NumberFromInt(1000),
				}}},
			},
			object.HASH_OBJ},

		{`val test = fn(x) { if(x < 0: -1; x > 200: 1; 0) }
			groupbyH([-123, 345, 89, 150, 1000, -34], by=test)`,
			[][]object.Object{
				{object.NumberFromInt(0), &object.List{Elements: []object.Object{
					object.NumberFromInt(89), object.NumberFromInt(150),
				}}},
				{object.NumberFromInt(1), &object.List{Elements: []object.Object{
					object.NumberFromInt(345), object.NumberFromInt(1000),
				}}},
				{object.NumberFromInt(-1), &object.List{Elements: []object.Object{
					object.NumberFromInt(-123), object.NumberFromInt(-34),
				}}},
			},
			object.HASH_OBJ},

		// groupbyH on string code points
		{`groupbyH(s2cp("abcdefg"), by=fn(i) { if(i < 100: 0; 1) })`,
			[][]object.Object{
				{object.NumberFromInt(0), &object.List{Elements: []object.Object{
					object.NumberFromInt(97), object.NumberFromInt(98), object.NumberFromInt(99),
				}}},
				{object.NumberFromInt(1), &object.List{Elements: []object.Object{
					object.NumberFromInt(100), object.NumberFromInt(101), object.NumberFromInt(102), object.NumberFromInt(103),
				}}},
			},
			object.HASH_OBJ,
		},

		// any
		{`any([10, 101], by=fn{> 100})`, true, object.BOOLEAN_OBJ},
		{`any([10, 99], by=fn{> 100})`, false, object.BOOLEAN_OBJ},
		{`any([10, 101], by=re/01/)`, true, object.BOOLEAN_OBJ},
		{`any([10, 99], by=re/02/)`, false, object.BOOLEAN_OBJ},
		{`any({1: 10, 2: 101}, by=fn{> 100})`, true, object.BOOLEAN_OBJ},
		{`any({1: 10, 2: 99}, by=fn{> 100})`, false, object.BOOLEAN_OBJ},
		{`any({1: 10, 2: 101}, by=re/01/)`, true, object.BOOLEAN_OBJ},
		{`any({1: 10, 2: 99}, by=re/02/)`, false, object.BOOLEAN_OBJ},

		// any with one argument (by truthiness)
		{`any([123, 345])`, true, object.BOOLEAN_OBJ},
		{`any([true, [123]])`, true, object.BOOLEAN_OBJ},
		{`any([null, [], false])`, false, object.BOOLEAN_OBJ},
		{`any({1: 123, 2: null})`, true, object.BOOLEAN_OBJ},
		{`any({1: [], 2: null})`, false, object.BOOLEAN_OBJ},

		// all
		{`all([10, 101], by=fn{> 7})`, true, object.BOOLEAN_OBJ},
		{`all([10, 101], by=fn{> 100})`, false, object.BOOLEAN_OBJ},
		{`all([101, 10], by=fn{> 100})`, false, object.BOOLEAN_OBJ},
		{`all([10, 101], by=re/01/)`, false, object.BOOLEAN_OBJ},
		{`all([10, 99], by=re/[0-9]/)`, true, object.BOOLEAN_OBJ},
		{`all({1: 10, 2: 101}, by=fn{> 7})`, true, object.BOOLEAN_OBJ},
		{`all({1: 10, 2: 101}, by=fn{> 100})`, false, object.BOOLEAN_OBJ},
		{`all({1: 10, 2: 101}, by=re/01/)`, false, object.BOOLEAN_OBJ},
		{`all({1: 10, 2: 99}, by=re/[0-9]/)`, true, object.BOOLEAN_OBJ},

		// all with one argument (by truthiness)
		{`all([123, 345])`, true, object.BOOLEAN_OBJ},
		{`all([true, [123]])`, true, object.BOOLEAN_OBJ},
		{`all([true, [123], false])`, false, object.BOOLEAN_OBJ},
		{`all({1: 123, 2: 890})`, true, object.BOOLEAN_OBJ},
		{`all({1: 123, 2: null})`, false, object.BOOLEAN_OBJ},

		// all or any with empty hash or list
		{`all({:}, by=re/[0-9]/)`, nil, object.NULL_OBJ},
		{`all({:}, by=fn{> 100})`, nil, object.NULL_OBJ},
		{`all([], by=re/[0-9]/)`, nil, object.NULL_OBJ},
		{`all([], by=fn{> 100})`, nil, object.NULL_OBJ},
		{`any({:}, by=re/[0-9]/)`, nil, object.NULL_OBJ},
		{`any({:}, by=fn{> 100})`, nil, object.NULL_OBJ},
		{`any([], by=re/[0-9]/)`, nil, object.NULL_OBJ},
		{`any([], by=fn{> 100})`, nil, object.NULL_OBJ},

		{`all([])`, nil, object.NULL_OBJ},
		{`all({:})`, nil, object.NULL_OBJ},
		{`any([])`, nil, object.NULL_OBJ},
		{`any({:})`, nil, object.NULL_OBJ},

		// join
		{`join(["abc", "123"], delim=",")`, "abc,123", object.STRING_OBJ},
		// join with auto-stringification
		{`join(series(1..7), delim=" ")`, "1 2 3 4 5 6 7", object.STRING_OBJ},

		{`join(["abc", "123"])`, "abc123", object.STRING_OBJ},
		{`join(series(7))`, "1234567", object.STRING_OBJ},

		// random
		{`val x = random(2); x == 1 or x == 2`, true, object.BOOLEAN_OBJ},
		{`val x = random(-2); x == -1 or x == -2`, true, object.BOOLEAN_OBJ},
		{`val x = random(7); x >= 1 and x <= 7`, true, object.BOOLEAN_OBJ},
		{`val x = random(100); x >= 1 and x <= 100`, true, object.BOOLEAN_OBJ},
		{`val x = random(-7); x >= -7 and x <= -1`, true, object.BOOLEAN_OBJ},
		{`val x = random(-100); x >= -100 and x <= -1`, true, object.BOOLEAN_OBJ},
		{`val x = random(1..7); x >= 1 and x <= 7`, true, object.BOOLEAN_OBJ},
		{`val x = random(0..7); x >= 0 and x <= 7`, true, object.BOOLEAN_OBJ},
		{`val x = random(-100..7); x >= -100 and x <= 7`, true, object.BOOLEAN_OBJ},
		{`val x = random(7..100); x >= 7 and x <= 100`, true, object.BOOLEAN_OBJ},
		{`val x = random(100..7); x >= 7 and x <= 100`, true, object.BOOLEAN_OBJ},

		// random with string
		{`val x = random("abc"); x >= 97 and x <= 99`, true, object.BOOLEAN_OBJ},
		// random with list
		{`val x = random(series(7 .. 14)); x >= 7 and x <= 14`, true, object.BOOLEAN_OBJ},
		// random with range
		{`val x = random(7 .. 14); x >= 7 and x <= 14`, true, object.BOOLEAN_OBJ},
		// random with hash
		{`val x = random({3: 7, 99: 8, 101: 12}); x >= 7 and x <= 12`, true, object.BOOLEAN_OBJ},

		// series
		{`series(1..3)`, []int{1, 2, 3}, object.LIST_OBJ},
		{`series(3..1)`, []int{3, 2, 1}, object.LIST_OBJ},
		{`series(1..3, inc=2)`, []int{1, 3}, object.LIST_OBJ},
		{`series(3..1, inc=-2)`, []int{3, 1}, object.LIST_OBJ},
		{`series(0..4, inc=2)`, []int{0, 2, 4}, object.LIST_OBJ},
		{`series(4..0, inc=-2)`, []int{4, 2, 0}, object.LIST_OBJ},

		{`series(1..3, asconly=true)`, []int{1, 2, 3}, object.LIST_OBJ},
		{`series(3..1, asconly=true)`, []int{}, object.LIST_OBJ},
		{`series(1..3, inc=2, asconly=true)`, []int{1, 3}, object.LIST_OBJ},
		{`series(3..1, asconly=true, inc=-2)`, []int{}, object.LIST_OBJ},
		{`series(0..4, asconly=true, inc=2)`, []int{0, 2, 4}, object.LIST_OBJ},
		{`series(4..0, asconly=true, inc=-2)`, []int{}, object.LIST_OBJ},

		{`string(series(1..3, inc=0.5))`, "[1, 1.5, 2.0, 2.5, 3.0]", object.STRING_OBJ},
		{`string(series(3..1, inc=-0.5))`, "[3, 2.5, 2.0, 1.5, 1.0]", object.STRING_OBJ},

		{`string(series(3.1..3.5, inc=0.1))`, "[3.1, 3.2, 3.3, 3.4, 3.5]", object.STRING_OBJ},

		// with large numbers (outside int64)
		{`string(series(9223372036854775808 .. 9223372036854775812, inc=2))`, 
			"[9223372036854775808, 9223372036854775810, 9223372036854775812]", object.STRING_OBJ},
		{`string(series(9223372036854775812 .. 9223372036854775808, inc=-2))`, 
			"[9223372036854775812, 9223372036854775810, 9223372036854775808]", object.STRING_OBJ},
		{`string(series(-9223372036854775812 .. -9223372036854775808, inc=2))`, 
			"[-9223372036854775812, -9223372036854775810, -9223372036854775808]", object.STRING_OBJ},
		{`string(series(-9223372036854775808 .. -9223372036854775812, inc=-2))`, 
			"[-9223372036854775808, -9223372036854775810, -9223372036854775812]", object.STRING_OBJ},

		// sort using implied operator function
		{"sort([16, 14, 16, 13, 12, 25, 36, 42, 29, 49], by=fn{<})",
			[]int{12, 13, 14, 16, 16, 25, 29, 36, 42, 49}, object.LIST_OBJ,
		},

		// sort other direction using function
		{"sort([16, 14, 16, 13, 12, 25, 36, 42, 29, 49], by=fn{>})",
			[]int{49, 42, 36, 29, 25, 16, 16, 14, 13, 12}, object.LIST_OBJ,
		},

		// sort from single parameter function
		{`sort(["abcd", "ab", "abc", "zzzzzzz"], by=len)`,
			[]string{"ab", "abc", "abcd", "zzzzzzz"}, object.LIST_OBJ,
		},

		// sort without a function
		{`sort([45, 2, 56, 12, 7, 12])`,
			[]int{2, 7, 12, 12, 45, 56}, object.LIST_OBJ,
		},

		// sort ranges
		{"sort(3 .. 7, by=fn{<})",
			[]int64{3, 7}, object.RANGE_OBJ,
		},
		{"sort(3 .. 7, by=fn{>})",
			[]int64{7, 3}, object.RANGE_OBJ,
		},

		{"sort(3..7)",
			[]int64{3, 7}, object.RANGE_OBJ,
		},
		{"sort(7..3)",
			[]int64{3, 7}, object.RANGE_OBJ,
		},

		// fold
		{`fold([], by=fn{*})`,
			nil,
			object.NULL_OBJ,
		},
		{`fold([7], by=fn{*})`,
			7,
			object.NUMBER_OBJ,
		},
		{`fold([7, 7], by=fn{*})`,
			49,
			object.NUMBER_OBJ,
		},
		{`fold([7, 7, 21], by=fn{*})`,
			1029,
			object.NUMBER_OBJ,
		},

		{`fold(1..4, by=fn{*})`,
			24,
			object.NUMBER_OBJ,
		},
		{`fold(4..1, by=fn{*})`,
			24,
			object.NUMBER_OBJ,
		},
	}

	runVmTests(t, tests, false, false)
}

func TestFilterFunctions(t *testing.T) {
	tests := []vmTestCase{
		// filter
		{`filter([16, 16, 25, 36, 42, 29, 49], by=re/1/)`,
			[]int{16, 16}, object.LIST_OBJ,
		},
		{`filter({1: 16, 2: 16, 3: 25, 4: 36, 5: 42, 6: 29, 7: 56}, by=re/[234]/)`,
			[][]object.Object{
				{object.NumberFromInt(3), object.NumberFromInt(25)},
				{object.NumberFromInt(4), object.NumberFromInt(36)},
				{object.NumberFromInt(5), object.NumberFromInt(42)},
				{object.NumberFromInt(6), object.NumberFromInt(29)},
			},
			object.HASH_OBJ,
		},

		// filter with 1 argument
		{`string(filter([[], 123, "abc", null]))`,
			`[123, "abc"]`, object.STRING_OBJ,
		},
		{`filter({1: null, 2: 16, 3: false, 4: true, 5: "yo"})`,
			[][]object.Object{
				{object.NumberFromInt(2), object.NumberFromInt(16)},
				{object.NumberFromInt(4), object.NativeBoolToObject(true)},
				{object.NumberFromInt(5), object.NewString("yo")},
			},
			object.HASH_OBJ,
		},

		// count
		{`count([], by=fn{> 100})`, "0", object.NUMBER_OBJ},
		{`count([79, 7, 9], by=fn{> 100})`, "0", object.NUMBER_OBJ},
		{`count([79, 7, 9], by=fn{< 100})`, "3", object.NUMBER_OBJ},
		{`count([123, 90, 170], by=fn{> 100})`, "2", object.NUMBER_OBJ},
		{`count(["yo", "abc", "zzz"], by=re/[0-9]/)`, "0", object.NUMBER_OBJ},
		{`count([123, "abc9", "zzz"], by=re/[0-9]/)`, "2", object.NUMBER_OBJ},
		{`count([123, "abc9", "zzz0"], by=re/[0-9]/)`, "3", object.NUMBER_OBJ},

		{`count({1: "yo", 2: "abc", 3: "zzz"}, by=re/[0-9]/)`, "0", object.NUMBER_OBJ},
		{`count({1: 123, 2: "abc9", 3: "zzz"}, by=re/[0-9]/)`, "2", object.NUMBER_OBJ},

		// count with 1 argument
		{`count([])`, 0, object.NUMBER_OBJ},
		{`count([79, 7, 9])`, 3, object.NUMBER_OBJ},
		{`count([true, false, null])`, 1, object.NUMBER_OBJ},
		{`count([true, false, null, 1, "abc", []])`, 3, object.NUMBER_OBJ},

		{`count({1: false, 2: "abc", 3: 123})`, 2, object.NUMBER_OBJ},
		{`count({1: [], 2: null, 3: false})`, 0, object.NUMBER_OBJ},
	}

	runVmTests(t, tests, false, false)
}

func TestMoreAndLessFunctions(t *testing.T) {
	tests := []vmTestCase{
		// more
		{`more({2: 1}, {7: 14})`,
			[][]object.Object{
				{object.NumberFromInt(2), object.NumberFromInt(1)},
				{object.NumberFromInt(7), object.NumberFromInt(14)},
			},
			object.HASH_OBJ,
		},

		{`more({2: 1}, {7: 14}, {1: 17, 3: 15})`,
			[][]object.Object{
				{object.NumberFromInt(1), object.NumberFromInt(17)},
				{object.NumberFromInt(2), object.NumberFromInt(1)},
				{object.NumberFromInt(3), object.NumberFromInt(15)},
				{object.NumberFromInt(7), object.NumberFromInt(14)},
			},
			object.HASH_OBJ,
		},

		// less
		{`less({34: 890, 45: 23}, of=34)`,
			[][]object.Object{
				{object.NumberFromInt(45), object.NumberFromInt(23)},
			},
			object.HASH_OBJ,
		},
		{`less({34: 890, 45: 23}, of=[34, 45])`,
			[][]object.Object{},
			object.HASH_OBJ,
		},

		{`less([4, 5, 6, 7])`,
			[]int{4, 5, 6},
			object.LIST_OBJ,
		},
		{`less([4, 5, 6, 7], of=1)`,
			[]int{5, 6, 7},
			object.LIST_OBJ,
		},
		{`less([4, 5, 6, 7], of=-1)`,
			[]int{4, 5, 6},
			object.LIST_OBJ,
		},
		{`less([4, 5, 6, 7], of=3)`,
			[]int{4, 5, 7},
			object.LIST_OBJ,
		},
		{`less([4, 5, 6, 7], of=3..4)`,
			[]int{4, 5},
			object.LIST_OBJ,
		},
		{`less([4, 5, 6, 7], of=[1, 4])`,
			[]int{5, 6},
			object.LIST_OBJ,
		},
		{`less([1, 2, 3, 4, 5, 6, 7], of=[1, 4..6])`,
			[]int{2, 3, 7},
			object.LIST_OBJ,
		},
		{`less([1, 2, 3, 4, 5, 6, 7], of=[-1, 2, -3])`,
			[]int{1, 3, 4, 6},
			object.LIST_OBJ,
		},

		{`less("ΑΣΔΦΞΚΛ")`,
			"ΑΣΔΦΞΚ",
			object.STRING_OBJ,
		},
		{`less("ΑΣΔΦΞΚΛ", of=2)`,
			"ΑΔΦΞΚΛ",
			object.STRING_OBJ,
		},
		{`less("ΑΣΔΦΞΚΛ", of=-2)`,
			"ΑΣΔΦΞΛ",
			object.STRING_OBJ,
		},
		{`less("ΑΣΔΦΞΚΛ", of=2..6)`,
			"ΑΛ",
			object.STRING_OBJ,
		},
		{`less("ΑΣΔΦΞΚΛ", of=-2..6)`,
			"ΑΣΔΦΞΛ",
			object.STRING_OBJ,
		},
		{`less("ΑΣΔΦΞΚΛ", of=-2..7)`,
			"ΑΣΔΦΞ",
			object.STRING_OBJ,
		},
		{`less("ΑΣΔΦΞΚΛ", of=-4..-5)`,
			"ΑΣΞΚΛ",
			object.STRING_OBJ,
		},
		{`less("ΑΣΔΦΞΚΛ", of=[3, 7])`,
			"ΑΣΦΞΚ",
			object.STRING_OBJ,
		},
		{`less("ΑΣΔΦΞΚΛ", of=[3, 5..6])`,
			"ΑΣΦΛ",
			object.STRING_OBJ,
		},
	}

	runVmTests(t, tests, false, false)
}

func TestStringConversionFunctions(t *testing.T) {
	tests := []vmTestCase{
		// ucase, lcase, tcase
		{`lcase("LANGUR")`, "langur", object.STRING_OBJ},
		{`lcase("langur")`, "langur", object.STRING_OBJ},
		{`ucase("LANGUR")`, "LANGUR", object.STRING_OBJ},
		{`ucase("langur")`, "LANGUR", object.STRING_OBJ},
		{`tcase("LANGUR")`, "LANGUR", object.STRING_OBJ},
		{`tcase("langur")`, "LANGUR", object.STRING_OBJ},

		{`lcase(65)`, 97, object.NUMBER_OBJ},
		{`lcase(64)`, 64, object.NUMBER_OBJ},
		{`ucase(97)`, 65, object.NUMBER_OBJ},
		{`ucase(64)`, 64, object.NUMBER_OBJ},
		{`tcase(97)`, 65, object.NUMBER_OBJ},
		{`tcase(64)`, 64, object.NUMBER_OBJ},

		{`tcase(452)`, 453, object.NUMBER_OBJ},
		{`tcase(454)`, 453, object.NUMBER_OBJ},
		{`ucase(453)`, 452, object.NUMBER_OBJ},
		{`lcase(452)`, 454, object.NUMBER_OBJ},
		{`lcase(453)`, 454, object.NUMBER_OBJ},

		// trim
		{`trim(" langur ")`, "langur", object.STRING_OBJ},
		{`ltrim(" langur ")`, "langur ", object.STRING_OBJ},
		{`rtrim(" langur ")`, " langur", object.STRING_OBJ},

		// cp2s, s2cp, s2s, s2gc
		{`cp2s(s2cp("ΑΣΔΦΞΚΛ"))`,
			"ΑΣΔΦΞΚΛ",
			object.STRING_OBJ,
		},
		{`cp2s(s2cp("ΑΣΔΦΞΚΛ", of=2..3))`,
			"ΣΔ",
			object.STRING_OBJ,
		},

		{`cp2s(s2cp("ΑΣΔΦΞΚΛ", of=2..3, alt=[97, 98]))`,
			"ΣΔ",
			object.STRING_OBJ,
		},
		{`cp2s(s2cp("ΑΣΔΦΞΚΛ", of=2..10, alt=[97, 98]))`,
			"ab",
			object.STRING_OBJ,
		},

		{`s2s("ΑΣΔΦΞΚΛ", of=1..7)`,
			"ΑΣΔΦΞΚΛ",
			object.STRING_OBJ,
		},
		{`s2s("ΑΣΔΦΞΚΛ", of=2..3)`,
			"ΣΔ",
			object.STRING_OBJ,
		},
		{`s2s("ΑΣΔΦΞΚΛ", of=2..3, alt="ac")`,
			"ΣΔ",
			object.STRING_OBJ,
		},
		{`s2s("ΑΣΔΦΞΚΛ", of=2..10, alt="ac")`,
			"ac",
			object.STRING_OBJ,
		},

		{`cp2s(97..100)`, "abcd", object.STRING_OBJ},
		{`cp2s(63..58)`, "?>=<;:", object.STRING_OBJ},
		{`cp2s([97..100, 63..58])`, "abcd?>=<;:", object.STRING_OBJ},
		{`cp2s([97..100, [63, 58..59]])`, "abcd?:;", object.STRING_OBJ},

		// s2b, b2s
		{`s2b("abc")`,
			[]int{97, 98, 99},
			object.LIST_OBJ,
		},
		{`b2s(s2b("abc"))`,
			"abc",
			object.STRING_OBJ,
		},

		{`s2b("ΑΣΔΦΞΚΛ")`,
			[]int{0xce, 0x91, 0xce, 0xa3, 0xce, 0x94, 0xce, 0xa6, 0xce, 0x9e, 0xce, 0x9a, 0xce, 0x9b},
			object.LIST_OBJ,
		},
		{`b2s([16xce, 16x91, 16xce, 16xa3, 16xce, 16x94, 16xce, 16xa6, 16xce, 16x9e, 16xce, 16x9a, 16xce, 16x9b])`,
			"ΑΣΔΦΞΚΛ",
			object.STRING_OBJ,
		},
		{`b2s(s2b("ΑΣΔΦΞΚΛ"))`,
			"ΑΣΔΦΞΚΛ",
			object.STRING_OBJ,
		},

		// s2gc
		{`cp2s(reverse(s2gc("Aa\u0341b")))`,
			"ba\u0341A",
			object.STRING_OBJ,
		},

		// s2n: string or code point to numbers; interpreted from base 36
		{
			`s2n('7')`,
			7,
			object.NUMBER_OBJ,
		},
		{
			`s2n('a')`,
			10,
			object.NUMBER_OBJ,
		},
		{
			`s2n('A')`,
			10,
			object.NUMBER_OBJ,
		},
		{
			`s2n('z')`,
			35,
			object.NUMBER_OBJ,
		},
		{
			`s2n('Z')`,
			35,
			object.NUMBER_OBJ,
		},
		{
			`s2n("123XYZ")`,
			[]int{1, 2, 3, 33, 34, 35},
			object.LIST_OBJ,
		},
		{
			`s2n("abcd")`,
			[]int{10, 11, 12, 13},
			object.LIST_OBJ,
		},
		{
			`s2n("0aBCdE")`,
			[]int{0, 10, 11, 12, 13, 14},
			object.LIST_OBJ,
		},
	}

	runVmTests(t, tests, false, false)
}

func TestMathFunctions(t *testing.T) {
	tests := []vmTestCase{
		// min
		{`min([3, 56, 12, -1, 34, 5])`, "-1", object.NUMBER_OBJ},
		{`min([3, 56, 12, 34, 5])`, "3", object.NUMBER_OBJ},
		{`min({1: 7, 2: 14, 3: 0})`, "0", object.NUMBER_OBJ},
		{`min({1: 7, 2: -14, 3: 0})`, "-14", object.NUMBER_OBJ},
		{`min(4..10)`, "4", object.NUMBER_OBJ},
		{`min(10..4)`, "4", object.NUMBER_OBJ},
		{`min("asdf")`, "97", object.NUMBER_OBJ},
		{`min(["abc", "ABC"])`, "ABC", object.STRING_OBJ},

		// min for positional parameter counts
		{`min(fn{})`, "0", object.NUMBER_OBJ},
		{`min(fn(x) {x})`, "1", object.NUMBER_OBJ},
		{`min(fn(x, y) { x + y })`, "2", object.NUMBER_OBJ},
		{`min(map)`, 1, object.NUMBER_OBJ},

		// max
		{`max([3, 56, 12, -1, 34, 5])`, "56", object.NUMBER_OBJ},
		{`max([3, 56, 12, 34, 56.5, 5])`, "56.5", object.NUMBER_OBJ},
		{`max({1: 7, 2: 14, 3: 0})`, "14", object.NUMBER_OBJ},
		{`max({1: 7, 2: -14, 3: 0})`, "7", object.NUMBER_OBJ},
		{`max(4..10)`, "10", object.NUMBER_OBJ},
		{`max(10..4)`, "10", object.NUMBER_OBJ},
		{`max("asdf")`, "115", object.NUMBER_OBJ},
		{`max(["abc", "ABC"])`, "abc", object.STRING_OBJ},

		// max for positional parameter counts
		{`max(fn{})`, "0", object.NUMBER_OBJ},
		{`max(fn(x) { x })`, "1", object.NUMBER_OBJ},
		{`max(fn(x, y) {x + y})`, "2", object.NUMBER_OBJ},
		{`max(map)`, -1, object.NUMBER_OBJ},

		// minmax
		{"minmax([3, 56, 12, -1, 34, 5])", []int64{-1, 56}, object.RANGE_OBJ},
		{"minmax([3, 56, 12, 34, 5])", []int64{3, 56}, object.RANGE_OBJ},
		{"minmax([56, 12, 7777, 34, 5])", []int64{5, 7777}, object.RANGE_OBJ},
		{"minmax({1: 7, 2: 14, 3: 0})", []int64{0, 14}, object.RANGE_OBJ},
		{"minmax(30..40)", []int64{30, 40}, object.RANGE_OBJ},
		{"minmax(40..30)", []int64{30, 40}, object.RANGE_OBJ},
		{`minmax("asdf")`, []int64{97, 115}, object.RANGE_OBJ},

		// minmax for positional parameter counts
		{`minmax(fn{})`, []int64{0, 0}, object.RANGE_OBJ},
		{`minmax(fn(x) { x })`, []int64{1, 1}, object.RANGE_OBJ},
		{`minmax(fn(x, y) {x + y})`, []int64{2, 2}, object.RANGE_OBJ},
		{`minmax(map)`, []int64{1, -1}, object.RANGE_OBJ},

		// mid
		{`mid([1, 2])`, "1.5", object.NUMBER_OBJ},
		{`mid([0, 7, 20, 3])`, "10", object.NUMBER_OBJ},
		{`mid(1..2)`, "1.5", object.NUMBER_OBJ},
		{`mid(0..20)`, "10", object.NUMBER_OBJ},
		{`mid({1: 1, 10: 2})`, "1.5", object.NUMBER_OBJ},
		{`mid({1: 0, 2: 7, 3: 20, 4: 3})`, "10", object.NUMBER_OBJ},

		// mean
		{`mean([1, 2])`, "1.5", object.NUMBER_OBJ},
		{`mean([0, 7, 20, 3])`, "7.5", object.NUMBER_OBJ},
		{`mean(1..2)`, "1.5", object.NUMBER_OBJ},
		{`mean(0..20)`, "10", object.NUMBER_OBJ},
		{`mean({1: 1, 10: 2})`, "1.5", object.NUMBER_OBJ},
		{`mean({1: 0, 2: 7, 3: 20, 4: 3})`, "7.5", object.NUMBER_OBJ},

		// truncate without padding zeroes
		{`trunc(123)`, 123, object.NUMBER_OBJ},
		{`trunc(123.0)`, 123, object.NUMBER_OBJ},
		{`trunc(123, places=2, zeroes=false)`, 123, object.NUMBER_OBJ},
		{`trunc(123.123456789)`, 123, object.NUMBER_OBJ},
		{`trunc(123.123456789, places=4, zeroes=false)`, "123.1234", object.NUMBER_OBJ},
		{`trunc(123.123456789, places=8, zeroes=false)`, "123.12345678", object.NUMBER_OBJ},
		{`trunc(123.123456789, places=9, zeroes=false)`, "123.123456789", object.NUMBER_OBJ},
		{`trunc(123.123456789, places=10, zeroes=false)`, "123.123456789", object.NUMBER_OBJ},
		{`trunc(123.123456789, places=12, zeroes=false)`, "123.123456789", object.NUMBER_OBJ},

		// truncate with padding zeros
		{`trunc(123, places=0)`, 123, object.NUMBER_OBJ},
		{`trunc(123.0, places=0)`, 123, object.NUMBER_OBJ},
		{`trunc(123, places=2)`, "123.00", object.NUMBER_OBJ},
		{`trunc(123.123456789, places=0)`, 123, object.NUMBER_OBJ},
		{`trunc(123.123456789, places=4)`, "123.1234", object.NUMBER_OBJ},
		{`trunc(123.123456789, places=8)`, "123.12345678", object.NUMBER_OBJ},
		{`trunc(123.123456789, places=9)`, "123.123456789", object.NUMBER_OBJ},
		{`trunc(123.123456789, places=10)`, "123.1234567890", object.NUMBER_OBJ},
		{`trunc(123.123456789, places=12)`, "123.123456789000", object.NUMBER_OBJ},

		// zeroes true, false, null
		{`trunc(123.200, places=7, zeroes=true)`, "123.2000000", object.NUMBER_OBJ},
		{`trunc(123.200, places=7, zeroes=false)`, "123.2", object.NUMBER_OBJ},
		{`trunc(123.200, places=7, zeroes=null)`, "123.200", object.NUMBER_OBJ},
		{`trunc(123.222, places=7, zeroes=true)`, "123.2220000", object.NUMBER_OBJ},
		{`trunc(123.222, places=7, zeroes=false)`, "123.222", object.NUMBER_OBJ},
		{`trunc(123.222, places=7, zeroes=null)`, "123.222", object.NUMBER_OBJ},

		// truncate on integer
		{`trunc(123, places=-1)`, 120, object.NUMBER_OBJ},
		{`trunc(123.4, places=-1)`, 120, object.NUMBER_OBJ},
		{`trunc(123, places=-2)`, 100, object.NUMBER_OBJ},
		{`trunc(123.4, places=-2)`, 100, object.NUMBER_OBJ},
		{`trunc(123, places=-3)`, 0, object.NUMBER_OBJ},
		{`trunc(123.4, places=-3)`, 0, object.NUMBER_OBJ},
		{`trunc(153, places=-2)`, 100, object.NUMBER_OBJ},
		{`trunc(153.4, places=-2)`, 100, object.NUMBER_OBJ},

		// round without padding zeroes
		{`round(123)`, 123, object.NUMBER_OBJ},
		{`round(123.4)`, 123, object.NUMBER_OBJ},
		{`round(123.7)`, 124, object.NUMBER_OBJ},
		{`round(123.0)`, 123, object.NUMBER_OBJ},
		{`round(123, places=2, zeroes=false)`, 123, object.NUMBER_OBJ},
		{`round(123.123456789)`, 123, object.NUMBER_OBJ},
		{`round(123.123456789, places=4, zeroes=false)`, "123.1235", object.NUMBER_OBJ},
		{`round(123.123456789, places=9, zeroes=false)`, "123.123456789", object.NUMBER_OBJ},
		{`round(123.123456789, places=10, zeroes=false)`, "123.123456789", object.NUMBER_OBJ},
		{`round(123.123456789, places=12, zeroes=false)`, "123.123456789", object.NUMBER_OBJ},

		// round with padding zeroes
		{`round(123, places=0)`, 123, object.NUMBER_OBJ},
		{`round(123.4, places=0)`, 123, object.NUMBER_OBJ},
		{`round(123.7, places=0)`, 124, object.NUMBER_OBJ},
		{`round(123.0, places=0)`, 123, object.NUMBER_OBJ},
		{`round(123, places=2)`, "123.00", object.NUMBER_OBJ},
		{`round(123.123456789, places=0)`, 123, object.NUMBER_OBJ},
		{`round(123.123456789, places=4)`, "123.1235", object.NUMBER_OBJ},
		{`round(123.123456789, places=9)`, "123.123456789", object.NUMBER_OBJ},
		{`round(123.123456789, places=10)`, "123.1234567890", object.NUMBER_OBJ},
		{`round(123.123456789, places=12)`, "123.123456789000", object.NUMBER_OBJ},

		// zeroes true, false, null
		{`round(123.200, places=7, zeroes=true)`, "123.2000000", object.NUMBER_OBJ},
		{`round(123.200, places=7, zeroes=false)`, "123.2", object.NUMBER_OBJ},
		{`round(123.200, places=7, zeroes=null)`, "123.200", object.NUMBER_OBJ},
		{`round(123.222, places=7, zeroes=true)`, "123.2220000", object.NUMBER_OBJ},
		{`round(123.222, places=7, zeroes=false)`, "123.222", object.NUMBER_OBJ},
		{`round(123.222, places=7, zeroes=null)`, "123.222", object.NUMBER_OBJ},

		// round on integer
		{`round(123, places=-1)`, 120, object.NUMBER_OBJ},
		{`round(123.4, places=-1)`, 120, object.NUMBER_OBJ},
		{`round(123, places=-2)`, 100, object.NUMBER_OBJ},
		{`round(123.4, places=-2)`, 100, object.NUMBER_OBJ},
		{`round(123, places=-3)`, 0, object.NUMBER_OBJ},
		{`round(123.4, places=-3)`, 0, object.NUMBER_OBJ},
		{`round(153, places=-2)`, 200, object.NUMBER_OBJ},
		{`round(153.4, places=-2)`, 200, object.NUMBER_OBJ},

		// test passing rounding mode to round() function
		{`round(123.5, places=0, mode=_round'halfeven)`, 124, object.NUMBER_OBJ},
		{`round(122.5, places=0, mode=_round'halfeven)`, 122, object.NUMBER_OBJ},
		{`round(123.5, places=0, mode=_round'halfawayfrom0)`, 124, object.NUMBER_OBJ},
		{`round(122.5, places=0, mode=_round'halfawayfrom0)`, 123, object.NUMBER_OBJ},

		{`val x = round(2.5, places=0, mode=_round'halfeven)
		  val y = round(2.5, places=0, mode=_round'halfawayfrom0)
		  val z = round(4.45, places=1, mode=_round'halfeven)
		  x + y + z`, "9.4", object.NUMBER_OBJ},

		// alternating rounding modes
		{`mode rounding = _round'halfeven
		  val x = round(2.5)
		  mode rounding = _round'halfawayfrom0
		  val y = round(2.5)
		  x + y`, 5, object.NUMBER_OBJ},

		{`mode rounding = _round'halfeven
		  val x = round(2.5)
		  mode rounding = _round'halfawayfrom0
		  val y = round(2.5)
		  mode rounding = _round'halfeven
		  val z = round(4.45, places=1)
		  x + y + z`, "9.4", object.NUMBER_OBJ},

		// gcd
		{`gcd([64, 32])`, 32, object.NUMBER_OBJ},
		{`gcd([16, 64, 32])`, 16, object.NUMBER_OBJ},
		{`gcd([24, 64, 32])`, 8, object.NUMBER_OBJ},
		{`gcd([24, 64, 32, 1024])`, 8, object.NUMBER_OBJ},
		{`gcd([25, 15])`, 5, object.NUMBER_OBJ},
		{`gcd([25, 17])`, 1, object.NUMBER_OBJ},
		{`gcd([25, 17, 15])`, 1, object.NUMBER_OBJ},

		{`gcd([861, 1113])`, 21, object.NUMBER_OBJ},
		{`gcd([1113, 861])`, 21, object.NUMBER_OBJ},
		{`gcd([1113, 789789777])`, 21, object.NUMBER_OBJ},
		{`gcd([789789777, 1113])`, 21, object.NUMBER_OBJ},

		// if truly works with more than 2, should be able to rearrange and get same result
		{`gcd([64, 24, 32])`, 8, object.NUMBER_OBJ},
		{`gcd([32, 64, 24])`, 8, object.NUMBER_OBJ},
		{`gcd([32, 64, 24, 1024])`, 8, object.NUMBER_OBJ},
		{`gcd([1024, 64, 32, 24])`, 8, object.NUMBER_OBJ},

		{`gcd([861, 1113, 42])`, 21, object.NUMBER_OBJ},
		{`gcd([1113, 861, 42])`, 21, object.NUMBER_OBJ},
		{`gcd([42, 1113, 861])`, 21, object.NUMBER_OBJ},

		{`gcd([42, 1113, 7])`, 7, object.NUMBER_OBJ},
		{`gcd([42, 1113, 861, 21])`, 21, object.NUMBER_OBJ},

		{`gcd([861, 1113, 41])`, 1, object.NUMBER_OBJ},
		{`gcd([1113, 861, 41])`, 1, object.NUMBER_OBJ},
		{`gcd([41, 1113, 861])`, 1, object.NUMBER_OBJ},

		// lcm
		{`lcm([1, 2])`, 2, object.NUMBER_OBJ},
		{`lcm([3, 9])`, 9, object.NUMBER_OBJ},
		{`lcm([3, 7])`, 21, object.NUMBER_OBJ},
		{`lcm([9, 3])`, 9, object.NUMBER_OBJ},
		{`lcm([7, 3])`, 21, object.NUMBER_OBJ},
		{`lcm([1, 2, 8, 3])`, 24, object.NUMBER_OBJ},
		{`lcm([1, 2, 8, 3])`, 24, object.NUMBER_OBJ},
		{`lcm([2, 7, 3, 9, 4])`, 252, object.NUMBER_OBJ},
		{`lcm([12, 15, 10, 75])`, 300, object.NUMBER_OBJ},
	}

	runVmTests(t, tests, false, false)
}

func TestTrigonometryFunctions(t *testing.T) {
	// may need revision if the implementation changes
	tests := []vmTestCase{
		{`sin(0)`, 0, object.NUMBER_OBJ},
		{`simplify(sin(0.1))`, "0.099833416646828152572461017173779584656", object.NUMBER_OBJ},
		{`simplify(sin(0.7))`, "0.644217687237691058301057310266969994992", object.NUMBER_OBJ},
		{`simplify(sin(777))`, "-0.85555119309214226986078503275661905999367629103431303756153801046391965453583965662742470367595236888969155024663158003487124467067994824234713922520290687807900706449633582331336174645559874922595769975789002846836398973209171687155587553719175786849797776390235555598637159025563776602157894412346084150471137245038441572723491633575836821432745114375825788867188494987568516322721607040276986355987893081729739523472864573436269521531982421875", object.NUMBER_OBJ},

		{`simplify(cos(0))`, 1, object.NUMBER_OBJ},
		{`simplify(cos(0.1))`, "0.99500416527802576608985619424209225912318", object.NUMBER_OBJ},
		{`simplify(cos(0.7))`, "0.76484218728448842674297597659422650053982", object.NUMBER_OBJ},
		{`simplify(cos(777))`, "-0.5177182206554178203639958289014534248739119197942031270107690298334799415007007493695487850601437860368814347086122102956576793825542159782548317881872096475162194440297207454396108284484038090083185218018846646878174201136372602881999039658327656057026499717825509626244148721224284961962441925104056517782875917849352703577942213917986399590621529709815250599585120368819671079214424037460510895950919640625", object.NUMBER_OBJ},

		{`tan(0)`, 0, object.NUMBER_OBJ},
		{`simplify(tan(0.1))`, "0.1003346720854505450570729988386166", object.NUMBER_OBJ},
		{`simplify(tan(0.7))`, "0.8422883804630794435307217508544428", object.NUMBER_OBJ},
		{`simplify(tan(777))`, "1.652542172475669651802597792023833", object.NUMBER_OBJ},

		{`atan(0)`, 0, object.NUMBER_OBJ},
		{`simplify(atan(0.1))`, "0.0996686524911620275499155861903632", object.NUMBER_OBJ},
		{`simplify(atan(0.7))`, "0.610725964389208587959903042496619943574322890974527438702908613132", object.NUMBER_OBJ},
		{`simplify(atan(777))`, "1.569509326218479000787624487068413876447717987705425760996466178539", object.NUMBER_OBJ},
	}

	runVmTests(t, tests, false, false)
}

func TestOperatorImpliedFunctions(t *testing.T) {
	tests := []vmTestCase{
		{`fold([1, 2, 3, 4], by=fn{+})`,
			10, object.NUMBER_OBJ,
		},
		{`fold([1, 2, 3, 4], by=fn{*})`,
			24, object.NUMBER_OBJ,
		},
		{`fold([16, 2, 2], by=fn{^/})`,
			2, object.NUMBER_OBJ,
		},
		{`fold(["w", "h", "a", "t", "?"], by=fn{~})`,
			"what?", object.STRING_OBJ,
		},
		{`fold([true, true, false], by=fn{and})`,
			false, object.BOOLEAN_OBJ,
		},
		{`fold([true, true, false], by=fn{or?})`,
			true, object.BOOLEAN_OBJ,
		},
		{`fold([true, true, null], by=fn{or?})`,
			nil, object.NULL_OBJ,
		},
		{`fold( [true, true, null], by=fn{or})`,
			true, object.BOOLEAN_OBJ,
		},
		{`fold([true, true, false], by=fn{xor})`,
			false, object.BOOLEAN_OBJ,
		},
		{`fold([true, false, false], by=fn{xor})`,
			true, object.BOOLEAN_OBJ,
		},
	}

	runVmTests(t, tests, false, false)
}

func TestNilLeftPartiallyImpliedFunctions(t *testing.T) {
	tests := []vmTestCase{
		{`map([1, 3, 9], by=fn{+7})`,
			[]int{8, 10, 16}, object.LIST_OBJ,
		},
		{`map([1, 3, 9], by=fn{-7})`,
			[]int{-6, -4, 2}, object.LIST_OBJ,
		},
		{`map([1, 3, 9], by=fn{* 7})`,
			[]int{7, 21, 63}, object.LIST_OBJ,
		},

		{`val f = fn{nor false}; f(false)`,
			true, object.BOOLEAN_OBJ,
		},
		{`val f = fn{nor false}; f(true)`,
			false, object.BOOLEAN_OBJ,
		},
		{`val f = fn{xor false}; f(true)`,
			true, object.BOOLEAN_OBJ,
		},
		{ // as a closure
			`val x = false; val f = fn{xor x}; f(true)`,
			true, object.BOOLEAN_OBJ,
		},
	}

	runVmTests(t, tests, false, false)
}

func TestIsAndIsNotOperators(t *testing.T) {
	tests := []vmTestCase{
		{`123 is number`,
			true, object.BOOLEAN_OBJ},
		{`123 is not number`,
			false, object.BOOLEAN_OBJ},
		{`123 is list`,
			false, object.BOOLEAN_OBJ},
		{`val x = 123
			x is number`,
			true, object.BOOLEAN_OBJ},
		{`val x = 123
			x is not number`,
			false, object.BOOLEAN_OBJ},

		{`123+1i is complex`,
			true, object.BOOLEAN_OBJ},
		{`123+1i is not complex`,
			false, object.BOOLEAN_OBJ},
		{`123+1i is list`,
			false, object.BOOLEAN_OBJ},
		{`val x = 123+1i
			x is complex`,
			true, object.BOOLEAN_OBJ},
		{`val x = 123+1i
			x is not complex`,
			false, object.BOOLEAN_OBJ},

		{`1..2 is range`,
			true, object.BOOLEAN_OBJ},
		{`1..2 is not range`,
			false, object.BOOLEAN_OBJ},
		{`1..2 is list`,
			false, object.BOOLEAN_OBJ},
		{`val x = 1..2
			x is range`,
			true, object.BOOLEAN_OBJ},
		{`val x = 1..2
			x is not range`,
			false, object.BOOLEAN_OBJ},

		{`true is bool`,
			true, object.BOOLEAN_OBJ},
		{`true is not bool`,
			false, object.BOOLEAN_OBJ},
		{`true is list`,
			false, object.BOOLEAN_OBJ},
		{`val x = true
			x is bool`,
			true, object.BOOLEAN_OBJ},
		{`val x = true
			x is not bool`,
			false, object.BOOLEAN_OBJ},

		{`"asdf" is string`,
			true, object.BOOLEAN_OBJ},
		{`"asdf" is not string`,
			false, object.BOOLEAN_OBJ},
		{`"asdf" is list`,
			false, object.BOOLEAN_OBJ},
		{`val x = "asdf"
			x is string`,
			true, object.BOOLEAN_OBJ},
		{`val x = "asdf"
			x is not string`,
			false, object.BOOLEAN_OBJ},

		{`re// is regex`,
			true, object.BOOLEAN_OBJ},
		{`re// is not regex`,
			false, object.BOOLEAN_OBJ},
		{`re// is list`,
			false, object.BOOLEAN_OBJ},
		{`val x = re//
			x is regex`,
			true, object.BOOLEAN_OBJ},
		{`val x = re//
			x is not regex`,
			false, object.BOOLEAN_OBJ},

		{`dt// is datetime`,
			true, object.BOOLEAN_OBJ},
		{`dt// is not datetime`,
			false, object.BOOLEAN_OBJ},
		{`dt// is list`,
			false, object.BOOLEAN_OBJ},
		{`val x = dt//
			x is datetime`,
			true, object.BOOLEAN_OBJ},
		{`val x = dt//
			x is not datetime`,
			false, object.BOOLEAN_OBJ},

		{`dr// is duration`,
			true, object.BOOLEAN_OBJ},
		{`dr// is not duration`,
			false, object.BOOLEAN_OBJ},
		{`dr// is list`,
			false, object.BOOLEAN_OBJ},
		{`val x = dr//
			x is duration`,
			true, object.BOOLEAN_OBJ},
		{`val x = dr//
			x is not duration`,
			false, object.BOOLEAN_OBJ},

		{`[] is list`,
			true, object.BOOLEAN_OBJ},
		{`[] is not list`,
			false, object.BOOLEAN_OBJ},
		{`[] is number`,
			false, object.BOOLEAN_OBJ},
		{`val x = []
			x is list`,
			true, object.BOOLEAN_OBJ},
		{`val x = []
			x is not list`,
			false, object.BOOLEAN_OBJ},

		{`{:} is hash`,
			true, object.BOOLEAN_OBJ},
		{`{:} is not hash`,
			false, object.BOOLEAN_OBJ},
		{`{:} is list`,
			false, object.BOOLEAN_OBJ},
		{`val x = {:}
			x is hash`,
			true, object.BOOLEAN_OBJ},
		{`val x = {:}
			x is not hash`,
			false, object.BOOLEAN_OBJ},

		{`fn{+} is fn`,
			true, object.BOOLEAN_OBJ},
		{`fn{+} is not fn`,
			false, object.BOOLEAN_OBJ},
		{`mapX is fn`,
			true, object.BOOLEAN_OBJ},
		{`mapX is not fn`,
			false, object.BOOLEAN_OBJ},
		{`12 is fn`,
			false, object.BOOLEAN_OBJ},
		{`12 is not fn`,
			true, object.BOOLEAN_OBJ},
	}

	runVmTests(t, tests, false, false)
}

func TestTypeConversion(t *testing.T) {
	tests := []vmTestCase{
		{`bool(0)`,
			false, object.BOOLEAN_OBJ},
		{`bool(123)`,
			true, object.BOOLEAN_OBJ},
		{`bool("")`,
			false, object.BOOLEAN_OBJ},
		{`bool("A")`,
			true, object.BOOLEAN_OBJ},

		{`string(123)`,
			"123", object.STRING_OBJ},
		{`number("123.45")`,
			"123.45", object.NUMBER_OBJ},
		{`complex(123.45, 7)`,
			"123.45+7i", object.COMPLEX_OBJ},

		{`string(65519, fmt=16)`,
			"ffef", object.STRING_OBJ},
		{`number("ffef", fmt=16)`,
			65519, object.NUMBER_OBJ},

		{`duration(0) is duration`,
			true, object.BOOLEAN_OBJ},
		{`datetime(0) is datetime`,
			true, object.BOOLEAN_OBJ},
		{`hash(dt//) is hash`,
			true, object.BOOLEAN_OBJ},

		// unbounded list
		{`string 123`,
			"123", object.STRING_OBJ},
		{`number string(123)`,
			"123", object.NUMBER_OBJ},

		{`len string(123, fmt=16)`,
			2, object.NUMBER_OBJ},
	}

	runVmTests(t, tests, false, false)
}

func TestExecT(t *testing.T) {
	testStr := `execT("ls")`
	if system.Type == system.WINDOWS {
		testStr = `execT("dir")`
	}
	result := oneResult(t, 1, testStr, false, false)

	s, ok := result.(*object.String)
	if !ok {
		t.Fatalf("Expected String object, received=%s", result.TypeString())
	}

	if !strings.Contains(s.String(), "vm_test.go") {
		t.Errorf(`String returned did not contain "vm_test.go"`)
	}
}

// func TestParameterMutability(t *testing.T) {
// 	tests := []vmTestCase{
// 		{`
// 			fn(var x, var y) {
// 				if y > 10 { return x + y }
// 				x += 2; y += 1
// 				fn((x, y))
// 			}(7, 0)
// 			`,
// 			"40",
// 			object.NUMBER_OBJ,
// 		},

// 		{`
// 			fn(x, var y) {
// 				y += 10
// 				x + y
// 			}(7, 3)
// 			`,
// 			"20",
// 			object.NUMBER_OBJ,
// 		},
// 	}

// 	runVmTests(t, tests, false, false)
// }

func TestTransliterate(t *testing.T) {
	tests := []vmTestCase{
		{`tran("abcd", by='a'..'z', with='A'..'Z')`, "ABCD", object.STRING_OBJ},
		{`tran("abcd", by=["bc", "d", "a"], with=["ED", " now", "A"])`, "AED now", object.STRING_OBJ},
		{`tran("abcd", by=["bc", "d", "a"], with=["ed", " now", "a"])`, "aed now", object.STRING_OBJ},
		{`tran("try this", by="something", with="SOMETHING")`, "Try THIS", object.STRING_OBJ},
		{`tran("/0no//way/1man", by=fw"// /0 /1", with=fw"/ ! ?")`, "!no/way?man", object.STRING_OBJ},
		{`tran("𝄞€ЖöA", by="AöЖ€𝄞", with='A'..'E')`, "EDCBA", object.STRING_OBJ},
		{`tran("CDAB", with="AöЖ€𝄞", by='A'..'E')`, "Ж€Aö", object.STRING_OBJ},

		// using hash
		{`tran("CDAB", with={"A": "A", "B": "ö", "C": "Ж", "D": "€", "E": "𝄞"})`, "Ж€Aö", object.STRING_OBJ},

		// with delimiter
		{`tran("abcd", by='a'..'z', with='A'..'Z', delim="7")`, "A7B7C7D", object.STRING_OBJ},
		{`tran("abc3d", by='a'..'z', with='A'..'Z', delim="7")`, "A7B7C737D", object.STRING_OBJ},

		{`tran("cab", with={"a": ".-", "b": "-...", "c": "-.-."}, delim=" ")`, "-.-. .- -...", object.STRING_OBJ},
		{`tran("cabs", with={"a": ".-", "b": "-...", "c": "-.-."}, delim=" ")`, "-.-. .- -... s", object.STRING_OBJ},
		{`tran("zcabs2", with={"a": ".-", "b": "-...", "c": "-.-."}, delim=" ")`, "z -.-. .- -... s2", object.STRING_OBJ},
	}

	runVmTests(t, tests, false, false)
}

func TestRe2(t *testing.T) {
	tests := []vmTestCase{
		{"re/abc/ is regex", true, object.BOOLEAN_OBJ},
		{"qs/abc/ is regex", false, object.BOOLEAN_OBJ},

		{`matching(" abc ", by=re/a.*c/)`, true, object.BOOLEAN_OBJ},
		{`matching(" abc ", by=re/a.*d/)`, false, object.BOOLEAN_OBJ},

		// regex functions accepting non-strings (auto-stringification)
		{`matching(123.0, by=RE/\d+\.\d+/)`, true, object.BOOLEAN_OBJ},
		{`matching(123, by=RE/\d+\.\d+/)`, false, object.BOOLEAN_OBJ},

		{`match(" abc ", by=re/a.*c/)`, "abc", object.STRING_OBJ},
		{`match(" abc ", by=re/a.*c/, alt=7)`, "abc", object.STRING_OBJ},
		{`match(" abc ", by=re/a.*d/)`, nil, object.NULL_OBJ},
		{`match(" abc ", by=re/a.*d/, alt=7)`, "7", object.NUMBER_OBJ},

		{`matches("abc azc aec ", by=re/a.*?c/)`, []string{"abc", "azc", "aec"}, object.LIST_OBJ},
		{`matches("abc azc aec ", by=re/a.*?z/)`, []string{"abc az"}, object.LIST_OBJ},
		{`matches("abc azc aec ", by=re/a.*?Z/)`, []string{}, object.LIST_OBJ},

		{`matches("abc azc aec ", by=re/^a.*?c/)`, []string{"abc"}, object.LIST_OBJ},
		{`matches("abc azc aec ", by=re/^a.*?c/, max=7)`, []string{"abc"}, object.LIST_OBJ},
		{`matches("abc azc aec ", by=re/^a.*?Z/)`, []string{}, object.LIST_OBJ},
		{`matches("abc azc aec ", by=re/[a-c]/, max=6)`, []string{"a", "b", "c", "a", "c", "a"}, object.LIST_OBJ},
		{`matches("abc azc aec ", by=re/[a-c]/, max=2)`, []string{"a", "b"}, object.LIST_OBJ},

		{`replace(" abc abc abc abc ", by=re/a.*?c/, with="7")`, " 7 7 7 7 ", object.STRING_OBJ},
		{`replace(" abc abc abc abc ", by=re/a.*?c/, with="7", max=1)`, " 7 abc abc abc ", object.STRING_OBJ},
		{`replace(" abc abc abc abc ", by=re/a.*?c/, with="7", max=2)`, " 7 7 abc abc ", object.STRING_OBJ},
		{`replace(" abc abc abc abc ", by=re/a.*?c/, with="7", max=3)`, " 7 7 7 abc ", object.STRING_OBJ},
		{`replace(" abc abc abc abc ", by=re/a.*?c/, with="7", max=4)`, " 7 7 7 7 ", object.STRING_OBJ},

		{`replace("abc abc abc abc ", by=re/^a.*?c/, with="7")`, "7 abc abc abc ", object.STRING_OBJ},
		{`replace("abc abc abc abc ", by=re/^a.*?c/, with="7", max=7)`, "7 abc abc abc ", object.STRING_OBJ},

		{`replace("abc azc aec afc ", by=re/a(.*?)c/, with="A$1", max=7)`, "Ab Az Ae Af ", object.STRING_OBJ},
		{`replace("abc azc aec afc ", by=re/a(.*?)c/, with="A$1", max=1)`, "Ab azc aec afc ", object.STRING_OBJ},
		{`replace("abc azc aec afc ", by=re/a(.*?)c/, with="A $1 Z")`, "A b Z A z Z A e Z A f Z ", object.STRING_OBJ},

		{`replace("abc azc aec afc ", by=re/a(.*?)c/, with="A$1", max=7, interp=false)`, "A$1 A$1 A$1 A$1 ", object.STRING_OBJ},

		// passing a function to replace
		{`replace("abc azc aec afc ", by=re/[e-z]/, with=fn(s) { ucase(s) } )`, "abc aZc aEc aFc ", object.STRING_OBJ},
		{`replace("abc azc aec afc ", by=re/./, with=fn(s) { ucase(s) } )`, "ABC AZC AEC AFC ", object.STRING_OBJ},
		{`replace("abc azc aec afc ", by=re/a./, with=fn(s) { s~"AAA" })`, "abAAAc azAAAc aeAAAc afAAAc ", object.STRING_OBJ},
		{`replace("abc azc aec afc ", by=re/[a-e]/, with=fn(s) {"Z"})`, "ZZZ ZzZ ZZZ ZfZ ", object.STRING_OBJ},
		{`replace("abc azc aec afc ", by=re/[a-e]/, with=fn(s) {"Z"}, max=2)`, "ZZc azc aec afc ", object.STRING_OBJ},
		{`replace("abc azc aec afc ", by=re/[a-e]/, with=fn(s) {"Z"}, max=1)`, "Zbc azc aec afc ", object.STRING_OBJ},
		{`replace("abc azc aec afc ", by=re/c/, with=fn(s) {"Z"}, max=1)`, "abZ azc aec afc ", object.STRING_OBJ},

		// passing multiple things to replace
		{`replace("abc azc aec afc ", by="c", with=["Z", _])`, "abZ azc aeZ afc ", object.STRING_OBJ},
		{`replace("abc azc aec afc ", by=re/[ab]/, with=[fn(s) { ucase s }, fn(s) { s~"Y" }])`, "AbYc Azc aYec Afc ", object.STRING_OBJ},
		{`replace("abc azc aec afc ", by=re/[ab]/, with=[fn(s) { ucase s }, fn(s) { s~"Y" }], max=2)`, "AbYc azc aec afc ", object.STRING_OBJ},
		{`replace("abc azc aec afc ", by=re/[ab]/, with=[fn(s) { ucase s }, fn(s) { s~"Y" }], max=3)`, "AbYc Azc aec afc ", object.STRING_OBJ},
		{`replace("abc azc aec afc ", by=re/[ab]/, with=["Z", fn(s) { ucase s }], interp=false)`, "ZBc Zzc Aec Zfc ", object.STRING_OBJ},
		// {`replace("abc azc aec afc ", by=re/[ab]/, with=["$1", fn(s) { ucase s }])`, "aBc azc Aec afc ", object.STRING_OBJ},

		// passing nothing to replace (no replacement string or function; default ZLS)
		{`replace("abc azc aec afc ", by=re/[e-z]/)`, "abc ac ac ac ", object.STRING_OBJ},

		{`split("5841755193", delim=re/[789]/)`, []string{"5", "41", "551", "3"}, object.LIST_OBJ},
		{`split("84175519", delim=re/[789]/)`, []string{"", "41", "551", ""}, object.LIST_OBJ},
		{`split("84175519", delim=re/[789]/, max=2)`, []string{"", "4175519"}, object.LIST_OBJ},
		{`split("84175519", delim=re/[789]/, max=1)`, []string{"84175519"}, object.LIST_OBJ},

		{`split("abc", delim=re//, max=4)`, []string{"a", "b", "c"}, object.LIST_OBJ},
		{`split("abc", delim=re//, max=3)`, []string{"a", "b", "c"}, object.LIST_OBJ},
		{`split("abc", delim=re//, max=2)`, []string{"a", "bc"}, object.LIST_OBJ},
		{`split("abc", delim=re//, max=1)`, []string{"abc"}, object.LIST_OBJ},
		{`split("abc", delim=re//, max=0)`, []string{}, object.LIST_OBJ},
		{`split("abc", delim=re//)`, []string{"a", "b", "c"}, object.LIST_OBJ},
		{`split("ασδ", delim="σ")`, []string{"α", "δ"}, object.LIST_OBJ},

		{`split("144000", delim=RE/\./)`, []string{"144000"}, object.LIST_OBJ},
		{`split("144.000", delim=RE/\./)`, []string{"144", "000"}, object.LIST_OBJ},

		{`split("don't.you.know", delim=RE/\./)`, []string{"don't", "you", "know"}, object.LIST_OBJ},
		{`split("don't.you.know", delim=RE/Z/)`, []string{"don't.you.know"}, object.LIST_OBJ},

		{`split("zzz", delim=re"a")`, []string{"zzz"}, object.LIST_OBJ},
		{`split("zzz", delim=re"a", max=0)`, []string{}, object.LIST_OBJ},
		{`split("zzz", delim="a")`, []string{"zzz"}, object.LIST_OBJ},
		{`split("zzz", delim="a", max=0)`, []string{}, object.LIST_OBJ},

		{`split("", delim=re"a")`, []string{""}, object.LIST_OBJ},
		{`split("", delim="a")`, []string{""}, object.LIST_OBJ},
		{`split("", delim=re"")`, []string{}, object.LIST_OBJ},
		{`split("", delim="")`, []string{}, object.LIST_OBJ},
		{`split("", delim=3)`, []string{}, object.LIST_OBJ}, // ?

		{`submatch("asdfzzto", by=re"(a.).+(zz)(t)")`, []string{"as", "zz", "t"}, object.LIST_OBJ},
		{`submatch("asdfzz", by=re"(a.).+(zz)(t)")`, []string{}, object.LIST_OBJ},

		{`submatchH(" abcd: peaceInJerusalem ", by=RE/(?P<key>\w+)\s*:\s*(?P<value>\w+)/)`,
			[][]object.Object{
				{object.NumberFromInt(0), object.NewString("abcd: peaceInJerusalem")},
				{object.NumberFromInt(1), object.NewString("abcd")},
				{object.NumberFromInt(2), object.NewString("peaceInJerusalem")},
				{object.NewString("key"), object.NewString("abcd")},
				{object.NewString("value"), object.NewString("peaceInJerusalem")},
			},
			object.HASH_OBJ,
		},

		{`submatches("asdfzzmnazmnopzz", by=re/(a.).+?(z)z/)`, [][]string{{"as", "z"}, {"az", "z"}}, object.LIST_OBJ},
		{`submatches("asdfmnazmnopzz", by=re/(a.).+?(z)z/)`, [][]string{{"as", "z"}}, object.LIST_OBJ},
		{`submatches("asdfmnazmnop", by=re/(a.).+?(z)z/)`, [][]string{}, object.LIST_OBJ},

		{`index("asdfzzto", by=re"(a.).+(zz)(t)")`, []int64{1, 7}, object.RANGE_OBJ},
		{`index("asdfzzto", by=re"zz")`, []int64{5, 6}, object.RANGE_OBJ},
		{`index("aaasdfzzto", by=re"z")`, []int64{7, 7}, object.RANGE_OBJ},

		{`indices(7467334300, by=re/[73]+/)`, [][]int64{{1, 1}, {4, 6}, {8, 8}}, object.LIST_OBJ},
		{`indices("cdefghabab12", by=re/ab/)`, [][]int64{{7, 8}, {9, 10}}, object.LIST_OBJ},
		{`indices("cdefghazaz12", by=re/ab/)`, [][]int64{}, object.LIST_OBJ},

		{`subindex("sdfbbzzmnazmnop", by=re/(a.)/)`, [][]int64{{10, 11}}, object.LIST_OBJ},
		{`subindex("asdfbbzzmnazmnopzz", by=re/(a.).+?(z)z/)`, [][]int64{{1, 2}, {7, 7}}, object.LIST_OBJ},
		{`subindex("sdfbbzzmnzmnop", by=re/(a.)/)`, [][]int64{}, object.LIST_OBJ},

		{`subindices("sdfbbzzmnazmnop", by=re/(a.)/)`, [][][]int64{{{10, 11}}}, object.LIST_OBJ},

		{`subindices("asdfbbzzmnazmnopzz", by=re/(a.).+?(z)z/)`, [][][]int64{{{1, 2}, {7, 7}}, {{11, 12}, {17, 17}}}, object.LIST_OBJ},
		{`subindices("asdfbbzzmnazmnopzz", by=re/(a.).+?(z)z/, max=1)`, [][][]int64{{{1, 2}, {7, 7}}}, object.LIST_OBJ},
		{`subindices("asdfbbzzmnazmnopzz", by=re/(a.).+?(z)z/, max=0)`, [][][]int64{}, object.LIST_OBJ},

		{`subindices("sdfbbzzmnazmnopzz", by=re/(a.).+?(z)z/)`, [][][]int64{{{10, 11}, {16, 16}}}, object.LIST_OBJ},
		{`subindices("sdfbbzzmnazmnop", by=re/(a.).+?(z)z/)`, [][][]int64{}, object.LIST_OBJ},

		{`reEsc(QS"\(abc)+")`, `\\\(abc\)\+`, object.STRING_OBJ},
		// including free-spacing meta-characters...
		{`reEsc("\\(\x0D\x0Aabc #)\x09+")`, `\\\(\r\nabc\ \#\)\t\+`, object.STRING_OBJ},

		// characters added for free-spacing mode escaped and handled well?
		{`matching("\n", by=reCompile("(?x:" ~ reEsc("\n") ~ ")"))`, true, object.BOOLEAN_OBJ},
		{`matching("", by=reCompile("(?x:" ~ reEsc("\n") ~ ")"))`, false, object.BOOLEAN_OBJ},
		{`matching("\n", by=reCompile(reEsc("\n")))`, true, object.BOOLEAN_OBJ},
		{`matching("", by=reCompile(reEsc("\n")))`, false, object.BOOLEAN_OBJ},
		{`matching("\r", by=reCompile("(?x:" ~ reEsc("\r") ~ ")"))`, true, object.BOOLEAN_OBJ},
		{`matching("", by=reCompile("(?x:" ~ reEsc("\r") ~ ")"))`, false, object.BOOLEAN_OBJ},
		{`matching("\r", by=reCompile(reEsc("\r")))`, true, object.BOOLEAN_OBJ},
		{`matching("", by=reCompile(reEsc("\r")))`, false, object.BOOLEAN_OBJ},
		{`matching("\t", by=reCompile("(?x:" ~ reEsc("\t") ~ ")"))`, true, object.BOOLEAN_OBJ},
		{`matching("", by=reCompile("(?x:" ~ reEsc("\t") ~ ")"))`, false, object.BOOLEAN_OBJ},
		{`matching("\t", by=reCompile(reEsc("\t")))`, true, object.BOOLEAN_OBJ},
		{`matching("", by=reCompile(reEsc("\t")))`, false, object.BOOLEAN_OBJ},
		{`matching(" ", by=reCompile("(?x:" ~ reEsc(" ") ~ ")"))`, true, object.BOOLEAN_OBJ},
		{`matching("", by=reCompile("(?x:" ~ reEsc(" ") ~ ")"))`, false, object.BOOLEAN_OBJ},
		{`matching(" ", by=reCompile(reEsc(" ")))`, true, object.BOOLEAN_OBJ},
		{`matching("", by=reCompile(reEsc(" ")))`, false, object.BOOLEAN_OBJ},
		{`matching("#", by=reCompile("(?x:" ~ reEsc("#") ~ ")"))`, true, object.BOOLEAN_OBJ},
		{`matching("", by=reCompile("(?x:" ~ reEsc("#") ~ ")"))`, false, object.BOOLEAN_OBJ},
		{`matching("#", by=reCompile(reEsc("#")))`, true, object.BOOLEAN_OBJ},
		{`matching("", by=reCompile(reEsc("#")))`, false, object.BOOLEAN_OBJ},
	}

	runVmTests(t, tests, false, false)
}

func TestSubMatchesHashList(t *testing.T) {
	str := `submatchesH(" abcd: ", by=RE/(?P<key>\w+)\s*:\s*(?P<value>\w+)/)`
	expect := [][][]object.Object{}
	result := oneResult(t, 1, str, false, false)
	err := testListOfHashesObject(expect, result)
	if err != nil {
		t.Errorf("testListOfHashesObject failed: %s", err)
	}

	str = `submatchesH(" abcd: peaceInJerusalem ", by=RE/(?P<key>\w+)\s*:\s*(?P<value>\w+)/)`
	expect = [][][]object.Object{{
		{object.NumberFromInt(0), object.NewString("abcd: peaceInJerusalem")},
		{object.NumberFromInt(1), object.NewString("abcd")},
		{object.NumberFromInt(2), object.NewString("peaceInJerusalem")},
		{object.NewString("key"), object.NewString("abcd")},
		{object.NewString("value"), object.NewString("peaceInJerusalem")},
	}}
	result = oneResult(t, -1, str, false, false)
	err = testListOfHashesObject(expect, result)
	if err != nil {
		t.Errorf("testListOfHashesObject failed: %s", err)
	}

	str = `submatchesH(" first: youknow ; second : youdontknow ", by=RE/(?P<key>\w+)\s*:\s*(?P<value>\w+)/)`
	expect = [][][]object.Object{
		{
			{object.NumberFromInt(0), object.NewString("first: youknow")},
			{object.NumberFromInt(1), object.NewString("first")},
			{object.NumberFromInt(2), object.NewString("youknow")},
			{object.NewString("key"), object.NewString("first")},
			{object.NewString("value"), object.NewString("youknow")},
		},
		{
			{object.NumberFromInt(0), object.NewString("second : youdontknow")},
			{object.NumberFromInt(1), object.NewString("second")},
			{object.NumberFromInt(2), object.NewString("youdontknow")},
			{object.NewString("key"), object.NewString("second")},
			{object.NewString("value"), object.NewString("youdontknow")},
		},
	}
	result = oneResult(t, 1, str, false, false)
	err = testListOfHashesObject(expect, result)
	if err != nil {
		t.Errorf("testListOfHashesObject failed: %s", err)
	}
}

func TestRe2Modifiers(t *testing.T) {
	tests := []vmTestCase{
		// https://github.com/google/re2/wiki/Syntax (see "flags")
		// http://rexegg.com/regex-modifiers.html

		// case insensitive
		{`matching("ABC", by=re/a.*c/)`, false, object.BOOLEAN_OBJ},
		{`matching("ABC", by=re:i/a.*c/)`, true, object.BOOLEAN_OBJ},
		{`matching("ABC", by=re:i/a.*z/)`, false, object.BOOLEAN_OBJ},

		// case insensitive with interpolation
		{`matching("AB2C", by=re/a.*{{1+1}}c/)`, false, object.BOOLEAN_OBJ},
		{`matching("AB2C", by=re:i/a.*{{1+1}}c/)`, true, object.BOOLEAN_OBJ},
		{`matching("AB2C", by=re:i/a.*{{1+1}}z/)`, false, object.BOOLEAN_OBJ},

		// single line (a.k.a. DOTALL mode)
		{`matching("a\nc", by=re/a.c/)`, false, object.BOOLEAN_OBJ},
		{`matching("a\nc", by=re:s/a.c/)`, true, object.BOOLEAN_OBJ},

		// multiline mode, which is NOT the opposite of single line mode
		{`matching("\nabc", by=re/^a.c/)`, false, object.BOOLEAN_OBJ},
		{`matching("\nabc", by=re:m/^a.c/)`, true, object.BOOLEAN_OBJ},

		{`matching("abc\n", by=re(a.c$))`, false, object.BOOLEAN_OBJ},
		{`matching("abc\n", by=re:m(a.c$))`, true, object.BOOLEAN_OBJ},

		// ungreedy mode (reverse "greediness"/"laziness" of quantifiers)
		{`match("1234567", by=RE/\d+/)`, "1234567", object.STRING_OBJ},
		{`match("1234567", by=RE:U/\d+/)`, "1", object.STRING_OBJ},

		{`match("1234567", by=RE/\d+?/)`, "1", object.STRING_OBJ},
		{`match("1234567", by=RE:U/\d+?/)`, "1234567", object.STRING_OBJ},

		// combined
		{`matching("a\nC", by=re:s:i/a.c/)`, true, object.BOOLEAN_OBJ},
		{`matching("a\nC", by=re:s/a.c/)`, false, object.BOOLEAN_OBJ},
		{`match("a\nC\n", by=re:s:m:U:i/a.c.*$/)`, "a\nC", object.STRING_OBJ},
	}

	runVmTests(t, tests, false, false)
}

func TestRe2BlockQuotesAndFreeSpacingMode(t *testing.T) {
	tests := []vmTestCase{
		{`val r = RE:block END
\d+\.\d+
END
matching(123.0, by=r)`, true, object.BOOLEAN_OBJ},

		// using lead modifier, which trims leading space on each line
		{`val r = RE:lead:block END
			\d+\.\d+
			abc
			END
			matching("123.0\nabc", by=r)`, true, object.BOOLEAN_OBJ},

		{`val r = re:block END
\\d+\\.\\d+
END
		matching(123.0, by=r)`, true, object.BOOLEAN_OBJ},

		// free-spacing mode
		{`val r = re:x/ abc /
		  matching("abc", by=r)`, true, object.BOOLEAN_OBJ},

		{`val r = re:x:block END
			abc
END
		  matching("abc", by=r)`, true, object.BOOLEAN_OBJ},

		{`val r = re:x:block END
			abc
			# an intwesting comment
END
		  matching("abc", by=r)`, true, object.BOOLEAN_OBJ},

		{`val r = re:x:block END
			abc
			# an intwesting comment
			123[ ]
END
		  matching("abc123 ", by=r)`, true, object.BOOLEAN_OBJ},

		{`val r = RE:x:block END
			abc\#
			# an intwesting comment
			(?-x: you know )  	# with a non-free-spacing section in the middle
			123\ 				# can't stop commenting
		  END
		  matching("abc# you know 123 ", by=r)`, true, object.BOOLEAN_OBJ},

		// free-spacing mode with interpolation
		// esc modifier on re literal will escape all interpolations
		{`val x = "yo yo"
		  val r = re:x:esc:block END
			abc {{x}}
			# an intwesting comment
			123
		  END
		  matching("abcyo yo123", by=r)`, true, object.BOOLEAN_OBJ},

		{`val x = "hey hey"
		  val r = re:x:block END
			abc {{x}}
			# an intwesting comment
			123
     END
		  matching("abcheyhey123", by=r)`, true, object.BOOLEAN_OBJ},
	}

	runVmTests(t, tests, false, false)
}

func TestInterpolationIntoNonFreeSpacingRegex(t *testing.T) {
	tests := []vmTestCase{
		{ // non-free-spacing regex interpolated into a non-free-spacing regex literal
			`val re1 = re/yo joe/
		     val re2 = re/{{re1}}/
		     matching("yo joes", by=re2)`,
			true, object.BOOLEAN_OBJ,
		},
		{ // free-spacing regex interpolated into a non-free-spacing regex literal
			`val re1 = re:x/yo joe/
		     val re2 = re/{{re1}}/
		     matching("yojoes", by=re2)`,
			true, object.BOOLEAN_OBJ,
		},
		{ // string interpolated into a non-free-spacing regex literal without escaping
			`val re1 = "yo joe"
		     val re2 = re/{{re1}}/
		     matching("yo joes", by=re2)`,
			true, object.BOOLEAN_OBJ,
		},
		{ // string interpolated into a non-free-spacing regex literal with escaping
			`val re1 = "yo joe"
		     val re2 = re:esc/{{re1}}/
		     matching("yo joes", by=re2)`,
			true, object.BOOLEAN_OBJ,
		},
		{ // string interpolated into a non-free-spacing regex literal with escaping
			`val re1 = "yo joe"
		     val re2 = re/{{re1:esc}}/
		     matching("yo joes", by=re2)`,
			true, object.BOOLEAN_OBJ,
		},
	}

	runVmTests(t, tests, false, false)
}

func TestInterpolationIntoFreeSpacingRegex(t *testing.T) {
	tests := []vmTestCase{
		{ // non-free-spacing regex interpolated into a free-spacing regex literal
			`val re1 = re/yo joe/
		     val re2 = re:x/{{re1}}/
		     matching("yo joes", by=re2)`,
			true, object.BOOLEAN_OBJ,
		},
		{ // free-spacing regex interpolated into a free-spacing regex literal
			`val re1 = re:x/yo joe/
		     val re2 = re:x/{{re1}}/
		     matching("yojoes", by=re2)`,
			true, object.BOOLEAN_OBJ,
		},
		{ // string interpolated into a free-spacing regex literal without escaping
			`val re1 = "yo joe"
		     val re2 = re:x/{{re1}}/
		     matching("yojoes", by=re2)`,
			true, object.BOOLEAN_OBJ,
		},
		{ // string interpolated into a free-spacing regex literal with escaping
			`val re1 = "yo joe"
		     val re2 = re:esc:x/{{re1}}/
		     matching("yo joes", by=re2)`,
			true, object.BOOLEAN_OBJ,
		},
		{ // string interpolated into a free-spacing regex literal with escaping
			`val re1 = "yo joe"
		     val re2 = re:x/{{re1:esc}}/
		     matching("yo joes", by=re2)`,
			true, object.BOOLEAN_OBJ,
		},
	}

	runVmTests(t, tests, false, false)
}

func TestRegexFunctionsWithPlainStrings(t *testing.T) {
	// some of the same functions with plain strings instead of regexes
	tests := []vmTestCase{
		{`replace(" abc abc ", by="abc", with="7", max=1)`, " 7 abc ", object.STRING_OBJ},
		{`replace(" abc abc ", by="abc", with="7")`, " 7 7 ", object.STRING_OBJ},

		// passing nothing to replace (no replacement string or function; default ZLS)
		{`replace("abc azc aec afc ", by=" a")`, "abczcecfc ", object.STRING_OBJ},

		// replace with function
		{`replace(" abc abc ", by="b", with=fn(s) { ucase s })`, " aBc aBc ", object.STRING_OBJ},
		{`replace(" abc abc ", by="abc", with=fn(s) { "ZZZ" })`, " ZZZ ZZZ ", object.STRING_OBJ},

		{`split("don't.ya.know", delim=".")`, []string{"don't", "ya", "know"}, object.LIST_OBJ},
		{`split("abc", delim="b")`, []string{"a", "c"}, object.LIST_OBJ},
		{`split("ασδ", delim="σ")`, []string{"α", "δ"}, object.LIST_OBJ},

		// split with default ZLS delimiter
		{`split("ασδ")`, []string{"α", "σ", "δ"}, object.LIST_OBJ},
		{`split(3.14)`, []string{"3", ".", "1", "4"}, object.LIST_OBJ},

		{`matching("basdfabcklsdf", by="abc")`, true, object.BOOLEAN_OBJ},
		{`matching("basdfabcklsdf", by="abC")`, false, object.BOOLEAN_OBJ},

		{`index("basdfabcklsdf", by="abc")`, []int64{6, 8}, object.RANGE_OBJ},
		{`index("basdfabcklsdf", by="abC")`, nil, object.NULL_OBJ},
		{`index("basdfaZcklsdf", by="Z")`, []int64{7, 7}, object.RANGE_OBJ},

		{`indices(7467337300, by="73")`, [][]int64{{4, 5}, {7, 8}}, object.LIST_OBJ},
		{`indices(7467337300, by="7")`, [][]int64{{1, 1}, {4, 4}, {7, 7}}, object.LIST_OBJ},
		{`indices(7467337300, by="7", max=4)`, [][]int64{{1, 1}, {4, 4}, {7, 7}}, object.LIST_OBJ},
		{`indices(7467337300, by="7", max=3)`, [][]int64{{1, 1}, {4, 4}, {7, 7}}, object.LIST_OBJ},
		{`indices(7467337300, by="7", max=2)`, [][]int64{{1, 1}, {4, 4}}, object.LIST_OBJ},
		{`indices(7467337300, by="7", max=1)`, [][]int64{{1, 1}}, object.LIST_OBJ},
		{`indices(7467337300, by="a")`, [][]int64{}, object.LIST_OBJ},
	}

	runVmTests(t, tests, false, false)
}

func TestSplitByNumber(t *testing.T) {
	tests := []vmTestCase{
		{`split("ασδ", delim=1)`, []string{"α", "σ", "δ"}, object.LIST_OBJ},
		{`split("ασδ", delim=1, max=2)`, []string{"α", "σδ"}, object.LIST_OBJ},
		{`split("ασδ", delim=-1, max=2)`, []string{"ασ", "δ"}, object.LIST_OBJ},
		{`split("ασδ", delim=2)`, []string{"ασ", "δ"}, object.LIST_OBJ},
		{`split("ασδ", delim=-2)`, []string{"α", "σδ"}, object.LIST_OBJ},

		{`split("123456789", delim=1)`, []string{"1", "2", "3", "4", "5", "6", "7", "8", "9"}, object.LIST_OBJ},
		{`split("123456789", delim=2)`, []string{"12", "34", "56", "78", "9"}, object.LIST_OBJ},
		{`split("1234567890", delim=2)`, []string{"12", "34", "56", "78", "90"}, object.LIST_OBJ},
		{`split("123456789", delim=3)`, []string{"123", "456", "789"}, object.LIST_OBJ},
		{`split("1234567890", delim=3)`, []string{"123", "456", "789", "0"}, object.LIST_OBJ},
		{`split("1234567890", delim=9)`, []string{"123456789", "0"}, object.LIST_OBJ},
		{`split("1234567890", delim=10)`, []string{"1234567890"}, object.LIST_OBJ},
		{`split("1234567890", delim=12)`, []string{"1234567890"}, object.LIST_OBJ},

		{`split("123456789", delim=3, max=2)`, []string{"123", "456789"}, object.LIST_OBJ},
		{`split("1234567890", delim=3, max=2)`, []string{"123", "4567890"}, object.LIST_OBJ},
		{`split("1234567890", delim=3, max=3)`, []string{"123", "456", "7890"}, object.LIST_OBJ},
		{`split("1234567890", delim=3, max=4)`, []string{"123", "456", "789", "0"}, object.LIST_OBJ},
		{`split("1234567890", delim=3, max=5)`, []string{"123", "456", "789", "0"}, object.LIST_OBJ},

		{`split("1234567890123456789012345678901234567890", delim=3, max=2)`,
			[]string{"123", "4567890123456789012345678901234567890"}, object.LIST_OBJ},

		{`split("1234567890123456789012345678901234567890", delim=3, max=3)`,
			[]string{"123", "456", "7890123456789012345678901234567890"}, object.LIST_OBJ},

		{`split("1234567890123456789012345678901234567890", delim=3, max=4)`,
			[]string{"123", "456", "789", "0123456789012345678901234567890"}, object.LIST_OBJ},

		{`split("1234567890123456789012345678901234567890", delim=3, max=12)`,
			[]string{"123", "456", "789", "012", "345", "678", "901", "234", "567", "890", "123", "4567890"}, object.LIST_OBJ},

		{`split("1234567890123456789012345678901234567890", delim=3, max=13)`,
			[]string{"123", "456", "789", "012", "345", "678", "901", "234", "567", "890", "123", "456", "7890"}, object.LIST_OBJ},

		{`split("1234567890123456789012345678901234567890", delim=3, max=14)`,
			[]string{"123", "456", "789", "012", "345", "678", "901", "234", "567", "890", "123", "456", "789", "0"}, object.LIST_OBJ},

		{`split("1234567890123456789012345678901234567890", delim=3)`,
			[]string{"123", "456", "789", "012", "345", "678", "901", "234", "567", "890", "123", "456", "789", "0"}, object.LIST_OBJ},

		{`split("123456789", delim=-3, max=2)`, []string{"123456", "789"}, object.LIST_OBJ},
		{`split("1234567890", delim=-3, max=2)`, []string{"1234567", "890"}, object.LIST_OBJ},
		{`split("1234567890", delim=-3, max=3)`, []string{"1234", "567", "890"}, object.LIST_OBJ},
		{`split("1234567890", delim=-3, max=4)`, []string{"1", "234", "567", "890"}, object.LIST_OBJ},
		{`split("1234567890", delim=-3, max=5)`, []string{"1", "234", "567", "890"}, object.LIST_OBJ},

		{`split("1234567890123456789012345678901234567890", delim=-3, max=2)`,
			[]string{"1234567890123456789012345678901234567", "890"}, object.LIST_OBJ},

		{`split("1234567890123456789012345678901234567890", delim=-3, max=3)`,
			[]string{"1234567890123456789012345678901234", "567", "890"}, object.LIST_OBJ},

		{`split("1234567890123456789012345678901234567890", delim=-3, max=4)`,
			[]string{"1234567890123456789012345678901", "234", "567", "890"}, object.LIST_OBJ},

		{`split("1234567890123456789012345678901234567890", delim=-3, max=12)`,
			[]string{"1234567", "890", "123", "456", "789", "012", "345", "678", "901", "234", "567", "890"}, object.LIST_OBJ},

		{`split("1234567890123456789012345678901234567890", delim=-3, max=13)`,
			[]string{"1234", "567", "890", "123", "456", "789", "012", "345", "678", "901", "234", "567", "890"}, object.LIST_OBJ},

		{`split("1234567890123456789012345678901234567890", delim=-3, max=14)`,
			[]string{"1", "234", "567", "890", "123", "456", "789", "012", "345", "678", "901", "234", "567", "890"}, object.LIST_OBJ},

		{`split("1234567890123456789012345678901234567890", delim=-3)`,
			[]string{"1", "234", "567", "890", "123", "456", "789", "012", "345", "678", "901", "234", "567", "890"}, object.LIST_OBJ},

		{`split("1234567890", delim=-12)`, []string{"1234567890"}, object.LIST_OBJ},
		{`split("1234567890", delim=-10)`, []string{"1234567890"}, object.LIST_OBJ},
		{`split("1234567890", delim=-9)`, []string{"1", "234567890"}, object.LIST_OBJ},
		{`split("123456789", delim=-3)`, []string{"123", "456", "789"}, object.LIST_OBJ},
		{`split("1234567890", delim=-3)`, []string{"1", "234", "567", "890"}, object.LIST_OBJ},
		{`split("123456789", delim=-2)`, []string{"1", "23", "45", "67", "89"}, object.LIST_OBJ},
		{`split("1234567890", delim=-2)`, []string{"12", "34", "56", "78", "90"}, object.LIST_OBJ},
		{`split("123456789", delim=-1)`, []string{"1", "2", "3", "4", "5", "6", "7", "8", "9"}, object.LIST_OBJ},

		// Do something practical with it.
		{`join(split("1234567890", delim=-3), delim=",")`, "1,234,567,890", object.STRING_OBJ},

		{`"2x" ~ join(map(split("{{2 ^ 63 - 1 : 2x}}", delim=-8), by=fn(x) { "{{x:8(0)}}" }), delim="_")`,
			"2x01111111_11111111_11111111_11111111_11111111_11111111_11111111_11111111", object.STRING_OBJ},
	}

	runVmTests(t, tests, false, false)
}

func TestDateTimeLiterals(t *testing.T) {
	tests := []vmTestCase{
		{`dt/2020-03-24T17:51:25-05:00/`, "2020-03-24T17:51:25-05:00", object.DATETIME_OBJ},
		{`dt/2020-03-24 17:51:25-09:15/`, "2020-03-24T17:51:25-09:15", object.DATETIME_OBJ},
		{`dt/2020-03-24 17:51:25+09:15/`, "2020-03-24T17:51:25+09:15", object.DATETIME_OBJ},
		{`dt/2020-03-24 17:51:25+07:30/`, "2020-03-24T17:51:25+07:30", object.DATETIME_OBJ},
		{`dt/2020-03-24 17:51:25-07:30/`, "2020-03-24T17:51:25-07:30", object.DATETIME_OBJ},
		{`dt/2020-03-24T17:51:25Z/`, "2020-03-24T17:51:25Z", object.DATETIME_OBJ},

		{`dt/2020-03-24T17:51-05:00/`, "2020-03-24T17:51:00-05:00", object.DATETIME_OBJ},
		{`dt/2020-03-24 17-05:00/`, "2020-03-24T17:00:00-05:00", object.DATETIME_OBJ},
		{`dt/2020-03-24 17Z/`, "2020-03-24T17:00:00Z", object.DATETIME_OBJ},

		// with fractions on seconds
		{`string dt/2020-03-24 17:51:25.123+07:30/, fmt="2006-01-02T15:04:05.999999999Z07:00"`,
			"2020-03-24T17:51:25.123+07:30", object.STRING_OBJ},
		{`string dt/2020-03-24 17:51:25.123456789+07:30/, fmt="2006-01-02T15:04:05.999999999Z07:00"`,
			"2020-03-24T17:51:25.123456789+07:30", object.STRING_OBJ},
		{`string dt/2020-03-24 17:51:25.0001+07:30/, fmt="2006-01-02T15:04:05.999999999Z07:00"`,
			"2020-03-24T17:51:25.0001+07:30", object.STRING_OBJ},
	}

	runVmTests(t, tests, false, false)
}

func TestDurations(t *testing.T) {
	tests := []vmTestCase{
		{`number(dr/1D/)`, "86400000000000", object.NUMBER_OBJ},
		{`number(dr/1DT1H/)`, "90000000000000", object.NUMBER_OBJ},
		{`number(dr/T23M10S/)`, "1390000000000", object.NUMBER_OBJ},

		// now allowing easier to read formatting
		{`number(dr/T 23m 10s/)`, "1390000000000", object.NUMBER_OBJ},

		// with fractional seconds
		{`number(dr/T23M10.1S/)`, "1390100000000", object.NUMBER_OBJ},
		{`number(dr/T23M10.001S/)`, "1390001000000", object.NUMBER_OBJ},
		{`number(dr/T23M10.123456789S/)`, "1390123456789", object.NUMBER_OBJ},
		{`number(dr/T23M10.000000009S/)`, "1390000000009", object.NUMBER_OBJ},
	}

	runVmTests(t, tests, false, false)
}

func TestDateTimeMath(t *testing.T) {
	tests := []vmTestCase{
		// date-time + duration
		{`dt/2020-03-24 17Z/ + dr/1D/`, "2020-03-25T17:00:00Z", object.DATETIME_OBJ},
		{`dt/2020-03-31 17Z/ + dr/2D/`, "2020-04-02T17:00:00Z", object.DATETIME_OBJ},
		{`dt/2020-03-31 17Z/ - dr/2D/`, "2020-03-29T17:00:00Z", object.DATETIME_OBJ},
		{`dt/2020-03-31 17Z/ - dr/2DT12H30M/`, "2020-03-29T04:30:00Z", object.DATETIME_OBJ},

		// whole years, etc.
		{`dt/2020-03-31 17Z/ + dr/1Y/`, "2021-03-31T17:00:00Z", object.DATETIME_OBJ},
		{`dt/2020-03-31 17Z/ - dr/1Y/`, "2019-03-31T17:00:00Z", object.DATETIME_OBJ},

		{`dt/2020-03-15 17Z/ + dr/1M/`, "2020-04-15T17:00:00Z", object.DATETIME_OBJ},
		{`dt/2020-03-15 17Z/ - dr/1M/`, "2020-02-15T17:00:00Z", object.DATETIME_OBJ},

		{`dt/2020-03-15 17Z/ + dr/1D/`, "2020-03-16T17:00:00Z", object.DATETIME_OBJ},
		{`dt/2020-03-15 17Z/ - dr/1D/`, "2020-03-14T17:00:00Z", object.DATETIME_OBJ},

		{`dt/2020-03-15 17Z/ + dr/T1H/`, "2020-03-15T18:00:00Z", object.DATETIME_OBJ},
		{`dt/2020-03-15 17Z/ - dr/T1H/`, "2020-03-15T16:00:00Z", object.DATETIME_OBJ},

		{`dt/2020-03-15 17Z/ + dr/T1M/`, "2020-03-15T17:01:00Z", object.DATETIME_OBJ},
		{`dt/2020-03-15 17Z/ - dr/T1M/`, "2020-03-15T16:59:00Z", object.DATETIME_OBJ},

		{`dt/2020-03-15 17Z/ + dr/T1S/`, "2020-03-15T17:00:01Z", object.DATETIME_OBJ},
		{`dt/2020-03-15 17Z/ - dr/T1S/`, "2020-03-15T16:59:59Z", object.DATETIME_OBJ},

		{`dt/2020-03-15 17Z/ + dr/T0.1S/`, "2020-03-15T17:00:00.1Z", object.DATETIME_OBJ},
		{`dt/2020-03-15 17Z/ - dr/T0.1S/`, "2020-03-15T16:59:59.9Z", object.DATETIME_OBJ},

		{`dt/2020-03-15 17Z/ + dr/T0.000000001S/`, "2020-03-15T17:00:00.000000001Z", object.DATETIME_OBJ},
		{`dt/2020-03-15 17Z/ - dr/T0.000000001S/`, "2020-03-15T16:59:59.999999999Z", object.DATETIME_OBJ},

		// date-time + nanoseconds
		{`dt/2020-03-15 17Z/ + 1`, "2020-03-15T17:00:00.000000001Z", object.DATETIME_OBJ},
		{`dt/2020-03-15 17Z/ - 1`, "2020-03-15T16:59:59.999999999Z", object.DATETIME_OBJ},
		{`dt/2020-03-15 17Z/ + 604800000000000 == dt/2020-03-22 17Z/`, true, object.BOOLEAN_OBJ},

		// date-time - date-time, producing a duration
		{`dt/2015-05-01/ - dt/2016-06-02T01:01:01.000000001/ == dr/1Y 1M 1D T 1H 1M 1.000000001S/`, true, object.BOOLEAN_OBJ},
		{`dt/2016-06-02T01:01:01.000000001/ - dt/2015-05-01/ == dr/1Y 1M 1D T 1H 1M 1.000000001S/`, true, object.BOOLEAN_OBJ},
		{`dt/2015-01-02/ - dt/2015-02-01/ == dr/30D/`, true, object.BOOLEAN_OBJ},
		{`dt/2016-01-02/ - dt/2016-02-01/ == dr/30D/`, true, object.BOOLEAN_OBJ},
		{`dt/2015-02-02/ - dt/2015-03-01/ == dr/27D/`, true, object.BOOLEAN_OBJ},
		{`dt/2016-02-02/ - dt/2016-03-01/ == dr/28D/`, true, object.BOOLEAN_OBJ},
		{`dt/2015-02-11/ - dt/2016-01-12/ == dr/11M 1D/`, true, object.BOOLEAN_OBJ},
		{`dt/2015-01-11/ - dt/2015-03-10/ == dr/1M 30D/`, true, object.BOOLEAN_OBJ},
		{`dt/2015-12-31/ - dt/2015-03-01/ == dr/9M 30D/`, true, object.BOOLEAN_OBJ},
		{`dt/2015-12-31/ - dt/2016-03-01/ == dr/2M 1D/`, true, object.BOOLEAN_OBJ},
		{`dt/2015-12-31/ - dt/2016-02-28/ == dr/1M 28D/`, true, object.BOOLEAN_OBJ},
	}

	runVmTests(t, tests, false, false)
}

func TestDateTimeOutputFormatting(t *testing.T) {
	tests := []vmTestCase{
		// output format strings...
		// Apparently, the Go time package expects exactly the following numbers and formats in a format string.
		// Anything else will not yield the results you expect.
		// year: 2006 or 06
		// month: 01 or 1
		// month name: Jan or January
		// month day: 02 or 2
		// weekday name: Mon or Monday
		// hour: 03 or 3 or 15
		// minute: 04 or 4
		// second: 05 or 5
		// AM/PM: PM or pm
		// time zone offset: -07:00 or -0700 or -07
		// time zone name: MST

		{`val x = dt/2020-03-24 17:51:25-05:00/
		  string(x)`, "2020-03-24T17:51:25-05:00", object.STRING_OBJ},
		{`val x = dt'2008-01-01 12:30:12'
		  string(x, fmt="2006")`, "2008", object.STRING_OBJ},
		{`val x = dt"2008-01-01 12:30:12"
		  string(x, fmt="2006")`, "2008", object.STRING_OBJ},
		{`val x = dt/2008-01-01 12:30/
		  string(x, fmt="2006")`, "2008", object.STRING_OBJ},
		{`val x = dt(2008-01-01 12)
		  string(x, fmt="2006")`, "2008", object.STRING_OBJ},

		{`val x = dt/2008-01-01 11/
		  string(x, fmt="pm")`, "am", object.STRING_OBJ},
		{`val x = dt/2008-01-01 13/
		  string(x, fmt="pm")`, "pm", object.STRING_OBJ},

		{`val x = dt/2008-07-22 11/
		  string(x, fmt="Jan pm 01 05 3 January")`, "Jul am 07 00 11 July", object.STRING_OBJ},
	}

	runVmTests(t, tests, false, false)
}

func TestDateTimeComparisons(t *testing.T) {
	tests := []vmTestCase{
		{`dt/2008-01-01 11/ == dt/2008-01-01 11/`,
			true, object.BOOLEAN_OBJ},
		{`dt/2008-01-01 11/ != dt/2008-01-01 11/`,
			false, object.BOOLEAN_OBJ},
		{`dt/2008-01-01 11/ == dt/2010-01-01 11/`,
			false, object.BOOLEAN_OBJ},
		{`dt/2008-01-01 11/ != dt/2010-01-01 11/`,
			true, object.BOOLEAN_OBJ},

		{`dt/2008-01-01 11/ > dt/2008-01-01 11/`,
			false, object.BOOLEAN_OBJ},
		{`dt/2008-01-01 11/ < dt/2008-01-01 11/`,
			false, object.BOOLEAN_OBJ},
		{`dt/2008-01-01 11/ >= dt/2008-01-01 11/`,
			true, object.BOOLEAN_OBJ},
		{`dt/2008-01-01 11/ <= dt/2008-01-01 11/`,
			true, object.BOOLEAN_OBJ},

		{`dt/2008-01-01 11/ > dt/2010-01-01 11/`,
			false, object.BOOLEAN_OBJ},
		{`dt/2008-01-01 11/ < dt/2010-01-01 11/`,
			true, object.BOOLEAN_OBJ},
		{`dt/2008-01-01 11/ >= dt/2010-01-01 11/`,
			false, object.BOOLEAN_OBJ},
		{`dt/2008-01-01 11/ <= dt/2010-01-01 11/`,
			true, object.BOOLEAN_OBJ},

		{`dt/2018-01-01 11/ > dt/2008-01-01 11/`,
			true, object.BOOLEAN_OBJ},
		{`dt/2018-01-01 11/ < dt/2008-01-01 11/`,
			false, object.BOOLEAN_OBJ},
		{`dt/2018-01-01 11/ >= dt/2008-01-01 11/`,
			true, object.BOOLEAN_OBJ},
		{`dt/2018-01-01 11/ <= dt/2008-01-01 11/`,
			false, object.BOOLEAN_OBJ},
	}

	runVmTests(t, tests, false, false)
}

func TestDateTimeFunction(t *testing.T) {
	tests := []vmTestCase{
		{`dt/2020-03-13/ == datetime("2020-03-13")`,
			// no format string: defaults to local time zone, just as date-time literal does
			true, object.BOOLEAN_OBJ,
		},
		{`dt/2020-03-13 00:00+00/ == datetime("Mar 13, 2020", fmt="Jan 02, 2006")`,
			// with format string: defaults to UTC time zone (by Go time library)
			true, object.BOOLEAN_OBJ,
		},

		// test by round-trip
		{`var x = dt//
		  var y = hash(x)
		  datetime(y) == x`,
			true, object.BOOLEAN_OBJ,
		},
		{`var x = dt//
		  var y = number(x)
		  datetime(y) == x`,
			true, object.BOOLEAN_OBJ,
		},
		{`var x = dt/2020-03-13/
		  # Mar. 13, 2020, the day I heard the US was shutting down
		  # accounting for nanoseconds by using a set date with 0 nanoseconds
		  var y = string(x)
		  datetime(y) == x`,
			true, object.BOOLEAN_OBJ,
		},

		// ISO shortened form from a string
		// short form not allowed for literals
		{`dt/2020-03-13/ == datetime("20200313")`,
			true, object.BOOLEAN_OBJ,
		},
		{`dt/2020-03-13T12:03:13/ == datetime("20200313T120313")`,
			true, object.BOOLEAN_OBJ,
		},
	}

	runVmTests(t, tests, false, false)
}

func TestDurationBasicComparisons(t *testing.T) {
	// naive, testing first by years, then months, etc.
	tests := []vmTestCase{
		{`dr/10Y/ == dr/10Y/`, true, object.BOOLEAN_OBJ},
		{`dr/10Y/ > dr/9Y/`, true, object.BOOLEAN_OBJ},
		{`dr/10Y/ >= dr/9Y/`, true, object.BOOLEAN_OBJ},
		{`dr/10Y/ == dr/9Y/`, false, object.BOOLEAN_OBJ},
		{`dr/8Y/ > dr/9Y/`, false, object.BOOLEAN_OBJ},
		{`dr/8Y/ >= dr/9Y/`, false, object.BOOLEAN_OBJ},

		{`dr/10M/ == dr/10M/`, true, object.BOOLEAN_OBJ},
		{`dr/10M/ > dr/9M/`, true, object.BOOLEAN_OBJ},
		{`dr/10M/ >= dr/9M/`, true, object.BOOLEAN_OBJ},
		{`dr/10M/ == dr/9M/`, false, object.BOOLEAN_OBJ},
		{`dr/8M/ > dr/9M/`, false, object.BOOLEAN_OBJ},
		{`dr/8M/ >= dr/9M/`, false, object.BOOLEAN_OBJ},

		{`dr/10D/ == dr/10D/`, true, object.BOOLEAN_OBJ},
		{`dr/10D/ > dr/9D/`, true, object.BOOLEAN_OBJ},
		{`dr/10D/ >= dr/9D/`, true, object.BOOLEAN_OBJ},
		{`dr/10D/ == dr/9D/`, false, object.BOOLEAN_OBJ},
		{`dr/8D/ > dr/9D/`, false, object.BOOLEAN_OBJ},
		{`dr/8D/ >= dr/9D/`, false, object.BOOLEAN_OBJ},

		{`dr/T10H/ == dr/T10H/`, true, object.BOOLEAN_OBJ},
		{`dr/T10H/ > dr/T9H/`, true, object.BOOLEAN_OBJ},
		{`dr/T10H/ >= dr/T9H/`, true, object.BOOLEAN_OBJ},
		{`dr/T10H/ == dr/T9H/`, false, object.BOOLEAN_OBJ},
		{`dr/T8H/ > dr/T9H/`, false, object.BOOLEAN_OBJ},
		{`dr/T8H/ >= dr/T9H/`, false, object.BOOLEAN_OBJ},

		{`dr/T10M/ == dr/T10M/`, true, object.BOOLEAN_OBJ},
		{`dr/T10M/ > dr/T9M/`, true, object.BOOLEAN_OBJ},
		{`dr/T10M/ >= dr/T9M/`, true, object.BOOLEAN_OBJ},
		{`dr/T10M/ == dr/T9M/`, false, object.BOOLEAN_OBJ},
		{`dr/T8M/ > dr/T9M/`, false, object.BOOLEAN_OBJ},
		{`dr/T8M/ >= dr/T9M/`, false, object.BOOLEAN_OBJ},

		{`dr/T10S/ == dr/T10S/`, true, object.BOOLEAN_OBJ},
		{`dr/T10S/ > dr/T9S/`, true, object.BOOLEAN_OBJ},
		{`dr/T10S/ >= dr/T9S/`, true, object.BOOLEAN_OBJ},
		{`dr/T10S/ == dr/T9S/`, false, object.BOOLEAN_OBJ},
		{`dr/T8S/ > dr/T9S/`, false, object.BOOLEAN_OBJ},
		{`dr/T8S/ >= dr/T9S/`, false, object.BOOLEAN_OBJ},

		{`dr/T0.9S/ == dr/T0.9S/`, true, object.BOOLEAN_OBJ},
		{`dr/T0.9S/ > dr/T0.8S/`, true, object.BOOLEAN_OBJ},
		{`dr/T0.9S/ >= dr/T0.8S/`, true, object.BOOLEAN_OBJ},
		{`dr/T0.8S/ == dr/T0.9S/`, false, object.BOOLEAN_OBJ},
		{`dr/T0.8S/ > dr/T0.9S/`, false, object.BOOLEAN_OBJ},
		{`dr/T0.8S/ >= dr/T0.9S/`, false, object.BOOLEAN_OBJ},

		{`dr/T0.9S/ <= dr/T0.9S/`, true, object.BOOLEAN_OBJ},
		{`dr/T0.8S/ <= dr/T0.9S/`, true, object.BOOLEAN_OBJ},

		// "P" optional
		{`dr/1Y/ == dr/P1Y/`, true, object.BOOLEAN_OBJ},
		{`dr/7W/ == dr/P7W/`, true, object.BOOLEAN_OBJ},
	}

	runVmTests(t, tests, false, false)
}

func TestDurationNaiveComparisons(t *testing.T) {
	// naive, testing first by years, then months, etc.
	tests := []vmTestCase{
		{`dr/1Y/ == dr/12M/`, false, object.BOOLEAN_OBJ},
		{`dr/1Y/ > dr/1M/`, true, object.BOOLEAN_OBJ},

		{`dr/2Y/ > dr/1Y 5M/`, true, object.BOOLEAN_OBJ},
		{`dr/1Y/ > dr/1Y 5M/`, false, object.BOOLEAN_OBJ},
		{`dr/1Y/ == dr/1Y 5M/`, false, object.BOOLEAN_OBJ},

		{`dr/1Y/ > dr/5000M/`, true, object.BOOLEAN_OBJ},
		{`dr/2Y/ >= dr/1Y 5M/`, true, object.BOOLEAN_OBJ},
		{`dr/1Y/ <= dr/1Y 5M/`, true, object.BOOLEAN_OBJ},
		{`dr/1Y/ >= dr/1Y 5M/`, false, object.BOOLEAN_OBJ},
	}

	runVmTests(t, tests, false, false)
}

func TestDurationMath(t *testing.T) {
	tests := []vmTestCase{
		{`dr/2Y/ + dr/1Y5M/ == dr/3Y5M/`, true, object.BOOLEAN_OBJ},
		{`dr/2Y/ + dr/1Y5M/ == dr/3Y6M/`, false, object.BOOLEAN_OBJ},

		{`dr/2Y T 21M/ + dr/T 13H 3.2S/ == dr/2Y T 13H 21M 3.2S/`, true, object.BOOLEAN_OBJ},
		{`dr/2Y T 21M/ + dr/T 13H 3.2S/ == dr/2Y T 13H 21M 3.1S/`, false, object.BOOLEAN_OBJ},
		{`dr/2Y T 21M/ + dr/T 13H 3.2S/ == dr/T 13H 21M 3.2S/`, false, object.BOOLEAN_OBJ},
		{`dr/2Y T 21M/ + dr/T 13H 3.2S/ == dr/2Y T 13H 21M/`, false, object.BOOLEAN_OBJ},
	}

	runVmTests(t, tests, false, false)
}

func TestDurationConversion(t *testing.T) {
	tests := []vmTestCase{
		{`duration(hash(dr/12Y 3M/)) == dr/12Y 3M/`, true, object.BOOLEAN_OBJ},
		{`duration(hash(dr/12Y 3M 5D T 7H 32M 4.789S/)) == dr/12Y 3M 5D T 7H 32M 4.789S/`, true, object.BOOLEAN_OBJ},
		{`duration(hash(dr/12Y 3M 5D T 7H 32M 4.009S/)) == dr/12Y 3M 5D T 7H 32M 4.009S/`, true, object.BOOLEAN_OBJ},
		{`duration(hash(dr/12Y 3M 5D T 7H 32M 4.000000009S/)) == dr/12Y 3M 5D T 7H 32M 4.000000009S/`, true, object.BOOLEAN_OBJ},
		{`duration(hash(dr/12Y 3M 5D T 7H 32M 4.000000009S/)) != dr/12Y 3M 5D T 7H 32M 4.000000001S/`, true, object.BOOLEAN_OBJ},
		{`duration(dr/12Y 3M/) == dr/12Y 3M/`, true, object.BOOLEAN_OBJ}, // pass through
		{`duration("12Y 3M") == dr/12Y 3M/`, true, object.BOOLEAN_OBJ},   // from string
		{`duration("P12Y3M") == dr/12Y 3M/`, true, object.BOOLEAN_OBJ},   // from string
		{`duration(1200) == dr/T0.0000012S/`, true, object.BOOLEAN_OBJ},  // from integer

		{`string(dr/T0.000000009S/) == "PT0.000000009S"`, true, object.BOOLEAN_OBJ},
		{`string(dr/T0.00000009S/) == "PT0.00000009S"`, true, object.BOOLEAN_OBJ},
		{`string(dr/12Y 3M 5D T 7H 32M 4.000000009S/) == "P12Y3M5DT7H32M4.000000009S"`, true, object.BOOLEAN_OBJ},

		// if converting between total nanoseconds and duration ...
		// there's no perfect way to define this ...
		// 31557600000000000 nanoseconds == 1 "year" or 365.25 days
		// 2629800000000000 nanoseconds == 1 "month" or 30.4375 days

		{`duration(31557600000000000) == dr/1Y/`, true, object.BOOLEAN_OBJ},
		{`duration(2629800000000000) == dr/1M/`, true, object.BOOLEAN_OBJ},
		{`duration(604800000000000) == dr/1W/`, true, object.BOOLEAN_OBJ},
		{`duration(86400000000000) == dr/1D/`, true, object.BOOLEAN_OBJ},
		{`duration(3600000000000) == dr/T1H/`, true, object.BOOLEAN_OBJ},
		{`duration(60000000000) == dr/T1M/`, true, object.BOOLEAN_OBJ},
		{`duration(1000000000) == dr/T1S/`, true, object.BOOLEAN_OBJ},
		{`duration(1) == dr/T0.000000001S/`, true, object.BOOLEAN_OBJ},

		{`31557600000000000 == number(dr/1Y/)`, true, object.BOOLEAN_OBJ},
		{`2629800000000000 == number(dr/1M/)`, true, object.BOOLEAN_OBJ},
		{`604800000000000 == number(dr/1W/)`, true, object.BOOLEAN_OBJ},
		{`86400000000000 == number(dr/1D/)`, true, object.BOOLEAN_OBJ},
		{`3600000000000 == number(dr/T1H/)`, true, object.BOOLEAN_OBJ},
		{`60000000000 == number(dr/T1M/)`, true, object.BOOLEAN_OBJ},
		{`1000000000 == number(dr/T1S/)`, true, object.BOOLEAN_OBJ},
		{`1 == number(dr/T0.000000001S/)`, true, object.BOOLEAN_OBJ},

		{`duration(2 * 31557600000000000 + 3 * 2629800000000000 + 7 * 86400000000000) == 
			dr/2Y 3M 7D/`, true, object.BOOLEAN_OBJ},

		{`2 * 31557600000000000 + 3 * 2629800000000000 + 7 * 86400000000000 == 
			number(dr/2Y 3M 7D/)`, true, object.BOOLEAN_OBJ},

		{`duration(dr/T1M/) == dr/T1M/`, true, object.BOOLEAN_OBJ},
		{`number(dr/T1M/) == dr/T1M/`, false, object.BOOLEAN_OBJ},
	}

	runVmTests(t, tests, false, false)
}

func TestCallingBuiltInsFromBuiltIns(t *testing.T) {
	tests := []vmTestCase{
		{`map([[], "Jim", {7: 14}], by=len)`, []int{0, 3, 1}, object.LIST_OBJ},

		{`map({1: "", 7: "abc", 14: [1, 2]}, by=len)`,
			[][]object.Object{
				{object.NumberFromInt(1), object.NumberFromInt(0)},
				{object.NumberFromInt(7), object.NumberFromInt(3)},
				{object.NumberFromInt(14), object.NumberFromInt(2)},
			},
			object.HASH_OBJ},

		{`map [[], "Jim", {7: 14}], by=len`, []int{0, 3, 1}, object.LIST_OBJ},

		// sort from single parameter function
		{`sort(["abcd", "ab", "abc", "zzzzzzz"], by=len)`,
			[]string{"ab", "abc", "abcd", "zzzzzzz"}, object.LIST_OBJ,
		},
	}

	runVmTests(t, tests, false, false)
}

func TestCallingRegexFromBuiltIns(t *testing.T) {
	tests := []vmTestCase{
		{`filter([16, 16, 25, 36, 42, 29, 49], by=re/[19]/)`,
			[]int{16, 16, 29, 49}, object.LIST_OBJ,
		},
	}

	runVmTests(t, tests, false, false)
}

func TestCallingCompiledFunctionsFromBuiltIns(t *testing.T) {
	tests := []vmTestCase{
		{`map([7, 9, 10], by=fn(x) {x * 2})`, []int{14, 18, 20}, object.LIST_OBJ},

		{`val x = fn(y) { (y ^/ 2) >= 4 }
			filter([16, 14, 16, 13, 12, 25, 36, 42, 29, 49], by=x)`,
			[]int{16, 16, 25, 36, 42, 29, 49}, object.LIST_OBJ,
		},

		{`val x = fn(y) { y >= 49 }
			filter({1: 49, 2: 35, 3: 70}, by=x)`,
			[][]object.Object{
				{object.NumberFromInt(1), object.NumberFromInt(49)},
				{object.NumberFromInt(3), object.NumberFromInt(70)},
			},
			object.HASH_OBJ,
		},

		// call a built-in from a compiled function that is called from a built-in...
		{`val x = fn(y) { len(split(y ^/ 2, delim=".")) == 1 }	
		  # true if the square root of a number is an integer
		  # There are probably better ways to check this.
	
			filter([16, 14, 16, 13, 12, 25, 36, 42, 29, 49], by=x)`,
			[]int{16, 16, 25, 36, 49}, object.LIST_OBJ,
		},

		{`sort([16, 14, 16, 13, 12, 25, 36, 42, 29, 49], by=fn(a, b) { a < b })`,
			[]int{12, 13, 14, 16, 16, 25, 29, 36, 42, 49}, object.LIST_OBJ,
		},
	}

	runVmTests(t, tests, false, false)
}

func TestZipFunction(t *testing.T) {
	tests := []vmTestCase{
		// zip without a function
		{"zip([], [])",
			[]int{}, object.LIST_OBJ,
		},
		{"zip([1, 3], [2, 4])",
			[]int{1, 2, 3, 4}, object.LIST_OBJ,
		},
		{"zip([1, 3], [2, 4], [6, 7])",
			[]int{1, 2, 6, 3, 4, 7}, object.LIST_OBJ,
		},

		{"zip([1, 2, 3], 2..4)",
			[]int{1, 2, 2, 3, 3, 4}, object.LIST_OBJ,
		},
		{"zip(1..3, 2..4)",
			[]int{1, 2, 2, 3, 3, 4}, object.LIST_OBJ,
		},

		// zip with a function
		{"zip([], [], by=fn(x, y) {[x, y]})",
			[]int{}, object.LIST_OBJ,
		},
		{"zip([1, 3], [2, 4], by=fn(x, y) { [x, y] })",
			[]int{1, 2, 3, 4}, object.LIST_OBJ,
		},
		{"zip([1, 3], [2, 4], by=fn(x, y) { [x + 7, y * 3] })",
			[]int{8, 6, 10, 12}, object.LIST_OBJ,
		},
		{"zip([1, 3], [2, 4], by=fn(x, y) { if(y > 3: 123; [x, y]) })",
			[]int{1, 2, 123}, object.LIST_OBJ,
		},

		{"zip(1..3, 7..9, by=fn(x, y) { [x, y] })", // function redundant in this case, but just a test
			[]int{1, 7, 2, 8, 3, 9}, object.LIST_OBJ,
		},
		{"zip(1..3, 7..9, by=fn(x, y) { if(y rem 2 == 0: []; [x, y]) })",
			[]int{1, 7, 3, 9}, object.LIST_OBJ,
		},
	}

	runVmTests(t, tests, false, false)
}

func TestMapFunction(t *testing.T) {
	tests := []vmTestCase{
		{`map([7, 9, 10], by=fn(x) { x * 2 })`, []int{14, 18, 20}, object.LIST_OBJ},
		{`map(1..3, by=fn(x) { x * 2 })`, []int{2, 4, 6}, object.LIST_OBJ},
		{`map(7..3, by=fn(x) { x * 2 })`, []int{14, 12, 10, 8, 6}, object.LIST_OBJ},

		{`map({1: 7, 2: 9, 3: 10}, by=fn(x) { x * 2 })`,
			[][]object.Object{
				{object.NumberFromInt(1), object.NumberFromInt(14)},
				{object.NumberFromInt(2), object.NumberFromInt(18)},
				{object.NumberFromInt(3), object.NumberFromInt(20)},
			},
			object.HASH_OBJ,
		},

		{`val x = fn(y) { if y >? 21 { y } else { -y } }
			map([7, 21, 35, 49], by=x)`,
			[]int{-7, -21, 35, 49}, object.LIST_OBJ,
		},

		// map multiple
		{`map([7, 9, 10], [7, 11, 14], by=fn{*})`, []int{49, 99, 140}, object.LIST_OBJ},
		{`map(1..3, [7, 8, 9], by=fn{*})`, []int{7, 16, 27}, object.LIST_OBJ},
		{`map(1..3, 9..7, by=fn{*})`, []int{9, 16, 21}, object.LIST_OBJ},

		// map with list of functions
		{ // multiply every second one by 2
			`val f = fn{*2}
		 	 map([7, 9, 10], by=[fn(x) { x }, f])`,
			[]int{7, 18, 10}, object.LIST_OBJ},

		{ // multiply every second one by 2
			`val f = fn{*2}
		 	 map([7, 9, 10, 11, 12, 13, 14, 15], by=[fn(x) { x }, f])`,
			[]int{7, 18, 10, 22, 12, 26, 14, 30}, object.LIST_OBJ},
		{ // multiply every second one by 2; use no-op for first
			`map([7, 9, 10, 11, 12, 13, 14, 15], by=[_, fn{*2}])`,
			[]int{7, 18, 10, 22, 12, 26, 14, 30}, object.LIST_OBJ},

		{ // multiply every third one by 2; use no-op for first and second
			`map([7, 9, 10, 11, 12, 13, 14, 15], by=[_, _, fn{*2}])`,
			[]int{7, 9, 20, 11, 12, 26, 14, 15}, object.LIST_OBJ},

		// map multiple with list of functions
		{`map([7, 9, 10], [13, 14, 15], by=[fn{*}, fn{+}])`,
			[]int{91, 23, 150}, object.LIST_OBJ},
		{`map([7, 9, 10, 11], [13, 14, 15, 16], by=[fn{*}, fn{+}])`,
			[]int{91, 23, 150, 27}, object.LIST_OBJ},

		// map multiple including a no-op
		{`string(map([7, 9, 10, 11], [13, 14, 15, 16], by=[fn{*}, _]))`,
			`[91, [9, 14], 150, [11, 16]]`, object.STRING_OBJ},
	}

	runVmTests(t, tests, false, false)
}

func TestMapXFunction(t *testing.T) {
	tests := []vmTestCase{
		{`mapX([1, 2], [3, 4], by=fn(x...) {x})`,
			[][]int{{1, 3}, {1, 4}, {2, 3}, {2, 4}},
			object.LIST_OBJ,
		},
		{`mapX([1, 2, 3], 30, [500, 100], by=fn(x...) {x})`,
			[][]int{{1, 30, 500}, {1, 30, 100}, {2, 30, 500}, {2, 30, 100}, {3, 30, 500}, {3, 30, 100}},
			object.LIST_OBJ,
		},
		{`mapX(fw/a b/, fw/c d/, by=fn(x...) {x})`,
			[][]string{{"a", "c"}, {"a", "d"}, {"b", "c"}, {"b", "d"}},
			object.LIST_OBJ,
		},
		{`mapX([1, 2], by=fn(x...) {x})`,
			[][]int{{1}, {2}},
			object.LIST_OBJ,
		},
		{`mapX(7, by=fn(x...) {x})`,
			[][]int{{7}},
			object.LIST_OBJ,
		},
		{`mapX(7, [14, 21], by=fn(x...) {x})`,
			[][]int{{7, 14}, {7, 21}},
			object.LIST_OBJ,
		},

		{`mapX([1, 2], [3, 4], by=fn{*})`,
			[]int{3, 4, 6, 8},
			object.LIST_OBJ,
		},
		{`mapX([1, 2, 3], 30, [500, 100], by=fn(x, y, z) { x + y + z })`,
			[]int{531, 131, 532, 132, 533, 133},
			object.LIST_OBJ,
		},
		{`mapX(fw/a b/, fw/c d/, by=fn{~})`,
			[]string{"ac", "ad", "bc", "bd"},
			object.LIST_OBJ,
		},
		{`mapX(fw/a b/, fw/c d/, fw/e f/, by=fn(x, y, z) { x ~ y ~ z })`,
			[]string{"ace", "acf", "ade", "adf", "bce", "bcf", "bde", "bdf"},
			object.LIST_OBJ,
		},
	}

	runVmTests(t, tests, false, false)
}

func TestArgumentExpansion(t *testing.T) {
	tests := []vmTestCase{
		{`zip([[7], [14], [21]]...)`,
			[]int{7, 14, 21},
			object.LIST_OBJ,
		},

		{`zip(mapX([7], [14, 21], by=fn(x...) {x})...)`,
			[]int{7, 7, 14, 21},
			object.LIST_OBJ,
		},

		{`zip([1, 2], mapX([7], [14, 21], by=fn(x...) {x})...)`,
			[]int{1, 7, 7, 2, 14, 21},
			object.LIST_OBJ,
		},

		{`val x = [[21, 7], [42, 14], [96, 35]]
		  zip(x...)`,
			[]int{21, 42, 96, 7, 14, 35},
			object.LIST_OBJ,
		},
	}

	runVmTests(t, tests, false, false)
}

func TestParameterExpansion(t *testing.T) {
	tests := []vmTestCase{
		{`val f = fn(x...) { x }
		  f(1234)`,
			[]int{1234}, object.LIST_OBJ,
		},
		{`val f = fn(x...) { x }
		  f(1234, 5678)`,
			[]int{1234, 5678}, object.LIST_OBJ,
		},
		{`val f = fn(a, x...) { x }
		  f(7, 21, 32, 45)`,
			[]int{21, 32, 45}, object.LIST_OBJ,
		},
		{`val f = fn(a, x...) { x ~ [a] }
		  f(7, 21, 32, 45)`,
			[]int{21, 32, 45, 7}, object.LIST_OBJ,
		},
		{`val f = fn(a, x...[2..3]) { x ~ [a] }
		  f(7, 21, 32, 45)`,
			[]int{21, 32, 45, 7}, object.LIST_OBJ,
		},
		{`val f = fn(a, x ...[3]) { x ~ [a] }
		  f(7, 21, 32, 45)`,
			[]int{21, 32, 45, 7}, object.LIST_OBJ,
		},
		{`val f = fn(a, x...[1..4]) { x ~ [a] }
		  f(7, 21, 32, 45)`,
			[]int{21, 32, 45, 7}, object.LIST_OBJ,
		},
		{`val f = fn(a, x...[0..4]) { x ~ [a] }
		  f(7, 21, 32, 45)`,
			[]int{21, 32, 45, 7}, object.LIST_OBJ,
		},
		{`val f = fn(a, x...[1..]) { x ~ [a] }
		  f(7, 21, 32, 45)`,
			[]int{21, 32, 45, 7}, object.LIST_OBJ,
		},

		{`val f = fn(a, x...[0..]) { x ~ [a] }
		  f(7, 21, 32, 45)`,
			[]int{21, 32, 45, 7}, object.LIST_OBJ,
		},
		{`val f = fn(a, x...[0..]) { x ~ [a] }
		  f(7)`,
			[]int{7}, object.LIST_OBJ,
		},

		{`val f = fn(a, x...[0..]) { x ~ [a] }
		  f(7)`,
			[]int{7}, object.LIST_OBJ,
		},
		{`val f = fn(x...[0..]) { x }
		  f()`,
			[]int{}, object.LIST_OBJ,
		},
	}

	runVmTests(t, tests, false, false)
}

func TestParameterExpansionMinMax(t *testing.T) {
	tests := []vmTestCase{
		{`val f = fn(x...) { x }
		  max(f)`,
			common.ArgCountMax, object.NUMBER_OBJ,
		},
		{`val f = fn(x...) { x }
		  min(f)`,
			0, object.NUMBER_OBJ,
		},
		{`val f = fn(x...) { x }
		  minmax(f)`,
			[]int64{0, common.ArgCountMax}, object.RANGE_OBJ,
		},

		{`val f = fn(a, x...) { x }
		  max(f)`,
			common.ArgCountMax, object.NUMBER_OBJ,
		},
		{`val f = fn(a, x...) { x }
		  min(f)`,
			1, object.NUMBER_OBJ,
		},
		{`val f = fn(a, x...) { x }
		  minmax(f)`,
			[]int64{1, common.ArgCountMax}, object.RANGE_OBJ,
		},

		{`val f = fn(x...[0..]) { x }
		  max(f)`,
			common.ArgCountMax, object.NUMBER_OBJ,
		},
		{`val f = fn(x...[0..]) { x }
		  min(f)`,
			0, object.NUMBER_OBJ,
		},
		{`val f = fn(x...[0..]) { x }
		  minmax(f)`,
			[]int64{0, common.ArgCountMax}, object.RANGE_OBJ,
		},

		{`val f = fn(a, b, x...[0..]) { x }
		  max(f)`,
			common.ArgCountMax, object.NUMBER_OBJ,
		},
		{`val f = fn(a, b, x...[0..]) { x }
		  min(f)`,
			2, object.NUMBER_OBJ,
		},
		{`val f = fn(a, b, x...[0..]) { x }
		  minmax(f)`,
			[]int64{2, common.ArgCountMax}, object.RANGE_OBJ,
		},

		{`val f = fn(x...[2..10]) { x }
		  max(f)`,
			10, object.NUMBER_OBJ,
		},
		{`val f = fn(x...[2..10]) { x }
		  min(f)`,
			2, object.NUMBER_OBJ,
		},
		{`val f = fn(x...[2..10]) { x }
		  minmax(f)`,
			[]int64{2, 10}, object.RANGE_OBJ,
		},

		{`val f = fn(a, b, x...[2..10]) { x }
		  max(f)`,
			12, object.NUMBER_OBJ,
		},
		{`val f = fn(a, b, x...[2..10]) { x }
		  min(f)`,
			4, object.NUMBER_OBJ,
		},
		{`val f = fn(a, b, x...[2..10]) { x }
		  minmax(f)`,
			[]int64{4, 12}, object.RANGE_OBJ,
		},
	}

	runVmTests(t, tests, false, false)
}

func TestFoldAndFoldFromFunctions(t *testing.T) {
	tests := []vmTestCase{
		// fold using function with implied parameters
		{"fold([16, 14, 16, 13, 12, 25, 36], by=fn{*})",
			"503193600", object.NUMBER_OBJ,
		},

		// factorial using fold() and series()
		{"fold(series(7), by=fn{*})",
			"5040", object.NUMBER_OBJ,
		},

		// fold with multiple functions
		{"fold(series(7), by=[fn{*}, fn{+}])",
			"157", object.NUMBER_OBJ,
		},

		// fold from starting value
		{`fold(
			[10, 2, 3, 4],
			by=fn{*}, 
			init=111, 
			)`,
			"26640", object.NUMBER_OBJ,
		},
		{`fold(
			5..7,
			by=fn{*}, 
			init=111, 
			)`,
			"23310", object.NUMBER_OBJ,
		},

		// fold on multiple lists (or ranges); requires starting value (can't use fold() for multiple lists)
		{`fold(
			[1, 2, 3], 
			[10, 5, 6],
			by=fn(from, a, b) { from + a * b },
			init=1, 
			)`,
			"39", object.NUMBER_OBJ,
		},
		{`fold(
			[1, 2, 3], 
			[10, 5, 6],
			[3.2, 7, 7],
			by=fn(from, a, b, c) { from + a * b - c },
			init=1, 
			)`,
			"21.8", object.NUMBER_OBJ,
		},

		{`fold(
			1 .. 3, 
			[10, 5, 6],
			[3.2, 7, 7],
			by=fn(from, a, b, c) { from + a * b - c },
			init=1, 
			)`,
			"21.8", object.NUMBER_OBJ,
		},

		// fold from with list of functions
		{`fold(
			[10, 2, 3, 4],
			by=[fn{*}], 
			init=111, 
			)`,
			"26640", object.NUMBER_OBJ,
		},
		{`fold(
			[10, 2, 3, 4],
			by=[fn{*}, fn{+}], 
			init=111, 
			)`,
			"3340", object.NUMBER_OBJ,
		},

		// fold from with list of functions and multiple lists
		{`fold(
			[10, 2, 3, 4], 
			[7, 8, 9, 10],
			by=[fn(a, b, c) { a + b * c }, fn(a, b, c) { a+b+c }], 
			init=111, 
			)`,
			"232", object.NUMBER_OBJ,
		},
	}

	runVmTests(t, tests, false, false)
}

func TestValueFromBlock(t *testing.T) {
	tests := []vmTestCase{
		{`{ 7 } * 7`, "49", object.NUMBER_OBJ},
		{`{ var x = 7 } * 7`, "49", object.NUMBER_OBJ},
	}

	runVmTests(t, tests, false, false)
}

func TestClosures(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `
		val newClosure = fn(a) {
			fn(){ a }
		}
		val closure = newClosure(99)
		closure()
		`,
			expected:     "99",
			expectedType: object.NUMBER_OBJ,
		},
		{
			input: `
		val newAdder = fn(a, b) {
			fn(c) { a + b + c }
		}
		val adder = newAdder(1, 2)
		adder(8)
		`,
			expected:     "11",
			expectedType: object.NUMBER_OBJ,
		},
		{
			input: `
		val newAdder = fn(a, b) {
			val c = a + b
			fn(d) { c + d }
		}
		val adder = newAdder(1, 2)
		adder(8)
		`,
			expected:     "11",
			expectedType: object.NUMBER_OBJ,
		},
		{
			input: `
		val newAdderOuter = fn(a, b) {
			val c = a + b;
			fn(d) {
				val e = d + c;
				fn(f) { e + f; };
			};
		};
		val newAdderInner = newAdderOuter(1, 2)
		val adder = newAdderInner(3);
		adder(8);
		`,
			expected:     "14",
			expectedType: object.NUMBER_OBJ,
		},
		{
			input: `
		val a = 1;
		val newAdderOuter = fn(b) {
			fn(c) {
				fn(d) { a + b + c + d };
			};
		};
		val newAdderInner = newAdderOuter(2)
		val adder = newAdderInner(3);
		adder(8);
		`,
			expected:     "14",
			expectedType: object.NUMBER_OBJ,
		},
		{
			input: `
		val newClosure = fn(a, b) {
			val one = fn(){ a; };
			val two = fn(){ b; };
			fn() { one() + two(); };
		};
		val closure = newClosure(9, 90);
		closure();
		`,
			expected:     "99",
			expectedType: object.NUMBER_OBJ,
		},
	}

	runVmTests(t, tests, false, false)
}

func TestCallingClosuresFromBuiltins(t *testing.T) {
	tests := []vmTestCase{
		{
			// this one not a closure, but we test it here to prepare for the next test
			"map([1, 2, 3], by=fn{+2})",
			[]int{3, 4, 5}, object.LIST_OBJ,
		},
		{
			`val v = 120
			 val f = fn(x) { x - v }
			 map([1, 2, 3], by=f)`,
			[]int{-119, -118, -117}, object.LIST_OBJ,
		},
		{
			`val v = 120
			 map([1, 2, 3], by=fn(x) { x + v })`,
			[]int{121, 122, 123}, object.LIST_OBJ,
		},
	}

	runVmTests(t, tests, false, false)
}

func TestRecursiveFunctions(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `
				val fibonacci = fn(x number) { switch x {
					case < 2: x
					default: fibonacci(x - 1) + fibonacci(x - 2)
				}}
				fibonacci(15)
				`,
			expected:     "610",
			expectedType: object.NUMBER_OBJ,
		},

		{
			input: `
				val fact = fn(n) { if n > 1 { n * fact(n - 1) } else { 1 } }
				fact(7)
				`,
			expected:     "5040",
			expectedType: object.NUMBER_OBJ,
		},

		{ // recursive function enclosed in another function
			input: `
				fn() {
					val fact = fn(n) { if n > 1 { n * fact(n - 1) } else { 1 } }
					fact(7)
				}()
				`,
			expected:     "5040",
			expectedType: object.NUMBER_OBJ,
		},

		{
			input: `
				# This is just a test. We would use the built-in map() function instead.
				val zmap = fn(f, zlist) {
				    val iter = fn(remaining, accumulated) {
				        if len(remaining) == 0 {
				            accumulated
				        } else {
				            iter(less(remaining, of=1), more(accumulated, f(remaining[1])))
				       }
				    }
				    iter(zlist, [])
				}
				val doubled = zmap(fn{* 2}, [1, 2, 3, 4])
				`,
			expected:     []int{2, 4, 6, 8},
			expectedType: object.LIST_OBJ,
		},

		{
			input: `
				# This is just a test. We would use the built-in fold() or foldfrom() functions instead.
				val zfold = fn(zlist, initial, f) {
				    val iter = fn(zlist, result) { if len(zlist) == 0 {
			            result
			        } else {
			            iter(less(zlist, of=1), f(result, zlist[1]))
			        }}

				    iter(zlist, initial)
				}

				val sum = fn(zlist) { zfold(zlist, 0, fn{+}) }
				sum([1, 2, 3, 1])
				`,
			expected:     "7",
			expectedType: object.NUMBER_OBJ,
		},

		{ // more nested: call chickendinner from chickendinner
			input: `
				val winner = fn(x) {
					val winner2 = fn(x) {
						val chickendinner = fn(n) { if n > 1 { n * chickendinner(n - 1) } else { 1 } }
						chickendinner(x)
					}
					winner2(x)
				}
				winner(7)
			`,
			expected:     "5040",
			expectedType: object.NUMBER_OBJ,
		},
		{ // more nested 2: call winner2 from chickendinner
			input: `
				val winner = fn(x) {
					val winner2 = fn(x) {
						val chickendinner = fn(n) { if n > 1 { n * winner2(n - 1) } else { 1 } }
						chickendinner(x)
					}
					winner2(x)
				}
				winner(7)
			`,
			expected:     "5040",
			expectedType: object.NUMBER_OBJ,
		},
		{ // more nested 3: call winner from chickendinner
			input: `
				val winner = fn(x) {
					val winner2 = fn(x) {
						val chickendinner = fn(n) { if n > 1 { n * winner(n - 1) } else { 1 } }
						chickendinner(x)
					}
					winner2(x)
				}
				winner(7)
			`,
			expected:     "5040",
			expectedType: object.NUMBER_OBJ,
		},
	}

	runVmTests(t, tests, false, false)
}

func TestRecursiveFunctionsUsingSelfToken(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `
				val fibonacci = fn(x) { switch x {
					case 0, 1: x
					default: fn((x - 1)) + fn((x - 2))
				}}
				fibonacci(15)
				`,
			expected:     "610",
			expectedType: object.NUMBER_OBJ,
		},

		{
			input: `
				val fact = fn(n) { if n > 1 { n * fn((n - 1)) } else { 1 } }
				fact(7)
				`,
			expected:     "5040",
			expectedType: object.NUMBER_OBJ,
		},

		{ // recursion using self token without assignment and with implied parameter n
			input: `
				fn(n) {if n > 1 { n * fn((n - 1)) } else { 1 }}(7)
				`,
			expected:     "5040",
			expectedType: object.NUMBER_OBJ,
		},
	}

	runVmTests(t, tests, false, false)
}

func TestTryCatch(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `
				throw 123
				100

				catch[e] { switch e["msg"] {
					case "100": 1
					case "123": 2
					case "234": 3
					default: throw e
				}}
			`,
			expected:     "2",
			expectedType: object.NUMBER_OBJ,
		},

		{
			input: `
						100

						catch { switch _err["msg"] {
							case "100": 1
							case "123": 2
							case "234": 3
							default: throw
						}}
					`,
			expected:     "100",
			expectedType: object.NUMBER_OBJ,
		},

		{
			// no error
			input: `
				val x = 123
				catch { if _err["cat"] == "math" { 890 } else { 456 } }
				`,
			expected:     "123",
			expectedType: object.NUMBER_OBJ,
		},

		{
			input: `
				val x = 123 / 0
				catch { if _err["cat"] == "math" { 890 } else { 456 } }
				`,
			expected:     "890",
			expectedType: object.NUMBER_OBJ,
		},

		{
			input: `
				val x = 123 / 1
				catch {
					# print for analysis
					#write(_err["cat"], ": ", _err["msg"], " (", _err["src"], ")\N")

					if _err["cat"] == "math" { 890 } else { 456 }
				}
				`,
			expected:     "123",
			expectedType: object.NUMBER_OBJ,
		},

		{
			input: `
				val x = 123 / 0
				catch { if _err["cat"] == "math" { 890 } else { 456 } }
				catch { 789 }
				`,
			expected:     "890",
			expectedType: object.NUMBER_OBJ,
		},

		// simple catch
		{
			input: `
				val x = 123 / 0
				catch: 789
				`,
			expected:     "789",
			expectedType: object.NUMBER_OBJ,
		},
		{
			input: `
				val x = 123 / 0
				catch: if _err["cat"] == "math" { 890 } else { 456 }
				catch: 789
				`,
			expected:     "890",
			expectedType: object.NUMBER_OBJ,
		},
		{
			input: `
						100

						catch: switch _err["msg"] {
							case "100": 1
							case "123": 2
							case "234": 3
							default: throw
						}
					`,
			expected:     "100",
			expectedType: object.NUMBER_OBJ,
		},
		{
			input: `
						100

						catch[e]: switch e["msg"] {
							case "100": 1
							case "123": 2
							case "234": 3
							default: throw
						}
					`,
			expected:     "100",
			expectedType: object.NUMBER_OBJ,
		},

		{ // using same _err variable name in catches in sequence
			input: `
				val x = 123 / 0
				catch { if _err["cat"] == "math" { 890 } else { 456 } }
				val y = 789
				456
				78 / 0
				catch { y }
				`,
			expected:     "789",
			expectedType: object.NUMBER_OBJ,
		},

		{
			input: `
				val x = fn() {
					123 / 0
					catch { if _err["cat"] == "math" { 890 } else { 456 } }
					159
					val y = 789
					78 / 0
					catch { y }
				}()
				x - 89 + 77
				`,
			expected:     "777",
			expectedType: object.NUMBER_OBJ,
		},

		{ // return correctly from function frame out of a try section?
			input: `
				val tryme = fn() {
					if true { return 7 }
					catch { 90 }
					777
				}
				tryme()
				`,
			expected:     "7",
			expectedType: object.NUMBER_OBJ,
		},

		{ // return correctly from function frame out of a catch?
			input: `
				val tryme = fn() {
					1 / 0
					catch { return 90 }
					7
				}
				tryme()
				`,
			expected:     "90",
			expectedType: object.NUMBER_OBJ,
		},

		{ // embedded catch within catch reusing same name
			input: `
				1 / 0
				catch[e] {
					7 / 0
					catch[e] {
						80
					}
				}
				`,
			expected:     "80",
			expectedType: object.NUMBER_OBJ,
		},

		// fixed bug discovered January 2020
		// moving decoupling assignment into try block...
		{
			`val data = [[1, 2], [3, 4], [5, 6], [7, 8]]
			 var sum = 0
			 for test in data {
				 var x, y = test
				 sum += x + y

				 catch { writeln _err }
			 }
			 sum`,
			36, object.NUMBER_OBJ,
		},
	}

	runVmTests(t, tests, false, false)
}

func TestTryCatchElse(t *testing.T) {
	// catch else not available on *simple catch*
	tests := []vmTestCase{
		{
			input: `7 / 0
					catch {
						9
					} else {
						14
					}
					`,
			expected:     "9",
			expectedType: object.NUMBER_OBJ,
		},

		{
			input: `7 / 2
					catch {
						9
					} else {
						14
					}
					`,
			expected:     "14",
			expectedType: object.NUMBER_OBJ,
		},

		// using else if on catch
		{`1 / 2
		  catch {
				3
		  } else if 1 > 5 {
			    4
		  } else {
			 	7
		  }`,
			7,
			object.NUMBER_OBJ,
		},
		{`1 / 2
		  catch {
				3
		  } else if 1 < 5 {
			    4
		  } else {
			 	7
		  }`,
			4,
			object.NUMBER_OBJ,
		},
		{`1 / 2
		  catch {
				3
		  } else if 1 > 5 {
			    4
		  }`,
			nil,
			object.NULL_OBJ,
		},
		{`1 / 2
		  catch {
				3
		  } else if 1 < 5 {
			    4
		  }`,
			4,
			object.NUMBER_OBJ,
		},
		{`1 / 0
		  catch {
				3
		  } else if 1 > 5 {
			    4
		  } else {
			 	7
		  }`,
			3,
			object.NUMBER_OBJ,
		},
		{`1 / 0
		  catch {
				3
		  } else if 1 > 5 {
			    4
		  }`,
			3,
			object.NUMBER_OBJ,
		},
	}

	runVmTests(t, tests, false, false)
}

func TestLateBinding(t *testing.T) {
	tests := []vmTestCase{
		{
			input:        `_env["GOOS"]`,
			expected:     runtime.GOOS,
			expectedType: object.STRING_OBJ,
		},
		{
			input:        `_args`,
			expected:     []string{},
			expectedType: object.LIST_OBJ,
		},
		{
			input:        `_file`,
			expected:     "",
			expectedType: object.STRING_OBJ,
		},

		// ordering with late bindings
		{
			input:        `_file ~ string(_args)`,
			expected:     "[]",
			expectedType: object.STRING_OBJ,
		},
	}

	runVmTests(t, tests, false, false)
}

func TestVariableScoping(t *testing.T) {
	tests := []vmTestCase{
		// for catch blocks
		{
			input:        `val x = 123; 1 / 0; catch { val x = 78 }`,
			expected:     "78",
			expectedType: object.NUMBER_OBJ,
		},
		{
			input:        `val x = 123; 1 / 0; catch { val x = 78 }; x`,
			expected:     "123",
			expectedType: object.NUMBER_OBJ,
		},
		{
			input:        `var x = 123; 1 / 0; catch { x = 78 }; x`,
			expected:     "78",
			expectedType: object.NUMBER_OBJ,
		},
		{
			input:        `var x = 123; 1 / 1; catch { x = 78 }; x`,
			expected:     "123",
			expectedType: object.NUMBER_OBJ,
		},

		// reusing the same error variable in nested catches (not possible without scoping)
		{
			input:        `var x = 123; 1 / 1; catch { x = 78; 1 / 0; catch { _err["cat"] }}; x`,
			expected:     "123",
			expectedType: object.NUMBER_OBJ,
		},
		{
			input:        `var x = 123; 1 / 1; catch[e] { x = 78; 1 / 0; catch[e] { e["cat"] }}`,
			expected:     "1",
			expectedType: object.NUMBER_OBJ,
		},
		{
			input:        `var x = 123; 1 / 0; catch[e] { x = 78; 1 / 0; catch[e] { e["cat"] }}`,
			expected:     "math",
			expectedType: object.STRING_OBJ,
		},

		// for scope blocks
		{
			input:        `{ var x = 123; { var x = 7; x *= 2; var y = 12 }; x }`,
			expected:     "123",
			expectedType: object.NUMBER_OBJ,
		},
		{
			input:        `var x = 123; { var x = 7 }; x`,
			expected:     "123",
			expectedType: object.NUMBER_OBJ,
		},
		{
			input:        `{{{var x = 123; { var x = 7 }; x}}}`,
			expected:     "123",
			expectedType: object.NUMBER_OBJ,
		},
		{
			input:        `{{{var x = 123; { x = 7 }; x}}}`,
			expected:     "7",
			expectedType: object.NUMBER_OBJ,
		},
		{
			input:        `var x = 123; { x = 7 }; x`,
			expected:     "7",
			expectedType: object.NUMBER_OBJ,
		},
		{
			input:        `{ val y = 789; 78 / 0; catch { y } }`,
			expected:     "789",
			expectedType: object.NUMBER_OBJ,
		},
		{
			input:        `val x = 7; { val x = 123; switch x { case 123: 1; default: 2 } }; x`,
			expected:     "7",
			expectedType: object.NUMBER_OBJ,
		},
		{
			input:        `val x = 7; { val x = 123; switch x { case 123: 1; default: 2 } }`,
			expected:     "1",
			expectedType: object.NUMBER_OBJ,
		},
		{
			input:        `val x = 7; { val x = 123; switch x { case 100: 1; default: x } }`,
			expected:     "123",
			expectedType: object.NUMBER_OBJ,
		},

		// scoped if/switch expressions (contains declarations)
		{
			input:        `if val x = 123 { 1 } else { 2 }`,
			expected:     1,
			expectedType: object.NUMBER_OBJ,
		},
		{
			input:        `val x = 7; if val x = 123 { 1 } else { 2 }; x`,
			expected:     7,
			expectedType: object.NUMBER_OBJ,
		},
		{
			input:        `val x = 7; if true { val x = 789 } else { 2 }; x`,
			expected:     7,
			expectedType: object.NUMBER_OBJ,
		},
		{
			input:        `val x = 7; switch x { case 7: val x = 213 ; default: x }; x`,
			expected:     7,
			expectedType: object.NUMBER_OBJ,
		},

		// scope with jumps
		{
			input:        `if val x = [] { true; throw x } else { false }`,
			expected:     false,
			expectedType: object.BOOLEAN_OBJ,
		},
		{
			input:        `fn() { if val x = 7 { true; return x } else { false }}()`,
			expected:     "7",
			expectedType: object.NUMBER_OBJ,
		},
		{
			input:        `if false { true; throw 7 } else { false }`,
			expected:     false,
			expectedType: object.BOOLEAN_OBJ,
		},
		{
			input:        `fn() { if false { true; return 7 } else { false }}()`,
			expected:     false,
			expectedType: object.BOOLEAN_OBJ,
		},

		{
			input:        `if val x = [1] { true } else { false; throw 7 }`,
			expected:     true,
			expectedType: object.BOOLEAN_OBJ,
		},
		{
			input:        `fn() { if val x = 7 { x } else { false; return 9 }}()`,
			expected:     "7",
			expectedType: object.NUMBER_OBJ,
		},
		{
			input:        `fn() { if false { throw 7 } else { val x = true; return x }}()`,
			expected:     true,
			expectedType: object.BOOLEAN_OBJ,
		},
		{
			input:        `fn() { if false { true } else { val x = false; return x }}()`,
			expected:     false,
			expectedType: object.BOOLEAN_OBJ,
		},

		// scope per block of if/switch
		{
			input: `val x = 7
							switch x {
								case < 100: val y = 80
								default: val y = 70
							}`,
			expected:     "80",
			expectedType: object.NUMBER_OBJ,
		},
		{
			input: `val x = 7
					if x > 100 { val x = 80 } else { val x = 70 }`,
			expected:     "70",
			expectedType: object.NUMBER_OBJ,
		},

		// assignments in switch test expressions extracted and moved in front and into a scope block...
		// {
		// 	input:        `var x = 7; switch x = x + 1 { case 100: 1; case 200: 2; default: 3 }; x`,
		// 	expected:     "8",
		// 	expectedType: object.NUMBER_OBJ,
		// },

		// assignment within if expressions
		{
			input:        `var x = 7; if x = null { 0 } else { 1 }`,
			expected:     "1",
			expectedType: object.NUMBER_OBJ,
		},
		{
			input:        `var x = 7; if x = null { 0 } else { 1 }; x`,
			expected:     7,
			expectedType: object.NUMBER_OBJ,
		},
		{
			input:        `if var x = 0 { 0 } else if var x = 7 { x }`,
			expected:     "7",
			expectedType: object.NUMBER_OBJ,
		},
		{
			input:        `if var x { 0 } else if var x = [] { 7 }`,
			expected:     nil,
			expectedType: object.NULL_OBJ,
		},
		{
			input:        `var x = []; if x { 0 } else if x = [1, 2] { x }`,
			expected:     []int{1, 2},
			expectedType: object.LIST_OBJ,
		},
		{
			input:        `var x = []; if x = [1, 2] { x } else if x { 0 }`,
			expected:     []int{1, 2},
			expectedType: object.LIST_OBJ,
		},
		{
			input:        `var x = []; if x { 0 } else if x = [14, 7] { x[2] }`,
			expected:     "7",
			expectedType: object.NUMBER_OBJ,
		},
		{
			input:        `var x = [1]; if x { 0 } else if x = [14, 7] { x[2] }`,
			expected:     "0",
			expectedType: object.NUMBER_OBJ,
		},
		{
			input:        `var x = []; if x { 0 } else if x = submatch("abc", by=re/a(b)c/) { x[1] }`,
			expected:     "b",
			expectedType: object.STRING_OBJ,
		},
		{
			input:        `{ var x = []; if x { 0 } else if x = submatch("abc", by=re/a(b)c/) { x[1] }}`,
			expected:     "b",
			expectedType: object.STRING_OBJ,
		},

		{
			input:        `if x, y = submatch("abc", by=re/a(b)(c)/) { x ~ "!" ~ y }`,
			expected:     "b!c",
			expectedType: object.STRING_OBJ,
		},

		// Reassignments should be possible within deeper scopes (if mutable and not past a function boundary).
		{
			input:        `var x = 123; 1 / 0; catch { x = 78; 0 }; x`,
			expected:     "78",
			expectedType: object.NUMBER_OBJ,
		},
		{
			input:        `var x = 123; 1 / 1; catch { x = 78; 0 }; x`,
			expected:     "123",
			expectedType: object.NUMBER_OBJ,
		},

		{
			input:        `var x = 123; if true { x = 7 }; x`,
			expected:     "7",
			expectedType: object.NUMBER_OBJ,
		},
		{
			input:        `var x = 123; if false { x = 7 }; x`,
			expected:     "123",
			expectedType: object.NUMBER_OBJ,
		},
		{
			input:        `var x = 123; if true { x = 7 } else { val x = 89 }; x`,
			expected:     "7",
			expectedType: object.NUMBER_OBJ,
		},

		{
			input:        `var x = 123; { x = 7 }; x`,
			expected:     "7",
			expectedType: object.NUMBER_OBJ,
		},
		{
			input:        `var x = 123; { if true { x = 7 } }; x`,
			expected:     "7",
			expectedType: object.NUMBER_OBJ,
		},
		{
			input:        `var x = 123; { if false { x = 7 } }; x`,
			expected:     "123",
			expectedType: object.NUMBER_OBJ,
		},
	}

	runVmTests(t, tests, false, false)
}

func TestSemiDeepCopy(t *testing.T) {
	tests := []vmTestCase{
		// using append
		{`var x = [1, 2, 3]
		   var y = [x]
		   x ~= [4, 5, 6, 7]
		   y[1]`,
			[]int{1, 2, 3}, object.LIST_OBJ,
		},
		{`var x = "abcd "
		   var y = [x]
		   x ~= "you know"
		   y[1]`,
			"abcd ", object.STRING_OBJ,
		},
		{`var x = {1: 123}
		   var y = [x]
		   x ~= {2: 456}
		   len y[1]`,
			1, object.NUMBER_OBJ,
		},
		
		// value passed to function
		{`val x = [1, 2, 3]
		  fn(a) {
			var b = a
			b[2] = 7
		  }(x)
		  x[2]`,
			2, object.NUMBER_OBJ,
		},
		// {`val x = [1, 2, 3]
		//   fn(var a) {
		// 	a[2] = 7
		//   }(x)
		//   x[2]`,
		// 	2, object.NUMBER_OBJ,
		// },
	
	}

	runVmTests(t, tests, false, false)
}

func TestAssignmentContexts(t *testing.T) {
	tests := []vmTestCase{
		// for loop init (implicit declaration)
		{`var x = 1
		  for i = 1 ; i < 7; i += 1 { x += 1 }
		  x`,
			7, object.NUMBER_OBJ,
		},
		{`var x = 1
		  for i, j = 1, 2 ; j < 7; j += 1 { x += 1 }
		  x`,
			6, object.NUMBER_OBJ,
		},

		// for loop increment
		{`var x = [1]
		  for i = 1 ; i < 7; i, x[1] = i+1, x[1]+1 {  }
		  x[1]`,
			7, object.NUMBER_OBJ,
		},

		// for loop value (implicit declaration)
		{`for[=1] i = 1 ; i < 7; i += 1 { _for += 1 }`,
			7, object.NUMBER_OBJ,
		},
		{`for[x=1] i = 1 ; i < 7; i += 1 { x += 1 }`,
			7, object.NUMBER_OBJ,
		},
		{`len for[x] i = 1 ; i < 7; i += 1 { x ~= [1] }`,
			6, object.NUMBER_OBJ,
		},

		// if test (implicit declaration)
		{`var a = 7
		  if x = 7 { a += x }
		  a`,
			14, object.NUMBER_OBJ,
		},
		{`var x = ""
		  if k, v = submatch("abcd:fish", by=re/(.+):(.+)/) {
		  	x = k ~ "+" ~ v
		  }
		  x`,
			"abcd+fish", object.STRING_OBJ,
		},
	}

	runVmTests(t, tests, false, false)
}

func TestModuleCompile(t *testing.T) {
	tests := []vmTestCase{
		{`module
		 mode divMaxScale = 7
		 val _main = fn() { 2/3 }`,
			"0.6666667", object.NUMBER_OBJ,
		},

		{`module
		 val _main = fn() { mult(7, 2) }
		 val mult = fn{*}
		 `,
			14, object.NUMBER_OBJ,
		},
	}

	runVmTests(t, tests, false, false)
}

// func TestBindings(t *testing.T) {
// 	tests := []vmTestCase{
// 		{ ``,
// 			"",
// 			object.NUMBER_OBJ,
// 		},
// 	}

// 	runVmTests(t, tests, false, false)
// }
