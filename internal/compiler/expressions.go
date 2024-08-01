// langur/compiler/expressions.go

package compiler

import (
	"fmt"
	"langur/ast"
	"langur/object"
	"langur/opcode"
	"langur/token"
)

func (c *Compiler) compileIndexNode(node *ast.IndexNode) (ins opcode.Instructions, err error) {
	var b []byte

	// Get "left" node
	b, err = c.compileNode(node.Left, true)
	if err != nil {
		return
	}
	ins = append(ins, b...)

	// Get the index
	b, err = c.compileNode(node.Index, true)
	if err != nil {
		return
	}
	ins = append(ins, b...)

	if node.Alternate == nil {
		ins = append(ins, opcode.Make(opcode.OpIndex, 0)...)

	} else {
		// alternate for an invalid index
		var alt opcode.Instructions
		alt, err = c.compileNode(node.Alternate, true)
		if err != nil {
			return
		}
		ins = append(ins, opcode.Make(opcode.OpIndex, len(alt))...)
		ins = append(ins, alt...)
	}

	return
}

func (c *Compiler) wrapInstructions(ins opcode.Instructions) int {
	// NOTE: Call this before c.popVariableScope().
	compiled := &object.CompiledCode{
		InsPackage:         opcode.InsPackage{Instructions: ins},
		LocalBindingsCount: c.symbolTable.DefinitionCount,
	}
	return c.addConstant(compiled)
}
func (c *Compiler) wrapInstructionsWithExecute(ins opcode.Instructions) opcode.Instructions {
	// NOTE: Call this before c.popVariableScope().
	index := c.wrapInstructions(ins)
	return opcode.Make(opcode.OpExecute, index)
}

func (c *Compiler) compileBlock(node *ast.BlockNode, noValueIfEmpty bool) (
	ins opcode.Instructions, err error) {

	var bslc []byte

	if node.HasScope {
		// only wrap expressions containing declarations (as an efficiency improvement)
		if ast.NodeContainsFirstScopeLevelDeclaration(node) {
			defer func() {
				ins = c.wrapInstructionsWithExecute(ins)
				c.popVariableScope()
			}()
			c.pushVariableScope()
		}
	}

	if noValueIfEmpty && len(node.Statements) == 0 {
		ins = c.noValueIns

	} else {
		for i, s := range node.Statements {
			if i < len(node.Statements)-1 {
				bslc, err = c.compileNode(s, true)
			} else {
				// last node in Block; not to pop on last node of Block
				bslc, err = c.compileNode(s, false)
			}
			ins = append(ins, bslc...)

			if err != nil {
				return
			}
		}
	}
	return
}

func (c *Compiler) compilePrefixExpression(node *ast.PrefixExpressionNode) (ins opcode.Instructions, err error) {
	var b []byte
	b, err = c.compileNode(node.Right, true)
	if err != nil {
		return
	}

	code, _, _ := ocCodeFromAstCode(node.Operator.Code)

	switch node.Operator.Type {
	case token.NOT:
		ins = append(b, opcode.Make(opcode.OpLogicalNegation, code)...)
	case token.MINUS:
		ins = append(b, opcode.Make(opcode.OpNumericNegation)...)
	default:
		err = c.makeErr(node, fmt.Sprintf("Unknown prefix operator %s", token.TypeDescription(node.Operator.Type)))
	}

	return
}

func (c *Compiler) compileInfixExpression(node *ast.InfixExpressionNode) (ins opcode.Instructions, err error) {
	var left, right []byte

	code, isDatabaseOperation, _ := ocCodeFromAstCode(node.Operator.Code)

	// NOTE: negated in present form, ...
	// may not work and play well with database operation but so far not mixed
	op, negated, ok := InfixTokenToOpCode(node.Operator)
	if !ok {
		err = c.makeErr(node, fmt.Sprintf("no infix token to opcode conversion for %s", token.TypeDescription(node.Operator.Type)))
		return
	}

	left, err = c.compileNode(node.Left, true)
	if err != nil {
		return
	}

	rightTypeCode := ast.NodeToTypeCode(node.Right)
	rightIsType := rightTypeCode != 0

	if !rightIsType || node.Operator.Type != token.IS {
		right, err = c.compileNode(node.Right, true)
		if err != nil {
			return
		}
	}

	plain := func() (ins opcode.Instructions, err error) {
		ins = append(left, right...)
		ins = append(ins, opcode.Make(op)...)
		if negated {
			ins = append(ins, opcode.Make(opcode.OpLogicalNegation, 0)...)
		}
		return ins, nil
	}

	plainWithCode := func() (ins opcode.Instructions, err error) {
		ins = append(left, right...)
		ins = append(ins, opcode.Make(op, code)...)
		if negated {
			ins = append(ins, opcode.Make(opcode.OpLogicalNegation, 0)...)
		}
		return ins, nil
	}

	nonShortCircuiting := func() (ins opcode.Instructions, err error) {
		ins = append(left, right...)
		ins = append(ins, opcode.Make(op, code, 0)...)
		if negated {
			ins = append(ins, opcode.Make(opcode.OpLogicalNegation, 0)...)
		}
		return ins, nil
	}

	shortCircuiting := func() (ins opcode.Instructions, err error) {
		evalWithRight := opcode.Make(op, code, 0)

		// len(right)+len(evalWithRight) == opcodes to jump if left gives the answer
		ins = append(left, opcode.Make(op, code, len(right)+len(evalWithRight))...)

		// if we didn't short-circuit, must evaluate here...
		ins = append(ins, right...)
		ins = append(ins, evalWithRight...)

		if negated {
			ins = append(ins, opcode.Make(opcode.OpLogicalNegation, 0)...)
		}

		return ins, nil
	}

	either := func() (ins opcode.Instructions, err error) {
		// either: for operations that could have short-circuiting
		// but only when used as "database" (null propagating) operators
		if isDatabaseOperation {
			return shortCircuiting()
		}
		return nonShortCircuiting()
	}

	withTypeCode := func() (ins opcode.Instructions, err error) {
		tcode := 0 // 0 indicates requirement for right operand
		ins = left

		if rightIsType {
			tcode = int(rightTypeCode)
		} else {
			ins = append(ins, right...)
		}

		ins = append(ins, opcode.Make(op, tcode)...)

		if negated {
			ins = append(ins, opcode.Make(opcode.OpLogicalNegation, 0)...)
		}

		return ins, nil
	}

	switch node.Operator.Type {
	case token.APPEND:
		return plainWithCode()

	case token.IS:
		return withTypeCode()

	case token.RANGE,
		token.PLUS, token.MINUS,
		token.ASTERISK, token.SLASH,
		token.BACKSLASH, token.DOUBLESLASH,
		token.REMAINDER, token.MODULUS,
		token.POWER, token.ROOT,
		token.FORWARD,
		token.IN, token.OF:

		return plain()

	case token.AND, token.OR,
		token.NAND, token.NOR:

		return shortCircuiting()

	case token.EQUAL, token.NOT_EQUAL,
		token.GREATER_THAN, token.GT_OR_EQUAL,
		token.LESS_THAN, token.LT_OR_EQUAL,
		token.DIVISIBLE_BY, token.NOT_DIVISIBLE_BY,
		token.XOR, token.NXOR:

		return either()

	default:
		err = c.makeErr(node, fmt.Sprintf("unknown operator (%s)", token.TypeDescription(node.Operator.Type)))
	}

	return
}

func InfixTokenToOpCode(tok token.Token) (op opcode.OpCode, negated, ok bool) {
	ok = true
	negated = token.NegatedLiteral(tok.Literal)

	switch tok.Type {
	case token.APPEND:
		op = opcode.OpAppend
	case token.RANGE:
		op = opcode.OpRange
	case token.PLUS:
		op = opcode.OpAdd
	case token.MINUS:
		op = opcode.OpSubtract
	case token.ASTERISK:
		op = opcode.OpMultiply
	case token.SLASH:
		op = opcode.OpDivide
	case token.BACKSLASH:
		op = opcode.OpTruncateDivide
	case token.DOUBLESLASH:
		op = opcode.OpFloorDivide
	case token.REMAINDER:
		op = opcode.OpRemainder
	case token.MODULUS:
		op = opcode.OpModulus
	case token.POWER:
		op = opcode.OpPower
	case token.ROOT:
		op = opcode.OpRoot

	case token.EQUAL:
		op = opcode.OpEqual
	case token.NOT_EQUAL:
		op = opcode.OpNotEqual
	case token.GREATER_THAN:
		op = opcode.OpGreaterThan
	case token.GT_OR_EQUAL:
		op = opcode.OpGreaterThanOrEqual
	case token.LESS_THAN:
		op = opcode.OpLessThan
	case token.LT_OR_EQUAL:
		op = opcode.OpLessThanOrEqual

	case token.FORWARD:
		op = opcode.OpForward

	case token.DIVISIBLE_BY:
		op = opcode.OpDivisibleBy
	case token.NOT_DIVISIBLE_BY:
		op = opcode.OpNotDivisibleBy

	case token.AND:
		op = opcode.OpLogicalAnd
	case token.OR:
		op = opcode.OpLogicalOr

	case token.NAND:
		op = opcode.OpLogicalNAnd
	case token.NOR:
		op = opcode.OpLogicalNOr

	case token.XOR:
		op = opcode.OpLogicalXor
	case token.NXOR:
		op = opcode.OpLogicalNXor

	case token.IS:
		op = opcode.OpIs
	case token.IN:
		op = opcode.OpIn
	case token.OF:
		op = opcode.OpOf

	default:
		ok = false
	}

	return
}

func ocCodeFromAstCode(code int) (c int, isDataBaseOp, isComboOp bool) {
	if 0 != code&token.CODE_DB_OPERATOR {
		c = opcode.OC_Database_Op
		isDataBaseOp = true
	}
	if 0 != code&token.CODE_COMBINATION_ASSIGNMENT_OPERATOR {
		c |= opcode.OC_Combination_Op
		isComboOp = true
	}
	return
}
