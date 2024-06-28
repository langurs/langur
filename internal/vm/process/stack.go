// langur/vm/process/stack.go

package process

import (
	"langur/object"
)

// the general object stack
// locals still carried on frames

func (pr *Process) push(o object.Object) error {
	pr.stack = append(pr.stack, o)
	return nil
}

func (pr *Process) pushMultiple(objs []object.Object) error {
	pr.stack = append(pr.stack, objs...)
	return nil
}

func (pr *Process) pop() object.Object {
	L := len(pr.stack)
	pr.LastValue = pr.stack[L-1]
	pr.stack = pr.stack[:L-1]
	return pr.LastValue
}

func (pr *Process) popMultiple(count int) []object.Object {
	L := len(pr.stack)
	elements := object.CopyRefSlice(pr.stack[L-count : L])
	pr.stack = pr.stack[:L-count]
	return elements
}

// let us see what's at the top of the stack without removing it
func (pr *Process) look() object.Object {
	return pr.stack[len(pr.stack)-1]
}
