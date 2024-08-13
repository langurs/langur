// langur/vm/process/builtins_map.go

package process

import (
	"fmt"
	"langur/object"
)

// map, mapX

var bi_map = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:        "map",
		Description: "map(function, lists...); returns list (or hash) of values mapped to the given function from the given lists or hashes (one type only)",

		// TODO: update
		ParamPositional: []object.Parameter{
			object.Parameter{},
		},
		ParamExpansionMin: 2,
		ParamExpansionMax: -1,
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		const fnName = "map"

		// FIXME: update parameters/args
		args = args[0].(*object.List).Elements

		var fns []object.Object

		fn := args[0]
		if !object.IsCallable(fn) {
			arr, ok := fn.(*object.List)
			if !ok || len(arr.Elements) == 0 {
				return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected function or list of functions for first argument")
			}
			fns = arr.Elements
			for i, f := range fns {
				if f == object.NONE {
					fns[i] = nil
				} else if !object.IsCallable(f) {
					return object.NewError(object.ERR_ARGUMENTS, fnName, fmt.Sprintf("List element %d not callable or no-op", i+1))
				}
			}
			fn = fns[0] // initialiaze fn to first function
		}

		if len(args) > 2 {
			return mapMultiple(pr, fn, fns, args[1:])
		}

		fnn := 0 // current function number

		nextFunction := func() {
			if fns != nil {
				if fnn > len(fns)-2 {
					fnn = 0
				} else {
					fnn++
				}
				fn = fns[fnn]
			}
		}

		mapToList := func(elements []object.Object) object.Object {
			arr := &object.List{}

			for _, v := range elements {
				if fn == nil {
					// no op
					arr.Elements = append(arr.Elements, v)

				} else {
					result, err := pr.callback(fn, v)
					if err != nil {
						return object.NewError(object.ERR_GENERAL, fnName, err.Error())
					}
					arr.Elements = append(arr.Elements, result)
				}
				nextFunction()
			}
			return arr
		}

		switch arg := args[1].(type) {
		case *object.List:
			return mapToList(arg.Elements)

		case *object.Range:
			from, err := arg.ToList()
			if err != nil {
				return object.NewError(object.ERR_ARGUMENTS, fnName, err.Error())
			}
			return mapToList(from.Elements)

		case *object.Hash:
			elements := make([]object.Object, 0, len(arg.Pairs)*2)

			for _, kv := range arg.Pairs {
				if fn == nil {
					// no op
					elements = append(elements, kv.Key, kv.Value)

				} else {
					result, err := pr.callback(fn, kv.Value)
					if err != nil {
						return object.NewError(object.ERR_GENERAL, fnName, err.Error())
					}
					elements = append(elements, kv.Key, result)
				}
				nextFunction()
			}
			hash, err := object.NewHashFromSlice(elements, false)
			if err != nil {
				return object.NewError(object.ERR_GENERAL, fnName, err.Error())
			}
			return hash
		}

		return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected lists (or ranges) or hashes after first argument")
	},
}

// an extension of bi_map() for mapping multiple lists or hashes
// may be slower than using a single list or hash
func mapMultiple(
	pr *Process,
	fn object.Object, fns, args []object.Object) object.Object {

	const fnName = "map"

	var lists, hashes []object.Object
	var length int = -1

	for _, o := range args {
		switch o.(type) {
		case *object.List:
			if hashes != nil {
				return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected same type for multiple things to map (lists (or ranges) or hashes)")
			}
			if length > -1 && length != len(o.(*object.List).Elements) {
				return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected same size for multiple lists to map")
			}
			length = len(o.(*object.List).Elements)
			lists = append(lists, o)

		case *object.Range:
			// ranges converted to lists
			if hashes != nil {
				return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected same type for multiple things to map (lists (or ranges) or hashes)")
			}
			arr, err := o.(*object.Range).ToList()
			if err != nil {
				return object.NewError(object.ERR_ARGUMENTS, fnName, err.Error())
			}
			if length > -1 && length != len(arr.Elements) {
				return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected same size for multiple lists (or ranges) to map")
			}
			length = len(arr.Elements)
			lists = append(lists, arr)

		case *object.Hash:
			if lists != nil {
				return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected same type for multiple things to map (lists (or ranges) or hashes)")
			}
			hashes = append(hashes, o)

		default:
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected lists (or ranges) or hashes only to map")
		}
	}

	var items []object.Object
	fnn := 0 // function number

	nextFunction := func() {
		if fns != nil {
			if fnn > len(fns)-2 {
				fnn = 0
			} else {
				fnn++
			}
			fn = fns[fnn]
		}
	}

	if lists != nil {
		arr := &object.List{}
		for i := 0; i < length; i++ {
			items = nil
			for _, arr := range lists {
				items = append(items, arr.(*object.List).Elements[i])
			}
			if fn == nil {
				// no op
				arr.Elements = append(arr.Elements, &object.List{Elements: items})

			} else {
				result, err := pr.callback(fn, items...)
				if err != nil {
					return object.NewError(object.ERR_GENERAL, fnName, err.Error())
				}
				arr.Elements = append(arr.Elements, result)
			}
			nextFunction()
		}
		return arr
	}

	if hashes != nil {
		// Mapping to multiple hashes depends upon the keys of the first hash passed.
		elements := make([]object.Object, 0, len(hashes[0].(*object.Hash).Pairs))
		for _, kv := range hashes[0].(*object.Hash).Pairs {
			items = nil
			for _, h := range hashes {
				val, err := h.(*object.Hash).GetValue(kv.Key)
				if err != nil {
					return object.NewError(object.ERR_GENERAL, fnName, err.Error())
				}
				items = append(items, val)
			}
			if fn == nil {
				// no op
				elements = append(elements, kv.Key, &object.List{Elements: items})

			} else {
				result, err := pr.callback(fn, items...)
				if err != nil {
					return object.NewError(object.ERR_GENERAL, fnName, err.Error())
				}
				elements = append(elements, kv.Key, result)
			}
			nextFunction()
		}
		hash, err := object.NewHashFromSlice(elements, false)
		if err != nil {
			return object.NewError(object.ERR_GENERAL, fnName, err.Error())
		}
		return hash
	}

	return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected lists (or ranges) or hashes")
}

var bi_mapX = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:        "mapX",
		Description: "mapX(function, lists...); returns list of values mapped to the given function from the given lists",

		// TODO: update
		ParamPositional: []object.Parameter{
			object.Parameter{},
		},
		ParamExpansionMin: 2,
		ParamExpansionMax: -1,
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		const fnName = "mapX"

		// FIXME: update parameters/args
		args = args[0].(*object.List).Elements

		var fn object.Object
		var arrs []object.Object

		if object.IsCallable(args[0]) {
			fn = args[0]
			arrs = args[1:]

		} else {
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected function for first argument")

			// fn = nil
			// arrs = args

			// arr, ok := args[0].(*object.List)
			// if ok {
			// 	// not presently allowing no-ops or functions in first list to mapX()
			// 	for i, f := range arr.Elements {
			// 		if object.IsCallable(f) || f == object.NONE {
			// 			return object.NewError(object.ERR_ARGUMENTS, fnName, fmt.Sprintf("List element %d callable or no-op", i+1))
			// 		}
			// 	}
			// }
		}

		return crossMap(pr, fnName, fn, arrs...)
	},
}

func crossMap(
	pr *Process,
	fnName string,
	fn object.Object,
	arrs ...object.Object) object.Object {

	var lists []*object.List
	arr := &object.List{}

	for _, o := range arrs {
		switch o := o.(type) {
		case *object.List:
			if len(o.Elements) == 0 {
				// with any zero length lists, return an empty list
				return arr
			}
			lists = append(lists, o)

		case *object.Range:
			// ranges converted to lists
			arr, err := o.ToList()
			if err != nil {
				return object.NewError(object.ERR_ARGUMENTS, fnName, err.Error())
			}
			lists = append(lists, arr)

		default:
			lists = append(lists, &object.List{Elements: []object.Object{o}})
		}
	}

	var counters = make([]int, len(lists))

	done := false
	for !done {
		items := make([]object.Object, 0, len(counters))
		for i := 0; i < len(counters); i++ {
			items = append(items, lists[i].Elements[counters[i]])
		}

		if fn == nil {
			arr.Elements = append(arr.Elements, &object.List{Elements: items})
		} else {
			result, err := pr.callback(fn, items...)
			if err != nil {
				return object.NewError(object.ERR_GENERAL, fnName, err.Error())
			}
			arr.Elements = append(arr.Elements, result)
		}

		done = true
		for i := len(counters) - 1; i >= 0; i-- {
			counters[i]++
			if counters[i] > len(lists[i].Elements)-1 {
				counters[i] = 0
				continue
			}
			done = false
			break
		}
	}

	return arr
}
