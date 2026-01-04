// langur/main.go

// Copyright 2026 Anthony Davis
// See LICENSE.txt.
// This constitutes notice for all appropriate source files.

package main

import (
	"fmt"
	"io/ioutil"
	"langur/args"
	"langur/ast"
	"langur/bytecode"
	"langur/interactive"
	"langur/lexer"
	"langur/modes"
	"langur/object"
	"langur/parser"
	"langur/str"
	"langur/system"
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

func printErr(s string) {
	fmt.Fprint(os.Stderr, s)
}
func printLnErr(s string) {
	printErr(s + "\n")
}

func main() {
	var where *trace.Where

	defer func() {
		if p := recover(); p != nil {
			if printErrors {
				printLnErr(object.UnhandledPanicString(p))
				if printStackTrace {
					panic(p)
				}
			}
			os.Exit(system.GetExitStatus(system.ExitStatusFailedRun))
		}
	}()

	var compile_modes *modes.CompileModes = nil
	var vm_modes *modes.VmModes = nil

	// langur, langurArgs, file, fileArgs, err := args.OsArgsToArgs()
	_, langurArgs, file, _, err := args.OsArgsToArgs()
	if err != nil {
		printLnErr("langur: " + err.Error())
		os.Exit(system.GetExitStatus(system.ExitStatusFailedArgs))
	}

	compile_modes, err = modes.CompileModesFromArgs(langurArgs, system.OnWindows)
	if err != nil {
		printLnErr("langur: " + err.Error() + "\n\n" + use)
		os.Exit(system.GetExitStatus(system.ExitStatusFailedArgs))
	}

	if compile_modes.Help {
		printLnErr(fmt.Sprintf("langur %s (langurlang.org)\n\n %s\n%s",
			bytecode.LangurRev, use, args.GetArgsDescription()))

		os.Exit(system.GetExitStatus(system.ExitStatusHelp))
	}

	if file == "" {
		// interactive mode
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
				s := str.Limit(file, 100, "...")
				printLnErr(fmt.Sprintf("langur: error reading from file (%s): %s", s, err.Error()))
			}
			os.Exit(system.GetExitStatus(system.ExitStatusFailedReadFile))
		}
		source = string(b)
	}

	// Note: must check the parser and compiler for errors
	// Most lexer errors are passed to the parser, so they don't have to be checked here.
	lex, err := lexer.New(source, file, compile_modes)
	if err != nil {
		printLnErr("langur: lexer error: " + err.Error())
		os.Exit(system.GetExitStatus(system.ExitStatusFailedParse))
	}
	p := parser.New(lex, compile_modes)

	var program *ast.Program
	program, err = p.ParseProgram()
	if err != nil {
		printLnErr("langur: parsing error: " + err.Error())
	}

	if len(p.Errs) != 0 {
		if printErrors {
			printLnErr("langur: parsing errors")
			for _, msg := range p.Errs {
				printLnErr("\t" + msg.Error())
			}
		}
		os.Exit(system.GetExitStatus(system.ExitStatusFailedParse))
	}

	comp, err := ast.NewCompiler(compile_modes, true)
	if err != nil {
		if printErrors {
			printLnErr("langur: new compiler error: " + err.Error())

			if printCodeLocationTrace {
				tr := trace.LocationTrace(where, source, file)
				if tr != "" {
					printLnErr("\n" + tr)
				}
			}
		}
		os.Exit(system.GetExitStatus(system.ExitStatusFailedCompile))
	}

	_, err = program.Compile(comp)
	if err != nil {
		if printErrors {
			printLnErr("langur: compilation errors\n" + err.Error())

			if printCodeLocationTrace {
				tr := trace.LocationTrace(where, source, file)
				if tr != "" {
					printLnErr("\n" + tr)
				}
			}
		}
		os.Exit(system.GetExitStatus(system.ExitStatusFailedCompile))
	}

	if compile_modes.TestCompile {
		printLnErr("langur: no errors (parse and compile success)")
		os.Exit(system.GetExitStatus(system.ExitStatusTest))
	}

	byteCode := comp.ByteCode()
	machine := vm.New(byteCode, vm_modes)
	err, where = machine.Run()
	if err != nil {
		if printErrors {
			printLnErr("langur: vm errors\n" + err.Error())

			if printCodeLocationTrace {
				tr := trace.LocationTrace(where, source, file)
				if tr != "" {
					printLnErr("\n" + tr)
				}
			}
		}
		os.Exit(system.GetExitStatus(system.ExitStatusFailedRun))
	}

	os.Exit(0)
}
