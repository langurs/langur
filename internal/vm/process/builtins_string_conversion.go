// langur/vm/process/builtins_string_conversion.go

package process

import (
	"langur/cpoint"
	"langur/object"
	"langur/str"
	"unicode/utf8"
)

// cp2s, s2cp, s2s, s2gc
// s2b, b2s
// s2n

var bi_s2b = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:        "s2b",
		Description: "returns list of UTF-8 bytes from a langur string",

		ParamPositional: []object.Parameter{
			object.Parameter{ExternalName: "string"},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		// langur string to UTF-8 bytes
		const fnName = "s2b"

		s, ok := args[0].(*object.String)
		if !ok {
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected string for argument string")
		}
		bytes := s.ByteSlc()
		arr := &object.List{Elements: make([]object.Object, len(bytes))}
		for i, b := range bytes {
			arr.Elements[i] = object.NumberFromInt(int(b))
		}
		return arr
	},
}

var bi_b2s = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:        "b2s",
		Description: "converts a byte or list of UTF-8 bytes to a langur string",

		ParamPositional: []object.Parameter{
			object.Parameter{ExternalName: "bytes"},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		// UTF-8 bytes to langur string
		const fnName = "b2s"

		switch arg := args[0].(type) {
		case *object.Number:
			b, err := arg.ToByte()
			if err != nil {
				return object.NewError(object.ERR_ARGUMENTS, fnName, err.Error())
			}
			return utf8bytesToString(fnName, []byte{b})

		case *object.List:
			bSlc := make([]byte, len(arg.Elements))
			for i, v := range arg.Elements {
				var b byte
				var err error

				switch v := v.(type) {
				case *object.Number:
					b, err = v.ToByte()
					if err != nil {
						return object.NewError(object.ERR_ARGUMENTS, fnName, err.Error())
					}
				default:
					return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected integer or list of integers")
				}
				bSlc[i] = b
			}
			return utf8bytesToString(fnName, bSlc)
		}
		return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected integer or list of integers")
	},
}

func utf8bytesToString(fnName string, bSlc []byte) object.Object {
	if utf8.Valid(bSlc) {
		s, err := object.NewStringFromParts(bSlc)
		if err == nil {
			return s
		}
		return object.NewError(object.ERR_ARGUMENTS, fnName, err.Error())
		// return object.StringFromByteSlice(bSlc)
	}
	return object.NewError(object.ERR_ARGUMENTS, fnName, "Invalid UTF-8 byte sequence")
}

var bi_cp2s = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:        "cp2s",
		Description: "converts a code point (integer) or list of code points to a string",

		ParamPositional: []object.Parameter{
			object.Parameter{ExternalName: "cp"},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		// code point(s) to string
		const fnName = "cp2s"

		rSlc, err := object.CodePointsToFlatRuneSlice(args[0])
		if err != nil {
			return object.NewError(object.ERR_ARGUMENTS, fnName, err.Error())
		}
		s, err := object.NewStringFromParts(rSlc)
		if err == nil {
			return s
		}
		return object.NewError(object.ERR_ARGUMENTS, fnName, err.Error())
	},
}

var bi_s2cp = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:        "s2cp",
		Description: "indexes a string to a code point or a list of code points",

		ParamPositional: []object.Parameter{
			object.Parameter{ExternalName: "string"},
		},

		ParamByName: []object.Parameter{
			object.Parameter{ExternalName: "of"},
			object.Parameter{ExternalName: "alt"},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		// string to code point(s): indexes string and returns code point or list of code points
		const fnName = "s2cp"

		s, ok := args[0].(*object.String)
		if !ok {
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected string for first argument")
		}
		var result object.Object
		var err error

		index, alt := args[1], args[2]

		result, err = s.Index(index, false)
		if err != nil {
			if alt != nil {
				return alt
			}
			return object.NewError(object.ERR_ARGUMENTS, fnName, err.Error())
		}

		return result
	},
}

var bi_s2gc = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:        "s2gc",
		Description: "converts string to grapheme clusters list",

		ParamPositional: []object.Parameter{
			object.Parameter{ExternalName: "string"},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		// string to grapheme clusters
		const fnName = "s2gc"

		s, ok := args[0].(*object.String)
		if !ok {
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected string for argument string")
		}

		var clusters []object.Object

		graphemes := str.Graphemes(s.String())
		for _, gr := range graphemes {
			if len(gr) == 1 {
				// 1 code point
				clusters = append(clusters, object.NumberFromInt(int(gr[0])))
			} else {
				// a cluster; make a list
				clusters = append(clusters, &object.List{Elements: runeSlcToObjects(gr)})
			}
		}

		return &object.List{Elements: clusters}
	},
}

func runeSlcToObjects(rSlc []rune) []object.Object {
	list := make([]object.Object, len(rSlc))
	for i, r := range rSlc {
		list[i] = object.NumberFromInt(int(r))
	}
	return list
}

var bi_s2s = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:        "s2s",
		Description: "indexes a string to a string",

		ParamPositional: []object.Parameter{
			object.Parameter{ExternalName: "string"},
		},

		ParamByName: []object.Parameter{
			object.Parameter{ExternalName: "of"},
			object.Parameter{ExternalName: "alt"},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		// string to string: indexes string and returns string
		const fnName = "s2s"

		s, ok := args[0].(*object.String)
		if !ok {
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected string for first argument")
		}
		var result object.Object
		var err error

		index, alt := args[1], args[2]

		result, err = s.Index(index, true)
		if err != nil {
			if alt != nil {
				return alt
			}
			return object.NewError(object.ERR_ARGUMENTS, fnName, err.Error())
		}

		return result
	},
}

var bi_s2n = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:        "s2n",
		Description: "returns list of numbers from a langur string, interpreting 0-9, A-Z, and a-z as base 36 numbers",

		ParamPositional: []object.Parameter{
			object.Parameter{ExternalName: "string"},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		// langur string or code point to numbers (interpreted from base 36)
		const fnName = "s2n"

		var rSlc []rune
		var one bool

		switch arg := args[0].(type) {
		case *object.String:
			rSlc = arg.RuneSlc()

		case *object.Number:
			cp, err := arg.ToRune()
			if err != nil {
				return object.NewError(object.ERR_ARGUMENTS, fnName, "Invalid code point")
			}
			rSlc = []rune{cp}
			one = true

		default:
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected string or code point for argument string")
		}

		if one {
			n, err := cpoint.Base36ToNumber(rSlc[0])
			if err != nil {
				return object.NewError(object.ERR_ARGUMENTS, fnName, err.Error())
			}
			return object.NumberFromInt(n)
		}

		elements := make([]object.Object, len(rSlc))
		for i, cp := range rSlc {
			n, err := cpoint.Base36ToNumber(cp)
			if err != nil {
				return object.NewError(object.ERR_ARGUMENTS, fnName, err.Error())
			}
			elements[i] = object.NumberFromInt(n)
		}

		return &object.List{Elements: elements}
	},
}
