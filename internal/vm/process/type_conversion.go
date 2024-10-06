// langur/vm/process/type_conversion.go

package process

import (
	"langur/common"
	"langur/object"
	"langur/str"
)

// string, number, hash, datetime, duration, bool

var bi_string = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name: common.StringType,

		ParamPositional: []object.Parameter{
			object.Parameter{ExternalName: "from"},
		},

		ParamByName: []object.Parameter{
			object.Parameter{ExternalName: "fmt"},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		const fnName = common.StringType

		from, format := args[0], args[1]

		if format != nil {
			switch n := from.(type) {
			case *object.Number:
				base, ok := object.NumberToInt(format)
				if !ok {
					return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected integer for argument fmt when converting number to string")
				}

				i, err := n.ToInt()
				if err != nil {
					return object.NewError(object.ERR_ARGUMENTS, fnName, "Failed to convert number to string using the given base")
				}
				return object.NewString(str.IntToStr(i, base))

			case *object.DateTime:
				format, ok := format.(*object.String)
				if !ok {
					return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected string for argument fmt when converting datetime to string")
				}
				s := n.FormatString(format.String())
				return object.NewString(s)

			default:
				return object.NewError(object.ERR_ARGUMENTS, fnName, "Unexpected argument fmt for this conversion")
			}
		}
		return object.ToString(from)
	},
}

var bi_number = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name: common.NumberType,

		ParamPositional: []object.Parameter{
			object.Parameter{ExternalName: "from"},
		},

		ParamByName: []object.Parameter{
			object.Parameter{ExternalName: "fmt"},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		const fnName = common.NumberType

		from, format := args[0], args[1]

		var ok bool
		base := 10
		if format != nil {
			base, ok = object.NumberToInt(format)
			if !ok {
				return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected integer for argument fmt")
			}
		}
		n, ok := object.ToNumber(from, base)
		if !ok {
			return object.NewError(object.ERR_GENERAL, fnName, "Failed to convert to number")
		}
		return n
	},
}

var bi_hash = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name: common.HashType,

		ParamPositional: []object.Parameter{
			object.Parameter{ExternalName: "from"},
		},
		ParamExpansionMin: 1,
		ParamExpansionMax: 2,
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		const fnName = common.HashType

		args = args[0].(*object.List).Elements

		var list1, list2 *object.List
		var err error

		switch arg := args[0].(type) {
		case *object.List:
			list1 = arg

		case *object.Range:
			list1, err = arg.ToList()
			if err != nil {
				return object.NewError(object.ERR_ARGUMENTS, fnName, err.Error())
			}

		case *object.DateTime:
			if len(args) != 1 {
				return object.NewError(object.ERR_ARGUMENTS, fnName, "Unexpected second argument when given date-time")
			}
			return arg.ToHash()

		case *object.Duration:
			if len(args) != 1 {
				return object.NewError(object.ERR_ARGUMENTS, fnName, "Unexpected second argument when given duration")
			}
			return arg.ToHash()

		case *object.Hash:
			// pass-through hash
			if len(args) != 1 {
				return object.NewError(object.ERR_ARGUMENTS, fnName, "Unexpected second argument when given hash")
			}
			return arg

		default:
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected list or date-time for first argument")
		}

		if len(args) == 1 {
			hash, err := object.NewHashFromSlice(list1.Elements, false)
			if err != nil {
				return object.NewError(object.ERR_GENERAL, fnName, err.Error())
			}
			return hash
		}

		switch arg2 := args[1].(type) {
		case *object.List:
			list2 = arg2
		case *object.Range:
			list2, err = arg2.ToList()
			if err != nil {
				return object.NewError(object.ERR_ARGUMENTS, fnName, err.Error())
			}
		default:
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected list or range for second argument")
		}

		hash, err := object.NewHashFromParallelSlices(list1.Elements, list2.Elements, false)
		if err != nil {
			return object.NewError(object.ERR_GENERAL, fnName, err.Error())
		}
		return hash
	},
}

var bi_datetime = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name: common.DateTimeType,

		ParamPositional: []object.Parameter{
			object.Parameter{ExternalName: "from"},
		},

		ParamByName: []object.Parameter{
			object.Parameter{ExternalName: "fmt"},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		const fnName = common.DateTimeType

		from, format := args[0], args[1]

		fmtStr := ""

		switch from.(type) {
		case *object.String:
			switch format.(type) {
			case nil:
				// ok; no format string

			case *object.String:
				fmtStr = from.String()

			default:
				return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected string for argument fmt")
			}

		case *object.DateTime:
			switch format.(type) {
			case nil:
				// pass through unaltered
				return from

			case *object.String:
				// change time zone
				fmtStr = format.String()

			// case *object.Number:
			// offset in seconds?

			default:
				return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected string for argument fmt (time zone/location string) when converting datetime to datetime")
			}
		}

		dt, err := object.ToDateTime(from, fmtStr)
		if err != nil {
			return object.NewError(object.ERR_GENERAL, fnName, err.Error())
		}

		return dt
	},
}

var bi_duration = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name: common.DurationType,

		ParamPositional: []object.Parameter{
			object.Parameter{ExternalName: "from"},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		const fnName = common.DurationType

		switch arg := args[0].(type) {
		case *object.String:
			o, err := object.NewDurationFromString(arg.String())
			if err != nil {
				return object.NewError(object.ERR_GENERAL, fnName, err.Error())
			}
			return o

		case *object.Number:
			n, err := arg.ToInt64()
			if err != nil {
				return object.NewError(object.ERR_GENERAL, fnName, "Number not an INTEGER")
			}
			return object.NewDurationFromNanoseconds(n)

		case *object.Hash:
			o, err := arg.ToDuration()
			if err != nil {
				return object.NewError(object.ERR_GENERAL, fnName, err.Error())
			}
			return o

		case *object.Duration:
			return arg

		default:
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected string, number, hash, or duration")
		}
	},
}

var bi_bool = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name: common.BooleanType,

		ParamPositional: []object.Parameter{
			object.Parameter{ExternalName: "from"},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		return object.NativeBoolToObject(args[0].IsTruthy())
	},
}
