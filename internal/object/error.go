// langur/object/error.go

package object

import (
	"fmt"
	"langur/common"
	"langur/str"
	"langur/trace"
)

// The Error type fulfils the Error interface and the langur Object interface.

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

type Error struct {
	Contents *Hash
	Where    trace.Where
}

func (e *Error) Copy() Object {
	return &Error{Contents: e.Contents.Copy().(*Hash), Where: e.Where.Copy()}
}

// fulfilling the Object interface; not necessarily to be called
func (l *Error) Equal(eo2 Object) bool {
	r, ok := eo2.(*Error)
	if !ok {
		return false
	}
	return l.Contents.Equal(r.Contents) && l.Where.Equal(r.Where)
}

func (e *Error) Type() ObjectType {
	return ERROR_OBJ
}
func (e *Error) TypeString() string {
	return common.ErrorTypeName
}

func (e *Error) IsTruthy() bool {
	return false
}

func (e *Error) ReplString() string {
	return common.ErrorTypeName + " (" + e.Error() + ")"
}

func (e *Error) String() string {
	// langur string; should not happen
	return INTERNAL_OBJECT_ONLY
}

// also fulfilling the Go error interface...
func (e *Error) Error() string {
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
	if err == nil && history != noErrorHistory {
		hst = "*"
	}

	return cat + ": " + msg + " (" + src + ")" + hst
}

// NOTE: Error objects should always be generated with one of the methods here, ...
// ... to ensure the right fields are present with the right types.
// This should, among other things, prevent errors from generating errors (or worse, panics).

const unknownLinePos = -1

var noErrorHistory = NULL

func NewErrorFromAnything(err interface{}, source string) *Error {
	switch err := err.(type) {
	case *Error:
		return err
	case Object:
		obj := NewErrorFromObject(err)
		obj.Contents.WritePair(ERR_HASHKEY_SOURCE, NewString(source))
		return obj
	case error:
		return NewError(ERR_GENERAL, source, err.Error())
	case string:
		return NewError(ERR_GENERAL, source, err)
	default:
		return NewError(ERR_GENERAL, source, fmt.Sprintf("Unknown error type (%T)", err))
	}
}

func NewErrorFromObject(obj Object) *Error {
	switch obj := obj.(type) {
	case *Error:
		return obj
	case *Hash:
		return NewErrorFromHash(obj)
	case *String:
		return NewError(ERR_CUSTOM, "", obj.String())
	default:
		return NewError(ERR_CUSTOM, "", obj.String())
	}
}

func NewError(category, source, message string) *Error {
	hash := &Hash{}
	hash.WritePair(ERR_HASHKEY_CATEGORY, NewString(category))
	hash.WritePair(ERR_HASHKEY_SOURCE, NewString(source))
	hash.WritePair(ERR_HASHKEY_MESSAGE, NewString(message))
	hash.WritePair(ERR_HASHKEY_HISTORY, noErrorHistory)
	return &Error{Contents: hash}
}

// called when throwing a hash as an error
// enforces the field types
func NewErrorFromHash(hash *Hash) *Error {
	// Add required fields if not present, and enforce their type if they are.
	// Other fields are allowed (optional).
	enforceHashString(hash, ERR_HASHKEY_CATEGORY, NewString(ERR_GENERAL))
	enforceHashString(hash, ERR_HASHKEY_SOURCE, ZLS)
	enforceHashString(hash, ERR_HASHKEY_MESSAGE, NewString("Unknown Error"))
	enforceHashErrorHistory(hash, ERR_HASHKEY_HISTORY)
	return &Error{Contents: hash}
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

func enforceHashErrorHistory(hash *Hash, key Object) {
	// must be previous error or langur NULL
	val, err := hash.GetValue(key)
	if err == nil {
		if val.Type() == HASH_OBJ {
			hash.WritePair(key, NewErrorFromHash(val.(*Hash)))
		} else if val != noErrorHistory {
			hash.WritePair(key, NewErrorFromObject(val))
		}
	} else {
		hash.WritePair(key, noErrorHistory)
	}
}
