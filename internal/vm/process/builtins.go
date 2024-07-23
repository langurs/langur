// langur/vm/process/builtins.go

package process

import (
	"langur/common"
	"langur/modes"
	"langur/object"
)

func GetBuiltInByName(name string) *object.BuiltIn {
	for _, bi := range BuiltIns {
		if bi.FnSignature.Name == name {
			return bi
		}
	}
	return nil
}

func GetBuiltInImpurityStatus(name string) bool {
	for _, bi := range BuiltIns {
		if bi.FnSignature.Name == name {
			return bi.HasImpureEffects()
		}
	}
	return false
}

// the index of built-ins, listed here in alphabetical order
// All builtins have the same signature (except for the name, obviously).
//func bi_map(pr *Process, args ...object.Object) object.Object {}

type BuiltInFunction = func(pr *Process, args ...object.Object) object.Object

var BuiltIns []*object.BuiltIn

// using init() to avoid an initialization cycle
func init() {
	BuiltIns = []*object.BuiltIn{
		// internal built-ins
		// can be links to external built-ins
		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name: "_limit"},
			Fn:       bi__limit,
			ParamMin: 1, ParamMax: 1},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name: "_values"},
			Fn:       bi__values,
			ParamMin: 1, ParamMax: 1},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name: "_keys"},
			Fn:       bi_keys,
			ParamMin: 1, ParamMax: 1},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name: "_len"},
			Fn:       bi_len,
			ParamMin: 1, ParamMax: 1},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name: "_is_hash"},
			Fn:       bi_is_hash,
			ParamMin: 1, ParamMax: 1},

		// type conversion functions
		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name: common.BooleanType},
			Fn:       bi_bool,
			ParamMin: 1, ParamMax: 1},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name: common.DateTimeType},
			Fn:       bi_datetime,
			ParamMin: 1, ParamMax: 2},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name: common.DurationType},
			Fn:       bi_duration,
			ParamMin: 1, ParamMax: 1},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name: common.HashType},
			Fn:       bi_hash,
			ParamMin: 0, ParamMax: 2},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name: common.NumberType},
			Fn:       bi_number,
			ParamMin: 1, ParamMax: 2},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name: common.StringType},
			Fn:       bi_string,
			ParamMin: 1, ParamMax: 2},

		// external built-ins
		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:        "abs",
				Description: "abs(number); returns the absolute value of a number"},
			Fn:       bi_abs,
			ParamMin: 1, ParamMax: 1},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:        "all",
				Description: "all(validation, list); returns Boolean indicating whether the validation function or regex returns true for all elements of a list or hash, or null when given an empty list or hash"},
			Fn:       bi_all,
			ParamMin: 1, ParamMax: 2},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:        "any",
				Description: "any(validation, list); returns Boolean indicating whether the validation function or regex returns true for any elements of a list or hash, or null when given an empty list or hash"},
			Fn:       bi_any,
			ParamMin: 1, ParamMax: 2},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:          "appendfile",
				ImpureEffects: true,
				Description:   "appendfile(filename, string, permissions); appends string to specified file name (or writes new file if it doesn't exist); permissions optional (default 664); permissions in form of 8x644 (NOT 0644, which would give the wrong number)"},
			Fn:       bi_appendfile,
			ParamMin: 2, ParamMax: 3},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:        "atan",
				Description: "return arctangent of a number given in radians"},
			Fn:       bi_atan,
			ParamMin: 1, ParamMax: 1},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:        "b2s",
				Description: "converts a byte or list of UTF-8 bytes to a langur string"},
			Fn:       bi_b2s,
			ParamMin: 1, ParamMax: 1},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:        "benchmark",
				Description: "benchmark(function, times); runs function specified number of times (default 1), returning time elapsed (as string)"},
			Fn:       bi_benchmark,
			ParamMin: 1, ParamMax: 2},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:          "cd",
				Description:   "changes the current directory of the script; returns present working directory; has no effect on parent processes",
				ImpureEffects: true},
			Fn:       bi_cd,
			ParamMin: 0, ParamMax: 1},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:        "ceiling",
				Description: "returns least integer greater than or equal to input number"},
			Fn:       bi_ceiling,
			ParamMin: 1, ParamMax: 1},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:        "count",
				Description: "count(validation, list); returns count of values verified by given function or regex"},
			Fn:       bi_count,
			ParamMin: 1, ParamMax: 2},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:        "cos",
				Description: "return cosine of a number given in radians"},
			Fn:       bi_cos,
			ParamMin: 1, ParamMax: 1},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:        "cp2s",
				Description: "converts a code point (integer) or list of code points to a string"},
			Fn:       bi_cp2s,
			ParamMin: 1, ParamMax: 1},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:          "execT",
				ImpureEffects: true,
				Description:   "executes the given command string from a trusted source, returning a result or throwing an exception"},
			Fn:       bi_execT,
			ParamMin: 1, ParamMax: 1},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:          "execTH",
				ImpureEffects: true,
				Description:   "executes the given command string from a trusted source, returning a hash"},
			Fn:       bi_execTH,
			ParamMin: 1, ParamMax: 1},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:          "exit",
				ImpureEffects: true,
				Description:   "exits with the integer code given; 0 if no code is given; second arg as string to write to standard error, appending a newline"},
			Fn:       bi_exit,
			ParamMin: 0, ParamMax: 2},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:        "filter",
				Description: "filter(validation, list); returns list (or hash) of values verified by given function or regex, or an empty list or hash if there are no matches"},
			Fn:       bi_filter,
			ParamMin: 1, ParamMax: 2},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:        "first",
				Description: "returns first element in list or string, or start of range"},
			Fn:       bi_first,
			ParamMin: 1, ParamMax: 1},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:        "floor",
				Description: "returns greatest integer less than or equal to input number"},
			Fn:       bi_floor,
			ParamMin: 1, ParamMax: 1},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:        "fold",
				Description: "fold(function, list); returns list of values folded by the given function from the given list"},
			Fn:       bi_fold,
			ParamMin: 2, ParamMax: 2},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:        "foldfrom",
				Description: "foldfrom(function, init, lists...); returns list of values folded by the given function from the given lists; given function parameter count == number of lists + 1 for the result (result as first parameter in given function)"},
			Fn:       bi_foldfrom,
			ParamMin: 3, ParamMax: -1},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:        "gcd",
				Description: "returns the greatest common divisor of 2 or more integers"},
			Fn:       bi_gcd,
			ParamMin: 1, ParamMax: 1},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:        "group",
				Description: "group(by, list); groups list elements into list of lists as specified by first argument"},
			Fn:       bi_group,
			ParamMin: 1, ParamMax: 2},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:        "groupby",
				Description: "groupby(by, list); groups list elements into list of lists of lists as specified by first argument, including the values used to determine grouping"},
			Fn:       bi_groupby,
			ParamMin: 1, ParamMax: 2},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:        "groupbyH",
				Description: "groupbyH(by, list); groups list elements into hash as specified by first argument, using the values used to determine grouping as keys"},
			Fn:       bi_groupbyH,
			ParamMin: 2, ParamMax: 2},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:        "haskey",
				Description: "tells you if a key exists in something indexable; null if not indexable"},
			Fn:       bi_haskey,
			ParamMin: 2, ParamMax: 2},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:        "index",
				Description: "index(regex, anything, alternate); accepts regex and returns code point range for match, or returns null or alternate value (optional) for no match"},
			Fn:       bi_index,
			ParamMin: 2, ParamMax: 3},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:        "indices",
				Description: `indices(regex, anything, max); accepts regex and returns list of code point ranges for progressive matches (a.k.a. "global"), or empty list for no match; max optional`},
			Fn:       bi_indices,
			ParamMin: 2, ParamMax: 3},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:        "join",
				Description: "join(delim, list); joins list into a single string; uses auto-stringification on all list elements"},
			Fn:       bi_join,
			ParamMin: 1, ParamMax: 2},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:        "keys",
				Description: "returns the keys (as list) of a hash, or list or string indices"},
			Fn:       bi_keys,
			ParamMin: 1, ParamMax: 1},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:        "last",
				Description: "returns last element in list or string, or end of range"},
			Fn:       bi_last,
			ParamMin: 1, ParamMax: 1},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:        "lcase",
				Description: "converts string (or code point integer) to lowercase string"},
			Fn:       bi_lcase,
			ParamMin: 1, ParamMax: 1},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:        "lcm",
				Description: "returns the least common multiple of 2 or more integers"},
			Fn:       bi_lcm,
			ParamMin: 1, ParamMax: 1},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:        "len",
				Description: "returns the length (as integer) of a list, hash, or string (in code points)"},
			Fn:       bi_len,
			ParamMin: 1, ParamMax: 1},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:        "less",
				Description: "creates a new list or string, with 1 less element at the end, or an empty one if the length is already 0; may return empty list or string; also may create a new hash, with specified keys left out, or return original hash if specified keys not present"},
			Fn:       bi_less,
			ParamMin: 1, ParamMax: 2},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:        "ltrim",
				Description: "trims left-most Unicode whitespace in a string"},
			Fn:       bi_ltrim,
			ParamMin: 1, ParamMax: 1},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:        "map",
				Description: "map(function, lists...); returns list (or hash) of values mapped to the given function from the given lists or hashes (one type only)"},
			Fn:       bi_map,
			ParamMin: 2, ParamMax: -1},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:        "mapX",
				Description: "mapX(function, lists...); returns list of values mapped to the given function from the given lists"},
			Fn:       bi_mapX,
			ParamMin: 2, ParamMax: -1},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:        "match",
				Description: "match(regex, anything, alternate); accepts compiled regex and returns matching string, or returns null or alternate value (optional) for no match"},
			Fn:       bi_match,
			ParamMin: 2, ParamMax: 3},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:        "matches",
				Description: "matches(regex, anything, max); accepts compiled regex and returns list of progressive matches (empty list if no matches); max optional (defaults to -1 meaning infinite)"},
			Fn:       bi_matches,
			ParamMin: 2, ParamMax: 3},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:        "matching",
				Description: "matching(regex, anything); accepts compiled regex and returns Boolean indicating whether the string matches the pattern"},
			Fn:       bi_matching,
			ParamMin: 2, ParamMax: 2},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:        "max",
				Description: "returns maximum from a list, hash, range, string, or function; for string, max. code point; for function, max. parameter count (-1 for no max.)"},
			Fn:       bi_max,
			ParamMin: 1, ParamMax: 1},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:        "mean",
				Description: "returns mean (average) from given set of numbers"},
			Fn:       bi_mean,
			ParamMin: 1, ParamMax: 1},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:        "mid",
				Description: "returns mid-point from given set of numbers"},
			Fn:       bi_mid,
			ParamMin: 1, ParamMax: 1},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:        "min",
				Description: "returns minimum from a list, hash, range, string, or function; for string, min. code point; for function, min. parameter count"},
			Fn:       bi_min,
			ParamMin: 1, ParamMax: 1},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:        "minmax",
				Description: "returns range of minimum to maximum from a list, hash, range, string, or function; for string, min. code point; for function, min./max. parameter count"},
			Fn:       bi_minmax,
			ParamMin: 1, ParamMax: 1},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:        "more",
				Description: "more(list, items...); creates a new list or string, adding an item or items"},
			Fn:       bi_more,
			ParamMin: 1, ParamMax: -1},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:        "nfc",
				Description: "converts string to NFC form (Unicode normalization form)"},
			Fn:       bi_nfc,
			ParamMin: 1, ParamMax: 1},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:        "nfd",
				Description: "converts string to NFD form (Unicode normalization form)"},
			Fn:       bi_nfd,
			ParamMin: 1, ParamMax: 1},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:        "nfkc",
				Description: "converts string to NFKC form (Unicode normalization form)"},
			Fn:       bi_nfkc,
			ParamMin: 1, ParamMax: 1},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:        "nfkd",
				Description: "converts string to NFKD form (Unicode normalization form)"},
			Fn:       bi_nfkd,
			ParamMin: 1, ParamMax: 1},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:        "nn",
				Description: "nn(list, alternate); returns the first non-null value from a list, unless there are no non-null values, in which case it returns the alternate or an exception"},
			Fn:       bi_nn,
			ParamMin: 1, ParamMax: 2},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:          "prop",
				ImpureEffects: true,
				Description:   "returns hash of properties of file or directory if it exists at the given moment of execution; otherwise returns null"},
			Fn:       bi_prop,
			ParamMin: 1, ParamMax: 1},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:        "pseries",
				Description: "like series(), but returns positive series or empty list (given a negative range)"},
			Fn:       bi_pseries,
			ParamMin: 1, ParamMax: 2},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:        "random",
				Description: "returns random integer from a given range"},
			Fn:       bi_random,
			ParamMin: 1, ParamMax: 1},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:          "read",
				ImpureEffects: true,
				Description:   "read(prompt, validation, errmessage, maxattempts, alternate); reads from the console, validating the string is good by the regex or function passed, and giving the error message specified if the string is no good; If no alternate is given, this may ultimately generate an error."},
			Fn:       bi_read,
			ParamMin: 0, ParamMax: 5},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:          "readfile",
				ImpureEffects: true,
				Description:   "reads text of file name given, returning a string"},
			Fn:       bi_readfile,
			ParamMin: 1, ParamMax: 1},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:        "reCompile",
				Description: "compiles string pattern into re2 regex"},
			Fn:       bi_reCompile,
			ParamMin: 1, ParamMax: 1},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:        "reEsc",
				Description: "escapes re2 metacharacters in a pattern string"},
			Fn:       bi_reEsc,
			ParamMin: 1, ParamMax: 1},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:        "replace",
				Description: "replace(source, find, replace, max); accepts string or regex for find, and replaces portion of string with given replacement string; max optional"},
			Fn:       bi_replace,
			ParamMin: 2, ParamMax: 4},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:        "rest",
				Description: "removes the first element of a list or string or does nothing if the length is 0; may return empty list or string"},
			Fn:       bi_rest,
			ParamMin: 1, ParamMax: 1},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:        "reverse",
				Description: "returns the reversed list or range"},
			Fn:       bi_reverse,
			ParamMin: 1, ParamMax: 1},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:        "round",
				Description: "round(number, max, addzeroes, mode); rounds number to specified digits after decimal point; mode from the " + modes.RoundHashName + " hash"},
			Fn:       bi_round,
			ParamMin: 1, ParamMax: 4},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:        "rotate",
				Description: "rotates list elements or a number within a range"},
			Fn:       bi_rotate,
			ParamMin: 1, ParamMax: 3},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:        "rtrim",
				Description: "trims right-most Unicode whitespace in a string"},
			Fn:       bi_rtrim,
			ParamMin: 1, ParamMax: 1},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:        "s2b",
				Description: "returns list of UTF-8 bytes from a langur string"},
			Fn:       bi_s2b,
			ParamMin: 1, ParamMax: 1},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:        "s2cp",
				Description: "s2cp(string, index, alternate); indexes a string to a code point or a list of code points"},
			Fn:       bi_s2cp,
			ParamMin: 1, ParamMax: 3},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:        "s2gc",
				Description: "s2gc(string); converts string to grapheme clusters list"},
			Fn:       bi_s2gc,
			ParamMin: 1, ParamMax: 1},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:        "s2n",
				Description: "returns list of numbers from a langur string, interpreting 0-9, A-Z, and a-z as base 36 numbers"},
			Fn:       bi_s2n,
			ParamMin: 1, ParamMax: 1},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:        "s2s",
				Description: "s2s(string, index, alternate); indexes a string to a string"},
			Fn:       bi_s2s,
			ParamMin: 2, ParamMax: 3},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:        "series",
				Description: "series(range, increment); generates a list of numbers from a range and increment (optional, defaults to 1 or -1)"},
			Fn:       bi_series,
			ParamMin: 1, ParamMax: 2},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:        "simplify",
				Description: "simplifies number, removing trailing zeros"},
			Fn:       bi_simplify,
			ParamMin: 1, ParamMax: 1},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:        "sine",
				Description: "return sine of a number given in radians"},
			Fn:       bi_sine,
			ParamMin: 1, ParamMax: 1},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:          "sleep",
				ImpureEffects: true,
				Description:   "waits for the specified number of milliseconds"},
			Fn:       bi_sleep,
			ParamMin: 1, ParamMax: 1},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:        "sort",
				Description: "sort(function, list); returns a sorted list from the given list, comparing by the given function (taking two variables and returning a Boolean in the form of f(.a, .b) .a < .b, or with implied parameters in the form of f .a < .b)"},
			Fn:       bi_sort,
			ParamMin: 1, ParamMax: 2},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:        "split",
				Description: "split(delim, anything, max); accepts regex or string delimiter and splits string into a list of strings; max optional"},
			Fn:       bi_split,
			ParamMin: 1, ParamMax: 3},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:        "subindex",
				Description: `subindex(regex, anything); accepts regex and returns list of code point ranges for submatches, or empty list for no match`},
			Fn:       bi_subindex,
			ParamMin: 2, ParamMax: 2},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:        "subindices",
				Description: `subindices(regex, anything, max); accepts regex and returns list of lists of code point ranges for progressive submatches (a.k.a. "global"), or empty list for no match; max optional`},
			Fn:       bi_subindices,
			ParamMin: 2, ParamMax: 3},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:        "submatch",
				Description: "submatch(regex, anything); returns list of submatches (empty list if not a match)"},
			Fn:       bi_submatch,
			ParamMin: 2, ParamMax: 2},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:        "submatchH",
				Description: "submatchH(regex, anything); returns hash of submatches (empty hash if not a match)"},
			Fn:       bi_submatchH,
			ParamMin: 2, ParamMax: 2},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:        "submatches",
				Description: "submatches(regex, anything, max); returns list of lists of progressive submatches (empty list if not a match); max optional (defaults to -1 meaning infinite)"},
			Fn:       bi_submatches,
			ParamMin: 2, ParamMax: 3},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:        "submatchesH",
				Description: "submatchesH(regex, anything, max); returns list of hashes of progressive whole match and submatches (empty list if not a match); max optional (defaults to -1 meaning infinite)"},
			Fn:       bi_submatchesH,
			ParamMin: 2, ParamMax: 3},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:        "tan",
				Description: "return tangent of a number given in radians"},
			Fn:       bi_tan,
			ParamMin: 1, ParamMax: 1},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:        "tcase",
				Description: "converts string (or code point integer) to titlecase string"},
			Fn:       bi_tcase,
			ParamMin: 1, ParamMax: 1},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:        "ticks",
				Description: "returns Unix ticks in nanoseconds"},
			Fn:       bi_ticks,
			ParamMin: 0, ParamMax: 0},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:        "trim",
				Description: "trims Unicode whitespace around a string"},
			Fn:       bi_trim,
			ParamMin: 1, ParamMax: 1},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:        "trunc",
				Description: "trunc(number, max, addzeroes); truncate number to specified digits after decimal point"},
			Fn:       bi_trunc,
			ParamMin: 1, ParamMax: 3},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:        "ucase",
				Description: "converts string (or code point integer) to uppercase string"},
			Fn:       bi_ucase,
			ParamMin: 1, ParamMax: 1},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:          "write",
				ImpureEffects: true,
				Description:   "writes to the console"},
			Fn:       bi_write,
			ParamMin: 0, ParamMax: -1},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:          "writeErr",
				ImpureEffects: true,
				Description:   "writes to standard error"},
			Fn:       bi_write,
			ParamMin: 0, ParamMax: -1},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:          "writefile",
				ImpureEffects: true,
				Description:   "writefile(filename, string, permissions); writes string to specified file name; permissions optional (default 664); permissions in form of 8x644 (NOT 0644, which would give the wrong number)"},
			Fn:       bi_writefile,
			ParamMin: 2, ParamMax: 3},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:          "writeln",
				ImpureEffects: true,
				Description:   "writes to the console, adding a system newline at the end"},
			Fn:       bi_writeln,
			ParamMin: 0, ParamMax: -1},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:          "writelnErr",
				ImpureEffects: true,
				Description:   "writes to standard error, adding a system newline at the end"},
			Fn:       bi_writeln,
			ParamMin: 0, ParamMax: -1},

		&object.BuiltIn{
			FnSignature: &object.Signature{
				Name:        "zip",
				Description: "zips together lists; may optionally use a function (first argument)"},
			Fn:       bi_zip,
			ParamMin: 2, ParamMax: -1},
	}
}
