// langur/common/words.go
// langur type names and tokens

package common

const IdentifierPatternString = "[_a-zA-Z][a-zA-Z0-9_]*$"

// const IdentifierPatternString = "^[a-zA-Z][_a-zA-Z0-9]*|_+[a-zA-Z0-9][_a-zA-Z0-9]*$"

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
