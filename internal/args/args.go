// langur/args/args.go

package args

import (
	"langur/system"
	"os"
)

// NOTE: file may be file name/path or code to execute (using the -e option)
func OsArgsToArgs() (
	langur string, langurArgs []string, file string, fileArgs []string, err error) {

	if len(os.Args) > 0 {
		langur = os.Args[0]

		// args before filename go to langur / after the filename go to the script

		var optionMarker byte = '-'
		if system.OnWindows {
			optionMarker = '/'
		}

		fileArgAt := -1
		for i := 1; i < len(os.Args); i++ {
			// anything starting with an option marker preceding filename taken as compiler argument
			if os.Args[i][0] != optionMarker {
				fileArgAt = i
				break
			}
		}

		if len(os.Args) > 1 {
			if fileArgAt == -1 {
				// no filename passed
				langurArgs = os.Args[1:]

			} else {
				file = os.Args[fileArgAt]
				langurArgs = os.Args[1:fileArgAt]

				if len(os.Args) > fileArgAt+1 {
					fileArgs = os.Args[fileArgAt+1:]
				}
			}
		}
	}
	return
}

func GetArgsDescription() string {
	if system.OnWindows {
		return WindowsArgsDescription
	}
	return LinuxArgsDescription
}

var WindowsArgsDescription = `
 File arguments are available inside a script via the _args variable (a list of strings).

 Command line options are as follows.
  /c           test parse and compile; do not execute

  /e "code"    execute command-line code instead of file

  /?           print command-line options

  /W0123       warn on number literals starting with 0 ...
   ... so they're not confused for base 8 literals ...
   ... as used in many languages

  /w0123       disable this warning

   In langur, a base 8 literal is prefixed with 8x, ...
   ... such as 8x123 (NOT 0123).

  /Wsurrogate  warn on surrogate codes used as escape codes

  /wsurrogate  disable this warning

`

var LinuxArgsDescription = `
 File arguments are available inside a script via the _args variable (a list of strings).

 Command line options are as follows.
  -c           test parse and compile; do not execute

  -e "code"    execute command-line code instead of file

  --help       print command-line options

  -W0123       warn on number literals starting with 0 ...
   ... so they're not confused for base 8 literals ...
   ... as used in many languages

  -w0123       disable this warning

   In langur, a base 8 literal is prefixed with 8x, ...
   ... such as 8x123 (NOT 0123).

  -Wsurrogate  warn on surrogate codes used as escape codes

  -wsurrogate  disable this warning

`
