// langur/vm/process/interpolation.go

package process

import (
	"fmt"
	"langur/format"
	"langur/modes"
	"langur/object"
	"langur/regex"
	"langur/str"
)

func (pr *Process) format(code int) (result object.Object, err error) {
	// used for string interpolation modifiers
	// must be coordinated with compiler.compileInterpolationModifiers()

	switch code {
	case format.FORMAT_TYPE:
		original := pr.pop()
		return object.NewString(original.TypeString()), nil

	case format.FORMAT_ALIGN:
		things := pr.popMultiple(3)
		withCp := things[2]
		alignment := things[1]
		original := things[0]

		cp, ok := object.NumberToRune(withCp)
		if !ok {
			err = fmt.Errorf("Unable to convert integer for alignment code point for interpolation")
			return
		}
		align, ok := object.NumberToInt(alignment)
		if !ok {
			err = fmt.Errorf("Unable to convert integer for alignment distance for interpolation")
			return
		}

		result = object.NewString(str.Pad(original.String(), align, cp))

	case format.FORMAT_LIMIT:
		things := pr.popMultiple(3)
		internal := things[2].String()
		limit := things[1]
		original := things[0]

		limits, ok := object.NumberToInt(limit)
		if !ok {
			err = fmt.Errorf("Unable to convert integer for limit code points for interpolation")
			return
		}

		result = object.NewString(str.Limit(original.String(), limits, internal))

	case format.FORMAT_LIMIT_GRAPHEMES:
		things := pr.popMultiple(3)
		internal := things[2].String()
		limit := things[1]
		original := things[0]

		limits, ok := object.NumberToInt(limit)
		if !ok {
			err = fmt.Errorf("Unable to convert integer for limit graphemes for interpolation")
			return
		}

		result = object.NewString(str.LimitGraphemes(original.String(), limits, internal))

	case format.FORMAT_TRUNCATE:
		things := pr.popMultiple(4)

		trimTrailingZeroes := things[3].(*object.Boolean).Value
		addTrailingZeroes := things[2].(*object.Boolean).Value
		max := things[1]
		original := things[0]

		m, ok := object.NumberToInt(max)
		if !ok {
			err = fmt.Errorf("Unable to convert integer for truncate max digits for interpolation")
			return
		}
		orig, ok := original.(*object.Number)
		if !ok {
			err = fmt.Errorf("Unable to convert number for truncate for interpolation")
			return
		}

		result, err = orig.Truncate(m, addTrailingZeroes, trimTrailingZeroes)

	case format.FORMAT_ROUND:
		things := pr.popMultiple(4)

		trimTrailingZeroes := things[3].(*object.Boolean).Value
		addTrailingZeroes := things[2].(*object.Boolean).Value
		max := things[1]
		original := things[0]

		m, ok := object.NumberToInt(max)
		if !ok {
			err = fmt.Errorf("Unable to convert integer for rounding max digits for interpolation")
			return
		}
		orig, ok := original.(*object.Number)
		if !ok {
			err = fmt.Errorf("Unable to convert number for rounding for interpolation")
			return
		}

		result, err = orig.RoundByMode(m, addTrailingZeroes, trimTrailingZeroes, modes.RoundingMode)

	case format.FORMAT_ESCAPE:
		things := pr.popMultiple(2)
		regexType := things[1]
		original := things[0]

		reType, ok := object.NumberToInt(regexType)
		if !ok {
			err = fmt.Errorf("Unable to convert integer for string/regex type for interpolation")
			return
		}
		result, err = object.EscString(original, regex.RegexType(reType))

	case format.FORMAT_HEX:
		things := pr.popMultiple(5)
		padWithZeroes := things[4].(*object.Boolean).Value
		requireSign := things[3].(*object.Boolean).Value
		uppercase := things[2].(*object.Boolean).Value
		minimum := things[1]
		original := things[0]

		addFractionalZeroes := true
		trimFractionalZeroes := false

		min, ok := object.NumberToInt(minimum)
		if !ok {
			err = fmt.Errorf("Unable to convert integer for hex minimum for interpolation")
			return
		}

		padWith := ' '
		if padWithZeroes {
			padWith = '0'
		}

		result, err = object.ToBaseString(
			original, uppercase, requireSign, true,
			addFractionalZeroes, trimFractionalZeroes, min, 0, 16, padWith)

	case format.FORMAT_BASE:
		things := pr.popMultiple(6)
		padWithZeroes := things[5].(*object.Boolean).Value
		requireSign := things[4].(*object.Boolean).Value
		uppercase := things[3].(*object.Boolean).Value
		minimum := things[2]
		base := things[1]
		original := things[0]

		addFractionalZeroes := true
		trimFractionalZeroes := false

		b, ok := object.NumberToInt(base)
		if !ok {
			err = fmt.Errorf("Unable to convert integer for base for interpolation")
			return
		}
		min, ok := object.NumberToInt(minimum)
		if !ok {
			err = fmt.Errorf("Unable to convert integer for base minimum width for interpolation")
			return
		}

		padWith := ' '
		if padWithZeroes {
			padWith = '0'
		}

		result, err = object.ToBaseString(
			original, uppercase, requireSign, true,
			addFractionalZeroes, trimFractionalZeroes, min, 0, b, padWith)

	case format.FORMAT_FIXED:
		things := pr.popMultiple(7)

		trimFractionalZeroes := things[6].(*object.Boolean).Value
		addFractionalZeroes := things[5].(*object.Boolean).Value
		padIntWithZeroes := things[4].(*object.Boolean).Value
		frac := things[3]
		integer := things[2]
		requireSign := things[1].(*object.Boolean).Value
		original := things[0]

		intMin, ok := object.NumberToInt(integer)
		if !ok {
			err = fmt.Errorf("Unable to convert integer minimum width for fixed point interpolation")
			return
		}
		fracRound, ok := object.NumberToInt(frac)
		if !ok {
			err = fmt.Errorf("Unable to convert fractional rounding for fixed point interpolation")
			return
		}

		padIntWith := ' '
		if padIntWithZeroes {
			padIntWith = '0'
		}

		result, err = object.ToBaseString(
			original, false, requireSign, true,
			addFractionalZeroes, trimFractionalZeroes, intMin, fracRound, 10, padIntWith)

	case format.FORMAT_SCIENTIFIC_NOTATION:
		things := pr.popMultiple(8)

		scaleExp := things[7]
		requireExpSign := things[6].(*object.Boolean).Value
		uppercase := things[5].(*object.Boolean).Value
		scaleTrimTrailingZeroes := things[4].(*object.Boolean).Value
		scaleAddTrailingZeroes := things[3].(*object.Boolean).Value
		scale := things[2]
		requireSign := things[1].(*object.Boolean).Value
		original := things[0]

		rescale := true
		sc := 0
		if scale == object.FALSE {
			rescale = false

		} else {
			var ok bool
			sc, ok = object.NumberToInt(scale)
			if !ok {
				err = fmt.Errorf("Unable to convert integer for scientific notation scale for interpolation")
				return
			}
		}

		scExp, ok := object.NumberToInt(scaleExp)
		if !ok {
			err = fmt.Errorf("Unable to convert integer for scientific notation exponent scale for interpolation")
			return
		}

		orig, ok := original.(*object.Number)
		if !ok {
			err = fmt.Errorf("Unable to convert number for scientific notation for interpolation")
			return
		}

		result = object.NewString(
			orig.ScientificNotation(uppercase, requireSign, requireExpSign, rescale, scaleAddTrailingZeroes, scaleTrimTrailingZeroes, sc, scExp))

	case format.FORMAT_CODE_POINT:
		rSlc, err := object.CodePointsToFlatRuneSlice(pr.pop())
		if err != nil {
			return nil, fmt.Errorf("Unable to convert for code point interpolation: %s", err.Error())
		}
		s, err := object.NewStringFromParts(rSlc)
		if err != nil {
			return nil, fmt.Errorf("Unable to convert for code point interpolation: %s", err.Error())
		}
		return s, nil

	case format.FORMAT_DATE_TIME:
		things := pr.popMultiple(2)
		dtformat := things[1].String()
		original := things[0]

		orig, ok := original.(*object.DateTime)
		if !ok {
			err = fmt.Errorf("Unable to convert to date-time for interpolation modification")
			return
		}

		return object.NewString(orig.FormatString(dtformat)), nil

	default:
		bug("format", "Invalid code for Format")
		err = fmt.Errorf("Invalid code for Format")
	}

	return
}
