// langur/vm/process/builtins_misc.go

package process

import (
	"fmt"
	"langur/object"
	"langur/str"
	"langur/system"
	"os"
	"time"
)

// exit, keys
// len, sleep, ticks, nn

var bi_exit = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:          "exit",
		ImpureEffects: true,
		Description:   "exits with the integer code given; msg as string to write to standard error, appending a newline, if code not 0",

		ParamPositional: []object.Parameter{
			object.Parameter{ExternalName: "code"},
		},

		ParamByName: []object.Parameter{
			object.Parameter{ExternalName: "msg", DefaultValue: object.ZLS},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		var err error
		code := 0 // 0 = success
		strArg := args[1]

		switch codeArg := args[0].(type) {
		case *object.Number:
			code, err = codeArg.ToInt()
			if err != nil {
				// failure to convert to native integer
				code = system.GetExitStatus(system.ExitStatusArgToExitBad)
			}
			code = system.FixExitStatus(code)

		case *object.Boolean:
			//  true: success (code 0)
			// false: general failure
			if !codeArg.Value {
				code = system.GetExitStatus(system.ExitStatusGeneral)
			}

		default:
			// invalid code argument to exit()
			code = system.GetExitStatus(system.ExitStatusArgToExitBad)
		}

		if code != 0 && strArg.IsTruthy() {
			// if non-zero return code, write string to standard error, appending a newline
			s := strArg.String()
			if pr.Modes.ConsoleTextMode {
				s = str.ReplaceNewLinesWithSystem(s)
			}
			if len(s) != 0 {
				fmt.Fprint(os.Stderr, s)
			}
		}
		os.Exit(code)

		// no need to return, but the compiler requires it...
		return object.NONE
	},
}

var bi_keys = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:        "keys",
		Description: "returns the keys (as list) of a hash, or list or string indices (always 1-based index)",

		ParamPositional: []object.Parameter{
			object.Parameter{ExternalName: "over"},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		switch over := args[0].(type) {
		case object.IIndex:
			return over.IndexKeys()

		default:
			return object.NewError(object.ERR_ARGUMENTS, "keys", "Expected indexable item")
		}
	},
}

var bi_len = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:        "len",
		Description: "returns the index count of an indexable item",

		ParamPositional: []object.Parameter{
			object.Parameter{ExternalName: "over"},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		switch arg := args[0].(type) {
		case object.IIndex:
			return object.NumberFromInt(arg.IndexCount())
		}
		return object.NewError(object.ERR_ARGUMENTS, "len", "Expected indexable item")
	},
}

var bi_nn = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:        "nn",
		Description: "returns the first non-null value from a list, unless there are no non-null values, in which case it returns the alternate or an exception",

		ParamPositional: []object.Parameter{
			object.Parameter{ExternalName: "over"},
		},

		ParamByName: []object.Parameter{
			object.Parameter{ExternalName: "alt"},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		list, ok := args[0].(*object.List)
		if !ok {
			return object.NewError(object.ERR_ARGUMENTS, "nn", "Expected list for first argument")
		}
		for _, v := range list.Elements {
			if v != object.NULL {
				return v
			}
		}

		if args[1] != nil {
			// return alternate
			return args[1]
		}
		return object.NewError(object.ERR_ARGUMENTS, "nn", "No suitable value found")
	},
}

var bi_sleep = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:          "sleep",
		ImpureEffects: true,
		Description:   "waits for the specified number of milliseconds",

		ParamPositional: []object.Parameter{
			object.Parameter{ExternalName: "ms"},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		d, ok := object.NumberToInt(args[0])
		if ok {
			if d < 1 {
				return object.FALSE
			}
			time.Sleep(time.Duration(d) * time.Millisecond)
			return object.TRUE
		}
		return object.NewError(object.ERR_ARGUMENTS, "sleep", "Expected number of milliseconds")
	},
}

var bi_ticks = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:        "ticks",
		Description: "returns Unix ticks in nanoseconds",
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		return object.NumberFromInt64(time.Now().UnixNano())
	},
}
