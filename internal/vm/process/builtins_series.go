// langur/vm/process/builtins_series.go

package process

import (
	"langur/object"
)

// series

var bi_series = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:        "series",
		Description: "generates a list of numbers from a range and increment (optional, defaults to 1 or -1)",

		ParamPositional: []object.Parameter{
			object.Parameter{ExternalName: "from"},
		},
		ParamByName: []object.Parameter{
			object.Parameter{ExternalName: "inc"},
			object.Parameter{ExternalName: "asconly", DefaultValue: object.FALSE},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		const fnName = "series"

		var err error
		var result object.Object

		from, increment := args[0], args[1]

		ascendingOnly, ok := args[2].(*object.Boolean)
		if !ok {
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected bool for argument asconly")
		}
		forAscendingSeries := ascendingOnly.Value
		correctIncrementSign := false

		// check increment
		var inc *object.Number

		switch e := increment.(type) {
		case nil:
			// no increment specified; default 1
			inc = object.One
			correctIncrementSign = true

		case *object.Number:
			if e.IsZero() {
				// can't use a zero increment, but not an error
				return &object.List{}
			}
			inc = e

		default:
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected number for argument inc")
		}

		// check source
		switch arg := from.(type) {
		case *object.Range:
			if forAscendingSeries && !arg.IsFlatOrAscending() {
				return &object.List{}
			}
			result, err = arg.ToList(inc, correctIncrementSign)
			
		case *object.Number:
			// number as implicit range
			result, err = arg.ToList(inc)
			
		default:
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected range or number for argument from")
		}

		if err != nil {
			return object.NewError(object.ERR_GENERAL, fnName, err.Error())
		}

		return result
	},
}
