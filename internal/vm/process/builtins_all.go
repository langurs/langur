// langur/vm/process/builtins_all.go

package process

import (
	"langur/object"
)

// all, any

var bi_all = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:        "all",
		Description: "returns bool indicating whether the validation function or regex returns true for all elements of a list or hash, or null when given an empty list or hash",

		// TODO: update
		ParamPositional: []object.Parameter{
			object.Parameter{},
		},
		ParamExpansionMin: 1,
		ParamExpansionMax: 2,
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		const fnName = "all"

		// FIXME: update parameters/args
		args = args[0].(*object.List).Elements

		var isRegex bool
		var re *object.Regex
		var fn, over object.Object

		if len(args) == 2 {
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
			over = args[1]

		} else {
			over = args[0]
		}

		var result object.Object
		var err error

		switch arg := over.(type) {
		case *object.List:
			if len(arg.Elements) == 0 {
				return object.NONE
			}

			if isRegex {
				for _, v := range arg.Elements {
					result, err = object.RegexMatchingOrError(re, v)
					if err != nil {
						return object.NewError(object.ERR_GENERAL, fnName, err.Error())
					}
					if result != object.TRUE {
						return object.FALSE
					}
				}

			} else if fn != nil {
				for _, v := range arg.Elements {
					result, err = pr.callback(fn, v)
					if err != nil {
						return object.NewError(object.ERR_GENERAL, fnName, err.Error())
					}
					if result != object.TRUE {
						return object.FALSE
					}
				}

			} else {
				for _, v := range arg.Elements {
					if !v.IsTruthy() {
						return object.FALSE
					}
				}
			}

		case *object.Hash:
			if len(arg.Pairs) == 0 {
				return object.NONE
			}

			if isRegex {
				for _, kv := range arg.Pairs {
					result, err = object.RegexMatchingOrError(re, kv.Value)
					if err != nil {
						return object.NewError(object.ERR_GENERAL, fnName, err.Error())
					}
					if result != object.TRUE {
						return object.FALSE
					}
				}

			} else if fn != nil {
				for _, kv := range arg.Pairs {
					result, err = pr.callback(fn, kv.Value)
					if err != nil {
						return object.NewError(object.ERR_GENERAL, fnName, err.Error())
					}
					if result != object.TRUE {
						return object.FALSE
					}
				}

			} else {
				for _, kv := range arg.Pairs {
					if !kv.Value.IsTruthy() {
						return object.FALSE
					}
				}
			}

		default:
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected list or hash for second argument")
		}

		return object.TRUE
	},
}

var bi_any = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:        "any",
		Description: "returns bool indicating whether the validation function or regex returns true for any elements of a list or hash, or null when given an empty list or hash",

		// TODO: update
		ParamPositional: []object.Parameter{
			object.Parameter{},
		},
		ParamExpansionMin: 1,
		ParamExpansionMax: 2,
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		const fnName = "any"

		// FIXME: update parameters/args
		args = args[0].(*object.List).Elements

		var isRegex bool
		var re *object.Regex
		var fn, over object.Object

		if len(args) == 2 {
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
			over = args[1]

		} else {
			over = args[0]
		}

		var result object.Object
		var err error

		switch arg := over.(type) {
		case *object.List:
			if len(arg.Elements) == 0 {
				return object.NONE
			}

			if isRegex {
				for _, v := range arg.Elements {
					result, err = object.RegexMatchingOrError(re, v)
					if err != nil {
						return object.NewError(object.ERR_GENERAL, fnName, err.Error())
					}
					if result == object.TRUE {
						return object.TRUE
					}
				}

			} else if fn != nil {
				for _, v := range arg.Elements {
					result, err = pr.callback(fn, v)
					if err != nil {
						return object.NewError(object.ERR_GENERAL, fnName, err.Error())
					}
					if result == object.TRUE {
						return object.TRUE
					}
				}

			} else {
				for _, v := range arg.Elements {
					if v.IsTruthy() {
						return object.TRUE
					}
				}
			}

		case *object.Hash:
			if len(arg.Pairs) == 0 {
				return object.NONE
			}

			if isRegex {
				for _, kv := range arg.Pairs {
					result, err = object.RegexMatchingOrError(re, kv.Value)
					if err != nil {
						return object.NewError(object.ERR_GENERAL, fnName, err.Error())
					}
					if result == object.TRUE {
						return object.TRUE
					}
				}

			} else if fn != nil {
				for _, kv := range arg.Pairs {
					result, err = pr.callback(fn, kv.Value)
					if err != nil {
						return object.NewError(object.ERR_GENERAL, fnName, err.Error())
					}
					if result == object.TRUE {
						return object.TRUE
					}
				}

			} else {
				for _, kv := range arg.Pairs {
					if kv.Value.IsTruthy() {
						return object.TRUE
					}
				}
			}

		default:
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected list or hash for second argument")
		}

		return object.FALSE
	},
}
