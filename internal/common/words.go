// langur/common/words.go
// langur type names and tokens

package common

const IdentifierRegexString = "[_a-zA-Z][a-zA-Z0-9_]*$"

// const IdentifierRegexString = "^[a-zA-Z][_a-zA-Z0-9]*|_+[a-zA-Z0-9][_a-zA-Z0-9]*$"

const (
	NumberType   = "number"
	ComplexType = "complex"
	RangeType    = "range"
	BooleanType  = "bool"
	StringType   = "string"
	RegexType    = "regex"
	DateTimeType = "datetime"
	DurationType = "duration"
	ListType     = "list"
	HashType     = "hash"

	NumberTypeName   = "Number"
	ComplexTypeName = "Complex"
	RangeTypeName    = "Range"
	BooleanTypeName  = "Boolean"
	StringTypeName   = "String"
	RegexTypeName    = "Regex"
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

	RegexRe2TokenLiteral = "re"
	RegexRE2TokenLiteral = "RE"

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
