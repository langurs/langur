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

		lit.PositionalParameters, lit.ByNameParameters = p.parseFunctionParameters([]token.Type{token.RPAREN}, longForm)

	} else if p.tok.Type == token.COLON {
		// no parameters
		// fn: ...
		p.advanceToken()

	} else {
		// fn x, y: ...
		lit.PositionalParameters, lit.ByNameParameters = p.parseFunctionParameters([]token.Type{token.COLON}, longForm)
	}

	if longForm {
		// optional explicit return type here (long form only)
		// fn(x, y) : string { ... }

		if p.tok.Type == tokenTypeBetweenVarNameAndType && p.peekTok.Type == token.IDENT {
			p.advanceToken()  // past the colon
			var code int
			lit.ReturnType, code = p.parseType()
			if code == 0 {
				p.addError("Unexpected identifier token; not a return type for function")
			}
		}

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

func (p *Parser) parseFunctionParameters(until []token.Type, longForm bool) (
	positional, byname []ast.Node) {

	for !token.InTypeSlice(p.tok.Type, until) {
		param, isByName := p.parseParameter(0, longForm)

		if isByName {
			byname = append(byname, param)

		} else {
			if len(byname) != 0 {
				p.addError("Cannot have positional parameter after parameter by name")
			}
			positional = append(positional, param)
		}

		if p.tok.Type == token.COMMA {
			p.advanceToken()

		} else if !token.InTypeSlice(p.tok.Type, until) {
			p.addError("Expected comma or closing parenthesis on parameter list")
			break
		}
	}

	if p.tok.Type != token.LBRACE {
		p.advanceToken()
	}

	return
}

func (p *Parser) parseParameter(level int, longForm bool) (param ast.Node, isByName bool) {
	var value, alias ast.Node
	var aliasTok token.Token

	// potential parts of parameter
	// 1. ... operator to indicate parameter expansion
	// 2.	[] brackets to indicate details of parameter expansion
	// 3. var keyword
	// 4. internal name (required)
	// 5. as keyword followed by external name
	// 6. : operator followed by explicit type (long form only)
	// 7. = operator followed by default value
	// ex.: var x as y : string = "asdf"

	parseIdentAliasAndAssignment := func() {
		param = p.parseIdentifier()

		if p.tok.Type == token.AS {
			// external name specified after as keyword
			isByName = true
			if level != 0 {
				p.addError("Unexpected alias on parameter expansion")
			}

			aliasTok = p.tok
			p.advanceToken() // past as keyword
			var ok bool
			alias, ok = p.parseWord()
			if !ok {
				p.addError("Error parsing alias for parameter (a as b)")
			}
			// param already set as ident
			// having an alias, set param as infix expression, a as b
			param = &ast.InfixExpressionNode{
				Token:    param.TokenInfo(),
				Left:     param,
				Operator: aliasTok,
				Right:    alias,
			}
		}

		if longForm && p.tok.Type == tokenTypeBetweenVarNameAndType {
			// explicit type
			p.advanceToken()
			_, code := p.parseType()
			if code != 0 {
				p.addError("This version of langur not set up to parse explicit parameter type")
				
			} else {
				p.addError("Expected parameter type following colon")
			}
		}

		if p.tok.Type == token.ASSIGN {
			// default value
			isByName = true
			if level != 0 {
				p.addError("Unexpected assignment on parameter expansion")
			}
			p.advanceToken()
			value = p.parseExpression(precedence_LOWEST)
		}
	}

	switch p.tok.Type {
	case token.IDENT:
		parseIdentAliasAndAssignment()
		if value != nil {
			param = ast.MakeAssignmentExpression(param, value, false)
			return
		}
		return

	case token.VAR:
		mutable := true
		p.advanceToken()
		parseIdentAliasAndAssignment()

		if value == nil {
			param = &ast.LineDeclarationNode{
				Token:      param.TokenInfo(),
				Mutable:    mutable,
				Assignment: param,
			}

		} else {
			param = ast.MakeDeclarationAssignmentExpression(param, value, false, mutable)
		}

		return

	case token.EXPANSION:
		if level > 0 {
			p.addError("Error parsing parameters")
			p.advanceToken()
			return
		}
		param = p.parseParameterExpansion(level, longForm)
		return

	default:
		p.addError(fmt.Sprintf("Invalid parameter type %s", token.TypeDescription(p.tok.Type)))
	}
	return
}

func (p *Parser) parseParameterExpansion(level int, longForm bool) ast.Node {
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
	continuation, isByName := p.parseParameter(level + 1, longForm)

	if isByName {
		p.addError("Cannot use parameter expansion on parameter by name")
	}

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
			args, _ := p.parseExpressionList(
				token.EndUnboundedArgumentList, token.COMMA, true, false, true)

			var err error
			expr.PositionalArgs, expr.ByNameArgs, err = ast.SplitArgumentSliceToPositionalAndByName(args)
			if err != nil {
				p.addError("Error determining positional arguments and by name arguments: " + err.Error())
			}

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
	args, _ := p.parseExpressionList(
		[]token.Type{token.RPAREN}, token.COMMA, true, true, false)
	// passes closing parenthesis

	var err error
	expr.PositionalArgs, expr.ByNameArgs, err = ast.SplitArgumentSliceToPositionalAndByName(args)
	if err != nil {
		p.addError("Error determining positional arguments and by name arguments: " + err.Error())
	}

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
