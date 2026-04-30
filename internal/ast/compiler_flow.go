// langur/ast/compiler_flow.go

package ast

import (
	"fmt"
	"langur/object"
	"langur/opcode"
)

func init() {
	// an assertion to be sure the compiler hasn't broken here
	if opcode.OP_JUMP_LEN != opcode.OP_JUMP_RELAY_LEN {
		bug("init", "opcode.OP_JUMP_LEN != opcode.OP_JUMP_RELAY_LEN: This means the old method of setting break/next jumps will not work (compiler must change).")
	}
}

func (c *Compiler) checkStatementCounts() (err error) {
	if c.breakStmtCount != 0 {
		err = c.makeErr(nil, fmt.Sprintf(`%d "break" statement(s) not accounted for`, c.breakStmtCount))
	} else if c.nextStmtCount != 0 {
		err = c.makeErr(nil, fmt.Sprintf(`%d "next" statement(s) not accounted for`, c.nextStmtCount))
	} else if c.fallthroughStmtCount != 0 {
		err = c.makeErr(nil, fmt.Sprintf(`%d "fallthrough" statement(s) not accounted for`, c.fallthroughStmtCount))
	}
	return
}

// next, break, and fallthrough compiled to OpJumpPlaceHolder
// now must find them and fix them as OpJump or if in another frame, as OpJumpRelay
func (c *Compiler) fixJumps(
	ins opcode.Instructions,
	conditionalJumps bool,
	lookForOperandCode int, decrementCnt *int,
	jumpLocal, jumpNonLocal, frameLevel int) opcode.Instructions {

	opJumpLocal := opcode.OpJump
	opJumpNonLocal := opcode.OpJumpRelay
	if conditionalJumps {
		opJumpLocal = opcode.OpJumpIfNotTruthy
		opJumpNonLocal = opcode.OpJumpRelayIfNotTruthy
	}

	// currently lacking an intermediate rep. (IR) compilation phase, ...
	// ... we convert the bytecode into slices first
	insSlc := opcode.OpCodesAndOperandsSliceOfInstructionSlices(ins)
	newIns := opcode.Instructions{}

	for _, piece := range insSlc {
		if frameLevel == 0 {
			jumpLocal -= len(piece)
		}
		newPiece := piece // unless we find out otherwise

		switch piece[0] {
		case opcode.OpJumpPlaceHolder:
			operand := int(opcode.ReadUInt32(piece[1:]))

			if operand == lookForOperandCode {
				if frameLevel == 0 {
					newPiece = opcode.Make(opJumpLocal, jumpLocal)

				} else {
					if jumpNonLocal == 0 {
						newPiece = opcode.Make(opJumpNonLocal, jumpLocal, frameLevel)
					} else {
						newPiece = opcode.Make(opJumpNonLocal, jumpNonLocal, frameLevel)
					}
				}
				if decrementCnt != nil {
					(*decrementCnt)--
				}
			}
			// else...
			// maybe another placeholder type ... pass it through...

		case opcode.OpExecute:
			index := opcode.ReadUInt16(piece[1:])

			// OpExecute not used for a function, so we know the CompiledCode is not a function...
			c.constants[index].(*object.CompiledCode).InsPackage.Instructions = c.fixJumps(
				c.constants[index].(*object.CompiledCode).InsPackage.Instructions, conditionalJumps, lookForOperandCode, decrementCnt, jumpLocal, jumpNonLocal, frameLevel+1)

		case opcode.OpTryCatch:
			tryIndex := opcode.ReadUInt16(piece[1:])
			catchIndex := opcode.ReadUInt16(piece[3:])
			elseIndex := opcode.ReadUInt16(piece[5:])

			c.constants[tryIndex].(*object.CompiledCode).InsPackage.Instructions = c.fixJumps(
				c.constants[tryIndex].(*object.CompiledCode).InsPackage.Instructions, conditionalJumps, lookForOperandCode, decrementCnt, jumpLocal, jumpNonLocal, frameLevel+1)

			c.constants[catchIndex].(*object.CompiledCode).InsPackage.Instructions = c.fixJumps(
				c.constants[catchIndex].(*object.CompiledCode).InsPackage.Instructions, conditionalJumps, lookForOperandCode, decrementCnt, jumpLocal, jumpNonLocal, frameLevel+1)

			if elseIndex != 0 {
				c.constants[elseIndex].(*object.CompiledCode).InsPackage.Instructions = c.fixJumps(
					c.constants[elseIndex].(*object.CompiledCode).InsPackage.Instructions, conditionalJumps, lookForOperandCode, decrementCnt, jumpLocal, jumpNonLocal, frameLevel+1)
			}
		}

		newIns = append(newIns, newPiece...)
	}

	return newIns
}
