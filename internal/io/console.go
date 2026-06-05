// langur/io/console.go

package io

import (
	"bufio"
	"fmt"
	"langur/str"
	"os"
)

func PrintErr(s string, replaceNewLines bool) {
	if replaceNewLines {
		s = str.ReplaceNewLinesWithSystem(s)
	}
	fmt.Fprint(os.Stderr, s)
}
func PrintLnErr(s string, replaceNewLines bool) {
	PrintErr(s + str.SysNewLine, replaceNewLines)
}

func Print(s string, replaceNewLines bool) {
	if replaceNewLines {
		s = str.ReplaceNewLinesWithSystem(s)
	}
	fmt.Fprint(os.Stdout, s)
}
func PrintLn(s string, replaceNewLines bool) {
	Print(s + str.SysNewLine, replaceNewLines)
}

func ReadLn(replaceNewLines bool) (s string, scanned bool) {
	scanner := bufio.NewScanner(os.Stdin)
	scanned = scanner.Scan()
	if !scanned {
		return
	}
	s = scanner.Text()
	if replaceNewLines {
		s = str.ReplaceNewLinesWithLinux(s)
	}
	return
}
