// langur/system/exitstatus.go

package system

import (
	"math"
)

// Linux exit status codes (specifically bash)
// https://tldp.org/LDP/abs/html/exitcodes.html
// https://www.gnu.org/software/bash/manual/html_node/Exit-Status.html
// https://www.redhat.com/sysadmin/exit-codes-demystified
const generalStatusError = 1

var exitStatus = map[string]int{
	"":                    generalStatusError,
	"noscript":            0,
	"help":                0,
	"test":                0,
	"failedargs":          2,
	"failedreadfile":      127,
	"failedparse":         126,
	"failedcompile":       126,
	"failedrun":           1,
	"argtoexitBad":        128,
	"argtoexitOutofrange": 255,
}

// Windows exit status codes
// https://learn.microsoft.com/en-us/windows/win32/debug/system-error-codes--0-499-
// It may be that there are better codes to use on Windows, so that these could be revised.
const generalStatusErrorWindows = 574

var exitStatusWindows = map[string]int{
	"":                    generalStatusErrorWindows,
	"noscript":            0,
	"help":                0,
	"test":                0,
	"failedargs":          160,
	"failedreadfile":      2,
	"failedparse":         575,
	"failedcompile":       575,
	"failedrun":           574,
	"argtoexitBad":        574,
	"argtoexitOutofrange": 574,
}

// NOTE: These functions should not panic.

func GetExitStatus(s string) int {
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
		n = GetExitStatus("argtoexitOutofrange")
	}
	return n
}
