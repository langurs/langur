// langur/vm/process/loadmodule.go

package process

import (
	"io/ioutil"
	"langur/regexp"
)

var fileNameExtRegex = regexp.MustCompile("\\.langur$")
var lastPartRegex = regexp.MustCompile("[~/]+$")

func (pr *Process) loadModule(mod, as string) error {
	// 1. locate the module
	// TODO: decide where/how imports will be found
	// for initial development, doing crude load; file has to be easily found

	// 2. read the file
	// for now, assuming UTF-8
	bSlc, err := ioutil.ReadFile(mod)
	if err != nil {
		return err
	}
	code := string(bSlc)

	_ = code

	// 3. attempt to parse and compile

	// 4. add base name to symbol table
	// if as alias not specified, must decipher from file
	if as == "" {
		as = mod
		if idx := fileNameExtRegex.FindStringIndex(as); idx != nil {
			as = as[:idx[0]]
		}
		if idx := lastPartRegex.FindStringIndex(as); idx != nil {
			as = as[idx[0]:]
		}
	}



	// 5. find "exported" things (public)

	return nil
}
