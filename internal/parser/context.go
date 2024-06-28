// langur/parser/context.go

package parser

type context int

const (
	context_statement context = iota
	context_switch_test
)

func (p *Parser) pushContext(c context) {
	p.contexts = append(p.contexts, c)
}
func (p *Parser) popContext() {
	p.contexts = p.contexts[:len(p.contexts)-1]
}
func (p *Parser) peekContext() context {
	return p.contexts[len(p.contexts)-1]
}
