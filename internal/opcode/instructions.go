// langur/opcode/instructions.go

package opcode

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"langur/trace"
	// "langur/token"
)

type Instructions []byte

func (ins Instructions) Copy() Instructions {
	newIns := make(Instructions, len(ins))
	copy(newIns, ins)
	return newIns
}

// TODO: use InsPackage instead of just Instructions, allowing meta-data to be included
type InsPackage struct {
	Instructions Instructions
	Where        trace.WhereSlice
}

func (ip InsPackage) Copy() InsPackage {
	return InsPackage{
		Instructions: ip.Instructions.Copy(),
		Where:        trace.CopyWhereSlice(ip.Where),
	}
}

func (ip InsPackage) Append(ip2 InsPackage) InsPackage {
	ins := append(ip.Instructions, ip2.Instructions...)
	where := trace.AppendWhereSlice(ip.Where, ip2.Where)
	return InsPackage{Instructions: ins, Where: where}
}

// TODO: TEST
// func MakePkg(op OpCode, tok token.Token, operands ...int) (pkg InsPackage) {
// 	pkg = InsPackage{}
// 	ins := Make(op, operands...)

// 	pkg.Where = make(trace.WhereSlice, len(ins))
// 	pkg.Where[0] = &(tok.Where)

// 	return pkg
// }

func Make(op OpCode, operands ...int) (ins Instructions) {
	// added make with error test as of 0.5.4, for things that are not a bug but a limitation, such as too many local variables
	var err error
	ins, err = MakeWithErrTest(op, operands...)
	if err != nil {
		bug("Make", err.Error())
	}
	return
}

func MakeWithErrTest(op OpCode, operands ...int) (ins Instructions, err error) {
	const fnName = "MakeWithErrTest"

	def, defined := definitions[op]
	if !defined {
		//return []byte{}
		bug(fnName, fmt.Sprintf("OpCode %d not defined", op))
	}

	if op == OpFormat {
		// variable width
	} else if len(operands) != len(def.OperandWidths) {
		bug(fnName, fmt.Sprintf("Operand Count Mismatch on OpCode %s, expected=%d, received=%d", def.Name, len(def.OperandWidths), len(operands)))
	}

	instuctionLen := 1
	for _, w := range def.OperandWidths {
		instuctionLen += w
	}

	instruction := make(Instructions, instuctionLen)
	instruction[0] = byte(op)

	const max1byteop = 255              // 2 ^ 8 - 1
	const max2byteop = 65535            // 2 ^ 16 - 1
	const max3byteop = 16777215         // 2 ^ 24 - 1
	const max4byteop int64 = 4294967295 // 2 ^ 32 - 1
	// Using int64 here doesn't fix everything for 32-bit systems, but allows us to run/test these things on them.

	offset := 1
	for i, o := range operands {
		w := def.OperandWidths[i]
		switch w {
		case 1:
			if o > max1byteop {
				if IsIndexedGetSetOpCode(op) && i == 0 {
					err = fmt.Errorf("%s index out of range (more than %d local variables)", DisplayName(op, false), max2byteop)
					return
				}
			}
			if o < 0 || o > max1byteop {
				bug(fnName, fmt.Sprintf("Operand %d on OpCode %s value (%d) out of range", i+1, def.Name, o))
			}
			instruction[offset] = uint8(o)

		case 2:
			if o > max2byteop {
				if IsIndexedGetSetOpCode(op) && i == 0 {
					err = fmt.Errorf("%s index out of range (more than %d global variables)", DisplayName(op, false), max2byteop)
					return
				}
			}
			if o < 0 || o > max2byteop {
				bug(fnName, fmt.Sprintf("Operand %d on OpCode %s value (%d) out of range", i+1, def.Name, o))
			}
			binary.BigEndian.PutUint16(instruction[offset:], uint16(o))

		case 3:
			if o < 0 || o > max3byteop {
				bug(fnName, fmt.Sprintf("Operand %d on OpCode %s value (%d) out of range", i+1, def.Name, o))
			}
			temp := []byte{0, 0, 0, 0}
			binary.BigEndian.PutUint32(temp, uint32(o))
			copy(instruction[offset:], temp[1:])

		case 4:
			if o < 0 || int64(o) > max4byteop {
				bug(fnName, fmt.Sprintf("Operand %d on OpCode %s value (%d) out of range", i+1, def.Name, o))
			}
			binary.BigEndian.PutUint32(instruction[offset:], uint32(o))

		default:
			bug(fnName, fmt.Sprintf("Operand %d on OpCode %s of unknown width %d", i+1, def.Name, w))
		}
		offset += w
	}

	return instruction, nil
}

func (ins Instructions) String() string {
	var out bytes.Buffer

	i := 0
	for i < len(ins) {
		def, err := Lookup(ins[i])
		if err != nil {
			bug("Instructions.String", err.Error())
			fmt.Fprintf(&out, "ERROR: %s\n", err)
			continue
		}

		operands, offset := ReadOperands(def, ins[i+1:])

		fmt.Fprintf(&out, "%04d %s\n", i, ins.FmtInstruction(def, operands))

		i += 1 + offset
	}

	return out.String()
}

func (ins Instructions) FmtInstruction(def *Definition, operands []int) string {
	operandCount := len(def.OperandWidths)

	if len(operands) != operandCount {
		bug("Instructions.FmtInstruction", fmt.Sprintf("Operand length %d does not match defined %d", len(operands), operandCount))
		return fmt.Sprintf("ERROR: operand length %d does not match defined %d\n", len(operands), operandCount)
	}

	var out bytes.Buffer

	out.WriteString(def.Name)
	for _, o := range operands {
		out.WriteString(fmt.Sprintf(" %d", o))
	}

	return out.String()
}

func ReadOperands(def *Definition, ins Instructions) ([]int, int) {
	operands := make([]int, len(def.OperandWidths))
	offset := 0

	for i, width := range def.OperandWidths {
		switch width {
		case 1:
			operands[i] = int(ins[offset])
		case 2:
			operands[i] = int(ReadUInt16(ins[offset:]))
		case 3:
			operands[i] = int(ReadUInt24(ins[offset:]))
		case 4:
			operands[i] = int(ReadUInt32(ins[offset:]))
		default:
			bug("ReadOperands", fmt.Sprintf("Operand width %d not accounted for", width))
		}

		offset += width
	}

	return operands, offset
}

func ReadUInt16(ins Instructions) uint16 {
	return binary.BigEndian.Uint16(ins)
}

func ReadUInt24(ins Instructions) uint32 {
	return binary.BigEndian.Uint32(append([]byte{0}, ins[:3]...))
}

func ReadUInt32(ins Instructions) uint32 {
	return binary.BigEndian.Uint32(ins)
}

func OpCodesAndOperandsSliceOfInstructionSlices(ins Instructions) []Instructions {
	iSlc := []Instructions{}

	i := 0
	for i < len(ins) {
		def, err := Lookup(ins[i])
		if err != nil {
			bug("OpCodesAndOperandsSliceOfInstructionSlices", err.Error())
			continue
		}
		operands, offset := ReadOperands(def, ins[i+1:])

		insPiece, err := MakeWithErrTest(ins[i], operands...)
		if err != nil {
			bug("OpCodesAndOperandsSliceOfInstructionSlices", err.Error())
		}
		iSlc = append(iSlc, insPiece)

		i += 1 + offset
	}

	return iSlc
}
