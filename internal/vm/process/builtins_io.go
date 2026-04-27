// langur/vm/process/builtins_io.go

package process

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"langur/object"
	"langur/str"
	"os"
)

// write, writeln, writeErr, writelnErr, read

var bi_write = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:          "write",
		ImpureEffects: true,
		Description:   "writes to the console",

		ParamPositional: []object.Parameter{
			object.Parameter{},
		},
		ParamExpansionMax: -1,
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		var out bytes.Buffer

		args = args[0].(*object.List).Elements

		for _, v := range args {
			out.WriteString(v.String())
		}

		s := out.String()
		if pr.Modes.ConsoleTextMode {
			s = str.ReplaceNewLinesWithSystem(s)
		}

		if len(s) == 0 {
			return object.NULL
		}
		fmt.Print(s)
		return object.TRUE
	},
}

var bi_writeln = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:          "writeln",
		ImpureEffects: true,
		Description:   "writes to the console, adding a system newline at the end",

		ParamPositional: []object.Parameter{
			object.Parameter{},
		},
		ParamExpansionMax: -1,
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		return bi_write.Fn.(BuiltInFunction)(pr,
			&object.List{
				Elements: append(args[0].(*object.List).Elements, object.NewString(str.SysNewLine))})
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
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		var out bytes.Buffer

		args = args[0].(*object.List).Elements

		for _, v := range args {
			out.WriteString(v.String())
		}

		s := out.String()
		if pr.Modes.ConsoleTextMode {
			s = str.ReplaceNewLinesWithSystem(s)
		}

		if len(s) == 0 {
			return object.NULL
		}
		fmt.Fprint(os.Stderr, s)
		return object.TRUE
	},
}

var bi_writelnErr = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:          "writelnErr",
		ImpureEffects: true,
		Description:   "writes to standard error, adding a system newline at the end",

		ParamPositional: []object.Parameter{
			object.Parameter{},
		},
		ParamExpansionMax: -1,
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		return bi_writeErr.Fn.(BuiltInFunction)(pr,
			&object.List{
				Elements: append(args[0].(*object.List).Elements, object.NewString(str.SysNewLine))})
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
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		const fnName = "read"

		var ok bool

		// Gather arguments.
		// "prompt" argument
		prompt := args[0].String()
		if pr.Modes.ConsoleTextMode {
			prompt = str.ReplaceNewLinesWithSystem(prompt)
		}

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
					return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected function or regex for validation argument")
				}
			}
		}

		// "errmsg" argument
		errMsg := args[2].String()
		if pr.Modes.ConsoleTextMode {
			errMsg = str.ReplaceNewLinesWithSystem(errMsg)
		}

		// "maxattempts" argument
		maxattempts, ok := object.NumberToInt(args[3])
		if !ok {
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected integer for maximum attempts")
		}

		// "alt" argument
		alternate := args[4]

		// parameters gathered...
		for i := 0; maxattempts == -1 || i < maxattempts; i++ {
			fmt.Print(prompt)
			input, err := readLine(os.Stdin)
			if err != nil {
				return object.NewError(object.ERR_GENERAL, fnName, err.Error())
			}

			if pr.Modes.ConsoleTextMode {
				input = str.ReplaceNewLinesWithLinux(input)
			}

			if validationByRegex || fn != nil {
				var verify object.Object

				if validationByRegex {
					verify, err = object.RegexMatching(re, input)
				} else {
					verify, err = pr.callback(fn, object.NewString(input))
				}
				if err != nil {
					return object.NewError(object.ERR_GENERAL, fnName, err.Error())
				}
				if verify == object.TRUE {
					return object.NewString(input)
				} else {
					fmt.Print(errMsg)
				}

			} else {
				return object.NewString(input)
			}
		}

		if alternate == nil {
			return object.NewError(object.ERR_GENERAL, fnName, "Input failed to match expected")
		}
		return alternate
	},
}

func readLine(in io.Reader) (string, error) {
	scanner := bufio.NewScanner(in)
	scanned := scanner.Scan()
	if !scanned {
		return "", fmt.Errorf("Unknown failure to scan input text")
	}
	return scanner.Text(), nil
}
