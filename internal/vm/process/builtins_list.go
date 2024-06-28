// langur/vm/process/builtins_list.go

package process

import (
	"langur/object"
	"langur/str"
	"strings"
)

// reverse, rotate
// less, more, rest

func bi_less(pr *Process, args ...object.Object) object.Object {
	const fnName = "less"

	switch arg := args[0].(type) {
	case *object.List:
		if len(arg.Elements) == 0 {
			return arg
		}
		if len(args) == 1 {
			// 1 less element
			newElements := make([]object.Object, len(arg.Elements)-1)
			copy(newElements, arg.Elements[:len(arg.Elements)-1])
			return &object.List{Elements: newElements}
		}
		// remove specific indices
		list, err := arg.RemoveIndices(args[1])
		if err != nil {
			return object.NewError(object.ERR_GENERAL, fnName, "Error removing indices from list: "+err.Error())
		}
		return list

	case *object.String:
		if arg.LenCP() == 0 {
			return arg
		}
		if len(args) == 1 {
			// 1 less code point
			cpSlc := arg.RuneSlc()
			s, err := object.NewStringFromParts(cpSlc[:len(cpSlc)-1])
			if err == nil {
				return s
			}
			return object.NewError(object.ERR_GENERAL, fnName, "Error removing indices from list: "+err.Error())
		}
		str, err := arg.RemoveIndices(args[1])
		if err != nil {
			return object.NewError(object.ERR_GENERAL, fnName, "Error removing indices from string: "+err.Error())
		}
		return str

	case *object.Hash:
		if len(args) < 2 {
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected second argument for hash (key or list of keys)")
		}
		if len(arg.Pairs) == 0 {
			return arg
		}
		hash, err := arg.RemoveKeys(args[1])
		if err != nil {
			return object.NewError(object.ERR_GENERAL, fnName, "Error removing keys from hash: "+err.Error())
		}
		return hash
	}

	return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected list, hash or string for first argument")
}

func bi_more(pr *Process, args ...object.Object) object.Object {
	const fnName = "more"

	if len(args) == 1 {
		return args[0]
	}

	switch arg := args[0].(type) {
	case *object.List:
		newElements := make([]object.Object, len(arg.Elements)+len(args)-1)
		copy(newElements, arg.Elements)
		copy(newElements[len(arg.Elements):], args[1:])
		return &object.List{Elements: newElements}

	case *object.String:
		ns := &strings.Builder{}
		ns.WriteString(arg.String())
		for i := 1; i < len(args); i++ {
			// items may be strings or code points (integers)
			switch item := args[i].(type) {
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
		hash := arg

		for i := 1; i < len(args); i++ {
			from, ok := args[i].(*object.Hash)
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
}

func bi_rest(pr *Process, args ...object.Object) object.Object {
	switch arg := args[0].(type) {
	case *object.List:
		if len(arg.Elements) == 0 {
			return arg
		}
		newElements := make([]object.Object, len(arg.Elements)-1)
		copy(newElements, arg.Elements[1:])
		return &object.List{Elements: newElements}

	case *object.String:
		if arg.LenCP() == 0 {
			return arg
		}
		cpSlc := arg.RuneSlc()
		return object.NewString(string(cpSlc[1:]))
	}

	return object.NewError(object.ERR_ARGUMENTS, "rest", "Expected list or string")
}

func bi_reverse(pr *Process, args ...object.Object) object.Object {
	const fnName = "reverse"

	switch arg := args[0].(type) {
	case *object.List:
		return &object.List{Elements: object.CopyAndReverseSlice(arg.Elements)}

	case *object.Hash:
		// reverse keys/values of hash if possible
		hash, err := arg.Reverse()
		if err != nil {
			return object.NewError(object.ERR_ARGUMENTS, fnName, err.Error())
		}
		return hash

	case *object.String:
		return object.NewString(str.Reverse(arg.String()))

	case *object.Range:
		return &object.Range{Start: arg.End, End: arg.Start}

	case *object.Number:
		return arg.Reverse()

	default:
		return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected list, hash, string, or range")
	}
}

func bi_rotate(pr *Process, args ...object.Object) object.Object {
	const fnName = "rotate"

	rotation := 1
	if len(args) > 1 {
		var ok bool
		rotation, ok = object.NumberToInt(args[1])
		if !ok {
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected integer distance to rotate for second argument")
		}
	}

	switch arg := args[0].(type) {
	case *object.List:
		if len(args) > 2 {
			return object.NewError(object.ERR_ARGUMENTS, fnName, "unexpected third argument")
		}
		rotation = determineRotation(rotation, len(arg.Elements))
		if rotation == 0 {
			return arg
		}
		newSlcL := object.CopySlice(arg.Elements[rotation:])
		newSlcR := object.CopySlice(arg.Elements[:rotation])
		return &object.List{Elements: append(newSlcL, newSlcR...)}

	case *object.String:
		if len(args) > 2 {
			return object.NewError(object.ERR_ARGUMENTS, fnName, "unexpected third argument")
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
		const expectedRng = "Expected range to rotate number within for third argument"

		var start, end int
		if len(args) > 2 {
			rng, ok := args[2].(*object.Range)
			if !ok {
				return object.NewError(object.ERR_ARGUMENTS, fnName, expectedRng)
			}
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
			rotation = determineRotation(rotation, end-start+1)
			if rotation == 0 {
				return arg
			}

		} else {
			return object.NewError(object.ERR_ARGUMENTS, fnName, expectedRng)
		}

		i, ok := object.NumberToInt(arg)
		if !ok {
			return object.NewError(object.ERR_ARGUMENTS, fnName, "expected integer")
		}
		if i < start || i > end {
			// number outside the range passed through
			return arg
		}
		i -= rotation
		if i < start {
			i = i + end - start + 1
		}
		return object.NumberFromInt(i)

	default:
		// return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected list or string for first argument")
		return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected list for first argument")
	}
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
