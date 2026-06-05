// langur/interactive/main.go

// See copyright notice at langur/internal/main.go.
// See LICENSE.txt.

// allowing REPL to be used "locally" with special settings (for testing) or ...
// to be run from langur command as "interactive," with a more restricted set of possibilities

// NOTE: Go allows a package to be either executable or importable (not both).
// Use only one of the following package names (normally set to interactive, not main).
// for local REPL only, use...
// package main			/// executable

// for interactive mode (normal), use...
package interactive		/// importable

import (
	"langur/io"
	"bytes"
	"fmt"
	"io/ioutil"
	"langur/ast"
	"langur/bytecode"
	"langur/lexer"
	"langur/modes"
	"langur/object"
	"langur/opcode"
	"langur/parser"
	"langur/str"
	"langur/symbol"
	"langur/token"
	"langur/trace"
	"langur/vm"
	"strings"
)

type InteractiveOptions struct{
	Prompt string

	PrintCodeLocationTrace bool

	printLexTokens bool

	printParseTokenRepresentation bool
	printParseNodes               bool
	printParsedVarNames           bool

	printCompiledInstructions  bool
	printCompiledConstants     bool

	PrintVmResultEscaped    bool
	PrintVmResultGoEscaped  bool
	PrintVmResultRaw        bool

	PrintVmResultDescriptions bool
}

// NOTE: options for local REPL; may freely change them here for testing
// These are NOT applied to running from the langur command ("interactive"), ... 
// ... which will use a different set of options.
var options = &InteractiveOptions{
	Prompt : ">> ",

	PrintCodeLocationTrace: true,

	printLexTokens : false,

	printParseTokenRepresentation : false,
	printParseNodes               : false,
	printParsedVarNames           : false,

	printCompiledInstructions : false,
	printCompiledConstants    : false,

	PrintVmResultEscaped   : true,
	PrintVmResultGoEscaped : false,
	PrintVmResultRaw       : false,	
	PrintVmResultDescriptions: true,
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

func replPrintLn(s string) {
	io.PrintLn(s, vmModes.ConsoleTextMode)
}
func replPrint(s string) {
	io.Print(s, vmModes.ConsoleTextMode)
}
func replReadLn() (string, error) {
	return io.ReadLn(vmModes.ConsoleTextMode)
}

// for REPL not run from langur command (not "interactive" mode)
func main() {
	const loadFile = ""

	defer func() {
		if p := recover(); p != nil {
			replPrintLn(object.UnhandledPanicString(p))

			// NOTE: since not a command line REPL, okay to print a stack trace
			replPrint("Print stack trace? y/n: ")
			answer, _ := replReadLn()
			if answer == "y" || answer == "Y" {
				panic(p)
			} else {
				return
			}
		}
	}()

	resetEnvironment()

	if loadFile != "" {
		replPrintLn(fmt.Sprintf("loading file (%s)...", loadFile))
		b, err := ioutil.ReadFile(loadFile)

		if err == nil {
			repl(string(b), options)
		} else {
			replPrintLn(fmt.Sprintf("failed to load file: %s\n", err.Error()))
		}
		firstRun = false
	}

	loop(options)
}

// from langur command ("interactive")
func Interactive(opts *InteractiveOptions) {
	resetEnvironment()
	loop(opts)
}

func loop(opts *InteractiveOptions) {
	replPrintLn(fmt.Sprintf("langur %s (langurlang.org)\n", bytecode.LangurRev))
	replPrintLn("Type “exit()” or press ctrl-D to quit.")
	replPrintLn("Type “reset()” for a new environment.")

	for {
		replPrint(opts.Prompt)
		line, err := replReadLn()
		if err != nil {
			// don't print error; may be ctrl-D
			replPrintLn("")
			return
		}
		line = strings.TrimSpace(line)

		switch line {
		case "":
			continue

		case "exit":
			replPrintLn("Type exit() to quit.")
			continue

		case "exit()":
			// exit(): normally requires a parameter, but okay without for REPL
			return

		// FIXME: "reset" not a reserved keyword; therefore could potentially conflict with variable name
		case "reset":
			replPrintLn("Type reset() to reset the environment.")
			continue

		case "reset()":
			resetEnvironment()
			replPrintLn("Environment Reset")
			continue
		}

		repl(line, opts)
		firstRun = false
	}
}

func repl(source string, opts *InteractiveOptions) {
	var lex *lexer.Lexer
	var p *parser.Parser
	var program *ast.Program
	var comp *ast.Compiler
	var byteCode *bytecode.ByteCode
	var machine *vm.VM
	var err error

	var where *trace.Where

	defer func() {
		if err != nil && opts.PrintCodeLocationTrace {
			tr := trace.LocationTrace(where, source, "")
			if tr != "" {
				replPrint("\n" + tr)
			}
		}
	}()

	if opts.printLexTokens {
		// print lexical tokens
		lex, err = lexer.New(source, "RLPL", compileModes)
		if err == nil {
			replPrintLn("Tokens")
			for tok, err := lex.NextToken(); tok.Type != token.EOF; tok, err = lex.NextToken() {
				if err != nil {
					replPrintLn(err.Error())
					return
				}
				replPrintLn(fmt.Sprintf("%+v", tok.String()))
			}
		}
	}

	lex, err = lexer.New(source, "REPL", compileModes)
	if err != nil {
		replPrintLn(err.Error())
		return
	}

	if opts.printParseTokenRepresentation || opts.printParseNodes ||
		opts.printCompiledConstants || opts.printCompiledInstructions ||
		opts.PrintVmResultRaw || opts.PrintVmResultEscaped || opts.PrintVmResultGoEscaped {

		p = parser.New(lex, compileModes)
		program, err = p.ParseProgram()
		if err != nil {
			replPrintLn("Parser Error: " + err.Error())
		}

		if len(p.Errs) != 0 {
			replPrintLn("Parser Errors")
			for _, msg := range p.Errs {
				replPrintLn("\t"+msg.Error())
			}
		}
	}

	if opts.printParseTokenRepresentation {
		replPrintLn("Parsed Token Representation")
		replPrintLn(program.TokenRepresentation())
	}

	if opts.printParseNodes {
		replPrintLn("Nodes")
		replPrintLn(program.String())
	}

	if opts.printParsedVarNames {
		replPrintLn("Variable Names Used")
		for i := range program.VarNamesUsed {
			replPrintLn(program.VarNamesUsed[i])
		}
	}

	if p != nil && len(p.Errs) != 0 {
		return
	}

	if opts.printCompiledInstructions || opts.printCompiledConstants ||
		opts.PrintVmResultRaw || opts.PrintVmResultEscaped || opts.PrintVmResultGoEscaped {

		comp, err = ast.NewCompilerWithState(symbolTable, constants, compileModes, firstRun)
		if err != nil {
			replPrintLn("New Compiler Error: " + err.Error())

		} else {
			if firstRun {
				_, err = program.Compile(comp)
			} else {
				_, err = program.CompileAnother(comp)
			}
			if err != nil {
				replPrintLn("Compile Errors\n" + err.Error())
			}
	
			byteCode = comp.ByteCode()
			if opts.printCompiledInstructions {
				replPrintLn("ByteCode Instructions\n" +
					InstructionsString(byteCode.StartCode.InsPackage.Instructions, byteCode.Constants))
			}
			if opts.printCompiledConstants {
				replPrintLn("ByteCode Constants")
				for i := range byteCode.Constants {
					replPrintLn(fmt.Sprintf("%d: %s", i, byteCode.Constants[i].ReplString()))
				}
			}
		}

		if err != nil {
			return
		}

		constants = byteCode.Constants
	}

	if opts.PrintVmResultRaw || opts.PrintVmResultEscaped || opts.PrintVmResultGoEscaped {
		machine = vm.NewWithGlobalStore(byteCode, globals, vmModes)

		err, where = machine.Run()
		if err != nil {
			replPrintLn("VM Errors\n" + err.Error())
			return
		}
		result := machine.LastValue()

		vmModes = machine.LastModes() // so modes persist in the REPL

		if result == nil {
			replPrintLn("VM Result Nil (bug?)")
			return
		}
		if opts.PrintVmResultEscaped {
			if opts.PrintVmResultDescriptions {
				replPrint("langur escaped result: ")
			}
			replPrintLn(str.Escape(result.String()))
		}

		if opts.PrintVmResultGoEscaped {
			if opts.PrintVmResultDescriptions {
				replPrint("Go escaped result: ")
			}
			replPrintLn(str.EscapeGo(result.String()))
		}

		if opts.PrintVmResultRaw {
			if opts.PrintVmResultDescriptions {
				replPrint("raw result string: ")
			}
			replPrintLn(result.String())
		}
	}
}

// strings including type of constant
func InstructionsString(ins opcode.Instructions, constants []object.Object) string {
	var sb bytes.Buffer

	i := 0
	for i < len(ins) {
		deftypenum := ins[i]
		def, err := opcode.Lookup(deftypenum)
		if err != nil {
			fmt.Fprintf(&sb, "ERROR: %s\n", err)
			continue
		}

		operands, offset := opcode.ReadOperands(def, ins[i+1:])

		switch deftypenum {
		case opcode.OpConstant:
			// include the constant type string
			fmt.Fprintf(&sb, "%04d %s (%s)\n", i, ins.FmtInstruction(def, operands), constants[operands[0]].TypeString())
		default:
			fmt.Fprintf(&sb, "%04d %s\n", i, ins.FmtInstruction(def, operands))
		}

		i += 1 + offset
	}

	return sb.String()
}
