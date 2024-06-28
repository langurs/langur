// langur/vm/process/builtins_regex_re2.go

package process

import (
	"langur/object"
	"langur/regex"
)

// reCompile, reEsc

// re2 functions (see also builtin_regex.go)...

func bi_reCompile(pr *Process, args ...object.Object) object.Object {
	pattern, ok := args[0].(*object.String)
	if !ok {
		return object.NewError(object.ERR_ARGUMENTS, "reCompile", "Expected string for pattern")
	}
	re, err := object.NewRegex(pattern.String(), regex.RE2)
	if err != nil {
		return object.NewError(object.ERR_ARGUMENTS, "reCompile", err.Error())
	}
	return re
}

func bi_reEsc(pr *Process, args ...object.Object) object.Object {
	pattern := object.ToString(args[0])

	newStrObj, err := object.EscString(pattern, regex.RE2)
	if err != nil {
		return object.NewError(object.ERR_ARGUMENTS, "reEsc", err.Error())
	}
	return newStrObj
}
