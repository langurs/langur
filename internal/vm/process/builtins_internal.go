// langur/vm/process/builtins_internal.go

package process

import (
	"langur/object"
)

// _limit, _values, _keys, _len, _ishash

var bi__limit = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name: "_limit",

		ParamPositional: []object.Parameter{
			object.Parameter{ExternalName: "over"},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
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
	},
}

var bi__values = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name: "_values",

		ParamPositional: []object.Parameter{
			object.Parameter{ExternalName: "over"},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		// for for in loop values
		const fnName = "for_in"

		var start, end, num int64

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
			list, err := over.ToList()
			if err != nil {
				return object.NewError(object.ERR_ARGUMENTS, fnName, err.Error())
			}
			return list

		case *object.Range:
			list, err := over.ToList()
			if err != nil {
				return object.NewError(object.ERR_ARGUMENTS, fnName, err.Error())
			}
			return list

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
	},
}

var bi__keys = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:            "_keys",
		ParamPositional: bi_keys.FnSignature.ParamPositional,
		ParamByName:     bi_keys.FnSignature.ParamByName,
	},
	Fn: bi_keys.Fn,
}

var bi__len = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:            "_len",
		ParamPositional: bi_len.FnSignature.ParamPositional,
		ParamByName:     bi_len.FnSignature.ParamByName,
	},
	Fn: bi_len.Fn,
}

var bi__ishash = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name: "_is_hash",
		ParamPositional: []object.Parameter{
			object.Parameter{},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		return object.NativeBoolToObject(args[0].Type() == object.HASH_OBJ)
	},
}
