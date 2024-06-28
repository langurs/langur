// langur/parser/strings.go

package parser

import (
	"fmt"
	"langur/ast"
	"langur/format"
	"langur/regex"
	"langur/token"
)

func (p *Parser) parseRegex() ast.Node {
	expr := &ast.RegexNode{
		Token: p.tok,
	}

	if p.tok.Type == token.REGEX_RE2 {
		expr.RegexType = regex.RE2

	} else {
		bug("parseRegex", "Unknown regex type")
		p.addError("Unknown regex type")
		p.advanceToken()
		return nil
	}

	expr.Pattern = p.parseStringForLiteral(p.tok.Code&token.CODE_ESC_ALL_INTERPOLATION != 0)

	return expr
}

func (p *Parser) parseDateTime() ast.Node {
	expr := &ast.DateTimeNode{Token: p.tok}
	expr.Pattern = p.parseStringForLiteral(false)
	return expr
}

func (p *Parser) parseDuration() ast.Node {
	expr := &ast.DurationNode{Token: p.tok}
	expr.Pattern = p.parseStringForLiteral(false)
	return expr
}

func (p *Parser) parseString() ast.Node {
	return p.parseStringForLiteral(false)
}

func (p *Parser) parseStringForLiteral(escAllInterpolations bool) ast.Node {
	strNode := &ast.StringNode{
		Token: p.tok,
	}

	if p.tok.Attachments == nil {
		// no interpolation
		strNode.Values = []string{p.tok.Literal}
		strNode.Interpolations = nil

	} else {
		pieces := p.tok.Attachments
		sv, iv := 0, 0
		for i := 0; i < len(pieces); i++ {
			if i%3 == 0 {
				// string
				sv++
				str, ok := pieces[i].(string)
				if !ok {
					bug("parseStringForLiteral", fmt.Sprintf("String interpolation value %d type wrong (%T); expected string", sv, pieces[i]))
					return strNode
				}
				strNode.Values = append(strNode.Values, str)

			} else {
				// token slice (interpolated)
				iv++
				p.parseInterpolation(strNode, pieces, i, iv, escAllInterpolations)
				i++
			}
		}
	}

	p.advanceToken()
	return strNode
}

func (p *Parser) parseInterpolation(
	strNode *ast.StringNode, pieces []interface{}, i, iv int, escInterpolation bool) {

	// token slice (interpolated)
	tokSlc, ok := pieces[i].([]token.Token)
	if !ok {
		bug("parseStringForLiteral", fmt.Sprintf("String interpolation value %d type wrong (%T); expected token slice", iv, pieces[i]))
		return
	}
	i++
	interpModifiers, ok := pieces[i].([]string)
	if !ok {
		bug("parseStringForLiteral", fmt.Sprintf("String interpolation modifiers %d type wrong (%T); expected string slice", iv, pieces[i]))
		return
	}

	if len(interpModifiers) == 0 || interpModifiers[len(interpModifiers)-1] != format.MODSTRING_ESCAPE {
		// does not end with escape modifier
		switch {
		case escInterpolation:
			interpModifiers = append(interpModifiers, format.MODSTRING_ESCAPE)
		}
	}

	node, err := ParseExpressionTokens(tokSlc)
	if err != nil {
		p.addError(fmt.Sprintf("String interpolation value %d failed parsing (%s)", iv, err))
		p.advanceToken()
		return
	}
	strNode.Interpolations = append(strNode.Interpolations,
		&ast.InterpolatedNode{
			Token:     p.tok,
			Value:     node,
			Modifiers: interpModifiers,
		},
	)
}
