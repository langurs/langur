// builtins_fold.go

package process

import (
	"fmt"
	"langur/object"
)

// fold,  zip

var bi_fold = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:        "fold",
		Description: "returns list of values folded by the given function from the given list",

		ParamPositional: []object.Parameter{
			object.Parameter{ExternalName: "over"},
		},
		ParamExpansionMin: 1,
		ParamExpansionMax: -1,

		ParamByName: []object.Parameter{
			object.Parameter{ExternalName: "by", Required: true},
			object.Parameter{ExternalName: "init"},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		const fnName = "fold"

		var fns []object.Object
		fn := args[1]
		if !object.IsCallable(fn) {
			fnArr, _ := fn.(*object.List)
			if len(fnArr.Elements) == 0 {
				return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected function or list of functions")
			}
			fns = fnArr.Elements
			for i, f := range fns {
				if !object.IsCallable(f) {
					return object.NewError(object.ERR_ARGUMENTS, fnName, fmt.Sprintf("Function list element %d not callable", i+1))
				}
			}
			fn = fns[0] // initialiaze fn to first function
		}

		result := args[2] // nil if init not passed; checked later

		lists := args[0].(*object.List).Elements // We know it's a list, because parameter expansion made it so.
		var list *object.List
		if len(lists) == 1 {
			switch arg := lists[0].(type) {
			case *object.List:
				list = arg

			case *object.Range:
				from, err := arg.ToList(object.One)
				if err != nil {
					return object.NewError(object.ERR_ARGUMENTS, fnName, err.Error())
				}
				list = from

			default:
				return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected list or range")
			}
			if len(list.Elements) == 0 {
				// empty list
				// FIXME: ? return initialization if not nil ?
				return object.NONE
			}

		} else {
			// more than 1 list
			return foldBetweenLists(pr, lists, fns, fn, result)
		}

		start := 0
		if result == nil {
			result = list.Elements[0]
			start = 1
		}

		var err error
		fnn := 0
		for i := start; i < len(list.Elements); i++ {
			result, err = pr.callback(fn, result, list.Elements[i])
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

// an extension of fold() for folding between lists
func foldBetweenLists(
	pr *Process,
	lists, fns []object.Object,
	fn, result object.Object) object.Object {

	const fnName = "fold"

	if result == nil {
		return object.NewError(object.ERR_ARGUMENTS, fnName, "Cannot fold between lists without initialization")
	}

	var length int = -1

	for i, o := range lists {
		switch arg := o.(type) {
		case *object.List:
			if length > -1 && length != len(arg.Elements) {
				return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected same size for multiple lists (or ranges) to fold from")
			}
			length = len(arg.Elements)

		case *object.Range:
			from, err := arg.ToList(object.One)
			if err != nil {
				return object.NewError(object.ERR_ARGUMENTS, fnName, err.Error())
			}
			if length > -1 && length != len(from.Elements) {
				return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected same size for multiple lists (or ranges) to fold from")
			}
			length = len(from.Elements)
			lists[i] = from

		default:
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected lists (or ranges) only for foldfrom")
		}
	}
	if length == 0 {
		// empty lists
		// FIXME: ? return starting result or "no data" ?
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
		Description: "zips together lists; may optionally use a function",

		ParamPositional: []object.Parameter{
			object.Parameter{ExternalName: "over"},
		},
		ParamExpansionMin: 2,
		ParamExpansionMax: -1,

		ParamByName: []object.Parameter{
			object.Parameter{ExternalName: "by"},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		const fnName = "zip"

		fn := args[1]
		if fn != nil {
			if !object.IsCallable(fn) {
				return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected function")
			}
		}

		length := -1

		lists := args[0].(*object.List).Elements
		for i, o := range lists {
			switch arg := o.(type) {
			case *object.List:
				if length > -1 && length != len(arg.Elements) {
					return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected same size for multiple lists (or ranges) to zip from")
				}
				length = len(arg.Elements)

			case *object.Range:
				from, err := arg.ToList(object.One)
				if err != nil {
					return object.NewError(object.ERR_ARGUMENTS, fnName, err.Error())
				}
				if length > -1 && length != len(from.Elements) {
					return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected same size for multiple lists (or ranges) to zip from")
				}
				length = len(from.Elements)
				lists[i] = from

			default:
				return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected lists only to zip")
			}
		}

		list := &object.List{}
		var items []object.Object
		if fn != nil {
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
					list.Elements = append(list.Elements, r.Elements...)
				default:
					list.Elements = append(list.Elements, result)
				}
			}

		} else {
			// no function; straight-up zip
			for i := 0; i < length; i++ {
				items = nil
				for _, arr2 := range lists {
					items = append(items, arr2.(*object.List).Elements[i])
				}
				list.Elements = append(list.Elements, items...)
			}
		}

		return list
	},
}
