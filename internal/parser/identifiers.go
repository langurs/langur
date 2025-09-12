// langur/parser/identifiers.go

package parser

import (
	"langur/object"
	"langur/ast"
	"langur/common"
	"langur/regexp"
	"langur/token"
)

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

func (p *Parser) parseType() (ast.Node, object.ObjectType) {
	tt := p.tok.Type
	t, ok := p.parseWord()
	if !ok || tt != token.IDENT {
		p.addError("Expected identifier token")
		p.advanceToken()
		return t, 0
	}
	return t, ast.NodeToLangurType(t)
}

func (p *Parser) checkParseType() (ast.Node, object.ObjectType) {
	_, ok := object.TypeNameToType[p.tok.Literal]
	if ok {
		return p.parseType()
	}
	return nil, 0
}
