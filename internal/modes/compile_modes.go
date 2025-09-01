// langur/modes/compile_modes.go

package modes

import (
	"fmt"
	"langur/str"
)

type CompileModes struct {
	WarnOnIntegerLiteralsStartingWithZero bool
	ExecuteSourceStringInsteadOfFile      bool
	TestCompile                           bool
	WarnOnSurrogateCodes                  bool
	Help                                  bool
}

// Integer literals starting with 0 might be confused for base 8 numbers.
// Base 8 literals in langur start with 8x, such as 8x123, not 0123.
const Default_WarnOnIntegerLiteralsStartingWithZero = true

const Default_WarnOnSurrogateCodes = true

func NewCompileModes() *CompileModes {
	return &CompileModes{
		WarnOnIntegerLiteralsStartingWithZero: Default_WarnOnIntegerLiteralsStartingWithZero,
		WarnOnSurrogateCodes:                  Default_WarnOnSurrogateCodes,
	}
}

func (m *CompileModes) Copy() *CompileModes {
	return &CompileModes{
		WarnOnIntegerLiteralsStartingWithZero: m.WarnOnIntegerLiteralsStartingWithZero,
		WarnOnSurrogateCodes:                  m.WarnOnSurrogateCodes,
	}
}

// NOTE: Also update the argument descriptions (at langur/args.go).
func CompileModesFromArgs(args []string, useSlash bool) (m *CompileModes, err error) {
	m = NewCompileModes()

	for i, flag := range args {
		switch flag {
		case "-W0123", "/W0123":
			m.WarnOnIntegerLiteralsStartingWithZero = true

		case "-w0123", "/w0123":
			m.WarnOnIntegerLiteralsStartingWithZero = false

		case "-Wsurrogate", "/Wsurrogate":
			m.WarnOnSurrogateCodes = true

		case "-wsurrogate", "/wsurrogate":
			m.WarnOnSurrogateCodes = false

		case "-e", "/e":
			// execute command line script rather than file
			// must be last
			if i != len(args)-1 {
				err = fmt.Errorf("Execute flag must be last")
			}
			m.ExecuteSourceStringInsteadOfFile = true

		case "-c", "/c":
			m.TestCompile = true
			if m.Help {
				err = fmt.Errorf("Invalid mix of command line flags")
				return
			}

		case "--help", "/?":
			m.Help = true
			if m.TestCompile || m.ExecuteSourceStringInsteadOfFile {
				err = fmt.Errorf("Invalid mix of command line flags")
				return
			}

		default:
			err = fmt.Errorf(
				"Unexpected command line flag: %s",
				str.ReformatInput(flag))
			return
		}
	}
	return
}
