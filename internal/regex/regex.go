// langur/regex/regex.go

package regex

// not to be confused with the regexp package, which is the golang re2 package
// also not to be confused with langur/object/regex.go

import (
	"langur/common"
)

type RegexType int

const (
	NONE RegexType = iota
	RE2
)

func (rt RegexType) String() string {
	switch rt {
	case NONE:
		return "?"
	case RE2:
		return "Re2"
	}
	return "?"
}

func (rt RegexType) LiteralString() string {
	switch rt {
	case NONE:
		return "?"
	case RE2:
		return common.RegexRe2TokenLiteral
	}
	return "?"
}
