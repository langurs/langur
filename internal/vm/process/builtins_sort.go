// langur/vm/process/builtins_sort.go

package process

import (
	"langur/object"
	"langur/opcode"
)

// sort

var bi_sort = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:        "sort",
		Description: "sort(function, list); returns a sorted list from the given list, comparing by the given function (taking two variables and returning a Boolean in the form of f(.a, .b) .a < .b, or with implied parameters in the form of f .a < .b)",

		// TODO: update
		ParamPositional: []object.Parameter{
			object.Parameter{},
		},
		ParamExpansionMin: 1,
		ParamExpansionMax: 2,
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		const fnName = "sort"

		// FIXME: update parameters/args
		args = args[0].(*object.List).Elements

		var fn, over object.Object
		var pmax int

		if len(args) == 1 {
			pmax = 0
			over = args[0]
		} else {
			fn = args[0]
			if !object.IsCallable(fn) {
				return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected callable for first argument when passed 2 arguments")
			}
			pmax = object.ParamMax(fn)
			if pmax == -1 {
				// if a function that takes an "unlimited" number of parameters, pass 2
				pmax = 2
			} else if pmax < 1 || pmax > 2 {
				return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected callable that may be passed 1 or 2 arguments")
			}
			over = args[1]
		}

		arr, isList := over.(*object.List)
		if isList {
			var sorted []object.Object
			var err error

			if pmax == 0 {
				sorted, err = quickSort(arr.Elements)
				if err != nil {
					return object.NewError(object.ERR_GENERAL, fnName, err.Error())
				}

			} else if pmax == 1 {
				sorted, err = quickSortFromSingleParameterFunction(pr, fn, arr.Elements)
				if err != nil {
					return object.NewError(object.ERR_GENERAL, fnName, err.Error())
				}

			} else {
				sorted, err = quickSortFromTwoParameterFunction(pr, fn, arr.Elements)
				if err != nil {
					return object.NewError(object.ERR_GENERAL, fnName, err.Error())
				}
			}
			return &object.List{Elements: sorted}

		} else {
			rng, isRange := over.(*object.Range)
			if isRange {
				var less object.Object
				var err error

				if pmax == 0 {
					less, err = object.InfixComparison(opcode.OpLessThan, rng.Start, rng.End, 0)
					if err != nil {
						return object.NewError(object.ERR_GENERAL, fnName, err.Error())
					}

				} else if pmax == 1 {
					first, err := pr.callback(fn, rng.Start)
					if err != nil {
						return object.NewError(object.ERR_GENERAL, fnName, err.Error())
					}
					second, err := pr.callback(fn, rng.End)
					if err != nil {
						return object.NewError(object.ERR_GENERAL, fnName, err.Error())
					}

					less, err = object.InfixComparison(opcode.OpLessThan, first, second, 0)
					if err != nil {
						return object.NewError(object.ERR_GENERAL, fnName, err.Error())
					}

				} else {
					less, err = pr.callback(fn, rng.Start, rng.End)
					if err != nil {
						return object.NewError(object.ERR_GENERAL, fnName, err.Error())
					}
				}

				if less == object.TRUE {
					return &object.Range{Start: rng.Start, End: rng.End}
				} else {
					return &object.Range{Start: rng.End, End: rng.Start}
				}
			} else {
				_, isNumber := over.(*object.Number)
				if isNumber {
					return over
				}
			}
		}

		return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected list or range")
	},
}

func quickSortFromTwoParameterFunction(
	pr *Process, fn object.Object, elements []object.Object) (
	[]object.Object, error) {

	if len(elements) < 2 {
		return elements, nil
	}
	left := make([]object.Object, 0, len(elements)-1)
	right := make([]object.Object, 0, len(elements)-1)
	pivot := elements[len(elements)-1]

	for _, v := range elements[:len(elements)-1] {
		less, err := pr.callback(fn, v, pivot)
		if err != nil {
			return nil, err
		}
		if less == object.TRUE {
			left = append(left, v)
		} else {
			right = append(right, v)
		}
	}

	var err error
	left, err = quickSortFromTwoParameterFunction(pr, fn, left)
	if err != nil {
		return nil, err
	}
	right, err = quickSortFromTwoParameterFunction(pr, fn, right)
	if err != nil {
		return nil, err
	}
	return append(append(left, pivot), right...), nil
}

func quickSortFromSingleParameterFunction(
	pr *Process, fn object.Object, elements []object.Object) (
	[]object.Object, error) {

	if len(elements) < 2 {
		return elements, nil
	}
	left := make([]object.Object, 0, len(elements)-1)
	right := make([]object.Object, 0, len(elements)-1)
	pivot := elements[len(elements)-1]

	for _, v := range elements[:len(elements)-1] {
		first, err := pr.callback(fn, v)
		if err != nil {
			return nil, err
		}
		second, err := pr.callback(fn, pivot)
		if err != nil {
			return nil, err
		}

		less, err := object.InfixComparison(opcode.OpLessThan, first, second, 0)
		if err != nil {
			return nil, err
		}

		if less == object.TRUE {
			left = append(left, v)
		} else {
			right = append(right, v)
		}
	}

	var err error
	left, err = quickSortFromSingleParameterFunction(pr, fn, left)
	if err != nil {
		return nil, err
	}
	right, err = quickSortFromSingleParameterFunction(pr, fn, right)
	if err != nil {
		return nil, err
	}
	return append(append(left, pivot), right...), nil
}

func quickSort(elements []object.Object) (
	[]object.Object, error) {

	if len(elements) < 2 {
		return elements, nil
	}
	left := make([]object.Object, 0, len(elements)-1)
	right := make([]object.Object, 0, len(elements)-1)
	pivot := elements[len(elements)-1]

	for _, v := range elements[:len(elements)-1] {
		less, err := object.InfixComparison(opcode.OpLessThan, v, pivot, 0)
		if err != nil {
			return nil, err
		}

		if less == object.TRUE {
			left = append(left, v)
		} else {
			right = append(right, v)
		}
	}

	var err error
	left, err = quickSort(left)
	if err != nil {
		return nil, err
	}
	right, err = quickSort(right)
	if err != nil {
		return nil, err
	}
	return append(append(left, pivot), right...), nil
}
