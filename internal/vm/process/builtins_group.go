// langur/vm/process/builtins_group.go

package process

import (
	"langur/object"
)

// group, groupby, groupbyH

var bi_group = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:        "group",
		Description: "groups list elements into list of lists as specified by argument by",

		ParamPositional: []object.Parameter{
			object.Parameter{ExternalName: "over", Type: object.LIST_OBJ},
		},
		ParamByName: []object.Parameter{
			object.Parameter{ExternalName: "by"},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		const fnName = "group"

		over, by := args[0], args[1]

		var integer int
		var fn object.Object
		var ok bool
		useTruthiness := false

		switch by := by.(type) {
		case *object.Number:
			integer, ok = object.NumberToInt(by)
			if !ok || integer == 0 {
				return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected function or non-zero integer for argument by")
			}

		case nil:
			// not passed; okay
			useTruthiness = true

		case *object.CompiledCode, *object.BuiltIn:
			fn = by

		default:
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected function or non-zero integer for argument by")
		}

		switch over := over.(type) {
		case *object.List:
			if fn != nil || useTruthiness {
				return groupByFunctionOrTruthiness(fnName, pr, fn, false, over)
			}

			// group by integer
			groupList := &object.List{}
			start := 0

			if integer < 0 {
				// negative integer; "start from right" by changing where we start from the left
				integer = -integer
				start = len(over.Elements) % integer
				if start != 0 {
					// not evenly divided
					groupList.Elements = append(groupList.Elements,
						&object.List{Elements: object.CopySlice(over.Elements[:start])})
				}
			}

			for i := start; i < len(over.Elements); i += integer {
				end := i + integer
				if end > len(over.Elements) {
					end = len(over.Elements)
				}
				groupList.Elements = append(groupList.Elements,
					&object.List{Elements: object.CopySlice(over.Elements[i:end])})
			}
			return groupList

		default:
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected list for argument over")
		}
	},
}

var bi_groupby = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:        "groupby",
		Description: "groups list elements into list of lists of lists as specified by argument by, including the values used to determine grouping",

		ParamPositional: []object.Parameter{
			object.Parameter{ExternalName: "over", Type: object.LIST_OBJ},
		},
		ParamByName: []object.Parameter{
			object.Parameter{ExternalName: "by", Type: object.COMPILED_CODE_OBJ},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		const fnName = "groupby"

		over, by := args[0], args[1]

		var fn object.Object

		switch by := by.(type) {
		case *object.CompiledCode, *object.BuiltIn:
			fn = by

		case nil:
			// not passed; okay

		default:
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected function for argument by")
		}

		switch over := over.(type) {
		case *object.List:
			return groupByFunctionOrTruthiness(fnName, pr, fn, true, over)

		default:
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected list for argument over")
		}
	},
}

var bi_groupbyH = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:        "groupbyH",
		Description: "groups list elements into hash as specified by argument by, using the values used to determine grouping as keys",

		ParamPositional: []object.Parameter{
			object.Parameter{ExternalName: "over", Type: object.LIST_OBJ},
		},
		ParamByName: []object.Parameter{
			object.Parameter{ExternalName: "by", Type: object.COMPILED_CODE_OBJ},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		const fnName = "groupbyH"

		over, by := args[0], args[1]

		var fn object.Object

		switch by := by.(type) {
		case *object.CompiledCode, *object.BuiltIn:
			fn = by

		case nil:
			// not passed; okay

		default:
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected function for argument by")
		}

		switch over := over.(type) {
		case *object.List:
			return groupByIntoHash(fnName, pr, fn, over)

		default:
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected list for argument over")
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
