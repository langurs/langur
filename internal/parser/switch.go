// langur/parser/switch.go

package parser

import (
	"fmt"
	"langur/ast"
	"langur/token"
)

func (p *Parser) possiblyParseAlternateInfixOp(tok token.Token) token.Token {
	if p.tok.Type == token.LBRACKET && p.tok.CpDiff == 0 {
		// such as switch[and] ...
		p.advanceToken()
		if token.IsInfixOp(p.tok.Type) && p.tok.CpDiff == 0 {
			tok = p.parseInfixOperator()

		} else {
			p.addError("Expected infix operator, without space, such as switch[and] ...")
		}

		if p.tok.Type == token.RBRACKET && p.tok.CpDiff == 0 {
			p.advanceToken()
		} else {
			p.addError("Expected closing square bracket without space after infix operator, such as switch[and] ...")
		}
		if p.tok.CpDiff == 0 {
			p.addError("Expected space after closing square bracket for switch alternate operator, such as switch[and] ...")
		}
	}

	return tok
}

func (p *Parser) parseSwitchExpression() ast.Node {
	var cd ast.CaseDo

	sw := &ast.SwitchNode{Token: p.tok}

	// skip "switch" token
	p.advanceToken()
	shortenedForm := p.tok.Type == token.LPAREN && p.tok.CpDiff == 0

	// use default *or* operator or parse logical operator (used between case conditions)

	sw.DefaultLogicalOp = p.possiblyParseAlternateInfixOp(
		token.Token{Literal: "or", Type: token.OR})

	// variables or other test expressions to compare against
	// may be partial expressions (1 side nil)
	var exprs []ast.Node
	if shortenedForm {
		// skip opening parenthesis
		p.advanceToken()

		exprs, _ = p.parseExpressionList([]token.Type{token.SEMICOLON}, token.COMMA, false, false, false)
		// NOT past the opening semicolon (used to delimit cases)

	} else {
		// push context to differentiate between starting a hash literal after an infix operator, and the end of test expression list
		p.pushContext(context_expression_switch_test)

		exprs, _ = p.parseExpressionList([]token.Type{token.LBRACE}, token.COMMA, false, true, false)
		// now past the opening left brace {

		p.popContext()
	}

	var declarationInternalNumber = 1
	possiblyConvertTestExprToSysAssignment := func(node ast.Node) ast.Node {
		// to be extracted at the end if converted to declaration...
		if ast.ShouldConvertSwitchTestToSystemAssignment(node) {
			id := ast.NewVariableNode(
				node.TokenInfo(),
				fmt.Sprintf("_test%d_", declarationInternalNumber),
				true,
			)
			node = ast.MakeDeclarationAssignmentExpression(id, node, true, false)
			declarationInternalNumber++
		}
		return node
	}

	// check if partial test expressions used for comparison such as switch != x { ... }
	for _, e := range exprs {
		expr := ast.PartialExpr{}

		var done bool
		switch e := e.(type) {
		case *ast.InfixExpressionNode:
			if e.Left == nil {
				expr.Expr = possiblyConvertTestExprToSysAssignment(e.Right)
				expr.Op = e.Operator
				done = true

			} else if e.Right == nil {
				expr.Expr = possiblyConvertTestExprToSysAssignment(e.Left)
				expr.Op = e.Operator
				expr.LeftOperand = true
				done = true
			}
		}

		if !done {
			expr.Expr = possiblyConvertTestExprToSysAssignment(e)
			expr.Op = token.DefaultCompOp
			expr.LeftOperand = true
		}

		sw.Expressions = append(sw.Expressions, expr)
	}

	// now parse cases...
	parseOtherConditions := func() {
		nvc, _ := p.parseExpressionList([]token.Type{token.COLON}, token.COMMA, false, true, false)

		if len(nvc) == 0 {
			p.addError("Expected non-variable test expressions for switch case condition")
		}

		for _, c := range nvc {
			cd.OtherConditions = append(cd.OtherConditions, c)
		}
	}

	for {
		if p.tok.Type == token.EOF {
			p.addError("EOF reached without closing switch expression")
			return sw
		}

		cd.MatchConditions = nil
		cd.OtherConditions = nil

		if !shortenedForm && p.tok.Type == token.CASE ||
			shortenedForm && p.tok.Type == token.SEMICOLON {

			// skip the case or semicolon token
			p.advanceToken()

			// logical operator used between matches in a single case ...
			// case[and] ...
			cd.MatchLogicalOp = p.possiblyParseAlternateInfixOp(sw.DefaultLogicalOp)
			cd.OtherLogicalOp = sw.DefaultLogicalOp

			// parse test conditions
			if len(sw.Expressions) > 0 {
				var end token.Type

				if shortenedForm {
					var conditionsOrDefault []ast.Node
					// don't know if it's a set of test expressions or a default until we parse
					conditionsOrDefault, end = p.parseExpressionList(
						[]token.Type{token.COLON, token.SEMICOLON, token.RPAREN}, token.COMMA, false, true, false)

					if end == token.RPAREN {
						// is the default (no condition)
						if len(conditionsOrDefault) != 1 {
							p.addError("Use 1 expression for default in shortened form switch")
							return sw
						}

						cd.Do = ast.BlockOrAsBlock(conditionsOrDefault[0])
						sw.CasesAndActions = append(sw.CasesAndActions, cd)
						break

					} else if end == token.SEMICOLON {
						p.addError("Cannot use alternate test conditions for shortened form switch")
						return sw

					} else {
						cd.MatchConditions = conditionsOrDefault
					}

				} else {
					// not shortened form
					cd.MatchConditions, end = p.parseExpressionList(
						[]token.Type{token.COLON, token.SEMICOLON}, token.COMMA, false, true, false)
				}

				if len(cd.MatchConditions) > len(sw.Expressions) && len(sw.Expressions) > 1 {
					p.addError("More test conditions than test expressions specified (only allowed with 1 test expression)")
					return sw
				}

				switch end {
				case token.SEMICOLON:
					// non-variable conditions
					// logical operator used between matches and other conditions in a single case
					// (same as match operator by default)
					if token.IsInfixLogicalOp(p.tok.Type) {
						cd.OtherLogicalOp = p.parseInfixOperator()
					}

					parseOtherConditions()

				case token.COLON:
					// if closed with colon and nil right (partial) expression, ...
					// ... check last condition for space between operator and colon ...
					// ... (clarity and future proofing in case of operators using the colon).
					condInfix, condIsInfix := cd.MatchConditions[len(cd.MatchConditions)-1].(*ast.InfixExpressionNode)
					if condIsInfix && condInfix.Right == nil &&
						p.prevTok.CpDiff == 0 {
						// the previous token being the colon
						p.addError("Space required between operator and colon ending test condition (in case)")
					}
				}

			} else if shortenedForm {
				p.addError("Cannot use non-variable switch for shortened form switch")
				return sw

			} else {
				// no variables
				parseOtherConditions()
			}

			// parse action
			if shortenedForm {
				cd.Do = ast.BlockOrAsBlock(p.parseExpression(precedence_LOWEST))

				// disallowing explicit fallthrough on shortened form
				_, isFallThrough := cd.Do.(*ast.BlockNode).Statements[len(cd.Do.(*ast.BlockNode).Statements)-1].(*ast.FallThroughNode)
				if isFallThrough {
					p.addError("Explicit fallthrough not allowed in shortened form switch expression")
					return sw
				}

			} else {
				statements, _ := p.parseStatements(
					[]token.Type{token.CASE, token.DEFAULT, token.RBRACE},
					nil, false, false)

				cd.Do = ast.ListToBlock(statements)
			}

			sw.CasesAndActions = append(sw.CasesAndActions, cd)

		} else if !shortenedForm && p.tok.Type == token.DEFAULT {
			p.advanceToken()
			if p.tok.Type != token.COLON {
				p.addError("Expected colon after default in switch expression")
				return sw
			}
			p.advanceToken()

			statements, _ := p.parseStatements([]token.Type{token.RBRACE}, nil, false, true)
			cd.Do = ast.ListToBlock(statements)

			sw.CasesAndActions = append(sw.CasesAndActions, cd)
			break

		} else if !shortenedForm && p.tok.Type == token.RBRACE ||
			shortenedForm && p.tok.Type == token.RPAREN {

			p.advanceToken()
			break

		} else if shortenedForm {
			p.addError("Expected semicolon or closing parenthesis in shortened form switch expression")
			return sw

		} else {
			p.addError("Expected case or default in switch expression")
			return sw
		}
	}

	// passing switch by ref. so it can be changed and receiving any extracted nodes
	declarationsAndAssignments, err := ast.ExtractDeclarationsAndAssignmentsForSwitchTests(sw)
	if err != nil {
		p.addError(err.Error())
		return sw
	}

	// convert to if node (compiler never sees switch node)
	ifnode, err := ast.ConvertSwitchNodeToIfNode(sw, token.DefaultCompOp)
	if err != nil {
		p.addError(err.Error())
		return ifnode
	}

	if declarationsAndAssignments != nil {
		// move test expression declarations and assignments to front of a new scope block containing the switch expression
		block := &ast.BlockNode{HasScope: true}
		for _, decl := range declarationsAndAssignments {
			block.Statements = append(block.Statements, decl)
		}
		block.Statements = append(block.Statements, ifnode)
		return block
	}

	return ifnode
}
