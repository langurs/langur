// langur/object/pattern.go

package object

import (
	"fmt"
	"langur/common"
	"langur/opcode"
	"langur/pattern"
	"langur/regexp" // a modified copy of Go's standard regexp (re2) package
	"langur/str"
)

var codeToPatternType = map[int]pattern.PatternType{
	opcode.OC_Pattern_None: pattern.NONE,
	opcode.OC_Pattern_Re2:  pattern.RE2,
}

type Pattern struct {
	Compiled    interface{}
	Pattern     string
	PatternType pattern.PatternType
	FreeSpacing bool
}

func (re *Pattern) Copy() Object {
	var newComp interface{}

	switch comp := re.Compiled.(type) {
	case *regexp.Regexp:
		newComp = comp.Copy()
	default:
		bug("Pattern.Copy()", "Missing Copy() method for pattern type")
		// newComp nil
	}

	return &Pattern{
		Compiled:    newComp,
		Pattern:     re.Pattern,
		PatternType: re.PatternType,
		FreeSpacing: re.FreeSpacing,
	}
}

func (l *Pattern) Equal(r2 Object) bool {
	r, ok := r2.(*Pattern)
	if !ok {
		return false
	}
	return l.PatternType == r.PatternType && l.Pattern == r.Pattern
}

func (r *Pattern) ComposedString() string {
	// TODO: update to escape quote mark used, or escape more than that
	return r.PatternType.LiteralString() + "/" + r.Pattern + "/"
}

func (r *Pattern) ReplString() string {
	return fmt.Sprintf("%s %s %q",
		common.PatternTypeName, r.PatternType.String(), r.Pattern)
}

func (r *Pattern) String() string {
	return r.Pattern
}

func (r *Pattern) Type() ObjectType {
	return PATTERN_OBJ
}
func (r *Pattern) TypeString() string {
	return common.PatternTypeName
}

func (r *Pattern) IsTruthy() bool {
	return len(r.Pattern) != 0
}

func NewPatternByOpCode(pattern string, code int) (result Object, err error) {
	patternType, ok := codeToPatternType[code]
	if !ok {
		return nil, fmt.Errorf("Unknown Pattern Type")
	}
	return NewPattern(pattern, patternType)
}

func NewPattern(p string, patternType pattern.PatternType) (result Object, err error) {
	reggie := &Pattern{Pattern: p, PatternType: patternType}

	if patternType == pattern.RE2 {
		compiled, err := regexp.Compile(p)
		if err != nil {
			return reggie, err
		}
		reggie.Compiled = compiled
		return reggie, nil
	}
	return nil, fmt.Errorf("Unknown Pattern Type")
}

func EscStringByOpCode(obj Object, code int) (result Object, err error) {
	patternType, ok := codeToPatternType[code]
	if !ok {
		return nil, fmt.Errorf("Unknown Pattern Type")
	}
	return EscString(obj, patternType)
}

func EscString(obj Object, patternType pattern.PatternType) (result Object, err error) {
	var strObj Object
	strObj, err = AutoString(obj)
	if err != nil {
		return
	}
	if patternType == pattern.NONE {
		return NewString(str.Escape(strObj.String())), nil

	} else if patternType == pattern.RE2 {
		return NewString(regexp.QuoteMeta(strObj.String())), nil
	}
	return nil, fmt.Errorf("Unknown Escape Type for String Object")
}

func PatternMatchingOrError(re *Pattern, o2 Object) (Object, error) {
	strObj, err := AutoString(o2)
	if err != nil {
		return nil, err
	}
	return PatternMatching(re, strObj.String())
}
