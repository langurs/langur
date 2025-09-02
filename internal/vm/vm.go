// langur/vm.go

package vm

import (
	"langur/trace"
	"fmt"
	"langur/args"
	"langur/bytecode"
	"langur/modes"
	"langur/object"
	"langur/vm/process"
	"os"
	"strings"
)

func bug(fnName, s string) {
	// panic for now; may change later
	panic("VM Bug: " + s)
}

type VM struct {
	process *process.Process
	late    []string
}

func New(byteCode *bytecode.ByteCode, m *modes.VmModes) *VM {
	vm := &VM{
		process: process.New(byteCode.Constants, byteCode.StartCode, m),
		late:    byteCode.Late,
	}

	// set default modes
	// FIXME: not safe for concurrency
	object.SetDivisionMaxScaleMode(vm.process.Modes.DivisionMaxScale)

	return vm
}

// used by the REPL
func NewWithGlobalStore(byteCode *bytecode.ByteCode, globals []object.Object, m *modes.VmModes) *VM {
	vm := New(byteCode, m)
	vm.process.SetStartFrameLocals(globals)
	return vm
}

func (vm *VM) Run() (err error, where *trace.Where) {
	// to push late-binding assignments onto the stack before executing the global frame, ...
	// ... which should already contain the opcodes to retrieve them
	var late []object.Object
	
	late, err = vm.gatherLateBindings()
	if err != nil {
		return
	}
	
	_, _, err = vm.process.RunFrame(nil, late)
	return
}

func (vm *VM) gatherLateBindings() (late []object.Object, err error) {
	// NOTE: These are coordinated with a compiler list of late bindings (must be paired with the right opcodes).

	if vm.late == nil {
		return nil, nil
	}

	// langur, langurArgs, file, fileArgs, err := OsArgsToArgs()
	_, _, file, fileArgs, err := args.OsArgsToArgs()
	if err != nil {
		return nil, err
	}

	for _, v := range vm.late {
		switch v {
		case "_env":
			env := &object.Hash{}
			for _, kv := range os.Environ() {
				keyval := strings.Split(kv, "=")
				env.WritePair(object.NewString(keyval[0]), object.NewString(keyval[1]))
			}
			late = append(late, env)

		case "_args":
			// script arguments, not compiler/VM arguments
			args := &object.List{}
			for _, s := range fileArgs {
				args.Elements = append(args.Elements, object.NewString(s))
			}
			late = append(late, args)

		case "_file":
			late = append(late, object.NewString(file))

		default:
			bug("vm.Run", "Unknown late binding "+v)
			err = fmt.Errorf("Unknown late binding %q", v)
		}
	}

	return
}

// for the REPL and for testing
func (vm *VM) LastValue() object.Object {
	return vm.process.LastValue
}

func (vm *VM) LastModes() *modes.VmModes {
	return vm.process.Modes.Copy()
}
