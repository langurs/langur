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

var RegexNewLineString = "\x0D\x0A|[\x0D\x0A\u0085\u2028\u2029]"
var RegexNewLine = regexp.MustCompile(RegexNewLineString)

func ReplaceNewLinesWithSystem(s string) string {
	return RegexNewLine.ReplaceAllString(s, SysNewLine)
}

func ReplaceNewLinesWithLinux(s string) string {
	return RegexNewLine.ReplaceAllString(s, "\x0A")
}
