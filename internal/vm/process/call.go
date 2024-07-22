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
		// find split between positional and optional arguments
		// compiler already verified that optional arguments all come after positional
		optIndex := -1
		for i := len(args) - 1; i > -1; i-- {
			switch args[i].(type) {
			case *object.NameValue:
				optIndex = i
			default:
				break
			}
		}

		if optIndex == -1 {
			args, err = reformArgumentsBySignature(args, nil, code.FnSignature)
		} else {
			args, err = reformArgumentsBySignature(args[:optIndex], args[optIndex:], code.FnSignature)
		}
		if err != nil {
			return
		}
	}

	// make and run the frame
	fr := pr.newFrame(code, baseFr, args)
	defer pr.releaseFrame(fr)
	fnReturn, relay, err = pr.RunFrame(fr, late)
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
	positional, byname []object.Object, sig *object.Signature) (
	args []object.Object, err error) {

	// do parameter compression if applicable
	positional, err = parameterCompression(sig, positional)
	if err != nil {
		return
	}

	// check positional argument counts
	if len(positional) != len(sig.ParamPositional) {
		err = object.NewError(object.ERR_ARGUMENTS, sig.Name,
			fmt.Sprintf("Positional Argument/Parameter Count Mismatch, expected=%s, received=%d",
				sig.MinMaxString(), len(positional)))
		return
	}

	if byname == nil && sig.ParamByName == nil {
		// no optional arguments passed and no optional parameters
		args = positional
		return
	}

	args = make([]object.Object, len(positional)+len(sig.ParamByName))
	copy(args, positional)

	// ---- now to work on optional parameters ----
	argPtr := len(positional)

	// check / pick up optional parameter values in order listed in signature
	// order relevant for compiled functions, as they will be picked up in the same order
	for _, param := range sig.ParamByName {
		found := false
		for i := range byname {
			nv := byname[i].(*object.NameValue)
			if param.ExternalName == nv.Name {
				found = true
				// use passed optional value from argument
				args[argPtr] = nv.Value
				break
			}
		}
		if !found {
			// no argument found for this optional parameter; use default
			args[argPtr] = param.DefaultValue
		}
		argPtr++
	}

	// check if any invalid optional arguments passed
	for i := range byname {
		nv := byname[i].(*object.NameValue)
		found := false
		for _, param := range sig.ParamByName {
			if param.ExternalName == nv.Name {
				found = true
				break
			}
		}
		if !found {
			err = fmt.Errorf("Invalid optional argument passed (%s)", str.ReformatInput(nv.Name))
			return
		}
	}

	return
}

// may convert last positional arguments in a function call into a list
func parameterCompression(sig *object.Signature, args []object.Object) (
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
