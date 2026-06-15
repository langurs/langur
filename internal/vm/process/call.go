// langur/vm/process/call.go

package process

import (
	"fmt"
	"langur/object"
	"langur/str"
)

// call from VM (opcodes)
func (pr *Process) executeFunctionCall(fr *frame,
	positionalCount, keywordCount int, argExpansion bool) (
	fnReturn object.Object, err error) {

	// having put function to call onto stack after arguments, must pop it first
	fn := pr.pop()
	keyword := pr.popMultiple(keywordCount)
	positional := pr.popMultiple(positionalCount)

	if argExpansion {
		switch pre := positional[len(positional)-1].(type) {
		case *object.List:
			positional = append(positional[:len(positional)-1], pre.Elements...)

		default:
			err = fmt.Errorf("Expected list for argument expansion")
			return
		}
	}

	switch fn := fn.(type) {
	case *object.CompiledCode:
		if !fn.HasImpureEffects() &&
			(object.SliceHasImpureEffects(positional...) ||
				object.SliceHasImpureEffects(keyword...)) {

			err = fmt.Errorf("Cannot pass value with impure effects as argument to function not declared as having impure effects")
			return
		}

		fnReturn, _, err = pr.runCompiledCode(fn, fr, positional, keyword, nil)

	case *object.BuiltIn:
		if object.SliceHasImpureEffects(positional...) ||
			object.SliceHasImpureEffects(keyword...) {
			return nil, fmt.Errorf("Cannot pass impure values as arguments to built-in functions")
		}

		fnReturn, err = pr.callBuiltIn(fn, positional, keyword)

	default:
		err = fmt.Errorf("Call operation on non-function (%s)", fn.TypeString())
	}

	return
}

// callback from built-in functions
// TODO: include keyword paramaters for callback
func (pr *Process) callback(
	fn object.Object,
	positional ...object.Object) (

	fnReturn object.Object, err error) {

	switch fn := fn.(type) {
	case *object.BuiltIn:
		fnReturn, err = pr.callBuiltIn(fn, positional, nil)
		return

	case *object.CompiledCode:
		fnReturn, _, err = pr.runCompiledCode(fn, pr.currentFrame, positional, nil, nil)
		return
	}

	err = fmt.Errorf("Not a callable object (%s)", fn.TypeString())
	return
}

func (pr *Process) runCompiledCode(
	code *object.CompiledCode, baseFr *frame,
	positional, keyword, late []object.Object) (
	fnReturn object.Object, relay *jumpRelay, err error) {

	var args []object.Object
	if code.FnSignature != nil {
		args, err = reformArgumentsBySignature(positional, keyword, code.FnSignature)
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

func (pr *Process) callBuiltIn(bi *object.BuiltIn, positional, keyword []object.Object) (
	result object.Object, err error) {

	defer func() {
		if pr.Modes.GoPanicToLangurException {
			if p := recover(); p != nil {
				err = object.NewExceptionFromAnything(p, "panic:"+bi.FullName())
			}
		}
	}()

	var args []object.Object
	args, err = reformArgumentsBySignature(positional, keyword, bi.FnSignature)
	if err != nil {
		return
	}

	// type assertion required on interface{} here
	// (interface{} required to prevent a circular reference)
	result = bi.Fn.(BuiltInFunction)(pr, args...)

	// if received an Exception Object (from a built-in function), ... 
	// ... swap so that error is second value returned from this function
	if r, isException := result.(*object.Exception); isException {
		return nil, r // result.(*object.Exception)
	} else {
		return result, nil
	}
}

func reformArgumentsBySignature(
	positional, keyword []object.Object, sig *object.Signature) (
	args []object.Object, err error) {

	// do parameter compression if applicable
	positional, err = parameterCompression(sig, positional)
	if err != nil {
		return
	}

	// check positional argument counts after expansion/compression
	if len(positional) != len(sig.ParamPositional) {
		err = object.NewException(object.ERR_ARGUMENTS, sig.Name,
			fmt.Sprintf("Positional argument/parameter count mismatch, expected=%s, received=%d",
				sig.MinMaxString(), len(positional)))
		return
	}

	for argPtr, param := range sig.ParamPositional {
		// check positional parameters for adherance to explicit typing (in the signature)
		// already checked counts
		// NOTE: explicit typing not accepted on parameter expansion; If this changes, it will need to be accounted for.
		if param.Type != 0 {
			if param.Type != positional[argPtr].Type() {
				argTypeName := object.TypeToTypeName(positional[argPtr].Type())
				paramTypeName := object.TypeToTypeName(param.Type)
				err = object.NewException(object.ERR_ARGUMENTS, sig.Name,
					fmt.Sprintf("Argument %d type (%s) does not match parameter %s type (%s)", 
						argPtr+1, argTypeName, str.Nz(param.InternalName, param.ExternalName), paramTypeName))
				return
			}
		}

		if param.Mutable {
			// var keyword used on parameter; treat as copy on import
			positional[argPtr] = positional[argPtr].Copy()
		}
	}

	if keyword == nil && sig.ParamKeyword == nil {
		// no keyword arguments passed and none expected
		args = positional
		return
	}

	// ---- now to work on keyword parameters ----
	args = make([]object.Object, len(positional)+len(sig.ParamKeyword))
	copy(args, positional)

	// keyword parameters always following positional parameters
	argPtr := len(positional)

	// check / pick up keyword parameter values in order listed in signature
	// order relevant, as they will be picked up in order by the function
	for _, param := range sig.ParamKeyword {
		found := false
		for i := range keyword {
			nv := keyword[i].(*object.NameValue)
			if param.ExternalName == nv.Name {
				found = true
				// use passed optional value from argument
				args[argPtr] = nv.Value
				break
			}
		}

		if !found {
			if param.Required {
				err = object.NewException(object.ERR_ARGUMENTS, sig.Name,
					fmt.Sprintf("Required keyword parameter (%s) not passed", param.ExternalName))
				return
			}

			// optional parameter without argument; use default
			args[argPtr] = param.DefaultValue

			// NOTE: nil default okay for built-ins
			// Use nil default for built-ins to check if passed a value.
			// Compiled functions will use another means.
		}

		// check keyword parameters for adherance to explicit typing (in the signature).
		if found && param.Type != 0 {
			if param.Type != args[argPtr].Type() {
				argTypeName := object.TypeToTypeName(args[argPtr].Type())
				paramTypeName := object.TypeToTypeName(param.Type)
				err = object.NewException(object.ERR_ARGUMENTS, sig.Name,
					fmt.Sprintf("Argument %s type (%s) does not match parameter %s type (%s)", 
						param.ExternalName, argTypeName, param.ExternalName, paramTypeName))
				return
			}
		}

		if param.Mutable {
			// var keyword used on parameter; treat as copy on import
			args[argPtr] = args[argPtr].Copy()
		}

		argPtr++
	}

	// check if any invalid optional arguments passed
	for i := range keyword {
		nv := keyword[i].(*object.NameValue)

		found := false
		for _, param := range sig.ParamKeyword {
			if param.ExternalName == nv.Name {
				found = true
				break
			}
		}
		if !found {
			err = object.NewException(object.ERR_ARGUMENTS, sig.Name,
				fmt.Sprintf("Invalid optional argument (%s) passed", str.ReformatInput(nv.Name)))
			return
		}
	}

	return
}

// convert last positional arguments in a function call into a list
func parameterCompression(sig *object.Signature, positional []object.Object) (
	params []object.Object, err error) {

	params = positional
	if sig.ParamExpansionMax != 0 || sig.ParamExpansionMin != 0 {
		var last []object.Object
		diff := len(positional) - len(sig.ParamPositional) + 1
		if diff > 0 && len(positional) != 0 {
			// parameter compression required
			// must copy before use in append following
			last = object.CopyRefSlice(positional[len(sig.ParamPositional)-1:])
			params = append(positional[:len(sig.ParamPositional)-1], &object.List{Elements: last})
	
			if sig.ParamExpansionMax != -1 && len(last) > sig.ParamExpansionMax {
				err = object.NewException(object.ERR_ARGUMENTS, sig.Name, 
					fmt.Sprintf("Parameter expansion max (%d) exceeded (%d)", sig.ParamExpansionMax, len(last)))
			}
	
		} else if diff == 0 && sig.ParamExpansionMin == 0 {
			// received 0 and none required; since it's missing, add the empty list
			params = append(positional, object.EmptyList())
		}
	
		if len(last) < sig.ParamExpansionMin {
				err = object.NewException(object.ERR_ARGUMENTS, sig.Name, 
					fmt.Sprintf("Parameter expansion min (%d) not met (%d)", sig.ParamExpansionMin, len(last)))
		}
	}

	return
}
