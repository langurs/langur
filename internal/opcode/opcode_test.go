// langur/opcode/opcode_test.go

package opcode

import (
	"langur/token"
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
		{OpFunction, []int{65534, 255, 4}, []byte{byte(OpFunction), 255, 254, 255, 4}},
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

// TODO: test with trace information
func TestMakePkg(t *testing.T) {
	tests := []struct {
		op        OpCode
		operands  []int
		expected InsPackage
	}{
		{OpConstant, []int{65534}, InsPackage{Instructions: []byte{byte(OpConstant), 255, 254}}},
		{OpAdd, []int{}, InsPackage{Instructions: []byte{byte(OpAdd)}}},
		{OpFunction, []int{65534, 255, 4}, InsPackage{Instructions: []byte{byte(OpFunction), 255, 254, 255, 4}}},
	}

	tok := token.Token{}

	for _, tt := range tests {
		pkg, err := MakePkgWithErrTest(tok, tt.op, tt.operands...)
		if err != nil {
			t.Fatalf(err.Error())
		}

		if len(pkg.Instructions) != len(tt.expected.Instructions) {
			t.Fatalf("Instruction has wrong length, expected=%d, received=%d",
				len(tt.expected.Instructions), len(pkg.Instructions))
		}

		for i, b := range tt.expected.Instructions {
			if pkg.Instructions[i] != tt.expected.Instructions[i] {
				t.Errorf("Wrong byte at pos %d, expected=%d, received=%d",
					i, b, pkg.Instructions[i])
			}
		}
	}
}

// TODO: test with trace information
func TestAppendPkg(t *testing.T) {
	pkg1 := InsPackage{Instructions: []byte{byte(OpConstant), 255, 254}}
	pkg2 := InsPackage{Instructions: []byte{byte(OpConstant), 250, 250, byte(OpAdd)}}
	expect := InsPackage{Instructions: []byte{byte(OpConstant), 255, 254, byte(OpConstant), 250, 250, byte(OpAdd)}}

	pkg3 := pkg1.Append(pkg2)

	if len(pkg3.Instructions) != len(expect.Instructions) {
		t.Fatalf("Instruction has wrong length, expected=%d, received=%d",
			len(expect.Instructions), len(pkg3.Instructions))
	}

	for i, b := range expect.Instructions {
		if pkg3.Instructions[i] != expect.Instructions[i] {
			t.Errorf("Wrong byte at pos %d, expected=%d, received=%d",
				i, b, pkg3.Instructions[i])
		}
	}
}

func TestInstructionsString(t *testing.T) {
	instructions := []Instructions{
		Make(OpAdd),
		Make(OpConstant, 2),
		Make(OpConstant, 65535),
		Make(OpFunction, 65535, 255, 255),
		Make(OpJumpRelay, 16777215, 7),
	}

	expected := `0000 Add
0001 Constant 2
0004 Constant 65535
0007 Function 65535 255 255
0012 JumpRelay 16777215 7
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
		{OpFunction, []int{65535, 255, 255}, 4},

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
