// langur/parser/precedence.go

package parser

import (
	"langur/token"
)

type precedence int

const (
	// NOTE: Order here is important. This sets operator precedence.

	precedence_LOWEST precedence = iota

	precedence_ASSIGNMENT // .x = 123

	// C places bitwise xor between bitwise and and or,...
	// so we follow that precedence for logical xor.
	precedence_LOGICAL_OR          // or, nor
	precedence_LOGICAL_EQUIVALENCE // xor, nxor
	precedence_LOGICAL_AND         // and, nand

	precedence_EQUALITY // ==  !=

	// logical negation operator below Boolean comparison
	// (can't do mathematical comparison after logical negation)
	// higher precedence than equality
	precedence_LOGICAL_NEGATION

	precedence_BOOLEAN // < > <= >= div ndiv
	// is/is not: should have higher precedence than logical ops
	// in/of/not in/not of

	precedence_FORWARD // ->

	// append strings, lists, or hashes
	precedence_APPEND

	// range precedence between mathematical and equality operators
	precedence_RANGE

	precedence_SUM     // + -
	precedence_PRODUCT // x / \ // rem mod

	// NOTE(0.12): changed order of the following 2
	precedence_PREFIX    // -X
	precedence_EXPONENTS // ^  ^/

	precedence_AFTER // function call () or indexing []
)

var infixPrecedences = map[token.Type]precedence{
	token.ASSIGN:  precedence_ASSIGNMENT,
	token.FORWARD: precedence_FORWARD,

	token.EQUAL:     precedence_EQUALITY,
	token.NOT_EQUAL: precedence_EQUALITY,

	token.IS:               precedence_BOOLEAN,
	token.LESS_THAN:        precedence_BOOLEAN,
	token.GREATER_THAN:     precedence_BOOLEAN,
	token.LT_OR_EQUAL:      precedence_BOOLEAN,
	token.GT_OR_EQUAL:      precedence_BOOLEAN,
	token.DIVISIBLE_BY:     precedence_BOOLEAN,
	token.NOT_DIVISIBLE_BY: precedence_BOOLEAN,

	token.IN: precedence_BOOLEAN,
	token.OF: precedence_BOOLEAN,

	// for not in/not of
	// should be same precedence as IN and OF tokens
	token.NOT: precedence_BOOLEAN,

	token.RANGE:  precedence_RANGE,
	token.APPEND: precedence_APPEND,

	token.PLUS:        precedence_SUM,
	token.MINUS:       precedence_SUM,
	token.TIMES:       precedence_PRODUCT,
	token.DIVIDE:      precedence_PRODUCT,
	token.DIVIDEINT:   precedence_PRODUCT,
	token.DIVIDEFLOOR: precedence_PRODUCT,
	token.REMAINDER:   precedence_PRODUCT,
	token.MODULUS:     precedence_PRODUCT,
	token.EXPONENT:    precedence_EXPONENTS,
	token.ROOT:        precedence_EXPONENTS,

	token.AND:  precedence_LOGICAL_AND,
	token.OR:   precedence_LOGICAL_OR,
	token.NAND: precedence_LOGICAL_AND,
	token.NOR:  precedence_LOGICAL_OR,
	token.XOR:  precedence_LOGICAL_EQUIVALENCE,
	token.NXOR: precedence_LOGICAL_EQUIVALENCE,

	token.LPAREN:   precedence_AFTER,
	token.LBRACKET: precedence_AFTER,
}

func getInfixPrecedence(tt token.Type) precedence {
	if p, ok := infixPrecedences[tt]; ok {
		return p
	}
	return precedence_LOWEST
}

var prefixPrecedences = map[token.Type]precedence{
	token.NOT: precedence_LOGICAL_NEGATION,
}

func getPrefixPrecedence(tt token.Type) precedence {
	if p, ok := prefixPrecedences[tt]; ok {
		return p
	}
	return precedence_PREFIX
}
