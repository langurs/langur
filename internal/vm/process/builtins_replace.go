// builtins_replace.go

package process

import (
	"langur/cpoint"
	"fmt"
	"langur/object"
	"langur/str"
	"strings"
)

// replace, tran

// for both regex and plain string replacements
// replacement as string or function taking one parameter
// for progressive and single replacements
var bi_replace = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:        "replace",
		Description: "accepts string or regex for find, and replaces portion of string with given replacement string",

		ParamPositional: []object.Parameter{
			object.Parameter{ExternalName: "anything"},
		},

		ParamByName: []object.Parameter{
			object.Parameter{ExternalName: "by", Required: true},
			object.Parameter{ExternalName: "with", DefaultValue: object.ZLS},
			object.Parameter{ExternalName: "max", DefaultValue: object.IndicatorNoMax, Type: object.NUMBER_OBJ},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		const fnName = "replace"

		var find *object.String
		var ok bool

		src := object.ToString(args[0])

		re, isRegex := args[1].(*object.Regex)
		if !isRegex {
			find, ok = args[1].(*object.String)
			if !ok {
				return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected string or regex for argument by")
			}
		}

		var fn object.Object
		var fns []object.Object

		var replacementString string
		switch repl := args[2].(type) {
		case *object.String:
			replacementString = repl.String()

		case *object.CompiledCode, *object.BuiltIn:
			fn = repl

		case *object.List:
			fns = repl.Elements
			if len(fns) == 0 {
				return object.NewError(object.ERR_ARGUMENTS, fnName,
					"Expected string or function or list of functions for argument with")
			}
			for i, f := range fns {
				if f == object.NONE {
					fns[i] = nil
				} else if !object.IsCallable(f) && f.Type() != object.STRING_OBJ {
					return object.NewError(object.ERR_ARGUMENTS, fnName, fmt.Sprintf("List element %d not callable, no-op, or string", i+1))
				}
			}
			fn = fns[0] // initialiaze fn to first function

		default:
			return object.NewError(object.ERR_ARGUMENTS, fnName,
				"Expected string or function or list of functions for argument with")
		}

		max, err := args[3].(*object.Number).ToInt()
		if err != nil {
			return object.NewError(object.ERR_ARGUMENTS, fnName, err.Error())
		}

		if isRegex {
			if fn == nil && fns == nil {
				result, err := object.RegexReplace(src.String(), re, replacementString, max)
				if err != nil {
					return object.NewError(object.ERR_GENERAL, fnName, err.Error())
				}
				return result
			}
			return regexReplaceWithFunctionsAndStrings(pr, src.String(), re, fn, fns, max)
		}

		if fn == nil && fns == nil {
			return object.NewString(strings.Replace(src.String(), find.String(), replacementString, max))
		}
		return stringReplaceWithFunctionsAndStrings(pr, src.String(), find.String(), fn, fns, max)
	},
}

func regexReplaceWithFunctionsAndStrings(
	pr *Process, src string, re *object.Regex,
	fn object.Object, fns []object.Object, max int) object.Object {

	const fnName = "replace"

	arr, err := object.RegexSplitAndKeep(re, src, max)
	if err != nil {
		return object.NewError(object.ERR_GENERAL, fnName, err.Error())
	}

	return replaceWithFunctionsAndStrings(pr, arr.(*object.List).Elements, re, fn, fns)
}

func stringReplaceWithFunctionsAndStrings(
	pr *Process, src, find string,
	fn object.Object, fns []object.Object, max int) object.Object {

	const fnName = "replace"

	arr, err := object.StringSplitAndKeep(find, src, max)
	if err != nil {
		return object.NewError(object.ERR_GENERAL, fnName, err.Error())
	}
	return replaceWithFunctionsAndStrings(pr, arr.(*object.List).Elements, nil, fn, fns)
}

func replaceWithFunctionsAndStrings(
	pr *Process, elements []object.Object, re *object.Regex,
	fn object.Object, fns []object.Object) object.Object {

	const fnName = "replace"

	fnn := 0
	for i := 1; i < len(elements); i += 2 {
		switch fn.(type) {
		case nil:
			// no-op

		case *object.String:
			if re == nil {
				elements[i] = fn
			} else {
				// TODO: To account for submatch interpolation ($1, etc.)
				return object.NewError(object.ERR_ARGUMENTS, fnName, "Current implementation unable to use strings for multiple replacements with regex")
			}

		default:
			result, err := pr.callback(fn, elements[i])
			if err != nil {
				return object.NewError(object.ERR_GENERAL, fnName, err.Error())
			}
			elements[i] = result
		}

		if fns != nil {
			// next function
			if fnn > len(fns)-2 {
				fnn = 0
			} else {
				fnn++
			}
			fn = fns[fnn]
		}
	}

	return object.StringConcat(elements)
}

var bi_tran = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:        "tran",
		Description: "transliteration by strings, code points, or graphemes; may use lists; or may use single hash for argument with",

		ParamPositional: []object.Parameter{
			object.Parameter{ExternalName: "anything"},
		},

		ParamByName: []object.Parameter{
			object.Parameter{ExternalName: "by"},
			object.Parameter{ExternalName: "with", Required: true},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		const fnName = "tran"

		src := args[0].String()
		var sb strings.Builder

		var list1, list2 []string
		var err error

		if args[1] == nil {
			// no argument by; with must be hash
			h, ok := args[2].(*object.Hash)
			if ok {
				keys, values := h.IndexKeys(), h.Values()
				list1, err = listToGraphemeStringSlice(keys)
				if err != nil {
					return object.NewError(object.ERR_ARGUMENTS, fnName, fmt.Sprintf("Error on keys of hash argument with: %s", err.Error()))
				}
				list2, err = listToGraphemeStringSlice(values)
				if err != nil {
					return object.NewError(object.ERR_ARGUMENTS, fnName, fmt.Sprintf("Error on values of hash argument with: %s", err.Error()))
				}

			} else {
				return object.NewError(object.ERR_ARGUMENTS, fnName, "Argument with must be hash when argument by not passed")
			}

		} else {
			list1, err = listToGraphemeStringSlice(args[1])
			if err != nil {
				return object.NewError(object.ERR_ARGUMENTS, fnName, fmt.Sprintf("Error on argument by: %s", err.Error()))
			}
			list2, err = listToGraphemeStringSlice(args[2])
			if err != nil {
				return object.NewError(object.ERR_ARGUMENTS, fnName, fmt.Sprintf("Error on argument with: %s", err.Error()))
			}
		}

		// should have 2 lists of equal length
		if len(list1) != len(list2) {
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected same number of items for lists in arguments by and with")
		}

		for len(src) != 0 {
			match := false
			for i, find := range list1 {
				if len(src) >= len(find) && src[:len(find)] == find {
					match = true
					sb.WriteString(list2[i])
					src = src[len(find):]
					break // for i, find
				}
			}
			if !match {
				// no match; advance by 1 code point
				r, bc, _ := cpoint.Decode(&src, 0)
				sb.WriteRune(r)
				src = src[bc:]
			}
		}
		
		return object.NewString(sb.String())
	},
}

func listToGraphemeStringSlice(left object.Object) (elements []string, err error) {
	switch left := left.(type) {
	case *object.String:
		elements = str.GraphemeStringSlice(left.String())
		
	case *object.Range:
		rSlc, err := object.CodePointsToFlatRuneSlice(left)
		if err != nil {
			return nil, err
		}
		for _, r := range rSlc {
			elements = append(elements, string(r))
		}
	
	case *object.List:
		for _, e := range left.Elements {
			switch e.(type) {
			case *object.String:
				elements = append(elements, e.String())		

			default:
				rSlc, err := object.CodePointsToFlatRuneSlice(e)
				if err != nil {
					return nil, err
				}
				elements = append(elements, string(rSlc))		
			}
		}

	default:
		err = fmt.Errorf("Expected string, list, or range")
	}
	
	return
}
