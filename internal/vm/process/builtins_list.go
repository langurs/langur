// langur/vm/process/builtins_list.go

package process

import (
	"langur/object"
	"langur/str"
	"strings"
)

// reverse, rotate
// less, more

var bi_less = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:        "less",
		Description: "creates a new list or string, with 1 less element at the end, or an empty one if the length is already 0; may return empty list or string; also may create a new hash, with specified keys left out, or return original hash if specified keys not present",

		ParamPositional: []object.Parameter{
			object.Parameter{ExternalName: "over"},
		},

		ParamByName: []object.Parameter{
			object.Parameter{ExternalName: "of"},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		const fnName = "less"

		of := args[1]

		switch over := args[0].(type) {
		case *object.List:
			if len(over.Elements) == 0 {
				return over
			}
			if of == nil {
				// 1 less element
				newElements := make([]object.Object, len(over.Elements)-1)
				copy(newElements, over.Elements[:len(over.Elements)-1])
				return &object.List{Elements: newElements}
			}

			// remove specific indices
			list, err := over.Index(of, true, false)
			if err != nil {
				return object.NewError(object.ERR_GENERAL, fnName, "Error removing indices from list: "+err.Error())
			}
			return list

		case *object.String:
			if over.LenCP() == 0 {
				return over
			}
			if of == nil {
				// 1 less code point
				cpSlc := over.RuneSlc()
				s, err := object.NewStringFromParts(cpSlc[:len(cpSlc)-1])
				if err == nil {
					return s
				}
				return object.NewError(object.ERR_GENERAL, fnName, "Error removing indices from list: "+err.Error())
			}

			str, err := over.Index(of, true, true)
			if err != nil {
				return object.NewError(object.ERR_GENERAL, fnName, "Error removing indices from string: "+err.Error())
			}
			return str

		case *object.Hash:
			if of == nil {
				return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected key or list of keys to remove from hash")
			}
			if len(over.Pairs) == 0 {
				return over
			}
			hash, err := over.RemoveKeys(of)
			if err != nil {
				return object.NewError(object.ERR_GENERAL, fnName, "Error removing keys from hash: "+err.Error())
			}
			return hash
		}

		return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected list, hash or string")
	},
}

var bi_more = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:        "more",
		Description: "creates a new list or string, adding an item or items",

		ParamPositional: []object.Parameter{
			object.Parameter{ExternalName: "with"},
			object.Parameter{ExternalName: "add"},
		},
		ParamExpansionMin: 1,
		ParamExpansionMax: -1,
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		const fnName = "more"

		add := args[1].(*object.List).Elements

		switch with := args[0].(type) {
		case *object.List:
			newElements := make([]object.Object, len(with.Elements)+len(add))
			copy(newElements, with.Elements)
			copy(newElements[len(with.Elements):], add)
			return &object.List{Elements: newElements}

		case *object.String:
			ns := &strings.Builder{}
			ns.WriteString(with.String())
			for i := range add {
				// items may be strings or code points (integers)
				switch item := add[i].(type) {
				case *object.String:
					ns.WriteString(item.String())

				case *object.Number:
					r, err := item.ToRune()
					if err == nil {
						ns.WriteRune(r)
					} else {
						return object.NewError(object.ERR_ARGUMENTS, fnName, "error adding to string")
					}

				default:
					return object.NewError(object.ERR_ARGUMENTS, fnName, "error adding to string")
				}
			}
			return object.NewString(ns.String())

		case *object.Hash:
			hash := with

			for i := range add {
				from, ok := add[i].(*object.Hash)
				if !ok {
					return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected hashes to add to hash")
				}
				for _, kv := range from.Pairs {
					// adding values not already present in original hash
					if hash.KeyExists(kv.Key) {
						return object.NewError(object.ERR_ARGUMENTS, fnName, "Duplicate keys in adding hashes (Use append to overwrite)")
					}
					err := hash.WritePair(kv.Key, kv.Value)
					if err != nil {
						return object.NewError(object.ERR_GENERAL, fnName, err.Error())
					}
				}
			}

			return hash
		}

		return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected list, hash, or string for first argument")
	},
}

var bi_reverse = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:        "reverse",
		Description: "returns the reversed list or range",

		ParamPositional: []object.Parameter{
			object.Parameter{ExternalName: "over"},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		const fnName = "reverse"

		switch over := args[0].(type) {
		case *object.List:
			return &object.List{Elements: object.CopyAndReverseSlice(over.Elements)}

		case *object.Hash:
			// reverse keys/values of hash if possible
			hash, err := over.Reverse()
			if err != nil {
				return object.NewError(object.ERR_ARGUMENTS, fnName, err.Error())
			}
			return hash

		case *object.String:
			return object.NewString(str.Reverse(over.String()))

		case *object.Range:
			return &object.Range{Start: over.End, End: over.Start}

		case *object.Number:
			return over.Reverse()

		default:
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected list, hash, string, or range")
		}
	},
}

var bi_rotate = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:        "rotate",
		Description: "rotates list elements or a number within a range",

		ParamPositional: []object.Parameter{
			object.Parameter{ExternalName: "over"},
		},

		ParamByName: []object.Parameter{
			object.Parameter{ExternalName: "distance"},
			object.Parameter{ExternalName: "range"},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		const fnName = "rotate"

		distance := 1
		if args[1] != nil {
			var ok bool
			distance, ok = object.NumberToInt(args[1])
			if !ok {
				return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected integer distance to rotate")
			}
		}

		theRange := args[2]
		var rng *object.Range
		var ok bool
		const expectedRng = "Expected integer range to rotate number within"
		if theRange != nil {
			rng, ok = theRange.(*object.Range)
			if !ok {
				return object.NewError(object.ERR_ARGUMENTS, fnName, expectedRng)
			}
		}

		switch over := args[0].(type) {
		case *object.List:
			if theRange != nil {
				return object.NewError(object.ERR_ARGUMENTS, fnName, "unexpected range argument")
			}
			distance = determineRotation(distance, len(over.Elements))
			if distance == 0 {
				return over
			}
			newSlcL := object.CopySlice(over.Elements[distance:])
			newSlcR := object.CopySlice(over.Elements[:distance])
			return &object.List{Elements: append(newSlcL, newSlcR...)}

		case *object.String:
			if theRange != nil {
				return object.NewError(object.ERR_ARGUMENTS, fnName, "unexpected range argument")
			}

			// TODO: determine rotation on strings; see notes on reverse() for string by Unicode rules
			// rotate by grapheme instead?
			return object.NewError(object.ERR_ARGUMENTS, fnName, "string rotate not developed yet (by code point or grapheme?)")

			// simple rotation on string code points (not graphemes)
			// 	rotation = determineRotation(rotation, arg.LenCP())
			// 	if rotation == 0 {
			// 		return arg
			// 	}
			// 	codePoints := []rune(arg.Value)
			// 	newSlcL := codePoints[rotation:]
			// 	newSlcR := codePoints[:rotation]
			// 	return object.StringFromCpSlice(append(newSlcL, newSlcR...))

		case *object.Number:
			var start, end int
			if theRange != nil {
				start, ok = object.NumberToInt(rng.Start)
				if !ok {
					return object.NewError(object.ERR_ARGUMENTS, fnName, expectedRng)
				}
				end, ok = object.NumberToInt(rng.End)
				if !ok {
					return object.NewError(object.ERR_ARGUMENTS, fnName, expectedRng)
				}
				if start > end {
					start, end = end, start
				}
				distance = determineRotation(distance, end-start+1)
				if distance == 0 {
					return over
				}

			} else {
				return object.NewError(object.ERR_ARGUMENTS, fnName, expectedRng)
			}

			i, ok := object.NumberToInt(over)
			if !ok {
				return object.NewError(object.ERR_ARGUMENTS, fnName, "expected integer")
			}
			if i < start || i > end {
				// number outside the range passed through
				return over
			}
			i -= distance
			if i < start {
				i = i + end - start + 1
			}
			return object.NumberFromInt(i)

		default:
			// return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected list or string for first argument")
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected list for first argument")
		}
	},
}

func determineRotation(rotatation, length int) int {
	// accounts for negative and/or excessive rotation numbers
	if length < 2 {
		return 0
	}
	isNeg := false
	if rotatation < 0 {
		rotatation = -rotatation
		isNeg = true
	}
	if rotatation > length {
		rotatation = rotatation % length
	}
	if isNeg {
		return length - rotatation
	}
	return rotatation
}
