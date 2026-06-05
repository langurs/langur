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

func ReadLn(replaceNewLines bool) (string, error) {
	scanner := bufio.NewScanner(os.Stdin)
	scanned := scanner.Scan()
	if !scanned {
		return "", fmt.Errorf("Unknown failure to scan input text")
	}
	text := scanner.Text()
	if replaceNewLines {
		text = str.ReplaceNewLinesWithLinux(text)
	}
	return text, nil
}
