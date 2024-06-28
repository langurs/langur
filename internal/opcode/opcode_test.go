// langur/opcode/opcode_test.go

package opcode

import (
	"testing"
)

func TestMake(t *testing.T) {
	tests := []struct {
		op        OpCode
		operands  []int
		expeceted []byte
	}{
		{OpConstant, []int{65534}, []byte{byte(OpConstant), 255, 254}},
		{OpAdd, []int{}, []byte{byte(OpAdd)}},
		{OpClosure, []int{65534, 255}, []byte{byte(OpClosure), 255, 254, 255}},
	}

	for _, tt := range tests {
		instruction, err := MakeWithErrTest(tt.op, tt.operands...)
		if err != nil {
			t.Fatalf(err.Error())
		}

		if len(instruction) != len(tt.expeceted) {
			t.Fatalf("Instruction has wrong length, expected=%d, received=%d",
				len(tt.expeceted), len(instruction))
		}

		for i, b := range tt.expeceted {
			if instruction[i] != tt.expeceted[i] {
				t.Errorf("Wrong byte at pos %d, expected=%d, received=%d",
					i, b, instruction[i])
			}
		}
	}
}

func TestInstructionsString(t *testing.T) {
	instructions := []Instructions{
		Make(OpAdd),
		Make(OpConstant, 2),
		Make(OpConstant, 65535),
		Make(OpClosure, 65535, 255),
		Make(OpJumpRelay, 16777215, 7),
	}

	expected := `0000 Add
0001 Constant 2
0004 Constant 65535
0007 Closure 65535 255
0011 JumpRelay 16777215 7
`

	appended := Instructions{}
	for _, ins := range instructions {
		appended = append(appended, ins...)
	}

	if appended.String() != expected {
		t.Errorf("instructions String() not formatted as expected\nexpected=%q\nreceived=%q",
			expected, appended.String())
	}
}

func TestReadOperands(t *testing.T) {
	tests := []struct {
		op        OpCode
		operands  []int
		bytesRead int
	}{
		{OpJump, []int{2147483647}, 4}, // max value of signed 32-bit integer
		{OpJump, []int{4294967295}, 4}, // max value of unsigned 32-bit integer
		{OpJump, []int{77765535}, 4},
		{OpConstant, []int{65535}, 2},
		{OpClosure, []int{65535, 255}, 3},

		// {TEST, []int{9223372036854775807}, 8}, // max value of signed 64-bit integer
	}

	for _, tt := range tests {
		instruction, err := MakeWithErrTest(tt.op, tt.operands...)
		if err != nil {
			t.Fatalf(err.Error())
		}

		def, err := Lookup(tt.op)
		if err != nil {
			t.Fatalf("definition not found for %q\n", err)
		}

		operandsRead, n := ReadOperands(def, instruction[1:])
		if n != tt.bytesRead {
			t.Fatalf("n wrong, expected=%d, received=%d", tt.bytesRead, n)
		}

		for i, want := range tt.operands {
			if operandsRead[i] != want {
				t.Errorf("operand wrong, expected=%d, received=%d", want, operandsRead[i])
			}
		}
	}
}
