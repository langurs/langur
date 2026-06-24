// langur/ast/compiler_interpolation_modifiers.go

package ast

import (
	"fmt"
	"langur/common"
	"langur/cpoint"
	"langur/format"
	"langur/object"
	"langur/opcode"
	"langur/pattern"
	"langur/regexp"
	"langur/str"
)

// NOTE: Coordinate these things with the use of opFormat (in the VM).

var modifierPatternForFn = regexp.MustCompile(`^` + common.FunctionTokenLiteral +
	` (?P<fn>` + common.IdentifierPatternString + `)$`)

// var modifierPatternForFn = regexp.MustCompile(`^-> (?P<fn>` + common.IdentifierPatternString + `)$`)

var modifierPatternForDateTime = regexp.MustCompile(
	`^` + common.DateTimeTokenLiteral + `\((?P<format>.+)\)$|^` +
		common.DateTimeTokenLiteral + ` (?P<var>` + common.IdentifierPatternString + `)$`)

var modifierPatternForTruncate = regexp.MustCompile(
	`^t(?:(?P<max>-?[0-9]+)(?P<trailingZeroes>[!\-])?)?$`)
var modifierPatternForRounding = regexp.MustCompile(
	`^r(?:(?P<max>-?[0-9]+)(?P<trailingZeroes>[!\-])?)?$`)

var modifierPatternForAlign = regexp.MustCompile(
	`^(?P<align>-?[1-9][0-9]*)(?:\((?:(?P<withcp>.)|(?P<withcpnum>[0-9a-fA-F]{2,8}))\))?$`)

// var modifierPatternForAlignGraphemes = regexp.MustCompile(
// 	`^g(?P<align>-?[1-9][0-9]*)(?:\((?:(?P<withcp>.)|(?P<withcpnum>[0-9a-fA-F]{2,8}))\))?$`)

var modifierPatternForLimit = regexp.MustCompile(
	`^L(?P<limit>-?[1-9][0-9]*)(?:\((?P<internal>[^)]*)\))?$`)

var modifierPatternForLimitGraphemes = regexp.MustCompile(
	`^Lg(?P<limit>-?[1-9][0-9]*)(?:\((?P<internal>[^)]*)\))?$`)

var modifierPatternForFixed = regexp.MustCompile(`(?x)
	^
	(?P<sign>[+])?
	(?P<base>[1-9][0-9]*)?
	(?P<uc>[xX])
	(?:(?P<padint>0)?(?P<int>[0-9]+))?
	(?:(?P<point>[.,])(?P<frac>[0-9]+)(?P<trailingZeroes>[!\-])?)?
	$`)

var modifierPatternForScientificNotation = regexp.MustCompile(`(?x)
	^
	(?P<sign>[+])?
	(?:
		# scientific notation without explicitly being engineering notation
		(?P<first>1)(?P<point>[.,])(?P<scale>\d+)(?P<trailingZeroes>[!\-])?
		|
		# engineering notation; parts after "3" optional; eng. notation not the default
		(?P<first>3)((?P<point>[.,])?(?P<trailingZeroes>[!\-])?)?
	)?
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

func (c *Compiler) compileInterpolationModifiers(node Node, modifiers []string, patternType pattern.PatternType) (
	pkg opcode.InsPackage, err error) {

	for _, mod := range modifiers {
		temp, err := c.compileInterpolationModifierIns(node, mod, patternType)
		if err != nil {
			return temp, err
		}
		
		pkg = pkg.Append(temp)
	}

	return
}

func (c *Compiler) compileInterpolationModifierIns(
	node Node, mod string, patternType pattern.PatternType) (
	pkg opcode.InsPackage, err error) {

	if mod == format.MODSTRING_ESCAPE {
		patternType := object.NumberFromInt(int(patternType))
		pkg = pkg.Append(c.constantIns(patternType))
		pkg = pkg.Append(opcode.MakePkg(node.TokenInfo(), opcode.OpFormat, format.FORMAT_ESCAPE))

	} else if mod == "T" {
		pkg = pkg.Append(opcode.MakePkg(node.TokenInfo(), opcode.OpFormat, format.FORMAT_TYPE))

	} else if mod == "cp" {
		pkg = pkg.Append(opcode.MakePkg(node.TokenInfo(), opcode.OpFormat, format.FORMAT_CODE_POINT))

	} else if m := modifierPatternForAlign.FindStringSubmatch(mod); m != nil {
		return c.compileModifierInsForAlignment(node, m)

		// } else if m := modifierPatternForAlignGraphemes.FindStringSubmatch(mod); m != nil {
		// 	return c.compileModifierInsForAlignmentGraphemes(node, m)

	} else if m := modifierPatternForLimit.FindStringSubmatch(mod); m != nil {
		return c.compileModifierInsForLimit(node, m)

	} else if m := modifierPatternForLimitGraphemes.FindStringSubmatch(mod); m != nil {
		return c.compileModifierInsForLimitGraphemes(node, m)

	} else if m := modifierPatternForTruncate.FindStringSubmatch(mod); m != nil {
		return c.compileModifierInsForTruncate(node, m)

	} else if m := modifierPatternForRounding.FindStringSubmatch(mod); m != nil {
		return c.compileModifierInsForRounding(node, m)

	} else if m := modifierPatternForFn.FindStringSubmatch(mod); m != nil {
		return c.compileModifierInsForCustomFn(node, m)

	} else if m := modifierPatternForFixed.FindStringSubmatch(mod); m != nil {
		return c.compileModifierInsForFixedNotation(node, m)

	} else if m := modifierPatternForScientificNotation.FindStringSubmatch(mod); m != nil {
		return c.compileModifierInsForScientificNotation(node, m)

	} else if m := modifierPatternForDateTime.FindStringSubmatch(mod); m != nil {
		return c.compileModifierInsForDateTime(node, m)

	} else {
		err = c.makeErr(node, fmt.Sprintf("Unknown/invalid interpolation modifier (%s)", mod))
	}

	return
}

func (c *Compiler) compileModifierInsForAlignment(node Node, m []string) (
	pkg opcode.InsPackage, err error) {

	names := modifierPatternForAlign.SubexpNames()

	var align *object.Number
	align, err = object.NumberFromString(subMatchByName("align", m, names))
	if err != nil {
		err = c.makeErr(node, "Error processing code point alignment number for interpolation modifier")
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

	pkg = pkg.Append(c.constantIns(align))
	pkg = pkg.Append(c.constantIns(object.NumberFromInt(int(cp))))
	pkg = pkg.Append(opcode.MakePkg(node.TokenInfo(), opcode.OpFormat, format.FORMAT_ALIGN))

	return
}

func (c *Compiler) compileModifierInsForLimit(node Node, m []string) (
	pkg opcode.InsPackage, err error) {

	names := modifierPatternForLimit.SubexpNames()

	var limit *object.Number
	limit, err = object.NumberFromString(subMatchByName("limit", m, names))
	if err != nil {
		err = c.makeErr(node, "Error processing limit number for interpolation modifier")
		return
	}

	// overflow indicator
	// default to zero-length string
	internalOverflowIndicator := subMatchByName("internal", m, names)

	pkg = pkg.Append(c.constantIns(limit))
	pkg = pkg.Append(c.constantIns(object.NewString(internalOverflowIndicator)))
	pkg = pkg.Append(opcode.MakePkg(node.TokenInfo(), opcode.OpFormat, format.FORMAT_LIMIT))

	return
}

func (c *Compiler) compileModifierInsForLimitGraphemes(node Node, m []string) (
	pkg opcode.InsPackage, err error) {

	names := modifierPatternForLimitGraphemes.SubexpNames()

	var limit *object.Number
	limit, err = object.NumberFromString(subMatchByName("limit", m, names))
	if err != nil {
		err = c.makeErr(node, "Error processing grapheme limit number for interpolation modifier")
		return
	}

	// overflow indicator
	// default to zero-length string
	internalOverflowIndicator := subMatchByName("internal", m, names)

	pkg = pkg.Append(c.constantIns(limit))
	pkg = pkg.Append(c.constantIns(object.NewString(internalOverflowIndicator)))
	pkg = pkg.Append(opcode.MakePkg(node.TokenInfo(), opcode.OpFormat, format.FORMAT_LIMIT_GRAPHEMES))

	return
}

func (c *Compiler) compileModifierInsForCustomFn(node Node, m []string) (
	pkg opcode.InsPackage, err error) {

	names := modifierPatternForFn.SubexpNames()

	// custom formatting function
	var customFn opcode.InsPackage
	customFn, err = c.resolveAndGetInstructions(node, subMatchByName("fn", m, names))
	if err != nil {
		err = c.makeErr(node, fmt.Sprintf("Error processing custom function retrieval for interpolation modifier: %s", err))
		return
	}
	pkg = pkg.Append(customFn)
	pkg = pkg.Append(opcode.MakePkg(node.TokenInfo(), opcode.OpCall, 1, 0))

	return
}

func (c *Compiler) compileModifierInsForTruncate(node Node, m []string) (
	pkg opcode.InsPackage, err error) {

	names := modifierPatternForTruncate.SubexpNames()

	var max *object.Number
	mm := subMatchByName("max", m, names)
	if mm == "" {
		max = object.NumberFromInt(0)
	} else {
		max, err = object.NumberFromString(mm)
	}
	if err != nil {
		err = c.makeErr(node, fmt.Sprintf("Error processing truncate maximum digits for interpolation modifier: %s", err))
		return
	}

	// add codes in order
	// truncate to
	pkg = pkg.Append(c.constantIns(max))

	// add trailing zeroes?
	trailing := subMatchByName("trailingZeroes", m, names)
	if trailing == "!" {
		pkg = pkg.Append(opcode.MakePkg(node.TokenInfo(), opcode.OpFalse))
	} else {
		pkg = pkg.Append(opcode.MakePkg(node.TokenInfo(), opcode.OpTrue))
	}

	// trim trailing zeroes?
	if trailing == "-" {
		pkg = pkg.Append(opcode.MakePkg(node.TokenInfo(), opcode.OpTrue))
	} else {
		pkg = pkg.Append(opcode.MakePkg(node.TokenInfo(), opcode.OpFalse))
	}

	pkg = pkg.Append(opcode.MakePkg(node.TokenInfo(), opcode.OpFormat, format.FORMAT_TRUNCATE))

	return
}

func (c *Compiler) compileModifierInsForRounding(node Node, m []string) (
	pkg opcode.InsPackage, err error) {

	names := modifierPatternForRounding.SubexpNames()

	var max *object.Number
	mm := subMatchByName("max", m, names)
	if mm == "" {
		max = object.NumberFromInt(0)
	} else {
		max, err = object.NumberFromString(mm)
	}
	if err != nil {
		err = c.makeErr(node, fmt.Sprintf("Error processing rounding maximum digits for interpolation modifier: %s", err))
		return
	}

	// add codes in order
	// round to
	pkg = pkg.Append(c.constantIns(max))

	// add trailing zeroes?
	trailing := subMatchByName("trailingZeroes", m, names)
	if trailing == "!" {
		pkg = pkg.Append(opcode.MakePkg(node.TokenInfo(), opcode.OpFalse))
	} else {
		pkg = pkg.Append(opcode.MakePkg(node.TokenInfo(), opcode.OpTrue))
	}

	// trim trailing zeroes?
	if trailing == "-" {
		pkg = pkg.Append(opcode.MakePkg(node.TokenInfo(), opcode.OpTrue))
	} else {
		pkg = pkg.Append(opcode.MakePkg(node.TokenInfo(), opcode.OpFalse))
	}

	pkg = pkg.Append(opcode.MakePkg(node.TokenInfo(), opcode.OpFormat, format.FORMAT_ROUND))

	return
}

func (c *Compiler) compileModifierInsForFixedNotation(node Node, m []string) (
	pkg opcode.InsPackage, err error) {

	var integer, frac *object.Number

	names := modifierPatternForFixed.SubexpNames()

	// base from 2 to 36
	b := subMatchByName("base", m, names)
	if b == "" {
		// none specified; is hexadecimal
		b = "16"
	}
	var base *object.Number
	base, err = object.NumberFromString(b)
	if err != nil {
		err = c.makeErr(node, fmt.Sprintf("Error processing base for interpolation modifier: %s", err))
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
		err = c.makeErr(node, fmt.Sprintf("Error processing integer for fixed point interpolation modifier: integer cannot be 0 (maybe you meant 1?)"))
		return
	}
	integer, err = object.NumberFromString(mm)
	if err != nil {
		err = c.makeErr(node, fmt.Sprintf("Error processing integer for fixed point interpolation modifier: %s", err))
		return
	}

	mm = subMatchByName("frac", m, names)
	if mm == "" {
		frac = object.Zero

	} else {
		if b != "10" {
			err = c.makeErr(node, "Fractional only valid with base 10")
			return
		}

		frac, err = object.NumberFromString(mm)
		if err != nil {
			err = c.makeErr(node, fmt.Sprintf("Error processing fractional for fixed point interpolation modifier: %s", err))
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
		pkg = pkg.Append(opcode.MakePkg(node.TokenInfo(), opcode.OpTrue))
	} else {
		pkg = pkg.Append(opcode.MakePkg(node.TokenInfo(), opcode.OpFalse))
	}

	// base
	pkg = pkg.Append(c.constantIns(base))

	// uppercase?
	if subMatchByName("uc", m, names) == "X" {
		pkg = pkg.Append(opcode.MakePkg(node.TokenInfo(), opcode.OpTrue))
	} else {
		pkg = pkg.Append(opcode.MakePkg(node.TokenInfo(), opcode.OpFalse))
	}

	// integer, decimal point, and fractional
	pkg = pkg.Append(c.constantIns(integer))
	pkg = pkg.Append(c.constantIns(point))
	pkg = pkg.Append(c.constantIns(frac))

	// pad left (integer portion) with zeroes?
	if padIntWithZeroes {
		pkg = pkg.Append(opcode.MakePkg(node.TokenInfo(), opcode.OpTrue))
	} else {
		pkg = pkg.Append(opcode.MakePkg(node.TokenInfo(), opcode.OpFalse))
	}

	// add trailing zeroes?
	if trailing == "!" {
		pkg = pkg.Append(opcode.MakePkg(node.TokenInfo(), opcode.OpFalse))
	} else {
		pkg = pkg.Append(opcode.MakePkg(node.TokenInfo(), opcode.OpTrue))
	}

	// trim trailing zeroes?
	if trailing == "-" {
		pkg = pkg.Append(opcode.MakePkg(node.TokenInfo(), opcode.OpTrue))
	} else {
		pkg = pkg.Append(opcode.MakePkg(node.TokenInfo(), opcode.OpFalse))
	}

	pkg = pkg.Append(opcode.MakePkg(node.TokenInfo(), opcode.OpFormat, format.FORMAT_FIXED))

	return
}

func (c *Compiler) compileModifierInsForScientificNotation(node Node, m []string) (
	pkg opcode.InsPackage, err error) {

	names := modifierPatternForScientificNotation.SubexpNames()

	var scale object.Object
	mm := subMatchByName("scale", m, names)
	if mm == "" {
		// not using -1 here, as we use negative numbers to mean something else
		scale = object.FALSE
	} else {
		scale, err = object.NumberFromString(mm)
	}
	if err != nil {
		err = c.makeErr(node, fmt.Sprintf("Error processing scale for scientific interpolation modifier: %s", err))
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
		err = c.makeErr(node, fmt.Sprintf("Error processing exponent scale for scientific interpolation modifier: %s", err))
		return
	}

	// add codes to stack in order
	// sign required?
	if subMatchByName("sign", m, names) == "+" {
		pkg = pkg.Append(opcode.MakePkg(node.TokenInfo(), opcode.OpTrue))
	} else {
		pkg = pkg.Append(opcode.MakePkg(node.TokenInfo(), opcode.OpFalse))
	}

	// first digit; 3 indicating engineering notation
	// 1 or nothing for standard scientific notation
	first := subMatchByName("first", m, names)
	switch first {
	case "3":
		err = c.makeErr(node, "Engineering notation not set up yet")
		return
	}

	// decimal point
	pkg = pkg.Append(c.constantIns(point))

	// scale
	pkg = pkg.Append(c.constantIns(scale))

	// add trailing zeroes on scale?
	trailing := subMatchByName("trailingZeroes", m, names)
	if trailing == "!" {
		pkg = pkg.Append(opcode.MakePkg(node.TokenInfo(), opcode.OpFalse))
	} else {
		pkg = pkg.Append(opcode.MakePkg(node.TokenInfo(), opcode.OpTrue))
	}

	// trim trailing zeroes on scale?
	if trailing == "-" {
		pkg = pkg.Append(opcode.MakePkg(node.TokenInfo(), opcode.OpTrue))
	} else {
		pkg = pkg.Append(opcode.MakePkg(node.TokenInfo(), opcode.OpFalse))
	}

	// capitalize E?
	if subMatchByName("uc", m, names) == "E" {
		// uppercase
		pkg = pkg.Append(opcode.MakePkg(node.TokenInfo(), opcode.OpTrue))
	} else {
		pkg = pkg.Append(opcode.MakePkg(node.TokenInfo(), opcode.OpFalse))
	}

	// exponent sign required?
	if subMatchByName("expsign", m, names) == "+" {
		pkg = pkg.Append(opcode.MakePkg(node.TokenInfo(), opcode.OpTrue))
	} else {
		pkg = pkg.Append(opcode.MakePkg(node.TokenInfo(), opcode.OpFalse))
	}

	// exponent scale
	pkg = pkg.Append(c.constantIns(scaleExp))
	
	pkg = pkg.Append(opcode.MakePkg(node.TokenInfo(), opcode.OpFormat, format.FORMAT_SCIENTIFIC_NOTATION))

	return
}

func (c *Compiler) compileModifierInsForDateTime(node Node, m []string) (
	pkg opcode.InsPackage, err error) {

	names := modifierPatternForDateTime.SubexpNames()

	dtvar := subMatchByName("var", m, names)
	if dtvar == "" {
		// format string
		pkg = pkg.Append(c.constantIns(object.NewString(subMatchByName("format", m, names))))

	} else {
		// a variable to use as a format string
		var fmtStringVar opcode.InsPackage
		fmtStringVar, err = c.resolveAndGetInstructions(node, dtvar)
		if err != nil {
			err = c.makeErr(node, fmt.Sprintf("Error processing variable retrieval for date-time interpolation modifier: %s", err))
			return
		}
		pkg = pkg.Append(fmtStringVar)
	}
	pkg = pkg.Append(opcode.MakePkg(node.TokenInfo(), opcode.OpFormat, format.FORMAT_DATE_TIME))

	return
}
