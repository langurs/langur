// langur/vm/process/builtins_number.go

package process

import (
	"fmt"
	"langur/modes"
	"langur/object"
)

// abs
// ceiling, floor
// gcd, lcm
// max, min, minmax
// mean, mid
// round, trunc, simplify

var bi_abs = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:        "abs",
		Description: "returns the absolute value of a number",

		ParamPositional: []object.Parameter{
			object.Parameter{ExternalName: "num"},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		switch num := args[0].(type) {
		case *object.Number:
			return num.Abs()
		}
		return object.NewError(object.ERR_ARGUMENTS, "abs", "Expected a number")
	},
}

var bi_ceiling = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:        "ceiling",
		Description: "returns least integer greater than or equal to input number",

		ParamPositional: []object.Parameter{
			object.Parameter{ExternalName: "num"},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		switch num := args[0].(type) {
		case *object.Number:
			return num.Ceiling()
		}
		return object.NewError(object.ERR_ARGUMENTS, "ceiling", "Expected a number")
	},
}

var bi_floor = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:        "floor",
		Description: "returns greatest integer less than or equal to input number",

		ParamPositional: []object.Parameter{
			object.Parameter{ExternalName: "num"},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		switch num := args[0].(type) {
		case *object.Number:
			return num.Floor()
		}
		return object.NewError(object.ERR_ARGUMENTS, "floor", "Expected a number")
	},
}

var bi_gcd = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:        "gcd",
		Description: "returns the greatest common divisor of 2 or more integers",

		ParamPositional: []object.Parameter{
			object.Parameter{ExternalName: "over"},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		const fnName = "gcd"
		// greatest common divisor

		var elements []object.Object
		var numbers []*object.Number
		var err error

		arr, ok := args[0].(*object.List)
		if !ok {
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected list of numbers")
		}
		elements = arr.Elements

		numbers = make([]*object.Number, len(elements))
		for i, v := range elements {
			a, ok := v.(*object.Number)
			if !ok {
				return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected list of numbers only")
			}
			if !a.IsInteger() {
				// TODO: non-integers with testing
				return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected list of integer numbers only")
			}
			if a.IsZero() {
				return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected list of non-zero numbers only")
			}
			numbers[i] = a.Abs()
		}

		var b *object.Number

		switch len(numbers) {
		case 0:
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected non-empty list")
		case 1:
			return elements[0]
		case 2:
			b, err = object.Gcd(numbers[0], numbers[1])
			if err != nil {
				return object.NewError(object.ERR_GENERAL, fnName, err.Error())
			}
		default:
			// more than 2
			b = numbers[0]
			for i := 1; i < len(numbers); i++ {
				b, err = object.Gcd(b, numbers[i])
				if err != nil {
					return object.NewError(object.ERR_GENERAL, fnName, err.Error())
				}
			}
		}

		return b
	},
}

var bi_lcm = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:        "lcm",
		Description: "returns the least common multiple of 2 or more integers",

		ParamPositional: []object.Parameter{
			object.Parameter{ExternalName: "over"},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		const fnName = "lcm"
		// least common multiple

		var elements []object.Object
		var numbers []*object.Number
		var err error

		arr, ok := args[0].(*object.List)
		if !ok {
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected list of numbers")
		}
		elements = arr.Elements

		numbers = make([]*object.Number, len(elements))
		for i, v := range elements {
			a, ok := v.(*object.Number)
			if !ok {
				return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected list of numbers only")
			}
			if !a.IsInteger() {
				return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected list of integer numbers only")
			}
			if a.IsZero() {
				return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected list of non-zero numbers only")
			}
			numbers[i] = a.Abs()
		}

		var b *object.Number

		switch len(numbers) {
		case 0:
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected non-empty list")
		case 1:
			return elements[0]
		case 2:
			b, err = object.Lcm(numbers[0], numbers[1])
			if err != nil {
				return object.NewError(object.ERR_GENERAL, fnName, err.Error())
			}
		default:
			// more than 2
			b = numbers[0]
			for i := 1; i < len(numbers); i++ {
				b, err = object.Lcm(b, numbers[i])
				if err != nil {
					return object.NewError(object.ERR_GENERAL, fnName, err.Error())
				}
			}
		}

		return b
	},
}

var bi_max = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:        "max",
		Description: "returns maximum from a list, hash, range, string, or function; for string, max. code point; for function, max. parameter count (-1 for no max.)",

		ParamPositional: []object.Parameter{
			object.Parameter{ExternalName: "over"},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		const fnName = "max"

		var elements []object.Object

		switch c := args[0].(type) {
		case *object.List:
			elements = c.Elements

		case *object.Range:
			elements = []object.Object{c.Start, c.End}

		case *object.Hash:
			elements = make([]object.Object, len(c.Pairs))
			i := 0
			for _, kv := range c.Pairs {
				elements[i] = kv.Value
				i++
			}

		case *object.String:
			rSlc := c.RuneSlc()
			elements = make([]object.Object, len(rSlc))
			for i := range rSlc {
				elements[i] = object.NumberFromRune(rSlc[i])
			}

		case *object.CompiledCode, *object.BuiltIn:
			return object.NumberFromInt(object.ParamMax(c))

		default:
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected list, hash, range, string, or function")
		}

		if len(elements) == 0 {
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected non-empty list, hash, range, string, or a function")
		}

		max := elements[0]
		for i := 1; i < len(elements); i++ {
			gt, ok := object.GreaterThan(elements[i], max)
			if !ok {
				return object.NewError(object.ERR_ARGUMENTS, fnName,
					fmt.Sprintf("Could not compare %s with %s", max.TypeString(), elements[i].TypeString()))
			}
			if gt {
				max = elements[i]
			}
		}

		return max
	},
}

var bi_min = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:        "min",
		Description: "returns minimum from a list, hash, range, string, or function; for string, min. code point; for function, min. parameter count",

		ParamPositional: []object.Parameter{
			object.Parameter{ExternalName: "over"},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		const fnName = "min"

		var elements []object.Object

		switch c := args[0].(type) {
		case *object.List:
			elements = c.Elements

		case *object.Range:
			elements = []object.Object{c.Start, c.End}

		case *object.Hash:
			elements = make([]object.Object, len(c.Pairs))
			i := 0
			for _, kv := range c.Pairs {
				elements[i] = kv.Value
				i++
			}

		case *object.String:
			rSlc := c.RuneSlc()
			elements = make([]object.Object, len(rSlc))
			for i := range rSlc {
				elements[i] = object.NumberFromRune(rSlc[i])
			}

		case *object.CompiledCode, *object.BuiltIn:
			return object.NumberFromInt(object.ParamMin(c))

		default:
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected list, hash, range, string, or function")
		}

		if len(elements) == 0 {
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected non-empty list, hash, range, string, or a function")
		}

		min := elements[0]
		for i := 1; i < len(elements); i++ {
			lt, ok := object.GreaterThan(min, elements[i])
			if !ok {
				return object.NewError(object.ERR_ARGUMENTS, fnName,
					fmt.Sprintf("Could not compare %s with %s", min.TypeString(), elements[i].TypeString()))
			}
			if lt {
				min = elements[i]
			}
		}

		return min
	},
}

var bi_minmax = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:        "minmax",
		Description: "returns range of minimum to maximum from a list, hash, range, string, or function; for string, min. code point; for function, min./max. parameter count",

		ParamPositional: []object.Parameter{
			object.Parameter{ExternalName: "over"},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		const fnName = "minmax"

		var elements []object.Object

		switch c := args[0].(type) {
		case *object.List:
			elements = c.Elements

		case *object.Range:
			elements = []object.Object{c.Start, c.End}

		case *object.Hash:
			elements = make([]object.Object, len(c.Pairs))
			i := 0
			for _, kv := range c.Pairs {
				elements[i] = kv.Value
				i++
			}

		case *object.String:
			rSlc := c.RuneSlc()
			elements = make([]object.Object, len(rSlc))
			for i := range rSlc {
				elements[i] = object.NumberFromRune(rSlc[i])
			}

		case *object.CompiledCode, *object.BuiltIn:
			return object.NewRange(object.NumberFromInt(object.ParamMin(c)), object.NumberFromInt(object.ParamMax(c)))

		default:
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected list, hash, range, string, or function")
		}

		if len(elements) == 0 {
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected non-empty list, hash, range, string, or a function")
		}

		min := elements[0]
		max := min
		for i := 1; i < len(elements); i++ {
			lt, ok := object.GreaterThan(min, elements[i])
			if !ok {
				return object.NewError(object.ERR_ARGUMENTS, fnName,
					fmt.Sprintf("Could not compare %s with %s", min.TypeString(), elements[i].TypeString()))
			}
			if lt {
				min = elements[i]
			}
			gt, ok := object.GreaterThan(elements[i], max)
			if !ok {
				return object.NewError(object.ERR_ARGUMENTS, fnName,
					fmt.Sprintf("Could not compare %s with %s", max.TypeString(), elements[i].TypeString()))
			}
			if gt {
				max = elements[i]
			}
		}

		return object.NewRange(min, max)
	},
}

var bi_mean = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:        "mean",
		Description: "returns mean (average) from given set of numbers",

		ParamPositional: []object.Parameter{
			object.Parameter{ExternalName: "over"},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		const fnName = "mean"

		var elements []object.Object

		switch c := args[0].(type) {
		case *object.List:
			elements = c.Elements

		case *object.Range:
			elements = []object.Object{c.Start, c.End}

		case *object.Hash:
			elements = make([]object.Object, len(c.Pairs))
			i := 0
			for _, kv := range c.Pairs {
				elements[i] = kv.Value
				i++
			}

		default:
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected list or hash, or a range")
		}

		if len(elements) == 0 {
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected non-empty list or hash, or a range")
		}

		var nums = make([]*object.Number, len(elements))

		for i := range elements {
			n, ok := elements[i].(*object.Number)
			if !ok {
				return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected numbers only")
			}
			nums[i] = n
		}

		return object.Mean(nums...)
	},
}

var bi_mid = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:        "mid",
		Description: "returns mid-point from given set of numbers",

		ParamPositional: []object.Parameter{
			object.Parameter{ExternalName: "over"},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		const fnName = "mid"

		var elements []object.Object

		switch c := args[0].(type) {
		case *object.List:
			elements = c.Elements

		case *object.Range:
			elements = []object.Object{c.Start, c.End}

		case *object.Hash:
			elements = make([]object.Object, len(c.Pairs))
			i := 0
			for _, kv := range c.Pairs {
				elements[i] = kv.Value
				i++
			}

		default:
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected list or hash, or a range")
		}

		if len(elements) == 0 {
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected non-empty list or hash, or a range")
		}

		var nums = make([]*object.Number, len(elements))

		for i := range elements {
			n, ok := elements[i].(*object.Number)
			if !ok {
				return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected numbers only")
			}
			nums[i] = n
		}

		return object.Mid(nums...)
	},
}

var bi_round = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:        "round",
		Description: "rounds number to specified digits after decimal point; mode from the " + modes.RoundHashName + " hash",

		ParamPositional: []object.Parameter{
			object.Parameter{ExternalName: "num"},
		},

		ParamByName: []object.Parameter{
			object.Parameter{ExternalName: "places", DefaultValue: object.Zero},
			object.Parameter{ExternalName: "addzeroes", DefaultValue: object.TRUE},
			object.Parameter{ExternalName: "mode"},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		const fnName = "round"

		n, ok := args[0].(*object.Number)
		if !ok {
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected number")
		}

		places, ok := object.NumberToInt(args[1])
		if !ok {
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected integer for argument places")
		}

		trim, ok := args[2].(*object.Boolean)
		if !ok {
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected bool for argument addzeroes")
		}
		addTrailingZeroes := trim.Value
		trimTrailingZeroes := false

		var mode modes.RoundingMode
		if args[3] == nil {
			// round by current mode
			mode = pr.Modes.Rounding

		} else {
			m, ok := object.NumberToInt(args[3])
			if !ok {
				return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected integer for fourth argument (from "+modes.RoundHashName+" hash")
			}
			mode = modes.RoundingMode(m)
		}

		num, err := n.RoundByMode(places, addTrailingZeroes, trimTrailingZeroes, mode)
		if err != nil {
			return object.NewError(object.ERR_GENERAL, fnName, err.Error())
		}

		return num
	},
}

var bi_trunc = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:        "trunc",
		Description: "truncate number to specified digits after decimal point",

		ParamPositional: []object.Parameter{
			object.Parameter{ExternalName: "num"},
		},

		ParamByName: []object.Parameter{
			object.Parameter{ExternalName: "places", DefaultValue: object.Zero},
			object.Parameter{ExternalName: "addzeroes", DefaultValue: object.TRUE},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		const fnName = "trunc"

		n, ok := args[0].(*object.Number)
		if !ok {
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected number for first argument")
		}

		places, ok := object.NumberToInt(args[1])
		if !ok {
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected integer for argument places")
		}

		trim, ok := args[2].(*object.Boolean)
		if !ok {
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected bool for argument addzeroes")
		}
		addTrailingZeroes := trim.Value
		trimTrailingZeroes := false

		num, err := n.Truncate(places, addTrailingZeroes, trimTrailingZeroes)

		if err != nil {
			return object.NewError(object.ERR_GENERAL, fnName, err.Error())
		}

		return num
	},
}

var bi_simplify = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:        "simplify",
		Description: "simplifies number, removing trailing zeros",

		ParamPositional: []object.Parameter{
			object.Parameter{ExternalName: "num"},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		n, ok := args[0].(*object.Number)
		if !ok {
			return object.NewError(object.ERR_ARGUMENTS, "simplify", "Expected number")
		}
		return n.Simplify()
	},
}
