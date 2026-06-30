// langur/common/words.go
// langur type names and tokens

package common

import (
	"fmt"
)

// maybe arbitrary, probably needs to be limited
const IdentifierLenMax = 128

// not including ^ or $ as may be part of a larger string
var IdentifierPatternString = fmt.Sprintf("[_a-zA-Z][a-zA-Z0-9_]{0,%d}", IdentifierLenMax-1)
var WordPatternString = "[a-zA-Z][a-zA-Z0-9_]*"

const (
	NumberType   = "number"
	ComplexType  = "complex"
	RangeType    = "range"
	BooleanType  = "bool"
	StringType   = "string"
	PatternType  = "pattern"
	DateTimeType = "datetime"
	DurationType = "duration"
	ListType     = "list"
	HashType     = "hash"

	NumberTypeName   = "Number"
	ComplexTypeName  = "Complex"
	RangeTypeName    = "Range"
	BooleanTypeName  = "Boolean"
	StringTypeName   = "String"
	PatternTypeName  = "Pattern"
	DateTimeTypeName = "DateTime"
	DurationTypeName = "Duration"
	ListTypeName     = "List"
	HashTypeName     = "Hash"

	NullTypeName      = "Null"
	ErrorTypeName     = "Error"
	NameValueTypeName = "NameValue"

	CompiledCodeTypeName = "Compiled"
	FuntionTypeName      = "Compiled"
	BuiltInTypeName      = "BuiltIn"
)

const (
	FunctionTokenLiteral = "fn"
	DateTimeTokenLiteral = "dt"
	DurationTokenLiteral = "dr"

	PatternRe2TokenLiteral = "re"
	PatternRE2TokenLiteral = "RE"

	NullTokenLiteral  = "null"
	TrueTokenLiteral  = "true"
	FalseTokenLiteral = "false"

	ZlsLiteral = "zls"

	// see token.NegatedLiteral()
	IsNotLiteral = "is not"
	NotInLiteral = "not in"
	NotOfLiteral = "not of"
	
	MainFnName = "_main"
)
