// langur/main.go

// Copyright 2026 Anthony Davis
// See LICENSE.txt.
// This constitutes notice for all appropriate source files.

package main

import (
	"langur/vm/process"
	"fmt"
	"io/ioutil"
	"langur/args"
	"langur/bytecode"
	"langur/compile"
	"langur/interactive"
	"langur/modes"
	"langur/object"
	"langur/str"
	"langur/system"
	"langur/text_io"
	"langur/trace"
	"langur/vm"
	"os"
)

const (
	use = "use: langur [OPTION, ...] SCRIPT [SCRIPTARG, ...]"

	printErrors = true
	printCodeLocationTrace = true

	// NOTE: printStackTrace should generally be false; might be abused otherwise?
	printStackTrace = false
)

var consoleTextMode = true

func exitMain(status system.ExitStatus, s string) {
	if s != "" {
		text_io.PrintLnErr(s, consoleTextMode)
	}
	code := system.GetExitStatus(status)
	os.Exit(code)
}

func main() {
	var where *trace.Where
	var msg string

	defer func() {
		if p := recover(); p != nil {
			if printErrors {
				text_io.PrintLnErr(object.UnhandledPanicString(p), consoleTextMode)
				if printStackTrace {
					panic(p)
				}
			}
			exitMain(system.ExitStatusFailedRun, "")
		}
	}()

	var compile_modes *modes.CompileModes = nil
	var vm_modes *modes.VmModes = nil

	// langur, langurArgs, file, fileArgs, err := args.OsArgsToArgs()
	_, langurArgs, file, _, err := args.OsArgsToArgs()
	if err != nil {
		exitMain(system.ExitStatusFailedArgs, "langur: " + err.Error())
	}

	compile_modes, err = modes.CompileModesFromArgs(langurArgs, system.OnWindows)
	if err != nil {
		exitMain(system.ExitStatusFailedArgs, "langur: " + err.Error() + "\n\n" + use)
	}
	consoleTextMode = compile_modes.CompilerConsoleTextMode

	if compile_modes.Help {
		msg := fmt.Sprintf("langur %s (langurlang.org)\n\n %s\n%s",
			bytecode.LangurRev, use, args.GetArgsDescription())

		exitMain(system.ExitStatusHelp, msg)
	}

	if file == "" {
		// interactive mode -- invoked from the command line
		// These are not the REPL options. Edit those in the interactive package.
		// Interactive mode should not give away as much information as the REPL can.
		opts := &interactive.InteractiveOptions{
			Prompt: ">> ", PrintVmResultRaw: true,
			PrintCodeLocationTrace: printCodeLocationTrace,
		}
		interactive.Interactive(opts)
		os.Exit(0)
	}

	source := ""
	if compile_modes.ExecuteSourceStringInsteadOfFile {
		source, file = file, ""

	} else {
		b, err := ioutil.ReadFile(file)
		if err != nil {
			if printErrors {
				msg = str.Limit(file, 100, "...")
				msg = fmt.Sprintf("langur: error reading from file (%s): %s", msg, err.Error())
			}
			exitMain(system.ExitStatusFailedReadFile, msg)
		}
		source = string(b)
	}

	var byteCode *bytecode.ByteCode
	var exitStatus system.ExitStatus
	byteCode, exitStatus, err = compile.ParseAndCompile(
		source, file, true,	printCodeLocationTrace,	compile_modes, process.BuiltIns)

	if err != nil {
		msg := ""
		if printErrors {
			msg = "langur: " + err.Error()
		}
		exitMain(exitStatus, msg)
	}

	if compile_modes.TestCompile {
		exitMain(system.ExitStatusTest, "langur: no errors (parse and compile success)")
	}

	machine := vm.New(byteCode, vm_modes)
	err, where = machine.Run()
	if err != nil {
		if printErrors {
			msg = "langur: vm errors\n" + err.Error()

			if printCodeLocationTrace {
				tr := trace.LocationTrace(where, source, file)
				if tr != "" {
					msg += "\n" + tr
				}
			}
		}
		exitMain(system.ExitStatusFailedRun, msg)
	}

	object.Exit(machine.LastValue(), nil, consoleTextMode)
}
