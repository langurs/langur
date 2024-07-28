// langur/common/where.go

package common

import (
	"fmt"
)

type Where struct {
	Line         int
	LinePosition int // in code points or "runes"
}

func (w Where) Copy() Where {
	return Where{
		Line:         w.Line,
		LinePosition: w.LinePosition,
	}
}

func (w Where) String() string {
	if w.Line == 0 {
		return "?"
	}
	return fmt.Sprintf("%d %d", w.Line, w.LinePosition)
}

func NewWhere(line, linePosition int) Where {
	return Where{Line: line, LinePosition: linePosition}
}

// func FindLocation(wSlc []*Where, start int) (line, linePosition int, ok bool) {
// }
