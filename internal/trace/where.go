// langur/trace/where.go

package trace

import (
	"fmt"
	"langur/regexp"
	"strings"
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

var splitLines = regexp.MustCompile("\n")

func (w Where) Trace(source string) string {
	lines := splitLines.Split(source, -1)

	if w.Line <= len(lines) && w.LinePosition > 0 {
		// using 1-based indexing
		line := lines[w.Line-1]
		space := strings.Repeat(" ", w.LinePosition-1)
		return line + "\n" + space + "^" + "\n"
	}

	return ""
}

func NewWhere(line, linePosition int) Where {
	return Where{Line: line, LinePosition: linePosition}
}

type WhereSlice = []*Where

func NewWhereAddress(line, linePosition int) *Where {
	return &Where{Line: line, LinePosition: linePosition}
}

// work back to first non-nil entry
func FindLocation(wSlc WhereSlice, from int) Where {
	if len(wSlc) > from {
		for i := from; i > -1; i-- {
			if wSlc[i] != nil {
				return *(wSlc[i])
			}
		}
	}
	return NewWhere(0, 0)
}

func CopyWhereSlice(wSlc WhereSlice) WhereSlice {
	newSlc := make(WhereSlice, len(wSlc))
	for i := range wSlc {
		if wSlc[i] != nil {
			newSlc[i] = NewWhereAddress(wSlc[i].Line, wSlc[i].LinePosition)
		}
	}
	return newSlc
}

func AppendWhereSlice(wSlc1, wSlc2 WhereSlice) WhereSlice {
	newWSlc := make(WhereSlice, len(wSlc1)+len(wSlc2))
	copy(newWSlc, wSlc1)
	copy(newWSlc[len(wSlc1):], wSlc2)
	return newWSlc
}
