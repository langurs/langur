// langur/vm/process/loadmodule.go

package process

import (
	"fmt"
	// "langur/compile"
	"langur/object"
	"io/ioutil"
)

func (pr *Process) loadModule(module object.Object, as string) error {

	return fmt.Errorf("Load module feature not developed yet")

	// 1. locate the module
	// TODO: decide where/how imports will be found
	// for initial development, doing crude load; file has to be easily found

	// TODO:
	mod := module.String()

	// 2. read the file
	// for now, assuming UTF-8
	bSlc, err := ioutil.ReadFile(mod)
	if err != nil {
		return fmt.Errorf("Error reading import file %q: %s", mod, err.Error())
	}
	source := string(bSlc)

	_ = source

	// 3. attempt to parse and compile
	// byteCode, _, err := compile.ParseAndCompile(
	// 	source, mod, true, true, compile_modes)

	// 4. add base name to symbol table
	if as == "" {
		return fmt.Errorf("Missing alias name for import %q", mod)
	}


	// 5. find "exported" things (public)

	return nil
}
