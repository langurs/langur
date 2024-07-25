// langur/token/rules.go

package token

func Closer(opener Type) Type {
	// This function does not attempt to decide what are valid opener/closer pairs for a given context. That must be done elsewhere.
	switch opener {
	case LPAREN:
		return RPAREN
	case LBRACE:
		return RBRACE
	case LBRACKET:
		return RBRACKET
	case LESS_THAN:
		return GREATER_THAN
	default:
		return opener
	}
}

func IsKeyword(s string) bool {
	_, ok := Keywords[s]
	return ok
}

func IsNumeric(tt Type) bool {
	return tt == INT || tt == FLOAT
}

func IsInfixComparisonOp(tt Type) bool {
	switch tt {
	case IN, OF,
		EQUAL, NOT_EQUAL,
		GREATER_THAN, LESS_THAN,
		GT_OR_EQUAL, LT_OR_EQUAL,
		DIVISIBLE_BY, NOT_DIVISIBLE_BY:

		return true
	}
	return false
}

func IsInfixLogicalOp(tt Type) bool {
	switch tt {
	case AND, OR, XOR, NXOR, NOR, NAND:
		return true
	}
	return false
}

func IsLogicalOp(tt Type) bool {
	if tt == NOT {
		return true
	}
	return IsInfixLogicalOp(tt)
}

func IsInfixValueOp(tt Type) bool {
	switch tt {
	case FORWARD,
		PLUS, MINUS,
		ASTERISK,
		SLASH, BACKSLASH, DOUBLESLASH,
		REMAINDER,
		MODULUS,
		POWER, ROOT:

		return true
	}
	return false
}

func IsInfixBitwiseOp(tt Type) bool {
	// none so far
	return false
}

func IsValueOp(tt Type) bool {
	return IsInfixValueOp(tt)
}

func IsInfixOp(tt Type) bool {
	return tt == APPEND || tt == RANGE || tt == IS || tt == AS ||
		IsInfixComparisonOp(tt) || IsInfixLogicalOp(tt) ||
		IsInfixValueOp(tt) || IsInfixBitwiseOp(tt)
}

func IsPrefixOp(tt Type) bool {
	return tt == MINUS || tt == NOT
}

func IsRightAssociativeOp(tt Type) bool {
	switch tt {
	case POWER, ROOT,
		ASSIGN:

		return true
	}
	return false
}

func ExpressionContinuationExpected(tt Type) bool {
	// used by parser
	// to prevent nonsense such as 123 567
	switch tt {
	case
		// false if...
		EOF,
		COLON, SEMICOLON, COMMA,
		RPAREN, RBRACE, RBRACKET,
		EXPANSION:

		return false
	}
	return true
}

func ImpliedExprTerminatorIfFollowedByNewline(tok Token) bool {
	if tok.AddImpliedSemicolonAtNewLine {
		// check for override of norm
		return true
	}

	// used by lexer to determine when to insert a semicolon
	tt := tok.Type
	switch tt {
	case
		// false if...
		INVALID, // prevents adding semicolon with leading white space (INVALID being the default token type)
		SEMICOLON, COLON,
		LPAREN, LBRACE, LBRACKET,
		COMMA,
		ASSIGN:

		return false

	case RANGE:
		// keeping ranges on a single line
		return true
	}

	// is and as operators not split across lines
	return tt == IS || tt == AS || !IsInfixOp(tt)
}

func MayStartFunctionArg(tt Type) bool {
	// to determine if built-in function call without parentheses (starts "unbounded" argument list)
	switch tt {
	case IDENT, INT, FLOAT,
		DATETIME, DURATION,
		STRING, QUOTE_STRING, ZLS,
		FREE_WORD_LIST,
		FUNCTION,
		SWITCH, IF,
		FOR, WHILE,
		TRUE, FALSE, NULL,
		REGEX_RE2,
		LPAREN, LBRACKET, LBRACE:

		return true
	}
	return IsPrefixOp(tt)
}

var EndUnboundedArgumentList = []Type{
	EOF, SEMICOLON, COLON, RBRACE, RPAREN, RBRACKET,
}

var EndUnboundedAssignmentExprList = []Type{
	EOF, SEMICOLON, COLON, RBRACE, RPAREN, RBRACKET, LBRACE,
}

// default comparison operator for switch expressions
var DefaultCompOp = Token{Type: EQUAL, Literal: "(==)"}

func MayPrecedeOpeningParenthesisWithoutSpacing(tok Token) bool {
	// lest they should look like function calls...
	// so there's no else(123) etc.
	switch tok.Type {
	case IF, SWITCH,
		FUNCTION,
		IDENT,
		NOT:

		return true

	case NONE, DATETIME, DURATION, STRING:
		return false
	}

	return !IsKeyword(tok.Literal)
}

func MayPrecedeOpeningBracketWithoutSpacing(tok Token) bool {
	// so there's no len[1, 2] etc.
	switch tok.Type {
	case FOR, WHILE, CASE:
		return true
	case NONE, DATETIME, DURATION, IDENT:
		return false
	}
	return !IsKeyword(tok.Literal)
}

func MayFollowDeclarationVariableWithoutSpacing(tok Token) bool {
	// used by parser
	// so we don't get nonsense like var x[1]
	switch tok.Type {
	case ASSIGN,
		RPAREN, RBRACE, RBRACKET,
		COMMA, COLON, SEMICOLON,
		EOF:

		return true
	}
	return IsInfixOp(tok.Type)
}

func MayBeUsedOnCombinationOperator(tt Type) bool {
	// += x= and= etc.
	return tt != FORWARD &&
		(tt == APPEND || tt == RANGE ||
			IsInfixValueOp(tt) || IsInfixLogicalOp(tt) || IsInfixBitwiseOp(tt))
}

func MayBeUsedAsDbOperator(tt Type) bool {
	// not? and? or? ==? etc.
	return (IsInfixComparisonOp(tt) || IsLogicalOp(tt)) &&
		tt != IN && tt != OF
}

func MayBeUsedForOperatorImpliedFunction(tt Type) bool {
	if tt == ASSIGN {
		return false
	}
	return tt != IS && tt != FORWARD && IsInfixOp(tt)
}

func MayPrecedeShorthandStringIndexing(tt Type) bool {
	switch tt {
	case IDENT, RBRACKET:
		return true
	}
	return false
}

func AllowNilRightExpression(tt Type) bool {
	// such as used for incomplete expressions in switch tests or case conditions
	switch tt {
	case COLON, SEMICOLON, COMMA, LBRACE, RBRACKET:
		return true
		// LBRACE for switch test
		// RBRACKET for [0..] expression
	}
	return false
}

func BeginsFlowBreakingStatement(tt Type) bool {
	// not including CATCH
	// MODE is a statement, but doesn't break the flow.
	// IF and SWITCH are flow control (and expressions), but don't break the flow.
	switch tt {
	case THROW, RETURN, BREAK, NEXT, FALLTHROUGH:
		return true
	}
	return false
}

func IsNoneBySymbol(tok Token) bool {
	return tok.Type == NONE && tok.Literal == "_"
}
