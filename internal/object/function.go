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
	return obj.Type() == COMPILED_CODE_OBJ && obj.(*CompiledCode).IsFunction
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

func ParamMin(obj Object) int {
	switch fn := obj.(type) {
	case *CompiledCode:
		if fn.ParamExpansionMin > 0 {
			return fn.ParamMin + fn.ParamExpansionMin - 1
		} else if fn.ParamExpansionMax != 0 {
			// We already know that ParamExpansionMin == 0 (from the previous test failing), ...
			// ... so if the maximum is not 0, we have an optional one at the end, so we subtract 1.
			return fn.ParamMin - 1
		}
		return fn.ParamMin
	case *BuiltIn:
		return fn.ParamMin
	}
	return NotCallable
}

func ParamMax(obj Object) int {
	switch fn := obj.(type) {
	case *CompiledCode:
		if fn.ParamExpansionMax == -1 {
			return -1
		}
		if fn.ParamExpansionMax > 0 {
			return fn.ParamMax + fn.ParamExpansionMax - 1
		}
		return fn.ParamMax
	case *BuiltIn:
		return fn.ParamMax
	}
	return NotCallable
}

type CompiledCode struct {
	Name          string
	IsFunction    bool
	ImpureEffects bool
	Instructions  opcode.Instructions

	ParamMin           int
	ParamMax           int
	LocalBindingsCount int // including parameters

	ParamExpansionMin int
	ParamExpansionMax int

	// "free" variables for closures
	Free []Object
}

func (cf *CompiledCode) HasImpureEffects() bool {
	return cf.ImpureEffects
}

func (cf *CompiledCode) Copy() Object {
	return &CompiledCode{
		Name:               cf.Name,
		IsFunction:         cf.IsFunction,
		ImpureEffects:      cf.ImpureEffects,
		Instructions:       cf.Instructions.Copy(),
		ParamMin:           cf.ParamMin,
		ParamMax:           cf.ParamMax,
		LocalBindingsCount: cf.LocalBindingsCount,
		ParamExpansionMin:  cf.ParamExpansionMin,
		ParamExpansionMax:  cf.ParamExpansionMax,
		Free:               CopySlice(cf.Free),
	}
}

func (cf *CompiledCode) Equal(cf2 Object) bool {
	return cf == cf2
}

func (cf *CompiledCode) Type() ObjectType {
	return COMPILED_CODE_OBJ
}
func (cf *CompiledCode) TypeString() string {
	if cf.IsFunction {
		return common.FuntionTypeName
	}
	// a string not likely to be seen in langur...
	return common.CompiledCodeTypeName
}

func (cf *CompiledCode) IsTruthy() bool {
	return !cf.ImpureEffects
}

func (cf *CompiledCode) String() string {
	if cf.IsFunction {
		var out bytes.Buffer

		out.WriteRune('(')
		out.WriteString("fn")
		if cf.ImpureEffects {
			out.WriteRune('*')
		}
		out.WriteRune(')')

		if cf.Name != "" {
			out.WriteString(" " + cf.Name)
		}

		return out.String()

	} else {
		// wouldn't likely happen
		return INTERNAL_OBJECT_ONLY
	}
}

func (cf *CompiledCode) ReplString() string {
	var out bytes.Buffer

	if cf.ImpureEffects {
		out.WriteString("Impure ")
	}

	if cf.IsFunction {
		out.WriteString(fmt.Sprintf(common.FuntionTypeName+" %s (%p)", cf.Name, cf))
	} else {
		out.WriteString(fmt.Sprintf(common.CompiledCodeTypeName+" (%p)", cf))
	}

	if cf.ParamMin > 0 || cf.ParamMax != 0 {
		out.WriteString(fmt.Sprintf("; Parameters: %d..%d", cf.ParamMin, cf.ParamMax))
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
	// Fn an interface{} here and type assertion in the process package to avoid an import cycle
	Fn            interface{}
	Library       string
	Name          string
	Description   string
	ParamMin      int
	ParamMax      int
	ImpureEffects bool
}

func (b *BuiltIn) Copy() Object {
	return &BuiltIn{
		Fn:            b.Fn,
		Library:       b.Library,
		Name:          b.Name,
		Description:   b.Description,
		ParamMin:      b.ParamMin,
		ParamMax:      b.ParamMax,
		ImpureEffects: b.ImpureEffects,
	}
}

func (b *BuiltIn) FullName() string {
	if b.Library == "" {
		return b.Name
	}
	return b.Library + "." + b.Name
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
