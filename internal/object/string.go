// langur/object/string.go

package object

import (
	"fmt"
	"langur/common"
	"langur/str"
)

// NOTE: using these functions to build strings, not &String{}

var ZLS = NewString("")

func NewStringFromParts(parts ...interface{}) (*String, error) {
	s, err := str.BuildString(parts)
	if err == nil {
		return &String{value: s}, nil
	}
	return nil, err
}

func NewString(s string) *String {
	return &String{value: s}
}

type String struct {
	value string
	lenCP int // length in code points ("runes" in Go)
}

func (s *String) Copy() Object {
	return &String{value: s.value, lenCP: s.lenCP}
}

func (l *String) Equal(s2 Object) bool {
	r, ok := s2.(*String)
	if !ok {
		return false
	}
	return l.value == r.value
}

func (l *String) GreaterThan(s2 Object) (bool, bool) {
	r, ok := s2.(*String)
	if !ok {
		return false, false
	}
	return l.value > r.value, true
}

func (s *String) ComposedString() string {
	// with quote marks
	// TODO: update to escape quote mark used, or escape more than that
	return `"` + s.value + `"`
}

func (s *String) ReplString() string {
	return fmt.Sprintf("%s %q", common.StringTypeName, s.value)
}

func (s *String) String() string {
	return s.value
}

func (s *String) Type() ObjectType {
	return STRING_OBJ
}
func (s *String) TypeString() string {
	return common.StringTypeName
}

func (s *String) IsTruthy() bool {
	return len(s.value) != 0
}

func (s *String) LenCP() int {
	// length of string in code points
	if s.lenCP == 0 {
		s.lenCP = len(s.RuneSlc())
	}
	return s.lenCP
}

func (s *String) RuneSlc() []rune {
	return []rune(s.value)
}

func (s *String) ByteSlc() []byte {
	return []byte(s.value)
}

func (s *String) IndexToCP(idx int) (rune, bool) {
	rSlc := s.RuneSlc()
	if idx < 1 || idx > len(rSlc) {
		return 0, false
	}
	return rSlc[idx-1], true
}

func (s *String) IndexRangeToCodePoints(start, end int) ([]rune, bool) {
	rSlc := s.RuneSlc()
	L := len(rSlc)
	if start < 1 || start > L {
		return []rune{}, false
	}
	if end < 1 || end > L {
		return []rune{}, false
	}

	if start > end {
		// high to low range
		// not a Unicode string reversal; just a simple code point reversal
		elements := make([]rune, 0, start-end+1)
		for i := start - 1; i > end-2; i-- {
			elements = append(elements, rSlc[i])
		}
		return elements, true
	} else {
		// low to high range
		return rSlc[start-1 : end], true
	}
}

func (s *String) RemoveIndices(indices Object) (*String, error) {
	// build new string without indices we want to remove
	cpSlc := []rune{}

	for i, cp := range s.RuneSlc() {
		excludeThis, err := intInObject(i+1, indices)
		if err != nil {
			return s, err
		}
		if !excludeThis {
			cpSlc = append(cpSlc, cp)
		}
	}

	return NewStringFromParts(cpSlc)
}
