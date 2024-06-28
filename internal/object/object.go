// langur/object/object.go

package object

import (
	"fmt"
	"langur/common"
	"langur/str"
)

// for strings that aren't supposed to be returned to the user
const INTERNAL_OBJECT_ONLY = "INTERNAL OBJECT"

type ObjectType byte

const (
	// 0 used as code to indicate type unknown, so it is skipped here
	_ ObjectType = iota

	NUMBER_OBJ
	BOOLEAN_OBJ
	NULL_OBJ

	STRING_OBJ
	REGEX_OBJ
	RANGE_OBJ

	DATETIME_OBJ
	DURATION_OBJ

	COMPILED_CODE_OBJ
	BUILTIN_FUNCTION_OBJ

	LIST_OBJ
	HASH_OBJ

	ERROR_OBJ
)

func AutoString(o Object) (Object, error) {
	// to have some discretion about what should recieve auto-stringification
	// no to lists, hashes, etc.
	switch o.(type) {
	case *String:
		return o, nil

	case *Number, *DateTime, *Duration, *Regex:
		return NewString(o.String()), nil

	default:
		return o, fmt.Errorf("No auto-stringification on type %s", o.TypeString())
	}
}

var TypeNameToType = map[string]ObjectType{
	common.NumberType:   NUMBER_OBJ,
	common.RangeType:    RANGE_OBJ,
	common.BooleanType:  BOOLEAN_OBJ,
	common.StringType:   STRING_OBJ,
	common.RegexType:    REGEX_OBJ,
	common.DateTimeType: DATETIME_OBJ,
	common.DurationType: DURATION_OBJ,
	common.ListType:     LIST_OBJ,
	common.HashType:     HASH_OBJ,
}

func Is(obj Object, otype Object) (bool, error) {
	switch t := otype.(type) {
	case *String:
		switch t.String() {
		case "callable":
			return IsCallable(obj), nil
		default:
			return false, fmt.Errorf("String %q not defined to determine type", str.ReformatInput(t.String()))
		}

		// case *BuiltIn:
		// 	ot, ok := TypeNameToType[t.Name]
		// 	if ok {
		// 		return obj.Type() == ot, nil
		// 	}

	}
	return false, fmt.Errorf("Expected string as definition to determine type")
}

func ComposedOrRegularString(obj Object) string {
	switch obj := obj.(type) {
	case IComposableString:
		return obj.ComposedString()
	default:
		return obj.String()
	}
}

func CopyRefSlice(objSlc []Object) []Object {
	copiedObjs := make([]Object, len(objSlc))
	copy(copiedObjs, objSlc)
	return copiedObjs
}

func CopySlice(objSlc []Object) []Object {
	copiedObjs := make([]Object, len(objSlc))
	for i := range objSlc {
		copiedObjs[i] = objSlc[i].Copy()
	}
	return copiedObjs
}

func CopyAndReverseSlice(objSlc []Object) []Object {
	lso := len(objSlc)
	newSlc := make([]Object, lso)
	for i := 0; i < lso/2; i++ {
		j := lso - i - 1
		newSlc[i], newSlc[j] = objSlc[j].Copy(), objSlc[i].Copy()
	}
	if lso%2 == 1 {
		// if odd, add missed middle object
		i := lso / 2
		newSlc[i] = objSlc[i].Copy()
	}
	return newSlc
}

func SliceHasImpureEffects(objSlc ...Object) bool {
	for i := range objSlc {
		switch obj := objSlc[i].(type) {
		case IDefilable:
			if obj.HasImpureEffects() {
				return true
			}
		}
	}
	return false
}
