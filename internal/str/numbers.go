// langur/str/numbers.go

package str

import (
	"fmt"
	"langur/cpoint"
	"langur/native"
	"strconv"
)

func IsBase(s string, base int) bool {
	for i, c := range s {
		if i == 0 && (c == '+' || c == '-') && len(s) > 1 {
			// continue
		} else {
			if !cpoint.IsDigitInBase(c, base) {
				return false
			}
		}
	}
	return true
}

func NumberWithBasePrefix(num string, base int) string {
	if len(num) == 0 {
		num = "0"
	} else if num[0] == '-' || num[0] == '+' {
		// sign in front of base
		return string(num[0]) + IntToStr(base, 10) + "x" + num[1:]
	}
	return IntToStr(base, 10) + "x" + num
}

func Int64ToStr(i int64, base int) string {
	return strconv.FormatInt(i, base)
}

func StrToInt64(s string, base int) (int64, error) {
	n64, err := strconv.ParseInt(s, base, 64)
	if err != nil {
		// change error message
		err = fmt.Errorf("Unable to convert string %q to int64", ReformatInput(s))
	}
	return n64, err
}

func StrToInt32(s string, base int) (int32, error) {
	n64, err := strconv.ParseInt(s, base, 32)
	if err != nil {
		// change error message
		err = fmt.Errorf("Unable to convert string %q to int32", ReformatInput(s))
	}
	return int32(n64), err
}

func IntToStr(i, base int) string {
	return strconv.FormatInt(int64(i), base)
}

func RuneToStr(r rune, base int) string {
	return strconv.FormatInt(int64(r), base)
}

func StrToInt(s string, base int) (int, error) {
	n64, err := strconv.ParseInt(s, base, 0)
	if err != nil {
		// change error message
		err = fmt.Errorf("Unable to convert string %q to int", ReformatInput(s))
	}
	return int(n64), err
}

func StrToRune(s string, base int) (rune, error) {
	n64, err := strconv.ParseInt(s, base, 64)

	var n rune
	var ok bool
	if err == nil {
		n, ok = native.ToRune(n64)
		if !ok {
			err = fmt.Errorf("Unable to string %q to rune (out of range)", n64)
		}

	} else {
		// change error message
		err = fmt.Errorf("Unable to convert string %q to rune", ReformatInput(s))
	}

	return n, err
}
