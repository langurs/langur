// langur/args/args.go

package args

import (
	"langur/system"
	"os"
)

// NOTE: script may be file name/path or a script to execute (using the -e option)
func OsArgsToArgs() (
	langur string, langurArgs []string, script string, scriptArgs []string, err error) {

	if len(os.Args) > 0 {
		langur = os.Args[0]

		// args before script name go to langur / after the script name go to the script

		var optionMarker byte = '-'
		if system.OnWindows {
			optionMarker = '/'
		}

		scriptArgAt := -1
		for i := 1; i < len(os.Args); i++ {
			// anything starting with an option marker preceding script name taken as compiler argument
			if os.Args[i][0] != optionMarker {
				scriptArgAt = i
				break
			}
		}

		if len(os.Args) > 1 {
			if scriptArgAt == -1 {
				// no script name passed
				langurArgs = os.Args[1:]

			} else {
				script = os.Args[scriptArgAt]
				langurArgs = os.Args[1:scriptArgAt]

				if len(os.Args) > scriptArgAt+1 {
					scriptArgs = os.Args[scriptArgAt+1:]
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
 Script arguments are available inside a script via the _args variable (a list of strings).

 Command line options are as follows.
  /c           test parse and compile; do not execute

  /e "script"  execute command-line script instead of file

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
 Script arguments are available inside a script via the _args variable (a list of strings).

 Command line options are as follows.
  -c           test parse and compile; do not execute

  -e "script"  execute command-line script instead of file

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
