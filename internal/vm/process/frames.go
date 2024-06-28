// langur/vm/process/frames.go

package process

import (
	"langur/object"
)

// When you exit a frame, you automatically "pop" local values.
// Frames are not just used for functions, but also for try and catch blocks, and for scope blocks.
// Not all frames have scope, so not all will have locals.

type frame struct {
	base   *frame
	code   *object.CompiledCode
	locals []object.Object
	inUse  bool
}

func (pr *Process) newFrame(
	code *object.CompiledCode, baseFr *frame,
	args []object.Object) *frame {

	// Using a managed allocation of frames gives us about a 30% improvement ...
	// ... in speed over allocating a new frame every time one is needed.
	// We do some cleaning when a frame is released.
	if pr.fap > len(pr.frameAlloc)-1 {
		// adding to a slice with no more room ...
		pr.frameAlloc = append(pr.frameAlloc, frame{})
	}

	fr := &pr.frameAlloc[pr.fap]
	pr.fap++

	fr.inUse = true
	fr.code = code
	fr.base = baseFr

	if code.LocalBindingsCount == 0 {
		fr.locals = nil
	} else {
		if code.LocalBindingsCount != len(fr.locals) {
			fr.locals = make([]object.Object, code.LocalBindingsCount)
			fr.locals = replaceNilInObjectSlice(fr.locals)
		}
		copy(fr.locals, args) // parameters as first locals
	}

	pr.currentFrame = fr
	return fr
}

func replaceNilInObjectSlice(oSlc []object.Object) []object.Object {
	// to ensure nil doesn't cause panics
	for i := range oSlc {
		if oSlc[i] == nil {
			oSlc[i] = object.NONE
		}
	}
	return oSlc
}

func (pr *Process) releaseFrame(fr *frame) {
	// clean some of frame at release
	if len(fr.locals) > 0 {
		for i := range fr.locals {
			fr.locals[i] = object.NONE
		}
	}

	pr.currentFrame = fr.base
	fr.base = nil
	fr.inUse = false

	// step down fap as possible
	for pr.fap > 0 && !pr.frameAlloc[pr.fap-1].inUse {
		pr.fap--
	}

	// frees some memory?
	if pr.fap*3 < len(pr.frameAlloc) {
		pr.frameAlloc = pr.frameAlloc[:pr.fap+1]
	}
}

// get/set values
// no functions to get/set globals b/c they are relatively simple

func (fr *frame) getSelf() (obj object.Object, err error) {
	if fr.code.IsFunction {
		return fr.code, nil
	}
	return fr.base.getSelf()
}

func (fr *frame) getFree(freeIndex int) (obj object.Object, err error) {
	if fr.code.IsFunction {
		return fr.code.Free[freeIndex], nil
	}
	return fr.base.getFree(freeIndex)
}

func (fr *frame) getLocal(localIndex int) (obj object.Object, err error) {
	if fr.code.LocalBindingsCount > 0 {
		return fr.locals[localIndex], nil
	}
	return fr.base.getLocal(localIndex)
}

func (fr *frame) setLocal(localIndex int, setTo object.Object) {
	if fr.code.LocalBindingsCount > 0 {
		fr.locals[localIndex] = setTo
	} else {
		fr.base.setLocal(localIndex, setTo)
	}
}

func (fr *frame) getNonLocal(localIndex, count int) (obj object.Object, err error) {
	if count == 0 {
		return fr.locals[localIndex], nil
	}
	return fr.base.getNonLocal(localIndex, count-1)
}

func (fr *frame) setNonLocal(localIndex, count int, setTo object.Object) {
	if count == 0 {
		fr.locals[localIndex] = setTo
	} else {
		fr.base.setNonLocal(localIndex, count-1, setTo)
	}
}

func (fr *frame) setLocalIndexedValue(localIndex int, objIndex, setTo object.Object) (err error) {
	if fr.code.LocalBindingsCount > 0 {
		var setObj object.Object
		setObj, err = object.SetIndex(fr.locals[localIndex], objIndex, setTo)
		if err != nil {
			return
		}
		fr.locals[localIndex] = setObj
		return nil
	} else {
		return fr.base.setLocalIndexedValue(localIndex, objIndex, setTo)
	}
}

func (fr *frame) setNonLocalIndexedValue(localIndex, count int, objIndex, setTo object.Object) (err error) {
	if count == 0 {
		var setObj object.Object
		setObj, err = object.SetIndex(fr.locals[localIndex], objIndex, setTo)
		if err != nil {
			return
		}
		fr.locals[localIndex] = setObj
		return nil
	} else {
		return fr.base.setNonLocalIndexedValue(localIndex, count-1, objIndex, setTo)
	}
}

func (fr *frame) getFnName() (string, bool) {
	if fr.code.IsFunction && fr.code.Name != "" {
		return fr.code.Name, true
	} else if fr.base != nil {
		return fr.base.getFnName()
	}
	return "", false
}
