// langur/object/exception.go

package object

import (
	"fmt"
	"langur/common"
	"langur/str"
	"langur/trace"
)

// The Exception type fulfils the Error interface and the langur Object interface.

const (
	// categories
	ERR_GENERAL   = "general"
	ERR_CUSTOM    = "custom"
	ERR_INDEX     = "index"
	ERR_MATH      = "math"
	ERR_ARGUMENTS = "args"
)

var ERR_HASHKEY_CATEGORY = NewString("cat")
var ERR_HASHKEY_SOURCE = NewString("src")
var ERR_HASHKEY_MESSAGE = NewString("msg")
var ERR_HASHKEY_HISTORY = NewString("hst")

type Exception struct {
	Contents *Hash
	Where    trace.Where
}

func (e *Exception) Copy() Object {
	return &Exception{Contents: e.Contents.Copy().(*Hash), Where: e.Where.Copy()}
}

// fulfilling the Object interface; not necessarily to be called
func (l *Exception) Equal(eo2 Object) bool {
	r, ok := eo2.(*Exception)
	if !ok {
		return false
	}
	return l.Contents.Equal(r.Contents) && l.Where.Equal(r.Where)
}

func (e *Exception) Type() ObjectType {
	return ERROR_OBJ
}
func (e *Exception) TypeString() string {
	return common.ErrorTypeName
}

func (e *Exception) IsTruthy() bool {
	return false
}

func (e *Exception) ReplString() string {
	return common.ErrorTypeName + " (" + e.Error() + ")"
}

func (e *Exception) String() string {
	// langur string; should not happen
	return INTERNAL_OBJECT_ONLY
}

// also fulfilling the Go error interface...
func (e *Exception) Error() string {
	// enforced types for these values already (see NOTE below)
	cat := ERR_GENERAL
	category, err := e.Contents.GetValue(ERR_HASHKEY_CATEGORY)
	if err == nil {
		cat = str.Escape(category.String())
	}

	src := ""
	source, err := e.Contents.GetValue(ERR_HASHKEY_SOURCE)
	if err == nil {
		src = str.Escape(source.String())
	}

	msg := "Unknown Error"
	message, err := e.Contents.GetValue(ERR_HASHKEY_MESSAGE)
	if err == nil {
		msg = str.Escape(message.String())
	}

	hst := ""
	history, err := e.Contents.GetValue(ERR_HASHKEY_HISTORY)
	if err == nil && history != noExceptionHistory {
		hst = "*"
	}

	return cat + ": " + msg + " (" + src + ")" + hst
}

// NOTE: Error objects should always be generated with one of the methods here, ...
// ... to ensure the right fields are present with the right types.
// This should, among other things, prevent errors from generating errors (or worse, panics).

const unknownLinePos = -1

var noExceptionHistory = NULL

func NewExceptionFromAnything(err interface{}, source string) *Exception {
	switch err := err.(type) {
	case *Exception:
		return err
	case Object:
		obj := NewExceptionFromObject(err)
		obj.Contents.WritePair(ERR_HASHKEY_SOURCE, NewString(source))
		return obj
	case error:
		return NewException(ERR_GENERAL, source, err.Error())
	case string:
		return NewException(ERR_GENERAL, source, err)
	default:
		return NewException(ERR_GENERAL, source, fmt.Sprintf("Unknown error type (%T)", err))
	}
}

func NewExceptionFromObject(obj Object) *Exception {
	switch obj := obj.(type) {
	case *Exception:
		return obj
	case *Hash:
		return NewExceptionFromHash(obj)
	case *String:
		return NewException(ERR_CUSTOM, "", obj.String())
	default:
		return NewException(ERR_CUSTOM, "", obj.String())
	}
}

func NewException(category, source, message string) *Exception {
	hash := &Hash{}
	hash.WritePair(ERR_HASHKEY_CATEGORY, NewString(category))
	hash.WritePair(ERR_HASHKEY_SOURCE, NewString(source))
	hash.WritePair(ERR_HASHKEY_MESSAGE, NewString(message))
	hash.WritePair(ERR_HASHKEY_HISTORY, noExceptionHistory)
	return &Exception{Contents: hash}
}

// called when throwing a hash as an error
// enforces the field types
func NewExceptionFromHash(hash *Hash) *Exception {
	// Add required fields if not present, and enforce their type if they are.
	// Other fields are allowed (optional).
	enforceHashString(hash, ERR_HASHKEY_CATEGORY, NewString(ERR_GENERAL))
	enforceHashString(hash, ERR_HASHKEY_SOURCE, ZeroLengthString())
	enforceHashString(hash, ERR_HASHKEY_MESSAGE, NewString("Unknown Error"))
	enforceHashExceptionHistory(hash, ERR_HASHKEY_HISTORY)
	return &Exception{Contents: hash}
}

func enforceHashString(hash *Hash, key Object, altValue Object) {
	val, err := hash.GetValue(key)
	if err == nil {
		// key found; enforce type
		if val.Type() != STRING_OBJ {
			hash.WritePair(key, NewString(val.String()))
		}
	} else {
		hash.WritePair(key, altValue)
	}
}

// func enforceHashNumber(hash *Hash, key Object, altValue Object) {
// 	val, err := hash.GetValue(key)
// 	if err == nil {
// 		// key found; enforce type
// 		if val.Type() != NUMBER_OBJ {
// 			var ns string
// 			if val.Type() == STRING_OBJ {
// 				ns = val.(*String).Value
// 			} else {
// 				ns = val.String()
// 			}
// 			i, err := str.StrToInt(ns, 10)
// 			if err != nil {
// 				i = 0
// 			}
// 			hash.WritePair(err_HASHKEY_LINE, NumberFromInt(i))
// 		}
// 	} else {
// 		hash.WritePair(key, altValue)
// 	}
// }

func enforceHashExceptionHistory(hash *Hash, key Object) {
	// must be previous error or langur NULL
	val, err := hash.GetValue(key)
	if err == nil {
		if val.Type() == HASH_OBJ {
			hash.WritePair(key, NewExceptionFromHash(val.(*Hash)))
		} else if val != noExceptionHistory {
			hash.WritePair(key, NewExceptionFromObject(val))
		}
	} else {
		hash.WritePair(key, noExceptionHistory)
	}
}
