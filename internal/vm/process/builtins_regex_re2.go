// langur/vm/process/builtins_pattern_re2.go

package process

import (
	"langur/object"
	"langur/pattern"
)

// reCompile, reEsc

// re2 functions (see also builtin_pattern.go)...

var bi_reCompile = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:        "reCompile",
		Description: "compiles string pattern into re2 pattern",

		ParamPositional: []object.Parameter{
			object.Parameter{ExternalName: "from", Type: object.STRING_OBJ},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		re, err := object.NewPattern(args[0].String(), pattern.RE2)
		if err != nil {
			return object.NewException(object.ERR_ARGUMENTS, "reCompile", err.Error())
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
		p := object.ToString(args[0])

		newStrObj, err := object.EscString(p, pattern.RE2)
		if err != nil {
			return object.NewException(object.ERR_ARGUMENTS, "reEsc", err.Error())
		}
		return newStrObj
	},
}
