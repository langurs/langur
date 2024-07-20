// langur/vm/process/call.go

package process

import (
	"fmt"
	"langur/object"
	"langur/str"
)

// call from VM (opcodes)
func (pr *Process) executeFunctionCall(fr *frame, argCount int, argExpansion bool) (
	fnReturn object.Object, err error) {

	// having put function to call onto stack after arguments, must pop it first
	fn := pr.pop()
	args := pr.popMultiple(argCount)

	if argExpansion {
		switch pre := args[len(args)-1].(type) {
		case *object.List:
			args = append(args[:len(args)-1], pre.Elements...)

		default:
			return nil, fmt.Errorf("Expected list for argument expansion")
		}
	}

	switch fn := fn.(type) {
	case *object.CompiledCode:
		if !fn.HasImpureEffects() && object.SliceHasImpureEffects(args...) {
			return nil, fmt.Errorf("Cannot pass function with impure effects as argument to function not declared as having impure effects")
		}

		fnReturn, _, err = pr.runCompiledCode(fn, fr, args, nil)

	case *object.BuiltIn:
		if object.SliceHasImpureEffects(args...) {
			return nil, fmt.Errorf("Cannot pass impure functions as arguments to built-in functions")
		}

		fnReturn, err = pr.callBuiltIn(fn, args)

	default:
		return nil, fmt.Errorf("Call operation on non-function (%s)", fn.TypeString())
	}

	return fnReturn, err
}

func (pr *Process) runCompiledCode(
	code *object.CompiledCode, baseFr *frame, args, late []object.Object) (
	fnReturn object.Object, relay *jumpRelay, err error) {

	if code.FnSignature != nil {
		// convert args if necessary
		args, err = parameterCompression(code, args)
		if err != nil {
			return
		}

		// check arg counts after expansion/compression, not including optional parameters
		argCnt := len(args)
		max := len(code.FnSignature.ParamPositional)
		if argCnt != max && max > -1 {
			name := "fn"
			if code.FnSignature.Name != "" {
				name = code.FnSignature.Name
			}
			return nil, nil,
				object.NewError(object.ERR_ARGUMENTS, name,
					fmt.Sprintf("Argument/Parameter Count Mismatch, expected=%s, received=%d",
						code.FnSignature.MinMaxString(), argCnt))
		}
	}

	// make and run the frame
	fr := pr.newFrame(code, baseFr, args)
	defer pr.releaseFrame(fr)
	fnReturn, relay, err = pr.RunFrame(fr, late)
	return
}

// may convert last positional arguments in a function call into a list
func parameterCompression(fn *object.CompiledCode, args []object.Object) (
	params []object.Object, err error) {

	params = args
	if fn.FnSignature.ParamExpansionMax != 0 || fn.FnSignature.ParamExpansionMin != 0 {
		var last []object.Object
		max := len(fn.FnSignature.ParamPositional)
		diff := len(args) - max + 1

		if diff > 0 && len(args) != 0 {
			// must copy before use in append following
			last = object.CopyRefSlice(args[max-1:])
			params = append(args[:max-1], &object.List{Elements: last})

			if fn.FnSignature.ParamExpansionMax != -1 && len(last) > fn.FnSignature.ParamExpansionMax {
				err = fmt.Errorf("Parameter expansion max (%d) exceeded (%d)", fn.FnSignature.ParamExpansionMax, len(last))
			}

		} else if diff == 0 && fn.FnSignature.ParamExpansionMin == 0 {
			// not a required parameter; since it's missing, add an empty list
			params = append(args, &object.List{})
		}

		if len(last) < fn.FnSignature.ParamExpansionMin {
			err = fmt.Errorf("Parameter expansion min (%d) not met (%d)", fn.FnSignature.ParamExpansionMin, len(last))
		}
	}

	return
}

func (pr *Process) callBuiltIn(bi *object.BuiltIn, args []object.Object) (
	result object.Object, err error) {

	defer func() {
		if pr.Modes.GoPanicToLangurException {
			if p := recover(); p != nil {
				err = object.NewErrorFromAnything(p, "panic:"+bi.Name)
			}
		}
	}()

	if BuiltInArgCountMismatch(bi, len(args)) {
		return nil, object.NewError(object.ERR_ARGUMENTS, bi.Name,
			fmt.Sprintf("Argument/parameter count mismatch; expected=%s, received=%d",
				object.ParamExpectedString(bi), len(args)))
	}

	// type assertion required on interface{} here
	result = bi.Fn.(BuiltInFunction)(pr, args...)
	if result.Type() == object.ERROR_OBJ {
		return nil, result.(*object.Error)
	} else {
		return result, nil
	}
}

func BuiltInArgCountMismatch(bi *object.BuiltIn, count int) bool {
	if count < bi.ParamMin || bi.ParamMax != -1 && count > bi.ParamMax {
		return true
	}
	return false
}

// callback from built-in functions
func (pr *Process) call(fn object.Object,
	args ...object.Object) (

	object.Object, error) {

	switch fn := fn.(type) {
	case *object.BuiltIn:
		return pr.callBuiltIn(fn, args)

	case *object.CompiledCode:
		result, _, err := pr.runCompiledCode(fn, pr.currentFrame, args, nil)
		return result, err
	}

	return nil, fmt.Errorf("Not a callable object (%s)", fn.TypeString())
}

func reformArgumentsBySignature(
	args *object.ArgumentPackage, sig *object.Signature) (
	args2 []object.Object, err error) {

	var positional []object.Object

	// do parameter compression if applicable
	positional, err = parameterCompressionNEW(sig, args.ArgsPositional)
	if err != nil {
		return
	}

	// check positional argument counts
	if len(positional) != len(sig.ParamPositional) {
		return nil,
			object.NewError(object.ERR_ARGUMENTS, sig.Name,
				fmt.Sprintf("Positional Argument/Parameter Count Mismatch, expected=%s, received=%d",
					sig.MinMaxString(), len(positional)))
	}

	args2 = make([]object.Object, 0, len(positional)+len(sig.ParamByName))
	copy(positional, args2)

	// ---- now to work on optional parameters ----

	// check / pick up optional parameter values in order listed in signature
	// order relevant for compiled functions, as they will be picked up in the same order
	for _, param := range sig.ParamByName {
		found := false
		for argExtName, argValue := range args.ArgsByName {
			if param.ExternalName == argExtName {
				found = true
				// use passed optional value from argument
				args2 = append(args2, argValue)
			}
		}
		if !found {
			// no argument found for this one; all is well; use default
			args2 = append(args2, param.DefaultValue)
		}
	}

	// check if any invalid optional arguments passed
	for argExtName := range args.ArgsByName {
		found := false
		for _, param := range sig.ParamByName {
			if param.ExternalName == argExtName {
				found = true
			}
		}
		if !found {
			err = fmt.Errorf("Invalid optional argument passed (%s)", str.ReformatInput(argExtName))
			return
		}
	}

	return
}

// may convert last arguments in a function call into a list
func parameterCompressionNEW(sig *object.Signature, args []object.Object) (
	params []object.Object, err error) {

	params = args
	if sig.ParamExpansionMax != 0 || sig.ParamExpansionMin != 0 {
		var last []object.Object
		diff := len(args) - len(sig.ParamPositional) + 1
		if diff > 0 && len(args) != 0 {
			// must copy before use in append following
			last = object.CopyRefSlice(args[len(sig.ParamPositional)-1:])
			params = append(args[:len(sig.ParamPositional)-1], &object.List{Elements: last})

			if sig.ParamExpansionMax != -1 && len(last) > sig.ParamExpansionMax {
				err = fmt.Errorf("Parameter expansion max (%d) exceeded (%d)", sig.ParamExpansionMax, len(last))
			}

		} else if diff == 0 && sig.ParamExpansionMin == 0 {
			// not a required parameter; since it's missing, add an empty list
			params = append(args, &object.List{})
		}

		if len(last) < sig.ParamExpansionMin {
			err = fmt.Errorf("Parameter expansion min (%d) not met (%d)", sig.ParamExpansionMin, len(last))
		}
	}

	return
}
