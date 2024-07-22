// langur/parser/functions.go

package parser

import (
	"fmt"
	"langur/ast"
	"langur/token"
)

func (p *Parser) parseFunction() ast.Node {
	impureEffects := false

	tok := p.tok
	p.advanceToken() // past the fn token

	if tok.Type != token.FUNCTION {
		p.addError("Expected function token")
		return nil
	}

	if p.tok.Literal == "*" {
		if p.tok.CpDiff == 0 {
			impureEffects = true
			p.advanceToken()
		} else {
			p.addError("Expected no space between fn and * tokens")
			return nil
		}
	}

	lit := &ast.FunctionNode{
		Token:         tok,
		ImpureEffects: impureEffects,
	}

	if specialFn := p.maybeParseSpecialFunction(); specialFn != nil {
		if impureEffects {
			p.addError("Special function marked as having impure effects")
		}
		return specialFn
	}

	longForm := false
	if p.tok.Type == token.LPAREN {
		// directly attached parentheses for parameters
		// fn() { ... }
		longForm = true

		if p.tok.CpDiff != 0 {
			p.addError("Expected parenthesis directly after fn token")
			return lit
		}
		p.advanceToken()

		// ... or could be self-reference call fn((x, y))
		if p.tok.Type == token.LPAREN {
			return p.finishSelfReferenceCall()
		}

		lit.Parameters = p.parseFunctionParameters([]token.Type{token.RPAREN})

	} else if p.tok.Type == token.COLON {
		// no parameters
		// fn: ...
		p.advanceToken()

	} else {
		// fn .x, .y: ...
		lit.Parameters = p.parseFunctionParameters([]token.Type{token.COLON})
	}

	if longForm {
		// NOTE: potential optional explicit return type here
		// fn(x, y) string { ... }

		if p.tok.Type != token.LBRACE {
			p.addError("Expected left brace { to start function body for long form")
		}
	}

	if p.tok.CpDiff > 1 {
		p.addError("0..1 spaces expected between fn() tokens and function body")
	}

	lit.Body, _ = p.parseBlockOrIntoBlock()

	return lit
}

func (p *Parser) parseFunctionParameters(until []token.Type) (params []ast.Node) {
	for !token.InTypeSlice(p.tok.Type, until) {
		params = append(params, p.parseParameter(0))

		if p.tok.Type == token.COMMA {
			p.advanceToken()

		} else if !token.InTypeSlice(p.tok.Type, until) {
			p.addError("Expected comma or closing parenthis on parameter list")
			break
		}
	}

	if p.tok.Type != token.LBRACE {
		p.advanceToken()
	}

	return params
}

func (p *Parser) parseParameter(level int) ast.Node {
	var ident, value, alias ast.Node
	var aliasTok token.Token

	parseIdentAliasAndAssignment := func() {
		ident = p.parseIdentifier()
		if p.tok.Type == token.AS {
			if level != 0 {
				p.addError("Unexpected alias on parameter expansion")
			}
			aliasTok = p.tok
			p.advanceToken()
			alias = p.parseIdentifier()
		}
		if p.tok.Type == token.ASSIGN {
			if level != 0 {
				p.addError("Unexpected assignment on parameter expansion")
			}
			p.advanceToken()
			value = p.parseExpression(precedence_LOWEST)
			if alias != nil {
				ident = &ast.InfixExpressionNode{
					Token:    ident.TokenInfo(),
					Left:     ident,
					Operator: aliasTok,
					Right:    alias,
				}
			}
		}
	}

	switch p.tok.Type {
	case token.IDENT:
		parseIdentAliasAndAssignment()
		if value != nil {
			return ast.MakeAssignmentExpression(ident, value, false)
		}
		if alias != nil {
			p.addError("Expected assignment after alias in parameter")
		}
		return ident

	case token.VAR:
		mutable := true
		p.advanceToken()
		parseIdentAliasAndAssignment()
		if value != nil {
			return ast.MakeDeclarationAssignmentExpression(ident, value, false, mutable)
		}
		if alias != nil {
			p.addError("Expected assignment after alias in parameter")
		}
		return &ast.LineDeclarationNode{
			Token:      ident.TokenInfo(),
			Mutable:    mutable,
			Assignment: ident,
		}

	case token.EXPANSION:
		if level > 0 {
			p.addError("Error parsing parameters")
			p.advanceToken()
			return nil
		}
		return p.parseExpansion(level)

	default:
		p.addError(fmt.Sprintf("Invalid parameter type %s", token.TypeDescription(p.tok.Type)))
	}
	return nil
}

func (p *Parser) parseExpansion(level int) ast.Node {
	exp := p.tok
	p.advanceToken() // past ... expansion token

	var limits ast.Node
	switch p.tok.Type {
	case token.LBRACKET:
		p.advanceToken()
		limits = p.parseExpression(precedence_LOWEST)
		if p.tok.Type != token.RBRACKET {
			p.addError("Expected closing bracket for expansion limit expression")
		}
		p.advanceToken()
	}
	continuation := p.parseParameter(level + 1)

	return &ast.ExpansionNode{
		Token:        exp,
		Limits:       limits,
		Continuation: continuation,
	}
}

func (p *Parser) maybeParseSpecialFunction() ast.Node {
	// already past fn token
	if p.tok.Type == token.LBRACE &&
		p.peekTok.Type == token.RBRACE {
		// the empty function without parameters
		if p.tok.CpDiff != 0 || p.peekTok.CpDiff != 0 {
			p.addError("No space expected on empty function body and no parameters")
		}
		lit := &ast.FunctionNode{Token: p.tok}
		p.advanceToken()
		p.advanceToken()
		return lit
	}

	if p.tok.Type != token.LBRACE ||
		p.tok.CpDiff != 0 { // no spacing in fn{
		// not an operator implied functionast
		return nil
	}
	p.advanceToken() // past opening {
	if p.tok.CpDiff != 0 {
		p.addError("Expected operator with no spacing after opening curly brace for operator implied function")
	}

	if p.peekTok.Type == token.RBRACE &&
		token.IsNumeric(p.tok.Type) && p.tok.Literal[0] == '-' {

		// remove minus sign from number literal
		p.tok.Literal = p.tok.Literal[1:]
		// generate and pass a minus operator
		op := p.tok
		op.Literal, op.Type, op.Code = "(-)", token.MINUS, 0
		return p.parseNilLeftPartiallyImpliedFunction(op)

	} else if token.MayBeUsedForOperatorImpliedFunction(p.tok.Type) {
		op := p.parseInfixOperator()

		if p.tok.Type == token.RBRACE {
			lit, err := ast.MakeFunctionFromOperator(op, nil, nil)
			if err != nil {
				p.addError(fmt.Sprintf("Error generating operator implied function: %s", err.Error()))
			}

			if p.tok.CpDiff != 0 {
				p.addError("Expected closing curly brace with no spacing for operator implied function")
			}
			p.advanceToken() // past closing }

			return lit

		} else {
			return p.parseNilLeftPartiallyImpliedFunction(op)
		}
	}

	p.addError("Invalid use of function fn{...} tokens")
	return nil
}

func (p *Parser) parseNilLeftPartiallyImpliedFunction(op token.Token) *ast.FunctionNode {
	// nil left partially implied function
	// already past the operator
	if token.IsInfixOp(p.tok.Type) {
		p.addError("Expected expression after operator for nil left partially implied function")
		return nil
	}
	right := p.parseExpression(precedence_LOWEST)
	lit, err := ast.MakeFunctionFromOperator(op, nil, right)
	if err != nil {
		p.addError(fmt.Sprintf("Error generating nil left partially implied function: %s", err.Error()))
	}
	if p.tok.Type != token.RBRACE {
		p.addError("Expected closing curly brace on nil left partially implied function")
	}
	p.advanceToken() // past clsing }

	return lit
}

// possibly a call without parentheses
func (p *Parser) parsePossibleUnboundedCall(ident ast.Node) (ast.Node, bool) {
	if p.tok.CpDiff != 0 && token.MayStartFunctionArg(p.tok.Type) {
		if !p.likelyInfixPosition() {
			expr := &ast.CallNode{Token: p.tok, Function: ident}
			expr.Args, _ = p.parseExpressionList(
				token.EndUnboundedArgumentList, token.COMMA, true, false, true)
			return expr, true
		}
	}
	return ident, false
}

func (p *Parser) parseParenthesizedCallExpression(fn ast.Node) ast.Node {
	if p.tok.Type != token.LPAREN {
		p.addError("Expected opening parenthesis on this call expression")
		return nil
	}
	if p.tok.CpDiff > 0 {
		p.addError("Call expression may not have spacing between token and opening parenthesis")
		p.advanceToken()
		return nil
	}

	expr := &ast.CallNode{Token: p.tok, Function: fn}
	p.advanceToken() // past the opening parenthesis
	expr.Args, _ = p.parseExpressionList(
		[]token.Type{token.RPAREN}, token.COMMA, true, true, false)
	// passes closing parenthesis

	return expr
}

func (p *Parser) finishSelfReferenceCall() ast.Node {
	// fn((x, y))
	// already past first opening parenthesis
	if p.tok.Type == token.LPAREN {
		if p.tok.CpDiff != 0 {
			p.addError("Expected second opening parenthesis directly after first")
			return nil
		}
		// don't advance token

	} else {
		p.addError("Expected second opening parenthesis")
	}

	expr := p.parseParenthesizedCallExpression(&ast.SelfNode{Token: p.tok})

	// already past first closing parenthesis
	if p.tok.Type == token.RPAREN {
		if p.tok.CpDiff != 0 {
			p.addError("Expected second closing parenthesis directly after first")
			return nil
		}
		p.advanceToken()

	} else {
		p.addError("Expected second closing parenthesis")
	}

	return expr
}
