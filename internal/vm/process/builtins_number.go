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

func bi_abs(pr *Process, args ...object.Object) object.Object {
	switch num := args[0].(type) {
	case *object.Number:
		return num.Abs()
	}
	return object.NewError(object.ERR_ARGUMENTS, "abs", "Expected a number")
}

func bi_ceiling(pr *Process, args ...object.Object) object.Object {
	switch num := args[0].(type) {
	case *object.Number:
		return num.Ceiling()
	}
	return object.NewError(object.ERR_ARGUMENTS, "ceiling", "Expected a number")
}

func bi_floor(pr *Process, args ...object.Object) object.Object {
	switch num := args[0].(type) {
	case *object.Number:
		return num.Floor()
	}
	return object.NewError(object.ERR_ARGUMENTS, "floor", "Expected a number")
}

func bi_gcd(pr *Process, args ...object.Object) object.Object {
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
}

func bi_lcm(pr *Process, args ...object.Object) object.Object {
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
}

func bi_max(pr *Process, args ...object.Object) object.Object {
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
}

func bi_min(pr *Process, args ...object.Object) object.Object {
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
}

func bi_minmax(pr *Process, args ...object.Object) object.Object {
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
}

func bi_mean(pr *Process, args ...object.Object) object.Object {
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
}

func bi_mid(pr *Process, args ...object.Object) object.Object {
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
}

func bi_round(pr *Process, args ...object.Object) object.Object {
	const fnName = "round"

	n, ok := args[0].(*object.Number)
	if !ok {
		return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected number for first argument")
	}

	max := 0
	if len(args) > 1 {
		var err error
		switch e := args[1].(type) {
		case *object.Number:
			max, err = e.ToInt()
			if err != nil {
				return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected integer for second argument")
			}
		default:
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected integer for second argument")
		}
	}

	trimTrailingZeroes := false
	if len(args) > 2 {
		trim, ok := args[2].(*object.Boolean)
		if !ok {
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected bool for third argument (whether to trim trailing zeroes)")
		}
		trimTrailingZeroes = trim.Value
	}

	var num *object.Number
	var err error
	if len(args) > 3 {
		mode, ok := object.NumberToInt(args[3])
		if !ok {
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected integer for fourth argument (from "+modes.RoundHashName+" hash")
		}
		num, err = n.RoundBy(max, mode)

	} else {
		// round by current mode
		num, err = n.Round(max, trimTrailingZeroes)
	}

	if err != nil {
		return object.NewError(object.ERR_GENERAL, fnName, err.Error())
	}

	return num
}

func bi_trunc(pr *Process, args ...object.Object) object.Object {
	const fnName = "trunc"

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

	num, err := n.Truncate(max)

	if err != nil {
		return object.NewError(object.ERR_GENERAL, fnName, err.Error())
	}

	return num
}

func bi_simplify(pr *Process, args ...object.Object) object.Object {
	n, ok := args[0].(*object.Number)
	if !ok {
		return object.NewError(object.ERR_ARGUMENTS, "simplify", "Expected number")
	}
	return n.Simplify()
}
