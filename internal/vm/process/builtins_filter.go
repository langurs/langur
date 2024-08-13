// langur/vm/process/builtins_filter.go

package process

import (
	"langur/object"
)

// filter, count

// return all values (as list or hash) returning true from passed regex or function
var bi_filter = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:        "filter",
		Description: "returns list (or hash) of values verified by given function or regex, or an empty list or hash if there are no matches",

		// TODO: update
		ParamPositional: []object.Parameter{
			object.Parameter{},
		},
		ParamExpansionMin: 1,
		ParamExpansionMax: 2,
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		const fnName = "filter"

		// FIXME: update parameters/args
		args = args[0].(*object.List).Elements

		var isRegex bool
		var re *object.Regex
		var fn, over object.Object

		if len(args) == 1 {
			over = args[0]
		} else {
			over = args[1]
			if object.IsCallable(args[0]) {
				fn = args[0]
			} else {
				var ok bool
				re, ok = args[0].(*object.Regex)
				if !ok {
					return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected regex or callable for first argument")
				}
				isRegex = true
			}
		}

		var result object.Object
		var err error

		switch arg := over.(type) {
		case *object.List:
			newArr := &object.List{}
			if isRegex {
				for _, v := range arg.Elements {
					result, err = object.RegexMatchingOrError(re, v)
					if err != nil {
						return object.NewError(object.ERR_GENERAL, fnName, err.Error())
					}
					if result == object.TRUE {
						newArr.Elements = append(newArr.Elements, v)
					}
				}

			} else if fn != nil {
				for _, v := range arg.Elements {
					result, err = pr.callback(fn, v)
					if err != nil {
						return object.NewError(object.ERR_GENERAL, fnName, err.Error())
					}
					if result == object.TRUE {
						newArr.Elements = append(newArr.Elements, v)
					}
				}

			} else {
				for _, v := range arg.Elements {
					if v.IsTruthy() {
						newArr.Elements = append(newArr.Elements, v)
					}
				}
			}
			return newArr

		case *object.Hash:
			elements := make([]object.Object, 0, len(arg.Pairs)*2)
			if isRegex {
				for _, kv := range arg.Pairs {
					result, err = object.RegexMatchingOrError(re, kv.Value)
					if err != nil {
						return object.NewError(object.ERR_GENERAL, fnName, err.Error())
					}
					if result == object.TRUE {
						elements = append(elements, kv.Key, kv.Value)
					}
				}

			} else if fn != nil {
				for _, kv := range arg.Pairs {
					result, err = pr.callback(fn, kv.Value)
					if err != nil {
						return object.NewError(object.ERR_GENERAL, fnName, err.Error())
					}
					if result == object.TRUE {
						elements = append(elements, kv.Key, kv.Value)
					}
				}

			} else {
				for _, kv := range arg.Pairs {
					if kv.Value.IsTruthy() {
						elements = append(elements, kv.Key, kv.Value)
					}
				}
			}

			hash, err := object.NewHashFromSlice(elements, false)
			if err != nil {
				return object.NewError(object.ERR_GENERAL, fnName, err.Error())
			}
			return hash
		}

		return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected list or hash for second argument")
	},
}

// like filter(), but returning a count instead
var bi_count = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:        "count",
		Description: "returns count of values verified by given function or regex",

		// TODO: update
		ParamPositional: []object.Parameter{
			object.Parameter{},
		},
		ParamExpansionMin: 1,
		ParamExpansionMax: 2,
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		const fnName = "count"

		// FIXME: update parameters/args
		args = args[0].(*object.List).Elements

		var isRegex bool
		var re *object.Regex
		var fn, over object.Object
		var count int64

		if len(args) == 1 {
			over = args[0]
		} else {
			over = args[1]
			if object.IsCallable(args[0]) {
				fn = args[0]
			} else {
				var ok bool
				re, ok = args[0].(*object.Regex)
				if !ok {
					return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected regex or callable for first argument")
				}
				isRegex = true
			}
		}

		var result object.Object
		var err error

		switch arg := over.(type) {
		case *object.List:
			if isRegex {
				for _, v := range arg.Elements {
					result, err = object.RegexMatchingOrError(re, v)
					if err != nil {
						return object.NewError(object.ERR_GENERAL, fnName, err.Error())
					}
					if result == object.TRUE {
						count++
					}
				}

			} else if fn != nil {
				for _, v := range arg.Elements {
					result, err = pr.callback(fn, v)
					if err != nil {
						return object.NewError(object.ERR_GENERAL, fnName, err.Error())
					}
					if result == object.TRUE {
						count++
					}
				}

			} else {
				for _, v := range arg.Elements {
					if v.IsTruthy() {
						count++
					}
				}
			}

		case *object.Hash:
			if isRegex {
				for _, kv := range arg.Pairs {
					result, err = object.RegexMatchingOrError(re, kv.Value)
					if err != nil {
						return object.NewError(object.ERR_GENERAL, fnName, err.Error())
					}
					if result == object.TRUE {
						count++
					}
				}

			} else if fn != nil {
				for _, kv := range arg.Pairs {
					result, err = pr.callback(fn, kv.Value)
					if err != nil {
						return object.NewError(object.ERR_GENERAL, fnName, err.Error())
					}
					if result == object.TRUE {
						count++
					}
				}

			} else {
				for _, kv := range arg.Pairs {
					if kv.Value.IsTruthy() {
						count++
					}
				}
			}

		default:
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected list or hash for second argument")
		}

		return object.NumberFromInt64(count)
	},
}
