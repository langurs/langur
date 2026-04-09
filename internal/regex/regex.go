// langur/regex/regex.go

package regex

// not to be confused with the regexp package, which is the golang re2 package
// also not to be confused with langur/object/regex.go

import (
	"fmt"
	"langur/regexp"
	"langur/str"
	"langur/common"
)

type RegexType int

const (
	NONE RegexType = iota
	RE2
)

func (rt RegexType) Escape(s string) (string, error) {
	switch rt {
	case NONE:
		return str.Escape(s), nil
	case RE2:
		return regexp.QuoteMeta(s), nil
	}
	return "?", fmt.Errorf("Cannot escape meta characters; unknown regex type")
}

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
