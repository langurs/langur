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
	"bufio"
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
	"os"
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
}

// for REPL not run from langur command (not "interactive" mode)
func main() {
	const loadFile = ""

	defer func() {
		if p := recover(); p != nil {
			fmt.Println(object.UnhandledPanicString(p))

			// NOTE: since not a command line REPL, okay to print a stack trace
			fmt.Print("Print stack trace? y/n: ")
			answer, _ := readLine(false)
			if answer == "y" || answer == "Y" {
				panic(p)
			} else {
				return
			}
		}
	}()

	firstRun = true

	if loadFile != "" {
		fmt.Printf("loading file (%s)...\n", loadFile)
		b, err := ioutil.ReadFile(loadFile)

		if err == nil {
			repl(string(b), options)
		} else {
			fmt.Printf("failed to load file: %s\n", err.Error())
		}
		firstRun = false
	}

	loop(options)
}

// from langur command ("interactive")
func Interactive(opts *InteractiveOptions) {
	firstRun = true
	loop(opts)
}

func loop(opts *InteractiveOptions) {
	fmt.Printf("langur %s (langurlang.org)\n", bytecode.LangurRev)
	fmt.Println("Type “exit()” or press ctrl-D to quit.")
	fmt.Println("Type “reset()” for a new environment.")

	resetEnvironment()

	for {
		fmt.Print(opts.Prompt)
		line, ok := readLine(true)
		if !ok {
			return
		}
		line = strings.TrimSpace(line)

		switch line {
		case "":
			continue

		case "exit":
			fmt.Print("Type exit() to quit.\n")
			continue

		case "exit()":
			// exit(): normally requires a parameter, but okay without for REPL
			return

		// FIXME: "reset" not a reserved keyword; therefore could potentially conflict with variable name
		case "reset":
			fmt.Print("Type reset() to reset the environment.\n")
			continue

		case "reset()":
			resetEnvironment()
			firstRun = true
			fmt.Print("Environment Reset\n")
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
				fmt.Printf("\n" + tr)
			}
		}
	}()

	if opts.printLexTokens {
		// print lexical tokens
		lex, err = lexer.New(source, "RLPL", compileModes)
		if err == nil {
			fmt.Println("Tokens")
			for tok, err := lex.NextToken(); tok.Type != token.EOF; tok, err = lex.NextToken() {
				if err != nil {
					fmt.Print(err.Error())
					return
				}
				fmt.Printf("%+v\n", tok.String())
			}
		}
	}

	lex, err = lexer.New(source, "REPL", compileModes)
	if err != nil {
		fmt.Print(err.Error())
		return
	}

	if opts.printParseTokenRepresentation || opts.printParseNodes ||
		opts.printCompiledConstants || opts.printCompiledInstructions ||
		opts.PrintVmResultRaw || opts.PrintVmResultEscaped || opts.PrintVmResultGoEscaped {

		p = parser.New(lex, compileModes)
		program, err = p.ParseProgram()
		if err != nil {
			fmt.Printf("Parser Error: %s", err.Error())
		}

		if len(p.Errs) != 0 {
			fmt.Println("Parser Errors")
			for _, msg := range p.Errs {
				fmt.Println("\t"+msg.Error())
			}
		}
	}

	if opts.printParseTokenRepresentation {
		fmt.Println("Parsed Token Representation")
		fmt.Println(program.TokenRepresentation())
	}

	if opts.printParseNodes {
		fmt.Println("Nodes")
		fmt.Println(program.String())
	}

	if opts.printParsedVarNames {
		fmt.Println("Variable Names Used")
		for i := range program.VarNamesUsed {
			fmt.Println(program.VarNamesUsed[i])
		}
	}

	if p != nil && len(p.Errs) != 0 {
		return
	}

	if opts.printCompiledInstructions || opts.printCompiledConstants ||
		opts.PrintVmResultRaw || opts.PrintVmResultEscaped || opts.PrintVmResultGoEscaped {

		comp, err = ast.NewCompilerWithState(symbolTable, constants, compileModes, firstRun)
		if err != nil {
			fmt.Print(fmt.Sprintf("New Compiler Error: %s", err.Error()))

		} else {
			if firstRun {
				_, err = program.Compile(comp)
			} else {
				_, err = program.CompileAnother(comp)
			}
			if err != nil {
				fmt.Printf("Compile Errors\n%s\n", err)
			}
	
			byteCode = comp.ByteCode()
			if opts.printCompiledInstructions {
				fmt.Printf("ByteCode Instructions\n%s\n",
					InstructionsString(byteCode.StartCode.InsPackage.Instructions, byteCode.Constants))
			}
			if opts.printCompiledConstants {
				fmt.Println("ByteCode Constants")
				for i := range byteCode.Constants {
					fmt.Printf("%d: %s\n", i, byteCode.Constants[i].ReplString())
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
			fmt.Printf("VM Errors\n%s\n", err)
			return
		}
		result := machine.LastValue()

		vmModes = machine.LastModes() // so modes persist in the REPL

		if result == nil {
			fmt.Println("VM Result Nil (bug?)")
			return
		}
		if opts.PrintVmResultEscaped {
			if opts.PrintVmResultDescriptions {
				fmt.Print("langur escaped result: ")
			}
			fmt.Println(str.Escape(result.String()))
		}

		if opts.PrintVmResultGoEscaped {
			if opts.PrintVmResultDescriptions {
				fmt.Print("Go escaped result: ")
			}
			fmt.Println(str.EscapeGo(result.String()))
		}

		if opts.PrintVmResultRaw {
			if opts.PrintVmResultDescriptions {
				fmt.Print("raw result string: ")
			}
			fmt.Println(result.String())
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

func readLine(fixNewLines bool) (text string, scanned bool) {
	scanner := bufio.NewScanner(os.Stdin)
	scanned = scanner.Scan()
	if !scanned {
		return
	}
	text = scanner.Text()

	// allow input from plain text editor, ...
	// ... which seems to insist on using Unicode line endings for copying ...
	// ... even when no Unicode line endings present in the original text
	if fixNewLines {
		text = strings.Replace(text, "\u2029", "\n", -1)
		text = strings.Replace(text, "\u2028", "\n", -1)
	}

	return
}
