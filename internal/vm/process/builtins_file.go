// langur/vm/process/builtins_file.go

package process

import (
	"io/ioutil"
	"langur/object"
	"os"
)

// NOTE: These deal with bytes as plain text only and assuming UTF-8.
// Options may be added to make them more flexible, using optional parameters.

// readfile, writefile, appendfile

var bi_readfile = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:          "readfile",
		ImpureEffects: true,
		Description:   "reads text of file name given, returning a string",
		ParamPositional: []object.Parameter{
			object.Parameter{ExternalName: "file", Type: object.STRING_OBJ},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		const fnName = "readfile"

		filename := args[0].String()
		bSlc, err := ioutil.ReadFile(filename)
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
		ParamPositional: []object.Parameter{
			object.Parameter{ExternalName: "file", Type: object.STRING_OBJ},
			object.Parameter{ExternalName: "contents", Type: object.STRING_OBJ},
		},
		ParamByName: []object.Parameter{
			object.Parameter{ExternalName: "perm", Type: object.NUMBER_OBJ},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		const fnName = "writefile"

		file := args[0].String()
		contents := args[1].String()
		perm := args[2]

		permissions := pr.Modes.NewFilePermissions
		if perm != nil {
			p, ok := object.NumberToInt(perm)
			if !ok {
				return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected integer for permissions; example: 8x664 (NOT 0664, BTW)")
			}
			if p < 0 || p > 0777 {
				return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected integer for permissions in range of 0 to 8x777 (NOT 0777, BTW)")
			}
			permissions = os.FileMode(p)
		}

		err := ioutil.WriteFile(file, []byte(contents), permissions)
		if err != nil {
			return object.NewError(object.ERR_GENERAL, fnName, err.Error())
		}
		// returns number of bytes written
		return object.NumberFromInt(len(contents))
	},
}

var bi_appendfile = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:          "appendfile",
		ImpureEffects: true,
		Description:   "appends string to specified file name (or writes new file if it doesn't exist); new file permissions optional (default 664); permissions in form of 8x644 (NOT 0644, which would give the wrong number)",
		ParamPositional: []object.Parameter{
			object.Parameter{ExternalName: "file", Type: object.STRING_OBJ},
			object.Parameter{ExternalName: "contents", Type: object.STRING_OBJ},
		},
		ParamByName: []object.Parameter{
			object.Parameter{ExternalName: "perm", Type: object.NUMBER_OBJ},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		const fnName = "appendfile"

		file := args[0].String()
		contents := args[1].String()
		perm := args[2]

		// Permissions only apply if creating a new file.
		permissions := pr.Modes.NewFilePermissions
		if perm != nil {
			p, ok := object.NumberToInt(perm)
			if !ok {
				return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected integer for permissions; example: 8x664 (NOT 0664, BTW)")
			}
			if p < 0 || p > 0777 {
				return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected integer for permissions in range of 0 to 8x777 (NOT 0777, BTW)")
			}
			permissions = os.FileMode(p)
		}

		f, err := os.OpenFile(file, os.O_APPEND|os.O_WRONLY|os.O_CREATE, permissions)
		defer f.Close()

		if err != nil {
			return object.NewError(object.ERR_GENERAL, fnName, err.Error())
		}

		if _, err = f.WriteString(contents); err != nil {
			return object.NewError(object.ERR_GENERAL, fnName, err.Error())
		}

		// returns number of bytes written
		return object.NumberFromInt(len(contents))
	},
}
