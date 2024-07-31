// langur/vm/process/process.go

package process

import (
	"fmt"
	"langur/modes"
	"langur/object"
)

func bug(fnName, s string) {
	// panic for now; may change later
	panic("Process Bug: " + s)
}

type Process struct {
	constants    []object.Object
	startFrame   *frame
	currentFrame *frame

	// see vm/frames.go
	frameAlloc []frame
	fap        int

	// general object stack; not used for variables
	stack     []object.Object
	LastValue object.Object

	Modes *modes.VmModes
}

type jumpRelay struct {
	Jump  int
	Level int
	Value object.Object
}

func New(
	constants []object.Object,
	startCode *object.CompiledCode,
	m *modes.VmModes) *Process {

	pr := &Process{
		constants: constants,
	}
	if m == nil {
		pr.Modes = modes.NewVmModes()
	} else {
		// NOTE: Concurrency would likely require us to copy modes for a new process.
		pr.Modes = m.Copy()
	}
	pr.startFrame = pr.newFrame(startCode, nil, nil)
	return pr
}

func (pr *Process) SetStartFrameLocals(slc []object.Object) {
	pr.startFrame.locals = slc
}

func (pr *Process) executeTryCatch(fr *frame, tryIndex, catchIndex, elseIndex int) (
	fnReturn object.Object, relay *jumpRelay, err error) {

	tryCode := pr.constants[tryIndex].(*object.CompiledCode)

	fnReturn, relay, err = pr.runCompiledCode(tryCode, fr, nil, nil, nil)
	if err == nil {
		if elseIndex != 0 {
			// else block used on catch to run only when there is no exception
			elseCode := pr.constants[elseIndex].(*object.CompiledCode)
			fnReturn, relay, err = pr.runCompiledCode(elseCode, fr, nil, nil, nil)
		}

	} else {
		catchCode := pr.constants[catchIndex].(*object.CompiledCode)

		// catch error returned from try frame
		// The error variable is a langur hash that is guaranteed to contain certain keys, even if they have no data.
		errObj := object.NewErrorFromAnything(err, "")
		// The compiler already set up this late-binding assignment in the catch code, ...
		// ... so we pass the exception hash to be pushed onto the stack of the catch frame (errObj.Contents)

		fnReturn, relay, err =
			pr.runCompiledCode(catchCode, fr, nil, nil, []object.Object{errObj.Contents})
	}

	return
}

func (pr *Process) throw(fr *frame, what object.Object) error {
	errObj := object.NewErrorFromObject(what)

	// write function name source to hash if source not already set (is ZLS)
	src, err := errObj.Contents.GetValue(object.ERR_HASHKEY_SOURCE)
	if err != nil {
		err = fmt.Errorf("Error retrieving Error Object Source: %s", err.Error())
		bug("Process.throw", err.Error())
		return err
	}
	if src.String() == "" {
		errFnName, ok := fr.getFnName()
		if ok {
			errObj.Contents.WritePair(object.ERR_HASHKEY_SOURCE, object.NewString(errFnName))
		}
	}
	return errObj
}

func (pr *Process) pushFunction(constIndex, freeCount, optionalsCount int) error {
	constant := pr.constants[constIndex]
	compiledFn, ok := constant.(*object.CompiledCode)
	if !ok {
		bug("pushFunction", fmt.Sprintf("Not a function: %T", constant))
		return fmt.Errorf("Not a function: %s", constant.TypeString())
	}

	if optionalsCount != 0 {
		// for any optional parameter defaults that weren't set at compile-time
		// name/value pairs
		optionals := pr.popMultiple(optionalsCount * 2)

		for i := 0; i < len(optionals); i += 2 {
			// name should be a string object
			err := compiledFn.FnSignature.SetParamDefault(optionals[i].String(), optionals[i+1])
			if err != nil {
				return err
			}
		}
	}

	if freeCount != 0 {
		// copy to prevent one closure from clobbering another closure's free slice ...
		// ... when there is a single *object.CompiledCode, but more than one definition
		compiledFn = compiledFn.Copy().(*object.CompiledCode)
		compiledFn.Free = pr.popMultiple(freeCount)
	}

	return pr.push(compiledFn)
}
