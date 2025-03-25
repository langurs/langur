// langur/repl/main.go
// langur REPL

// Copyright 2025 Anthony Davis
// See LICENSE file.
// This constitutes notice for all appropriate source files.

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"langur/ast"
	"langur/bytecode"
	"langur/compiler"
	"langur/lexer"
	"langur/modes"
	"langur/object"
	"langur/opcode"
	"langur/parser"
	"langur/str"
	"langur/symbol"
	"langur/token"
	"langur/vm"
	"os"
	"strings"
)

type replOptions struct{
	PROMPT string

	printLexTokens bool

	printParseTokenRepresentation bool
	printParseNodes               bool
	printParsedVarNames           bool

	printCompiledInstructions  bool
	printCompiledConstants     bool

	printVmResultEscaped    bool
	printVmResultGoEscaped  bool
	printVmResultRaw        bool
}

var options = &replOptions{
	PROMPT : ">> ",

	printLexTokens : false,

	printParseTokenRepresentation : false,
	printParseNodes               : false,
	printParsedVarNames           : false,

	printCompiledInstructions : false,
	printCompiledConstants    : false,

	printVmResultEscaped   : true,
	printVmResultGoEscaped : false,
	printVmResultRaw       : false,	
}

// with a 2-byte operand on OpGetGlobal and OpSetGlobal...
const GlobalStackMax = 65536

var (
	// for saving the environment in our REPL loop
	constants    []object.Object
	globals      []object.Object
	symbolTable  *symbol.SymbolTable
	vmModes      *modes.VmModes
	compileModes *modes.CompileModes
	firstRun 	 bool
)

func resetEnvironment() {
	constants = []object.Object{}
	globals = make([]object.Object, GlobalStackMax)
	symbolTable = symbol.NewSymbolTable(nil, modes.NewCompileModes())
	vmModes = modes.NewVmModes()
	compileModes = modes.NewCompileModes()
	firstRun = true
}

const loadFile = ""

func main() {
	REPL(os.Stdin, os.Stdout, options)
}

func readLine(in io.Reader, fixNewLines bool) string {
	scanner := bufio.NewScanner(in)
	scanned := scanner.Scan()
	if !scanned {
		panic("failed to scan input text")
	}
	text := scanner.Text()

	// allow input from plain text editor, which seems to insist on using Unicode line endings for copying even when no Unicode line endings present in the original text
	if fixNewLines {
		text = strings.Replace(text, "\u2029", "\n", -1)
		text = strings.Replace(text, "\u2028", "\n", -1)
	}

	return text
}

func REPL(in io.Reader, out io.Writer, opts *replOptions) {
	defer func() {
		if p := recover(); p != nil {
			fmt.Fprintf(out, object.UnhandledPanicString(p))
			fmt.Fprintln(out)

			// NOTE: since not a command line REPL (so far), okay to print a stack trace
			fmt.Fprintf(out, "Print stack trace? y/n: ")
			answer := readLine(in, false)
			if answer == "y" || answer == "Y" {
				panic(p)
			} else {
				return
			}
		}
	}()

	firstRun = true

	if loadFile != "" {
		fmt.Fprintf(out, "loading file (%s)...\n", loadFile)
		b, err := ioutil.ReadFile(loadFile)

		if err == nil {
			repl(string(b), out, true, opts)
		} else {
			fmt.Fprintf(out, "failed to load file: %s\n", err.Error())
		}
		firstRun = false
	}

	fmt.Printf("This is the REPL for langur %s (langurlang.org).\n", bytecode.LangurRev)
	fmt.Fprintf(out, "Type “exit()” to quit.\n")
	fmt.Fprintf(out, "Type “reset()” for a new environment.\n")

	resetEnvironment()

	for {
		fmt.Fprintf(out, opts.PROMPT)
		line := readLine(in, true)

		switch line {
		case "":
			continue

		case "exit":
			fmt.Fprintf(out, "Type exit() to quit.\n")
			continue

		case "exit()":
			// exit(): normally requires a parameter, but okay without for REPL
			return

		case "reset()":
			resetEnvironment()
			fmt.Fprintf(out, "Environment Reset\n")
			continue
		}

		repl(line, out, firstRun, opts)
		firstRun = false
	}
}

func repl(source string, out io.Writer, firstRun bool, opts *replOptions) {
	var lex *lexer.Lexer
	var p *parser.Parser
	var program *ast.Program
	var comp *compiler.Compiler
	var byteCode *bytecode.ByteCode
	var machine *vm.VM
	var err error

	if opts.printLexTokens {
		// print lexical tokens
		lex, err = lexer.New(source, "RLPL", compileModes)
		if err == nil {
			io.WriteString(out, "Tokens\n")
			for tok, err := lex.NextToken(); tok.Type != token.EOF; tok, err = lex.NextToken() {
				if err != nil {
					fmt.Printf(err.Error())
					return
				}
				fmt.Fprintf(out, "%+v\n", tok.String())
			}
		}
	}

	lex, err = lexer.New(source, "REPL", compileModes)
	if err != nil {
		fmt.Fprintf(out, err.Error())
		return
	}

	if opts.printParseTokenRepresentation || opts.printParseNodes ||
		opts.printCompiledConstants || opts.printCompiledInstructions ||
		opts.printVmResultRaw || opts.printVmResultEscaped || opts.printVmResultGoEscaped {

		p = parser.New(lex, compileModes)
		program, err = p.ParseProgram()
		if err != nil {
			io.WriteString(out, fmt.Sprintf("Parser Error: %s", err.Error()))
		}

		if len(p.Errs) != 0 {
			io.WriteString(out, "Parser Errors\n")
			for _, msg := range p.Errs {
				io.WriteString(out, "\t"+msg.Error()+"\n")
			}
		}
	}

	if opts.printParseTokenRepresentation {
		io.WriteString(out, "Parsed Token Representation\n")

		io.WriteString(out, program.TokenRepresentation())
		io.WriteString(out, "\n")
	}

	if opts.printParseNodes {
		io.WriteString(out, "Nodes\n")

		io.WriteString(out, program.String())
		io.WriteString(out, "\n")
	}

	if opts.printParsedVarNames {
		io.WriteString(out, "Variable Names Used\n")
		for i := range program.VarNamesUsed {
			io.WriteString(out, program.VarNamesUsed[i])
			io.WriteString(out, "\n")
		}
	}

	if p != nil && len(p.Errs) != 0 {
		return
	}

	if opts.printCompiledInstructions || opts.printCompiledConstants ||
		opts.printVmResultRaw || opts.printVmResultEscaped || opts.printVmResultGoEscaped {

		comp, err = compiler.NewWithState(symbolTable, constants, compileModes)
		if err != nil {
			io.WriteString(out, fmt.Sprintf("Compile Error: %s", err.Error()))
		}

		if firstRun {
			err = comp.Compile(program, true)
		} else {
			err = comp.CompileAnother(program)
		}
		if err != nil {
			fmt.Fprintf(out, "Compile Errors\n%s\n", err)
		}

		byteCode = comp.ByteCode()
		if opts.printCompiledInstructions {
			fmt.Fprintf(out, "ByteCode Instructions\n%s\n",
				InstructionsString(byteCode.StartCode.InsPackage.Instructions, byteCode.Constants))
		}
		if opts.printCompiledConstants {
			fmt.Fprintf(out, "ByteCode Constants\n")
			for i := range byteCode.Constants {
				fmt.Fprintf(out, "%d: %s\n", i, byteCode.Constants[i].ReplString())
			}
		}

		if err != nil {
			return
		}

		constants = byteCode.Constants
	}

	if opts.printVmResultRaw || opts.printVmResultEscaped || opts.printVmResultGoEscaped {
		machine = vm.NewWithGlobalStore(byteCode, globals, vmModes)

		err = machine.Run()
		if err != nil {
			fmt.Fprintf(out, "VM Errors\n%s\n", err)
			return
		}
		result := machine.LastValue()

		vmModes = machine.LastModes() // so modes persist in the REPL

		if result == nil {
			io.WriteString(out, "VM Result Nil (bug?)\n")
			return
		}
		if opts.printVmResultEscaped {
			io.WriteString(out, "langur escaped result: ")
			io.WriteString(out, str.Escape(result.String()))
			io.WriteString(out, "\n")
		}

		if opts.printVmResultGoEscaped {
			io.WriteString(out, "Go escaped result: ")
			io.WriteString(out, str.EscapeGo(result.String()))
			io.WriteString(out, "\n")
		}

		if opts.printVmResultRaw {
			io.WriteString(out, "raw result string: ")
			io.WriteString(out, result.String())
			io.WriteString(out, "\n")
		}
	}
}

// strings including type of constant
func InstructionsString(ins opcode.Instructions, constants []object.Object) string {
	var out bytes.Buffer

	i := 0
	for i < len(ins) {
		deftypenum := ins[i]
		def, err := opcode.Lookup(deftypenum)
		if err != nil {
			fmt.Fprintf(&out, "ERROR: %s\n", err)
			continue
		}

		operands, offset := opcode.ReadOperands(def, ins[i+1:])

		switch deftypenum {
		case opcode.OpConstant:
			// include the constant type string
			fmt.Fprintf(&out, "%04d %s (%s)\n", i, ins.FmtInstruction(def, operands), constants[operands[0]].TypeString())
		default:
			fmt.Fprintf(&out, "%04d %s\n", i, ins.FmtInstruction(def, operands))
		}

		i += 1 + offset
	}

	return out.String()
}
