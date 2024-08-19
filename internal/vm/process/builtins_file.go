// langur/vm/process/builtins_file.go

package process

import (
	"io/ioutil"
	"langur/object"
	"os"
)

// NOTE: These essentially deal with bytes as plain text only.
// Options may be added to make them more flexible, using optional parameters.

// readfile, writefile, appendfile

var bi_readfile = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:          "readfile",
		ImpureEffects: true,
		Description:   "reads text of file name given, returning a string",
		ParamPositional: []object.Parameter{
			object.Parameter{ExternalName: "file"},
		},
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
		ParamPositional: []object.Parameter{
			object.Parameter{ExternalName: "file"},
			object.Parameter{ExternalName: "contents"},
		},
		ParamByName: []object.Parameter{
			object.Parameter{ExternalName: "perm"},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		const fnName = "writefile"

		file, contents, perm := args[0], args[1], args[2]

		permissions := pr.Modes.NewFilePermissions

		filename, ok := file.(*object.String)
		if !ok {
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected string for file name and path")
		}
		sObj, ok := contents.(*object.String)
		if !ok {
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected string for contents for writing to file")
		}
		s := sObj.String()

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

		err := ioutil.WriteFile(filename.String(), []byte(s), permissions)
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
		Description:   "appends string to specified file name (or writes new file if it doesn't exist); new file permissions optional (default 664); permissions in form of 8x644 (NOT 0644, which would give the wrong number)",
		ParamPositional: []object.Parameter{
			object.Parameter{ExternalName: "file"},
			object.Parameter{ExternalName: "contents"},
		},
		ParamByName: []object.Parameter{
			object.Parameter{ExternalName: "perm"},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		const fnName = "appendfile"

		file, contents, perm := args[0], args[1], args[2]

		// Permissions only apply if creating a new file.
		permissions := pr.Modes.NewFilePermissions

		filename, ok := file.(*object.String)
		if !ok {
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected string for file name and path")
		}
		sObj, ok := contents.(*object.String)
		if !ok {
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected string for contents for appending to file")
		}
		s := sObj.String()

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

		f, err := os.OpenFile(filename.String(), os.O_APPEND|os.O_WRONLY|os.O_CREATE, permissions)
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
