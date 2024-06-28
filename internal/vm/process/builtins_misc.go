// langur/vm/process/builtins_misc.go

package process

import (
	"langur/object"
	"langur/system"
	"os"
	"time"
)

// benchmark, exit, first, haskey, keys
// last, len, sleep, ticks, nn

func bi_benchmark(pr *Process, args ...object.Object) object.Object {
	const fnName = "benchmark"

	fn := args[0]
	if !object.IsCallable(fn) {
		return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected callable for first argument")
	}
	times := 1
	if len(args) > 1 {
		n, ok := object.NumberToInt(args[1])
		if !ok {
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected integer for second argument")
		}
		times = n
	}
	start := time.Now()
	for i := 0; i < times; i++ {
		_, err := pr.call(fn)
		if err != nil {
			return object.NewError(object.ERR_GENERAL, fnName, err.Error())
		}
	}
	return object.NewString(time.Since(start).String())
}

func bi_exit(pr *Process, args ...object.Object) object.Object {
	var err error
	code := 0 // 0 = success
	var str object.Object

	if len(args) > 0 {
		switch arg1 := args[0].(type) {
		case *object.Number:
			code, err = arg1.ToInt()
			if err != nil {
				// failure to convert to native integer
				code = system.GetExitStatus("argtoexitBad")
			}
			code = system.FixExitStatus(code)

		case *object.Boolean:
			//  true: success (code 0)
			// false: general failure
			if !arg1.Value {
				code = system.GetExitStatus("")
			}

		default:
			// invalid argument to exit()
			code = system.GetExitStatus("argtoexitBad")
		}

		if len(args) > 1 {
			str = args[1]
		}
	}

	if str != nil && code != 0 {
		// if non-zero return code, write string to standard error, appending a newline
		bi_writelnErr(pr, str)
	}
	os.Exit(code)

	// no need to return, but the compiler requires it...
	return object.NONE
}

// start of range or first element in list or string
func bi_first(pr *Process, args ...object.Object) object.Object {
	switch arg := args[0].(type) {
	case *object.List:
		if len(arg.Elements) > 0 {
			return arg.Elements[0]
		} else {
			return object.NewError(object.ERR_INDEX, "first", "Index out of range")
		}

	case *object.String:
		cp, ok := arg.IndexToCP(1)
		if !ok {
			return object.NewError(object.ERR_INDEX, "first", "String index out of range")
		}
		return object.NumberFromInt(int(cp))

	case *object.Range:
		return arg.Start
	}

	return object.NewError(object.ERR_ARGUMENTS, "first", "Expected list, string or range")
}

func bi_haskey(pr *Process, args ...object.Object) object.Object {
	obj, indexable := args[0].(object.IIndex)
	if indexable {
		return object.NativeBoolToObject(obj.IndexValid(args[1]))
	}
	return object.NONE
}

// return hash keys in a list, or keys of list, string, or range (always 1-based index)
func bi_keys(pr *Process, args ...object.Object) object.Object {
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
}

// end of range or last element in list or string
func bi_last(pr *Process, args ...object.Object) object.Object {
	switch arg := args[0].(type) {
	case *object.List:
		if len(arg.Elements) > 0 {
			return arg.Elements[len(arg.Elements)-1]
		} else {
			return object.NewError(object.ERR_INDEX, "last", "Index out of range")
		}

	case *object.String:
		if arg.LenCP() == 0 {
			return object.NewError(object.ERR_INDEX, "last", "String index out of range")
		}
		cpSlc := arg.RuneSlc()
		return object.NumberFromRune(cpSlc[len(cpSlc)-1])

	case *object.Range:
		return arg.End
	}

	return object.NewError(object.ERR_ARGUMENTS, "last", "Expected list, string or range")
}

func bi_len(pr *Process, args ...object.Object) object.Object {
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
}

// nn() non-null; returns first non-null object from lists passed, or alternate if there are none
func bi_nn(pr *Process, args ...object.Object) object.Object {
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
}

func bi_sleep(pr *Process, args ...object.Object) object.Object {
	d, ok := object.NumberToInt(args[0])
	if ok {
		if d < 1 {
			return object.FALSE
		}
		time.Sleep(time.Duration(d) * time.Millisecond)
		return object.TRUE
	}
	return object.NewError(object.ERR_ARGUMENTS, "sleep", "Expected number of milliseconds")
}

func bi_ticks(pr *Process, args ...object.Object) object.Object {
	return object.NumberFromInt64(time.Now().UnixNano())
}
