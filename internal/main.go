// langur/main.go

// Copyright 2024 Anthony Davis
// See LICENSE file.
// This constitutes notice for all appropriate source files.

package main

import (
	"fmt"
	"io/ioutil"
	"langur/args"
	"langur/ast"
	"langur/bytecode"
	"langur/compiler"
	"langur/lexer"
	"langur/modes"
	"langur/object"
	"langur/parser"
	"langur/str"
	"langur/system"
	"langur/vm"
	"os"
)

const use = "use: langur [OPTION, ...] SCRIPT [SCRIPTARG, ...]"

func main() {
	printErrors := true

	// NOTE: printStackTrace should generally be false; might be abused otherwise?
	printStackTrace := false

	defer func() {
		if p := recover(); p != nil {
			if printErrors {
				fmt.Println(object.UnhandledPanicString(p))
				if printStackTrace {
					panic(p)
				}
			}
			os.Exit(system.GetExitStatus("failedrun"))
		}
	}()

	var compile_modes *modes.CompileModes = nil
	var vm_modes *modes.VmModes = nil

	// langur, langurArgs, file, fileArgs, err := args.OsArgsToArgs()
	_, langurArgs, file, _, err := args.OsArgsToArgs()
	if err != nil {
		fmt.Print("langur: ")
		fmt.Println(err)
		os.Exit(system.GetExitStatus("failedargs"))
	}

	compile_modes, err = modes.CompileModesFromArgs(langurArgs, system.OnWindows)
	if err != nil {
		fmt.Print("langur: ")
		fmt.Println(err)
		fmt.Println()
		fmt.Println(use)
		os.Exit(system.GetExitStatus("failedargs"))
	}

	if compile_modes.Help {
		fmt.Printf("langur %s (langurlang.org)\n\n %s\n%s",
			bytecode.LangurRev, use, args.GetArgsDescription())

		os.Exit(system.GetExitStatus("help"))
	}

	if file == "" {
		fmt.Printf("langur %s (langurlang.org)\n", bytecode.LangurRev)
		fmt.Println(use)
		os.Exit(system.GetExitStatus("noscript"))
	}

	scriptString := ""
	if compile_modes.ExecuteScriptStringInsteadOfFile {
		scriptString, file = file, "-e"

	} else {
		b, err := ioutil.ReadFile(file)
		if err != nil {
			if printErrors {
				s := str.Limit(file, 100, "...")
				fmt.Print("langur: ")
				fmt.Printf("error reading from file (%s): %s\n", s, err.Error())
			}
			os.Exit(system.GetExitStatus("failedreadfile"))
		}
		scriptString = string(b)
	}

	// Note: must check the parser and compiler for errors
	// Most lexer errors are passed to the parser, so they don't have to be checked here.
	lex, err := lexer.New(scriptString, file, compile_modes)
	if err != nil {
		fmt.Print("langur: ")
		fmt.Println("lexer error: " + err.Error())
		os.Exit(system.GetExitStatus("failedparse"))
	}
	p := parser.New(lex, compile_modes)

	var program *ast.Program
	program, err = p.ParseProgram()
	if err != nil {
		fmt.Print("langur: ")
		fmt.Println("parsing error: " + err.Error())
	}

	if len(p.Errs) != 0 {
		if printErrors {
			fmt.Print("langur: ")
			fmt.Println("parsing errors")
			for _, msg := range p.Errs {
				fmt.Printf("\t" + msg.Error() + "\n")
			}
		}
		os.Exit(system.GetExitStatus("failedparse"))
	}

	comp, err := compiler.New(compile_modes)
	if err != nil {
		fmt.Print("langur: ")
		fmt.Printf("compilation error: %s", err.Error())
	}

	err = comp.Compile(program, false)
	if err != nil {
		if printErrors {
			fmt.Print("langur: ")
			fmt.Printf("compilation errors\n%s\n", err)
		}
		os.Exit(system.GetExitStatus("failedcompile"))
	}

	if compile_modes.TestCompile {
		fmt.Print("langur: ")
		fmt.Println("no errors (parse and compile success)")
		os.Exit(system.GetExitStatus("test"))
	}

	byteCode := comp.ByteCode()
	machine := vm.New(byteCode, vm_modes)
	err = machine.Run()
	if err != nil {
		if printErrors {
			fmt.Print("langur: ")
			fmt.Printf("vm errors\n%s\n", err)
		}
		os.Exit(system.GetExitStatus("failedrun"))
	}

	os.Exit(0)
}
