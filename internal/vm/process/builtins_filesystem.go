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
		Description:   "current directory/change directory: may change the current directory of the script; returns present working directory; has no effect on parent processes",
		ImpureEffects: true,

		ParamByName: []object.Parameter{
			object.Parameter{ExternalName: "path", Type: object.STRING_OBJ},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		const fnName = "cd"

		var err error
		var pwd string

		pathObj := args[0]

		// nil okay; would be no change of directory
		if pathObj != nil {
			err = os.Chdir(pathObj.String())
			if err != nil {
				return object.NewError(object.ERR_GENERAL, fnName, err.Error())
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
		Description:   "returns hash of properties of path file or directory if it exists at the given moment of execution; otherwise returns null",
		ImpureEffects: true,

		ParamPositional: []object.Parameter{
			object.Parameter{ExternalName: "path", Type: object.STRING_OBJ},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		const fnName = "prop"

		path := args[0].String()

		switch stat, err := os.Stat(path); {
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
