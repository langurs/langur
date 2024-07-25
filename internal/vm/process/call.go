// langur/vm/process/call.go

package process

import (
	"fmt"
	"langur/object"
	"langur/str"
)

// call from VM (opcodes)
func (pr *Process) executeFunctionCall(fr *frame,
	positionalCount, bynameCount int, argExpansion bool) (
	fnReturn object.Object, err error) {

	// having put function to call onto stack after arguments, must pop it first
	fn := pr.pop()
	byname := pr.popMultiple(bynameCount)
	positional := pr.popMultiple(positionalCount)

	if argExpansion {
		switch pre := positional[len(positional)-1].(type) {
		case *object.List:
			positional = append(positional[:len(positional)-1], pre.Elements...)

		default:
			return nil, fmt.Errorf("Expected list for argument expansion")
		}
	}

	switch fn := fn.(type) {
	case *object.CompiledCode:
		if !fn.HasImpureEffects() &&
			(object.SliceHasImpureEffects(positional...) ||
				object.SliceHasImpureEffects(byname...)) {

			return nil, fmt.Errorf("Cannot pass value with impure effects as argument to function not declared as having impure effects")
		}

		fnReturn, _, err = pr.runCompiledCode(fn, fr, positional, byname, nil)

	case *object.BuiltIn:
		if object.SliceHasImpureEffects(positional...) ||
			object.SliceHasImpureEffects(byname...) {
			return nil, fmt.Errorf("Cannot pass impure values as arguments to built-in functions")
		}

		fnReturn, err = pr.callBuiltIn(fn, positional, byname)

	default:
		return nil, fmt.Errorf("Call operation on non-function (%s)", fn.TypeString())
	}

	return fnReturn, err
}

func (pr *Process) runCompiledCode(
	code *object.CompiledCode, baseFr *frame,
	positional, byname, late []object.Object) (
	fnReturn object.Object, relay *jumpRelay, err error) {

	var args []object.Object
	if code.FnSignature != nil {
		args, err = reformArgumentsBySignature(positional, byname, code.FnSignature)
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

func (pr *Process) callBuiltIn(bi *object.BuiltIn, positional, byname []object.Object) (
	result object.Object, err error) {

	defer func() {
		if pr.Modes.GoPanicToLangurException {
			if p := recover(); p != nil {
				err = object.NewErrorFromAnything(p, "panic:"+bi.FullName())
			}
		}
	}()

	if BuiltInArgCountMismatch(bi, len(positional)) {
		return nil, object.NewError(object.ERR_ARGUMENTS, bi.FullName(),
			fmt.Sprintf("Positional argument/parameter count mismatch; expected=%s, received=%d",
				object.ParamExpectedString(bi), len(positional)))
	}

	// TODO: change how parameters are dealt with for built-in functions
	if len(byname) != 0 {
		return nil, object.NewError(object.ERR_ARGUMENTS, bi.FullName(), "Passed invalid argument(s) by name to built-in function")
	}

	// type assertion required on interface{} here
	result = bi.Fn.(BuiltInFunction)(pr, positional...)
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
	positional ...object.Object) (

	object.Object, error) {

	// TODO: callback with parameters by name

	switch fn := fn.(type) {
	case *object.BuiltIn:
		return pr.callBuiltIn(fn, positional, nil)

	case *object.CompiledCode:
		result, _, err := pr.runCompiledCode(fn, pr.currentFrame, positional, nil, nil)
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

	// check positional argument counts after expansion/compression
	if len(positional) != len(sig.ParamPositional) {
		err = object.NewError(object.ERR_ARGUMENTS, sig.Name,
			fmt.Sprintf("Positional argument/parameter count mismatch, expected=%s, received=%d",
				sig.MinMaxString(), len(positional)))
		return
	}

	if byname == nil && sig.ParamByName == nil {
		// no arguments passed by name and none expected
		args = positional
		return
	}

	args = make([]object.Object, len(positional)+len(sig.ParamByName))
	copy(args, positional)

	// ---- now to work on optional parameters ----
	argPtr := len(positional)

	// check / pick up parameter by name values in order listed in signature
	// order relevant internally, as they will be picked up in the same order
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
			// no argument found for this parameter by name
			if param.DefaultValue == nil {
				// no default set; treat as required by name
				err = object.NewError(object.ERR_ARGUMENTS, sig.Name,
					fmt.Sprintf("Required parameter by name (%s) not received", str.ReformatInput(param.ExternalName)))
				return
			}
			// optional parameter without argument; use default
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
			err = fmt.Errorf("Invalid optional argument (%s) passed", str.ReformatInput(nv.Name))
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
