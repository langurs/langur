// langur/parser/assignment.go

package parser

import (
	"langur/ast"
	"langur/token"
)

func (p *Parser) parseDeclaration() ast.Node {
	mutable := false
	public := false
	// tok := p.tok

	if p.tok.Type == token.PUBLIC {
		public = true
		p.advanceToken()
	}

	switch p.tok.Type {
	case token.VAL:
		p.advanceToken() // past the val token
	case token.VAR:
		mutable = true
		p.advanceToken() // past the var token
	default:
		p.addError("Expected val or var token")
	}

	return p.parseIdentifiersWithPotentialAssignments(
		false, false, !mutable, false, true, mutable, true, mutable, public)
}

// assignmentRequired: identifier alone allowed?
func (p *Parser) parseIdentifiersWithPotentialAssignments(
	mayIncludeIndexedAssignment, mayIncludeIndexedNonAssignment,
	assignmentRequired, mayBeComboOp,
	mayBeMultiAssign, convertIdentsListToAssignment,
	convertAssignmentToDeclaration, defaultDeclarationsMutable, public bool) ast.Node {

	var idents []ast.Node
	tok := p.tok

	if mayBeMultiAssign {
		idents = p.parseIdentifierList(mayIncludeIndexedAssignment || mayIncludeIndexedNonAssignment)

	} else {
		ident := p.parseIdentifier()

		// indexed?
		if p.tok.Type == token.LBRACKET {
			if mayIncludeIndexedAssignment || mayIncludeIndexedNonAssignment {
				ident = p.parseIndices(ident)
			} else if p.tok.CpDiff == 0 {
				p.addError("Unexpected indexing on identifier")
			}
		}

		idents = []ast.Node{ident}
	}

	if p.tok.Type == token.ASSIGN || token.IsComboOp(p.tok) {
		assign := p.parseAssignment(idents, mayIncludeIndexedAssignment, mayBeComboOp, mayBeMultiAssign)

		if convertAssignmentToDeclaration {
			decl, err := ast.AssignmentToDeclaration(assign, defaultDeclarationsMutable, public)
			if err != nil {
				p.addError(err.Error())
			}
			return decl
		}
		return assign

	} else if convertIdentsListToAssignment {
		values := make([]ast.Node, len(idents))
		for i := range values {
			values[i] = ast.NoValue
		}
		assign := &ast.AssignmentNode{Token: tok, Identifiers: idents, Values: values}

		if convertAssignmentToDeclaration {
			decl, err := ast.AssignmentToDeclaration(assign, defaultDeclarationsMutable, public)
			if err != nil {
				p.addError(err.Error())
			}
			return decl
		}
		return assign

	} else if assignmentRequired {
		p.addError("Expected assignment")
		return nil

	} else if len(idents) != 1 {
		// identifier list, but no assignment
		p.addError("Invalid identifier list (no assignment)")
		return nil
	}

	return idents[0]
}

func (p *Parser) parseAssignment(idents []ast.Node, mayIncludeIndices, mayBeComboOp, mayBeMultiAssign bool) ast.Node {
	if idents == nil {
		idents = p.parseIdentifierList(mayIncludeIndices)
	}

	combo := false

	switch p.tok.Type {
	case token.ASSIGN:
		p.advanceToken()

	default:
		if token.IsComboOp(p.tok) {
			if mayBeComboOp {
				combo = true
			} else {
				p.addError("Invalid combination operation")
			}

		} else {
			p.addError("Expected comma or assignment operator")
		}
	}

	if len(idents) == 1 {
		if combo {
			return p.parseCombinationAssignment(idents[0])
		}

		expr := &ast.AssignmentNode{
			Token:       p.tok,
			Identifiers: idents,
			Values:      []ast.Node{p.parseExpression(precedence_LOWEST)},
		}
		return p.fixNamesInFunctionAssignments(expr)

	} else {
		if combo {
			p.addError("Cannot use combination operator with multi-variable assignment")
		}

		if mayBeMultiAssign {
			return p.parseMultiVariableAssignmentValues(p.tok, idents)
		}
		p.addError("Unexpected identifier list")
		return nil
	}
}

func (p *Parser) fixNamesInFunctionAssignments(node *ast.AssignmentNode) ast.Node {
	// multiple idents, 1 value: nothing fixable/to fix
	// multiple values, 1 ident: nothing fixable/to fix
	// number of values matches number of idents, ...
	if len(node.Identifiers) == len(node.Values) {
		for i := range node.Values {
			if fn, ok := node.Values[i].(*ast.FunctionNode); ok {
				fn.Name = node.Identifiers[i].(*ast.IdentNode).Name
				node.Values[i] = fn
			}
		}
	}
	return node
}

func (p *Parser) parseMultiVariableAssignmentValues(tok token.Token, idents []ast.Node) ast.Node {
	values, _ := p.parseExpressionList(token.EndUnboundedAssignmentExprList, token.COMMA, false, false, false)
	if len(values) != 1 && len(values) != len(idents) {
		p.addError("Identifier/value count mismatch in multi-variable declaration assignment")
		return nil
	}

	assign := &ast.AssignmentNode{
		Token:       tok,
		Identifiers: idents,
		Values:      values,
	}

	return p.fixNamesInFunctionAssignments(assign)
}

func (p *Parser) parseCombinationAssignment(left ast.Node) ast.Node {
	op := p.parseInfixOperator()
	right := p.parseExpression(precedence_ASSIGNMENT)

	expr := &ast.InfixExpressionNode{Left: left, Operator: op, Right: right}

	// remove index alternate from left expression; remains in infix only, not in node to assign to
	if L, isIndexOp := left.(*ast.IndexNode); isIndexOp {
		if L.Alternate != nil {
			left = left.Copy()
			left.(*ast.IndexNode).Alternate = nil
		}
	}

	return ast.MakeAssignmentExpression(left, expr, false)
}

func (p *Parser) parseIdentifierList(mayIncludeIndices bool) (idents []ast.Node) {
	cnt := 0
	line := p.tok.Where.Line
	includesExpansion := false

	parseIdent := func() ast.Node {
		ident := p.parseIdentifier()
		if p.tok.Type == token.LBRACKET && p.tok.CpDiff == 0 {
			if mayIncludeIndices {
				ident = p.parseIndices(ident)
			} else {
				p.addError("Unexpected indexing in identifier list")
			}
		}
		cnt++
		return ident
	}

	for {
		if includesExpansion {
			p.addError("Expansion not on last identifier in list")
		}

		if p.tok.Type == token.EXPANSION {
			// position of expansion token changed (0.20)
			p.addError("Expansion must be specified after the variable name, not before")
			p.advanceToken()
		}

		switch p.tok.Type {
		case token.IDENT:
			idents = append(idents, parseIdent())

			if p.tok.Type == token.EXPANSION {
				includesExpansion = true
				expansion := p.parseExpansionPartial()
				if expansion.Limits != nil {
					// cannot use limits at present on variable expansion
					p.addError("Cannot set limits on variable expansion")
				}
				expansion.Continuation = idents[len(idents)-1]
				idents[len(idents)-1] = expansion
			}

		case token.NONE:
			idents = append(idents, p.parseNone())

		default:
			p.addError("Expected variables (or no-ops) only in identifier list")
		}

		if p.tok.Type == token.COMMA {
			// list continues
			p.advanceToken()
		} else {
			// end of list
			break
		}
	}

	if cnt == 0 {
		p.addError("Expected identifier(s) in list (cannot be all no-op)")
	}
	if p.tok.Where.Line != line {
		p.addError("Unexpected new line(s) in identifier list")
	}

	return
}

func (p *Parser) parseExpressionWithPotentialAssignment() ast.Node {
	if p.tok.Type == token.IDENT {
		expr := p.parseIdentifiersWithPotentialAssignments(
			true, true, false, true, true, false, false, false, false)

		if _, ok := expr.(*ast.IdentNode); ok {
			expr = p.parseContinuedExpression(expr, precedence_LOWEST)
		}
		return expr
	}
	return p.parseExpression(precedence_LOWEST)
}
