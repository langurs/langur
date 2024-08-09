// langur/vm/process/builtins_group.go

package process

import (
	"langur/object"
)

// group, groupby, groupbyH

var bi_group = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:        "group",
		Description: "group(by, list); groups list elements into list of lists as specified by first argument",

		// TODO: update
		ParamPositional: []object.Parameter{
			object.Parameter{},
		},
		ParamExpansionMin: 1,
		ParamExpansionMax: 2,
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		const fnName = "group"

		var integer int
		var fn, over object.Object
		var ok bool
		useTruthiness := false

		if len(args) == 2 {
			switch arg1 := args[0].(type) {
			case *object.Number:
				integer, ok = object.NumberToInt(arg1)
				if !ok || integer == 0 {
					return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected function or non-zero integer for first argument")
				}
			case *object.CompiledCode, *object.BuiltIn:
				fn = arg1
			default:
				return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected function or non-zero integer for first argument")
			}
			over = args[1]

		} else {
			over = args[0]
			useTruthiness = true
		}

		switch arg2 := over.(type) {
		case *object.List:
			if fn != nil || useTruthiness {
				return groupByFunctionOrTruthiness(fnName, pr, fn, false, arg2)
			}

			// group by integer
			groupArr := &object.List{}
			start := 0

			if integer < 0 {
				// negative integer; "start from right" by changing where we start from the left
				integer = -integer
				start = len(arg2.Elements) % integer
				if start != 0 {
					// not evenly divided
					groupArr.Elements = append(groupArr.Elements,
						&object.List{Elements: object.CopySlice(arg2.Elements[:start])})
				}
			}

			for i := start; i < len(arg2.Elements); i += integer {
				end := i + integer
				if end > len(arg2.Elements) {
					end = len(arg2.Elements)
				}
				groupArr.Elements = append(groupArr.Elements,
					&object.List{Elements: object.CopySlice(arg2.Elements[i:end])})
			}
			return groupArr

		default:
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected list or string for second argument")
		}
	},
}

var bi_groupby = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:        "groupby",
		Description: "groupby(by, list); groups list elements into list of lists of lists as specified by first argument, including the values used to determine grouping",

		// TODO: update
		ParamPositional: []object.Parameter{
			object.Parameter{},
		},
		ParamExpansionMin: 1,
		ParamExpansionMax: 2,
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		const fnName = "groupby"

		var fn, over object.Object

		if len(args) == 2 {
			switch arg1 := args[0].(type) {
			case *object.CompiledCode, *object.BuiltIn:
				fn = arg1
			default:
				return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected function for first argument")
			}
			over = args[1]

		} else {
			over = args[0]
		}

		switch arg2 := over.(type) {
		case *object.List:
			return groupByFunctionOrTruthiness(fnName, pr, fn, true, arg2)

		default:
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected list for second argument")
		}
	},
}

var bi_groupbyH = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:        "groupbyH",
		Description: "groupbyH(by, list); groups list elements into hash as specified by first argument, using the values used to determine grouping as keys",

		// TODO: update
		ParamPositional: []object.Parameter{
			object.Parameter{},
			object.Parameter{},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		const fnName = "groupbyH"

		var fn, over object.Object

		if len(args) == 2 {
			switch arg1 := args[0].(type) {
			case *object.CompiledCode, *object.BuiltIn:
				fn = arg1
			default:
				return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected function for first argument")
			}
			over = args[1]

		} else {
			over = args[0]
		}

		switch arg2 := over.(type) {
		case *object.List:
			return groupByIntoHash(fnName, pr, fn, arg2)

		default:
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected list for second argument")
		}
	},
}

func groupByFunctionOrTruthiness(
	fnName string, pr *Process,
	fn object.Object,
	groupByWithIds bool,
	arr *object.List) object.Object {

	type group struct {
		id   object.Object
		objs []object.Object
	}
	var groups []group
	var key object.Object
	var err error

	if fn == nil {
		// standardize truthiness grouping result
		groups = []group{
			group{id: object.TRUE}, group{id: object.FALSE},
		}
	}

	for _, obj := range arr.Elements {
		if fn == nil {
			if obj.IsTruthy() {
				key = object.TRUE
			} else {
				key = object.FALSE
			}

		} else {
			key, err = pr.callback(fn, obj)
			if err != nil {
				return object.NewError(object.ERR_GENERAL, fnName, err.Error())
			}
		}

		// see if key already in use
		groupId := -1
		for i, gr := range groups {
			if object.Equal(key, gr.id) {
				groupId = i
				break
			}
		}
		if groupId == -1 {
			// not found; make new group
			newGroup := group{id: key, objs: []object.Object{obj}}
			groups = append(groups, newGroup)
		} else {
			// add to existing group
			groups[groupId].objs = append(groups[groupId].objs, obj)
		}
	}

	if len(groups) == 2 && groups[0].id == object.FALSE && groups[1].id == object.TRUE {
		// standardize true/false grouping result
		groups[0], groups[1] = groups[1], groups[0]
	}

	newList := &object.List{}

	if groupByWithIds {
		for _, gr := range groups {
			newList.Elements = append(newList.Elements,
				&object.List{Elements: []object.Object{
					gr.id,
					&object.List{Elements: gr.objs}}},
			)
		}

	} else {
		for _, gr := range groups {
			newList.Elements = append(newList.Elements, &object.List{Elements: gr.objs})
		}
	}
	return newList
}

func groupByIntoHash(
	fnName string, pr *Process,
	fn object.Object,
	arr *object.List) object.Object {

	hash := &object.Hash{}
	var key object.Object
	var err error

	for _, obj := range arr.Elements {
		if fn == nil {
			key = obj
		} else {
			key, err = pr.callback(fn, obj)
			if err != nil {
				return object.NewError(object.ERR_GENERAL, fnName, err.Error())
			}
		}

		value, err := hash.GetValue(key)
		if err != nil {
			value = &object.List{Elements: []object.Object{obj}}
		} else {
			arr2 := value.(*object.List)
			arr2.Elements = append(arr2.Elements, obj)
			value = arr2
		}

		err = hash.WritePair(key, value)
		if err != nil {
			return object.NewError(object.ERR_GENERAL, fnName, err.Error())
		}
	}

	return hash
}
