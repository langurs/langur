// langur/parser/statements.go

package parser

import (
	"fmt"
	"langur/ast"
	"langur/lexer"
	"langur/modes"
	"langur/token"
)

func (p *Parser) parseStatement(eatSemicolon bool) ast.Node {
	var stmt ast.Node

	p.pushContext(context_statement)
	defer p.popContext()

	switch p.tok.Type {
	case token.RBRACE:
		stmt = p.parseLBrace()
		stmt = &ast.ExpressionStatementNode{
			Token:      stmt.TokenInfo(),
			Expression: stmt}

	case token.CATCH:
		// A catch in langur is neither an expression nor a statement (is part of an expression).
		// It should not be parsed as an expression itself.
		// It should be parsed in statement context.
		stmt = p.parseCatch()

	case token.THROW:
		stmt = p.parseThrowStatement()

	case token.BREAK:
		stmt = p.parseBreakStatement()
	case token.NEXT:
		stmt = p.parseNextStatement()
	case token.FALLTHROUGH:
		stmt = p.parseFallThroughStatement()

	case token.MODE:
		stmt = p.parseModeStatement()
	case token.RETURN:
		stmt = p.parseReturnStatement()

	case token.MODULE:
		stmt = p.parseModuleStatement()
	case token.IMPORT:
		stmt = p.parseImportStatement()

	case token.IDENT:
		stmt = p.parseIdentifierStatement()
	case token.VAL, token.VAR:
		stmt = p.parseDeclarationStatement()

	default:
		stmt = p.parseExpressionStatement(nil)
	}

	if eatSemicolon && p.tok.Type == token.SEMICOLON {
		p.advanceToken()
	}

	if p.stopNow {
		return nil
	}

	if p.prevTok.Type != token.SEMICOLON &&
		token.ExpressionContinuationExpected(p.tok.Type) {

		p.addError("Expected EOF, newline, semicolon, or other closing mark to end an expression (not " + token.TypeDescription(p.tok.Type) + ")")
	}

	return stmt
}

func (p *Parser) parseExpressionStatement(ident ast.Node) *ast.ExpressionStatementNode {
	if ident == nil {
		return &ast.ExpressionStatementNode{
			Token:      p.tok,
			Expression: p.parseExpression(precedence_LOWEST)}

	} else {
		return &ast.ExpressionStatementNode{
			Token:      ident.TokenInfo(),
			Expression: p.parseContinuedExpression(ident, precedence_LOWEST)}
	}
}

func (p *Parser) parseIdentifierStatement() *ast.ExpressionStatementNode {
	// could be an unbounded function call, an identifier expression, or an assignment
	var stmt *ast.ExpressionStatementNode

	// could be a single identifier or list, and may include indexed identifiers
	idents := p.parseIdentifierList(true)

	switch p.tok.Type {
	case token.ASSIGN:
		// assignment
		stmt = p.parseAssignmentStatement(idents)

	default:
		if token.IsComboOp(p.tok) {
			// combination assignment
			stmt = p.parseAssignmentStatement(idents)

		} else {
			if len(idents) == 1 {
				// a function call with unbounded arguments?
				call, ok := p.parsePossibleUnboundedCall(idents[0])
				if ok {
					stmt = &ast.ExpressionStatementNode{Token: p.tok, Expression: call}
				} else {
					stmt = p.parseExpressionStatement(idents[0])
				}

			} else {
				p.addError("Invalid identifier list (used without assignment)")
			}
		}
	}

	return stmt
}

func (p *Parser) parseAssignmentStatement(idents []ast.Node) *ast.ExpressionStatementNode {
	return &ast.ExpressionStatementNode{
		Token: p.tok, Expression: p.parseAssignment(idents, true, true, true)}
}

func (p *Parser) parseDeclarationStatement() *ast.ExpressionStatementNode {
	return &ast.ExpressionStatementNode{
		Token: p.tok, Expression: p.parseDeclaration()}
}

func (p *Parser) parseModuleStatement() *ast.ModuleNode {
	stmt := &ast.ModuleNode{Token: p.tok}
	p.advanceToken()

	mod, ok := p.parseWord()
	if ok {
		stmt.Name = mod.Name
	}

	if p.tok.Type == token.ASTERISK {
		if p.tok.CpDiff != 0 {
			p.addError("Asterisk * must immediately follow module keyword")
		}
		stmt.ImpureEffects = true
		p.advanceToken()
	}

	return stmt
}

func (p *Parser) parseImportStatement() *ast.ImportNode {
	stmt := &ast.ImportNode{Token: p.tok}
	p.advanceToken()

	for {
		module := ""
		as := ""

		mod, ok := p.parseWord()
		if ok {
			module = mod.Name

			if p.tok.Type == token.AS {
				p.advanceToken()

				mas, ok := p.parseWord()
				if ok {
					as = mas.Name
					p.checkIdentifierName(as)
					p.addToIdentifiersUsed(as)

				} else {
					p.addError("Expected identifier after as keyword")
					break
				}

			} else {
				p.checkIdentifierName(module)
				p.addToIdentifiersUsed(module)
			}

		} else {
			p.addError("Expected identifier")
			break
		}

		stmt.Modules = append(stmt.Modules, ast.ImportAs{Import: module, As: as})

		if p.tok.Type == token.EOF || p.tok.Type == token.SEMICOLON {
			break

		} else if p.tok.Type == token.COMMA {
			p.advanceToken()
			continue

		} else {
			p.addError("Invalid import list")
			break
		}
	}

	return stmt
}

func (p *Parser) parseReturnStatement() *ast.ReturnNode {
	stmt := &ast.ReturnNode{Token: p.tok}
	p.advanceToken()
	stmt.ReturnValue = p.parseExpression(precedence_LOWEST)
	return stmt
}

func (p *Parser) parseThrowStatement() *ast.ThrowNode {
	stmt := &ast.ThrowNode{Token: p.tok}
	p.advanceToken()

	if p.prefixParseFns[p.tok.Type] != nil {
		stmt.Exception = p.parseExpression(precedence_LOWEST)
	} else {
		if len(p.exceptionVariableStack) > 0 {
			stmt.Exception = p.exceptionVariableStack[len(p.exceptionVariableStack)-1]
		} else {
			p.addError("Cannot throw with implicit value outside of catch block (must throw *something*)")
		}
	}
	return stmt
}

func (p *Parser) parseBreakStatement() *ast.BreakNode {
	stmt := &ast.BreakNode{Token: p.tok}
	p.advanceToken()

	if p.tok.Type == token.ASSIGN {
		p.advanceToken()
		stmt.Value = p.parseExpression(precedence_LOWEST)
	}
	return stmt
}

func (p *Parser) parseNextStatement() *ast.NextNode {
	stmt := &ast.NextNode{Token: p.tok}
	p.advanceToken()
	return stmt
}

func (p *Parser) parseFallThroughStatement() *ast.FallThroughNode {
	stmt := &ast.FallThroughNode{Token: p.tok}
	p.advanceToken()
	return stmt
}

func (p *Parser) parseModeStatement() (mode *ast.ModeNode) {
	mode = &ast.ModeNode{}

	p.advanceToken()
	if p.prevTok.Type != token.MODE {
		p.addError("Expected mode token")
		return
	}

	tok := p.tok
	name := p.tok.Literal
	p.advanceToken() // past name token

	if tok.Type != token.IDENT && tok.Type != token.INT {
		p.addError("Expected mode name after mode token")
		return
	}
	if p.tok.Type != token.ASSIGN {
		p.addError("Expected equals sign after mode name")
		return
	}
	p.advanceToken() // past the equals sign

	modeNumber, ok := modes.ModeNames[name]
	if !ok {
		p.addError(fmt.Sprintf("Unknown mode name %s", name))
		return
	}

	var setting ast.Node
	if p.tok.Type == token.DEFAULT {
		p.advanceToken() // past default token

		// set mode to its default
		defaultSubLexString, ok := modes.DefaultSubLexString[modeNumber]
		if !ok {
			bug("parseModeStatement", fmt.Sprintf("Failed to lookup default for %s", name))
			p.addError(fmt.Sprintf("Failed to lookup default for %s", name))
			return
		}
		tokSlc, err := lexer.LexString(defaultSubLexString, "", p.Modes)
		if err != nil {
			bug("parseModeStatement", err.Error())
			p.addError(err.Error())
			return
		}
		setting, err = ParseExpressionTokens(tokSlc)
		if err != nil {
			bug("parseModeStatement", err.Error())
			p.addError(err.Error())
			return
		}

	} else {
		setting = p.parseExpression(precedence_LOWEST)
	}

	mode = &ast.ModeNode{
		Token:   tok,
		Name:    name,
		Setting: setting,
	}
	return
}

func (p *Parser) parseCatch() ast.Node {
	// Catch is not an expression of itself, but part of a Try/Catch expression.
	// Try section to be filled out later
	catch := &ast.TryCatchNode{Token: p.tok, Try: nil}

	// past the "catch" keyword
	p.advanceToken()

	if p.tok.Type == token.LBRACKET && p.tok.CpDiff == 0 {
		// new syntax (0.13.10) for explicit error variable name
		// catch[.e] ...
		p.advanceToken()

		if p.tok.CpDiff != 0 {
			p.addError("Expected variable after opening square bracket without space, such as catch[.e] ...")
		}

		switch p.tok.Type {
		case token.IDENT:
			// set error to variable named here
			catch.ExceptionVar = p.parseIdentifier()
		default:
			p.addError("Expected variable name after catch and opening square bracket")
		}

		if p.tok.Type == token.RBRACKET && p.tok.CpDiff == 0 {
			p.advanceToken()
		} else {
			p.addError("Expected closing square bracket without space after variable, such as catch[.e] ...")
		}
	}

	if catch.ExceptionVar == nil {
		// not set by new syntax
		switch p.tok.Type {
		case token.IDENT:
			// catch[.e]
			p.addError("To set an explicit exception variable, use square brackets, such as catch[.e]")

		default:
			// implied catch variable _err
			catch.ExceptionVar = ast.NewVariableNode(p.tok, "_err", true)
		}
	}

	p.exceptionVariableStack = append(p.exceptionVariableStack, catch.ExceptionVar)

	simpleCatch := false
	switch p.tok.Type {
	case token.LBRACE:
		catch.Catch = p.parseBlock()
	case token.COLON:
		catch.Catch = p.finishSimpleCatch()
		simpleCatch = true
	default:
		p.addError("Expected opening curly brace for catch body, or colon for simple catch body")
	}

	p.exceptionVariableStack = ast.Pop(p.exceptionVariableStack)

	if !simpleCatch {
		// parse else on catch (action for no exception)
		if p.tok.Type == token.ELSE && !p.tok.NewLinePrecedes {
			p.advanceToken()
			if p.tok.Type == token.LBRACE ||
				p.tok.Type == token.IF { // making it possible to use else if on catch

				if p.tok.Type == token.IF && p.peekTok.Type == token.LPAREN && p.peekTok.CpDiff == 0 {
					p.addError("Invalid use of parentheses on else if of catch block")
				}

				catch.Else, _ = p.parseBlockOrIntoBlock()
			} else {
				p.addError("Expected left brace for else on catch, or an else if")
			}
		}
	}

	return catch
}

func (p *Parser) finishSimpleCatch() ast.Node {
	// parse *simple catch* expression
	if p.tok.NewLinePrecedes {
		p.addError("Colon to indicate simple catch must be on same line as the catch keyword")
		return nil
	}
	p.advanceToken() // past the colon

	if p.tok.NewLinePrecedes {
		// at least for now; might be relaxed later...
		p.addError("Expression after colon must be on same line as the keyword for a simple catch")
		return nil
	}

	catchBlock, _ := p.parseBlockOrIntoBlockWithPotentialAssignment()
	return catchBlock
}

func (p *Parser) parseStatements(
	until []token.Type, first ast.Node,
	includesOpeningToken, passClosingToken bool) (

	nodes []ast.Node, closingtt token.Type) {

	eatSemicolon := !token.InTypeSlice(token.SEMICOLON, until)

	if includesOpeningToken {
		p.advanceToken()
	}

	var errCnt int

	if first != nil {
		nodes = []ast.Node{first}
	}

	for {
		p.tokenAdvanced = false
		errCnt = len(p.Errs)

		if p.checkUpdateStopNow() {
			break
		}

		if p.tok.Type == token.EOF {
			if !token.InTypeSlice(token.EOF, until) {
				p.addError("EOF reached without section closing mark")
				return
			}
			break
		}

		if token.InTypeSlice(p.tok.Type, until) {
			closingtt = p.tok.Type
			if passClosingToken {
				p.advanceToken()
			}
			break
		}

		next := p.parseStatement(eatSemicolon)

		if next == nil {
			//p.addError("Nil node returned in group expression")
			//return
		} else {
			nodes = append(nodes, next)

			if !eatSemicolon && p.tok.Type != token.SEMICOLON {
				p.addError("Expected semicolon")
			}
		}

		if !p.tokenAdvanced && p.tok.Type != token.EOF && errCnt == len(p.Errs) {
			// no new errors added, yet failed to advance
			bug("parseStatements", "Failed to advance the token position")
		}
	}

	var err error
	nodes, err = ast.FixTryCatchInNodeSlice(nodes, true)
	if err != nil {
		p.addError("Error checking for try/catch nodes: " + err.Error())
		return
	}

	return
}
