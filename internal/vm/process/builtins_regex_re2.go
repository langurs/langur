// langur/vm/process/builtins_regex_re2.go

package process

import (
	"langur/object"
	"langur/regex"
)

// reCompile, reEsc

// re2 functions (see also builtin_regex.go)...

var bi_reCompile = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:        "reCompile",
		Description: "compiles string pattern into re2 regex",

		ParamPositional: []object.Parameter{
			object.Parameter{ExternalName: "from", Type: object.STRING_OBJ},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		re, err := object.NewRegex(args[0].String(), regex.RE2)
		if err != nil {
			return object.NewError(object.ERR_ARGUMENTS, "reCompile", err.Error())
		}
		return re
	},
}

var bi_reEsc = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:        "reEsc",
		Description: "escapes re2 metacharacters in a pattern string",

		ParamPositional: []object.Parameter{
			object.Parameter{ExternalName: "from"},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		pattern := object.ToString(args[0])

		newStrObj, err := object.EscString(pattern, regex.RE2)
		if err != nil {
			return object.NewError(object.ERR_ARGUMENTS, "reEsc", err.Error())
		}
		return newStrObj
	},
}
