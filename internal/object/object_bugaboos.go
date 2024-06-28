// langur/object/object_bugaboos.go

package object

import (
	"fmt"
)

func bug(fnName, s string) {
	panic(s)
}

// This is not just used for bugs at the moment.
// This is used to convert a decimal package panic into an error.
func PanicToError(p interface{}) error {
	switch e := p.(type) {
	case error:
		return e
	case Object:
		return fmt.Errorf(e.String())
	case string:
		return fmt.Errorf(e)
	case fmt.Stringer:
		return fmt.Errorf(e.String())
	default:
		return fmt.Errorf("Unknown error type (%T)", e)
	}
}

func UnhandledPanicString(p interface{}) string {
	switch e := p.(type) {
	case Object:
		return fmt.Sprintf("Unhandled Panic: %s", e.String())
	case error:
		return fmt.Sprintf("Unhandled Panic: %s", e.Error())
	case string, fmt.Stringer:
		return fmt.Sprintf("Unhandled Panic: %s", e)
	default:
		return fmt.Sprintf("Unhandled Panic of Type %T", p)
	}
}
