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
		Instructions:       ins,
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

	code := ocCodeFromAstCode(node.Operator.Code)

	switch node.Operator.Type {
	case token.NOT:
		ins = append(b, opcode.Make(opcode.OpLogicalNegation, code)...)
	case token.MINUS:
		ins = append(b, opcode.Make(opcode.OpNumericNegation)...)
	default:
		err = makeErr(node, fmt.Sprintf("Unknown prefix operator %s", token.TypeDescription(node.Operator.Type)))
	}

	return
}

func ocCodeFromAstCode(code int) int {
	c := 0
	if 0 != code&token.CODE_DB_OPERATOR {
		c = opcode.OC_Database_Op
	}
	if 0 != code&token.CODE_COMBINATION_ASSIGNMENT_OPERATOR {
		c |= opcode.OC_Combination_Op
	}
	return c
}

func (c *Compiler) compileInfixExpression(node *ast.InfixExpressionNode) (ins opcode.Instructions, err error) {
	var left, right []byte

	code := ocCodeFromAstCode(node.Operator.Code)
	isDatabaseOperation := 0 != node.Operator.Code&token.CODE_DB_OPERATOR

	// NOTE: isNegatedOperation: in present form, may not work and play well with database operation but so far not mixed
	isNegatedOperation := token.NegatedLiteral(node.Operator.Literal)

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

	plain := func(op opcode.OpCode) (ins opcode.Instructions, err error) {
		ins = append(left, right...)
		ins = append(ins, opcode.Make(op)...)
		if isNegatedOperation {
			ins = append(ins, opcode.Make(opcode.OpLogicalNegation, 0)...)
		}
		return ins, nil
	}

	plainWithCode := func(op opcode.OpCode) (ins opcode.Instructions, err error) {
		ins = append(left, right...)
		ins = append(ins, opcode.Make(op, code)...)
		if isNegatedOperation {
			ins = append(ins, opcode.Make(opcode.OpLogicalNegation, 0)...)
		}
		return ins, nil
	}

	nonShortCircuiting := func(op opcode.OpCode) (ins opcode.Instructions, err error) {
		ins = append(left, right...)
		ins = append(ins, opcode.Make(op, code, 0)...)
		if isNegatedOperation {
			ins = append(ins, opcode.Make(opcode.OpLogicalNegation, 0)...)
		}
		return ins, nil
	}

	shortCircuiting := func(op opcode.OpCode) (ins opcode.Instructions, err error) {
		evalWithRight := opcode.Make(op, code, 0)

		// len(right)+len(evalWithRight) == opcodes to jump if left gives the answer
		ins = append(left, opcode.Make(op, code, len(right)+len(evalWithRight))...)

		// if we didn't short-circuit, must evaluate here...
		ins = append(ins, right...)
		ins = append(ins, evalWithRight...)

		if isNegatedOperation {
			ins = append(ins, opcode.Make(opcode.OpLogicalNegation, 0)...)
		}

		return ins, nil
	}

	either := func(op opcode.OpCode) (ins opcode.Instructions, err error) {
		// either: for operations that could have short-circuiting
		// but only when used as "database" (null propagating) operators
		if isDatabaseOperation {
			return shortCircuiting(op)
		}
		return nonShortCircuiting(op)
	}

	withTypeCode := func(op opcode.OpCode) (ins opcode.Instructions, err error) {
		tcode := 0 // 0 indicates requirement for right operand
		ins = left

		if rightIsType {
			tcode = int(rightTypeCode)
		} else {
			ins = append(ins, right...)
		}

		ins = append(ins, opcode.Make(op, tcode)...)

		if isNegatedOperation {
			ins = append(ins, opcode.Make(opcode.OpLogicalNegation, 0)...)
		}

		return ins, nil
	}

	switch node.Operator.Type {
	case token.APPEND:
		return plainWithCode(opcode.OpAppend)
	case token.RANGE:
		return plain(opcode.OpRange)
	case token.PLUS:
		return plain(opcode.OpAdd)
	case token.MINUS:
		return plain(opcode.OpSubtract)
	case token.ASTERISK:
		return plain(opcode.OpMultiply)
	case token.SLASH:
		return plain(opcode.OpDivide)
	case token.BACKSLASH:
		return plain(opcode.OpTruncateDivide)
	case token.DOUBLESLASH:
		return plain(opcode.OpFloorDivide)
	case token.REMAINDER:
		return plain(opcode.OpRemainder)
	case token.MODULUS:
		return plain(opcode.OpModulus)
	case token.POWER:
		return plain(opcode.OpPower)
	case token.ROOT:
		return plain(opcode.OpRoot)

	case token.EQUAL:
		return either(opcode.OpEqual)
	case token.NOT_EQUAL:
		return either(opcode.OpNotEqual)
	case token.GREATER_THAN:
		return either(opcode.OpGreaterThan)
	case token.GT_OR_EQUAL:
		return either(opcode.OpGreaterThanOrEqual)
	case token.LESS_THAN:
		return either(opcode.OpLessThan)
	case token.LT_OR_EQUAL:
		return either(opcode.OpLessThanOrEqual)

	case token.FORWARD:
		return plain(opcode.OpForward)

	case token.DIVISIBLE_BY:
		return either(opcode.OpDivisibleBy)
	case token.NOT_DIVISIBLE_BY:
		return either(opcode.OpNotDivisibleBy)

	case token.AND:
		return shortCircuiting(opcode.OpLogicalAnd)
	case token.OR:
		return shortCircuiting(opcode.OpLogicalOr)

	case token.NAND:
		return shortCircuiting(opcode.OpLogicalNAnd)
	case token.NOR:
		return shortCircuiting(opcode.OpLogicalNOr)

	case token.XOR:
		return either(opcode.OpLogicalXor)
	case token.NXOR:
		return either(opcode.OpLogicalNXor)

	case token.IS:
		return withTypeCode(opcode.OpIs)
	case token.IN:
		return plain(opcode.OpIn)
	case token.OF:
		return plain(opcode.OpOf)

	default:
		err = makeErr(node, fmt.Sprintf("unknown operator (%s)", token.TypeDescription(node.Operator.Type)))
	}

	return
}
