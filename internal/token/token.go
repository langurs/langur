// langur/token/token.go

package token

import (
	"langur/common"
	"strings"
)

type Type string

// NOTE: If adding token types, we must update token rules accordingly (see rules.go).
// We also need to see if it needs an entry in the precedences map for the parser (mostly infix operators, though not all).

const (
	INVALID Type = ""
	EOF          = "EOF"

	DOT = "DOT"

	MODULE = "MODULE"
	IMPORT = "IMPORT"
	AS     = "AS"

	IDENT    = "IDENT"
	INT      = "INT"
	FLOAT    = "FLOAT"
	NONE     = "NONE"
	DATETIME = "DATETIME"
	DURATION = "DURATION"

	MODE = "MODE"

	CATCH = "CATCH"
	THROW = "THROW"

	STRING    = "STRING"
	REGEX_RE2 = "REGEX_RE2"

	VAR    = "VAR"
	VAL    = "VAL"
	ASSIGN = "ASSIGN"

	APPEND = "APPEND"

	PLUS        = "PLUS"
	MINUS       = "MINUS"
	ASTERISK    = "ASTERISK"
	SLASH       = "SLASH"
	BACKSLASH   = "BACKSLASH"
	DOUBLESLASH = "DOUBLESLASH"
	REMAINDER   = "REMAINDER"
	MODULUS     = "MODULUS"
	POWER       = "POWER"
	ROOT        = "ROOT"

	LESS_THAN    = "LESS_THAN"
	GREATER_THAN = "GREATER_THAN"
	LT_OR_EQUAL  = "LT_OR_EQUAL"
	GT_OR_EQUAL  = "GT_OR_EQUAL"

	EQUAL     = "EQUAL"
	NOT_EQUAL = "NOT_EQUAL"

	FORWARD = "FORWARD"

	DIVISIBLE_BY     = "DIVISIBLE_BY"
	NOT_DIVISIBLE_BY = "NOT_DIVISIBLE_BY"

	RANGE = "RANGE"
	COLON = "COLON"

	COMMA     = "COMMA"
	SEMICOLON = "SEMICOLON"

	LPAREN   = "LPAREN"
	RPAREN   = "RPAREN"
	LBRACE   = "LBRACE"
	RBRACE   = "RBRACE"
	LBRACKET = "LBRACKET"
	RBRACKET = "RBRACKET"

	FUNCTION  = "FUNCTION"
	EXPANSION = "EXPANSION"
	RETURN    = "RETURN"

	WHILE = "WHILE"
	FOR   = "FOR"
	IN    = "IN"
	OF    = "OF"
	BREAK = "BREAK"
	NEXT  = "NEXT"

	SWITCH      = "SWITCH"
	CASE        = "CASE"
	DEFAULT     = "DEFAULT"
	FALLTHROUGH = "FALLTHROUGH"

	IF   = "IF"
	ELSE = "ELSE"

	TRUE  = "TRUE"
	FALSE = "FALSE"
	NULL  = "NULL"

	AND  = "AND"
	OR   = "OR"
	NAND = "NAND"
	NOR  = "NOR"

	XOR  = "XOR"
	NXOR = "NXOR"

	NOT = "NOT"

	IS = "IS"

	QUOTE_STRING   = "QUOTE_STRING"
	FREE_WORD_LIST = "FREE_WORD_LIST"

	RESERVED = "RESERVED"

	ZLS = "ZLS"
)

func IsComboOp(tok Token) bool {
	return 0 != tok.Code&CODE_COMBINATION_ASSIGNMENT_OPERATOR
}

func InterpretEscapeSequences(lit string) bool {
	return lit == strings.ToLower(lit)
}

func NegatedLiteral(lit string) bool {
	switch lit {
	case common.IsNotLiteral, common.NotInLiteral, common.NotOfLiteral:
		return true
	}
	return false
}

var Keywords = map[string]Type{
	// not including built-in functions (except type names)
	// All other keywords should be listed here.

	common.TrueTokenLiteral:  TRUE,
	common.FalseTokenLiteral: FALSE,
	common.NullTokenLiteral:  NULL,

	"module": MODULE,
	"import": IMPORT,
	"as":     AS,

	"is": IS,

	"mode": MODE,

	"var": VAR,
	"val": VAL,

	"catch": CATCH,
	"throw": THROW,

	common.FunctionTokenLiteral: FUNCTION,
	"return":                    RETURN,

	"while": WHILE,
	"for":   FOR,
	"in":    IN,
	"of":    OF,
	"break": BREAK,
	"next":  NEXT,

	"mod":  MODULUS,
	"rem":  REMAINDER,
	"div":  DIVISIBLE_BY,
	"ndiv": NOT_DIVISIBLE_BY,

	"switch":      SWITCH,
	"case":        CASE,
	"default":     DEFAULT,
	"fallthrough": FALLTHROUGH,

	"if":   IF,
	"else": ELSE,

	"and":  AND,
	"or":   OR,
	"nand": NAND,
	"nor":  NOR,
	"xor":  XOR,  // logical non-equivalence
	"nxor": NXOR, // logical equivalence

	"not": NOT,

	"qs": QUOTE_STRING,
	"QS": QUOTE_STRING,
	"fw": FREE_WORD_LIST,
	"FW": FREE_WORD_LIST,

	common.RegexRe2TokenLiteral: REGEX_RE2,
	common.RegexRE2TokenLiteral: REGEX_RE2,
	common.DateTimeTokenLiteral: DATETIME,
	common.DurationTokenLiteral: DURATION,

	common.NumberType:   IDENT,
	common.RangeType:    IDENT,
	common.BooleanType:  IDENT,
	common.StringType:   IDENT,
	common.RegexType:    IDENT,
	common.DateTimeType: IDENT,
	common.DurationType: IDENT,
	common.ListType:     IDENT,
	common.HashType:     IDENT,

	common.ZlsLiteral: ZLS,

	"class":  RESERVED,
	"pipe":   RESERVED,
	"run":    RESERVED,
	"select": RESERVED,
	"using":  RESERVED,
	"struct": RESERVED,
	"iface":  RESERVED,

	"i8":   RESERVED,
	"i16":  RESERVED,
	"i32":  RESERVED,
	"i64":  RESERVED,
	"i128": RESERVED,
	"int":  RESERVED,

	"u8":   RESERVED,
	"u16":  RESERVED,
	"u32":  RESERVED,
	"u64":  RESERVED,
	"u128": RESERVED,
	"uint": RESERVED,

	"d128": RESERVED,
	"b32":  RESERVED,
	"b64":  RESERVED,
}
