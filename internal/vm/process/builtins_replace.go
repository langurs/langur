// builtins_replace.go

package process

import (
	"fmt"
	"langur/object"
	"strings"
)

// replace

// for both regex and plain string replacements
// replacement as string or function taking one parameter
// for progressive and single replacements
var bi_replace = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:        "replace",
		Description: "replace(source, find, replace, max); accepts string or regex for find, and replaces portion of string with given replacement string; max optional",

		// TODO: update
		ParamPositional: []object.Parameter{
			object.Parameter{},
		},
		ParamExpansionMin: 2,
		ParamExpansionMax: 4,
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
				return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected string or regex for second argument (find)")
			}
		}

		max := -1
		var fn object.Object
		var fns []object.Object

		// defaulting to ZLS (replace with nothing)
		var replacementString string

		if len(args) > 2 {
			switch repl := args[2].(type) {
			case *object.String:
				replacementString = repl.String()

			case *object.CompiledCode, *object.BuiltIn:
				fn = repl

			case *object.List:
				fns = repl.Elements
				if len(fns) == 0 {
					return object.NewError(object.ERR_ARGUMENTS, fnName,
						"Expected string or function or list of functions for third argument (replacement)")
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
					"Expected string or function or list of functions for third argument (replacement)")
			}

			if len(args) > 3 {
				count, ok := args[3].(*object.Number)
				if !ok {
					return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected integer for fourth argument")
				}
				var err error
				max, err = count.ToInt()
				if err != nil {
					return object.NewError(object.ERR_ARGUMENTS, fnName, err.Error())
				}
			}
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
