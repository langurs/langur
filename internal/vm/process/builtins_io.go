// langur/vm/process/builtins_io.go

package process

import (
	"bytes"
	"langur/io"
	"langur/object"
	"langur/str"
)

// write, writeln, writeErr, writelnErr, read

var newLine = object.NewString(str.SysNewLine)

var bi_write = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:          "write",
		ImpureEffects: true,
		Description:   "writes to the console",

		ParamPositional: []object.Parameter{
			object.Parameter{},
		},
		ParamExpansionMax: -1,

		ParamByName: []object.Parameter{
			object.Parameter{ExternalName: "textmode", Type: object.BOOLEAN_OBJ},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		var out bytes.Buffer

		args = args[0].(*object.List).Elements

		// "textmode" argument
		textMode := pr.Modes.ConsoleTextMode
		if args[1] != nil {
			textMode = args[1].IsTruthy()
		}

		for _, v := range args {
			out.WriteString(v.String())
		}

		s := out.String()
		if len(s) == 0 {
			return object.NULL
		}

		io.Print(s, textMode)
		return object.TRUE
	},
}

var bi_writeln = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:          "writeln",
		ImpureEffects: true,
		Description:   "writes to the console, adding a newline at the end",

		ParamPositional: []object.Parameter{
			object.Parameter{},
		},
		ParamExpansionMax: -1,

		ParamByName: []object.Parameter{
			object.Parameter{ExternalName: "textmode", Type: object.BOOLEAN_OBJ},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		textMode := &object.NameValue{Name: "textmode", Value: args[1]}

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
			object.Parameter{},
		},
		ParamExpansionMax: -1,

		ParamByName: []object.Parameter{
			object.Parameter{ExternalName: "textmode", Type: object.BOOLEAN_OBJ},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		var out bytes.Buffer

		args = args[0].(*object.List).Elements

		// "textmode" argument
		textMode := pr.Modes.ConsoleTextMode
		if args[1] != nil {
			textMode = args[1].IsTruthy()
		}

		for _, v := range args {
			out.WriteString(v.String())
		}

		s := out.String()
		if len(s) == 0 {
			return object.NULL
		}

		io.PrintErr(s, textMode)
		return object.TRUE
	},
}

var bi_writelnErr = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:          "writelnErr",
		ImpureEffects: true,
		Description:   "writes to standard error, adding a newline at the end",

		ParamPositional: []object.Parameter{
			object.Parameter{},
		},
		ParamExpansionMax: -1,

		ParamByName: []object.Parameter{
			object.Parameter{ExternalName: "textmode", Type: object.BOOLEAN_OBJ},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		textMode := &object.NameValue{Name: "textmode", Value: args[1]}

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

		ParamByName: []object.Parameter{
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
			io.Print(prompt, textMode)
			input, err := io.ReadLn(textMode)
			if err != nil {
				return object.NewException(object.ERR_GENERAL, fnName, err.Error())
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
					io.Print(errMsg, textMode)
				}

			} else {
				return object.NewString(input)
			}
		}

		if alternate == nil {
			return object.NewException(object.ERR_GENERAL, fnName, "Input failed to match expected")
		}
		return alternate
	},
}
