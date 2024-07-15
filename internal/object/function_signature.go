// langur/object/function_signature.go

package object

import (
	"langur/opcode"
)

// NOTE: signatures to be used for both compiled functions and built-in functions

type Parameter struct {
	InternalName string // variable name within function; may change without affecting API
	ExternalName string // API / call name for optional parameter
	Mutable      bool

	DefaultValue             Object // for optional parameter; nil for positional parameters
	DefaultValueInstructions opcode.Instructions
}

func (p Parameter) Copy() Parameter {
	return Parameter{
		InternalName:             p.InternalName,
		ExternalName:             p.ExternalName,
		Mutable:                  p.Mutable,
		DefaultValue:             CopyOrNil(p.DefaultValue),
		DefaultValueInstructions: p.DefaultValueInstructions.Copy(),
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

type Signature struct {
	ParamPositional []Parameter
	ParamByName     []Parameter
}

func (s *Signature) Copy() *Signature {
	return &Signature{
		ParamPositional: copyParamList(s.ParamPositional),
		ParamByName:     copyParamList(s.ParamByName),
	}
}
