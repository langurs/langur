// langur/interactive/main.go

// See copyright notice at langur/internal/main.go.

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
	"io"
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

func PrintLocationTrace(where *trace.Where, source, file string) {
	if where != nil {
		fmt.Printf("\ntraced to [%s] in file \"%s\"...\n", where.String(), file)
		fmt.Printf(where.Trace(source))
	}
}

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

var in, out = os.Stdin, os.Stdout

// for REPL not run from langur command (not "interactive" mode)
func main() {
	const loadFile = ""

	defer func() {
		if p := recover(); p != nil {
			fmt.Fprintf(out, object.UnhandledPanicString(p))
			fmt.Fprintln(out)

			// NOTE: since not a command line REPL, okay to print a stack trace
			fmt.Fprintf(out, "Print stack trace? y/n: ")
			answer, _ := readLine(in, false)
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
			repl(string(b), options)
		} else {
			fmt.Fprintf(out, "failed to load file: %s\n", err.Error())
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
	fmt.Fprintf(out, "Type “exit()” or press ctrl-D to quit.\n")
	fmt.Fprintf(out, "Type “reset()” for a new environment.\n")

	resetEnvironment()

	for {
		fmt.Fprintf(out, opts.Prompt)
		line, ok := readLine(in, true)
		if !ok {
			return
		}
		line = strings.TrimSpace(line)

		switch line {
		case "":
			continue

		case "exit":
			fmt.Fprintf(out, "Type exit() to quit.\n")
			continue

		case "exit()":
			// exit(): normally requires a parameter, but okay without for REPL
			return

		// FIXME: "reset" not a reserved keyword; therefore could potentially conflict with variable name
		case "reset":
			fmt.Fprintf(out, "Type reset() to reset the environment.\n")
			continue

		case "reset()":
			resetEnvironment()
			firstRun = true
			fmt.Fprintf(out, "Environment Reset\n")
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
			PrintLocationTrace(where, source, "")
		}
	}()

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
		opts.PrintVmResultRaw || opts.PrintVmResultEscaped || opts.PrintVmResultGoEscaped {

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
		opts.PrintVmResultRaw || opts.PrintVmResultEscaped || opts.PrintVmResultGoEscaped {

		comp, err = ast.NewCompilerWithState(symbolTable, constants, compileModes, firstRun)
		if err != nil {
			io.WriteString(out, fmt.Sprintf("New Compiler Error: %s", err.Error()))

		} else {
			if firstRun {
				_, err = program.Compile(comp)
			} else {
				_, err = program.CompileAnother(comp)
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
			fmt.Fprintf(out, "VM Errors\n%s\n", err)
			return
		}
		result := machine.LastValue()

		vmModes = machine.LastModes() // so modes persist in the REPL

		if result == nil {
			io.WriteString(out, "VM Result Nil (bug?)\n")
			return
		}
		if opts.PrintVmResultEscaped {
			if opts.PrintVmResultDescriptions {
				io.WriteString(out, "langur escaped result: ")
			}
			io.WriteString(out, str.Escape(result.String()))
			io.WriteString(out, "\n")
		}

		if opts.PrintVmResultGoEscaped {
			if opts.PrintVmResultDescriptions {
				io.WriteString(out, "Go escaped result: ")
			}
			io.WriteString(out, str.EscapeGo(result.String()))
			io.WriteString(out, "\n")
		}

		if opts.PrintVmResultRaw {
			if opts.PrintVmResultDescriptions {
				io.WriteString(out, "raw result string: ")
			}
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

func readLine(in io.Reader, fixNewLines bool) (text string, scanned bool) {
	scanner := bufio.NewScanner(in)
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
