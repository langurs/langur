// langur/vm/process/builtins_file.go

package process

import (
	"io/ioutil"
	"langur/object"
	"os"
)

// NOTE: These essentially deal with bytes as plain text only (UTF-8).
// Options may be added to make them more flexible, using optional parameters.

// readfile, writefile, appendfile

var bi_readfile = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:            "readfile",
		ImpureEffects:   true,
		Description:     "reads text of file name given, returning a string",
		ParamPositional: []object.Parameter{object.Parameter{ExternalName: "file"}},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		const fnName = "readfile"

		filename, ok := args[0].(*object.String)
		if !ok {
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected string for file name and path")
		}

		bSlc, err := ioutil.ReadFile(filename.String())
		if err != nil {
			return object.NewError(object.ERR_GENERAL, fnName, err.Error())
		}
		s := string(bSlc)

		return object.NewString(s)
	},
}

var bi_writefile = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:          "writefile",
		ImpureEffects: true,
		Description:   "writes string to specified file name; permissions optional (default 664); permissions in form of 8x644 (NOT 0644, which would give the wrong number)",

		// TODO: update
		ParamPositional: []object.Parameter{
			object.Parameter{},
		},
		ParamExpansionMin: 2,
		ParamExpansionMax: 3,
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		const fnName = "writefile"

		perm := pr.Modes.NewFilePermissions

		filename, ok := args[0].(*object.String)
		if !ok {
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected string for first argument for file name and path")
		}
		sObj, ok := args[1].(*object.String)
		if !ok {
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected string for second argument for writing to file")
		}
		s := sObj.String()

		if len(args) > 2 {
			p, ok := object.NumberToInt(args[2])
			if !ok {
				return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected integer for third argument (permissions); example: 8x664")
			}
			perm = os.FileMode(p)
		}

		err := ioutil.WriteFile(filename.String(), []byte(s), perm)
		if err != nil {
			return object.NewError(object.ERR_GENERAL, fnName, err.Error())
		}
		// returns number of bytes written
		return object.NumberFromInt(len(s))
	},
}

var bi_appendfile = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:          "appendfile",
		ImpureEffects: true,
		Description:   "appends string to specified file name (or writes new file if it doesn't exist); permissions optional (default 664); permissions in form of 8x644 (NOT 0644, which would give the wrong number)",

		// TODO: update
		ParamPositional: []object.Parameter{
			object.Parameter{},
		},
		ParamExpansionMin: 2,
		ParamExpansionMax: 3,
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		const fnName = "appendfile"

		perm := pr.Modes.NewFilePermissions

		filename, ok := args[0].(*object.String)
		if !ok {
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected string for first argument for file name and path")
		}
		sObj, ok := args[1].(*object.String)
		if !ok {
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected string for second argument for appending to file")
		}
		s := sObj.String()

		if len(args) > 2 {
			p, ok := object.NumberToInt(args[2])
			if !ok {
				return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected integer for third argument (permissions); example: 8x664")
			}
			perm = os.FileMode(p)
		}

		f, err := os.OpenFile(filename.String(), os.O_APPEND|os.O_WRONLY|os.O_CREATE, perm)
		defer f.Close()

		if err != nil {
			return object.NewError(object.ERR_GENERAL, fnName, err.Error())
		}

		if _, err = f.WriteString(s); err != nil {
			return object.NewError(object.ERR_GENERAL, fnName, err.Error())
		}

		// returns number of bytes written
		return object.NumberFromInt(len(s))
	},
}
