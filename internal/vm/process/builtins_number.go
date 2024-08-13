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
		Description: "abs(number); returns the absolute value of a number",

		// TODO: update
		ParamPositional: []object.Parameter{
			object.Parameter{},
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

		// TODO: update
		ParamPositional: []object.Parameter{
			object.Parameter{},
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

		// TODO: update
		ParamPositional: []object.Parameter{
			object.Parameter{},
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

		// TODO: update
		ParamPositional: []object.Parameter{
			object.Parameter{},
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

		// TODO: update
		ParamPositional: []object.Parameter{
			object.Parameter{},
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

		// TODO: update
		ParamPositional: []object.Parameter{
			object.Parameter{},
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

		// TODO: update
		ParamPositional: []object.Parameter{
			object.Parameter{},
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

		// TODO: update
		ParamPositional: []object.Parameter{
			object.Parameter{},
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

		// TODO: update
		ParamPositional: []object.Parameter{
			object.Parameter{},
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

		// TODO: update
		ParamPositional: []object.Parameter{
			object.Parameter{},
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
		Description: "round(number, max, addzeroes, mode); rounds number to specified digits after decimal point; mode from the " + modes.RoundHashName + " hash",

		// TODO: update
		ParamPositional: []object.Parameter{
			object.Parameter{},
		},
		ParamExpansionMin: 1,
		ParamExpansionMax: 4,
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		const fnName = "round"

		// FIXME: update parameters/args
		args = args[0].(*object.List).Elements

		n, ok := args[0].(*object.Number)
		if !ok {
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected number for first argument")
		}

		max := 0
		if len(args) > 1 {
			max, ok = object.NumberToInt(args[1])
			if !ok {
				return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected integer for second argument")
			}
		}

		addTrailingZeroes := true
		if len(args) > 2 {
			trim, ok := args[2].(*object.Boolean)
			if !ok {
				return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected bool for third argument (whether to add trailing zeroes)")
			}
			addTrailingZeroes = trim.Value
		}

		trimTrailingZeroes := false

		var num *object.Number
		var err error
		if len(args) > 3 {
			mode, ok := object.NumberToInt(args[3])
			if !ok {
				return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected integer for fourth argument (from "+modes.RoundHashName+" hash")
			}

			num, err = n.RoundByMode(max, addTrailingZeroes, trimTrailingZeroes, modes.RoundingMode(mode))

		} else {
			// round by current mode
			num, err = n.RoundByMode(max, addTrailingZeroes, trimTrailingZeroes, pr.Modes.Rounding)
		}

		if err != nil {
			return object.NewError(object.ERR_GENERAL, fnName, err.Error())
		}

		return num
	},
}

var bi_trunc = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:        "trunc",
		Description: "trunc(number, max, addzeroes); truncate number to specified digits after decimal point",

		// TODO: update
		ParamPositional: []object.Parameter{
			object.Parameter{},
		},
		ParamExpansionMin: 1,
		ParamExpansionMax: 3,
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		const fnName = "trunc"

		// FIXME: update parameters/args
		args = args[0].(*object.List).Elements

		n, ok := args[0].(*object.Number)
		if !ok {
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected number for first argument")
		}

		max := 0
		if len(args) > 1 {
			max, ok = object.NumberToInt(args[1])
			if !ok {
				return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected integer for second argument")
			}
		}

		addTrailingZeroes := true
		if len(args) > 2 {
			trim, ok := args[2].(*object.Boolean)
			if !ok {
				return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected bool for third argument (whether to add trailing zeroes)")
			}
			addTrailingZeroes = trim.Value
		}

		trimTrailingZeroes := false

		num, err := n.Truncate(max, addTrailingZeroes, trimTrailingZeroes)

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

		// TODO: update
		ParamPositional: []object.Parameter{
			object.Parameter{},
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
