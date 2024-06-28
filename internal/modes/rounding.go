// /langur/modes/rounding.go

package modes

var RoundingMode = Default_Rounding

// NOTE: if adding new rounding modes, you must edit ...
// langur/compiler/bindings.go
// vm/process.(*Process).setMode()

const (
	Round_halfAwayFromZero = iota
	Round_halfEven
)

const RoundHashName = "_round"

var RoundHashModeNames = map[int]string{
	// These names should follow the rules for langur shorthand string indexing.
	Round_halfAwayFromZero: "halfawayfrom0",
	Round_halfEven:         "halfeven",
}
