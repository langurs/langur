// /langur/str/split.go

package str

import (
	"fmt"
	"strings"
)

func SplitByNumber(s string, num, max int) ([]string, error) {
	if num == 0 {
		return nil, fmt.Errorf("Cannot split by 0")
	}

	rSlc := []rune(s)
	if max == 0 || len(rSlc) == 0 {
		return nil, nil
	}

	indices := [][]int{}
	begin := 0

	if num < 0 {
		num = -num
		first := len(rSlc) % num

		if max > -1 && max*num < len(rSlc) {
			first = len(rSlc) - num*(max-1)
		}

		if first != 0 {
			// not evenly divided by num, or max limits
			if num <= len(rSlc) {
				end := len(string(rSlc[begin:first]))
				indices = append(indices, []int{begin, end})
				begin = end
			}
		}
	}

	cpSlc := []rune{} // section slice

	for i := range rSlc {
		if max > -1 && len(indices)+1 > max {
			break
		}

		cpSlc = append(cpSlc, rSlc[i])
		if len(cpSlc) == num {
			// one section read
			end := begin + len(string(cpSlc))
			indices = append(indices, []int{begin, end})

			// reset
			begin = end
			cpSlc = nil
		}
	}
	if begin < len(s) {
		// have not picked up everything yet
		if max > -1 && len(indices)+1 > max {
			// already maxed out; reset end of last index
			indices[len(indices)-1][1] = len(s)
		} else {
			// pick up last pieces when uneven
			indices = append(indices, []int{begin, len(s)})
		}
	}

	return SlicesByIndices(s, indices)
}

func SlicesByIndices(s string, indices [][]int) ([]string, error) {
	sSlc := []string{}
	for i := range indices {
		sSlc = append(sSlc, s[indices[i][0]:indices[i][1]])
	}
	return sSlc, nil
}

func SplitAndKeep(s string, indices [][]int) ([]string, error) {
	// keeping all the parts
	sSlc := make([]string, len(indices)*2+1)

	begin := 0
	j := 0
	for i := range indices {
		sSlc[j] = s[begin:indices[i][0]]
		j++
		sSlc[j] = s[indices[i][0]:indices[i][1]]
		j++
		begin = indices[i][1]
	}
	sSlc[j] = s[begin:]

	return sSlc, nil
}

func SplitAndKeepZeroLengthDelim(s string, indices [][]int) ([]string, error) {
	sSlc, err := SplitAndKeep(s, indices)
	if len(sSlc) != 0 {
		return sSlc[1 : len(sSlc)-1], err
	}
	return sSlc, err
}

func ProgressiveIndices(find, s string, max int) [][]int {
	var indices [][]int
	offset := 0
	chkstr := s

	for i := 0; max == -1 || i < max; i++ {
		index := strings.Index(chkstr, find)
		if index == -1 {
			break
		}
		plus := index + len(find)

		startEnd := []int{offset + index, offset + plus}
		indices = append(indices, startEnd)

		offset += plus
		chkstr = chkstr[plus:]
	}
	return indices
}
