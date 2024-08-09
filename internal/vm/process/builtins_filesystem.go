// langur/vm/process/builtins_filesystem.go

package process

import (
	"langur/object"
	"os"
)

// cd, prop

var bi_cd = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:          "cd",
		Description:   "may change the current directory of the script; returns present working directory; has no effect on parent processes",
		ImpureEffects: true,

		// TODO: update
		ParamPositional: []object.Parameter{
			object.Parameter{},
		},
		ParamExpansionMax: 1,
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		const fnName = "cd"

		var err error
		var pwd string

		if len(args) != 0 {
			switch arg := args[0].(type) {
			case *object.String:
				err = os.Chdir(arg.String())
				if err != nil {
					return object.NewError(object.ERR_GENERAL, fnName, err.Error())
				}

			default:
				return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected string for new path")
			}
		}

		// return the present working directory
		pwd, err = os.Getwd()
		if err == nil {
			return object.NewString(pwd)
		}
		return object.NewError(object.ERR_GENERAL, fnName, err.Error())
	},
}

var bi_prop = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:          "prop",
		ImpureEffects: true,
		Description:   "returns hash of properties of file or directory if it exists at the given moment of execution; otherwise returns null",
		ParamPositional: []object.Parameter{
			object.Parameter{ExternalName: "file"},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		const fnName = "prop"

		var s string

		switch arg := args[0].(type) {
		case *object.String:
			s = arg.String()
		default:
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected string")
		}

		switch stat, err := os.Stat(s); {
		case err != nil:
			return object.NULL
		default:
			hashkv := make([]object.Object, 10)

			hashkv[0] = object.NewString("name")
			hashkv[1] = object.NewString(stat.Name())

			hashkv[2] = object.NewString("size")
			hashkv[3] = object.NumberFromInt64(stat.Size())

			hashkv[4] = object.NewString("isdir")
			hashkv[5] = object.NativeBoolToObject(stat.IsDir())

			hashkv[6] = object.NewString("perm")
			hashkv[7] = object.NumberFromInt(int(stat.Mode().Perm() & 0777))

			hashkv[8] = object.NewString("mod")
			hashkv[9] = object.NewDateTime(stat.ModTime(), true)

			hash, err := object.NewHashFromSlice(hashkv, false)
			if err != nil {
				return object.NewError(object.ERR_GENERAL, fnName, err.Error())
			}
			return hash
		}
	},
}
