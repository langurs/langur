// langur/vm/process/builtins_string.go

package process

import (
	"langur/cpoint"
	"langur/object"
	"strings"

	"golang.org/x/text/unicode/norm"
)

// join
// lcase, tcase, ucase
// trim, ltrim, rtrim
// nfc, nfd, nfkc, nfkd
// also see builtins/regex

func bi_lcase(pr *Process, args ...object.Object) object.Object {
	switch arg := args[0].(type) {
	case *object.String:
		return object.NewString(strings.ToLower(arg.String()))

	case *object.Number:
		n, err := arg.ToRune()
		if err == nil {
			return object.NumberFromRune(cpoint.Lcase(n))
		}
		return object.NewError(object.ERR_ARGUMENTS, "lcase", "Integer outside expected range")
	}
	return object.NewError(object.ERR_ARGUMENTS, "lcase", "Expected string or integer")
}

func bi_tcase(pr *Process, args ...object.Object) object.Object {
	switch arg := args[0].(type) {
	case *object.String:
		return object.NewString(strings.ToTitle(arg.String()))

	case *object.Number:
		n, err := arg.ToRune()
		if err == nil {
			return object.NumberFromRune(cpoint.Tcase(n))
		}
		return object.NewError(object.ERR_ARGUMENTS, "tcase", "Integer outside expected range")
	}
	return object.NewError(object.ERR_ARGUMENTS, "tcase", "Expected string or integer")
}

func bi_ucase(pr *Process, args ...object.Object) object.Object {
	switch arg := args[0].(type) {
	case *object.String:
		return object.NewString(strings.ToUpper(arg.String()))

	case *object.Number:
		n, err := arg.ToRune()
		if err == nil {
			return object.NumberFromRune(cpoint.Ucase(n))
		}
		return object.NewError(object.ERR_ARGUMENTS, "ucase", "Integer outside expected range")
	}
	return object.NewError(object.ERR_ARGUMENTS, "ucase", "Expected string or integer")
}

func bi_trim(pr *Process, args ...object.Object) object.Object {
	s, ok := args[0].(*object.String)
	if !ok {
		return object.NewError(object.ERR_ARGUMENTS, "trim", "Expected string")
	}
	return object.NewString(strings.TrimFunc(s.String(), cpoint.IsTrimmable))
}

func bi_ltrim(pr *Process, args ...object.Object) object.Object {
	s, ok := args[0].(*object.String)
	if !ok {
		return object.NewError(object.ERR_ARGUMENTS, "ltrim", "Expected string")
	}
	return object.NewString(strings.TrimLeftFunc(s.String(), cpoint.IsTrimmable))
}

func bi_rtrim(pr *Process, args ...object.Object) object.Object {
	s, ok := args[0].(*object.String)
	if !ok {
		return object.NewError(object.ERR_ARGUMENTS, "rtrim", "Expected string")
	}
	return object.NewString(strings.TrimRightFunc(s.String(), cpoint.IsTrimmable))
}

func bi_join(pr *Process, args ...object.Object) object.Object {
	// delimiter defaulting to zls
	var delim string
	var arr *object.List
	var ok bool

	if len(args) == 1 {
		arr, ok = args[0].(*object.List)
		if !ok {
			return object.NewError(object.ERR_ARGUMENTS, "join", "Expected list of things to join")
		}

	} else {
		d, ok := args[0].(*object.String)
		if !ok {
			return object.NewError(object.ERR_ARGUMENTS, "join", "Expected string for delimiter")
		}
		delim = d.String()

		arr, ok = args[1].(*object.List)
		if !ok {
			return object.NewError(object.ERR_ARGUMENTS, "join", "Expected list of things to join")
		}
	}

	var sb strings.Builder
	for i, e := range arr.Elements {
		if i > 0 {
			// add delimiter
			sb.WriteString(delim)
		}

		s, err := object.AutoString(e)
		if err != nil {
			return object.NewError(object.ERR_ARGUMENTS, "join", err.Error())
		}
		sb.WriteString(s.String())
	}

	return object.NewString(sb.String())
}

// Unicode normalizations...

func bi_nfc(pr *Process, args ...object.Object) object.Object {
	strObj, ok := args[0].(*object.String)
	if !ok {
		return object.NewError(object.ERR_ARGUMENTS, "nfc", "Expected string")
	}
	return object.NewString(norm.NFC.String(strObj.String()))
}

func bi_nfd(pr *Process, args ...object.Object) object.Object {
	strObj, ok := args[0].(*object.String)
	if !ok {
		return object.NewError(object.ERR_ARGUMENTS, "nfd", "Expected string")
	}
	return object.NewString(norm.NFD.String(strObj.String()))
}

func bi_nfkc(pr *Process, args ...object.Object) object.Object {
	strObj, ok := args[0].(*object.String)
	if !ok {
		return object.NewError(object.ERR_ARGUMENTS, "nfkc", "Expected string")
	}
	return object.NewString(norm.NFKC.String(strObj.String()))
}

func bi_nfkd(pr *Process, args ...object.Object) object.Object {
	strObj, ok := args[0].(*object.String)
	if !ok {
		return object.NewError(object.ERR_ARGUMENTS, "nfkd", "Expected string")
	}
	return object.NewString(norm.NFKD.String(strObj.String()))
}
