// langur/vm/process/builtins_io.go

package process

import (
	"bytes"
	"langur/object"
	"langur/str"
	"langur/text_io"
)

// write, writeln
// writeErr, writelnErr
// read

var newLine = object.NewString(str.SysNewLine)

var bi_write = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:          "write",
		ImpureEffects: true,
		Description:   "writes to the console",

		ParamPositional: []object.Parameter{
			object.Parameter{ExternalName: "text"},
		},
		ParamExpansionMax: -1,

		ParamKeyword: []object.Parameter{
			object.Parameter{ExternalName: "textmode", Type: object.BOOLEAN_OBJ},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		var out bytes.Buffer

		// "textmode" argument
		textMode := pr.Modes.ConsoleTextMode
		if args[1] != nil {
			textMode = args[1].IsTruthy()
		}

		args = args[0].(*object.List).Elements

		for _, v := range args {
			out.WriteString(v.String())
		}

		s := out.String()
		if len(s) == 0 {
			return object.NULL
		}

		text_io.Print(s, textMode)
		return object.TRUE
	},
}

var bi_writeln = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:          "writeln",
		ImpureEffects: true,
		Description:   "writes to the console, adding a newline at the end",

		ParamPositional: []object.Parameter{
			object.Parameter{ExternalName: "text"},
		},
		ParamExpansionMax: -1,

		ParamKeyword: []object.Parameter{
			object.Parameter{ExternalName: "textmode", Type: object.BOOLEAN_OBJ},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		var textMode object.Object
		if args[1] != nil {
			textMode = args[1].Copy()
		}

		// WARN: Calling a built-in directly has the potential to violate its signature.
		// Using Process.callback() is safer since it will check the arguments against the signature.
		return bi_write.Fn.(BuiltInFunction)(pr,
			&object.List{Elements: append(args[0].(*object.List).Elements, newLine)},
			textMode)
	},
}

var bi_writeErr = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:          "writeErr",
		ImpureEffects: true,
		Description:   "writes to standard error",

		ParamPositional: []object.Parameter{
			object.Parameter{ExternalName: "text"},
		},
		ParamExpansionMax: -1,

		ParamKeyword: []object.Parameter{
			object.Parameter{ExternalName: "textmode", Type: object.BOOLEAN_OBJ},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		var out bytes.Buffer

		// "textmode" argument
		textMode := pr.Modes.ConsoleTextMode
		if args[1] != nil {
			textMode = args[1].IsTruthy()
		}

		args = args[0].(*object.List).Elements

		for _, v := range args {
			out.WriteString(v.String())
		}

		s := out.String()
		if len(s) == 0 {
			return object.NULL
		}

		text_io.PrintErr(s, textMode)
		return object.TRUE
	},
}

var bi_writelnErr = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:          "writelnErr",
		ImpureEffects: true,
		Description:   "writes to standard error, adding a newline at the end",

		ParamPositional: []object.Parameter{
			object.Parameter{ExternalName: "text"},
		},
		ParamExpansionMax: -1,

		ParamKeyword: []object.Parameter{
			object.Parameter{ExternalName: "textmode", Type: object.BOOLEAN_OBJ},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		var textMode object.Object
		if args[1] != nil {
			textMode = args[1].Copy()
		}

		// WARN: Calling a built-in directly has the potential to violate its signature.
		// Using Process.callback() is safer since it will check the arguments against the signature.
		return bi_writeErr.Fn.(BuiltInFunction)(pr,
			&object.List{Elements: append(args[0].(*object.List).Elements, newLine)},
			textMode)
	},
}

var bi_read = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:          "read",
		ImpureEffects: true,
		Description:   "reads from the console, validating the string is good by the regex or function passed, and giving the error message specified if the string is no good; If no alternate is given, this may ultimately generate an error.",

		ParamKeyword: []object.Parameter{
			object.Parameter{ExternalName: "prompt", Type: object.STRING_OBJ, DefaultValue: object.ZeroLengthString()},
			object.Parameter{ExternalName: "validation"},
			object.Parameter{ExternalName: "errmsg", Type: object.STRING_OBJ, DefaultValue: object.ZeroLengthString()},
			object.Parameter{ExternalName: "maxattempts", Type: object.NUMBER_OBJ, DefaultValue: object.One},
			object.Parameter{ExternalName: "alt"},
			object.Parameter{ExternalName: "textmode", Type: object.BOOLEAN_OBJ},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		const fnName = "read"

		var ok bool
		var err error

		// Gather arguments.
		// "prompt" argument
		prompt := args[0].String()

		// "validation" argument
		var fn object.Object
		var re *object.Regex
		validationByRegex := false

		if args[1] != nil {
			re, ok = args[1].(*object.Regex)
			if ok {
				validationByRegex = true
			} else {
				fn = args[1]
				if !object.IsCallable(fn) {
					return object.NewException(object.ERR_ARGUMENTS, fnName, "Expected function or regex for validation argument")
				}
			}
		}

		// "errmsg" argument
		errMsg := args[2].String()

		// "maxattempts" argument
		maxattempts, ok := object.NumberToInt(args[3])
		if !ok {
			return object.NewException(object.ERR_ARGUMENTS, fnName, "Expected integer for maximum attempts")
		}

		// "alt" argument
		alternate := args[4]

		// "textmode" argument
		textMode := pr.Modes.ConsoleTextMode
		if args[5] != nil {
			textMode = args[5].IsTruthy()
		}

		// parameters gathered...
		for i := 0; maxattempts == -1 || i < maxattempts; i++ {
			stop := false
			text_io.Print(prompt, textMode)

			var input string
			input, ok = text_io.ReadLn(textMode)
			if !ok {
				// FIXME:? user pressed ctrl-D?; action to take?
				// stop for loop after verification
				stop = true
			}

			if validationByRegex || fn != nil {
				var verify object.Object

				if validationByRegex {
					verify, err = object.RegexMatching(re, input)
				} else {
					verify, err = pr.callback(fn, object.NewString(input))
				}
				if err != nil {
					return object.NewException(object.ERR_GENERAL, fnName, err.Error())
				}
				if verify == object.TRUE {
					return object.NewString(input)
				} else {
					text_io.Print(errMsg, textMode)
				}

			} else {
				return object.NewString(input)
			}

			if stop {
				break
			}
		}

		if alternate == nil {
			return object.NewException(object.ERR_GENERAL, fnName, "Input failed to match expected")
		}
		return alternate
	},
}
