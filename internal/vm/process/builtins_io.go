// langur/vm/process/builtins_io.go

package process

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"langur/object"
	"langur/str"
	"os"
)

// write, writeln, writeErr, writelnErr, read

func bi_write(pr *Process, args ...object.Object) object.Object {
	var out bytes.Buffer

	for _, v := range args {
		out.WriteString(v.String())
	}

	s := out.String()
	if pr.Modes.ConsoleTextMode {
		s = str.ReplaceNewLinesWithSystem(s)
	}

	if len(s) == 0 {
		return object.NULL
	}
	fmt.Print(s)
	return object.TRUE
}

func bi_writeln(pr *Process, args ...object.Object) object.Object {
	return bi_write(pr, append(args, object.NewString(str.SysNewLine))...)
}

func bi_writeErr(pr *Process, args ...object.Object) object.Object {
	var out bytes.Buffer

	for _, v := range args {
		out.WriteString(v.String())
	}

	s := out.String()
	if pr.Modes.ConsoleTextMode {
		s = str.ReplaceNewLinesWithSystem(s)
	}

	if len(s) == 0 {
		return object.NULL
	}
	fmt.Fprint(os.Stderr, s)
	return object.TRUE
}

func bi_writelnErr(pr *Process, args ...object.Object) object.Object {
	return bi_writeErr(pr, append(args, object.NewString(str.SysNewLine))...)
}

func bi_read(pr *Process, args ...object.Object) object.Object {
	var fn, alternate object.Object
	var re *object.Regex
	prompt := ""
	errMsg := ""
	maxattempts := 1
	validationByRegex := false

	// gather parameters
	if len(args) > 0 {
		p, ok := args[0].(*object.String)
		if !ok {
			return object.NewError(object.ERR_ARGUMENTS, "read", "Expected string for prompt")
		}
		prompt = p.String()

		if pr.Modes.ConsoleTextMode {
			prompt = str.ReplaceNewLinesWithSystem(prompt)
		}

		if len(args) > 1 {
			re, ok = args[1].(*object.Regex)
			if ok {
				validationByRegex = true
			} else {
				fn = args[1]
				if !object.IsCallable(fn) {
					return object.NewError(object.ERR_ARGUMENTS, "read", "Expected callable or regex for second argument")
				}
			}

			if len(args) > 2 {
				e, ok := args[2].(*object.String)
				if !ok {
					return object.NewError(object.ERR_ARGUMENTS, "read", "Expected string for error message")
				}
				errMsg = e.String()

				if pr.Modes.ConsoleTextMode {
					errMsg = str.ReplaceNewLinesWithSystem(errMsg)
				}

				if len(args) > 3 {
					maxattempts, ok = object.NumberToInt(args[3])
					if !ok {
						return object.NewError(object.ERR_ARGUMENTS, "read", "Expected integer for maximum attempts")
					}

					if len(args) > 4 {
						// alternate return value instead of an exception
						alternate = args[4]
					}
				}
			}
		}
	}

	// parameters gathered...
	for i := 0; maxattempts == -1 || i < maxattempts; {
		fmt.Print(prompt)
		line, err := readLine(os.Stdin)
		if err != nil {
			return object.NewError(object.ERR_GENERAL, "read", err.Error())
		}

		if pr.Modes.ConsoleTextMode {
			line = str.ReplaceNewLinesWithLinux(line)
		}

		if validationByRegex || fn != nil {
			var verify object.Object

			if validationByRegex {
				verify, err = object.RegexMatching(re, line)
			} else {
				verify, err = pr.call(fn, object.NewString(line))
			}
			if err != nil {
				return object.NewError(object.ERR_GENERAL, "read", err.Error())
			}
			if verify == object.TRUE {
				return object.NewString(line)
			} else {
				fmt.Print(errMsg)
			}
		} else {
			return object.NewString(line)
		}

		if maxattempts > -1 {
			i++
		}
	}

	if alternate == nil {
		return object.NewError(object.ERR_GENERAL, "read", "Input failed to match expected")
	}
	return alternate
}

func readLine(in io.Reader) (string, error) {
	scanner := bufio.NewScanner(in)
	scanned := scanner.Scan()
	if !scanned {
		return "", fmt.Errorf("Unknown failure to scan input text")
	}
	return scanner.Text(), nil
}
