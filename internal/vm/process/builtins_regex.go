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
// split

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

		s := args[0].String()

		re, isRegex := args[1].(*object.Regex)
		if !isRegex {
			check, ok = args[1].(*object.String)
			if !ok {
				return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected string or regex for argument by")
			}
		}

		if isRegex {
			success, err := object.RegexMatching(re, s)
			if err != nil {
				return object.NewError(object.ERR_GENERAL, fnName, err.Error())
			}
			return success
		}

		return object.NativeBoolToObject(strings.Contains(s, check.String()))
	},
}

var bi_match = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:        "match",
		Description: "accepts compiled regex and returns matching string, or returns null or alternate value (optional) for no match",

		ParamPositional: []object.Parameter{
			object.Parameter{ExternalName: "anything"},
		},

		ParamByName: []object.Parameter{
			object.Parameter{ExternalName: "by", Required: true, Type: object.REGEX_OBJ},
			object.Parameter{ExternalName: "alt"},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		const fnName = "match"

		s := args[0].String()
		re := args[1].(*object.Regex)

		result, err := object.RegexMatchOnce(re, s)
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
		Description: "accepts compiled regex and returns list of progressive matches (empty list if no matches)",

		ParamPositional: []object.Parameter{
			object.Parameter{ExternalName: "anything"},
		},

		ParamByName: []object.Parameter{
			object.Parameter{ExternalName: "by", Required: true, Type: object.REGEX_OBJ},
			object.Parameter{ExternalName: "max", DefaultValue: object.IndicatorNoMax, Type: object.NUMBER_OBJ},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		const fnName = "matches"

		s := args[0].String()
		re := args[1].(*object.Regex)

		max, err := args[2].(*object.Number).ToInt()
		if err != nil {
			return object.NewError(object.ERR_ARGUMENTS, fnName, err.Error())
		}

		list, err := object.RegexMatchProgressive(re, s, max)
		if err != nil {
			return object.NewError(object.ERR_GENERAL, fnName, err.Error())
		}
		return list
	},
}

var bi_submatch = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:        "submatch",
		Description: "returns list of submatches (empty list if not a match)",

		ParamPositional: []object.Parameter{
			object.Parameter{ExternalName: "anything"},
		},

		ParamByName: []object.Parameter{
			object.Parameter{ExternalName: "by", Required: true, Type: object.REGEX_OBJ},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		const fnName = "submatch"

		s := args[0].String()
		re := args[1].(*object.Regex)

		result, err := object.RegexSubMatches(re, s)
		if err != nil {
			return object.NewError(object.ERR_GENERAL, fnName, err.Error())
		}
		return result
	},
}

var bi_submatchH = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:        "submatchH",
		Description: "returns hash of submatches (empty hash if not a match)",

		ParamPositional: []object.Parameter{
			object.Parameter{ExternalName: "anything"},
		},

		ParamByName: []object.Parameter{
			object.Parameter{ExternalName: "by", Required: true, Type: object.REGEX_OBJ},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		const fnName = "submatchH"

		s := args[0].String()
		re := args[1].(*object.Regex)

		result, err := object.RegexSubMatchesHash(re, s)
		if err != nil {
			return object.NewError(object.ERR_GENERAL, fnName, err.Error())
		}
		return result
	},
}

var bi_submatches = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:        "submatches",
		Description: "returns list of lists of progressive submatches (empty list if not a match)",

		ParamPositional: []object.Parameter{
			object.Parameter{ExternalName: "anything"},
		},

		ParamByName: []object.Parameter{
			object.Parameter{ExternalName: "by", Required: true, Type: object.REGEX_OBJ},
			object.Parameter{ExternalName: "max", DefaultValue: object.IndicatorNoMax, Type: object.NUMBER_OBJ},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		const fnName = "submatches"

		s := args[0].String()
		re := args[1].(*object.Regex)

		cnt, err := args[2].(*object.Number).ToInt()
		if err != nil {
			return object.NewError(object.ERR_ARGUMENTS, fnName, err.Error())
		}

		result, err := object.RegexProgressiveSubMatches(re, s, cnt)
		if err != nil {
			return object.NewError(object.ERR_GENERAL, fnName, err.Error())
		}
		return result
	},
}

var bi_submatchesH = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:        "submatchesH",
		Description: "returns list of hashes of progressive whole match and submatches (empty list if not a match)",

		ParamPositional: []object.Parameter{
			object.Parameter{ExternalName: "anything"},
		},

		ParamByName: []object.Parameter{
			object.Parameter{ExternalName: "by", Required: true, Type: object.REGEX_OBJ},
			object.Parameter{ExternalName: "max", DefaultValue: object.IndicatorNoMax, Type: object.NUMBER_OBJ},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		const fnName = "submatchesH"

		s := args[0].String()
		re := args[1].(*object.Regex)

		cnt, err := args[2].(*object.Number).ToInt()
		if err != nil {
			return object.NewError(object.ERR_ARGUMENTS, fnName, err.Error())
		}

		result, err := object.RegexProgressiveSubMatchesHashList(re, s, cnt)
		if err != nil {
			return object.NewError(object.ERR_GENERAL, fnName, err.Error())
		}
		return result
	},
}

var bi_split = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:        "split",
		Description: "accepts regex or string delimiter and splits anything into a list of strings",

		ParamPositional: []object.Parameter{
			object.Parameter{ExternalName: "anything"},
		},

		ParamByName: []object.Parameter{
			object.Parameter{ExternalName: "delim", DefaultValue: object.ZeroLengthString()},
			object.Parameter{ExternalName: "max", DefaultValue: object.IndicatorNoMax, Type: object.NUMBER_OBJ},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		const fnName = "split"

		// default delimiter as ZLS
		var delim, s string
		var countEach int
		var isRegex, isCountEach bool
		var re *object.Regex

		s = args[0].String()

		// check for regex/string/integer count to split by
		switch by := args[1].(type) {
		case *object.Regex:
			re, isRegex = by, true
		case *object.String:
			delim = by.String()
		case *object.Number:
			countEach, isCountEach = object.NumberToInt(by)
			if !isCountEach {
				return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected string, regex, or integer for argument delim")
			}
		default:
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected string, regex, or integer for argument delim")
		}

		count, ok := args[2].(*object.Number)
		if !ok {
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected integer for argument max")
		}
		max, err := count.ToInt()
		if err != nil {
			return object.NewError(object.ERR_GENERAL, fnName, err.Error())
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
		Description: "accepts regex and returns code point range for match, or returns null or alternate value (optional) for no match",

		ParamPositional: []object.Parameter{
			object.Parameter{ExternalName: "anything"},
		},

		ParamByName: []object.Parameter{
			object.Parameter{ExternalName: "by", Required: true},
			object.Parameter{ExternalName: "alt"},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		const fnName = "index"

		var sub *object.String
		var ok bool

		s := args[0].String()

		re, isRegex := args[1].(*object.Regex)
		if !isRegex {
			sub, ok = args[1].(*object.String)
			if !ok {
				return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected string or regex for argument by")
			}
		}

		var result object.Object
		var err error

		if isRegex {
			result, err = object.RegexIndex(re, s)
		} else {
			result, err = object.StringIndex(sub.String(), s)
		}

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

var bi_indices = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:        "indices",
		Description: `accepts regex and returns list of code point ranges for progressive matches (a.k.a. "global"), or empty list for no match`,

		ParamPositional: []object.Parameter{
			object.Parameter{ExternalName: "anything"},
		},

		ParamByName: []object.Parameter{
			object.Parameter{ExternalName: "by", Required: true},
			object.Parameter{ExternalName: "max", DefaultValue: object.IndicatorNoMax, Type: object.NUMBER_OBJ},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		const fnName = "indices"

		var sub *object.String
		var ok bool

		s := args[0].String()

		re, isRegex := args[1].(*object.Regex)
		if !isRegex {
			sub, ok = args[1].(*object.String)
			if !ok {
				return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected string or regex for argument by")
			}
		}

		cnt, err := args[2].(*object.Number).ToInt()
		if err != nil {
			return object.NewError(object.ERR_GENERAL, fnName, err.Error())
		}

		var result object.Object
		if isRegex {
			result, err = object.RegexProgressiveIndices(re, s, cnt)
		} else {
			result, err = object.StringProgressiveIndices(sub.String(), s, cnt)
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
		Description: `accepts regex and returns list of code point ranges for submatches, or empty list for no match`,

		ParamPositional: []object.Parameter{
			object.Parameter{ExternalName: "anything"},
		},

		ParamByName: []object.Parameter{
			object.Parameter{ExternalName: "by", Required: true, Type: object.REGEX_OBJ},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		const fnName = "subindex"

		s := args[0].String()
		re := args[1].(*object.Regex)

		result, err := object.RegexSubMatchesIndices(re, s)
		if err != nil {
			return object.NewError(object.ERR_GENERAL, fnName, err.Error())
		}
		return result
	},
}

var bi_subindices = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:        "subindices",
		Description: `accepts regex and returns list of lists of code point ranges for progressive submatches (a.k.a. "global"), or empty list for no match`,

		ParamPositional: []object.Parameter{
			object.Parameter{ExternalName: "anything"},
		},

		ParamByName: []object.Parameter{
			object.Parameter{ExternalName: "by", Required: true, Type: object.REGEX_OBJ},
			object.Parameter{ExternalName: "max", DefaultValue: object.IndicatorNoMax, Type: object.NUMBER_OBJ},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		const fnName = "subindices"

		s := args[0].String()
		re := args[1].(*object.Regex)

		cnt, err := args[2].(*object.Number).ToInt()
		if err != nil {
			return object.NewError(object.ERR_GENERAL, fnName, err.Error())
		}

		result, err := object.RegexProgressiveSubMatchesIndices(re, s, cnt)
		if err != nil {
			return object.NewError(object.ERR_GENERAL, fnName, err.Error())
		}
		return result
	},
}
