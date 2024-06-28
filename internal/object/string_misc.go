// langur/object/string_misc.go

package object

import (
	"langur/str"
	"strings"
)

// plain string counterpart to RegexIndex()
func StringIndex(sub, s string) (Object, error) {
	start := strings.Index(s, sub)

	if start == -1 {
		return NONE, nil
	}

	start, end, err := str.CodeUnitToCodePointRange(s, start, start+len(sub))
	return &Range{Start: NumberFromInt(start), End: NumberFromInt(end)}, err
}

// plain string counterpart to RegexProgressiveIndices()
func StringProgressiveIndices(find, s string, max int) (Object, error) {
	arr := &List{}
	offset := 0
	chkstr := s

	for i := 0; max == -1 || i < max; i++ {
		index := strings.Index(chkstr, find)
		if index == -1 {
			break
		}
		plus := index + len(find)

		start, end, err := str.CodeUnitToCodePointRange(s, offset+index, offset+plus)
		if err != nil {
			return nil, err
		}
		arr.Elements = append(arr.Elements, &Range{Start: NumberFromInt(start), End: NumberFromInt(end)})

		offset += plus
		chkstr = chkstr[plus:]
	}
	return arr, nil
}

func StringSplitAndKeep(find, s string, max int) (Object, error) {
	indices := str.ProgressiveIndices(find, s, max)
	var sSlc []string
	var err error

	if find == "" {
		sSlc, err = str.SplitAndKeepZeroLengthDelim(s, indices)
	} else {
		sSlc, err = str.SplitAndKeep(s, indices)
	}
	if err != nil {
		return nil, err
	}
	return StringSliceToList(sSlc), nil
}

func StringConcat(objSlc []Object) Object {
	sb := strings.Builder{}
	for _, obj := range objSlc {
		sb.WriteString(obj.String())
	}
	return NewString(sb.String())
}

func StringSliceToList(sSlc []string) Object {
	arr := &List{Elements: make([]Object, len(sSlc))}
	for i := range sSlc {
		arr.Elements[i] = NewString(sSlc[i])
	}
	return arr
}

func ToString(obj Object) *String {
	switch obj := obj.(type) {
	case *String:
		return obj
	default:
		return NewString(obj.String())
	}
}

func ExpectString(obj Object) (string, bool) {
	switch obj := obj.(type) {
	case *String:
		return obj.String(), true
	}
	return "", false
}

func ToStringFromSlice(objSlc []Object) *String {
	var sb strings.Builder
	for _, obj := range objSlc {
		sb.WriteString(obj.String())
	}
	return NewString(sb.String())
}

func SliceToStringSlice(objSlc []Object) []string {
	newSlc := make([]string, len(objSlc))
	for i := range objSlc {
		newSlc[i] = objSlc[i].String()
	}
	return newSlc
}
