// langur/object/function.go

// CompiledCode and BuiltIn object types
// CompiledCode objects not strictly functions, as they have other uses

package object

import (
	"bytes"
	"fmt"
	"langur/common"
	"langur/opcode"
)

func IsCallable(obj Object) bool {
	return obj.Type() == BUILTIN_FUNCTION_OBJ || IsCompiledFunction(obj)
}

func IsCompiledFunction(obj Object) bool {
	return obj.Type() == COMPILED_CODE_OBJ && obj.(*CompiledCode).IsFunction()
}

// -1 used to indicate no maximum
const NotCallable = -2

func ParamExpectedString(obj Object) string {
	min, max := ParamMin(obj), ParamMax(obj)
	if min == NotCallable {
		return "N/A"
	}
	return fmt.Sprintf("%d..%d", min, max)
}

// minimum, including parameter expansion
func ParamMin(obj Object) int {
	switch fn := obj.(type) {
	case *CompiledCode:
		if fn.FnSignature == nil {
			return 0
		}
		return fn.FnSignature.Min()

	case *BuiltIn:
		return fn.ParamMin
	}
	return NotCallable
}

// maximum, including parameter expansion
func ParamMax(obj Object) int {
	switch fn := obj.(type) {
	case *CompiledCode:
		if fn.FnSignature == nil {
			return 0
		}
		return fn.FnSignature.Max()

	case *BuiltIn:
		return fn.ParamMax
	}
	return NotCallable
}

type CompiledCode struct {
	FnSignature        *Signature
	Instructions       opcode.Instructions
	LocalBindingsCount int      // including parameters
	Free               []Object // "free" variables for closures
}

func (cf *CompiledCode) IsFunction() bool {
	return cf.FnSignature != nil
}

func (cf *CompiledCode) HasImpureEffects() bool {
	// TODO(?): check
	if cf.FnSignature == nil {
		return false
	}
	return cf.FnSignature.ImpureEffects
}

func (cf *CompiledCode) Copy() Object {
	return &CompiledCode{
		FnSignature:        cf.FnSignature.Copy(),
		Instructions:       cf.Instructions.Copy(),
		LocalBindingsCount: cf.LocalBindingsCount,
		Free:               CopyRefSlice(cf.Free),
	}
}

func (cf *CompiledCode) Equal(cf2 Object) bool {
	return cf == cf2
}

func (cf *CompiledCode) Type() ObjectType {
	return COMPILED_CODE_OBJ
}
func (cf *CompiledCode) TypeString() string {
	if cf.IsFunction() {
		return common.FuntionTypeName
	}
	// a string not likely to be seen in langur...
	return common.CompiledCodeTypeName
}

func (cf *CompiledCode) IsTruthy() bool {
	return !cf.HasImpureEffects()
}

func (cf *CompiledCode) String() string {
	if cf.IsFunction() {
		return cf.FnSignature.String()

	} else {
		// wouldn't likely happen
		return INTERNAL_OBJECT_ONLY
	}
}

func (cf *CompiledCode) ReplString() string {
	var out bytes.Buffer

	if cf.FnSignature.ImpureEffects {
		out.WriteString("Impure ")
	}

	if cf.IsFunction() {
		out.WriteString(fmt.Sprintf(common.FuntionTypeName+" %s (%p)", cf.FnSignature.Name, cf))
	} else {
		out.WriteString(fmt.Sprintf(common.CompiledCodeTypeName+" (%p)", cf))
	}

	if len(cf.FnSignature.ParamPositional) != 0 {
		out.WriteString(fmt.Sprintf("; Positional Parameters: %d..%d", ParamMin(cf), ParamMax(cf)))
	}
	if len(cf.FnSignature.ParamByName) != 0 {
		out.WriteString(fmt.Sprintf("; Parameters By Name: %d", len(cf.FnSignature.ParamByName)))
	}
	if cf.LocalBindingsCount > 0 {
		out.WriteString(fmt.Sprintf("; LocalBindingsCount: %d", cf.LocalBindingsCount))
	}
	if len(cf.Free) > 0 {
		out.WriteString(fmt.Sprintf("; FreeCount: %d", len(cf.Free)))
	}
	out.WriteString("\nInstructions\n")

	out.WriteString(cf.Instructions.String())

	return out.String()
}

// BUILT-IN FUNCTIONS

type BuiltIn struct {
	// Fn an interface{} here and type assertion in the process package to avoid an import cycle error
	Fn interface{}

	Name          string
	Description   string
	ParamMin      int
	ParamMax      int
	ImpureEffects bool
}

func (b *BuiltIn) Copy() Object {
	return &BuiltIn{
		Fn:            b.Fn,
		Name:          b.Name,
		Description:   b.Description,
		ParamMin:      b.ParamMin,
		ParamMax:      b.ParamMax,
		ImpureEffects: b.ImpureEffects,
	}
}

func (b *BuiltIn) FullName() string {
	return b.Name
}

func (b *BuiltIn) HasImpureEffects() bool {
	return b.ImpureEffects
}

func (l *BuiltIn) Equal(b2 Object) bool {
	r, ok := b2.(*BuiltIn)
	if !ok {
		return false
	}
	return l == r
}

func (b *BuiltIn) Type() ObjectType {
	return BUILTIN_FUNCTION_OBJ
}
func (b *BuiltIn) TypeString() string {
	return common.BuiltInTypeName
}

func (b *BuiltIn) IsTruthy() bool {
	return !b.ImpureEffects
}

func (b *BuiltIn) String() string {
	if b.Name[0] == '_' {
		// internal function names only start with underscore
		// likely won't happen, but shouldn't
		return INTERNAL_OBJECT_ONLY
	}

	var out bytes.Buffer

	out.WriteRune('(')
	if b.ImpureEffects {
		out.WriteString("impure ")
	}
	out.WriteString("builtin) " + b.FullName())

	return out.String()
}
func (b *BuiltIn) ReplString() string {
	var out bytes.Buffer

	if b.ImpureEffects {
		out.WriteString("Impure ")
	}
	out.WriteString(common.BuiltInTypeName)
	out.WriteString(" " + b.FullName())

	return out.String()
}
