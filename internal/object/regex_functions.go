// langur/object/regex_functions.go

package object

import (
	"fmt"
	"langur/regex"
	"langur/regexp" // a modified copy of Go's standard regexp (re2) package
	"langur/str"
)

func RegexMatching(re *Regex, s string) (Object, error) {
	if re.RegexType == regex.RE2 {
		compiled := re.Compiled.(*regexp.Regexp)
		return NativeBoolToObject(compiled.MatchString(s)), nil
	} else {
		return nil, fmt.Errorf("Unknown regex type")
	}
}

func RegexMatchOnce(re *Regex, s string) (Object, error) {
	if re.RegexType == regex.RE2 {
		compiled := re.Compiled.(*regexp.Regexp)
		if compiled.MatchString(s) {
			return NewString(compiled.FindString(s)), nil
		}
	} else {
		return nil, fmt.Errorf("Unknown regex type")
	}
	return NONE, nil
}

func RegexMatchProgressive(re *Regex, s string, max int) (Object, error) {
	var sSlc []string

	if re.RegexType == regex.RE2 {
		compiled := re.Compiled.(*regexp.Regexp)
		sSlc = compiled.FindAllString(s, max)
	} else {
		return nil, fmt.Errorf("Unknown regex type")
	}

	arr := &List{}
	for _, s := range sSlc {
		arr.Elements = append(arr.Elements, NewString(s))
	}
	return arr, nil
}

func RegexSubMatches(re *Regex, s string) (Object, error) {
	var sSlc []string

	if re.RegexType == regex.RE2 {
		compiled := re.Compiled.(*regexp.Regexp)
		sSlc = compiled.FindStringSubmatch(s)
	} else {
		return nil, fmt.Errorf("Unknown regex type")
	}

	arr := &List{}
	// not including whole match (first value)
	for i := 1; i < len(sSlc); i++ {
		arr.Elements = append(arr.Elements, NewString(sSlc[i]))
	}
	return arr, nil
}

func RegexSubMatchesHash(re *Regex, s string) (Object, error) {
	var sSlc, names []string

	if re.RegexType == regex.RE2 {
		compiled := re.Compiled.(*regexp.Regexp)
		sSlc = compiled.FindStringSubmatch(s)
		names = compiled.SubexpNames()
	} else {
		return nil, fmt.Errorf("Unknown regex type")
	}

	hash := &Hash{}
	// including whole match (first value 0)
	for i := 0; i < len(sSlc); i++ {
		value := NewString(sSlc[i])
		err := hash.WritePair(NumberFromInt(i), value)
		if err != nil {
			return hash, err
		}
		// add named subexpressions
		if names[i] != "" {
			err = hash.WritePair(NewString(names[i]), value)
			if err != nil {
				return hash, err
			}
		}
	}
	return hash, nil
}

func RegexProgressiveSubMatches(re *Regex, s string, max int) (Object, error) {
	var slcOfSlc [][]string

	if re.RegexType == regex.RE2 {
		compiled := re.Compiled.(*regexp.Regexp)
		slcOfSlc = compiled.FindAllStringSubmatch(s, max)
	} else {
		return nil, fmt.Errorf("Unknown regex type")
	}

	arr := &List{}
	if len(slcOfSlc) > 0 {
		for _, sSlc := range slcOfSlc {
			arr2 := &List{}
			// not including whole match (first value)
			for i := 1; i < len(sSlc); i++ {
				arr2.Elements = append(arr2.Elements, NewString(sSlc[i]))
			}
			arr.Elements = append(arr.Elements, arr2)
		}
	}
	return arr, nil
}

func RegexProgressiveSubMatchesHashList(re *Regex, s string, max int) (Object, error) {
	var slcOfSlc [][]string
	var names []string

	if re.RegexType == regex.RE2 {
		compiled := re.Compiled.(*regexp.Regexp)
		slcOfSlc = compiled.FindAllStringSubmatch(s, max)
		names = compiled.SubexpNames()
	} else {
		return nil, fmt.Errorf("Unknown regex type")
	}

	arr := &List{}
	if len(slcOfSlc) > 0 {
		for _, sSlc := range slcOfSlc {
			hash := &Hash{}
			// including whole match (first value 0)
			for i := 0; i < len(sSlc); i++ {
				value := NewString(sSlc[i])
				err := hash.WritePair(NumberFromInt(i), value)
				if err != nil {
					return nil, err
				}
				// add named subexpressions
				if names[i] != "" {
					err = hash.WritePair(NewString(names[i]), value)
					if err != nil {
						return nil, err
					}
				}
			}
			arr.Elements = append(arr.Elements, hash)
		}
	}
	return arr, nil
}

var prepForNoSubmatchInterpolation = regexp.MustCompile("\\$")

func RegexReplace(src string, re *Regex, repl string, max int, doSubmatchInterpolation bool) (Object, error) {
	var newStr string

	if !doSubmatchInterpolation {
		repl = prepForNoSubmatchInterpolation.ReplaceAllString(repl, "$$$$")
	}

	if re.RegexType == regex.RE2 {
		compiled := re.Compiled.(*regexp.Regexp)

		if max == -1 {
			newStr = compiled.ReplaceAllString(src, repl)

		} else {
			newStr = compiled.ReplaceString(src, repl, max)
		}
	} else {
		return nil, fmt.Errorf("Unknown regex type")
	}

	return NewString(newStr), nil
}

func RegexSplit(re *Regex, s string, max int) (Object, error) {
	var sSlc []string

	if re.RegexType == regex.RE2 {
		compiled := re.Compiled.(*regexp.Regexp)
		sSlc = compiled.Split(s, max)
	} else {
		return nil, fmt.Errorf("Unknown regex type")
	}

	return StringSliceToList(sSlc), nil
}

func RegexSplitAndKeep(re *Regex, s string, max int) (Object, error) {
	// keeping all the parts
	var indices [][]int
	var sSlc []string
	var err error

	if re.RegexType == regex.RE2 {
		compiled := re.Compiled.(*regexp.Regexp)
		indices = compiled.FindAllStringIndex(s, max)
	} else {
		return nil, fmt.Errorf("Unknown regex type")
	}

	if re.Pattern == "" {
		sSlc, err = str.SplitAndKeepZeroLengthDelim(s, indices)
	} else {
		sSlc, err = str.SplitAndKeep(s, indices)
	}
	if err != nil {
		return nil, err
	}
	return StringSliceToList(sSlc), nil
}

// regex counterpart to StringIndex()
func RegexIndex(re *Regex, s string) (Object, error) {
	var index []int

	if re.RegexType == regex.RE2 {
		compiled := re.Compiled.(*regexp.Regexp)
		index = compiled.FindStringIndex(s)
	} else {
		return nil, fmt.Errorf("Unknown regex type")
	}

	if len(index) == 0 {
		return NONE, nil
	}

	start, end, err := str.CodeUnitToCodePointRange(s, index[0], index[1])
	return &Range{Start: NumberFromInt(start), End: NumberFromInt(end)}, err
}

func RegexProgressiveIndices(re *Regex, s string, max int) (Object, error) {
	var indices [][]int

	if re.RegexType == regex.RE2 {
		compiled := re.Compiled.(*regexp.Regexp)
		indices = compiled.FindAllStringIndex(s, max)
	} else {
		return nil, fmt.Errorf("Unknown regex type")
	}

	arr := &List{}
	for i := range indices {
		start, end, err := str.CodeUnitToCodePointRange(s, indices[i][0], indices[i][1])
		if err != nil {
			return nil, err
		}
		arr.Elements = append(arr.Elements, &Range{Start: NumberFromInt(start), End: NumberFromInt(end)})
	}
	return arr, nil
}

func RegexSubMatchesIndices(re *Regex, s string) (Object, error) {
	var indices [][]int

	if re.RegexType == regex.RE2 {
		compiled := re.Compiled.(*regexp.Regexp)
		indices = compiled.FindAllStringSubmatchIndex(s, 1)
	} else {
		return nil, fmt.Errorf("Unknown regex type")
	}

	arr := &List{}
	if len(indices) > 0 {
		// ignoring whole match here
		for i := 2; i < len(indices[0]); i = i + 2 {
			start, end, err := str.CodeUnitToCodePointRange(s, indices[0][i], indices[0][i+1])
			if err != nil {
				return nil, err
			}
			arr.Elements = append(arr.Elements, &Range{Start: NumberFromInt(start), End: NumberFromInt(end)})
		}
	}
	return arr, nil
}

func RegexProgressiveSubMatchesIndices(re *Regex, s string, max int) (Object, error) {
	var indices [][]int

	if re.RegexType == regex.RE2 {
		compiled := re.Compiled.(*regexp.Regexp)
		indices = compiled.FindAllStringSubmatchIndex(s, max)
	} else {
		return nil, fmt.Errorf("Unknown regex type")
	}

	arr := &List{}
	for i := range indices {
		arr2 := &List{}
		// ignoring whole matches here
		for j := 2; j < len(indices[i]); j = j + 2 {
			start, end, err := str.CodeUnitToCodePointRange(s, indices[i][j], indices[i][j+1])
			if err != nil {
				return nil, err
			}
			arr2.Elements = append(arr2.Elements, &Range{Start: NumberFromInt(start), End: NumberFromInt(end)})
		}
		arr.Elements = append(arr.Elements, arr2)
	}
	return arr, nil
}
