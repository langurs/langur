// langur/ast/compiler_flow.go

package ast

import (
	"fmt"
	"langur/object"
	"langur/opcode"
	"langur/symbol"
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

func (c *Compiler) compileIfNode(node *IfNode) (pkg opcode.InsPackage, err error) {
	if node.TestsAndActions[len(node.TestsAndActions)-1].Test != nil {
		// no else/default section; add implicit else/default section of null or throw
		var def Node = NoValue
		if node.DefaultElse != nil {
			def = node.DefaultElse
		}
		node.TestsAndActions = append(node.TestsAndActions,
			TestDo{Test: nil, Do: &BlockNode{Statements: []Node{def}}})
	}

	type compiled struct {
		pkg opcode.InsPackage
		st  *symbol.SymbolTable
	}
	compiledTests := make([]compiled, len(node.TestsAndActions))
	compiledActions := make([]compiled, len(node.TestsAndActions))
	compiledTA := make([]compiled, len(node.TestsAndActions))

	/*
		opCodes (except for the great complication of scope frames)...
		test
		jump to next test if not truthy
		action
		jump to end
		...
		(rinse, repeat)
	*/

	/*
		Each section of if/else gets it's own scope. This compiles each to one of the following.
		1. no scope wrapping (no declarations in test or action)
		2. wrap scope over test and action (declarations in test and possibly in action)
		3. wrap scope over action only (declarations in action, but none in test)
	*/

	jumpToEndOpCodeLen := func(i int) int {
		// not adding a jump to the end if we already end with a fallthrough, break, or next
		if EndsWithDefiniteJump(node.TestsAndActions[i].Do.(*BlockNode).Statements) {
			return 0
		} else {
			return opcode.OP_JUMP_LEN
		}
	}

	// Compile tests first.
	for i, ta := range node.TestsAndActions {
		lastOne := i == len(node.TestsAndActions)-1

		compiledTests[i].st = nil

		if ta.Test == nil {
			if !lastOne {
				// Houston, we have a bug.
				if node.IsSwitchExpr {
					err = c.makeErr(node, "Default not last part of switch expression")
				} else {
					err = c.makeErr(node, "Else not last part of if/else expression")
				}
				bug("compileIfExpression", err.Error())
				return
			}

		} else {
			if NodeContainsFirstScopeLevelDeclaration(ta.Test) {
				// push and pop and save symbol table; wrap test/action together later
				c.pushVariableScope()
				compiledTests[i].pkg, err = ta.Test.Compile(c)
				compiledTests[i].st = c.symbolTable // save table for re-use
				c.popVariableScope()
			} else {
				// no scope on test
				compiledTests[i].pkg, err = ta.Test.Compile(c)
			}
			if err != nil {
				return
			}
		}
	}

	// Now compile the actions.
	for i, ta := range node.TestsAndActions {
		compiledActions[i].st = nil

		// push scope?
		if compiledTests[i].st != nil {
			// declarations in the test section and possibly in the action
			// using saved symbol table
			c.pushVariableScopeWithTable(compiledTests[i].st)
			compiledActions[i].st = c.symbolTable

		} else if NodeContainsFirstScopeLevelDeclaration(ta.Do) {
			// declarations in the action, but not the test section
			// using new symbol table
			c.pushVariableScope()
			compiledActions[i].st = c.symbolTable
		}

		compiledActions[i].pkg, err = ta.Do.Compile(c)
		if err != nil {
			return
		}

		if compiledActions[i].st != nil {
			if compiledTests[i].st == nil {
				// wrap only action into scope, not the test
				compiledActions[i].pkg = c.wrapInstructionsWithExecute(compiledActions[i].pkg, ta.Do.TokenInfo())
			}
			c.popVariableScope()
		}

		// set conditional jump over action
		if len(compiledTests[i].pkg.Instructions) > 0 {
			// not "else" or "default"
			if compiledTests[i].st == nil {
				compiledTests[i].pkg = compiledTests[i].pkg.Append(
					opcode.MakePkg(ta.Do.TokenInfo(), opcode.OpJumpIfNotTruthy, len(compiledActions[i].pkg.Instructions)+jumpToEndOpCodeLen(i)))

			} else {
				// going to have to add an OpJumpRelayIfNotTruthy b/c of test being buried in scope
				compiledTests[i].pkg = compiledTests[i].pkg.Append(
					opcode.MakePkg(ta.Do.TokenInfo(), opcode.OpJumpPlaceHolder, opcode.OC_PlaceHolder_IfElse_TestFailed))
			}
		}
	}

	// put it together
	for i := range node.TestsAndActions {
		lastOne := i == len(node.TestsAndActions)-1

		if node.IsSwitchExpr {
			// fix fallthrough on switch expressions only

			if compiledTests[i].st != nil {
				// If we allowed declarations within case statements, they would have to be included...
				// ...in scope wrapping, making it impossible to set a jump for fallthrough.
				err = c.makeErr(node, "Cannot use declarations in case statement of switch expression")
				return
			}

			// not looking for fallthrough in default section
			if !lastOne {
				compiledActions[i].pkg.Instructions = c.fixJumps(
					compiledActions[i].pkg.Instructions, false,
					opcode.OC_PlaceHolder_Fallthrough, &c.fallthroughStmtCount,

					len(compiledActions[i].pkg.Instructions)+ // over current action
						jumpToEndOpCodeLen(i)+ // over jump to end
						len(compiledTests[i+1].pkg.Instructions), // over next test

					0, 0)
			}
		}

		// put together test and action
		compiledTA[i].pkg = compiledTests[i].pkg.Append(compiledActions[i].pkg)
		compiledTA[i].st = compiledActions[i].st

		if compiledTests[i].st != nil {
			// wrap test and action into scope together
			c.pushVariableScopeWithTable(compiledTests[i].st)
			compiledTA[i].pkg = c.wrapInstructionsWithExecute(compiledTA[i].pkg, node.TestsAndActions[i].Test.TokenInfo())
			c.popVariableScope()

			compiledTA[i].pkg.Instructions = c.fixJumps(
				compiledTA[i].pkg.Instructions, true,
				opcode.OC_PlaceHolder_IfElse_TestFailed, nil,
				len(compiledTA[i].pkg.Instructions), jumpToEndOpCodeLen(i), 0)
		}

		// jump to end
		if !lastOne {
			if jumpToEndOpCodeLen(i) > 0 {
				compiledTA[i].pkg = compiledTA[i].pkg.Append(
					opcode.MakePkg(node.TestsAndActions[i].Test.TokenInfo(), opcode.OpJumpPlaceHolder, opcode.OC_PlaceHolder_IfElse_Exit))
			}
		}

		pkg = pkg.Append(compiledTA[i].pkg)
	}

	pkg.Instructions = c.fixJumps(pkg.Instructions, false, opcode.OC_PlaceHolder_IfElse_Exit, nil, len(pkg.Instructions), 0, 0)
	return
}
