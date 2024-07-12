// langur/compiler/interpolation_modifiers.go

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

// NOTE: Coordinate these things with the use of opFormat (in the VM).

var modifierRegexForFn = regexp.MustCompile(`^` + common.FunctionTokenLiteral +
	` (?P<fn>` + common.IdentifierRegexString + `)$`)

// var modifierRegexForFn = regexp.MustCompile(`^-> (?P<fn>` + common.IdentifierRegexString + `)$`)

var modifierRegexForDateTime = regexp.MustCompile(
	`^` + common.DateTimeTokenLiteral + `\((?P<format>.+)\)$|^` +
		common.DateTimeTokenLiteral + ` (?P<var>` + common.IdentifierRegexString + `)$`)

var modifierRegexForTruncate = regexp.MustCompile(
	`^t(?:(?P<max>-?[0-9]+)(?P<trailingZeroes>[!\-])?)?$`)
var modifierRegexForRounding = regexp.MustCompile(
	`^r(?:(?P<max>-?[0-9]+)(?P<trailingZeroes>[!\-])?)?$`)

var modifierRegexForAlign = regexp.MustCompile(
	`^(?P<align>-?[1-9][0-9]*)(?:\((?:(?P<withcp>.)|(?P<withcpnum>[0-9a-fA-F]{2,8}))\))?$`)

// var modifierRegexForAlignGraphemes = regexp.MustCompile(
// 	`^g(?P<align>-?[1-9][0-9]*)(?:\((?:(?P<withcp>.)|(?P<withcpnum>[0-9a-fA-F]{2,8}))\))?$`)

var modifierRegexForLimit = regexp.MustCompile(
	`^L(?P<limit>-?[1-9][0-9]*)(?:\((?P<internal>[^)]*)\))?$`)

var modifierRegexForLimitGraphemes = regexp.MustCompile(
	`^Lg(?P<limit>-?[1-9][0-9]*)(?:\((?P<internal>[^)]*)\))?$`)

var modifierRegexForFixed = regexp.MustCompile(`(?x)
	^
	(?P<sign>[+])?
	(?P<base>[1-9][0-9]*)?
	(?P<uc>[xX])
	(?:(?P<padint>0)?(?P<int>[0-9]+))?
	(?:(?P<point>[.,])(?P<frac>[0-9]+)(?P<trailingZeroes>[!\-])?)?
	$`)

var modifierRegexForScientificNotation = regexp.MustCompile(`(?x)
	^
	(?P<sign>[+])?
	(?:1(?P<point>[.,])(?P<scale>\d+)(?P<trailingZeroes>[!\-])?)?
	(?P<uc>[eE])
	(?P<expsign>[+])?
	(?P<scaleExp>\d+)?
	$`)

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

	// add codes in order
	// truncate to
	ins = append(ins, c.constantIns(max)...)

	// add trailing zeroes?
	trailing := subMatchByName("trailingZeroes", m, names)
	if trailing == "!" {
		ins = append(ins, opcode.Make(opcode.OpFalse)...)
	} else {
		ins = append(ins, opcode.Make(opcode.OpTrue)...)
	}

	// trim trailing zeroes?
	if trailing == "-" {
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

	// add codes in order
	// round to
	ins = append(ins, c.constantIns(max)...)

	// add trailing zeroes?
	trailing := subMatchByName("trailingZeroes", m, names)
	if trailing == "!" {
		ins = append(ins, opcode.Make(opcode.OpFalse)...)
	} else {
		ins = append(ins, opcode.Make(opcode.OpTrue)...)
	}

	// trim trailing zeroes?
	if trailing == "-" {
		ins = append(ins, opcode.Make(opcode.OpTrue)...)
	} else {
		ins = append(ins, opcode.Make(opcode.OpFalse)...)
	}

	ins = append(ins, opcode.Make(opcode.OpFormat, format.FORMAT_ROUND)...)

	return
}

func (c *Compiler) compileModifierInsForFixedNotation(node ast.Node, m []string) (
	ins opcode.Instructions, err error) {

	var integer, frac *object.Number

	names := modifierRegexForFixed.SubexpNames()

	// base from 2 to 36
	b := subMatchByName("base", m, names)
	if b == "" {
		// none specified; is hexadecimal
		b = "16"
	}
	var base *object.Number
	base, err = object.NumberFromString(b)
	if err != nil {
		err = makeErr(node, fmt.Sprintf("Error processing base for interpolation modifier: %s", err))
		return
	}

	padIntWithZeroes := false
	mm := subMatchByName("padint", m, names)
	if mm == "0" {
		padIntWithZeroes = true
	}

	mm = subMatchByName("int", m, names)
	if mm == "" {
		mm = "1"
	} else if mm == "0" {
		err = makeErr(node, fmt.Sprintf("Error processing integer for fixed point interpolation modifier: integer cannot be 0 (maybe you meant 1?)"))
		return
	}
	integer, err = object.NumberFromString(mm)
	if err != nil {
		err = makeErr(node, fmt.Sprintf("Error processing integer for fixed point interpolation modifier: %s", err))
		return
	}

	mm = subMatchByName("frac", m, names)
	if mm == "" {
		frac = object.Zero

	} else {
		if b != "10" {
			err = makeErr(node, "Fractional only valid with base 10")
			return
		}

		frac, err = object.NumberFromString(mm)
		if err != nil {
			err = makeErr(node, fmt.Sprintf("Error processing fractional for fixed point interpolation modifier: %s", err))
			return
		}
	}
	trailing := subMatchByName("trailingZeroes", m, names)

	point := object.NumberFromRune('.')
	p := subMatchByName("point", m, names)
	if p != "" {
		point = object.NumberFromInt(int(p[0]))
	}

	// add codes in order
	// sign required?
	if subMatchByName("sign", m, names) == "+" {
		ins = append(ins, opcode.Make(opcode.OpTrue)...)
	} else {
		ins = append(ins, opcode.Make(opcode.OpFalse)...)
	}

	// base
	ins = append(ins, c.constantIns(base)...)

	// uppercase?
	if subMatchByName("uc", m, names) == "X" {
		ins = append(ins, opcode.Make(opcode.OpTrue)...)
	} else {
		ins = append(ins, opcode.Make(opcode.OpFalse)...)
	}

	// integer, decimal point, and fractional
	ins = append(ins, c.constantIns(integer)...)
	ins = append(ins, c.constantIns(point)...)
	ins = append(ins, c.constantIns(frac)...)

	// pad left (integer portion) with zeroes?
	if padIntWithZeroes {
		ins = append(ins, opcode.Make(opcode.OpTrue)...)
	} else {
		ins = append(ins, opcode.Make(opcode.OpFalse)...)
	}

	// add trailing zeroes?
	if trailing == "!" {
		ins = append(ins, opcode.Make(opcode.OpFalse)...)
	} else {
		ins = append(ins, opcode.Make(opcode.OpTrue)...)
	}

	// trim trailing zeroes?
	if trailing == "-" {
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

	point := object.NumberFromRune('.')
	p := subMatchByName("point", m, names)
	if p != "" {
		point = object.NumberFromInt(int(p[0]))
	}

	var scaleExp object.Object
	mm = subMatchByName("scaleExp", m, names)
	if mm == "" {
		scaleExp = object.One
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

	// decimal point
	ins = append(ins, c.constantIns(point)...)

	// scale
	ins = append(ins, c.constantIns(scale)...)

	// add trailing zeroes on scale?
	trailing := subMatchByName("trailingZeroes", m, names)
	if trailing == "!" {
		ins = append(ins, opcode.Make(opcode.OpFalse)...)
	} else {
		ins = append(ins, opcode.Make(opcode.OpTrue)...)
	}

	// trim trailing zeroes on scale?
	if trailing == "-" {
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
