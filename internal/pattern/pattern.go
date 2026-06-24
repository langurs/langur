// langur/pattern/pattern.go

package pattern

import (
	"langur/common"
)

type PatternType int

const (
	NONE PatternType = iota
	RE2
)

func (rt PatternType) String() string {
	switch rt {
	case NONE:
		return "?"
	case RE2:
		return "Re2"
	}
	return "?"
}

func (rt PatternType) LiteralString() string {
	switch rt {
	case NONE:
		return "?"
	case RE2:
		return common.PatternRe2TokenLiteral
	}
	return "?"
}
