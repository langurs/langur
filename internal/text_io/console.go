// langur/text_io/console.go

package text_io

import (
	"io"
	"bufio"
	"fmt"
	"langur/str"
	"os"
	"unicode/utf8"
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

func ReadBytes(byteCount int, replaceNewLines bool) (s string, scanned bool) {
	buf := make([]byte, byteCount)
	_, err := io.ReadFull(os.Stdin, buf)
	scanned = err == nil
	s = string(buf)
	if replaceNewLines {
		s = str.ReplaceNewLinesWithLinux(s)
	}
	return
}

func ReadCodePoints(cpCount int, replaceNewLines bool) (s string, err error) {
	// Last we knew, the Unicode standard wouldn't have more than 4 bytes in a UTF-8 sequence.
	byteMax := 4

	r := bufio.NewReader(os.Stdin)

	var out []byte
	var cp []byte
	var b byte

	count := 0
	for count < cpCount {
		// get 1 byte
		b, err = r.ReadByte()
		if err != nil {
			break
		}

		// append to code point slice
		cp = append(cp, b)
		if utf8.FullRune(cp) {
			// valid code code point; append to total slice and count
			out = append(out, cp...)
			cp = nil
			count++
		}

		if len(cp) > byteMax {
			// b/c of potential errors in input, error out here
			err = fmt.Errorf("Too many bytes in UTF-8 sequence that do not create a single code point")
			break
		}
	}

	s = string(out)
	if replaceNewLines {
		s = str.ReplaceNewLinesWithLinux(s)
	}

	return
}
