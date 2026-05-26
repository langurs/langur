// langur/system/system.go

package system

import (
	"runtime"
)

var Type int

const (
	OTHER int = iota
	WINDOWS
	AMIGA
)

var OnWindows bool

func init() {
	// check and set this once
	switch runtime.GOOS {
	case "windows":
		Type = WINDOWS
		OnWindows = true

	case "amiga":
		Type = AMIGA

	default:
		Type = OTHER
	}
}
