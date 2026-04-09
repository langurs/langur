// langur/object/regex.go

package object

import (
	"fmt"
	"langur/common"
	"langur/opcode"
	"langur/regex"
	"langur/regexp" // a modified copy of Go's standard regexp (re2) package
	"langur/str"
)

var codeToReType = map[int]regex.RegexType{
	opcode.OC_Regex_None: regex.NONE,
	opcode.OC_Regex_Re2:  regex.RE2,
}

type Regex struct {
	Compiled    interface{}
	Pattern     string
	RegexType   regex.RegexType
	FreeSpacing bool
}

func (re *Regex) Copy() Object {
	var newComp interface{}

	switch comp := re.Compiled.(type) {
	case *regexp.Regexp:
		newComp = comp.Copy()
	default:
		bug("Regex.Copy()", "Missing Copy() method for regex type")
	}

	return &Regex{
		Compiled:    newComp,
		Pattern:     re.Pattern,
		RegexType:   re.RegexType,
		FreeSpacing: re.FreeSpacing,
	}
}

func (l *Regex) Equal(r2 Object) bool {
	r, ok := r2.(*Regex)
	if !ok {
		return false
	}
	return l.RegexType == r.RegexType && l.Pattern == r.Pattern
}

func (r *Regex) ComposedString() string {
	// TODO: update to escape quote mark used, or escape more than that
	return r.RegexType.LiteralString() + "/" + r.Pattern + "/"
}

func (r *Regex) ReplString() string {
	return fmt.Sprintf("%s %s %q",
		common.RegexTypeName, r.RegexType.String(), r.Pattern)
}

func (r *Regex) String() string {
	return r.Pattern
}

func (r *Regex) Type() ObjectType {
	return REGEX_OBJ
}
func (r *Regex) TypeString() string {
	return common.RegexTypeName
}

func (r *Regex) IsTruthy() bool {
	return len(r.Pattern) != 0
}

func NewRegexByOpCode(pattern string, code int) (result Object, err error) {
	reType, ok := codeToReType[code]
	if !ok {
		return nil, fmt.Errorf("Unknown Regex Type")
	}
	return NewRegex(pattern, reType)
}

func NewRegex(pattern string, regexType regex.RegexType) (result Object, err error) {
	reggie := &Regex{Pattern: pattern, RegexType: regexType}

	if regexType == regex.RE2 {
		compiled, err := regexp.Compile(pattern)
		if err != nil {
			return reggie, err
		}
		reggie.Compiled = compiled
		return reggie, nil
	}
	return nil, fmt.Errorf("Unknown Regex Type")
}

func EscStringByOpCode(obj Object, code int) (result Object, err error) {
	reType, ok := codeToReType[code]
	if !ok {
		return nil, fmt.Errorf("Unknown Regex Type")
	}
	return EscString(obj, reType)
}

func EscString(obj Object, reType regex.RegexType) (result Object, err error) {
	var strObj Object
	strObj, err = AutoString(obj)
	if err != nil {
		return
	}
	if reType == regex.NONE {
		return NewString(str.Escape(strObj.String())), nil

	} else if reType == regex.RE2 {
		return NewString(regexp.QuoteMeta(strObj.String())), nil
	}
	return nil, fmt.Errorf("Unknown Escape Type for String Object")
}

func RegexMatchingOrError(re *Regex, o2 Object) (Object, error) {
	strObj, err := AutoString(o2)
	if err != nil {
		return nil, err
	}
	return RegexMatching(re, strObj.String())
}
