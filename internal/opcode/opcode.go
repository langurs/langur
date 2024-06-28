// langur/opcode/opcode.go

package opcode

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

func bug(fnName, s string) {
	panic(s)
}

const OperandWidth_Jump = 4
const OperandWidth_ShortCircuitJump = 2
const OperandWidth_Constant = 2
const OperandWidth_Code = 1

var OP_JUMP_LEN = len(Make(OpJump, 0))
var OP_JUMP_RELAY_LEN = len(Make(OpJumpRelay, 0, 0))

const (
	// 8 bit flags (0x01, 0x02, 0x04, 0x08, ...)
	OC_Database_Op = 1 << iota
	OC_Combination_Op
)

const (
	OC_Regex_None = iota
	OC_Regex_Re2
)

const (
	OC_PlaceHolder_Break = iota
	OC_PlaceHolder_Next
	OC_PlaceHolder_Fallthrough
	OC_PlaceHolder_IfElse_Exit
	OC_PlaceHolder_IfElse_TestFailed
)

type Instructions []byte

func (ins Instructions) Copy() Instructions {
	newIns := make(Instructions, len(ins))
	copy(newIns, ins)
	return newIns
}

type OpCode = byte

const (
	NoOp OpCode = iota
	OpPop
	OpIs
	OpIn
	OpOf

	OpConstant
	OpClosure
	OpExecute

	OpMode

	OpRange
	OpList
	OpHash
	OpIndex

	OpJumpIfNotTruthy
	OpJump
	OpJumpBack
	OpJumpPlaceHolder
	OpJumpRelay
	OpJumpRelayIfNotTruthy

	OpTryCatch
	OpThrow

	OpCall
	OpCallWithExpansion
	OpReturnValue

	OpSetGlobal
	OpSetLocal
	OpSetNonLocal
	OpSetGlobalIndexedValue
	OpSetLocalIndexedValue
	OpSetNonLocalIndexedValue

	OpGetGlobal
	OpGetLocal
	OpGetNonLocal
	OpGetFree
	OpGetSelf

	OpTrue
	OpFalse
	OpNull

	OpAppend
	OpString
	OpRegex
	OpDateTime
	OpDuration
	OpFormat

	OpAdd
	OpSubtract
	OpMultiply
	OpDivide
	OpTruncateDivide
	OpFloorDivide
	OpRemainder
	OpModulus
	OpPower
	OpRoot

	OpEqual
	OpNotEqual
	OpGreaterThan
	OpGreaterThanOrEqual
	OpLessThan
	OpLessThanOrEqual

	OpForward

	OpDivisibleBy
	OpNotDivisibleBy

	OpNumericNegation
	OpLogicalNegation

	OpLogicalAnd
	OpLogicalOr
	OpLogicalNAnd
	OpLogicalNOr
	OpLogicalXor  // logical non-equivalence
	OpLogicalNXor // logical equivalence
)

type Definition struct {
	Name          string
	OperandWidths []int
}

func DisplayName(op OpCode, dbComp bool) string {
	def, err := Lookup(op)
	if err == nil {
		if dbComp {
			return def.Name + "?"
		}
		return def.Name
	}
	if dbComp {
		return "op?"
	}
	return "op"
}

var definitions = map[OpCode]*Definition{
	OpPop: {Name: "Pop"},

	// operands: object type number or 0
	OpIs: {Name: "Is", OperandWidths: []int{1}},
	OpIn: {Name: "In"},
	OpOf: {Name: "Of"},

	OpConstant: {Name: "Constant", OperandWidths: []int{OperandWidth_Constant}},
	OpExecute:  {Name: "Execute", OperandWidths: []int{OperandWidth_Constant}},

	// operands: constant, freecount
	OpClosure: {Name: "Closure", OperandWidths: []int{OperandWidth_Constant, 1}},

	OpMode: {Name: "Mode", OperandWidths: []int{1}},

	OpRange: {Name: "Range"},
	OpList:  {Name: "List", OperandWidths: []int{2}},
	OpHash:  {Name: "Hash", OperandWidths: []int{2}},
	OpIndex: {Name: "Index", OperandWidths: []int{OperandWidth_ShortCircuitJump}},

	OpJumpIfNotTruthy: {Name: "JumpIfNotTruthy", OperandWidths: []int{OperandWidth_Jump}},
	OpJump:            {Name: "Jump", OperandWidths: []int{OperandWidth_Jump}},
	OpJumpBack:        {Name: "JumpBack", OperandWidths: []int{OperandWidth_Jump}},

	// OpJumpRelay used to signal break, next, and fallthrough to another frame
	// OpJumpRelayIfNotTruthy used within scoped test of if/else expression
	// operands: jump, level
	OpJumpRelay:            {Name: "JumpRelay", OperandWidths: []int{3, 1}},
	OpJumpRelayIfNotTruthy: {Name: "JumpRelayIfNotTruthy", OperandWidths: []int{3, 1}},
	OpJumpPlaceHolder:      {Name: "JumpPlaceHolder", OperandWidths: []int{OperandWidth_Jump}}, // operand: code

	// operands: tryframe, catchframe, elseframe
	OpTryCatch: {Name: "TryCatch", OperandWidths: []int{OperandWidth_Constant, OperandWidth_Constant, OperandWidth_Constant}},
	OpThrow:    {Name: "Throw"},

	OpCall:              {Name: "Call", OperandWidths: []int{1}},
	OpCallWithExpansion: {Name: "CallWithExpansion", OperandWidths: []int{1}},
	OpReturnValue:       {Name: "ReturnValue"},

	OpSetGlobal:             {Name: "SetGlobal", OperandWidths: []int{2}},
	OpGetGlobal:             {Name: "GetGlobal", OperandWidths: []int{2}},
	OpSetGlobalIndexedValue: {Name: "SetGlobalIndexedValue", OperandWidths: []int{2}},

	OpSetLocal:             {Name: "SetLocal", OperandWidths: []int{1}},
	OpGetLocal:             {Name: "GetLocal", OperandWidths: []int{1}},
	OpSetLocalIndexedValue: {Name: "SetLocalIndexedValue", OperandWidths: []int{1}},

	// operands: index, level
	OpSetNonLocal:             {Name: "SetNonLocal", OperandWidths: []int{1, 1}},
	OpGetNonLocal:             {Name: "GetNonLocal", OperandWidths: []int{1, 1}},
	OpSetNonLocalIndexedValue: {Name: "SetNonLocalIndexedValue", OperandWidths: []int{1, 1}},

	OpGetFree: {Name: "GetFree", OperandWidths: []int{1}},
	OpGetSelf: {Name: "GetSelf"},

	OpTrue:  {Name: "True"},
	OpFalse: {Name: "False"},
	OpNull:  {Name: "Null"},

	OpAppend:   {Name: "Append", OperandWidths: []int{1}},
	OpString:   {Name: "String", OperandWidths: []int{2}},
	OpRegex:    {Name: "Regex", OperandWidths: []int{1}},
	OpDateTime: {Name: "DateTime"},
	OpDuration: {Name: "Duration"},
	OpFormat:   {Name: "Format", OperandWidths: []int{1}},

	OpAdd:            {Name: "Add"},
	OpSubtract:       {Name: "Subtract"},
	OpMultiply:       {Name: "Multiply"},
	OpDivide:         {Name: "Divide"},
	OpTruncateDivide: {Name: "TruncateDivide"},
	OpFloorDivide:    {Name: "FloorDivide"},
	OpRemainder:      {Name: "Remainder"},
	OpModulus:        {Name: "Modulus"},
	OpPower:          {Name: "Power"},
	OpRoot:           {Name: "Root"},

	OpEqual:              {Name: "Equal", OperandWidths: []int{OperandWidth_Code, OperandWidth_ShortCircuitJump}},
	OpNotEqual:           {Name: "NotEqual", OperandWidths: []int{OperandWidth_Code, OperandWidth_ShortCircuitJump}},
	OpGreaterThan:        {Name: "GreaterThan", OperandWidths: []int{OperandWidth_Code, OperandWidth_ShortCircuitJump}},
	OpGreaterThanOrEqual: {Name: "GreaterThanOrEqual", OperandWidths: []int{OperandWidth_Code, OperandWidth_ShortCircuitJump}},
	OpLessThan:           {Name: "LessThan", OperandWidths: []int{OperandWidth_Code, OperandWidth_ShortCircuitJump}},
	OpLessThanOrEqual:    {Name: "LessThanOrEqual", OperandWidths: []int{OperandWidth_Code, OperandWidth_ShortCircuitJump}},
	// have both less than and greater than opcodes b/c of null-propagating short-circuiting comparisons

	OpForward: {Name: "Forward"},

	OpDivisibleBy:    {Name: "DivisibleBy", OperandWidths: []int{OperandWidth_Code, OperandWidth_ShortCircuitJump}},
	OpNotDivisibleBy: {Name: "NotDivisibleBy", OperandWidths: []int{OperandWidth_Code, OperandWidth_ShortCircuitJump}},

	OpLogicalNegation: {Name: "LogicalNegation", OperandWidths: []int{OperandWidth_Code}},
	OpNumericNegation: {Name: "NumericNegation"},

	OpLogicalAnd:  {Name: "LogicalAnd", OperandWidths: []int{OperandWidth_Code, OperandWidth_ShortCircuitJump}},
	OpLogicalOr:   {Name: "LogicalOr", OperandWidths: []int{OperandWidth_Code, OperandWidth_ShortCircuitJump}},
	OpLogicalNAnd: {Name: "LogicalNAnd", OperandWidths: []int{OperandWidth_Code, OperandWidth_ShortCircuitJump}},
	OpLogicalNOr:  {Name: "LogicalNOr", OperandWidths: []int{OperandWidth_Code, OperandWidth_ShortCircuitJump}},
	OpLogicalXor:  {Name: "LogicalXor", OperandWidths: []int{OperandWidth_Code, OperandWidth_ShortCircuitJump}},
	OpLogicalNXor: {Name: "LogicalNXor", OperandWidths: []int{OperandWidth_Code, OperandWidth_ShortCircuitJump}},
}

func Lookup(op OpCode) (*Definition, error) {
	def, ok := definitions[op]
	if !ok {
		return nil, fmt.Errorf("OpCode %d undefined", op)
	}
	return def, nil
}

func isIndexedGetSetOpCode(op OpCode) bool {
	switch op {
	case OpGetGlobal, OpGetLocal, OpGetNonLocal,
		OpSetGlobal, OpSetLocal, OpSetNonLocal,
		OpSetGlobalIndexedValue, OpSetLocalIndexedValue, OpSetNonLocalIndexedValue:

		return true
	}
	return false
}

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
				if isIndexedGetSetOpCode(op) && i == 0 {
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
				if isIndexedGetSetOpCode(op) && i == 0 {
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
