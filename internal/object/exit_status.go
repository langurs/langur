// langur/object/exit_status.go

package object

import (
	"langur/system"
)

// should not panic/throw
func ObjectToExitCode(o Object) int {
	code := 0 // 0 = success
	var err error

	switch codeArg := o.(type) {
	case *Number:
		code, err = codeArg.ToInt()
		if err != nil {
			// failure to convert to native integer
			code = system.GetExitStatus(system.ExitStatusArgToExitBad)
		}
		code = system.FixExitStatus(code)

	case *Boolean:
		//  true: success (code 0)
		// false: general failure
		if !codeArg.Value {
			code = system.GetExitStatus(system.ExitStatusGeneral)
		}

	default:
		// invalid code argument for exit status
		code = system.GetExitStatus(system.ExitStatusArgToExitBad)
	}

	return code
}
