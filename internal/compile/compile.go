// langur/compile/compile.go

// lex, parse, and compile

package compile

import (
	"fmt"
	"langur/ast"
	"langur/bytecode"
	"langur/lexer"
	"langur/modes"
	"langur/object"
	"langur/parser"
	"langur/system"
	"langur/trace"
)

func ParseAndCompile(
	source, file string,
	runRemotely bool,
	printCodeLocationTrace bool,
	compile_modes *modes.CompileModes,
	builtins []*object.BuiltIn) (

	byteCode *bytecode.ByteCode,
	exitStatus system.ExitStatus, err error) {

	var where *trace.Where
	var msg string
	var lex *lexer.Lexer

	lex, err = lexer.New(source, file, compile_modes)
	if err != nil {
		exitStatus = system.ExitStatusFailedParse
		err = fmt.Errorf("lexer error: " + err.Error())
		return
	}
	p := parser.New(lex, compile_modes)

	var program *ast.Program
	program, err = p.ParseProgram()
	if err != nil {
		msg = "parsing error: " + err.Error()
		if len(p.Errs) != 0 {
			msg += "\n\nparsing errors..."
			for _, msg2 := range p.Errs {
				msg += "\n\t" + msg2.Error()
			}
		}
		exitStatus = system.ExitStatusFailedParse
		err = fmt.Errorf(msg)
		return
	}

	comp, err := ast.NewCompiler(compile_modes, builtins, true)
	comp.RunRemotely = runRemotely // not interactive mode/REPL or a test
	if err != nil {
		msg = "new compiler error: " + err.Error()
		if printCodeLocationTrace {
			tr := trace.LocationTrace(where, source, file)
			if tr != "" {
				msg += "\n" + tr
			}
		}
		exitStatus = system.ExitStatusFailedCompile
		err = fmt.Errorf(msg)
		return
	}

	_, err = program.Compile(comp)
	if err != nil {
		msg = "compilation errors\n" + err.Error()
		if printCodeLocationTrace {
			tr := trace.LocationTrace(where, source, file)
			if tr != "" {
				msg += "\n" + tr
			}
		}
		exitStatus = system.ExitStatusFailedCompile
		err = fmt.Errorf(msg)
		return
	}

	byteCode = comp.ByteCode()
	return
}
