// langur/parser/parser.go

package parser

// NOTE: Each terminal parse function is responsible for advancing the token position.

import (
	"bytes"
	"fmt"
	"langur/ast"
	"langur/lexer"
	"langur/modes"
	"langur/object"
	"langur/token"
)

func bug(fnName, s string) {
	panic("Parser bug: " + s)
}

const includeTracing = true

func (p *Parser) addError(err string) {
	if includeTracing {
		p.Errs = append(p.Errs, fmt.Errorf("[%s] %s\n\n%s", p.tok.Where.String(), err, p.tok.Where.Trace(p.lex.Source)))
	} else {
		p.Errs = append(p.Errs, fmt.Errorf("[%s] %s", p.tok.Where.String(), err))
	}
}

type (
	prefixParseFn  func() ast.Node
	infixParseFn   func(ast.Node) ast.Node
	postfixParseFn func(ast.Node) ast.Node
)

// 0 to disable
const STOP_AT_ERR_COUNT = 3

type Parser struct {
	lex       *lexer.Lexer
	tokSlc    []token.Token
	tokSlcPos int

	tok     token.Token
	prevTok token.Token
	peekTok token.Token
	Errs    []error

	prefixParseFns  map[token.Type]prefixParseFn
	infixParseFns   map[token.Type]infixParseFn
	postfixParseFns map[token.Type]postfixParseFn

	stopNow       bool
	tokenAdvanced bool

	exceptionVariableStack []ast.Node // throw with implicit value
	forLoopVariableStack   []ast.Node // set/break for loop with value

	identifiersUsed []string

	Modes *modes.CompileModes

	contexts []context
}

func New(lex *lexer.Lexer, m *modes.CompileModes) (p *Parser) {
	defer func() {
		if panik := recover(); panik != nil {
			p.addError(object.PanicToError(panik).Error())
		}
	}()

	p = &Parser{lex: lex, Modes: m}
	p.init()
	return p
}

func ErrorSliceToString(slc []error, precedeEach, delim string) string {
	var b bytes.Buffer

	if len(slc) != 0 {
		for i, msg := range slc {
			b.WriteString(precedeEach)
			b.WriteString(msg.Error())
			if i < len(slc)-1 {
				b.WriteString(delim)
			}
		}
	}

	return b.String()
}

func (p *Parser) init() {
	// read to tokens, so tok and peekTok are both set
	p.advanceToken()
	p.advanceToken()

	p.setParseFunctionMaps()
}

func (p *Parser) setParseFunctionMaps() {
	// map expression parsing functions (not for statements)
	p.prefixParseFns = map[token.Type]prefixParseFn{
		token.IDENT: p.parseIdentifier,

		token.INT:   p.parseNumber,
		token.FLOAT: p.parseNumber,
		token.MINUS: p.parsePrefixExpression,

		token.SWITCH: p.parseSwitchExpression,
		token.IF:     p.parseIfExpression,
		token.TRUE:   p.parseBoolean,
		token.FALSE:  p.parseBoolean,
		token.NULL:   p.parseNull,

		token.FUNCTION: p.parseFunction,
		token.LPAREN:   p.parseParenthesizedExpression,

		token.FOR:   p.parseForLoop,
		token.WHILE: p.parseWhileLoop,

		token.REGEX_RE2: p.parseRegex,
		token.DATETIME:  p.parseDateTime,
		token.DURATION:  p.parseDuration,
		token.STRING:    p.parseString,
		token.LBRACKET:  p.parseList,
		token.LBRACE:    p.parseLBrace,

		token.NONE: p.parseNone,
		token.NOT:  p.parsePrefixExpression,

		// The following are for parsing incomplete comparisons for switch expressions.
		token.IS:               p.parseInfixNilLeftExpression,
		token.EQUAL:            p.parseInfixNilLeftExpression,
		token.NOT_EQUAL:        p.parseInfixNilLeftExpression,
		token.FORWARD:          p.parseInfixNilLeftExpression,
		token.LESS_THAN:        p.parseInfixNilLeftExpression,
		token.GREATER_THAN:     p.parseInfixNilLeftExpression,
		token.LT_OR_EQUAL:      p.parseInfixNilLeftExpression,
		token.GT_OR_EQUAL:      p.parseInfixNilLeftExpression,
		token.DIVISIBLE_BY:     p.parseInfixNilLeftExpression,
		token.NOT_DIVISIBLE_BY: p.parseInfixNilLeftExpression,
	}

	// postfix functions
	p.postfixParseFns = map[token.Type]postfixParseFn{
		token.EXPANSION: p.parsePostfixExpression,
	}

	// infix functions
	p.infixParseFns = map[token.Type]infixParseFn{
		token.APPEND:      p.parseInfixExpression,
		token.PLUS:        p.parseInfixExpression,
		token.MINUS:       p.parseInfixExpression,
		token.ASTERISK:    p.parseInfixExpression,
		token.SLASH:       p.parseInfixExpression,
		token.BACKSLASH:   p.parseInfixExpression,
		token.DOUBLESLASH: p.parseInfixExpression,
		token.REMAINDER:   p.parseInfixExpression,
		token.MODULUS:     p.parseInfixExpression,
		token.POWER:       p.parseInfixExpression,
		token.ROOT:        p.parseInfixExpression,

		// token.NOT as infix for not in/not of; checked elsewhere that only used as such for infix (normally prefix)
		token.NOT: p.parseInfixExpression,
		token.IN:  p.parseInfixExpression,
		token.OF:  p.parseInfixExpression,

		token.IS:               p.parseInfixExpression,
		token.EQUAL:            p.parseInfixExpression,
		token.NOT_EQUAL:        p.parseInfixExpression,
		token.FORWARD:          p.parseInfixExpression,
		token.LESS_THAN:        p.parseInfixExpression,
		token.GREATER_THAN:     p.parseInfixExpression,
		token.LT_OR_EQUAL:      p.parseInfixExpression,
		token.GT_OR_EQUAL:      p.parseInfixExpression,
		token.DIVISIBLE_BY:     p.parseInfixExpression,
		token.NOT_DIVISIBLE_BY: p.parseInfixExpression,

		token.RANGE: p.parseInfixExpression,

		token.AND:  p.parseInfixExpression,
		token.OR:   p.parseInfixExpression,
		token.NAND: p.parseInfixExpression,
		token.NOR:  p.parseInfixExpression,
		token.XOR:  p.parseInfixExpression,
		token.NXOR: p.parseInfixExpression,

		token.LBRACKET: p.parseIndexExpression,
		token.LPAREN:   p.parseParenthesizedCallExpression,
	}
}

func ParseExpressionTokens(tokSlc []token.Token) (node ast.Node, err error) {
	defer func() {
		if panik := recover(); panik != nil {
			err = object.PanicToError(panik)
		}
	}()

	p := &Parser{}
	p.tokSlc = tokSlc
	p.init()

	node = p.parseExpression(precedence_LOWEST)
	if len(p.Errs) == 0 && p.tokSlcPos < len(p.tokSlc) {
		p.addError("Failed to parse all tokens as single expression")
	}

	errs := ErrorSliceToString(p.Errs, "", "; ")

	if errs != "" {
		err = fmt.Errorf(errs)
	}

	return
}

func (p *Parser) addToIdentifiersUsed(name string) {
	// for the moment, only adding system variable names; could allow others if useful
	if name[0] != '_' {
		return
	}
	for i := range p.identifiersUsed {
		if p.identifiersUsed[i] == name {
			return
		}
	}
	p.identifiersUsed = append(p.identifiersUsed, name)
}

func (p *Parser) checkIdentifierName(name string) bool {
	if len(name) > 1 && name[0] == '_' && name[len(name)-1] == '_' {
		// disallow user reading of internal system variable name (beginning and ending with underscore)
		p.addError(fmt.Sprintf("Cannot access identifier name %s", name))
		return false
	}
	return true
}

// THE STARTING POINT
func (p *Parser) ParseProgram() (program *ast.Program, err error) {
	defer func() {
		if panik := recover(); panik != nil {
			err = object.PanicToError(panik)
		}
	}()

	program = &ast.Program{Token: p.tok}
	program.Statements, _ = p.parseStatements([]token.Type{token.EOF}, nil, false, false)

	program.VarNamesUsed = p.identifiersUsed

	return
}

func (p *Parser) advanceToken() {
	p.prevTok = p.tok
	p.tok = p.peekTok

	if p.lex == nil {
		// using a token slice directly
		if len(p.tokSlc) == 0 {
			p.addError("No tokens")
		}
		if p.tokSlcPos < len(p.tokSlc) {
			p.peekTok = p.tokSlc[p.tokSlcPos]
			p.tokSlcPos++
		} else {
			p.peekTok = token.Token{Type: token.EOF}
		}

	} else {
		// using a lexer directly
		var err error
		p.peekTok, err = p.lex.NextToken()
		if err != nil {
			p.peekTok.AddTokenErr(err)
		}
	}
	p.tokenAdvanced = true

	// add lexer errors to the parser from the tokens
	if len(p.peekTok.Errs) > 0 {
		p.addError("Lex Errors: [" + p.peekTok.Errors("; ") + "]")
	}
}

func (p *Parser) checkUpdateStopNow() bool {
	if STOP_AT_ERR_COUNT > 0 && len(p.Errs) >= STOP_AT_ERR_COUNT {
		if !p.stopNow {
			p.addError(fmt.Sprintf("...Parsing stopping after %d errors", len(p.Errs)))
			p.stopNow = true
		}
	}
	return p.stopNow
}
