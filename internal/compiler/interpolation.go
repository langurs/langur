// langur/compiler/interpolation.go

package compiler

import (
	"fmt"
	"langur/ast"
	"langur/common"
	"langur/cpoint"
	"langur/format"
	"langur/object"
	"langur/opcode"
	"langur/regex"
	"langur/regexp"
	"langur/str"
)

// NOTE: Coordinate these things with opFormat (in the VM).

var modifierRegexForFn = regexp.MustCompile(`^` + common.FunctionTokenLiteral +
	` (?P<fn>` + common.IdentifierRegexString + `)$`)

// var modifierRegexForFn = regexp.MustCompile(`^-> (?P<fn>` + common.IdentifierRegexString + `)$`)

var modifierRegexForDateTime = regexp.MustCompile(
	`^` + common.DateTimeTokenLiteral + `\((?P<format>.+)\)$|^` +
		common.DateTimeTokenLiteral + ` (?P<var>` + common.IdentifierRegexString + `)$`)

var modifierRegexForTruncate = regexp.MustCompile(`^t(?:(?P<max>-?[0-9]+)(?P<trimTrailingZeroes>-)?)?$`)
var modifierRegexForRounding = regexp.MustCompile(`^r(?:(?P<max>-?[0-9]+)(?P<trimTrailingZeroes>-)?)?$`)

var modifierRegexForAlign = regexp.MustCompile(
	`^(?P<align>-?[1-9][0-9]*)(?:\((?:(?P<withcp>.)|(?P<withcpnum>[0-9a-fA-F]{2,8}))\))?$`)

// var modifierRegexForAlignGraphemes = regexp.MustCompile(
// 	`^g(?P<align>-?[1-9][0-9]*)(?:\((?:(?P<withcp>.)|(?P<withcpnum>[0-9a-fA-F]{2,8}))\))?$`)

var modifierRegexForLimit = regexp.MustCompile(
	`^L(?P<limit>-?[1-9][0-9]*)(?:\((?P<internal>[^)]*)\))?$`)

var modifierRegexForLimitGraphemes = regexp.MustCompile(
	`^Lg(?P<limit>-?[1-9][0-9]*)(?:\((?P<internal>[^)]*)\))?$`)

var modifierRegexForHex = regexp.MustCompile(`^(?P<sign>[+])?(?P<uc>[xX])(?P<min>[0-9]+)?$`)
var modifierRegexForBase = regexp.MustCompile(
	`^(?P<sign>[+])?(?P<base>[1-9][0-9]*)(?P<uc>[xX])(?:(?P<min>[0-9]+))?$`)
var modifierRegexForFixed = regexp.MustCompile(
	`^(?P<sign>[+])?10x(?P<int>[0-9]+)\.(?P<frac>[0-9]+)(?P<trimTrailingZeroes>-)?$`)

var modifierRegexForScientificNotation = regexp.MustCompile(
	`^(?P<sign>[+])?(?:(?P<scale>\d+)(?P<scaleTrimTrailingZeroes>-)?)?(?P<uc>[eE])(?P<expsign>[+])?(?P<scaleExp>\d+)?$`)

func subMatchByName(name string, subs, names []string) string {
	// used internally and assuming slices match
	for i := range subs {
		if names[i] == name {
			return subs[i]
		}
	}
	bug("subMatchByName", fmt.Sprintf("Unknown submatch name %q", name))
	return "ERROR"
}

func (c *Compiler) compileInterpolationModifiers(node ast.Node, modifiers []string, regexType regex.RegexType) (
	ins opcode.Instructions, err error) {

	// gather them first for special cases, such as a scientific notation modifier following a fixed point modifier
	var mods []opcode.Instructions

	for _, mod := range modifiers {
		temp, err := c.compileInterpolationModifierIns(node, mod, regexType)
		if err != nil {
			return temp, err
		}
		mods = append(mods, temp)
	}
	// TODO: check for scientific notation modifier following a fixed point modifier
	// That is, if a number doesn't fit within fixed point, optionally convert to scientific notation.
	for _, m := range mods {
		ins = append(ins, m...)
	}

	return
}

func (c *Compiler) compileInterpolationModifierIns(
	node ast.Node, mod string, regexType regex.RegexType) (
	ins opcode.Instructions, err error) {

	if mod == format.MODSTRING_ESCAPE {
		reType := object.NumberFromInt(int(regexType))
		ins = append(ins, c.constantIns(reType)...)
		ins = append(ins, opcode.Make(opcode.OpFormat, format.FORMAT_ESCAPE)...)

	} else if mod == "T" {
		ins = append(ins, opcode.Make(opcode.OpFormat, format.FORMAT_TYPE)...)

	} else if mod == "cp" {
		ins = append(ins, opcode.Make(opcode.OpFormat, format.FORMAT_CODE_POINT)...)

	} else if m := modifierRegexForAlign.FindStringSubmatch(mod); m != nil {
		return c.compileModifierInsForAlignment(node, m)

		// } else if m := modifierRegexForAlignGraphemes.FindStringSubmatch(mod); m != nil {
		// 	return c.compileModifierInsForAlignmentGraphemes(node, m)

	} else if m := modifierRegexForLimit.FindStringSubmatch(mod); m != nil {
		return c.compileModifierInsForLimit(node, m)

	} else if m := modifierRegexForLimitGraphemes.FindStringSubmatch(mod); m != nil {
		return c.compileModifierInsForLimitGraphemes(node, m)

	} else if m := modifierRegexForTruncate.FindStringSubmatch(mod); m != nil {
		return c.compileModifierInsForTruncate(node, m)

	} else if m := modifierRegexForRounding.FindStringSubmatch(mod); m != nil {
		return c.compileModifierInsForRounding(node, m)

	} else if m := modifierRegexForFn.FindStringSubmatch(mod); m != nil {
		return c.compileModifierInsForCustomFn(node, m)

	} else if m := modifierRegexForHex.FindStringSubmatch(mod); m != nil {
		return c.compileModifierInsForHex(node, m)

	} else if m := modifierRegexForBase.FindStringSubmatch(mod); m != nil {
		return c.compileModifierInsForCustomBase(node, m)

	} else if m := modifierRegexForFixed.FindStringSubmatch(mod); m != nil {
		return c.compileModifierInsForFixedNotation(node, m)

	} else if m := modifierRegexForScientificNotation.FindStringSubmatch(mod); m != nil {
		return c.compileModifierInsForScientificNotation(node, m)

	} else if m := modifierRegexForDateTime.FindStringSubmatch(mod); m != nil {
		return c.compileModifierInsForDateTime(node, m)

	} else {
		err = makeErr(node, fmt.Sprintf("Unknown/invalid interpolation modifier (%s)", mod))
	}

	return
}

func (c *Compiler) compileModifierInsForAlignment(node ast.Node, m []string) (
	ins opcode.Instructions, err error) {

	names := modifierRegexForAlign.SubexpNames()

	var align *object.Number
	align, err = object.NumberFromString(subMatchByName("align", m, names))
	if err != nil {
		err = makeErr(node, "Error processing code point alignment number for interpolation modifier")
		return
	}

	cp := ' ' // default to ASCII space
	withCp := subMatchByName("withcp", m, names)
	withCpNum := subMatchByName("withcpnum", m, names)
	if withCp != "" {
		cp, _, err = cpoint.Decode(&withCp, 0)
	} else if withCpNum != "" {
		cp, err = str.StrToRune(withCpNum, 16)
	}

	ins = append(ins, c.constantIns(align)...)
	ins = append(ins, c.constantIns(object.NumberFromInt(int(cp)))...)
	ins = append(ins, opcode.Make(opcode.OpFormat, format.FORMAT_ALIGN)...)

	return
}

func (c *Compiler) compileModifierInsForLimit(node ast.Node, m []string) (
	ins opcode.Instructions, err error) {

	names := modifierRegexForLimit.SubexpNames()

	var limit *object.Number
	limit, err = object.NumberFromString(subMatchByName("limit", m, names))
	if err != nil {
		err = makeErr(node, "Error processing limit number for interpolation modifier")
		return
	}

	// overflow indicator
	// default to zero-length string
	internalOverflowIndicator := subMatchByName("internal", m, names)

	ins = append(ins, c.constantIns(limit)...)
	ins = append(ins, c.constantIns(object.NewString(internalOverflowIndicator))...)
	ins = append(ins, opcode.Make(opcode.OpFormat, format.FORMAT_LIMIT)...)

	return
}

func (c *Compiler) compileModifierInsForLimitGraphemes(node ast.Node, m []string) (
	ins opcode.Instructions, err error) {

	names := modifierRegexForLimitGraphemes.SubexpNames()

	var limit *object.Number
	limit, err = object.NumberFromString(subMatchByName("limit", m, names))
	if err != nil {
		err = makeErr(node, "Error processing grapheme limit number for interpolation modifier")
		return
	}

	// overflow indicator
	// default to zero-length string
	internalOverflowIndicator := subMatchByName("internal", m, names)

	ins = append(ins, c.constantIns(limit)...)
	ins = append(ins, c.constantIns(object.NewString(internalOverflowIndicator))...)
	ins = append(ins, opcode.Make(opcode.OpFormat, format.FORMAT_LIMIT_GRAPHEMES)...)

	return
}

func (c *Compiler) compileModifierInsForCustomFn(node ast.Node, m []string) (
	ins opcode.Instructions, err error) {

	names := modifierRegexForFn.SubexpNames()

	// custom formatting function
	var customFn opcode.Instructions
	customFn, err = c.resolveAndGetInstructions(node, subMatchByName("fn", m, names))
	if err != nil {
		err = makeErr(node, fmt.Sprintf("Error processing custom function retrieval for interpolation modifier: %s", err))
		return
	}
	ins = append(ins, customFn...)
	ins = append(ins, opcode.Make(opcode.OpCall, 1)...)

	return
}

func (c *Compiler) compileModifierInsForTruncate(node ast.Node, m []string) (
	ins opcode.Instructions, err error) {

	names := modifierRegexForTruncate.SubexpNames()

	var max *object.Number
	mm := subMatchByName("max", m, names)
	if mm == "" {
		max = object.NumberFromInt(0)
	} else {
		max, err = object.NumberFromString(mm)
	}
	if err != nil {
		err = makeErr(node, fmt.Sprintf("Error processing truncate maximum digits for interpolation modifier: %s", err))
		return
	}

	ins = append(ins, c.constantIns(max)...)

	if subMatchByName("trimTrailingZeroes", m, names) == "-" {
		ins = append(ins, opcode.Make(opcode.OpTrue)...)
	} else {
		ins = append(ins, opcode.Make(opcode.OpFalse)...)
	}

	ins = append(ins, opcode.Make(opcode.OpFormat, format.FORMAT_TRUNCATE)...)

	return
}

func (c *Compiler) compileModifierInsForRounding(node ast.Node, m []string) (
	ins opcode.Instructions, err error) {

	names := modifierRegexForRounding.SubexpNames()

	var max *object.Number
	mm := subMatchByName("max", m, names)
	if mm == "" {
		max = object.NumberFromInt(0)
	} else {
		max, err = object.NumberFromString(mm)
	}
	if err != nil {
		err = makeErr(node, fmt.Sprintf("Error processing rounding maximum digits for interpolation modifier: %s", err))
		return
	}

	ins = append(ins, c.constantIns(max)...)

	if subMatchByName("trimTrailingZeroes", m, names) == "-" {
		ins = append(ins, opcode.Make(opcode.OpTrue)...)
	} else {
		ins = append(ins, opcode.Make(opcode.OpFalse)...)
	}

	ins = append(ins, opcode.Make(opcode.OpFormat, format.FORMAT_ROUND)...)

	return
}

func (c *Compiler) compileModifierInsForHex(node ast.Node, m []string) (
	ins opcode.Instructions, err error) {

	names := modifierRegexForHex.SubexpNames()

	padWithZeroes := false

	var min *object.Number
	mm := subMatchByName("min", m, names)
	if mm == "" {
		min, err = object.NumberFromString("0")
	} else {
		if mm[0] == '0' {
			padWithZeroes = true
		}
		min, err = object.NumberFromString(mm)
	}
	if err != nil {
		err = makeErr(node, fmt.Sprintf("Error processing hex minimum for interpolation modifier: %s", err))
		return
	}

	// add codes to stack
	ins = append(ins, c.constantIns(min)...)
	if subMatchByName("uc", m, names) == "X" {
		// uppercase
		ins = append(ins, opcode.Make(opcode.OpTrue)...)
	} else {
		ins = append(ins, opcode.Make(opcode.OpFalse)...)
	}
	if subMatchByName("sign", m, names) == "+" {
		ins = append(ins, opcode.Make(opcode.OpTrue)...)
	} else {
		ins = append(ins, opcode.Make(opcode.OpFalse)...)
	}

	if padWithZeroes {
		ins = append(ins, opcode.Make(opcode.OpTrue)...)
	} else {
		ins = append(ins, opcode.Make(opcode.OpFalse)...)
	}

	ins = append(ins, opcode.Make(opcode.OpFormat, format.FORMAT_HEX)...)

	return
}

func (c *Compiler) compileModifierInsForCustomBase(node ast.Node, m []string) (
	ins opcode.Instructions, err error) {

	names := modifierRegexForBase.SubexpNames()

	// custom base specified (from 2 to 36)
	var base *object.Number
	base, err = object.NumberFromString(subMatchByName("base", m, names))
	if err != nil {
		err = makeErr(node, fmt.Sprintf("Error processing base for interpolation modifier: %s", err))
		return
	}

	padWithZeroes := false

	var min *object.Number
	mm := subMatchByName("min", m, names)
	if mm == "" {
		min, err = object.NumberFromString("0")
	} else {
		if mm[0] == '0' {
			padWithZeroes = true
		}
		min, err = object.NumberFromString(mm)
	}
	if err != nil {
		err = makeErr(node, fmt.Sprintf("Error processing base minimum width for interpolation modifier: %s", err))
		return
	}

	// add codes to stack
	ins = append(ins, c.constantIns(base)...)
	ins = append(ins, c.constantIns(min)...)
	if subMatchByName("uc", m, names) == "X" {
		// uppercase
		ins = append(ins, opcode.Make(opcode.OpTrue)...)
	} else {
		ins = append(ins, opcode.Make(opcode.OpFalse)...)
	}
	if subMatchByName("sign", m, names) == "+" {
		ins = append(ins, opcode.Make(opcode.OpTrue)...)
	} else {
		ins = append(ins, opcode.Make(opcode.OpFalse)...)
	}

	if padWithZeroes {
		ins = append(ins, opcode.Make(opcode.OpTrue)...)
	} else {
		ins = append(ins, opcode.Make(opcode.OpFalse)...)
	}

	ins = append(ins, opcode.Make(opcode.OpFormat, format.FORMAT_BASE)...)

	return
}

func (c *Compiler) compileModifierInsForFixedNotation(node ast.Node, m []string) (
	ins opcode.Instructions, err error) {

	names := modifierRegexForFixed.SubexpNames()

	// fixed point
	var integer, frac *object.Number

	padIntWithZeroes := false

	mm := subMatchByName("int", m, names)
	integer, err = object.NumberFromString(mm)
	if err != nil {
		err = makeErr(node, fmt.Sprintf("Error processing integer for fixed point interpolation modifier: %s", err))
		return
	}
	if mm[0] == '0' {
		padIntWithZeroes = true
	}
	mm = subMatchByName("frac", m, names)
	frac, err = object.NumberFromString(mm)
	if err != nil {
		err = makeErr(node, fmt.Sprintf("Error processing fractional for fixed point interpolation modifier: %s", err))
		return
	}

	if subMatchByName("sign", m, names) == "+" {
		ins = append(ins, opcode.Make(opcode.OpTrue)...)
	} else {
		ins = append(ins, opcode.Make(opcode.OpFalse)...)
	}
	ins = append(ins, c.constantIns(integer)...)
	ins = append(ins, c.constantIns(frac)...)

	if padIntWithZeroes {
		ins = append(ins, opcode.Make(opcode.OpTrue)...)
	} else {
		ins = append(ins, opcode.Make(opcode.OpFalse)...)
	}

	if subMatchByName("trimTrailingZeroes", m, names) == "-" {
		ins = append(ins, opcode.Make(opcode.OpTrue)...)
	} else {
		ins = append(ins, opcode.Make(opcode.OpFalse)...)
	}

	ins = append(ins, opcode.Make(opcode.OpFormat, format.FORMAT_FIXED)...)

	return
}

func (c *Compiler) compileModifierInsForScientificNotation(node ast.Node, m []string) (
	ins opcode.Instructions, err error) {

	names := modifierRegexForScientificNotation.SubexpNames()

	var scale object.Object
	mm := subMatchByName("scale", m, names)
	if mm == "" {
		// not using -1 here, as we use negative numbers to mean something else
		scale = object.FALSE
	} else {
		scale, err = object.NumberFromString(mm)
	}
	if err != nil {
		err = makeErr(node, fmt.Sprintf("Error processing scale for e-notation interpolation modifier: %s", err))
		return
	}

	var scaleExp object.Object
	mm = subMatchByName("scaleExp", m, names)
	if mm == "" {
		scaleExp = object.NumberFromInt(1)
	} else {
		scaleExp, err = object.NumberFromString(mm)
	}
	if err != nil {
		err = makeErr(node, fmt.Sprintf("Error processing exponent scale for e-notation interpolation modifier: %s", err))
		return
	}

	// add codes to stack in order
	// sign required?
	if subMatchByName("sign", m, names) == "+" {
		ins = append(ins, opcode.Make(opcode.OpTrue)...)
	} else {
		ins = append(ins, opcode.Make(opcode.OpFalse)...)
	}

	// scale
	ins = append(ins, c.constantIns(scale)...)

	// include trailing zeroes on scale?
	if subMatchByName("scaleTrimTrailingZeroes", m, names) == "-" {
		ins = append(ins, opcode.Make(opcode.OpTrue)...)
	} else {
		ins = append(ins, opcode.Make(opcode.OpFalse)...)
	}

	// capitalize E?
	if subMatchByName("uc", m, names) == "E" {
		// uppercase
		ins = append(ins, opcode.Make(opcode.OpTrue)...)
	} else {
		ins = append(ins, opcode.Make(opcode.OpFalse)...)
	}

	// exponent sign required?
	if subMatchByName("expsign", m, names) == "+" {
		ins = append(ins, opcode.Make(opcode.OpTrue)...)
	} else {
		ins = append(ins, opcode.Make(opcode.OpFalse)...)
	}

	ins = append(ins, c.constantIns(scaleExp)...)
	ins = append(ins, opcode.Make(opcode.OpFormat, format.FORMAT_SCIENTIFIC_NOTATION)...)

	return
}

func (c *Compiler) compileModifierInsForDateTime(node ast.Node, m []string) (
	ins opcode.Instructions, err error) {

	names := modifierRegexForDateTime.SubexpNames()

	dtvar := subMatchByName("var", m, names)
	if dtvar == "" {
		// format string
		ins = append(ins, c.constantIns(object.NewString(subMatchByName("format", m, names)))...)

	} else {
		// a variable to use as a format string
		var fmtStringVar opcode.Instructions
		fmtStringVar, err = c.resolveAndGetInstructions(node, dtvar)
		if err != nil {
			err = makeErr(node, fmt.Sprintf("Error processing variable retrieval for date-time interpolation modifier: %s", err))
			return
		}
		ins = append(ins, fmtStringVar...)
	}
	ins = append(ins, opcode.Make(opcode.OpFormat, format.FORMAT_DATE_TIME)...)

	return
}
