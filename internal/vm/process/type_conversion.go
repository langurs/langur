// langur/vm/process/type_conversion.go

package process

import (
	"langur/common"
	"langur/object"
	"langur/str"
)

// used internally; might change later
func bi_is_hash(pr *Process, args ...object.Object) object.Object {
	return object.NativeBoolToObject(args[0].Type() == object.HASH_OBJ)
}

// string, number, hash, datetime, duration, bool

func bi_string(pr *Process, args ...object.Object) object.Object {
	const fnName = common.StringType

	if len(args) > 1 {
		switch n := args[0].(type) {
		case *object.Number:
			base, ok := object.NumberToInt(args[1])
			if !ok {
				return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected integer for second argument (base)")
			}

			i, err := n.ToInt()
			if err != nil {
				return object.NewError(object.ERR_ARGUMENTS, fnName, "Failed to convert number to string using the given base")
			}
			return object.NewString(str.IntToStr(i, base))

		case *object.DateTime:
			format, ok := args[1].(*object.String)
			if !ok {
				return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected string for second argument (date-time format)")
			}
			s := n.FormatString(format.String())
			return object.NewString(s)

		default:
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected number for first argument when given a base for second")
		}
	}
	return object.ToString(args[0])
}

func bi_number(pr *Process, args ...object.Object) object.Object {
	const fnName = common.NumberType

	var ok bool
	base := 10
	if len(args) > 1 {
		base, ok = object.NumberToInt(args[1])
		if !ok {
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected integer for second argument (base)")
		}
	}
	n, ok := object.ToNumber(args[0], base)
	if !ok {
		return object.NewError(object.ERR_GENERAL, fnName, "Failed to convert to number")
	}
	return n
}

func bi_hash(pr *Process, args ...object.Object) object.Object {
	const fnName = common.HashType

	if len(args) == 0 {
		// empty hash
		return &object.Hash{}
	}

	var arr1, arr2 *object.List
	var err error

	switch arg := args[0].(type) {
	case *object.List:
		arr1 = arg

	case *object.Range:
		arr1, err = arg.ToList()
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
		hash, err := object.NewHashFromSlice(arr1.Elements, false)
		if err != nil {
			return object.NewError(object.ERR_GENERAL, fnName, err.Error())
		}
		return hash
	}

	switch arg2 := args[1].(type) {
	case *object.List:
		arr2 = arg2
	case *object.Range:
		arr2, err = arg2.ToList()
		if err != nil {
			return object.NewError(object.ERR_ARGUMENTS, fnName, err.Error())
		}
	default:
		return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected list or range for second argument")
	}

	hash, err := object.NewHashFromParallelSlices(arr1.Elements, arr2.Elements, false)
	if err != nil {
		return object.NewError(object.ERR_GENERAL, fnName, err.Error())
	}
	return hash
}

func bi_datetime(pr *Process, args ...object.Object) object.Object {
	const fnName = common.DateTimeType

	format := ""
	var ok bool
	var dt *object.DateTime
	var err error

	switch args[0].(type) {
	case *object.String:
		if len(args) > 1 {
			format, ok = object.ExpectString(args[1])
			if !ok {
				return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected string for second argument (format string)")
			}
		}

	case *object.DateTime:
		if len(args) == 1 {
			// pass through unaltered
			return args[0]

		} else {
			// change time zone
			switch args[1].(type) {
			case *object.String:
				format = args[1].String()

			// case *object.Number:
			// offset in seconds?

			default:
				return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected string for second argument (time zone/location string)")
			}
		}
	}

	dt, err = object.ToDateTime(args[0], format)
	if err != nil {
		return object.NewError(object.ERR_GENERAL, fnName, err.Error())
	}

	return dt
}

func bi_duration(pr *Process, args ...object.Object) object.Object {
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
}

func bi_bool(pr *Process, args ...object.Object) object.Object {
	return object.NativeBoolToObject(args[0].IsTruthy())
}
