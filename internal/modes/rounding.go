// /langur/modes/rounding.go

package modes

import (
	"langur/decimal"
)

type RoundingMode int

// NOTE: if adding rounding modes, must also edit compiler/bindings.go
const (
	RoundHalfAwayFromZero RoundingMode = iota
	RoundHalfEven
)

const RoundHashName = "_round"

var RoundHashModeNames = map[RoundingMode]string{
	// These names should follow the rules for langur shorthand string indexing.
	RoundHalfAwayFromZero: "halfawayfrom0",
	RoundHalfEven:         "halfeven",
}

func LangurRoundingModeToDecimalRoundingMode(mode RoundingMode) decimal.RoundingMode {
	switch mode {
	case RoundHalfAwayFromZero:
		return decimal.RoundingMode_Default
	case RoundHalfEven:
		return decimal.RoundingMode_Bank
	default:
		bug("LangurRoundingModeToDecimalRoundingMode", "Unknown langur rounding mode")
		// uneccesary code here to satisfy the Go compiler...
		return decimal.RoundingMode_Default
	}
}

func bug(fnName, s string) {
	panic("Mode bug: " + s)
}
