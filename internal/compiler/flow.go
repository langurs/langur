// langur/compiler/flow.go

package compiler

import (
	"fmt"
	"langur/ast"
	"langur/object"
	"langur/opcode"
	"langur/symbol"
)

func init() {
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

func (c *Compiler) compileFor(node *ast.ForNode) (ins opcode.Instructions, err error) {
	var init, test, body, increment []byte

	c.pushVariableScope()
	defer func() {
		ins = c.wrapInstructionsWithExecute(ins)
		c.popVariableScope()
	}()

	// The 4 sections are...
	// 1. init
	// 2. test
	//	(conditionally jump out)
	// 3. body
	// 4. increment
	//	(jump back to test)
	// ... (as of 0.7+ ...)
	// 5. for loop value

	// Prior to 0.7, we would use init = c.noValueIns here to make sure something is on the stack.
	// Now, we set the for loop value a different way.

	for _, each := range node.Init {
		var i []byte
		i, err = c.compileNode(each, true)
		if err != nil {
			return
		}
		init = append(init, i...)
	}

	var loopValueInit opcode.Instructions
	loopValueInit, err = c.compileNode(node.LoopValueInit, true)
	if err != nil {
		return
	}
	init = append(init, loopValueInit...)

	loopValueVar := node.LoopValueInit.(*ast.ExpressionStatementNode).Expression.(*ast.LineDeclarationNode).Assignment.(*ast.AssignmentNode).Identifiers[0]
	// for setting break value when not specified as something else
	c.loopVarStack = append(c.loopVarStack, loopValueVar)
	defer func() {
		c.loopVarStack = c.loopVarStack[:len(c.loopVarStack)-1]
	}()

	if node.Test != nil {
		test, err = c.compileNode(node.Test, true)
		if err != nil {
			return
		}
	}

	body, err = c.compileNode(node.Body, true)
	if err != nil {
		return
	}

	for _, each := range node.Increment {
		var i []byte
		i, err = c.compileNode(each, true)
		if err != nil {
			return
		}
		increment = append(increment, i...)
	}

	body = c.fixJumps(body, false, opcode.OC_PlaceHolder_Next, &c.nextStmtCount, len(body), 0, 0)
	body = c.fixJumps(body, false, opcode.OC_PlaceHolder_Break, &c.breakStmtCount, len(body)+len(increment)+opcode.OP_JUMP_LEN, 0, 0)

	if len(test) > 0 {
		test = append(test, opcode.Make(opcode.OpJumpIfNotTruthy, len(body)+len(increment)+opcode.OP_JUMP_LEN)...)
	}
	// after increment, jump back to start of test section (or body if there is no test)
	increment = append(increment, opcode.Make(opcode.OpJumpBack, len(test)+len(body)+len(increment))...)

	ins = append(init, test...)
	ins = append(ins, body...)
	ins = append(ins, increment...)

	// append loop value to very end; vm will push onto stack before exiting frame
	// FIXME: ? do another way
	var loopValue opcode.Instructions
	loopValue, err = c.compileNode(c.loopVarStack[len(c.loopVarStack)-1], false)
	if err != nil {
		return
	}
	ins = append(ins, loopValue...)

	return
}

func (c *Compiler) compileBreak(node *ast.BreakNode) (ins opcode.Instructions, err error) {
	c.breakStmtCount++

	if len(c.loopVarStack) < 1 {
		return nil, c.makeErr(node, "Break declared outside of loop")
	}

	if node.Value == nil {
		// break with current for loop value
		ins, err = c.compileNode(c.loopVarStack[len(c.loopVarStack)-1], false)

	} else {
		// break with specified value
		// FIXME: redundancy when embedded in scope (no need to set variable)
		ins, err = c.compileNode(ast.MakeAssignmentExpression(c.loopVarStack[len(c.loopVarStack)-1], node.Value, false), false)
	}

	ins = append(ins, opcode.Make(opcode.OpJumpPlaceHolder, opcode.OC_PlaceHolder_Break)...)
	return
}

func (c *Compiler) compileNext(node *ast.NextNode) (ins opcode.Instructions, err error) {
	c.nextStmtCount++
	ins = opcode.Make(opcode.OpJumpPlaceHolder, opcode.OC_PlaceHolder_Next)

	// FIXME: redundancy for OpJumpRelay; don't need to push a value here
	ins = append(c.noValueIns, ins...)

	return
}

func (c *Compiler) compileFallthrough(node *ast.FallThroughNode) (ins opcode.Instructions, err error) {
	c.fallthroughStmtCount++
	ins = opcode.Make(opcode.OpJumpPlaceHolder, opcode.OC_PlaceHolder_Fallthrough)
	return
}

// if or switch expression: A switch node is translated by the parser into an if node, ....
// ... so the compiler never sees a switch node.
func (c *Compiler) compileIfExpression(node *ast.IfNode) (ins opcode.Instructions, err error) {

	if node.TestsAndActions[len(node.TestsAndActions)-1].Test != nil {
		// no else/default section; add implicit else/default section returning null
		node.TestsAndActions = append(node.TestsAndActions,
			ast.TestDo{Test: nil, Do: &ast.BlockNode{Statements: []ast.Node{ast.NoValue}}})
	}

	type compiled struct {
		ins opcode.Instructions
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
		if ast.EndsWithDefiniteJump(node.TestsAndActions[i].Do.(*ast.BlockNode).Statements) {
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
					bug("compileIfExpression", "Default not last part of switch expression")
					err = c.makeErr(node, "Default not last part of switch expression")
				} else {
					bug("compileIfExpression", "Else not last part of if/else expression")
					err = c.makeErr(node, "Else not last part of if/else expression")
				}
				return
			}

		} else {
			if ast.NodeContainsFirstScopeLevelDeclaration(ta.Test) {
				// push and pop and save symbol table; wrap test/action together later
				c.pushVariableScope()
				compiledTests[i].ins, err = c.compileNode(ta.Test, false)
				compiledTests[i].st = c.symbolTable // save table for re-use
				c.popVariableScope()
			} else {
				// no scope on test
				compiledTests[i].ins, err = c.compileNode(ta.Test, false)
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

		} else if ast.NodeContainsFirstScopeLevelDeclaration(ta.Do) {
			// declarations in the action, but not the test section
			// using new symbol table
			c.pushVariableScope()
			compiledActions[i].st = c.symbolTable
		}

		compiledActions[i].ins, err = c.compileBlock(ta.Do.(*ast.BlockNode), true)
		if err != nil {
			return
		}

		if compiledActions[i].st != nil {
			if compiledTests[i].st == nil {
				// wrap only action into scope, not the test
				compiledActions[i].ins = c.wrapInstructionsWithExecute(compiledActions[i].ins)
			}
			c.popVariableScope()
		}

		// set conditional jump over action
		if len(compiledTests[i].ins) > 0 {
			// not "else" or "default"
			if compiledTests[i].st == nil {
				compiledTests[i].ins = append(compiledTests[i].ins,
					opcode.Make(opcode.OpJumpIfNotTruthy, len(compiledActions[i].ins)+jumpToEndOpCodeLen(i))...)

			} else {
				// going to have to add an OpJumpRelayIfNotTruthy b/c of test being buried in scope
				compiledTests[i].ins = append(compiledTests[i].ins,
					opcode.Make(opcode.OpJumpPlaceHolder, opcode.OC_PlaceHolder_IfElse_TestFailed)...)
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
				compiledActions[i].ins = c.fixJumps(
					compiledActions[i].ins, false,
					opcode.OC_PlaceHolder_Fallthrough, &c.fallthroughStmtCount,

					len(compiledActions[i].ins)+ // over current action
						jumpToEndOpCodeLen(i)+ // over jump to end
						len(compiledTests[i+1].ins), // over next test

					0, 0)
			}
		}

		// put together test and action
		compiledTA[i].ins = append(compiledTests[i].ins, compiledActions[i].ins...)
		compiledTA[i].st = compiledActions[i].st

		if compiledTests[i].st != nil {
			// wrap test and action into scope together
			c.pushVariableScopeWithTable(compiledTests[i].st)
			compiledTA[i].ins = c.wrapInstructionsWithExecute(compiledTA[i].ins)
			c.popVariableScope()

			compiledTA[i].ins = c.fixJumps(
				compiledTA[i].ins, true,
				opcode.OC_PlaceHolder_IfElse_TestFailed, nil,
				len(compiledTA[i].ins), jumpToEndOpCodeLen(i), 0)
		}

		// jump to end
		if !lastOne {
			if jumpToEndOpCodeLen(i) > 0 {
				compiledTA[i].ins = append(compiledTA[i].ins,
					opcode.Make(opcode.OpJumpPlaceHolder, opcode.OC_PlaceHolder_IfElse_Exit)...)
			}
		}

		ins = append(ins, compiledTA[i].ins...)
	}

	ins = c.fixJumps(ins, false, opcode.OC_PlaceHolder_IfElse_Exit, nil, len(ins), 0, 0)
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

func (c *Compiler) compileTryCatch(node *ast.TryCatchNode) (ins opcode.Instructions, err error) {
	var try, catch, tcelse opcode.Instructions

	// The try frame doesn't have scope, but catch and else frames do.
	c.pushNonScope()
	try, err = c.compileNode(node.Try, false)
	c.popVariableScope()
	if err != nil {
		return
	}
	tryIndex := c.addConstant(&object.CompiledCode{InsPackage: opcode.InsPackage{Instructions: try}})

	// push scope for the catch frame, including the exception variable
	c.pushVariableScope()
	defer c.popVariableScope()

	var setException opcode.Instructions
	if node.ExceptionVar != nil {
		setException, err = c.compileNode(
			ast.MakeDeclarationAssignmentStatement(node.ExceptionVar, nil, true, false),
			true)

		if err != nil {
			return
		}
	}

	catch, err = c.compileNode(node.Catch, false)
	if err != nil {
		return
	}
	if node.ExceptionVar != nil {
		catch = append(setException, catch...)
	}
	catchIndex := c.wrapInstructions(catch)

	elseIndex := 0
	if node.Else != nil {
		// pop scope from catch; else with different scope
		c.popVariableScope()
		c.pushVariableScope()

		tcelse, err = c.compileNode(node.Else, false)
		elseIndex = c.wrapInstructions(tcelse)
		if elseIndex == 0 {
			bug("compileTryCatch", "elseIndex 0 (0 used as indicator for no else section)")
		}
	}

	ins = opcode.Make(opcode.OpTryCatch, tryIndex, catchIndex, elseIndex)
	return
}

func (c *Compiler) compileThrow(node *ast.ThrowNode) (ins opcode.Instructions, err error) {
	ins, err = c.compileNode(node.Exception, false)
	if err != nil {
		return
	}
	ins = append(ins, opcode.Make(opcode.OpThrow)...)
	return
}
