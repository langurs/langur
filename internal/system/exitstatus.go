// langur/system/exitstatus.go

package system

import (
	"math"
)

const (
	ExitStatusGeneral = iota
	ExitStatusNoScript
	ExitStatusHelp
	ExitStatusTest
	ExitStatusFailedArgs
	ExitStatusFailedReadFile
	ExitStatusFailedParse
	ExitStatusFailedCompile
	ExitStatusFailedRun
	ExitStatusArgToExitBad
	ExitStatusArgToExitOutOfRange
)

// Linux exit status codes (specifically bash)
// https://tldp.org/LDP/abs/html/exitcodes.html
// https://www.gnu.org/software/bash/manual/html_node/Exit-Status.html
// https://www.redhat.com/sysadmin/exit-codes-demystified
const generalStatusError = 1

var exitStatus = map[int]int{
	ExitStatusGeneral:             generalStatusError,
	ExitStatusNoScript:            0,
	ExitStatusHelp:                0,
	ExitStatusTest:                0,
	ExitStatusFailedArgs:          2,
	ExitStatusFailedReadFile:      127,
	ExitStatusFailedParse:         126,
	ExitStatusFailedCompile:       126,
	ExitStatusFailedRun:           1,
	ExitStatusArgToExitBad:        128,
	ExitStatusArgToExitOutOfRange: 255,
}

// Windows exit status codes
// https://learn.microsoft.com/en-us/windows/win32/debug/system-error-codes--0-499-
// It may be that there are better codes to use on Windows, so that these could be revised.
const generalStatusErrorWindows = 574

var exitStatusWindows = map[int]int{
	ExitStatusGeneral:             generalStatusErrorWindows,
	ExitStatusNoScript:            0,
	ExitStatusHelp:                0,
	ExitStatusTest:                0,
	ExitStatusFailedArgs:          160,
	ExitStatusFailedReadFile:      2,
	ExitStatusFailedParse:         575,
	ExitStatusFailedCompile:       575,
	ExitStatusFailedRun:           574,
	ExitStatusArgToExitBad:        574,
	ExitStatusArgToExitOutOfRange: 574,
}

// NOTE: These functions should not panic.

func GetExitStatus(s int) int {
	switch Type {
	case WINDOWS:
		status, ok := exitStatusWindows[s]
		if ok {
			return status
		}
		return generalStatusErrorWindows

	default:
		status, ok := exitStatus[s]
		if ok {
			return status
		}
		return generalStatusError
	}
}

func FixExitStatus(n int) int {
	var min, max int
	switch Type {
	case WINDOWS:
		min, max = math.MinInt32, math.MaxInt32
	default:
		min, max = 0, 255
	}
	if n < min || n > max {
		n = GetExitStatus(ExitStatusArgToExitOutOfRange)
	}
	return n
}
