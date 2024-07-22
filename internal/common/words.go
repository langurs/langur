// langur/common/words.go
// langur type names and tokens

package common

const IdentifierRegexString = "[a-zA-Z0-9][a-zA-Z0-9_]*"

const (
	NumberType   = "number"
	RangeType    = "range"
	BooleanType  = "bool"
	StringType   = "string"
	RegexType    = "regex"
	DateTimeType = "datetime"
	DurationType = "duration"
	ListType     = "list"
	HashType     = "hash"

	NumberTypeName   = "Number"
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
)
