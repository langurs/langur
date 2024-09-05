// langur/vm/process/builtins_regex.go

package process

import (
	"langur/object"
	"langur/str"
	"strings"
)

// Note: Several of these are not purely regex functions (may accept a plain string).

// matching, match, submatch, index, subindex
// matches, submatches, indices, subindices
// split, splitd

var bi_matching = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:        "matching",
		Description: "accepts compiled regex and returns Boolean indicating whether the string matches the pattern",

		ParamPositional: []object.Parameter{
			object.Parameter{ExternalName: "anything"},
		},

		ParamByName: []object.Parameter{
			object.Parameter{ExternalName: "by", Required: true},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		const fnName = "matching"

		var check *object.String
		var ok bool

		re, isRegex := args[1].(*object.Regex)
		if !isRegex {
			check, ok = args[1].(*object.String)
			if !ok {
				return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected string or regex for matching by argument")
			}
		}
		s := object.ToString(args[0])

		if isRegex {
			success, err := object.RegexMatching(re, s.String())
			if err != nil {
				return object.NewError(object.ERR_GENERAL, fnName, err.Error())
			}
			return success
		}

		return object.NativeBoolToObject(strings.Contains(s.String(), check.String()))
	},
}

var bi_match = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:        "match",
		Description: "accepts compiled regex and returns matching string, or returns null or alternate value (optional) for no match",

		ParamPositional: []object.Parameter{
			object.Parameter{ExternalName: "by"},
			object.Parameter{ExternalName: "anything"},
		},

		ParamByName: []object.Parameter{
			object.Parameter{ExternalName: "alt"},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		const fnName = "match"

		re, ok := args[0].(*object.Regex)
		if !ok {
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected regex for match by argument")
		}
		s := object.ToString(args[1])

		result, err := object.RegexMatchOnce(re, s.String())
		if err != nil {
			return object.NewError(object.ERR_GENERAL, fnName, err.Error())
		}
		if result == object.NONE && args[2] != nil {
			// no match
			// return alternate value
			return args[2]
		}
		return result
	},
}

var bi_matches = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:        "matches",
		Description: "matches(regex, anything, max); accepts compiled regex and returns list of progressive matches (empty list if no matches); max optional (defaults to -1 meaning infinite)",

		ParamPositional: []object.Parameter{
			object.Parameter{ExternalName: "by"},
			object.Parameter{ExternalName: "anything"},
		},

		ParamByName: []object.Parameter{
			object.Parameter{ExternalName: "max", DefaultValue: object.NegOne},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		const fnName = "matches"

		re, ok := args[0].(*object.Regex)
		if !ok {
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected regex for match by argument")
		}
		s := object.ToString(args[1])

		var err error
		count, ok := args[2].(*object.Number)
		if !ok {
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected integer for max count argument")
		}
		cnt, err := count.ToInt()
		if err != nil {
			return object.NewError(object.ERR_ARGUMENTS, fnName, err.Error())
		}

		arr, err := object.RegexMatchProgressive(re, s.String(), cnt)
		if err != nil {
			return object.NewError(object.ERR_GENERAL, fnName, err.Error())
		}
		return arr
	},
}

var bi_submatch = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:        "submatch",
		Description: "submatch(regex, anything); returns list of submatches (empty list if not a match)",

		// TODO: update
		ParamPositional: []object.Parameter{
			object.Parameter{},
			object.Parameter{},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		const fnName = "submatch"

		re, ok := args[0].(*object.Regex)
		if !ok {
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected regex for first argument")
		}
		s := object.ToString(args[1])

		result, err := object.RegexSubMatches(re, s.String())
		if err != nil {
			return object.NewError(object.ERR_GENERAL, fnName, err.Error())
		}
		return result
	},
}

var bi_submatchH = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:        "submatchH",
		Description: "submatchH(regex, anything); returns hash of submatches (empty hash if not a match)",

		// TODO: update
		ParamPositional: []object.Parameter{
			object.Parameter{},
			object.Parameter{},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		const fnName = "submatchH"

		re, ok := args[0].(*object.Regex)
		if !ok {
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected regex for first argument")
		}
		s := object.ToString(args[1])

		result, err := object.RegexSubMatchesHash(re, s.String())
		if err != nil {
			return object.NewError(object.ERR_GENERAL, fnName, err.Error())
		}
		return result
	},
}

var bi_submatches = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:        "submatches",
		Description: "submatches(regex, anything, max); returns list of lists of progressive submatches (empty list if not a match); max optional (defaults to -1 meaning infinite)",

		// TODO: update
		ParamPositional: []object.Parameter{
			object.Parameter{},
		},
		ParamExpansionMin: 2,
		ParamExpansionMax: 3,
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		const fnName = "submatches"

		// FIXME: update parameters/args
		args = args[0].(*object.List).Elements

		re, ok := args[0].(*object.Regex)
		if !ok {
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected regex for first argument")
		}
		s := object.ToString(args[1])

		cnt := -1
		if len(args) > 2 {
			count, ok := args[2].(*object.Number)
			if !ok {
				return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected integer for third argument")
			}
			var err error
			cnt, err = count.ToInt()
			if err != nil {
				return object.NewError(object.ERR_ARGUMENTS, fnName, err.Error())
			}
		}

		result, err := object.RegexProgressiveSubMatches(re, s.String(), cnt)
		if err != nil {
			return object.NewError(object.ERR_GENERAL, fnName, err.Error())
		}
		return result
	},
}

var bi_submatchesH = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:        "submatchesH",
		Description: "submatchesH(regex, anything, max); returns list of hashes of progressive whole match and submatches (empty list if not a match); max optional (defaults to -1 meaning infinite)",

		// TODO: update
		ParamPositional: []object.Parameter{
			object.Parameter{},
		},
		ParamExpansionMin: 2,
		ParamExpansionMax: 3,
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		const fnName = "submatchesH"

		// FIXME: update parameters/args
		args = args[0].(*object.List).Elements

		re, ok := args[0].(*object.Regex)
		if !ok {
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected regex for first argument")
		}
		s := object.ToString(args[1])

		cnt := -1
		if len(args) > 2 {
			count, ok := args[2].(*object.Number)
			if !ok {
				return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected integer for third argument")
			}
			var err error
			cnt, err = count.ToInt()
			if err != nil {
				return object.NewError(object.ERR_ARGUMENTS, fnName, err.Error())
			}
		}

		result, err := object.RegexProgressiveSubMatchesHashList(re, s.String(), cnt)
		if err != nil {
			return object.NewError(object.ERR_GENERAL, fnName, err.Error())
		}
		return result
	},
}

var bi_split = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:        "split",
		Description: "split(delim, anything, max); accepts regex or string delimiter and splits string into a list of strings; max optional",

		// TODO: update
		ParamPositional: []object.Parameter{
			object.Parameter{},
		},
		ParamExpansionMin: 1,
		ParamExpansionMax: 3,
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		const fnName = "split"

		// FIXME: update parameters/args
		args = args[0].(*object.List).Elements

		// default delimiter as ZLS
		var delim, s string
		var countEach int
		var isRegex, isCountEach bool
		var re *object.Regex

		if len(args) == 1 {
			// 1 argument, split into single code point strings
			s = args[0].String()

		} else {
			// check for regex/string/integer count to split by
			switch args[0].(type) {
			case *object.Regex:
				re, isRegex = args[0].(*object.Regex)
			case *object.String:
				delim = args[0].String()
			case *object.Number:
				countEach, isCountEach = object.NumberToInt(args[0])
				if !isCountEach {
					return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected string, regex, or integer for first argument")
				}
			default:
				return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected string, regex, or integer for first argument")
			}
			s = args[1].String()
		}

		max := -1
		if len(args) > 2 {
			count, ok := args[2].(*object.Number)
			if !ok {
				return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected integer for third argument")
			}
			var err error
			max, err = count.ToInt()
			if err != nil {
				return object.NewError(object.ERR_GENERAL, fnName, err.Error())
			}
		}

		if isRegex {
			result, err := object.RegexSplit(re, s, max)
			if err != nil {
				return object.NewError(object.ERR_GENERAL, fnName, err.Error())
			}
			return result

		} else if isCountEach {
			sSlc, err := str.SplitByNumber(s, countEach, max)
			if err != nil {
				return object.NewError(object.ERR_GENERAL, fnName, err.Error())
			}
			return object.StringSliceToList(sSlc)

		} else {
			return object.StringSliceToList(strings.SplitN(s, delim, max))
		}
	},
}

var bi_index = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:        "index",
		Description: "index(regex, anything, alternate); accepts regex and returns code point range for match, or returns null or alternate value (optional) for no match",

		// TODO: update
		ParamPositional: []object.Parameter{
			object.Parameter{},
		},
		ParamExpansionMin: 2,
		ParamExpansionMax: 3,
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		const fnName = "index"

		// FIXME: update parameters/args
		args = args[0].(*object.List).Elements

		var sub *object.String
		var ok bool

		re, isRegex := args[0].(*object.Regex)
		if !isRegex {
			sub, ok = args[0].(*object.String)
			if !ok {
				return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected string or regex for first argument")
			}
		}
		s := object.ToString(args[1])

		var result object.Object
		var err error

		if isRegex {
			result, err = object.RegexIndex(re, s.String())
		} else {
			result, err = object.StringIndex(sub.String(), s.String())
		}

		if err != nil {
			return object.NewError(object.ERR_GENERAL, fnName, err.Error())
		}
		if result == object.NONE && len(args) > 2 {
			// no match
			// return alternate value
			return args[2]
		}
		return result
	},
}

var bi_indices = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:        "indices",
		Description: `indices(regex, anything, max); accepts regex and returns list of code point ranges for progressive matches (a.k.a. "global"), or empty list for no match; max optional`,

		// TODO: update
		ParamPositional: []object.Parameter{
			object.Parameter{},
		},
		ParamExpansionMin: 2,
		ParamExpansionMax: 3,
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		const fnName = "indices"

		// FIXME: update parameters/args
		args = args[0].(*object.List).Elements

		var sub *object.String
		var ok bool

		re, isRegex := args[0].(*object.Regex)
		if !isRegex {
			sub, ok = args[0].(*object.String)
			if !ok {
				return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected string or regex for first argument")
			}
		}
		s := object.ToString(args[1])

		cnt := -1
		if len(args) > 2 {
			count, ok := args[2].(*object.Number)
			if !ok {
				return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected integer for third argument")
			}
			var err error
			cnt, err = count.ToInt()
			if err != nil {
				return object.NewError(object.ERR_GENERAL, fnName, err.Error())
			}
		}

		var result object.Object
		var err error

		if isRegex {
			result, err = object.RegexProgressiveIndices(re, s.String(), cnt)
		} else {
			result, err = object.StringProgressiveIndices(sub.String(), s.String(), cnt)
		}
		if err != nil {
			return object.NewError(object.ERR_GENERAL, fnName, err.Error())
		}
		return result
	},
}

var bi_subindex = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:        "subindex",
		Description: `subindex(regex, anything); accepts regex and returns list of code point ranges for submatches, or empty list for no match`,

		// TODO: update
		ParamPositional: []object.Parameter{
			object.Parameter{},
			object.Parameter{},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		const fnName = "subindex"

		re, ok := args[0].(*object.Regex)
		if !ok {
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected regex for first argument")
		}
		s := object.ToString(args[1])

		result, err := object.RegexSubMatchesIndices(re, s.String())
		if err != nil {
			return object.NewError(object.ERR_GENERAL, fnName, err.Error())
		}
		return result
	},
}

var bi_subindices = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:        "subindices",
		Description: `subindices(regex, anything, max); accepts regex and returns list of lists of code point ranges for progressive submatches (a.k.a. "global"), or empty list for no match; max optional`,

		// TODO: update
		ParamPositional: []object.Parameter{
			object.Parameter{},
		},
		ParamExpansionMin: 2,
		ParamExpansionMax: 3,
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		const fnName = "subindices"

		// FIXME: update parameters/args
		args = args[0].(*object.List).Elements

		re, ok := args[0].(*object.Regex)
		if !ok {
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected regex for first argument")
		}
		s := object.ToString(args[1])

		cnt := -1
		if len(args) > 2 {
			count, ok := args[2].(*object.Number)
			if !ok {
				return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected integer for third argument")
			}
			var err error
			cnt, err = count.ToInt()
			if err != nil {
				return object.NewError(object.ERR_GENERAL, fnName, err.Error())
			}
		}

		result, err := object.RegexProgressiveSubMatchesIndices(re, s.String(), cnt)
		if err != nil {
			return object.NewError(object.ERR_GENERAL, fnName, err.Error())
		}
		return result
	},
}
