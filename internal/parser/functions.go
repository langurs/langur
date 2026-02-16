// langur/parser/functions.go

package parser

import (
	"fmt"
	"langur/ast"
	"langur/object"
	"langur/token"
)

func (p *Parser) parseFunction() ast.Node {
	impureEffects := false

	tok := p.tok
	
	asType, _ := p.parseWord()	// will be returned if fn token by itself
	// past the fn token

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
	if p.tok.Type == token.LPAREN && p.tok.CpDiff == 0 {
		// directly attached parentheses for parameters
		// fn() { ... }
		longForm = true
		p.advanceToken()

		// ... could be self-reference call fn((x, y))
		if p.tok.Type == token.LPAREN {
			return p.finishSelfReferenceCall()
		}

		lit.PositionalParameters, lit.ByNameParameters = p.parseFunctionParameters([]token.Type{token.RPAREN}, longForm)

	} else if p.tok.Type == token.COLON && 
		p.checkContext() != context_expression_switch_condition {
		// not in a switch case condition
		
		// no parameters
		// fn: ...
		p.advanceToken()

	} else if p.tok.Type == token.IDENT {
		// fn x, y: ...
		lit.PositionalParameters, lit.ByNameParameters = p.parseFunctionParameters([]token.Type{token.COLON}, longForm)
	
	} else {
		// fn token by itself
		return asType
	}

	if longForm {
		// optional explicit return type here (long form only)
		// fn(x, y) string { ... }
		if p.tok.Type == token.IDENT {
			var code object.ObjectType
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
		param, isByName := p.parseParameter()

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

func (p *Parser) parseParameter() (param ast.Node, isByName bool) {
	// potential parts of parameter
	// 1. var keyword for mutable parameter
	// 2. internal name (required)
	// 3. as keyword followed by external name
	// 4. ... operator to indicate parameter expansion
	// 5.	[] brackets to indicate details of parameter expansion
	// 6. explicit type
	// 7. = operator followed by default value
	// ex.: var x as y string
	// ex.: x as y string = "asdf"
	// ex.: x ...[4..7]

	var expansion *ast.ExpansionNode
	if p.tok.Type == token.EXPANSION {
		// position of expansion token changed (0.20)
		p.advanceToken()
		p.addError("Parameter expansion must be specified after the parameter name, not before")
	}

	var value, vtype ast.Node

	switch p.tok.Type {
	case token.IDENT:
		param, isByName, expansion, vtype, value = p.parseIdentForParameter()
		if value != nil {
			param = ast.MakeAssignmentExpression(param, value, false)
			return
		}

	case token.VAR:
		mutable := true
		p.advanceToken()
		param, isByName, expansion, vtype, value = p.parseIdentForParameter()

		if value == nil {
			param = &ast.LineDeclarationNode{
				Token:      param.TokenInfo(),
				Mutable:    mutable,
				Assignment: param,
			}

		} else {
			param = ast.MakeDeclarationAssignmentExpression(param, value, false, mutable)
		}

	default:
		p.addError(fmt.Sprintf("Invalid parameter type %s", token.TypeDescription(p.tok.Type)))
	}

	if expansion != nil {
		if value != nil {
			p.addError("Unexpected assignment on parameter expansion")		
		} else if isByName {
			p.addError("Unexpected alias on parameter expansion")
		}
		if vtype != nil {
			p.addError("Unexpected explicit type on parameter expansion")
		}
		
		expansion.Continuation = param
		param = expansion
	}

	return
}

func (p *Parser) parseIdentForParameter() (param ast.Node, isByName bool, expansion *ast.ExpansionNode, vtype, value ast.Node) {
	var alias ast.Node
	var aliasTok token.Token

	param = p.parseIdentifier()

	if p.tok.Type == token.AS {
		// external name specified after as keyword
		isByName = true
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

	if p.tok.Type == token.EXPANSION {
		expansion = p.parseExpansionPartial()
	}

	vtype, code := p.checkParseType()
	if code != 0 {
		err := ast.AddTypeToIdent(param, vtype)
		if err != nil {
			p.addError(err.Error())
		}
	}

	if p.tok.Type == token.ASSIGN {
		// default value
		isByName = true
		p.advanceToken()
		value = p.parseExpression(precedence_LOWEST)
	}
	
	return
}

// partial in that it does not set the Continuation field yet (only sets Token and Limits)
func (p *Parser) parseExpansionPartial() *ast.ExpansionNode {
	expTok := p.tok
	
	if expTok.Type != token.EXPANSION {
		p.addError("Expected expansion token")
	}
	p.advanceToken() // past ... expansion token

	var limits ast.Node
	if p.tok.Type == token.LBRACKET {
		if p.tok.CpDiff != 0 {
			p.addError("Expected no space between expansion token and opening square bracket (expansion limits)")
		}
	
		p.advanceToken()
		limits = p.parseExpression(precedence_LOWEST)
		if p.tok.Type != token.RBRACKET {
			p.addError("Expected closing bracket for expansion limit expression")
		}
		p.advanceToken()
	}

	return &ast.ExpansionNode{
		Token:        expTok,
		Limits:       limits,
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
