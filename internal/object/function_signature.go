// langur/object/function_signature.go

package object

import (
	"fmt"
	"langur/str"
	"strings"
)

// signatures to be used for both compiled functions and built-in functions
type Signature struct {
	Name              string
	Description       string
	ImpureEffects     bool
	ParamPositional   []Parameter
	ParamByName       []Parameter
	ParamExpansionMin int
	ParamExpansionMax int
}

func (s *Signature) SetParamDefault(name string, defaultValue Object) error {
	for i := range s.ParamByName {
		if s.ParamByName[i].ExternalName == name {
			s.ParamByName[i].DefaultValue = defaultValue
			return nil
		}
	}
	return fmt.Errorf("Optional parameter (%s) not found to set default", str.ReformatInput(name))
}

// including parameter expansion
// not including optional parameters
func (s *Signature) Max() int {
	if s.ParamExpansionMax == -1 {
		return -1
	}
	if s.ParamExpansionMax > 0 {
		return len(s.ParamPositional) + s.ParamExpansionMax - 1
	}
	return len(s.ParamPositional)
}

// including parameter expansion
// not including optional parameters
func (s *Signature) Min() int {
	if s.ParamExpansionMin > 0 {
		return len(s.ParamPositional) + s.ParamExpansionMin - 1
	} else if s.ParamExpansionMax != 0 {
		// We already know that ParamExpansionMin == 0 (from the previous test failing), ...
		// ... so if the maximum is not 0, we have an optional one at the end, so we subtract 1.
		return len(s.ParamPositional) - 1
	}
	return len(s.ParamPositional)
}

func (s *Signature) MinMaxString() string {
	min, max := s.Min(), s.Max()
	if min == max {
		return fmt.Sprintf("%d", max)
	}
	return fmt.Sprintf("%d..%d", min, max)
}

func (s *Signature) Copy() *Signature {
	return &Signature{
		Name:              s.Name,
		Description:       s.Description,
		ImpureEffects:     s.ImpureEffects,
		ParamPositional:   copyParamList(s.ParamPositional),
		ParamByName:       copyParamList(s.ParamByName),
		ParamExpansionMin: s.ParamExpansionMin,
		ParamExpansionMax: s.ParamExpansionMax,
	}
}

func copyParamList(pl []Parameter) []Parameter {
	if pl == nil {
		return nil
	}
	newPl := make([]Parameter, len(pl))
	for i := range pl {
		newPl[i] = pl[i].Copy()
	}
	return newPl
}

func (s *Signature) String() string {
	var sb strings.Builder

	if s.Name == "" {
		sb.WriteString("fn")
	} else {
		sb.WriteString("(fn)")
		sb.WriteString(s.Name)
	}

	if s.ImpureEffects {
		sb.WriteRune('*')
	}

	sb.WriteRune('(')

	for i, p := range s.ParamPositional {
		lastPositional := i == len(s.ParamPositional)-1

		if lastPositional &&
			(s.ParamExpansionMin != 0 || s.ParamExpansionMax != 0) {
			sb.WriteString("...[")
			sb.WriteString(str.IntToStr(s.ParamExpansionMin, 10))
			sb.WriteString("..")
			if s.ParamExpansionMax != -1 {
				sb.WriteString(str.IntToStr(s.ParamExpansionMax, 10))
			}
			sb.WriteString("] ")
		}

		sb.WriteString(p.String())

		if !lastPositional || len(s.ParamByName) != 0 {
			sb.WriteString(", ")
		}
	}

	for i, p := range s.ParamByName {
		lastByName := i == len(s.ParamByName)-1
		sb.WriteString(p.String())
		if !lastByName {
			sb.WriteString(", ")
		}
	}

	sb.WriteRune(')')
	return sb.String()
}

type Parameter struct {
	InternalName string // variable name within function; may change without affecting API
	ExternalName string // API / call name for optional parameter
	Mutable      bool

	// default for optional parameter
	DefaultValue Object
}

func (p Parameter) Copy() Parameter {
	return Parameter{
		InternalName: p.InternalName,
		ExternalName: p.ExternalName,
		Mutable:      p.Mutable,
		DefaultValue: CopyOrNil(p.DefaultValue),
	}
}

func (p Parameter) String() string {
	var sb strings.Builder

	if p.Mutable {
		sb.WriteString("var ")
	}

	sb.WriteString(p.InternalName)
	if p.InternalName != p.ExternalName && p.ExternalName != "" {
		sb.WriteString(" as ")
		sb.WriteString(p.ExternalName)
	}

	if p.DefaultValue != nil {
		sb.WriteRune('=')
		sb.WriteString(p.DefaultValue.String())
	}

	return sb.String()
}
