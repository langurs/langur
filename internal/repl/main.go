// langur/repl/main.go
// langur REPL

// Copyright 2024 Anthony Davis
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
	"langur/vm/process"
	"os"
	"sort"
	"strings"
)

const (
	PROMPT = ">> "

	printLexTokens = false

	printParseTokenRepresentation = false
	printParseNodes               = false
	printParsedVarNames           = false

	printCompiledInstructions = true
	printCompiledConstants    = true

	printVmResultEscaped   = true
	printVmResultGoEscaped = false
	printVmResultRaw       = false
)

// with a 2-byte operand on OpGetGlobal and OpSetGlobal...
const GlobalStackMax = 65536

var (
	// for saving the environment in our REPL loop
	constants    []object.Object
	globals      []object.Object
	symbolTable  *symbol.SymbolTable
	vmModes      *modes.VmModes
	compileModes *modes.CompileModes
)

const loadFile = ""

func main() {
	fmt.Printf("This is the REPL for langur %s (langurlang.org).\n", bytecode.LangurRev)
	Start(os.Stdin, os.Stdout)
}

func readLine(in io.Reader) string {
	scanner := bufio.NewScanner(in)
	scanned := scanner.Scan()
	if !scanned {
		panic("failed to scan input text")
	}
	text := scanner.Text()

	// allow input from plain text editor, which seems to insist on using Unicode line endings for copying even when no Unicode line endings present in the original text
	text = strings.Replace(text, "\u2029", "\n", -1)
	text = strings.Replace(text, "\u2028", "\n", -1)

	return text
}

func Start(in io.Reader, out io.Writer) {
	defer func() {
		if p := recover(); p != nil {
			fmt.Fprintf(out, object.UnhandledPanicString(p))
			fmt.Fprintln(out)

			// NOTE: since not a command line REPL (so far), okay to print a stack trace
			fmt.Fprintf(out, "Print stack trace? y/n: ")
			answer := readLine(in)
			if answer == "y" || answer == "Y" {
				panic(p)
			} else {
				return
			}
		}
	}()

	firstRun := true

	resetEnvironment := func() {
		constants = []object.Object{}
		globals = make([]object.Object, GlobalStackMax)
		symbolTable = symbol.NewSymbolTable(nil, modes.NewCompileModes())
		firstRun = true
		vmModes = modes.NewVmModes()
		compileModes = modes.NewCompileModes()
	}

	if loadFile != "" {
		fmt.Fprintf(out, "loading file (%s)...\n", loadFile)
		b, err := ioutil.ReadFile(loadFile)

		if err == nil {
			repl(string(b), out, true)
		} else {
			fmt.Fprintf(out, "failed to load file: %s\n", err.Error())
		}
		firstRun = false
	}

	fmt.Fprintf(out, "Type “exit()” to quit.\n")
	fmt.Fprintf(out, "Type “reset()” for a new environment.\n")
	fmt.Fprintf(out, "Type “list()” to list built-in functions.\n")

	resetEnvironment()

	for {
		fmt.Fprintf(out, PROMPT)
		line := readLine(in)

		switch line {
		case "":
			continue

		case "exit":
			fmt.Fprintf(out, "Type exit() to quit.\n")
			continue

		case "exit()":
			// exit(): would work with this case, but leaving it for now
			return

		case "reset()":
			resetEnvironment()
			fmt.Fprintf(out, "Environment Reset\n")
			continue

		case "list()":
			var keys []string
			for _, k := range process.BuiltIns {
				// leave out internal built-ins
				if k.FnSignature.Name[0] != '_' {
					keys = append(keys, k.FnSignature.Name)
				}
			}
			sort.Strings(keys)
			fmt.Fprintf(out, "%d Built-in Functions\n", len(keys))
			for _, k := range keys {
				bi := process.GetBuiltInByName(k)
				fmt.Fprintf(out, " %s: %s\n", bi.FnSignature.Name, strings.Replace(bi.FnSignature.Description, "\n", "\n\t", -1))
			}
			continue
		}

		repl(line, out, firstRun)
		firstRun = false
	}
}

func repl(source string, out io.Writer, firstRun bool) {
	var lex *lexer.Lexer
	var p *parser.Parser
	var program *ast.Program
	var comp *compiler.Compiler
	var byteCode *bytecode.ByteCode
	var machine *vm.VM
	var err error

	if printLexTokens {
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

	if printParseTokenRepresentation || printParseNodes ||
		printCompiledConstants || printCompiledInstructions ||
		printVmResultRaw || printVmResultEscaped || printVmResultGoEscaped {

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

	if printParseTokenRepresentation {
		io.WriteString(out, "Parsed Token Representation\n")

		io.WriteString(out, program.TokenRepresentation())
		io.WriteString(out, "\n")
	}

	if printParseNodes {
		io.WriteString(out, "Nodes\n")

		io.WriteString(out, program.String())
		io.WriteString(out, "\n")
	}

	if printParsedVarNames {
		io.WriteString(out, "Variable Names Used\n")
		for i := range program.VarNamesUsed {
			io.WriteString(out, program.VarNamesUsed[i])
			io.WriteString(out, "\n")
		}
	}

	if p != nil && len(p.Errs) != 0 {
		return
	}

	if printCompiledInstructions || printCompiledConstants ||
		printVmResultRaw || printVmResultEscaped || printVmResultGoEscaped {

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
		if printCompiledInstructions {
			fmt.Fprintf(out, "ByteCode Instructions\n%s\n",
				InstructionsString(byteCode.StartCode.InsPackage.Instructions, byteCode.Constants))
		}
		if printCompiledConstants {
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

	if printVmResultRaw || printVmResultEscaped || printVmResultGoEscaped {
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
		if printVmResultEscaped {
			io.WriteString(out, "langur escaped result: ")
			io.WriteString(out, str.Escape(result.String()))
			io.WriteString(out, "\n")
		}

		if printVmResultGoEscaped {
			io.WriteString(out, "Go escaped result: ")
			io.WriteString(out, str.EscapeGo(result.String()))
			io.WriteString(out, "\n")
		}

		if printVmResultRaw {
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
