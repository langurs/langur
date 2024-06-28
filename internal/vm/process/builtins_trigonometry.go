// langur/vm/process/builtins_trigonometry.go

package process

import (
	"langur/object"
)

// tan, atan, sine, cos

func bi_tan(pr *Process, args ...object.Object) object.Object {
	const fnName = "tan"

	switch n := args[0].(type) {
	case *object.Number:
		return n.Tangent()
	}
	return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected a number")
}

func bi_atan(pr *Process, args ...object.Object) object.Object {
	const fnName = "atan"

	switch n := args[0].(type) {
	case *object.Number:
		return n.ArcTangent()
	}
	return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected a number")
}

func bi_sine(pr *Process, args ...object.Object) object.Object {
	const fnName = "sine"

	switch n := args[0].(type) {
	case *object.Number:
		return n.Sine()
	}
	return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected a number")
}

func bi_cos(pr *Process, args ...object.Object) object.Object {
	const fnName = "cos"

	switch n := args[0].(type) {
	case *object.Number:
		return n.Cosine()
	}
	return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected a number")
}
