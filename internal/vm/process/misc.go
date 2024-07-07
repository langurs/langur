// langur/vm/process/misc.go

package process

import (
	"fmt"
	"langur/format"
	"langur/modes"
	"langur/object"
	"langur/regex"
	"langur/str"
	"math"
	"os"
)

func (pr *Process) setMode(code int, setting object.Object) error {
	switch code {
	case modes.MODE_DIVISION_MAX_SCALE:
		i, ok := object.NumberToInt(setting)
		if !ok {
			return fmt.Errorf("Expected integer for mode division max scale, not %s", setting.TypeString())
		}
		if i < 0 || i > math.MaxInt32-2 {
			return fmt.Errorf("Integer %d for mode division max scale out of range", i)
		}
		pr.Modes.DivisionMaxScale = i
		// FIXME: not safe for concurrency
		return object.SetDivisionMaxScaleMode(i)

	case modes.MODE_ROUNDING:
		i, ok := object.NumberToInt(setting)
		if !ok {
			return fmt.Errorf("Expected rounding mode (from %s hash), not %s", modes.RoundHashName, setting.TypeString())
		}
		// ensure it's in the enumeration
		_, ok = modes.RoundHashModeNames[i]
		if !ok {
			return fmt.Errorf("Unknown rounding mode")
		}
		pr.Modes.Rounding = i
		// FIXME: not safe for concurrency
		modes.RoundingMode = i

	case modes.MODE_CONSOLE_TEXT_MODE:
		b, ok := setting.(*object.Boolean)
		if !ok {
			return fmt.Errorf("Expected Boolean for mode console text mode")
		}
		pr.Modes.ConsoleTextMode = b.Value

	case modes.MODE_NEW_FILE_PERMISSIONS:
		i, ok := object.NumberToInt(setting)
		if !ok {
			return fmt.Errorf("Expected integer for mode new file permissions, not %s", setting.TypeString())
		}
		if i < 0 || i > 0o777 {
			return fmt.Errorf("Integer 8x%s for mode new file permissions out of range", str.IntToStr(i, 8))
		}
		pr.Modes.NewFilePermissions = os.FileMode(i)

	case modes.MODE_NOW_INCLUDES_NANO:
		b, ok := setting.(*object.Boolean)
		if !ok {
			return fmt.Errorf("Expected Boolean for mode now includes nanoseconds")
		}
		pr.Modes.NowIncludesNano = b.Value

	default:
		bug("setMode", fmt.Sprintf("Unknown mode setting %d", code))
		return fmt.Errorf("Unknown mode setting %d", code)
	}

	return nil
}

func (pr *Process) format(code int) (result object.Object, err error) {
	// used for string interpolation modifiers
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
		things := pr.popMultiple(3)

		trimTrailingZeroes := things[2].(*object.Boolean).Value
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

		result, err = orig.Truncate(m, trimTrailingZeroes)

	case format.FORMAT_ROUND:
		things := pr.popMultiple(3)

		trimTrailingZeroes := things[2].(*object.Boolean).Value
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

		result, err = orig.Round(m, trimTrailingZeroes)

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

		result, err = object.ToBaseString(original, uppercase, requireSign, true, trimFractionalZeroes, min, 0, 16, padWith)

	case format.FORMAT_BASE:
		things := pr.popMultiple(6)
		padWithZeroes := things[5].(*object.Boolean).Value
		requireSign := things[4].(*object.Boolean).Value
		uppercase := things[3].(*object.Boolean).Value
		minimum := things[2]
		base := things[1]
		original := things[0]

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

		result, err = object.ToBaseString(original, uppercase, requireSign, true, trimFractionalZeroes, min, 0, b, padWith)

	case format.FORMAT_FIXED:
		things := pr.popMultiple(5)
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

		// TODO
		trimFractionalZeroes := false

		padIntWith := ' '
		if padIntWithZeroes {
			padIntWith = '0'
		}

		result, err = object.ToBaseString(original, false, requireSign, true, trimFractionalZeroes, intMin, fracRound, 10, padIntWith)

	case format.FORMAT_SCIENTIFIC_NOTATION:
		things := pr.popMultiple(7)

		scaleExp := things[6]
		requireExpSign := things[5].(*object.Boolean).Value
		uppercase := things[4].(*object.Boolean).Value
		scaleTrimTrailingZeroes := things[3].(*object.Boolean).Value
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
			orig.ScientificNotation(uppercase, requireSign, requireExpSign, rescale, scaleTrimTrailingZeroes, sc, scExp))

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
