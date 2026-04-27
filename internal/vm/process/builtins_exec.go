// langur/vm/process/builtins_exec.go

package process

import (
	"langur/object"
	"langur/regexp"
	"langur/str"
	"langur/system"
	"os/exec"
)

// execT, execTH
// executed trusted source string

var bi_execT = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:          "execT",
		ImpureEffects: true,
		Description:   "executes a command from a trusted source string, returning a result or throwing an exception",
		ParamPositional: []object.Parameter{
			object.Parameter{ExternalName: "source", Type: object.STRING_OBJ},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		const fnName = "execT"

		out, err := execCmd(args[0].String())

		// result or exception
		if err == nil {
			return object.NewString(out)
		} else {
			es := err.Error()
			err := object.NewError(object.ERR_GENERAL, fnName,
				es+": "+formatErrString(out))

			// add the exit status to the hash
			err.Contents.WritePair(object.NewString("status"),
				object.NumberFromInt(getExitStatusFromExecError(es)))

			return err
		}
	},
}

func execCmd(s string) (string, error) {
	var out []byte
	var err error

	if system.Type == system.WINDOWS {
		out, err = exec.Command("cmd.exe", "/C", s).CombinedOutput()
	} else {
		out, err = exec.Command("sh", "-c", s).CombinedOutput()
	}
	return string(out), err
}

var bi_execTH = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:          "execTH",
		ImpureEffects: true,
		Description:   "executes a command from a trusted source string, returning a hash",
		ParamPositional: []object.Parameter{
			object.Parameter{ExternalName: "source", Type: object.STRING_OBJ},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		const fnName = "execTH"

		out, err := execCmd(args[0].String())

		hash := &object.Hash{}

		if err == nil {
			hash.WritePair(object.NewString("result"), object.NewString(out))
			hash.WritePair(object.NewString("error"), object.NONE)
			hash.WritePair(object.NewString("status"), object.Zero)

		} else {
			hash.WritePair(object.NewString("result"), object.ZeroLengthString())
			hash.WritePair(object.NewString("error"), object.NewString(formatErrString(out)))
			hash.WritePair(object.NewString("status"),
				object.NumberFromInt(getExitStatusFromExecError(err.Error())))
		}
		return hash
	},
}

var getExitStatusRegex = regexp.MustCompile("^exit status (-?\\d+)")

func getExitStatusFromExecError(over string) int {
	ns := getExitStatusRegex.FindStringSubmatch(over)
	if len(ns) == 0 || ns[1] == "" {
		return system.GetExitStatus(system.ExitStatusGeneral)
	}
	code, err := str.StrToInt(ns[1], 10)
	if err != nil {
		return system.GetExitStatus(system.ExitStatusGeneral)
	}
	return code
}

func formatErrString(s string) string {
	// remove preceding "sh: "
	if len(s) > 3 && s[0:4] == "sh: " {
		s = s[4:]
	}
	// remove trailing newline
	if len(s) > 0 && s[len(s)-1] == '\n' {
		s = s[:len(s)-1]
	}
	if len(s) > 0 && s[len(s)-1] == '\r' {
		s = s[:len(s)-1]
	}
	return s
}
