// langur/io/console.go

package io

import (
	"bufio"
	"fmt"
	"io"
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

func ReadLine(in io.Reader) (string, error) {
	scanner := bufio.NewScanner(in)
	scanned := scanner.Scan()
	if !scanned {
		return "", fmt.Errorf("Unknown failure to scan input text")
	}
	return scanner.Text(), nil
}
