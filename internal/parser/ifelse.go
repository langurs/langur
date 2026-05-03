// langur/parser/ifelse.go

package parser

import (
	"fmt"
	"langur/ast"
	"langur/token"
)

func (p *Parser) parseIfExpression() ast.Node {
	var ta ast.TestDo

	expr := &ast.IfNode{Token: p.tok}

	expr.TestsAndActions = make([]ast.TestDo, 0, 2)

	// skip the "if" token
	p.advanceToken()

	// check for shortened form if expression...
	if p.tok.Type == token.LPAREN && p.tok.CpDiff == 0 {
		return p.parseShortenedFormIfExpression()
	}

	// parse first test expression
	ta.Test = p.parseIfTest()

	switch p.tok.Type {
	case token.LBRACE:
		// parse first action
		ta.Do = p.parseBlock()

	case token.COLON:
		if p.checkContext() == context_expression_switch_case {
			// both simple if and case using a colon; no confusion...
			p.addError("Cannot use simple if within switch case test; use shortened form if instead")
		}
		return p.finishSimpleIf(expr, ta)

	default:
		p.addError("Expected opening curly brace for if expression body, or colon for simple if expression body")
		return nil
	}

	// add test and action to the slice
	expr.TestsAndActions = append(expr.TestsAndActions, ta)

	for p.tok.Type == token.ELSE {
		// skip else token
		p.advanceToken()

		if p.tok.Type == token.IF {
			// skip the if token
			p.advanceToken()

			// parse next test expression
			ta.Test = p.parseIfTest()
		} else {
			// final else, no test
			ta.Test = nil
		}

		// parse next action
		if p.tok.Type != token.LBRACE {
			p.addError("Expected opening curly brace for if/else expression body")
			return nil
		}
		ta.Do = p.parseBlock()

		// add test and action to the slice
		expr.TestsAndActions = append(expr.TestsAndActions, ta)

		// on last else, break from loop
		if ta.Test == nil {
			break
		}
	}

	return expr
}

func (p *Parser) parseIfTest() ast.Node {
	var expr ast.Node

	if p.tok.Type == token.VAL || p.tok.Type == token.VAR {
		expr = p.parseDeclaration()

	} else if p.tok.Type == token.IDENT || p.tok.Type == token.NONE {
		expr := p.parseIdentifiersWithPotentialAssignments(
			false, true, false, false, true, false, true, false, false,
		)

		switch p.tok.Type {
		case token.COLON, token.LBRACE:
			return expr
		default:
			return p.parseContinuedExpression(expr, precedence_LOWEST)
		}

	} else {
		expr = p.parseExpression(precedence_LOWEST)
	}

	return expr
}

func (p *Parser) finishSimpleIf(expr *ast.IfNode, ta ast.TestDo) ast.Node {
	// parse *simple if* expression
	if p.tok.NewLinePrecedes {
		p.addError("Colon to indicate simple if expression must be on same line as the test")
		return nil
	}
	p.advanceToken() // past the colon

	if p.tok.NewLinePrecedes {
		// at least for now; might be relaxed later...
		p.addError("Expression after colon must be on same line as the test for a simple if")
		return nil
	}

	ta.Do, _ = p.parseBlockOrIntoBlockWithPotentialAssignment()
	expr.TestsAndActions = append(expr.TestsAndActions, ta)
	return expr
}

func (p *Parser) parseShortenedFormIfExpression() ast.Node {
	var ta ast.TestDo

	expr := &ast.IfNode{Token: p.tok}

	expr.TestsAndActions = make([]ast.TestDo, 0, 2)

	addTA := func() {
		ta.Do = ast.BlockOrAsBlock(ta.Do)
		// add test and action to the slice
		expr.TestsAndActions = append(expr.TestsAndActions, ta)
	}

	// skip opening (
	p.advanceToken()
	ta.Test = p.parseExpression(precedence_LOWEST)

	if p.tok.Type != token.COLON {
		p.addError("Expected colon between test condition and action for shortened form if expression")
		return nil
	}
	// skip colon :
	p.advanceToken()

	ta.Do = p.parseExpression(precedence_LOWEST)
	addTA()

	for p.tok.Type != token.RPAREN {
		if p.tok.Type != token.SEMICOLON {
			p.addError(fmt.Sprintf("Expected semicolon between sections of shortened form if expression (not %s)", token.TypeDescription(p.tok.Type)))
			break
		}
		// skip semicolon
		p.advanceToken()

		// not knowing if the next sub-expression is a test or final action (else), just parse it first...
		testOrDo := p.parseExpression(precedence_LOWEST)

		if p.tok.Type == token.COLON {
			p.advanceToken()
			ta.Test = testOrDo
			ta.Do = p.parseExpression(precedence_LOWEST)
		} else if p.tok.Type == token.RPAREN {
			// final else, no test
			ta.Test = nil
			ta.Do = testOrDo
		} else {
			p.addError("Expected colon or closing parenthesis within shortened form if")
			break
		}
		addTA()

		// on last else, break from loop
		if ta.Test == nil {
			break
		}
	}
	// skip closing parenthesis
	p.advanceToken()

	return expr
}
