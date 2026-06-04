// langur/object/exit_status.go

package object

import (
	"langur/io"
	"langur/system"
	"os"
)

// should not panic/throw
func objectToExitCode(o Object) int {
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

func Exit(codeObj, msgObj Object, consoleTextMode bool) {
	code := objectToExitCode(codeObj)

	if code != 0 && msgObj != nil && msgObj.IsTruthy() {
		// if non-zero return code, write string to standard error, appending a newline
		s := msgObj.String()
		if len(s) != 0 {
			io.PrintLnErr(s, consoleTextMode)
		}
	}
	os.Exit(code)
}
