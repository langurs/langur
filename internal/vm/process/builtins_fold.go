// builtins_fold.go

package process

import (
	"fmt"
	"langur/object"
)

// fold, foldfrom, zip

var bi_fold = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:        "fold",
		Description: "fold(function, list); returns list of values folded by the given function from the given list",

		// TODO: update
		ParamPositional: []object.Parameter{
			object.Parameter{},
			object.Parameter{},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		const fnName = "fold"

		var fns []object.Object

		fn := args[0]
		if !object.IsCallable(fn) {
			fnArr, _ := fn.(*object.List)
			if len(fnArr.Elements) == 0 {
				return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected function or list of functions for first argument")
			}
			fns = fnArr.Elements
			for i, f := range fns {
				if !object.IsCallable(f) {
					return object.NewError(object.ERR_ARGUMENTS, fnName, fmt.Sprintf("List element %d not callable", i+1))
				}
			}
			fn = fns[0] // initialiaze fn to first function
		}

		var arr *object.List

		switch arg := args[1].(type) {
		case *object.List:
			arr = arg

		case *object.Range:
			from, err := arg.ToList()
			if err != nil {
				return object.NewError(object.ERR_ARGUMENTS, fnName, err.Error())
			}
			arr = from

		default:
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected list (or range)")
		}
		if len(arr.Elements) == 0 {
			// empty list
			return object.NONE
		}

		var err error
		fnn := 0

		result := arr.Elements[0]
		for i := 1; i < len(arr.Elements); i++ {
			result, err = pr.callback(fn, result, arr.Elements[i])
			if err != nil {
				return object.NewError(object.ERR_GENERAL, fnName, err.Error())
			}

			if fns != nil {
				// next function
				if fnn > len(fns)-2 {
					fnn = 0
				} else {
					fnn++
				}
				fn = fns[fnn]
			}
		}

		return result
	},
}

var bi_foldfrom = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:        "foldfrom",
		Description: "foldfrom(function, init, lists...); returns list of values folded by the given function from the given lists; given function parameter count == number of lists + 1 for the result (result as first parameter in given function)",

		// TODO: update
		ParamPositional: []object.Parameter{
			object.Parameter{},
		},
		ParamExpansionMin: 3,
		ParamExpansionMax: -1,
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		const fnName = "foldfrom"

		// FIXME: update parameters/args
		args = args[0].(*object.List).Elements

		if len(args) == 3 {
			// starting result and single list
			var arr *object.List

			switch arg := args[2].(type) {
			case *object.List:
				arr = arg

			case *object.Range:
				from, err := arg.ToList()
				if err != nil {
					return object.NewError(object.ERR_ARGUMENTS, fnName, err.Error())
				}
				arr = from

			default:
				return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected list (or range) for third argument")
			}

			arr.Elements = append([]object.Object{args[1]}, arr.Elements...)
			return bi_fold.Fn.(BuiltInFunction)(pr, args[0], arr)
		}
		return foldMultiple(pr, args...)
	},
}

// an extension of bi_foldfrom() for folding on multiple lists
// may be slower than using a single list
func foldMultiple(pr *Process, args ...object.Object) object.Object {
	const fnName = "foldfrom"

	var fns []object.Object

	fn := args[0]
	if !object.IsCallable(fn) {
		fnArr, _ := fn.(*object.List)
		if len(fnArr.Elements) == 0 {
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected function or list of functions for first argument")
		}
		fns = fnArr.Elements
		for i, f := range fns {
			if !object.IsCallable(f) {
				return object.NewError(object.ERR_ARGUMENTS, fnName, fmt.Sprintf("List element %d not callable", i+1))
			}
		}
		fn = fns[0] // initialiaze fn to first function
	}

	result := args[1]

	var length int = -1
	var lists []object.Object

	for _, o := range args[2:] {
		switch arg := o.(type) {
		case *object.List:
			if length > -1 && length != len(arg.Elements) {
				return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected same size for multiple lists (or ranges) to fold from")
			}
			length = len(arg.Elements)
			lists = append(lists, arg)

		case *object.Range:
			from, err := arg.ToList()
			if err != nil {
				return object.NewError(object.ERR_ARGUMENTS, fnName, err.Error())
			}
			if length > -1 && length != len(from.Elements) {
				return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected same size for multiple lists (or ranges) to fold from")
			}
			length = len(from.Elements)
			lists = append(lists, from)

		default:
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected lists (or ranges) only for foldfrom")
		}
	}
	if length == 0 {
		// empty lists
		// return starting result or "no data" ??????
		return object.NONE
	}

	gather := func(idx int) []object.Object {
		elements := []object.Object{}
		for _, arrObj := range lists {
			elements = append(elements, arrObj.(*object.List).Elements[idx])
		}
		return elements
	}

	var err error
	fnn := 0

	for i := 0; i < length; i++ {
		result, err = pr.callback(fn, append([]object.Object{result}, gather(i)...)...)
		if err != nil {
			return object.NewError(object.ERR_GENERAL, fnName, err.Error())
		}

		if fns != nil {
			// next function
			if fnn > len(fns)-2 {
				fnn = 0
			} else {
				fnn++
			}
			fn = fns[fnn]
		}
	}

	return result
}

var bi_zip = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:        "zip",
		Description: "zips together lists; may optionally use a function (first argument)",

		// TODO: update
		ParamPositional: []object.Parameter{
			object.Parameter{},
		},
		ParamExpansionMin: 2,
		ParamExpansionMax: -1,
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		const fnName = "zip"

		// FIXME: update parameters/args
		args = args[0].(*object.List).Elements

		fn := args[0]
		useFn := object.IsCallable(fn)

		start := 0
		if useFn {
			start = 1
		}

		length := -1

		var lists []object.Object

		for _, o := range args[start:] {
			switch arg := o.(type) {
			case *object.List:
				if length > -1 && length != len(arg.Elements) {
					return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected same size for multiple lists to zip")
				}
				length = len(arg.Elements)
				lists = append(lists, arg)

			case *object.Range:
				from, err := arg.ToList()
				if err != nil {
					return object.NewError(object.ERR_ARGUMENTS, fnName, err.Error())
				}
				if length > -1 && length != len(from.Elements) {
					return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected same size for multiple lists (or ranges) to zip from")
				}
				length = len(from.Elements)
				lists = append(lists, from)

			default:
				return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected lists only to zip")
			}
		}

		arr := &object.List{}
		var items []object.Object
		if useFn {
			// use function to determine zip values
			for i := 0; i < length; i++ {
				items = nil
				for _, arr := range lists {
					items = append(items, arr.(*object.List).Elements[i])
				}
				result, err := pr.callback(fn, items...)
				if err != nil {
					return object.NewError(object.ERR_GENERAL, fnName, err.Error())
				}
				switch r := result.(type) {
				case *object.List:
					arr.Elements = append(arr.Elements, r.Elements...)
				default:
					arr.Elements = append(arr.Elements, result)
				}
			}

		} else {
			// no function; straight-up zip
			for i := 0; i < length; i++ {
				items = nil
				for _, arr2 := range lists {
					items = append(items, arr2.(*object.List).Elements[i])
				}
				arr.Elements = append(arr.Elements, items...)
			}
		}

		return arr
	},
}
