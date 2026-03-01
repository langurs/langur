// langur/test/speed_test.go

package test

import (
	"fmt"
	"langur/object"
	"testing"
	"time"
)

// to *very generally* see how changes affect the speed of the VM
// runtimes vary between tests

func TestSpeedByFibonacci(t *testing.T) {
	tests := []vmTestCase{
		{
			`val fibonacci = fn(x) { if(x < 2: x ; fn((x - 1)) + fn((x - 2))) }
		 	fibonacci(15)`,
			"610",
			object.NUMBER_OBJ,
		},
	}

	runVmTests(t, tests, false, true)
}

func TestSpeedIntegerAddition(t *testing.T) {
	tests := []vmTestCase{
		{
			`30000 + 40000`,
			"70000",
			object.NUMBER_OBJ,
		},
	}

	runVmTests(t, tests, false, true)
}

func TestSpeedIntegerMultiplication(t *testing.T) {
	tests := []vmTestCase{
		{
			`7000 * 7000`,
			"49000000",
			object.NUMBER_OBJ,
		},
	}

	runVmTests(t, tests, false, true)
}

func TestSpeedDecimalMultiplication(t *testing.T) {
	tests := []vmTestCase{
		{
			`7.123456789 * 7.123456789`,
			"50.743636624750190521",
			object.NUMBER_OBJ,
		},
	}

	runVmTests(t, tests, false, true)
}

func TestSpeedIntegerDivision(t *testing.T) {
	tests := []vmTestCase{
		{
			`150 \ 7`,
			"21",
			object.NUMBER_OBJ,
		},
	}

	runVmTests(t, tests, false, true)
}

func TestSpeedIntegerComparison(t *testing.T) {
	tests := []vmTestCase{
		{
			`for i = 1; i < 1000; i += 1 {
				10000 == 10000
			 }`,
			nil,
			object.NULL_OBJ,
		},
	}

	runVmTests(t, tests, false, true)
}

func TestSpeedDecimalComparison(t *testing.T) {
	tests := []vmTestCase{
		{
			`for i = 1; i < 1000; i += 1 {
				1.412456789789456 == 1.412456789789456
			 }`,
			nil,
			object.NULL_OBJ,
		},
	}

	runVmTests(t, tests, false, true)
}

func TestSpeedForLoop(t *testing.T) {
	tests := []vmTestCase{
		{
			`for i = 1; i < 10000; i += 1 {
			 }`,
			nil,
			object.NULL_OBJ,
		},
	}

	runVmTests(t, tests, false, true)
}

func TestSpeedListAppend(t *testing.T) {
	tests := []vmTestCase{
		{
			`len for[=[]] i = 1; i < 100; i += 1 {
				_for = more(_for, i)
			 }`,
			99,
			object.NUMBER_OBJ,
		},
		{
			`len for i = 1; i < 100; i += 1 {
				_for = _for ~ [i]
			 }`,
			99,
			object.NUMBER_OBJ,
		},
		{
			`len for i = 1; i < 100; i += 1 {
				_for ~= [i]
			 }`,
			99,
			object.NUMBER_OBJ,
		},
	}

	runVmTests(t, tests, false, true)
}

func TestSpeedHashAppend(t *testing.T) {
	tests := []vmTestCase{
		{
			`len for[=[]] i = 1; i < 100; i += 1 {
				_for = more(_for, {i: string(i)})
			 }`,
			99,
			object.NUMBER_OBJ,
		},
		{
			`len for i = 1; i < 100; i += 1 {
				_for = _for ~ {i: string(i)}
			 }`,
			99,
			object.NUMBER_OBJ,
		},
		{
			`len for i = 1; i < 100; i += 1 {
				_for ~= {i: string(i)}
			 }`,
			99,
			object.NUMBER_OBJ,
		},
	}

	runVmTests(t, tests, false, true)
}

func TestSpeedStringAppend(t *testing.T) {
	tests := []vmTestCase{
		{
			`len for[=""] i = 1; i < 100; i += 1 {
				_for = more(_for, string(i))
			 }`,
			189,
			object.NUMBER_OBJ,
		},
		{
			`len for[=""] i = 1; i < 100; i += 1 {
				_for = _for ~ string(i)
			 }`,
			189,
			object.NUMBER_OBJ,
		},
		{
			`len for[=""] i = 1; i < 100; i += 1 {
				_for ~= string(i)
			 }`,
			189,
			object.NUMBER_OBJ,
		},
	}

	runVmTests(t, tests, false, true)
}

func TestSpeedStringNumberAppend(t *testing.T) {
	tests := []vmTestCase{
		{
			`len for[=""] i = 65; i < 98; i += 1 {
				_for = more(_for, i)
			 }`,
			33,
			object.NUMBER_OBJ,
		},
		{
			`len for[=""] i = 65; i < 98; i += 1 {
				_for = _for ~ i
			 }`,
			33,
			object.NUMBER_OBJ,
		},
		{
			`len for[=""] i = 65; i < 98; i += 1 {
				_for ~= i
			 }`,
			33,
			object.NUMBER_OBJ,
		},
	}

	runVmTests(t, tests, false, true)
}

func TestSpeedAllocation(t *testing.T) {
	tests := []vmTestCase{
		{
			`len for of 100 {
				_for = []
			 }`,
			0,
			object.NUMBER_OBJ,
		},
	}

	runVmTests(t, tests, false, true)
}

func TestSpeedIndexListByNumber(t *testing.T) {
	tests := []vmTestCase{
		{
			`val x = [0] * 1000
			 for of 10000 { x[700] }`,
			nil,
			object.NULL_OBJ,
		},
	}

	runVmTests(t, tests, false, true)
}

func TestSpeedIndexStringByNumber(t *testing.T) {
	tests := []vmTestCase{
		{
			`val x = cp2s('A' .. 'Z') * 10
			 for of 10000 { x[70] }`,
			nil,
			object.NULL_OBJ,
		},
	}

	runVmTests(t, tests, false, true)
}

func TestSpeedGoStringVsNumberEqualityComparisons(t *testing.T) {
	// curiosty about efficiency...
	iterations := 100000

	var start, total int64

	// equal strings comparison time
	start = time.Now().UnixNano()
	for i := 0; i < iterations; i++ {
		_ = "abcdyoyo" == "abcdyoyo"
	}
	total = time.Now().UnixNano() - start
	fmt.Printf("Go equal strings comparison time: %d\n", total)

	// inequal strings comparison time
	start = time.Now().UnixNano()
	for i := 0; i < iterations; i++ {
		_ = "abcdyoyo" == "abcdyoyo1"
	}
	total = time.Now().UnixNano() - start
	fmt.Printf("Go inequal strings comparison time: %d\n", total)

	// equal numbers comparison time
	start = time.Now().UnixNano()
	for i := 0; i < iterations; i++ {
		_ = 456789456123 == 456789456123
	}
	total = time.Now().UnixNano() - start
	fmt.Printf("Go equal numbers comparison time: %d\n", total)

	// inequal numbers comparison time
	start = time.Now().UnixNano()
	for i := 0; i < iterations; i++ {
		_ = 456789456123 == 4567894561239
	}
	total = time.Now().UnixNano() - start
	fmt.Printf("Go inequal numbers comparison time: %d\n", total)
}
