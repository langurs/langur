// langur/vm/process/runframe.go

package process

import (
	"fmt"
	"langur/object"
	"langur/opcode"
	"langur/trace"
)

// The fnReturn propogates langur return values out of frames and is handled by executeFunctionCall().
func (pr *Process) RunFrame(fr *frame, late []object.Object) (
	fnReturn object.Object,
	relay *jumpRelay,
	err error) {

	var errIP int
	var result object.Object
	retainLastValue := false

	if fr == nil {
		fr = pr.startFrame
	}

	// for repeated use
	ins := fr.code.InsPackage.Instructions

	// to reset the stack on exit
	// not used directly; using a Go slice
	sp := len(pr.stack)

	// late-binding assignments pushed onto stack last before executing frame ...
	// ... which should already contain the opcodes to retrieve the values
	pr.pushMultiple(late)

	defer func() {
		// catch panics and convert them to langur exceptions
		if pr.Modes.GoPanicToLangurException {
			if p := recover(); p != nil {
				name, _ := fr.getFnName()
				err = object.NewErrorFromAnything(p, "panic:"+name)
			}
		}

		if retainLastValue && fr != pr.startFrame {
			// ran out of instructions in this frame
			// not the global frame ...
			// not a return, exception, or jump relay
			// reset stack + add last value
			last := pr.look()
			pr.stack = pr.stack[:sp+1]
			pr.stack[sp] = last

		} else {
			// exiting global frame ...
			// or is a return, exception, or jump relay
			// reset the stack without adding to it
			pr.stack = pr.stack[:sp]
		}

		// to ensure to return an Error Object with Where field set...
		if err != nil {
			if e, isErrObj := err.(*object.Error); isErrObj {
				// only set Where field of Error Object if not already set
				// with 0 not being a valid line number (1-based)
				if e.Where.Line == 0 {
					e.Where = trace.FindLocation(fr.code.InsPackage.Where, errIP)
				}

			} else {
				// not an Error Object; create one so we can attach the location
				fnName, _ := fr.getFnName()
				err = object.NewErrorFromAnything(err, fnName)
				err.(*object.Error).Where = trace.FindLocation(fr.code.InsPackage.Where, errIP)
			}
		}
	}()

	for ip := 0; ip < len(ins); ip++ {
		errIP = ip	// in case of error, where did it happen?
		op := opcode.OpCode(ins[ip])

		switch op {
		case opcode.OpPop:
			pr.pop()

		case opcode.OpConstant:
			constIndex := opcode.ReadUInt16(ins[ip+1:])
			ip += 2
			err = pr.push(pr.constants[constIndex])

		case opcode.OpExecute:
			constIndex := int(opcode.ReadUInt16(ins[ip+1:]))
			ip += 2
			codeObj := pr.constants[constIndex].(*object.CompiledCode)
			fnReturn, relay, err = pr.runCompiledCode(codeObj, fr, nil, nil, nil)

		case opcode.OpFunction:
			constIndex := int(opcode.ReadUInt16(ins[ip+1:]))
			freeCount := int(ins[ip+3])
			// optionalsCount for optional default instructions (optional defaults that weren't pre-built)
			optionalsCount := int(ins[ip+4])
			ip += 4
			err = pr.pushFunction(constIndex, freeCount, optionalsCount)

		case opcode.OpSetGlobal:
			globalIndex := opcode.ReadUInt16(ins[ip+1:])
			ip += 2
			// look() doesn't pop, so that assignment is an expression
			pr.startFrame.locals[globalIndex] = pr.look()

		case opcode.OpGetGlobal:
			globalIndex := opcode.ReadUInt16(ins[ip+1:])
			ip += 2
			err = pr.push(pr.startFrame.locals[globalIndex])

		case opcode.OpSetLocal:
			localIndex := int(ins[ip+1])
			ip += 1
			// look() doesn't pop, so that assignment is an expression
			fr.setLocal(localIndex, pr.look())

		case opcode.OpGetLocal:
			localIndex := int(ins[ip+1])
			ip += 1
			result, err = fr.getLocal(localIndex)
			if err == nil {
				err = pr.push(result)
			}

		case opcode.OpSetNonLocal:
			index := int(ins[ip+1])
			level := int(ins[ip+2])
			ip += 2

			// look() doesn't pop, so that assignment is an expression
			fr.setNonLocal(index, level, pr.look())

		case opcode.OpGetNonLocal:
			index := int(ins[ip+1])
			level := int(ins[ip+2])
			ip += 2

			result, err = fr.getNonLocal(index, level)
			if err == nil {
				err = pr.push(result)
			}

		case opcode.OpSetDefine:
			// for setting indexed values
			// ... and for dot notation (future use)
			objIdx := pr.pop()
			target := pr.pop()
			// look() doesn't pop, so that assignment is an expression
			setTo := pr.look()
 
			err = setDefine(target, objIdx, setTo)

		case opcode.OpGetFree:
			freeIndex := int(ins[ip+1])
			ip += 1
			result, err = fr.getFree(freeIndex)
			if err == nil {
				err = pr.push(result)
			}

		case opcode.OpGetSelf:
			result, err = fr.getSelf()
			if err == nil {
				err = pr.push(result)
			}

		case opcode.OpJump:
			// ip + opcode.OP_JUMP_LEN + ... since it is a relative offset
			pos := ip + opcode.OP_JUMP_LEN + int(opcode.ReadUInt32(ins[ip+1:]))

			// - 1 since ip will be incremented by the loop
			ip = pos - 1

		case opcode.OpJumpBack:
			pos := ip - int(opcode.ReadUInt32(ins[ip+1:]))

			// - 1 since ip will be incremented by the loop
			ip = pos - 1

		case opcode.OpJumpIfNotTruthy:
			if pr.pop().IsTruthy() {
				// done; skip the operand
				ip += opcode.OperandWidth_Jump

			} else {
				pos := ip + opcode.OP_JUMP_LEN + int(opcode.ReadUInt32(ins[ip+1:]))

				// - 1 since ip will be incremented by the loop
				ip = pos - 1
			}

		case opcode.OpTrue:
			err = pr.push(object.TRUE)
		case opcode.OpFalse:
			err = pr.push(object.FALSE)
		case opcode.OpNull:
			err = pr.push(object.NULL)

		case opcode.OpIn, opcode.OpOf:
			code := int(ins[ip+1])
			ip += 1

			right := pr.pop()
			left := pr.pop()

			result, err = object.InfixComparison(op, left, right, code)
			if err == nil {
				err = pr.push(result)
			}

		case opcode.OpAdd, opcode.OpSubtract,
			opcode.OpMultiply, opcode.OpDivide,
			opcode.OpTruncateDivide, opcode.OpFloorDivide,
			opcode.OpRemainder, opcode.OpModulus,
			opcode.OpPower, opcode.OpRoot,

			opcode.OpRange:

			right := pr.pop()
			left := pr.pop()

			result, err = object.InfixNonLogicalOperation(op, left, right, 0)
			if err == nil {
				err = pr.push(result)
			}

		case opcode.OpAppend:
			code := int(ins[ip+1])
			ip += 1

			right := pr.pop()
			left := pr.pop()

			result, err = object.InfixNonLogicalOperation(op, left, right, code)
			if err == nil {
				err = pr.push(result)
			}

		case opcode.OpIs:
			code := int(ins[ip+1])
			tcode := int(ins[ip+2])
			ip += 2

			var is bool

			if tcode == 0 {
				// 0 indicates that we require a second operand; use object.Is() function
				right := pr.pop()
				left := pr.pop()
				is, err = object.Is(left, right)

			} else {
				// non-zero code indicates the type
				objTypeCode := int(pr.pop().Type())

				// using compiled fn code to check for either a built-in or compiled fn
				if objTypeCode == int(object.BUILTIN_FUNCTION_OBJ) {
					objTypeCode = int(object.COMPILED_CODE_OBJ)
				}

				is = objTypeCode == tcode
			}

			if err == nil {
				if 0 != code&opcode.OC_Negated_Op {
					is = !is
				}
				err = pr.push(object.NativeBoolToObject(is))
			}

		case opcode.OpLogicalAnd, opcode.OpLogicalOr,
			opcode.OpLogicalNAnd, opcode.OpLogicalNOr,
			opcode.OpLogicalNXor, opcode.OpLogicalXor:

			var left, right object.Object
			var ok bool

			// read in 2 operands
			code := int(ins[ip+1])
			ip += 1
			shortCircuitJump := int(opcode.ReadUInt16(ins[ip+1:]))
			ip += 2

			if shortCircuitJump == 0 {
				// not short-circuiting, or is second half
				right = pr.pop()
				left = pr.pop()

				result, err = object.InfixLogicalOperation(op, left, right, code)
				if err == nil {
					err = pr.push(result)
				}

			} else {
				// have left only; haven't evaluated right yet
				// just look; don't pop
				left = pr.look()
				result, ok = object.ShortCircuitingOperation(op, left, code)
				if ok {
					// short-circuit success
					// pop left now
					pr.pop()

					// jump over right evaluation
					ip += shortCircuitJump

					// push result
					err = pr.push(result)
				}
				// no result?
				// continues, starting evaluation of the right operand
			}

		case opcode.OpForward:
			right := pr.pop()
			left := pr.pop()

			var result object.Object
			if object.IsCallable(right) {
				// must be handled here, where we have access to the process
				// right operand as function to call; pass left as argument
				result, err = pr.callback(right, left)

			} else {
				result, err = object.InfixNonLogicalOperation(op, left, right, 0)
			}

			if err == nil {
				err = pr.push(result)
			}

		case opcode.OpEqual, opcode.OpNotEqual,
			opcode.OpGreaterThan, opcode.OpGreaterThanOrEqual,
			opcode.OpLessThan, opcode.OpLessThanOrEqual,
			opcode.OpDivisibleBy, opcode.OpNotDivisibleBy:

			var left, right object.Object
			var ok bool

			// read in 2 operands
			code := int(ins[ip+1])
			ip += 1
			// comparisons may have short-circuiting for null-propagating operators
			shortCircuitJump := int(opcode.ReadUInt16(ins[ip+1:]))
			ip += 2

			if shortCircuitJump == 0 {
				// not short-circuiting, or is second half
				right = pr.pop()
				left = pr.pop()

				result, err = object.InfixComparison(op, left, right, code)
				if err == nil {
					err = pr.push(result)
				}

			} else {
				// have left only; haven't evaluated right yet
				// just look; don't pop
				left = pr.look()
				result, ok = object.ShortCircuitingOperation(op, left, code)
				if ok {
					// short-circuit success
					// pop left now
					pr.pop()

					// jump over right evaluation
					ip += shortCircuitJump

					// push result
					err = pr.push(result)
				}
				// no result?
				// continues, starting evaluation of the right operand
			}

		case opcode.OpLogicalNegation:
			code := int(ins[ip+1])
			ip += 1

			result, err = object.LogicalNegation(pr.pop(), code)
			if err == nil {
				err = pr.push(result)
			}

		case opcode.OpNumericNegation:
			result, err = object.NumericNegation(pr.pop())
			if err == nil {
				err = pr.push(result)
			}

		case opcode.OpFormat:
			code := int(ins[ip+1])
			ip += 1

			result, err = pr.format(code)
			if err == nil {
				err = pr.push(result)
			}

		case opcode.OpString:
			elementCount := int(opcode.ReadUInt16(ins[ip+1:]))
			ip += 2
			err = pr.push(object.ToStringFromSlice(pr.popMultiple(elementCount)))

		case opcode.OpRegex:
			code := int(ins[ip+1])
			ip += 1

			obj := pr.pop()
			strObj, ok := obj.(*object.String)
			if ok {
				result, err = object.NewRegexByOpCode(strObj.String(), code)
				if err == nil {
					err = pr.push(result)
				}

			} else {
				err = fmt.Errorf("Expected string for regex pattern, received %s", obj.TypeString())
				bug("runFrame", err.Error())
			}

		case opcode.OpDateTime:
			code := int(ins[ip+1])
			ip += 1

			nowIncludesFractionalSeconds := 0 != code&opcode.OC_Fractional_Seconds
			
			obj := pr.pop()
			strObj, ok := obj.(*object.String)
			if ok {
				result, err = object.NewDateTimeFromLiteralString(strObj.String(), nowIncludesFractionalSeconds)
				if err == nil {
					err = pr.push(result)
				}

			} else {
				err = fmt.Errorf("Expected string for date-time pattern, received %s", obj.TypeString())
				bug("runFrame", err.Error())
			}

		case opcode.OpDuration:
			obj := pr.pop()
			strObj, ok := obj.(*object.String)
			if ok {
				result, err = object.NewDurationFromString(strObj.String())
				if err == nil {
					err = pr.push(result)
				}

			} else {
				err = fmt.Errorf("Expected string for duration pattern, received %s", obj.TypeString())
				bug("runFrame", err.Error())
			}

		case opcode.OpList:
			elementCount := int(opcode.ReadUInt16(ins[ip+1:]))
			ip += 2
			err = pr.push(&object.List{Elements: pr.popMultiple(elementCount)})

		case opcode.OpHash:
			elementCount := int(opcode.ReadUInt16(ins[ip+1:]))
			ip += 2

			result, err = object.NewHashFromSlice(pr.popMultiple(elementCount), false)
			if err == nil {
				err = pr.push(result)
			} else {
				err = object.NewError(object.ERR_INDEX, "", err.Error())
			}

		case opcode.OpIndex:
			shortCircuitJump := int(opcode.ReadUInt16(ins[ip+1:]))
			ip += 2

			index := pr.pop()
			left := pr.pop()

			jumpAlt := true

			result, err = object.Index(left, index)
			if err == nil {
				err = pr.push(result)

			} else {
				if result != nil && shortCircuitJump != 0 {
					// error indexing, but have an alternate; use it
					jumpAlt = false
					err = nil

				} else {
					// error indexing; no alternate
					err = object.NewError(object.ERR_INDEX, "", err.Error())
				}
			}

			if jumpAlt {
				// jump over alternate instructions
				ip += shortCircuitJump
			}

		case opcode.OpCall:
			positionalCount := int(ins[ip+1])
			bynameCount := int(ins[ip+2])
			ip += 2

			result, err = pr.executeFunctionCall(fr, positionalCount, bynameCount, false)
			if err == nil {
				err = pr.push(result)
			}

		case opcode.OpCallWithExpansion:
			positionalCount := int(ins[ip+1])
			bynameCount := int(ins[ip+2])
			ip += 2

			result, err = pr.executeFunctionCall(fr, positionalCount, bynameCount, true)
			if err == nil {
				err = pr.push(result)
			}

		case opcode.OpReturnValue:
			fnReturn = pr.pop()
			return

		case opcode.OpNameValue:
			value := pr.pop()
			name := pr.pop()

			result, err = object.NewNameValue(name, value)
			if err == nil {
				err = pr.push(result)
			}

		case opcode.OpJumpRelay:
			jump := int(opcode.ReadUInt24(ins[ip+1:]))
			level := int(ins[ip+4])

			relay = &jumpRelay{
				Jump:  jump,
				Level: level - 1,
				Value: pr.look(), // pass along, such as for break with value
			}
			return

		case opcode.OpJumpRelayIfNotTruthy:
			jump := int(opcode.ReadUInt24(ins[ip+1:]))
			level := int(ins[ip+4])

			if pr.pop().IsTruthy() {
				// done; skip operands
				ip += opcode.OperandWidth_Jump

			} else {
				relay = &jumpRelay{
					Jump:  jump,
					Level: level - 1,
				}
				return
			}

		case opcode.OpTryCatch:
			tryIndex := int(opcode.ReadUInt16(ins[ip+1:]))
			catchIndex := int(opcode.ReadUInt16(ins[ip+3:]))
			elseIndex := int(opcode.ReadUInt16(ins[ip+5:]))
			ip += 6

			fnReturn, relay, err = pr.executeTryCatch(fr, tryIndex, catchIndex, elseIndex)

		case opcode.OpThrow:
			err = pr.throw(fr, pr.pop())

		case opcode.OpMode:
			code := int(ins[ip+1])
			ip += 1
			setting := pr.pop()
			err = pr.setMode(code, setting)

		default:
			desc, err2 := opcode.Lookup(op)
			if err2 == nil {
				err = fmt.Errorf("OpCode %d (%s) not accounted for", op, desc.Name)
			} else {
				err = fmt.Errorf("Unknown OpCode %d", op)
			}

			bug("runFrame", err.Error())
		}

		if err != nil || fnReturn != nil {
			// fnReturn handled by executeFunctionCall(); just propagate from here
			return
		}

		if relay != nil {
			if relay.Level == 0 {
				ip += relay.Jump
				if relay.Value != nil {
					err = pr.push(relay.Value)
				}
				relay = nil
			} else {
				relay.Level--
				return
			}
		}

	} // instruction loop

	retainLastValue = true
	return
}
