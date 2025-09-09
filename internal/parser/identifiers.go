// langur/parser/identifiers.go

package parser

import (
	"langur/ast"
	"langur/common"
	"langur/regexp"
	"langur/token"
)

const tokenTypeBetweenVarNameAndType = token.COLON

func (p *Parser) parseIdentifier() ast.Node {
	tt := p.tok.Type
	identifier, ok := p.parseWord()
	if !ok || tt != token.IDENT {
		p.addError("Expected identifier")
		p.advanceToken()
		return identifier
	}

	p.checkIdentifierName(identifier.Name)
	p.addToIdentifiersUsed(identifier.Name)

	return identifier
}

var identifierRegex = regexp.MustCompile(common.IdentifierRegexString)

// a word token that may be an identifier or may be something else
func (p *Parser) parseWord() (*ast.IdentNode, bool) {
	if !identifierRegex.MatchString(p.tok.Literal) {
		return nil, false
	}

	identifier := ast.NewVariableNode(p.tok, p.tok.Literal, false)
	p.advanceToken()

	return identifier, true
}

func (p *Parser) parseType() (ast.Node, int) {
	tt := p.tok.Type
	t, ok := p.parseWord()
	if !ok || tt != token.IDENT {
		p.addError("Expected identifier token")
		p.advanceToken()
		return t, 0
	}
	return t, ast.NodeToLangurTypeCode(t)
}

func (p *Parser) parseIdentifierWithPossibleType() ast.Node {
	ident := p.parseIdentifier()

	if p.tok.Type == tokenTypeBetweenVarNameAndType {
		p.advanceToken()
		t, code := p.parseType()
		if code == 0 {
			p.addError("Expected type after delimiting token after new identifier")
		}
		
		if _, ok := ident.(*ast.IdentNode); ok {
			// set type
			ident.(*ast.IdentNode).Type = t
		}
	}
	
	return ident
}
