// langur/vm/process/builtins.go

package process

import (
	"langur/common"
	"langur/modes"
	"langur/object"
)

func GetBuiltInByName(name string) *object.BuiltIn {
	for _, bi := range BuiltIns {
		if bi.Name == name {
			return bi
		}
	}
	return nil
}

func GetBuiltInImpurityStatus(name string) bool {
	for _, bi := range BuiltIns {
		if bi.Name == name {
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

const (
	// TODO: libraries
	LIBRARY_MATH   = "math"
	LIBRARY_TRIG   = "trig"
	LIBRARY_STDIO  = "io"
	LIBRARY_STRING = "string"
	LIBRARY_REGEX  = "regex"
	LIBRARY_FILES  = "file"
	LIBRARY_FUNC   = "func"
	LIBRARY_EXEC   = "exec"
)

// using init() to avoid an initialization cycle
func init() {
	BuiltIns = []*object.BuiltIn{
		// internal built-ins
		// can be links to external built-ins
		&object.BuiltIn{
			Name: "_limit", Fn: bi__limit,
			ParamMin: 1, ParamMax: 1},

		&object.BuiltIn{
			Name: "_values", Fn: bi__values,
			ParamMin: 1, ParamMax: 1},

		&object.BuiltIn{
			Name: "_keys", Fn: bi_keys,
			ParamMin: 1, ParamMax: 1},

		&object.BuiltIn{
			Name: "_len", Fn: bi_len,
			ParamMin: 1, ParamMax: 1},

		&object.BuiltIn{
			Name: "_is_hash", Fn: bi_is_hash,
			ParamMin: 1, ParamMax: 1},

		// type conversion functions
		&object.BuiltIn{
			Name: common.BooleanType, Fn: bi_bool,
			ParamMin: 1, ParamMax: 1},

		&object.BuiltIn{
			Name: common.DateTimeType, Fn: bi_datetime,
			ParamMin: 1, ParamMax: 2},

		&object.BuiltIn{
			Name: common.DurationType, Fn: bi_duration,
			ParamMin: 1, ParamMax: 1},

		&object.BuiltIn{
			Name: common.HashType, Fn: bi_hash,
			ParamMin: 0, ParamMax: 2},

		&object.BuiltIn{
			Name: common.NumberType, Fn: bi_number,
			ParamMin: 1, ParamMax: 2},

		&object.BuiltIn{
			Name: common.StringType, Fn: bi_string,
			ParamMin: 1, ParamMax: 2},

		// external built-ins
		&object.BuiltIn{
			Name: "abs", Fn: bi_abs,
			Description: "abs(number); returns the absolute value of a number",
			ParamMin:    1, ParamMax: 1,
			Library: LIBRARY_MATH},

		&object.BuiltIn{
			Name: "all", Fn: bi_all,
			Description: "all(validation, list); returns Boolean indicating whether the validation function or regex returns true for all elements of a list or hash, or null when given an empty list or hash",
			ParamMin:    1, ParamMax: 2,
			Library: LIBRARY_FUNC},

		&object.BuiltIn{
			Name: "any", Fn: bi_any,
			Description: "any(validation, list); returns Boolean indicating whether the validation function or regex returns true for any elements of a list or hash, or null when given an empty list or hash",
			ParamMin:    1, ParamMax: 2,
			Library: LIBRARY_FUNC},

		&object.BuiltIn{
			Name: "appendfile", Fn: bi_appendfile,
			Description: "appendfile(filename, string, permissions); appends string to specified file name (or writes new file if it doesn't exist); permissions optional (default 664); permissions in form of 8x644 (NOT 0644, which would give the wrong number)",
			ParamMin:    2, ParamMax: 3,
			ImpureEffects: true,
			Library:       LIBRARY_FILES},

		&object.BuiltIn{
			Name: "atan", Fn: bi_atan,
			Description: "return arctangent of a number given in radians",
			ParamMin:    1, ParamMax: 1,
			Library: LIBRARY_TRIG},

		&object.BuiltIn{
			Name: "b2s", Fn: bi_b2s,
			Description: "converts a byte or list of UTF-8 bytes to a langur string",
			ParamMin:    1, ParamMax: 1,
			Library: LIBRARY_STRING},

		&object.BuiltIn{
			Name: "benchmark", Fn: bi_benchmark,
			Description: "benchmark(function, times); runs function specified number of times (default 1), returning time elapsed (as string)",
			ParamMin:    1, ParamMax: 2,
			Library: LIBRARY_FUNC},

		&object.BuiltIn{
			Name: "cd", Fn: bi_cd,
			Description: "changes the current directory of the script; returns present working directory; has no effect on parent processes",
			ParamMin:    0, ParamMax: 1,
			ImpureEffects: true},

		&object.BuiltIn{
			Name: "ceiling", Fn: bi_ceiling,
			Description: "returns least integer greater than or equal to input number",
			ParamMin:    1, ParamMax: 1,
			Library: LIBRARY_MATH},

		&object.BuiltIn{
			Name: "count", Fn: bi_count,
			Description: "count(validation, list); returns count of values verified by given function or regex",
			ParamMin:    1, ParamMax: 2},

		&object.BuiltIn{
			Name: "cos", Fn: bi_cos,
			Description: "return cosine of a number given in radians",
			ParamMin:    1, ParamMax: 1,
			Library: LIBRARY_TRIG},

		&object.BuiltIn{
			Name: "cp2s", Fn: bi_cp2s,
			Description: "converts a code point (integer) or list of code points to a string",
			ParamMin:    1, ParamMax: 1,
			Library: LIBRARY_STRING},

		&object.BuiltIn{
			Name: "execT", Fn: bi_execT,
			Description: "executes the given command string from a trusted source, returning a result or throwing an exception",
			ParamMin:    1, ParamMax: 1,
			ImpureEffects: true,
			Library:       LIBRARY_EXEC},

		&object.BuiltIn{
			Name: "execTH", Fn: bi_execTH,
			Description: "executes the given command string from a trusted source, returning a hash",
			ParamMin:    1, ParamMax: 1,
			ImpureEffects: true,
			Library:       LIBRARY_EXEC},

		&object.BuiltIn{
			Name: "exit", Fn: bi_exit,
			Description: "exits with the integer code given; 0 if no code is given; second arg as string to write to standard error, appending a newline",
			ParamMin:    0, ParamMax: 2,
			ImpureEffects: true},

		&object.BuiltIn{
			Name: "filter", Fn: bi_filter,
			Description: "filter(validation, list); returns list (or hash) of values verified by given function or regex, or an empty list or hash if there are no matches",
			ParamMin:    1, ParamMax: 2,
			Library: LIBRARY_FUNC},

		&object.BuiltIn{
			Name: "first", Fn: bi_first,
			Description: "returns first element in list or string, or start of range",
			ParamMin:    1, ParamMax: 1},

		&object.BuiltIn{
			Name: "floor", Fn: bi_floor,
			Description: "returns greatest integer less than or equal to input number",
			ParamMin:    1, ParamMax: 1,
			Library: LIBRARY_MATH},

		&object.BuiltIn{
			Name: "fold", Fn: bi_fold,
			Description: "fold(function, list); returns list of values folded by the given function from the given list",
			ParamMin:    2, ParamMax: 2,
			Library: LIBRARY_FUNC},

		&object.BuiltIn{
			Name: "foldfrom", Fn: bi_foldfrom,
			Description: "foldfrom(function, init, lists...); returns list of values folded by the given function from the given lists; given function parameter count == number of lists + 1 for the result (result as first parameter in given function)",
			ParamMin:    3, ParamMax: -1,
			Library: LIBRARY_FUNC},

		&object.BuiltIn{
			Name: "gcd", Fn: bi_gcd,
			Description: "returns the greatest common divisor of 2 or more integers",
			ParamMin:    1, ParamMax: 1,
			Library: LIBRARY_MATH},

		&object.BuiltIn{
			Name: "group", Fn: bi_group,
			Description: "group(by, list); groups list elements into list of lists as specified by first argument",
			ParamMin:    1, ParamMax: 2},

		&object.BuiltIn{
			Name: "groupby", Fn: bi_groupby,
			Description: "groupby(by, list); groups list elements into list of lists of lists as specified by first argument, including the values used to determine grouping",
			ParamMin:    1, ParamMax: 2},

		&object.BuiltIn{
			Name: "groupbyH", Fn: bi_groupbyH,
			Description: "groupbyH(by, list); groups list elements into hash as specified by first argument, using the values used to determine grouping as keys",
			ParamMin:    2, ParamMax: 2},

		&object.BuiltIn{
			Name: "haskey", Fn: bi_haskey,
			Description: "tells you if a key exists in something indexable; null if not indexable",
			ParamMin:    2, ParamMax: 2},

		&object.BuiltIn{
			Name: "index", Fn: bi_index,
			Description: "index(regex, anything, alternate); accepts regex and returns code point range for match, or returns null or alternate value (optional) for no match",
			ParamMin:    2, ParamMax: 3,
			Library: LIBRARY_REGEX},

		&object.BuiltIn{
			Name: "indices", Fn: bi_indices,
			Description: `indices(regex, anything, max); accepts regex and returns list of code point ranges for progressive matches (a.k.a. "global"), or empty list for no match; max optional`,
			ParamMin:    2, ParamMax: 3,
			Library: LIBRARY_REGEX},

		&object.BuiltIn{
			Name: "join", Fn: bi_join,
			Description: "join(delim, list); joins list into a single string; uses auto-stringification on all list elements",
			ParamMin:    1, ParamMax: 2,
			Library: LIBRARY_STRING},

		&object.BuiltIn{
			Name: "keys", Fn: bi_keys,
			Description: "returns the keys (as list) of a hash, or list or string indices",
			ParamMin:    1, ParamMax: 1},

		&object.BuiltIn{
			Name: "last", Fn: bi_last,
			Description: "returns last element in list or string, or end of range",
			ParamMin:    1, ParamMax: 1},

		&object.BuiltIn{
			Name: "lcase", Fn: bi_lcase,
			Description: "converts string (or code point integer) to lowercase string",
			ParamMin:    1, ParamMax: 1,
			Library: LIBRARY_STRING},

		&object.BuiltIn{
			Name: "lcm", Fn: bi_lcm,
			Description: "returns the least common multiple of 2 or more integers",
			ParamMin:    1, ParamMax: 1,
			Library: LIBRARY_MATH},

		&object.BuiltIn{
			Name: "len", Fn: bi_len,
			Description: "returns the length (as integer) of a list, hash, or string (in code points)",
			ParamMin:    1, ParamMax: 1},

		&object.BuiltIn{
			Name: "less", Fn: bi_less,
			Description: "creates a new list or string, with 1 less element at the end, or an empty one if the length is already 0; may return empty list or string; also may create a new hash, with specified keys left out, or return original hash if specified keys not present",
			ParamMin:    1, ParamMax: 2},

		&object.BuiltIn{
			Name: "ltrim", Fn: bi_ltrim,
			Description: "trims left-most Unicode whitespace in a string",
			ParamMin:    1, ParamMax: 1,
			Library: LIBRARY_STRING},

		&object.BuiltIn{
			Name: "map", Fn: bi_map,
			Description: "map(function, lists...); returns list (or hash) of values mapped to the given function from the given lists or hashes (one type only)",
			ParamMin:    2, ParamMax: -1,
			Library: LIBRARY_FUNC},

		&object.BuiltIn{
			Name: "mapX", Fn: bi_mapX,
			Description: "mapX(function, lists...); returns list of values mapped to the given function from the given lists",
			ParamMin:    2, ParamMax: -1,
			Library: LIBRARY_FUNC},

		&object.BuiltIn{
			Name: "match", Fn: bi_match,
			Description: "match(regex, anything, alternate); accepts compiled regex and returns matching string, or returns null or alternate value (optional) for no match",
			ParamMin:    2, ParamMax: 3,
			Library: LIBRARY_REGEX},

		&object.BuiltIn{
			Name: "matches", Fn: bi_matches,
			Description: "matches(regex, anything, max); accepts compiled regex and returns list of progressive matches (empty list if no matches); max optional (defaults to -1 meaning infinite)",
			ParamMin:    2, ParamMax: 3,
			Library: LIBRARY_REGEX},

		&object.BuiltIn{
			Name: "matching", Fn: bi_matching,
			Description: "matching(regex, anything); accepts compiled regex and returns Boolean indicating whether the string matches the pattern",
			ParamMin:    2, ParamMax: 2,
			Library: LIBRARY_REGEX},

		&object.BuiltIn{
			Name: "max", Fn: bi_max,
			Description: "returns maximum from a list, hash, range, string, or function; for string, max. code point; for function, max. parameter count (-1 for no max.)",
			ParamMin:    1, ParamMax: 1},

		&object.BuiltIn{
			Name: "mean", Fn: bi_mean,
			Description: "returns mean (average) from given set of numbers",
			ParamMin:    1, ParamMax: 1,
			Library: LIBRARY_MATH},

		&object.BuiltIn{
			Name: "mid", Fn: bi_mid,
			Description: "returns mid-point from given set of numbers",
			ParamMin:    1, ParamMax: 1,
			Library: LIBRARY_MATH},

		&object.BuiltIn{
			Name: "min", Fn: bi_min,
			Description: "returns minimum from a list, hash, range, string, or function; for string, min. code point; for function, min. parameter count",
			ParamMin:    1, ParamMax: 1},

		&object.BuiltIn{
			Name: "minmax", Fn: bi_minmax,
			Description: "returns range of minimum to maximum from a list, hash, range, string, or function; for string, min. code point; for function, min./max. parameter count",
			ParamMin:    1, ParamMax: 1},

		&object.BuiltIn{
			Name: "more", Fn: bi_more,
			Description: "more(list, items...); creates a new list or string, adding an item or items",
			ParamMin:    1, ParamMax: -1},

		&object.BuiltIn{
			Name: "nfc", Fn: bi_nfc,
			Description: "converts string to NFC form (Unicode normalization form)",
			ParamMin:    1, ParamMax: 1,
			Library: LIBRARY_STRING},

		&object.BuiltIn{
			Name: "nfd", Fn: bi_nfd,
			Description: "converts string to NFD form (Unicode normalization form)",
			ParamMin:    1, ParamMax: 1,
			Library: LIBRARY_STRING},

		&object.BuiltIn{
			Name: "nfkc", Fn: bi_nfkc,
			Description: "converts string to NFKC form (Unicode normalization form)",
			ParamMin:    1, ParamMax: 1,
			Library: LIBRARY_STRING},

		&object.BuiltIn{
			Name: "nfkd", Fn: bi_nfkd,
			Description: "converts string to NFKD form (Unicode normalization form)",
			ParamMin:    1, ParamMax: 1,
			Library: LIBRARY_STRING},

		&object.BuiltIn{
			Name: "nn", Fn: bi_nn,
			Description: "nn(list, alternate); returns the first non-null value from a list, unless there are no non-null values, in which case it returns the alternate or an exception",
			ParamMin:    1, ParamMax: 2},

		&object.BuiltIn{
			Name: "prop", Fn: bi_prop,
			Description: "returns hash of properties of file or directory if it exists at the given moment of execution; otherwise returns null",
			ParamMin:    1, ParamMax: 1,
			ImpureEffects: true,
			Library:       LIBRARY_FILES},

		&object.BuiltIn{
			Name: "pseries", Fn: bi_pseries,
			Description: "like series(), but returns positive series or empty list (given a negative range)",
			ParamMin:    1, ParamMax: 2},

		&object.BuiltIn{
			Name: "random", Fn: bi_random,
			Description: "returns random integer from a given range",
			ParamMin:    1, ParamMax: 1},

		&object.BuiltIn{
			Name: "read", Fn: bi_read,
			Description: "read(prompt, validation, errmessage, maxattempts, alternate); reads from the console, validating the string is good by the regex or function passed, and giving the error message specified if the string is no good; If no alternate is given, this may ultimately generate an error.",
			ParamMin:    0, ParamMax: 5,
			ImpureEffects: true,
			Library:       LIBRARY_STDIO},

		&object.BuiltIn{
			Name: "readfile", Fn: bi_readfile,
			Description: "reads text of file name given, returning a string",
			ParamMin:    1, ParamMax: 1,
			ImpureEffects: true,
			Library:       LIBRARY_FILES},

		&object.BuiltIn{
			Name: "reCompile", Fn: bi_reCompile,
			Description: "compiles string pattern into re2 regex",
			ParamMin:    1, ParamMax: 1,
			Library: LIBRARY_REGEX},

		&object.BuiltIn{
			Name: "reEsc", Fn: bi_reEsc,
			Description: "escapes re2 metacharacters in a pattern string",
			ParamMin:    1, ParamMax: 1,
			Library: LIBRARY_REGEX},

		&object.BuiltIn{
			Name: "replace", Fn: bi_replace,
			Description: "replace(source, find, replace, max); accepts string or regex for find, and replaces portion of string with given replacement string; max optional",
			ParamMin:    2, ParamMax: 4,
			Library: LIBRARY_REGEX},

		&object.BuiltIn{
			Name: "rest", Fn: bi_rest,
			Description: "removes the first element of a list or string or does nothing if the length is 0; may return empty list or string",
			ParamMin:    1, ParamMax: 1},

		&object.BuiltIn{
			Name: "reverse", Fn: bi_reverse,
			Description: "returns the reversed list or range",
			ParamMin:    1, ParamMax: 1},

		&object.BuiltIn{
			Name: "round", Fn: bi_round,
			Description: "round(number, max, trim, mode); rounds number to specified digits after decimal point; max < 0 means not to pad with extra trailing zeros; mode from the " + modes.RoundHashName + " hash",
			ParamMin:    1, ParamMax: 4,
			Library: LIBRARY_MATH},

		&object.BuiltIn{
			Name: "rotate", Fn: bi_rotate,
			Description: "rotates list elements or a number within a range",
			ParamMin:    1, ParamMax: 3},

		&object.BuiltIn{
			Name: "rtrim", Fn: bi_rtrim,
			Description: "trims right-most Unicode whitespace in a string",
			ParamMin:    1, ParamMax: 1,
			Library: LIBRARY_STRING},

		&object.BuiltIn{
			Name: "s2b", Fn: bi_s2b,
			Description: "returns list of UTF-8 bytes from a langur string",
			ParamMin:    1, ParamMax: 1,
			Library: LIBRARY_STRING},

		&object.BuiltIn{
			Name: "s2cp", Fn: bi_s2cp,
			Description: "s2cp(string, index, alternate); indexes a string to a code point or a list of code points",
			ParamMin:    1, ParamMax: 3,
			Library: LIBRARY_STRING},

		&object.BuiltIn{
			Name: "s2gc", Fn: bi_s2gc,
			Description: "s2gc(string); converts string to grapheme clusters list",
			ParamMin:    1, ParamMax: 1,
			Library: LIBRARY_STRING},

		&object.BuiltIn{
			Name: "s2n", Fn: bi_s2n,
			Description: "returns list of numbers from a langur string, interpreting 0-9, A-Z, and a-z as base 36 numbers",
			ParamMin:    1, ParamMax: 1,
			Library: LIBRARY_STRING},

		&object.BuiltIn{
			Name: "s2s", Fn: bi_s2s,
			Description: "s2s(string, index, alternate); indexes a string to a string",
			ParamMin:    2, ParamMax: 3,
			Library: LIBRARY_STRING},

		&object.BuiltIn{
			Name: "series", Fn: bi_series,
			Description: "series(range, increment); generates a list of numbers from a range and increment (optional, defaults to 1 or -1)",
			ParamMin:    1, ParamMax: 2},

		&object.BuiltIn{
			Name: "simplify", Fn: bi_simplify,
			Description: "simplifies number, removing trailing zeros",
			ParamMin:    1, ParamMax: 1,
			Library: LIBRARY_MATH},

		&object.BuiltIn{
			Name: "sine", Fn: bi_sine,
			Description: "return sine of a number given in radians",
			ParamMin:    1, ParamMax: 1,
			Library: LIBRARY_TRIG},

		&object.BuiltIn{
			Name: "sleep", Fn: bi_sleep,
			Description: "waits for the specified number of milliseconds",
			ParamMin:    1, ParamMax: 1,
			ImpureEffects: true},

		&object.BuiltIn{
			Name: "sort", Fn: bi_sort,
			Description: "sort(function, list); returns a sorted list from the given list, comparing by the given function (taking two variables and returning a Boolean in the form of f(.a, .b) .a < .b, or with implied parameters in the form of f .a < .b)",
			ParamMin:    1, ParamMax: 2},

		&object.BuiltIn{
			Name: "split", Fn: bi_split,
			Description: "split(delim, anything, max); accepts regex or string delimiter and splits string into a list of strings; max optional",
			ParamMin:    1, ParamMax: 3,
			Library: LIBRARY_STRING},

		&object.BuiltIn{
			Name: "subindex", Fn: bi_subindex,
			Description: `subindex(regex, anything); accepts regex and returns list of code point ranges for submatches, or empty list for no match`,
			ParamMin:    2, ParamMax: 2,
			Library: LIBRARY_REGEX},

		&object.BuiltIn{
			Name: "subindices", Fn: bi_subindices,
			Description: `subindices(regex, anything, max); accepts regex and returns list of lists of code point ranges for progressive submatches (a.k.a. "global"), or empty list for no match; max optional`,
			ParamMin:    2, ParamMax: 3,
			Library: LIBRARY_REGEX},

		&object.BuiltIn{
			Name: "submatch", Fn: bi_submatch,
			Description: "submatch(regex, anything); returns list of submatches (empty list if not a match)",
			ParamMin:    2, ParamMax: 2,
			Library: LIBRARY_REGEX},

		&object.BuiltIn{
			Name: "submatchH", Fn: bi_submatchH,
			Description: "submatchH(regex, anything); returns hash of submatches (empty hash if not a match)",
			ParamMin:    2, ParamMax: 2,
			Library: LIBRARY_REGEX},

		&object.BuiltIn{
			Name: "submatches", Fn: bi_submatches,
			Description: "submatches(regex, anything, max); returns list of lists of progressive submatches (empty list if not a match); max optional (defaults to -1 meaning infinite)",
			ParamMin:    2, ParamMax: 3,
			Library: LIBRARY_REGEX},

		&object.BuiltIn{
			Name: "submatchesH", Fn: bi_submatchesH,
			Description: "submatchesH(regex, anything, max); returns list of hashes of progressive whole match and submatches (empty list if not a match); max optional (defaults to -1 meaning infinite)",
			ParamMin:    2, ParamMax: 3,
			Library: LIBRARY_REGEX},

		&object.BuiltIn{
			Name: "tan", Fn: bi_tan,
			Description: "return tangent of a number given in radians",
			ParamMin:    1, ParamMax: 1,
			Library: LIBRARY_TRIG},

		&object.BuiltIn{
			Name: "tcase", Fn: bi_tcase,
			Description: "converts string (or code point integer) to titlecase string",
			ParamMin:    1, ParamMax: 1,
			Library: LIBRARY_STRING},

		&object.BuiltIn{
			Name: "ticks", Fn: bi_ticks,
			Description: "returns Unix ticks in nanoseconds",
			ParamMin:    0, ParamMax: 0},

		&object.BuiltIn{
			Name: "trim", Fn: bi_trim,
			Description: "trims Unicode whitespace around a string",
			ParamMin:    1, ParamMax: 1,
			Library: LIBRARY_STRING},

		&object.BuiltIn{
			Name: "trunc", Fn: bi_trunc,
			Description: "trunc(number, max, trim); truncate number to specified digits after decimal point; max < 0 means not to pad with extra trailing zeros",
			ParamMin:    1, ParamMax: 3,
			Library: LIBRARY_MATH},

		&object.BuiltIn{
			Name: "ucase", Fn: bi_ucase,
			Description: "converts string (or code point integer) to uppercase string",
			ParamMin:    1, ParamMax: 1,
			Library: LIBRARY_STRING},

		&object.BuiltIn{
			Name: "write", Fn: bi_write,
			Description: "writes to the console",
			ParamMin:    0, ParamMax: -1,
			ImpureEffects: true,
			Library:       LIBRARY_STDIO},

		&object.BuiltIn{
			Name: "writeErr", Fn: bi_write,
			Description: "writes to standard error",
			ParamMin:    0, ParamMax: -1,
			ImpureEffects: true,
			Library:       LIBRARY_STDIO},

		&object.BuiltIn{
			Name: "writefile", Fn: bi_writefile,
			Description: "writefile(filename, string, permissions); writes string to specified file name; permissions optional (default 664); permissions in form of 8x644 (NOT 0644, which would give the wrong number)",
			ParamMin:    2, ParamMax: 3,
			ImpureEffects: true,
			Library:       LIBRARY_FILES},

		&object.BuiltIn{
			Name: "writeln", Fn: bi_writeln,
			Description: "writes to the console, adding a system newline at the end",
			ParamMin:    0, ParamMax: -1,
			ImpureEffects: true,
			Library:       LIBRARY_STDIO},

		&object.BuiltIn{
			Name: "writelnErr", Fn: bi_writeln,
			Description: "writes to standard error, adding a system newline at the end",
			ParamMin:    0, ParamMax: -1,
			ImpureEffects: true,
			Library:       LIBRARY_STDIO},

		&object.BuiltIn{
			Name: "zip", Fn: bi_zip,
			Description: "zips together lists; may optionally use a function (first argument)",
			ParamMin:    2, ParamMax: -1,
			Library: LIBRARY_FUNC},
	}
}
