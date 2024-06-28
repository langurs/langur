// langur/system/system.go

package system

import (
	"runtime"
)

var Type int

const (
	OTHER int = iota
	WINDOWS
)

func init() {
	// check and set this once
	if runtime.GOOS == "windows" {
		Type = WINDOWS
		OnWindows = true
	} else {
		Type = OTHER
	}
}

var OnWindows bool
