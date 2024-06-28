// langur/str/str.go

package str

import (
	"fmt"
	"langur/cpoint"
	"strings"

	// for grapheme cluster segmentation
	"github.com/rivo/uniseg"
)

func PadRight(s string, somuch int, with rune) string {
	if len([]rune(s)) < somuch {
		return s + strings.Repeat(string(with), somuch-len([]rune(s)))
	}
	return s
}

func PadLeft(s string, somuch int, with rune) string {
	if len([]rune(s)) < somuch {
		return strings.Repeat(string(with), somuch-len([]rune(s))) + s
	}
	return s
}

func Pad(s string, somuch int, with rune) string {
	if somuch > 0 {
		return PadLeft(s, somuch, with)
	}
	return PadRight(s, -somuch, with)
}

func RemoveTrailing(s string, trailer byte) string {
	for i := len(s) - 1; i > 0; i-- {
		if s[i] != trailer {
			return s[:i+1]
		}
	}
	return s[:1]
}

func ReformatInput(s string) string {
	// may do more than this if deemed necessary...
	return LimitGraphemes(s, 40, "...")
}

func Limit(s string, limit int, internalOverflowIndicator string) string {
	// limit by code points
	if limit < 0 {
		return limitLeft(s, -limit, internalOverflowIndicator)
	}
	return limitRight(s, limit, internalOverflowIndicator)
}

func limitRight(s string, limit int, internalOverflowIndicator string) string {
	// limit right by code points
	rSlc := []rune(s)
	if len(rSlc) > limit {
		overflowSlc := []rune(internalOverflowIndicator)
		if limit < len(overflowSlc) {
			return internalOverflowIndicator
		}
		cpLimit := limit - len(overflowSlc)
		rSlc = append(rSlc[:cpLimit], overflowSlc...)
	}
	return string(rSlc)
}

func limitLeft(s string, limit int, internalOverflowIndicator string) string {
	// limit left by code points
	rSlc := []rune(s)
	if len(rSlc) > limit {
		overflowSlc := []rune(internalOverflowIndicator)
		if limit < len(overflowSlc) {
			return internalOverflowIndicator
		}
		cpLimit := limit - len(overflowSlc)
		rSlc = append(overflowSlc, rSlc[len(rSlc)-cpLimit:]...)
	}
	return string(rSlc)
}

func LimitGraphemes(s string, limit int, internalOverflowIndicator string) string {
	if limit < 0 {
		return limitLeftGraphemes(s, -limit, internalOverflowIndicator)
	}
	return limitRightGraphemes(s, limit, internalOverflowIndicator)
}

func limitLeftGraphemes(s string, limit int, internalOverflowIndicator string) string {
	grSlc := Graphemes(s)
	if len(grSlc) > limit {
		overflowSlc := Graphemes(internalOverflowIndicator)
		if limit < len(overflowSlc) {
			return internalOverflowIndicator
		}
		grLimit := limit - len(overflowSlc)
		grSlc = append(overflowSlc, grSlc[len(grSlc)-grLimit:]...)
	}
	return GraphemesToString(grSlc)
}

func limitRightGraphemes(s string, limit int, internalOverflowIndicator string) string {
	grSlc := Graphemes(s)
	if len(grSlc) > limit {
		overflowSlc := Graphemes(internalOverflowIndicator)
		if limit < len(overflowSlc) {
			return internalOverflowIndicator
		}
		grLimit := limit - len(overflowSlc)
		grSlc = append(grSlc[:grLimit], overflowSlc...)
	}
	return GraphemesToString(grSlc)
}

func SubStr(s string, start, limit int) (string, error) {
	rSlc, err := RuneSlice(s, start, limit)
	if err != nil {
		return "", err
	}
	return string(rSlc), nil
}

func RuneSlice(s string, start, limit int) (rSlc []rune, err error) {
	rSlc = []rune(s)

	if start > len(rSlc)-1 || start < 1 {
		return []rune(""), fmt.Errorf("Start greater than string code point length")
	}

	rSlc = rSlc[start-1:]
	if limit > -1 && len(rSlc) > limit {
		rSlc = rSlc[:limit]
	}
	return
}

func CodeUnitToCodePointRange(from string, start, end int) (s int, e int, err error) {
	// From a 0-based code unit range with exclusive top-end, ...
	// ... we need to create a 1-based code point range with inclusive top-end.
	s, err = cpoint.CodePointPosFromCodeUnitPos(&from, start)
	if err != nil {
		return
	}
	e, err = cpoint.CodePointPosFromCodeUnitPos(&from, end)
	e--
	return
}

func CopySlice(original []string) []string {
	if original == nil {
		return original
	}
	newSlc := make([]string, len(original))
	copy(newSlc, original)
	return newSlc
}

func CopySliceOfSlice(original [][]string) [][]string {
	if original == nil {
		return original
	}
	newSlc := make([][]string, len(original))
	for i := range original {
		copy(newSlc[i], CopySlice(original[i]))
	}
	return newSlc
}

func IsInSlice(str string, slc []string) bool {
	for i := range slc {
		if str == slc[i] {
			return true
		}
	}
	return false
}

func Balanced(s string) bool {
	openers := [][]rune{
		[]rune{0x2066, 0x2067, 0x2068},
		[]rune{0x202A, 0x202B, 0x202D, 0x202E},
	}
	closers := [][]rune{
		[]rune{0x2069},
		[]rune{0x202C},
	}

	return BalancedCodePoints(s, openers, closers)
}

func BalancedCodePoints(s string, openers [][]rune, closers [][]rune) bool {
	const noMatch = -1

	if len(openers) != len(closers) {
		return false
	}
	if len(openers) == 0 {
		return true
	}

	var stack = []int{}

	getSetIdx := func(cp rune, set [][]rune) int {
		for i := range set {
			if cpoint.InSlice(cp, set[i]) {
				return i
			}
		}
		return noMatch
	}
	getStackIdx := func() int {
		if len(stack) == 0 {
			return noMatch
		}
		return stack[len(stack)-1]
	}

	rSlc := []rune(s)
	for _, cp := range rSlc {
		openerIdx := getSetIdx(cp, openers)
		closerIdx := getSetIdx(cp, closers)
		stackIdx := getStackIdx()

		if closerIdx != noMatch {
			if closerIdx != stackIdx {
				// closer doesn't match last opener, if any
				return false
			}
			// pop
			stack = stack[:len(stack)-1]
		}
		if openerIdx != noMatch {
			// push
			stack = append(stack, openerIdx)
		}
	}

	return len(stack) == 0
}

func Reverse(s string) string {
	// reverse string paying attention to grapheme clusters
	return uniseg.ReverseString(s)
}

func Graphemes(s string) [][]rune {
	graphemes := [][]rune{}
	bSlc := []byte(s)
	state := -1
	var gr []byte
	for len(bSlc) > 0 {
		gr, bSlc, _, state = uniseg.FirstGraphemeCluster(bSlc, state)
		graphemes = append(graphemes, []rune(string(gr)))
	}
	return graphemes
}

func GraphemesToString(grSlc [][]rune) string {
	var sb strings.Builder

	for _, slc := range grSlc {
		sb.WriteString(string(slc))
	}

	return sb.String()
}
