// langur/vm/process/builtins_internal.go

package process

import (
	"langur/object"
)

// _limit, _values

func bi__limit(pr *Process, args ...object.Object) object.Object {
	// for for of loop limit
	const fnName = "for_of"

	count := 0
	var ok bool

	switch over := args[0].(type) {
	case *object.Number:
		count, ok = object.NumberToInt(over)
		if !ok {
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected integer")
		}

	case *object.List:
		count = len(over.Elements)

	case *object.Range:
		var start, end int
		start, ok = object.NumberToInt(over.Start)
		if !ok {
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected integer range")
		}
		end, ok = object.NumberToInt(over.End)
		if !ok {
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected integer range")
		}
		count = end - start + 1

	case *object.String:
		count = len(over.String())

	default:
		return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected list, integer, integer range, or string")
	}

	// negative range or number on for of gives no result; not an error
	if count < 0 {
		count = 0
	}

	return object.NumberFromInt(count)
}

func bi__values(pr *Process, args ...object.Object) object.Object {
	// for for in loop values
	const fnName = "for_in"

	var start, end, num int64
	var err error

	switch over := args[0].(type) {
	case *object.List, *object.String:
		return over

	case *object.Hash:
		arr := &object.List{Elements: make([]object.Object, len(over.Pairs))}
		for i, kv := range over.Pairs {
			arr.Elements[i] = kv.Value
		}
		return arr

	case *object.Number:
		// number as implicit ascending range
		end, err = over.ToInt64()
		if err != nil {
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected integer")
		}

		if end == 0 {
			// done
			return &object.List{}

		} else if end < 0 {
			// negative number
			start = end
			end = -1

		} else {
			start = 1
		}

	case *object.Range:
		switch e := over.Start.(type) {
		case *object.Number:
			start, err = e.ToInt64()
			if err != nil {
				return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected integer range")
			}
		default:
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected integer range")
		}
		switch e := over.End.(type) {
		case *object.Number:
			end, err = e.ToInt64()
			if err != nil {
				return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected integer range")
			}
		default:
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected integer range")
		}

	default:
		return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected list, hash, integer, integer range, or string")
	}

	num = start
	numbers := []object.Object{}

	if start > end {
		// descending range
		for {
			numbers = append(numbers, object.NumberFromInt64(num))
			num--
			if num < end {
				break
			}
		}

	} else {
		for {
			numbers = append(numbers, object.NumberFromInt64(num))
			num++
			if num > end {
				break
			}
		}
	}

	return &object.List{Elements: numbers}
}
