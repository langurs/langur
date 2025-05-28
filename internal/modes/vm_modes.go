// langur/modes/vm_modes.go

package modes

import (
	"os"
)

// NOTE: with concurrency, should be per process, with new process inheriting settings (copy) from calling context

type ModeNumber = int

// modes settable from source code
const (
	// numbers to use as opcode operands
	// NOTE: Currently, these must be handled individually in the vm/process setMode() method.
	MODE_DIVISION_MAX_SCALE ModeNumber = iota
	MODE_NEW_FILE_PERMISSIONS
	MODE_CONSOLE_TEXT_MODE
	MODE_ROUNDING
)

var ModeNames = map[string]ModeNumber{
	"divMaxScale":     MODE_DIVISION_MAX_SCALE,
	"consoleText":     MODE_CONSOLE_TEXT_MODE,
	"newFilePerm":     MODE_NEW_FILE_PERMISSIONS,
	"rounding":        MODE_ROUNDING,
}

const Default_DivisionMaxScale = 33
const Default_ConsoleTextMode = false
const Default_NewFilePerm os.FileMode = 0664 // in langur, 8x664
const Default_Rounding = RoundHalfAwayFromZero

var DefaultSubLexString = map[ModeNumber]string{
	MODE_DIVISION_MAX_SCALE:   "33",
	MODE_CONSOLE_TEXT_MODE:    "false",
	MODE_NEW_FILE_PERMISSIONS: "8x664",
	MODE_ROUNDING:             RoundHashName + "'" + RoundHashModeNames[Default_Rounding],
}

// modes not settable from source code so far
// NOTE: if Default_GoPanicToLangurException set to false for debugging, ...
// ... should be set back to true after
const Default_GoPanicToLangurException = true

type VmModes struct {
	NewFilePermissions       os.FileMode
	GoPanicToLangurException bool
	ConsoleTextMode          bool
	DivisionMaxScale         int
	Rounding                 RoundingMode
}

func NewVmModes() *VmModes {
	return &VmModes{
		NewFilePermissions:       Default_NewFilePerm,
		GoPanicToLangurException: Default_GoPanicToLangurException,
		ConsoleTextMode:          Default_ConsoleTextMode,
		DivisionMaxScale:         Default_DivisionMaxScale,
		Rounding:                 Default_Rounding,
	}
}

func (m *VmModes) Copy() *VmModes {
	return &VmModes{
		NewFilePermissions:       m.NewFilePermissions,
		GoPanicToLangurException: m.GoPanicToLangurException,
		ConsoleTextMode:          m.ConsoleTextMode,
		DivisionMaxScale:         m.DivisionMaxScale,
		Rounding:                 m.Rounding,
	}
}
