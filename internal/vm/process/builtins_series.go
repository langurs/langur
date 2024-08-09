// langur/vm/process/builtins_series.go

package process

import (
	"langur/object"
)

// series, pseries

var bi_series = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:        "series",
		Description: "series(range, increment); generates a list of numbers from a range and increment (optional, defaults to 1 or -1)",

		// TODO: update
		ParamPositional: []object.Parameter{
			object.Parameter{},
		},
		ParamExpansionMin: 1,
		ParamExpansionMax: 2,
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		return series(pr, "series", args, false)
	},
}

var bi_pseries = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:        "pseries",
		Description: "like series(), but returns positive series or empty list (given a negative range)",

		// TODO: update
		ParamPositional: []object.Parameter{
			object.Parameter{},
		},
		ParamExpansionMin: 1,
		ParamExpansionMax: 2,
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		// positive series or empty list
		return series(pr, "pseries", args, true)
	},
}

func series(pr *Process, fnName string, args []object.Object, forAscendingSeries bool) object.Object {
	var start, end *object.Number
	var ok bool

	switch arg := args[0].(type) {
	case *object.Range:
		start, ok = arg.Start.(*object.Number)
		if !ok {
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected number for start of range")
		}
		end, ok = arg.End.(*object.Number)
		if !ok {
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected number for end of range")
		}

	case *object.Number:
		// number as implicit range
		end = arg

		if end.Equal(object.Zero) {
			// done
			return &object.List{}

		} else {
			gt, ok := object.GreaterThan(object.One, end)
			if !ok {
				return object.NewError(object.ERR_ARGUMENTS, fnName, "Error checking arguments")
			}
			if gt {
				// negative number
				if forAscendingSeries {
					// done; "positive" series only
					return &object.List{}
				}
				start = end
				end = object.NumberFromInt(-1)

			} else {
				start = object.NumberFromInt(1)
			}
		}

	default:
		return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected range or integer for first argument")
	}

	descending, _ := start.GreaterThan(end)

	if descending && forAscendingSeries {
		// done; ascending series only
		return &object.List{}
	}

	// check increment
	var inc *object.Number

	if len(args) > 1 {
		switch e := args[1].(type) {
		case *object.Number:
			if e.IsZero() {
				// can't use a zero increment, but not an error
				return &object.List{}
			}
			inc = e

		default:
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected number for second argument")
		}
	}

	if inc == nil {
		// no increment specified; default 1 or -1
		if descending {
			inc = object.NumberFromInt(-1)
		} else {
			inc = object.One
		}
	}

	// start and end the same
	if start.Equal(end) {
		return &object.List{Elements: []object.Object{start}}
	}

	if descending == inc.IsPositive() {
		return object.NewError(object.ERR_ARGUMENTS, fnName,
			"Expected ascending range with positive increment, or descending range with negative increment")
	}

	elements := []object.Object{}

	num := start

	if descending {
		for {
			elements = append(elements, num)
			n2 := num.Add(inc)
			if n2 == nil {
				return object.NewError(object.ERR_MATH, fnName, "Error decrementing series")
			}
			num = n2.(*object.Number)
			if gt, _ := end.GreaterThan(num); gt {
				break
			}
		}

	} else {
		for {
			elements = append(elements, num)
			n2 := num.Add(inc)
			if n2 == nil {
				return object.NewError(object.ERR_MATH, fnName, "Error incrementing series")
			}
			num = n2.(*object.Number)
			if gt, _ := num.GreaterThan(end); gt {
				break
			}
		}
	}

	return &object.List{Elements: elements}
}
