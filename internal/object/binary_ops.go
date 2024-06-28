// langur/object/binary_ops.go
// "binary" as in 2 operands, not as in bitwise operations

package object

import (
	"fmt"
	"langur/opcode"
)

// Normal comparison considers true that null == null.
// Database comparison ==? with null on either side returns null.
func isDatabaseOperation(code int) bool {
	return 0 != code&opcode.OC_Database_Op
}
func isCombinationOperation(code int) bool {
	return 0 != code&opcode.OC_Combination_Op
}

func invalidOpString(op opcode.OpCode, dbComp bool, left, right Object) string {
	if right == nil {
		return fmt.Sprintf("Invalid operation: %s %s",
			opcode.DisplayName(op, dbComp), left.TypeString())
	}
	return fmt.Sprintf("Invalid operation: %s %s %s",
		left.TypeString(), opcode.DisplayName(op, dbComp), right.TypeString())
}

func invalidOpError(op opcode.OpCode, dbComp bool, left, right Object) error {
	return fmt.Errorf(invalidOpString(op, dbComp, left, right))
}

func ShortCircuitingOperation(op opcode.OpCode, left Object, code int) (
	result Object, haveResult bool) {

	// left only; haven't evaluated the right yet
	if isDatabaseOperation(code) {
		// database comparison; either side null, return null
		if left == NULL {
			return NULL, true
		}

	} else {
		// not a database comparison
		switch op {
		case opcode.OpLogicalAnd:
			if !left.IsTruthy() {
				return FALSE, true
			}
		case opcode.OpLogicalNAnd:
			if !left.IsTruthy() {
				return TRUE, true
			}
		case opcode.OpLogicalOr:
			if left.IsTruthy() {
				return TRUE, true
			}
		case opcode.OpLogicalNOr:
			if left.IsTruthy() {
				return FALSE, true
			}
		}
	}

	// We don't know the answer yet.
	return NONE, false
}

func BinaryLogicalOperation(op opcode.OpCode, left, right Object, code int) (Object, error) {
	dbComp := isDatabaseOperation(code)
	if dbComp {
		// database comparison; either side null, return null
		if left == NULL || right == NULL {
			return NULL, nil
		}
	}

	var b bool
	var err error

	switch op {
	case opcode.OpLogicalAnd:
		b = left.IsTruthy() && right.IsTruthy()
	case opcode.OpLogicalOr:
		b = left.IsTruthy() || right.IsTruthy()
	case opcode.OpLogicalXor:
		b = left.IsTruthy() != right.IsTruthy()

	case opcode.OpLogicalNAnd:
		b = !(left.IsTruthy() && right.IsTruthy())
	case opcode.OpLogicalNOr:
		b = !(left.IsTruthy() || right.IsTruthy())
	case opcode.OpLogicalNXor:
		b = left.IsTruthy() == right.IsTruthy()

	default:
		err = invalidOpError(op, dbComp, left, right)
	}

	return NativeBoolToObject(b), err
}

func BinaryOperation(op opcode.OpCode, left, right Object, code int) (result Object, err error) {
	// not a "logical" operation ...
	dbComp := false

	defer func() {
		if panik := recover(); panik != nil {
			err = PanicToError(panik)
			if left.Type() == NUMBER_OBJ && right.Type() == NUMBER_OBJ {
				err = AsMathError(err, "")
			}
		}
	}()

	switch op {
	case opcode.OpAdd:
		left, ok := left.(IAdd)
		if ok {
			result = left.Add(right)
		}

	case opcode.OpSubtract:
		left, ok := left.(ISubtract)
		if ok {
			result = left.Subtract(right)
		}

	case opcode.OpMultiply:
		left, ok := left.(IMultiply)
		if ok {
			result = left.Multiply(right)
		}

	case opcode.OpDivide:
		left, ok := left.(IDivide)
		if ok {
			result = left.Divide(right)
		}

	case opcode.OpTruncateDivide:
		left, ok := left.(IDivideTruncate)
		if ok {
			result = left.DivideTruncate(right)
		}

	case opcode.OpFloorDivide:
		left, ok := left.(IDivideFloor)
		if ok {
			result = left.DivideFloor(right)
		}

	case opcode.OpRemainder:
		left, ok := left.(IRemainder)
		if ok {
			result = left.Remainder(right)
		}

	case opcode.OpModulus:
		left, ok := left.(IModulus)
		if ok {
			result = left.Modulus(right)
		}

	case opcode.OpPower:
		left, ok := left.(IPower)
		if ok {
			result = left.Power(right)
		}

	case opcode.OpRoot:
		left, ok := left.(IRoot)
		if ok {
			result = left.Root(right)
		}

	case opcode.OpAppend:
		if left == NONE {
			if right == NONE {
				result = right

			} else {
				right, ok := right.(IAppendToNone)
				if ok {
					result = right.AppendToNone()
				}
			}

		} else {
			left, ok := left.(IAppend)
			if ok {
				result = left.Append(right)
			}
		}

	case opcode.OpRange:
		result = NewRange(left, right)

	case opcode.OpForward:
		right, ok := right.(IForward)
		if ok {
			result = right.Forward(left)
		}
	}

	switch result.(type) {
	case nil:
		return nil, invalidOpError(op, dbComp, left, right)

	case error:
		return nil, result.(error)
	}

	return result, nil
}

func AsMathError(err error, source string) error {
	// already checked that err != nil
	msg := err.Error()
	return NewError(ERR_MATH, source, msg)
}

func BinaryComparison(
	op opcode.OpCode, left, right Object,
	code int) (
	Object, error) {

	if left == NULL || right == NULL {
		if isDatabaseOperation(code) {
			// database comparison; either side null, return null
			return NULL, nil
		}
	}

	var match bool
	comparable := true // unless we find out otherwise

	switch op {
	case opcode.OpEqual:
		match = Equal(left, right)
	case opcode.OpNotEqual:
		match = !Equal(left, right)

	case opcode.OpGreaterThan:
		match, comparable = GreaterThan(left, right)
	case opcode.OpGreaterThanOrEqual:
		match, comparable = GreaterOrEqual(left, right)

	case opcode.OpLessThan:
		match, comparable = GreaterThan(right, left)
	case opcode.OpLessThanOrEqual:
		match, comparable = GreaterOrEqual(right, left)

	case opcode.OpDivisibleBy:
		match, comparable = DivisibleBy(left, right)

	case opcode.OpNotDivisibleBy:
		match, comparable = DivisibleBy(left, right)
		match = !match

	case opcode.OpIn:
		match, comparable = Contains(left, right)

	case opcode.OpOf:
		var err error
		var result Object
		result, err = Index(right, left)
		comparable = result != nil
		match = err == nil

	default:
		comparable = false
	}

	if !comparable {
		return nil, fmt.Errorf("Invalid comparison operation: %s %s %s",
			left.TypeString(),
			opcode.DisplayName(op, isDatabaseOperation(code)),
			right.TypeString())
	}

	return NativeBoolToObject(match), nil
}

func Equal(left, right Object) bool {
	if left == right {
		return true
		// This takes care of Booleans and null, which always use the same pre-defined object constants.
		// ... It may also take care of some other cases.

	} else {
		return left.Equal(right)
	}
}

func GreaterThan(l, r Object) (yes, comparable bool) {
	left, ok := l.(IGreaterThan)
	if !ok {
		return false, false
	}
	return left.GreaterThan(r)
}

func GreaterOrEqual(l, r Object) (yes, comparable bool) {
	left, ok := l.(IGreaterThan)
	if !ok {
		// even if they could be "equal" we return false for greater or equal ...
		// ... if not comparable as to whether one is greater
		return false, false
	}
	yes, comparable = left.GreaterThan(r)
	if yes {
		return yes, comparable
	}
	return l.Equal(r), comparable
}

func DivisibleBy(l, r Object) (yes, comparable bool) {
	left, ok := l.(IDivisibleBy)
	if !ok {
		return false, false
	}
	return left.DivisibleBy(r)
}

func Contains(l, r Object) (yes, comparable bool) {
	right, ok := r.(IContains)
	if !ok {
		return false, false
	}
	return right.Contains(l)
}

// result as nil indicates invalid operation; not to use alternate value
// result as left for most errors
func Index(left, index Object) (result Object, err error) {
	switch left := left.(type) {
	case IIndex:
		return left.Index(index, false)
	case *Null:
		return left, fmt.Errorf("Index on null without alternate")
	default:
		return nil, fmt.Errorf("Index operation not supported on %s", left.TypeString())
	}
}

func SetIndex(left, index, setTo Object) (Object, error) {
	switch left := left.(type) {
	case IIndex:
		return left.SetIndex(index, setTo)
	default:
		return nil, fmt.Errorf("Cannot set index value of type %s", left.TypeString())
	}
}

func LogicalNegation(operand Object, code int) (Object, error) {
	if operand == NULL && isDatabaseOperation(code) {
		// for DB, Negation of null results in null.
		return NULL, nil
	}
	return NativeBoolToObject(!operand.IsTruthy()), nil
}

func NumericNegation(operand Object) (Object, error) {
	switch n := operand.(type) {
	case *Number:
		return n.Negate(), nil
	default:
		return nil, fmt.Errorf("Unsupported type for numeric negation: %s",
			operand.TypeString())
	}
}
