// langur/opcode/opcode.go

package opcode

import (
	"fmt"
	"langur/token"
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
	OC_Index_inverse_Op
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

type OpCode = byte

const (
	NoOp OpCode = iota
	OpPop
	OpIs
	OpIn
	OpOf

	OpConstant
	OpFunction
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
	OpNameValue

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

	// operands: constant, freecount, variables by name count
	OpFunction: {Name: "Function", OperandWidths: []int{OperandWidth_Constant, 1, 1}},

	OpMode: {Name: "Mode", OperandWidths: []int{1}},

	OpRange: {Name: "Range"},
	OpList:  {Name: "List", OperandWidths: []int{2}},
	OpHash:  {Name: "Hash", OperandWidths: []int{2}},

	// operands: code, short-circuit jump
	OpIndex: {Name: "Index", OperandWidths: []int{1, OperandWidth_ShortCircuitJump}},

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

	// operands: positionalCount, bynameCount
	OpCall:              {Name: "Call", OperandWidths: []int{1, 1}},
	OpCallWithExpansion: {Name: "CallWithExpansion", OperandWidths: []int{1, 1}},

	OpReturnValue: {Name: "ReturnValue"},
	OpNameValue:   {Name: "NameValue"},

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

func IsIndexedGetSetOpCode(op OpCode) bool {
	switch op {
	case OpGetGlobal, OpGetLocal, OpGetNonLocal,
		OpSetGlobal, OpSetLocal, OpSetNonLocal,
		OpSetGlobalIndexedValue, OpSetLocalIndexedValue, OpSetNonLocalIndexedValue:

		return true
	}
	return false
}

func InfixTokenToOpCode(tok token.Token) (op OpCode, negated, ok bool) {
	ok = true
	negated = token.NegatedLiteral(tok.Literal)

	switch tok.Type {
	case token.APPEND:
		op = OpAppend
	case token.RANGE:
		op = OpRange
	case token.PLUS:
		op = OpAdd
	case token.MINUS:
		op = OpSubtract
	case token.ASTERISK:
		op = OpMultiply
	case token.SLASH:
		op = OpDivide
	case token.BACKSLASH:
		op = OpTruncateDivide
	case token.DOUBLESLASH:
		op = OpFloorDivide
	case token.REMAINDER:
		op = OpRemainder
	case token.MODULUS:
		op = OpModulus
	case token.POWER:
		op = OpPower
	case token.ROOT:
		op = OpRoot

	case token.EQUAL:
		op = OpEqual
	case token.NOT_EQUAL:
		op = OpNotEqual
	case token.GREATER_THAN:
		op = OpGreaterThan
	case token.GT_OR_EQUAL:
		op = OpGreaterThanOrEqual
	case token.LESS_THAN:
		op = OpLessThan
	case token.LT_OR_EQUAL:
		op = OpLessThanOrEqual

	case token.FORWARD:
		op = OpForward

	case token.DIVISIBLE_BY:
		op = OpDivisibleBy
	case token.NOT_DIVISIBLE_BY:
		op = OpNotDivisibleBy

	case token.AND:
		op = OpLogicalAnd
	case token.OR:
		op = OpLogicalOr

	case token.NAND:
		op = OpLogicalNAnd
	case token.NOR:
		op = OpLogicalNOr

	case token.XOR:
		op = OpLogicalXor
	case token.NXOR:
		op = OpLogicalNXor

	case token.IS:
		op = OpIs
	case token.IN:
		op = OpIn
	case token.OF:
		op = OpOf

	default:
		ok = false
	}

	return
}

func TokenCodeToOcCode(code int) (c int, isDataBaseOp, isComboOp bool) {
	if 0 != code&token.CODE_DB_OPERATOR {
		c = OC_Database_Op
		isDataBaseOp = true
	}
	if 0 != code&token.CODE_COMBINATION_ASSIGNMENT_OPERATOR {
		c |= OC_Combination_Op
		isComboOp = true
	}
	return
}
