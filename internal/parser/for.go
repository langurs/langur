// langur/parser/for.go

package parser

import (
	"langur/ast"
	"langur/token"
)

func (p *Parser) parseWhileLoop() ast.Node {
	if p.tok.Type != token.WHILE {
		p.addError("Expected while token")
		return nil
	}

	tok := p.tok
	p.advanceToken() // past while token

	// unless changed by user setting ...
	var loopValueVar ast.Node
	loopValueVar = ast.NewVariableNode(tok, "_while", true)
	loopValueInit := ast.MakeDeclarationAssignmentStatement(loopValueVar, ast.NoValue, true, true)

	if p.tok.Type == token.LBRACKET {
		loopValueVar, loopValueInit = p.parseExplicitForLoopValue(loopValueVar, loopValueInit)
	}

	if p.tok.Type == token.LPAREN && p.tok.CpDiff == 0 {
		p.addError("Illegal use of parenthesis on while loop token")
		return nil
	}

	f := &ast.ForNode{
		Token:         tok,
		LoopValueInit: loopValueInit,
	}
	f.Test = p.parseExpression(precedence_LOWEST)

	if p.tok.Type != token.LBRACE {
		p.addError("Expected opening curly brace")
		return f
	}

	p.forLoopVariableStack = append(p.forLoopVariableStack, loopValueVar)
	f.Body = &ast.ExpressionStatementNode{Expression: p.parseBlock()}
	p.forLoopVariableStack = ast.Pop(p.forLoopVariableStack)

	return f
}

func (p *Parser) parseForLoop() ast.Node {
	if p.tok.Type != token.FOR {
		p.addError("Expected for token")
		return nil
	}

	tok := p.tok
	p.advanceToken() // past for token

	// unless changed by user setting ...
	var loopValueVar ast.Node
	loopValueVar = ast.NewVariableNode(tok, "_for", true)
	loopValueInit := ast.MakeDeclarationAssignmentStatement(loopValueVar, ast.NoValue, true, true)

	if p.tok.Type == token.LBRACKET {
		loopValueVar, loopValueInit = p.parseExplicitForLoopValue(loopValueVar, loopValueInit)
	}

	if p.tok.Type == token.LPAREN && p.tok.CpDiff == 0 {
		p.addError("Illegal use of parenthesis on for loop token")
		return nil
	}

	if p.tok.Type == token.IN || p.tok.Type == token.OF {
		return p.finishForInOrForOf(tok, loopValueVar, loopValueInit)

	} else if p.tok.Type == token.IDENT &&
		(p.peekTok.Type == token.IN || p.peekTok.Type == token.OF) {

		return p.finishForInOrForOf(tok, loopValueVar, loopValueInit)

	} else if p.tok.Type == token.LBRACE {
		return p.finishUnlimitedLoop(tok, loopValueVar, loopValueInit)
	}

	return p.finishThreePartForLoop(tok, loopValueVar, loopValueInit)
}

func (p *Parser) finishThreePartForLoop(tok token.Token, loopValueVar, loopValueInit ast.Node) ast.Node {
	// 3 part loop
	f := &ast.ForNode{
		Token:         tok,
		LoopValueInit: loopValueInit,
	}

	// as of 0.10.0, now parsing init and increment sections as multi-variable assignments, ...
	// ... not as lists of assignments (a subtle, but important, breaking change)

	switch p.tok.Type {
	case token.LBRACE:
		// 1 part for loop uses while keyword
		p.addError("Use the while keyword for a 1 part (test only) loop")
		return f
	case token.SEMICOLON:
		// no init; is okay
	default:
		f.Init = p.parseForLoopInitialization()
	}

	p.advanceToken() // past the first semicolon
	if p.tok.Type != token.SEMICOLON {
		f.Test = p.parseExpression(precedence_LOWEST)
	}

	if p.tok.Type == token.SEMICOLON {
		p.advanceToken() // past the second semicolon
	} else {
		p.addError("Semicolon expected")
		return f
	}

	if p.tok.Type != token.LBRACE {
		f.Increment = p.parseForLoopIncrement()
		ast.ListToStatements(f.Increment)
	}

	p.forLoopVariableStack = append(p.forLoopVariableStack, loopValueVar)
	f.Body = &ast.ExpressionStatementNode{Expression: p.parseBlock()}
	p.forLoopVariableStack = ast.Pop(p.forLoopVariableStack)

	return f
}

func (p *Parser) finishUnlimitedLoop(tok token.Token, loopValueVar, loopValueInit ast.Node) ast.Node {
	f := &ast.ForNode{
		Token:         tok,
		LoopValueInit: loopValueInit,
	}

	p.forLoopVariableStack = append(p.forLoopVariableStack, loopValueVar)
	f.Body = &ast.ExpressionStatementNode{Expression: p.parseBlock()}
	p.forLoopVariableStack = ast.Pop(p.forLoopVariableStack)

	return f
}

func (p *Parser) finishForInOrForOf(tok token.Token, loopValueVar, loopValueInit ast.Node) ast.Node {
	// for in/for of
	// (not a 1 or 3 part for loop or non-variable for of loop)

	if p.tok.Type == token.IN {
		p.addError("Cannot use for in loop without variable")
		return nil
	}

	var variable ast.Node // defaulting to nil
	if p.tok.Type != token.OF {
		// not a for of loop without variable
		variable = p.parseIdentifier()
	}

	fio := &ast.ForInOfNode{
		Token:         tok,
		LoopValueInit: loopValueInit,
		Var:           variable,
	}

	if p.tok.Type == token.OF || variable == nil {
		fio.Of = true
	} else if p.tok.Type != token.IN {
		p.addError("Expected in or of for a for in/for of loop")
	}
	p.advanceToken()

	fio.Over = p.parseExpression(precedence_LOWEST)

	p.forLoopVariableStack = append(p.forLoopVariableStack, loopValueVar)
	fio.Body = &ast.ExpressionStatementNode{Expression: p.parseBlock()}
	p.forLoopVariableStack = ast.Pop(p.forLoopVariableStack)

	// compiler does not see for in/of node
	f, err := ast.ConvertForInOfNodeToForNode(fio)
	if err != nil {
		p.addError(err.Error())
	}
	return f
}

// explicitly pre-set for loop variable/value
func (p *Parser) parseExplicitForLoopValue(
	loopValueVar, loopValueInit ast.Node) (lvv, lvi ast.Node) {

	if p.tok.Type != token.LBRACKET {
		p.addError("Expected opening square bracket on for loop value initialization")
	}
	if p.tok.CpDiff != 0 {
		p.addError("Invalid use of opening bracket after for loop token (cannot have spacing)")
		return
	}

	p.advanceToken() // past opening [ bracket

	if p.tok.Type == token.ASSIGN {
		// specify for loop value assignment (using default _for)
		p.advanceToken() // past = sign
		loopValueInit = ast.MakeDeclarationAssignmentStatement(loopValueVar, p.parseExpression(precedence_LOWEST), true, true)

	} else if p.tok.Type == token.IDENT {
		loopValue := p.parseIdentifiersWithPotentialAssignments(
			false, false, false, false, false, true, false, true, false,
		)

		if assign, ok := loopValue.(*ast.AssignmentNode); ok {
			// specify for loop value variable with assignment
			if len(assign.Identifiers) != 1 || len(assign.Values) != 1 {
				p.addError("Expected 1 variable and 1 value only for explicit for loop value initialization")
				return
			}
			loopValueVar = assign.Identifiers[0].(*ast.IdentNode)
			loopValueInit = assign.Values[0]
			loopValueInit = ast.MakeDeclarationAssignmentStatement(loopValueVar, loopValueInit, true, true)

		} else {
			p.addError("Error parsing explicit for loop value initialization")
			return
		}
	}

	if p.tok.Type != token.RBRACKET {
		p.addError("Expected closing bracket after explicit for loop value initialization")
		return
	}
	p.advanceToken() // past closing ] bracket

	if p.tok.CpDiff == 0 && p.tok.Type != token.LPAREN {
		p.addError("Expected space after closing bracket on for loop value initialization")
		return
	}

	return loopValueVar, loopValueInit
}

func (p *Parser) parseForLoopInitialization() (init []ast.Node) {
	tok := p.tok
	idents := p.parseIdentifierList(false)

	if p.tok.Type == token.ASSIGN {
		p.advanceToken() // past equal sign
		initializers, ok := p.parseMultiVariableAssignmentValues(tok, idents).(*ast.AssignmentNode)
		if !ok {
			p.addError("Error parsing assignments in for loop initialization")
		}
		if len(initializers.Identifiers) != len(initializers.Values) {
			p.addError("For loop initialization requires same number of values as identifiers")
		}
		init = []ast.Node{initializers}

	} else {
		p.addError("Expected assignment in for loop initialization")
	}

	var err error
	init, err = ast.AssignmentsToDeclarations(init, true, false)
	if err != nil {
		p.addError(err.Error())
	}
	ast.ListToStatements(init)

	return
}

func (p *Parser) parseForLoopIncrement() []ast.Node {
	tok := p.tok
	idents := p.parseIdentifierList(true)

	if p.tok.Type == token.ASSIGN {
		p.advanceToken() // past equal sign
		return []ast.Node{p.parseMultiVariableAssignmentValues(tok, idents)}

	} else if token.IsComboOp(p.tok) {
		if len(idents) != 1 {
			p.addError("Combination operation can only be used with one identifier")
			return []ast.Node{}
		}
		return []ast.Node{p.parseCombinationAssignment(idents[0])}

	} else {
		p.addError("Expected assignment in for loop increment")
		return []ast.Node{}
	}
}
