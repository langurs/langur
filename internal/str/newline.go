// langur/str/newline.go

package str

import (
	"langur/regexp"
	"langur/system"
)

var SysNewLine string = "\n"

func init() {
	if system.Type == system.WINDOWS {
		SysNewLine = "\r\n"
	}
}

var PatternNewLineString = "\x0D\x0A|[\x0D\x0A\u0085\u2028\u2029]"
var PatternNewLine = regexp.MustCompile(PatternNewLineString)

func ReplaceNewLinesWithSystem(s string) string {
	return PatternNewLine.ReplaceAllString(s, SysNewLine)
}

func ReplaceNewLinesWithLinux(s string) string {
	return PatternNewLine.ReplaceAllString(s, "\x0A")
}
