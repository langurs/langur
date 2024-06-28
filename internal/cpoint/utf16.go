// langur/cpoint/utf16.go

package cpoint

import (
	"fmt"
)

const (
	twoByteMax = 0x10000

	lead_surrogate_start     = 0xD800
	lead_surrogate_end       = 0xDBFF
	trailing_surrogate_start = 0xDC00
	trailing_surrogate_end   = 0xDFFF

	lead_offset      = lead_surrogate_start - (0x10000 >> 10)
	surrogate_offset = 0x10000 - (lead_surrogate_start << 10) - trailing_surrogate_start
)

func IsSurrogate(cp rune) bool {
	return cp >= lead_surrogate_start && cp <= trailing_surrogate_end
}

// TODO: test and use the following functions

func cpToCpOrSurrogates(cp rune) (lead, trail rune, surrogates bool) {
	if cp > twoByteMax {
		lead = lead_offset + (cp >> 10)
		trail = trailing_surrogate_start + (cp & 0x3FF)
		surrogates = true

	} else {
		lead = cp
	}

	return
}

// ReplaceSurrogates()
// creates a UTF-32 rune slice from a (potentially) UTF-16 slice, replacing
// any surrogate pairs and throwing an error for invalid surrogates

// The problem with unicode/utf16/Decode is that it does not give an error
// for invalid surrogates (just replaces with substitution code point).
func ReplaceSurrogates(rSlc []rune) ([]rune, error) {
	var newSlc []rune
	var lead rune

	noLead := true

	for i := 0; i < len(rSlc); i++ {
		cp := rSlc[i]
		if noLead && cp >= lead_surrogate_start && cp <= lead_surrogate_end {
			lead = cp
			noLead = false
			continue

		} else if cp >= trailing_surrogate_start && cp <= trailing_surrogate_end {
			if noLead {
				return nil, fmt.Errorf("Independent trailing surrogate (%s) at code indexed at %d", Display(cp), i+1)
			}

			cp = (lead << 10) + cp + surrogate_offset
			noLead = true

			if newSlc == nil {
				// only allocating newSlc if there are surrogates
				newSlc = make([]rune, 0, len(rSlc)-1)
				copy(newSlc, rSlc[:i-1])
			}

			newSlc = append(newSlc, cp)

		} else if !noLead {
			return nil, fmt.Errorf("Independent lead surrogate (%s) at code indexed at %d", Display(lead), i)

		} else if newSlc != nil {
			// non-surrogate code; newSlc already initialized
			newSlc = append(newSlc, cp)
		}
	}

	if !noLead {
		return nil, fmt.Errorf("Independent lead surrogate (%s) at code indexed at %d", Display(lead), len(rSlc)+1)
	}

	if newSlc == nil {
		// no surrogates found; return original slice
		return rSlc, nil
	}

	return newSlc, nil
}
