// langur/vm/process/builtins_misc.go

package process

import (
	"langur/object"
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

		ParamKeyword: []object.Parameter{
			object.Parameter{ExternalName: "msg", DefaultValue: object.ZeroLengthString()},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		object.Exit(args[0], args[1], pr.Modes.ConsoleTextMode)

		// no need to return, but the Go compiler requires it...
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
			return object.NewException(object.ERR_ARGUMENTS, "keys", "Expected indexable item")
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
		return object.NewException(object.ERR_ARGUMENTS, "len", "Expected indexable item")
	},
}

var bi_nn = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:        "nn",
		Description: "returns the first non-null value from a list, unless there are no non-null values, in which case it returns the alternate or an exception",

		ParamPositional: []object.Parameter{
			object.Parameter{ExternalName: "over", Type: object.LIST_OBJ},
		},

		ParamKeyword: []object.Parameter{
			object.Parameter{ExternalName: "alt"},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		list := args[0].(*object.List)
		for _, v := range list.Elements {
			if v != object.NULL {
				return v
			}
		}

		if args[1] != nil {
			// return alternate
			return args[1]
		}
		return object.NewException(object.ERR_ARGUMENTS, "nn", "No suitable value found")
	},
}

var bi_sleep = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:          "sleep",
		ImpureEffects: true,
		Description:   "waits for the specified number of milliseconds",

		ParamPositional: []object.Parameter{
			object.Parameter{ExternalName: "ms", Type: object.NUMBER_OBJ},
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
		return object.NewException(object.ERR_ARGUMENTS, "sleep", "Expected number of milliseconds")
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
