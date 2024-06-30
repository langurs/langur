// NOTE: added for langur: replace with max count

// These do replace with maximum count, which is missing from the standard regexp library.
// This is WITHOUT WARRANTY of any kind.

// copied and modified from the regexp/regexp.go file from go 1.22.4
// It's a small change, passing a maximum and checking it.

package regexp

import (
	"strings"
	"unicode/utf8"
)

// ReplaceString(): just a slight modification of the ReplaceAllString() receiver function
func (re *Regexp) ReplaceString(src, repl string, max int) string {
	n := 2
	if strings.Contains(repl, "$") {
		n = 2 * (re.numSubexp + 1)
	}
	b := re.replace(nil, src, n,
		func(dst []byte, match []int) []byte {
			return re.expand(dst, repl, nil, src, match)
		},
		max)
	return string(b)
}

// replace(): just a slight modification of the replaceAll() receiver function
func (re *Regexp) replace(
	bsrc []byte, src string, nmatch int, repl func(dst []byte, m []int) []byte, max int) []byte {

	lastMatchEnd := 0 // end position of the most recent match
	searchPos := 0    // position where we next look for a match
	var buf []byte
	var endPos int
	if bsrc != nil {
		endPos = len(bsrc)
	} else {
		endPos = len(src)
	}
	if nmatch > re.prog.NumCap {
		nmatch = re.prog.NumCap
	}

	// count so we can stop when maximum is reached
	cnt := 0
	doCnt := max >= 0

	var dstCap [2]int
	for searchPos <= endPos {
		if doCnt {
			// increment and check count so we can stop when maximum is reached
			cnt++
			if cnt > max {
				break
			}
		}

		a := re.doExecute(nil, bsrc, src, searchPos, nmatch, dstCap[:0])
		if len(a) == 0 {
			break // no more matches
		}

		// Copy the unmatched characters before this match.
		if bsrc != nil {
			buf = append(buf, bsrc[lastMatchEnd:a[0]]...)
		} else {
			buf = append(buf, src[lastMatchEnd:a[0]]...)
		}

		// Now insert a copy of the replacement string, but not for a
		// match of the empty string immediately after another match.
		// (Otherwise, we get double replacement for patterns that
		// match both empty and nonempty strings.)
		if a[1] > lastMatchEnd || a[0] == 0 {
			buf = repl(buf, a)
		}
		lastMatchEnd = a[1]

		// Advance past this match; always advance at least one character.
		var width int
		if bsrc != nil {
			_, width = utf8.DecodeRune(bsrc[searchPos:])
		} else {
			_, width = utf8.DecodeRuneInString(src[searchPos:])
		}
		if searchPos+width > a[1] {
			searchPos += width
		} else if searchPos+1 > a[1] {
			// This clause is only needed at the end of the input
			// string. In that case, DecodeRuneInString returns width=0.
			searchPos++
		} else {
			searchPos = a[1]
		}
	}

	// Copy the unmatched characters after the last match.
	if bsrc != nil {
		buf = append(buf, bsrc[lastMatchEnd:]...)
	} else {
		buf = append(buf, src[lastMatchEnd:]...)
	}

	return buf
}
