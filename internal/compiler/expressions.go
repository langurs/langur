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
	b, err = c.compileNode(node.Left)
	if err != nil {
		return
	}
	ins = append(ins, b...)

	// Get the index
	b, err = c.compileNode(node.Index)
	if err != nil {
		return
	}
	ins = append(ins, b...)

	if node.Alternate == nil {
		ins = append(ins, opcode.Make(opcode.OpIndex, 0)...)

	} else {
		// alternate for an invalid index
		var alt opcode.Instructions
		alt, err = c.compileNode(node.Alternate)
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
				bslc, err = c.compileNodeWithPopIfExprStmt(s)
			} else {
				// last node in Block; not to pop on last node of Block
				bslc, err = c.compileNode(s)
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
	b, err = c.compileNode(node.Right)
	if err != nil {
		return
	}

	code, _, _ := opcode.TokenCodeToOcCode(node.Operator.Code)

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

// check for complex number such as 1 + 1i
func (c *Compiler) checkForComplexNumber(node *ast.InfixExpressionNode, op opcode.OpCode) (ins opcode.Instructions, err error) {
	numeric := func(node ast.Node) int {
		switch node := node.(type) {
		case *ast.NumberNode:
			if node.Imaginary {
				return 2
			}
			return 1
		}
		return 0
	}
	if (op == opcode.OpAdd || op == opcode.OpSubtract) && 
		numeric(node.Left) == 1 &&
		numeric(node.Right) == 2 {
		
		ins, err = c.compileComplexNumber(node.Left, node.Right, op == opcode.OpSubtract)
	}
	return
}

func (c *Compiler) compileInfixExpression(node *ast.InfixExpressionNode) (ins opcode.Instructions, err error) {
	var left, right []byte

	code, isDatabaseOperation, _ := opcode.TokenCodeToOcCode(node.Operator.Code)

	// NOTE: negated in present form, ...
	// may not work and play well with database operation but so far not mixed
	op, negated, ok := opcode.InfixTokenToOpCode(node.Operator)
	if !ok {
		err = c.makeErr(node, fmt.Sprintf("no infix token to opcode conversion for %s", token.TypeDescription(node.Operator.Type)))
		return
	}

	if !negated && node.Operator.Code == 0 {
		ins, err = c.checkForComplexNumber(node, op)
		if ins != nil || err != nil {
			return
		}
	}

	left, err = c.compileNode(node.Left)
	if err != nil {
		return
	}

	rightTypeCode := ast.NodeToLangurTypeCode(node.Right)
	rightIsType := rightTypeCode != 0

	if !rightIsType || node.Operator.Type != token.IS {
		right, err = c.compileNode(node.Right)
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

	switch op {
	case opcode.OpAppend:
		return plainWithCode()

	case opcode.OpIs:
		return withTypeCode()

	case opcode.OpRange,
		opcode.OpAdd, opcode.OpSubtract,
		opcode.OpMultiply, opcode.OpDivide,
		opcode.OpTruncateDivide, opcode.OpFloorDivide,
		opcode.OpRemainder, opcode.OpModulus,
		opcode.OpPower, opcode.OpRoot,
		opcode.OpForward,
		opcode.OpIn, opcode.OpOf:

		return plain()

	case opcode.OpLogicalAnd, opcode.OpLogicalNAnd,
		opcode.OpLogicalOr, opcode.OpLogicalNOr:

		return shortCircuiting()

	case opcode.OpEqual, opcode.OpNotEqual,
		opcode.OpGreaterThan, opcode.OpGreaterThanOrEqual,
		opcode.OpLessThan, opcode.OpLessThanOrEqual,
		opcode.OpDivisibleBy, opcode.OpNotDivisibleBy,

		opcode.OpLogicalXor, opcode.OpLogicalNXor:

		return either()

	default:
		err = c.makeErr(node, fmt.Sprintf("unknown operator (%s)", token.TypeDescription(node.Operator.Type)))
	}

	return
}
