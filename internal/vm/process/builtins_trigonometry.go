// langur/vm/process/builtins_trigonometry.go

package process

import (
	"langur/object"
)

// tan, atan, sine, cos

var bi_tan = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:        "tan",
		Description: "return tangent of a number given in radians",

		// TODO: update
		ParamPositional: []object.Parameter{
			object.Parameter{},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		const fnName = "tan"

		switch n := args[0].(type) {
		case *object.Number:
			return n.Tangent()
		}
		return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected a number")
	},
}

var bi_atan = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:        "atan",
		Description: "return arctangent of a number given in radians",

		// TODO: update
		ParamPositional: []object.Parameter{
			object.Parameter{},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		const fnName = "atan"

		switch n := args[0].(type) {
		case *object.Number:
			return n.ArcTangent()
		}
		return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected a number")
	},
}

var bi_sine = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:        "sine",
		Description: "return sine of a number given in radians",

		// TODO: update
		ParamPositional: []object.Parameter{
			object.Parameter{},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		const fnName = "sine"

		switch n := args[0].(type) {
		case *object.Number:
			return n.Sine()
		}
		return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected a number")
	},
}

var bi_cos = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:        "cos",
		Description: "return cosine of a number given in radians",

		// TODO: update
		ParamPositional: []object.Parameter{
			object.Parameter{},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		const fnName = "cos"

		switch n := args[0].(type) {
		case *object.Number:
			return n.Cosine()
		}
		return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected a number")
	},
}
