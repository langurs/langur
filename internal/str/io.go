// langur/str/io.go

package str

import (
	"fmt"
	"os"
)

func PrintErr(s string, replaceNewLines bool) {
	if replaceNewLines {
		s = ReplaceNewLinesWithSystem(s)
	}
	fmt.Fprint(os.Stderr, s)
}
func PrintLnErr(s string, replaceNewLines bool) {
	PrintErr(s + SysNewLine, replaceNewLines)
}

func Print(s string, replaceNewLines bool) {
	if replaceNewLines {
		s = ReplaceNewLinesWithSystem(s)
	}
	fmt.Fprint(os.Stdout, s)
}
func PrintLn(s string, replaceNewLines bool) {
	Print(s + SysNewLine, replaceNewLines)
}
