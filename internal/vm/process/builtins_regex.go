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

// returns Boolean object
func bi_matching(pr *Process, args ...object.Object) object.Object {
	const fnName = "matching"

	var check *object.String
	var ok bool

	re, isRegex := args[0].(*object.Regex)
	if !isRegex {
		check, ok = args[0].(*object.String)
		if !ok {
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected string or regex for first argument")
		}
	}
	s := object.ToString(args[1])

	if isRegex {
		success, err := object.RegexMatching(re, s.String())
		if err != nil {
			return object.NewError(object.ERR_GENERAL, fnName, err.Error())
		}
		return success
	}

	return object.NativeBoolToObject(strings.Contains(s.String(), check.String()))
}

// returns match string, or null or 3rd argument (alternate)
func bi_match(pr *Process, args ...object.Object) object.Object {
	const fnName = "match"

	re, ok := args[0].(*object.Regex)
	if !ok {
		return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected regex for first argument")
	}
	s := object.ToString(args[1])

	result, err := object.RegexMatchOnce(re, s.String())
	if err != nil {
		return object.NewError(object.ERR_GENERAL, fnName, err.Error())
	}
	if result == object.NONE && len(args) > 2 {
		// no match
		// return alternate value
		return args[2]
	}
	return result
}

// progressive matching
// returns list of matches or empty list for no match
func bi_matches(pr *Process, args ...object.Object) object.Object {
	const fnName = "matches"

	re, ok := args[0].(*object.Regex)
	if !ok {
		return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected regex for first argument")
	}
	s := object.ToString(args[1])

	cnt := -1 // -1 == infinite (practically)
	if len(args) > 2 {
		var err error
		count, ok := args[2].(*object.Number)
		if !ok {
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected integer for third argument")
		}
		cnt, err = count.ToInt()
		if err != nil {
			return object.NewError(object.ERR_ARGUMENTS, fnName, err.Error())
		}
	}

	arr, err := object.RegexMatchProgressive(re, s.String(), cnt)
	if err != nil {
		return object.NewError(object.ERR_GENERAL, fnName, err.Error())
	}
	return arr
}

// returns single list of submatches or empty list for no match
func bi_submatch(pr *Process, args ...object.Object) object.Object {
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
}

// returns single hash of submatches or empty hash for no match
func bi_submatchH(pr *Process, args ...object.Object) object.Object {
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
}

// returns list of lists of submatches or empty list for no match
func bi_submatches(pr *Process, args ...object.Object) object.Object {
	const fnName = "submatches"

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
}

// returns list of hashes of submatches or empty list for no match
func bi_submatchesH(pr *Process, args ...object.Object) object.Object {
	const fnName = "submatchesH"

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
}

// for both regex and plain string delimiters
func bi_split(pr *Process, args ...object.Object) object.Object {
	const fnName = "split"

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
}

func bi_index(pr *Process, args ...object.Object) object.Object {
	const fnName = "index"

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
}

func bi_indices(pr *Process, args ...object.Object) object.Object {
	const fnName = "indices"

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
}

func bi_subindex(pr *Process, args ...object.Object) object.Object {
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
}

func bi_subindices(pr *Process, args ...object.Object) object.Object {
	const fnName = "subindices"

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
}
