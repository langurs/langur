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
func bi_execT(pr *Process, args ...object.Object) object.Object {
	const fnName = "execT"

	cmd, ok := args[0].(*object.String)
	if !ok {
		return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected string for command to execute")
	}

	out, err := execCmd(cmd.String())

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

func bi_execTH(pr *Process, args ...object.Object) object.Object {
	const fnName = "execTH"

	cmd, ok := args[0].(*object.String)
	if !ok {
		return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected string for command to execute")
	}

	out, err := execCmd(cmd.String())

	hash := &object.Hash{}

	if err == nil {
		hash.WritePair(object.NewString("result"), object.NewString(out))
		hash.WritePair(object.NewString("error"), object.NONE)
		hash.WritePair(object.NewString("status"), object.Zero)

	} else {
		hash.WritePair(object.NewString("result"), object.ZLS)
		hash.WritePair(object.NewString("error"), object.NewString(formatErrString(out)))
		hash.WritePair(object.NewString("status"),
			object.NumberFromInt(getExitStatusFromExecError(err.Error())))
	}
	return hash
}

var getExitStatusRegex = regexp.MustCompile("^exit status (-?\\d+)")

func getExitStatusFromExecError(arg string) int {
	ns := getExitStatusRegex.FindStringSubmatch(arg)
	if len(ns) == 0 || ns[1] == "" {
		return system.GetExitStatus("")
	}
	code, err := str.StrToInt(ns[1], 10)
	if err != nil {
		return system.GetExitStatus("")
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
