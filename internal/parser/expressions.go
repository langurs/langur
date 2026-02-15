// langur/parser/expressions.go

package parser

import (
	"fmt"
	"langur/ast"
	"langur/common"
	"langur/lexer"
	"langur/str"
	"langur/token"
	"strings"
)

func (p *Parser) likelyInfixPosition() bool {
	if p.tok.Type == token.NOT {
		return p.peekTok.Type == token.IN || p.peekTok.Type == token.OF

	} else if token.IsInfixOp(p.tok.Type) {
		switch p.tok.Type {
		case token.MINUS:
			// Is the minus a prefix or an infix?
			// using spacing to determine
			return p.peekTok.CpDiff != 0
		}
		return true

	} else {
		return false
	}
}

func (p *Parser) parseExpression(prec precedence) ast.Node {
	prefix := p.prefixParseFns[p.tok.Type]
	if prefix == nil {
		// some kind of error
		if p.tok.Type == token.SEMICOLON && p.tok.Literal == lexer.IMPLIED_EXPRESSION_TERMINATOR_LITERAL {
			p.addError("Illegal implied expression termination")
		} else {
			p.addError(fmt.Sprintf("Illegal use of token %q (%s)",
				str.ReformatInput(p.tok.Literal), token.TypeDescription(p.tok.Type)))
		}
		p.advanceToken()
		return nil
	}

	return p.parseContinuedExpression(prefix(), prec)
}

func (p *Parser) parseContinuedExpression(leftExp ast.Node, prec precedence) ast.Node {
	// checking if necessary to defer to rightmost operation...
	// ...becuase of either precedence or associativity
	for prec < getInfixPrecedence(p.tok.Type) ||
		prec == getInfixPrecedence(p.tok.Type) && token.IsRightAssociativeOp(p.tok.Type) {

		infix := p.infixParseFns[p.tok.Type]
		if infix == nil {
			break
		}

		leftExp = infix(leftExp)
	}

	postfix := p.postfixParseFns[p.tok.Type]
	if postfix != nil && p.tok.CpDiff == 0 {
		// might relax it later (or may depend on operator), but initially must be directly attached (no spacing)
		leftExp = postfix(leftExp)
	}

	return leftExp
}

func (p *Parser) parseList() ast.Node {
	if p.tok.CpDiff == 0 && !token.MayPrecedeOpeningBracketWithoutSpacing(p.prevTok) {
		// disallow len[1, 2] without a space
		p.addError("Space required between word and list")
	}

	list := &ast.ListNode{Token: p.tok}
	p.advanceToken()
	list.Elements, _ = p.parseExpressionList([]token.Type{token.RBRACKET}, token.COMMA, false, true, false)

	return list
}

func (p *Parser) parseHash() ast.Node {
	return p.finishParsingHash(nil)
}

func (p *Parser) finishParsingHash(firstKey ast.Node) ast.Node {
	hash := &ast.HashNode{Token: p.tok}

	if firstKey == nil {
		if p.tok.Type != token.LBRACE {
			p.addError("Expected opening brace for hash literal")
			return hash
		}

		// past the {
		p.advanceToken()

	} else {
		// break out of expression statement
		f, isExprStatement := firstKey.(*ast.ExpressionStatementNode)
		if isExprStatement {
			firstKey = f.Expression
		}
	}

	var key ast.Node
	for p.tok.Type != token.RBRACE && p.tok.Type != token.EOF {
		if firstKey == nil {
			key = p.parseExpression(precedence_LOWEST)
		} else {
			key = firstKey
			firstKey = nil
		}

		if p.tok.Type != token.COLON {
			p.addError("Not the expected token between key and value in hash literal")
			return hash
		}
		p.advanceToken()

		value := p.parseExpression(precedence_LOWEST)

		hash.Pairs = append(hash.Pairs, ast.KeyValuePair{Key: key, Value: value})

		if p.tok.Type != token.COMMA && p.tok.Type != token.RBRACE {
			p.addError("Expected comma or closing brace in hash literal")
			return hash
		}
		if p.tok.Type == token.COMMA {
			p.advanceToken()
		}
	}

	p.advanceToken()

	return hash
}

func (p *Parser) parseIndices(ident ast.Node) ast.Node {
	index := ident
	for p.tok.Type == token.LBRACKET {
		// might be multiple indices; thus the for loop
		index = p.parseIndexExpression(index)
	}
	return index
}

func (p *Parser) parseIndexExpression(left ast.Node) ast.Node {
	if p.tok.Type != token.LBRACKET {
		p.addError("Expected opening [ bracket on index expression")
		return nil
	}

	if p.tok.CpDiff > 0 {
		p.addError("Index expression may not have spacing between token and opening bracket")
		p.advanceToken()
		return nil
	}

	expr := &ast.IndexNode{Token: p.tok, Left: left}
	p.advanceToken()

	expr.Index = p.parseExpression(precedence_LOWEST)

	// altenative for index expressions
	if p.tok.Type == token.SEMICOLON {
		p.advanceToken()
		expr.Alternate = p.parseExpression(precedence_LOWEST)
	}

	if p.tok.Type != token.RBRACKET {
		p.addError("Expected closing ] bracket on index expression")
		return nil
	}
	p.advanceToken()

	return expr
}

func (p *Parser) parseParenthesizedExpression() ast.Node {
	if p.tok.Type != token.LPAREN {
		p.addError("Expected opening ( parenthesis")
	}

	p.advanceToken() // past opening (
	expr := p.parseExpression(precedence_LOWEST)

	if p.tok.Type != token.RPAREN {
		p.addError("Expected closing ) parenthesis")
	}
	p.advanceToken()

	return expr
}

func (p *Parser) parseBlock() ast.Node {
	if p.tok.Type != token.LBRACE {
		p.addError("Expected opening { brace")
	}
	statements, _ := p.parseStatements([]token.Type{token.RBRACE}, nil, true, true)
	return ast.ListToBlock(statements)
}

// left brace
// don't know what it is yet: scope block or hash
func (p *Parser) parseLBrace() ast.Node {
	tok := p.tok
	p.advanceToken() // past left brace {

	switch p.tok.Type {
	case token.RBRACE:
		// {}
		p.addError("Invalid use of {}; for an empty hash, use {:}")
		return nil

	case token.COLON:
		return p.finishParsingEmptyHash()

	case token.COMMA:
		// set (hypothetical type so far)
		// empty set as {,}
		p.addError("Unexpected comma")
		return nil
	}

	// parse first statement, not knowing if we have a scope block or something else
	first := p.parseStatement(false)
	var statements []ast.Node

	switch p.tok.Type {
	case token.COLON:
		// hash
		return p.finishParsingHash(first)

	case token.COMMA:
		// set (hypothetical type so far)
		p.addError("Unexpected comma")
		return nil

	case token.RBRACE:
		// scope block with one statement
		p.advanceToken()
		statements = []ast.Node{first}

	case token.SEMICOLON:
		// scope block with multiple statements
		p.advanceToken()
		statements, _ = p.parseStatements([]token.Type{token.RBRACE}, first, false, true)
	}

	// have scope block
	block := &ast.BlockNode{Token: tok, Statements: statements}
	block.HasScope = true

	// limit scope blocks to statement context
	// if p.checkContext() != context_unknown_block {
	// 	p.addError("Unexpected scope block in expression context")
	// }

	return block
}

func (p *Parser) finishParsingEmptyHash() ast.Node {
	// empty hash literal {:}
	if p.tok.CpDiff != 0 {
		p.addError("Expected no space between { and :")
	}
	p.advanceToken()
	if p.tok.Type == token.RBRACE {
		if p.tok.CpDiff != 0 {
			p.addError("Expected no space between : and }")
		}
		p.advanceToken()

	} else {
		p.addError("Expected } to close empty literal")
	}
	return &ast.HashNode{}
}

func (p *Parser) parseBlockOrIntoBlock() (node ast.Node, wasBlock bool) {
	if p.tok.Type == token.LBRACE {
		node, wasBlock = p.parseBlock(), true

	} else if token.BeginsFlowBreakingStatement(p.tok.Type) {
		node, wasBlock = &ast.BlockNode{Token: p.tok, Statements: []ast.Node{p.parseStatement(true)}}, false

	} else {
		// single expression for body (no braces required)
		node, wasBlock = &ast.BlockNode{Token: p.tok,
			Statements: []ast.Node{p.parseExpression(precedence_LOWEST)}}, false
	}
	return
}

func (p *Parser) parseBlockOrIntoBlockWithPotentialAssignment() (node ast.Node, wasBlock bool) {
	if p.tok.Type == token.LBRACE {
		node, wasBlock = p.parseBlock(), true

	} else if token.BeginsFlowBreakingStatement(p.tok.Type) {
		node, wasBlock = &ast.BlockNode{Token: p.tok, Statements: []ast.Node{p.parseStatement(true)}}, false

	} else {
		node, wasBlock = &ast.BlockNode{Token: p.tok,
			Statements: []ast.Node{p.parseExpressionWithPotentialAssignment()}}, false
	}
	return
}

func (p *Parser) parsePrefixExpression() ast.Node {
	expr := &ast.PrefixExpressionNode{Token: p.tok, Operator: p.tok}
	tt := p.tok.Type

	prec := getPrefixPrecedence(tt)
	p.advanceToken()

	if p.tok.CpDiff != 0 && token.IsValueOp(tt) {
		// minus directly attached to number?
		p.addError("Spacing between prefix operator and expression")
	}

	expr.Right = p.parseExpression(prec)

	if token.IsComboOp(expr.Operator) {
		// prefix combination operator, such as not= or not?=
		return ast.MakeAssignmentExpression(expr.Right, expr, false)

	} else if r, isNumber := expr.Right.(*ast.NumberNode); isNumber && tt == token.MINUS {
		// precedence already taken care of (evaluated Right Node with precedence)
		// fix straight-up number negation here (no need to evaluate it later)
		if r.Value[0] == '-' {
			r.Value = r.Value[1:]
		} else {
			r.Value = "-" + r.Value
		}
		return r
	}

	return expr
}

func (p *Parser) parsePostfixExpression(left ast.Node) ast.Node {
	expr := &ast.PostfixExpressionNode{Left: left, Token: p.tok, Operator: p.tok}
	p.advanceToken()
	return expr
}

func (p *Parser) parseInfixOperator() token.Token {
	tok := p.tok
	p.advanceToken() // pass the initial op token

	if tok.Type == token.NOT {
		if 0 != tok.Code {
			// can't have not? here
			p.addError("Invalid combination on use of not in or not of operator")
			return tok
		}

		switch p.tok.Type {
		case token.IN:
			tok.Type, tok.Literal = p.tok.Type, common.NotInLiteral
			p.advanceToken() // pass the in/of token
		case token.OF:
			tok.Type, tok.Literal = p.tok.Type, common.NotOfLiteral
			p.advanceToken() // pass the in/of token
		}
	}

	if !token.IsInfixOp(tok.Type) {
		p.addError(fmt.Sprintf("Expected infix operator, received %s", token.TypeDescription(tok.Type)))
		return tok
	}

	if tok.Type == token.IS && p.tok.Type == token.NOT {
		// the is not operator
		tok.Literal = common.IsNotLiteral // used as indicator in compiler

		if 0 != p.tok.Code {
			// can't have not? here
			p.addError("Invalid combination on use of is not operator")
			return tok
		}

		p.advanceToken() // pass the not token

		if p.tok.CpDiff == 0 && p.tok.Type != token.RBRACE && p.tok.Type != token.RBRACKET {
			// can't have not() here
			p.addError("Expected space between not of is not operator and right operand")
			return tok
		}
	}

	return tok
}

func (p *Parser) parseInfixNilLeftExpression() ast.Node {
	return p.parseInfixExpression(nil)
}

func (p *Parser) parseInfixExpression(left ast.Node) ast.Node {
	if token.IsComboOp(p.tok) {
		// combination operator used in expression context
		p.addError("Unexpected combination operator in expression context")
	}

	expr := &ast.InfixExpressionNode{
		Operator: p.parseInfixOperator(), Left: left}

	if left == nil {
		// comparison operations for switch expressions such as "> 123" or "is number"
		expr.Token = expr.Operator

	} else {
		expr.Token = left.TokenInfo()
	}

	if !token.AllowNilRightExpression(p.tok.Type) ||
		p.tok.Type == token.LBRACE && p.checkContext() != context_expression_switch_test {
		// context check to not confuse right operand { (such as beginning a hash)...
		// ... with { after switch test

		// // special case for *is fn* or *is not fn*
		// if expr.Operator.Type == token.IS && p.tok.Type == token.FUNCTION {
		// 	expr.Right, _ = p.parseWord()

		expr.Right = p.parseExpression(getInfixPrecedence(expr.Operator.Type))

		rightIsType := ast.NodeToLangurType(expr.Right) != 0

		if rightIsType && expr.Operator.Type == token.FORWARD {
			// change to call node - call on type, such as 123 -> string, meaning string(123)
			return &ast.CallNode{Token: expr.Token,
				Function: expr.Right, PositionalArgs: []ast.Node{expr.Left}}
		}
	}

	return expr
}

func (p *Parser) parseDotExpression(left ast.Node) ast.Node {
	expr := &ast.InfixExpressionNode{
		Token:    left.TokenInfo(),
		Operator: p.parseInfixOperator(),
		Left:     left,
	}

	if expr.Operator.CpDiff != 0 {
		p.addError("Expected no spacing before dot in dot expression")
	}

	var ok bool
	expr.Right, ok = p.parseWord()
	if !ok {
		p.addError("Expected word token to continue dot expression")
		return nil
	}

	if expr.Right.TokenInfo().CpDiff != 0 {
		p.addError("Expected no spacing after dot in dot expression")
	}

	return expr
}

func (p *Parser) parseNone() ast.Node {
	expr := &ast.NoneNode{Token: p.tok}
	p.advanceToken()
	return expr
}

func (p *Parser) parseBoolean() ast.Node {
	expr := &ast.BooleanNode{Token: p.tok, Value: p.tok.Type == token.TRUE}
	p.advanceToken()
	return expr
}

func (p *Parser) parseNull() ast.Node {
	expr := &ast.NullNode{Token: p.tok}
	p.advanceToken()
	return expr
}

func (p *Parser) parseNumber() ast.Node {
	var base int

	// removing underscores first ...
	numStr := strings.Replace(p.tok.Literal, "_", "", -1)

	imaginary := 0 != p.tok.Code&token.CODE_IMAGINARY_NUMBER

	if p.tok.Code2 == 0 {
		base = 10
	} else {
		base = p.tok.Code2
	}

	node := &ast.NumberNode{Token: p.tok, Value: numStr, Base: base, Imaginary: imaginary}
	p.advanceToken()

	return node
}

func (p *Parser) parseExpressionList(
	until []token.Type,
	delimiterTokType token.Type,
	forFunctionCall,
	passClosingToken, endBeforeCommaPrecedingNewLine bool) (

	nodes []ast.Node, closingtt token.Type) {

	nodes = []ast.Node{}

	for {
		closingtt = p.tok.Type

		if p.checkUpdateStopNow() {
			break
		}

		if p.tok.Type == token.EOF {
			if !token.InTypeSlice(token.EOF, until) {
				p.addError("EOF reached without closing expression list")
				return
			}
			break
		}

		if token.InTypeSlice(p.tok.Type, until) {
			// NOTE: if p.prevTok.Type == delimiterTokType be sure not to add a nil node
			// A comma might be used like it is in Go ...
			// ... at the end of a list item before a line return (doesn't end expression).

			if passClosingToken {
				p.advanceToken()
			}
			break
		}

		if p.tok.Type == delimiterTokType {
			p.addError("Expected expression in list, not free delimiter")
		}

		if forFunctionCall &&
			identifierRegex.MatchString(p.tok.Literal) &&
			p.peekTok.Type == token.ASSIGN {
			// for optional argument
			// externalname = value

			externalName, _ := p.parseWord() // not parsing as normal identifier
			p.advanceToken()                 // past assignment operator
			value := p.parseExpression(precedence_LOWEST)
			assign := ast.MakeAssignmentExpression(externalName, value, false)

			nodes = append(nodes, assign)

		} else {
			nodes = append(nodes, p.parseExpression(precedence_LOWEST))
		}

		if p.tok.Type == delimiterTokType {
			if endBeforeCommaPrecedingNewLine && p.peekTok.NewLinePrecedes {
				// unbounded argument list should end at a newline without consuming this trailing comma
				break
			}
			// good; keep going
			p.advanceToken()
			continue

		} else if !token.InTypeSlice(p.tok.Type, until) {
			p.addError("Expected delimiter or closing between expressions")
		}
	}

	return
}
