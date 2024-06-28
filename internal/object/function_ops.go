// langur/object/function_ops.go

package object

// Forwarding to a function is handled directly in the process.
// It must have access to the process to do a callback.

// func (fn *CompiledCode) Forward(o2 Object) Object {
// 	// for use in switch, testing with matching operator (~~) against a function
// 	if fn.IsFunction {
// 		return process.call(fn, pr, o2)
// 	}
// 	return nil
// }
