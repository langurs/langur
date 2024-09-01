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
		Description:   "exits with the integer code given; 0 if no code is given; msg as string to write to standard error, appending a newline, if code not 0",

		ParamByName: []object.Parameter{
			object.Parameter{ExternalName: "code", DefaultValue: object.Zero},
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
				code = system.GetExitStatus("argtoexitBad")
			}
			code = system.FixExitStatus(code)

		case *object.Boolean:
			//  true: success (code 0)
			// false: general failure
			if !codeArg.Value {
				code = system.GetExitStatus("")
			}

		default:
			// invalid code argument to exit()
			code = system.GetExitStatus("argtoexitBad")
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

		// TODO: update
		ParamPositional: []object.Parameter{
			object.Parameter{},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		count := 0

		switch over := args[0].(type) {
		case *object.Hash:
			arr := &object.List{}
			for _, kv := range over.Pairs {
				arr.Elements = append(arr.Elements, kv.Key)
			}
			return arr

		case *object.List:
			count = len(over.Elements)
		case *object.String:
			count = over.LenCP()
		case *object.Range:
			count = 2

		default:
			return object.NewError(object.ERR_ARGUMENTS, "keys", "Expected hash, list, string, or range")
		}

		numbers := make([]object.Object, count)

		for num := 1; num <= count; num++ {
			numbers[num-1] = object.NumberFromInt(num)
		}

		return &object.List{Elements: numbers}
	},
}

var bi_len = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:        "len",
		Description: "returns the length (as integer) of a list, hash, or string (in code points)",

		// TODO: update
		ParamPositional: []object.Parameter{
			object.Parameter{},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		// return integer giving the length
		switch arg := args[0].(type) {
		case *object.List:
			return object.NumberFromInt(len(arg.Elements))
		case *object.Hash:
			return object.NumberFromInt(len(arg.Pairs))
		case *object.String:
			// returns code point, not code unit length
			return object.NumberFromInt(arg.LenCP())
		case *object.Range:
			return object.NumberFromInt(2)
		}
		return object.NewError(object.ERR_ARGUMENTS, "len", "Expected list, hash, string, or range")
	},
}

var bi_nn = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:        "nn",
		Description: "nn(list, alternate); returns the first non-null value from a list, unless there are no non-null values, in which case it returns the alternate or an exception",

		// TODO: update
		ParamPositional: []object.Parameter{
			object.Parameter{},
		},
		ParamExpansionMin: 1,
		ParamExpansionMax: 2,
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		// FIXME: update parameters/args
		args = args[0].(*object.List).Elements

		arr, ok := args[0].(*object.List)
		if !ok {
			return object.NewError(object.ERR_ARGUMENTS, "nn", "Expected list for first argument")
		}
		for _, v := range arr.Elements {
			if v != object.NULL {
				return v
			}
		}
		if len(args) > 1 {
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

		// TODO: update
		ParamPositional: []object.Parameter{
			object.Parameter{},
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
