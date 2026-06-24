// langur/vm/process/builtins_all.go

package process

import (
	"langur/object"
)

// all, any

var bi_all = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:        "all",
		Description: "returns bool indicating whether the validation function or pattern returns true for all elements of a list or hash, or null when given an empty list or hash",
		ParamPositional: []object.Parameter{
			object.Parameter{ExternalName: "over"},
		},
		ParamKeyword: []object.Parameter{
			object.Parameter{ExternalName: "by"},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		const fnName = "all"

		var isPattern bool
		var re *object.Pattern

		over, by := args[0], args[1]

		if by != nil {
			if !object.IsCallable(by) {
				var ok bool
				re, ok = by.(*object.Pattern)
				if !ok {
					return object.NewException(object.ERR_ARGUMENTS, fnName, "Expected pattern or callable for validation argument")
				}
				isPattern = true
			}
		}

		var result object.Object
		var err error

		switch arg := over.(type) {
		case *object.List:
			if len(arg.Elements) == 0 {
				return object.NONE
			}

			if isPattern {
				for _, v := range arg.Elements {
					result, err = object.PatternMatchingOrError(re, v)
					if err != nil {
						return object.NewException(object.ERR_GENERAL, fnName, err.Error())
					}
					if result != object.TRUE {
						return object.FALSE
					}
				}

			} else if by != nil {
				for _, v := range arg.Elements {
					result, err = pr.callback(by, v)
					if err != nil {
						return object.NewException(object.ERR_GENERAL, fnName, err.Error())
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

			if isPattern {
				for _, kv := range arg.Pairs {
					result, err = object.PatternMatchingOrError(re, kv.Value)
					if err != nil {
						return object.NewException(object.ERR_GENERAL, fnName, err.Error())
					}
					if result != object.TRUE {
						return object.FALSE
					}
				}

			} else if by != nil {
				for _, kv := range arg.Pairs {
					result, err = pr.callback(by, kv.Value)
					if err != nil {
						return object.NewException(object.ERR_GENERAL, fnName, err.Error())
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
			return object.NewException(object.ERR_ARGUMENTS, fnName, "Expected list or hash")
		}

		return object.TRUE
	},
}

var bi_any = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:        "any",
		Description: "returns bool indicating whether the validation function or pattern returns true for any elements of a list or hash, or null when given an empty list or hash",
		ParamPositional: []object.Parameter{
			object.Parameter{ExternalName: "over"},
		},
		ParamKeyword: []object.Parameter{
			object.Parameter{ExternalName: "by"},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		const fnName = "any"

		var isPattern bool
		var re *object.Pattern

		over, by := args[0], args[1]

		if by != nil {
			if !object.IsCallable(by) {
				var ok bool
				re, ok = by.(*object.Pattern)
				if !ok {
					return object.NewException(object.ERR_ARGUMENTS, fnName, "Expected pattern or callable for validation argument")
				}
				isPattern = true
			}
		}

		var result object.Object
		var err error

		switch arg := over.(type) {
		case *object.List:
			if len(arg.Elements) == 0 {
				return object.NONE
			}

			if isPattern {
				for _, v := range arg.Elements {
					result, err = object.PatternMatchingOrError(re, v)
					if err != nil {
						return object.NewException(object.ERR_GENERAL, fnName, err.Error())
					}
					if result == object.TRUE {
						return object.TRUE
					}
				}

			} else if by != nil {
				for _, v := range arg.Elements {
					result, err = pr.callback(by, v)
					if err != nil {
						return object.NewException(object.ERR_GENERAL, fnName, err.Error())
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

			if isPattern {
				for _, kv := range arg.Pairs {
					result, err = object.PatternMatchingOrError(re, kv.Value)
					if err != nil {
						return object.NewException(object.ERR_GENERAL, fnName, err.Error())
					}
					if result == object.TRUE {
						return object.TRUE
					}
				}

			} else if by != nil {
				for _, kv := range arg.Pairs {
					result, err = pr.callback(by, kv.Value)
					if err != nil {
						return object.NewException(object.ERR_GENERAL, fnName, err.Error())
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
			return object.NewException(object.ERR_ARGUMENTS, fnName, "Expected list or hash")
		}

		return object.FALSE
	},
}
