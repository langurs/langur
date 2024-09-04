// langur/vm/process/builtins_random.go

package process

import (
	crand "crypto/rand"
	"langur/object"
	"math/big"
)

// random

var bi_random = &object.BuiltIn{
	FnSignature: &object.Signature{
		Name:        "random",
		Description: "returns random integer from a given range or item from a list, hash, or string",

		ParamPositional: []object.Parameter{
			object.Parameter{ExternalName: "over"},
		},
	},
	Fn: func(pr *Process, args ...object.Object) object.Object {
		const fnName = "random"

		var start, end int64
		var err error

		switch n := args[0].(type) {
		case *object.Range:
			switch e := n.Start.(type) {
			case *object.Number:
				start, err = e.ToInt64()
				if err != nil {
					return object.NewError(object.ERR_ARGUMENTS, fnName, "Integer range failure")
				}
			default:
				return object.NewError(object.ERR_ARGUMENTS, fnName, "Integer range failure")
			}

			switch e := n.End.(type) {
			case *object.Number:
				end, err = e.ToInt64()
				if err != nil {
					return object.NewError(object.ERR_ARGUMENTS, fnName, "Integer range failure")
				}
			default:
				return object.NewError(object.ERR_ARGUMENTS, fnName, "Integer range failure")
			}

		case *object.Number:
			end, err = n.ToInt64()
			if err != nil {
				return object.NewError(object.ERR_ARGUMENTS, fnName, "Integer range failure")
			}

			// Given a single number, we start with 1 (implied), just as langur uses 1-based indexing.

			if end == 0 {
				return object.NumberFromInt(0)
			} else if end < 0 {
				start = -1
			} else {
				start = 1
			}

		case *object.List:
			// return 1 element of list at random
			if len(n.Elements) == 0 {
				return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected non-empty list")
			}
			e, err := randomCryptoInteger(0, int64(len(n.Elements)))
			if err != nil {
				return object.NewError(object.ERR_GENERAL, fnName, err.Error())
			}
			return n.Elements[e]

		case *object.String:
			// return 1 code point of string at random
			codePoints := n.RuneSlc()
			if len(codePoints) == 0 {
				return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected non-zero-length string")
			}
			with, err := randomCryptoInteger(0, int64(len(codePoints)))
			if err != nil {
				return object.NewError(object.ERR_GENERAL, fnName, err.Error())
			}
			return object.NumberFromRune(codePoints[with])

		case *object.Hash:
			// return 1 element of hash at random
			if len(n.Pairs) == 0 {
				return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected non-empty hash")
			}
			values := []object.Object{}
			for _, kv := range n.Pairs {
				values = append(values, kv.Value)
			}
			e, err := randomCryptoInteger(0, int64(len(values)))
			if err != nil {
				return object.NewError(object.ERR_GENERAL, fnName, err.Error())
			}
			return values[e]

		default:
			return object.NewError(object.ERR_ARGUMENTS, fnName, "Expected integer or integer range argument, or non-empty list, hash, or string")
		}

		if start > end {
			start, end = end, start
		}

		result, err := randomCryptoInteger(start, end+1)
		if err != nil {
			return object.NewError(object.ERR_GENERAL, fnName, err.Error())
		}

		return object.NumberFromInt64(result)
	},
}

// using crypto/rand
// end: exclusive
func randomCryptoInteger(start, end int64) (int64, error) {
	resultBig, err := crand.Int(crand.Reader, big.NewInt(end-start))
	if err != nil {
		return 0, err
	}
	return start + resultBig.Int64(), nil
}
