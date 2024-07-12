// langur/decimal/langur.go

// methods added to shopspring/decimal for langur
// see also langur/decimal/langur_zeroes.go

package decimal

import (
	"langur/modes"
	"strconv"
	"strings"
	"unicode"
)

// create once to make computations faster
var One = NewFromInt(1)
var Two = NewFromInt(2)

func decThrow(s string) {
	panic("decimal exception: " + s)
}

func (d Decimal) ToInt64(trimTrailingZeros bool) (int64, bool) {
	if !d.IsInteger() {
		return 0, false
	}
	s := d.string(trimTrailingZeros)
	i, err := strconv.ParseInt(s, 10, 64)
	if err == nil {
		return i, true
	}
	return 0, false
}

func (d Decimal) ToInt32(trimTrailingZeros bool) (int32, bool) {
	if !d.IsInteger() {
		return 0, false
	}
	s := d.string(trimTrailingZeros)
	i, err := strconv.ParseInt(s, 10, 32)
	if err == nil {
		return int32(i), true
	}
	return 0, false
}

func (d Decimal) ToInt(trimTrailingZeros bool) (int, bool) {
	if !d.IsInteger() {
		return 0, false
	}
	s := d.string(trimTrailingZeros)
	i, err := strconv.ParseInt(s, 10, 0)
	if err == nil {
		return int(i), true
	}
	return 0, false
}

func (d Decimal) ToRune(trimtrailingzeros bool) (rune, bool) {
	if !d.IsInteger() {
		return 0, false
	}
	s := d.string(trimtrailingzeros)
	i, err := strconv.ParseInt(s, 10, 64)
	if err == nil {
		if i >= 0 && i <= unicode.MaxRune {
			return rune(i), true
		}
	}
	return 0, false
}

func (d Decimal) TrueMod(d2 Decimal) Decimal {
	// Modulus and remainder are not the same operation.
	// https://www.microsoft.com/en-us/research/publication/division-and-modulus-for-computer-scientists/
	// https://stackoverflow.com/questions/13683563/whats-the-difference-between-mod-and-remainder
	quo := d.Mod(d2)
	if quo.LessThan(Zero) {
		quo = quo.Add(d2.Abs())
	}
	return quo
}

// The rescale Boolean determines if the part left of e should be rounded.
func (d Decimal) ScientificNotation(
	capitalize, requireSign, requireExpSign,
	rescale, scaleAddTrailingZeroes, scaleTrimTrailingZeroes bool,
	scale, scaleExp int,
	decimalPoint rune) string {

	parts := d.StringParts()
	p1, p2 := parts[0], parts[1]

	sign := "+"
	if p1[0] == '-' {
		sign = "-"
		p1 = p1[1:]
	}

	exp := 0
	if len(p1) > 1 {
		// large number
		exp = len(p1) - 1
		p2 = p1[1:] + p2
		p1 = p1[:1]

	} else if p1 == "0" && p2 != "" {
		// small number
		// find first non-zero in p2
		for i, c := range p2 {
			if c != '0' {
				p1 = string(c)
				if len(p2) == i+1 {
					p2 = "0"
				} else {
					p2 = p2[i+1:]
				}
				exp = -i - 1
				break
			}
		}
	}

	p1p2 := p1
	if p2 != "" {
		// use dot here; replace later if necessary
		p1p2 += "." + p2
	}

	if rescale && len(p2) != scale {
		p1p2 = RequireFromString(p1p2).
			RoundByMode(int32(scale),
				scaleAddTrailingZeroes, scaleTrimTrailingZeroes,
				modes.RoundingMode).
			string(false)

	} else {
		if scaleAddTrailingZeroes && len(p2) < scale {
			if p1p2 == "0" {
				p1p2 += string(decimalPoint)
			}
			p1p2 += strings.Repeat("0", scale-len(p2))
		}

		if scaleTrimTrailingZeroes {
			if p1p2 != "0" {
				p1p2 = strings.TrimRight(p1p2, "0")
			}
			if p1p2[len(p1p2)-1] == '.' {
				// removes hanging dot
				p1p2 = p1p2[:len(p1p2)-1]
			}
		}
	}

	e := "e"
	if capitalize {
		e = "E"
	}

	if !requireSign && sign == "+" {
		sign = ""
	}

	expSign := ""
	expStr := strconv.FormatInt(int64(exp), 10)
	expLen := len(expStr)
	if exp < 0 {
		expSign = "-"
		expStr = expStr[1:]
		expLen--
	} else if requireExpSign {
		expSign = "+"
	}
	if expLen < scaleExp {
		expStr = strings.Repeat("0", scaleExp-expLen) + expStr
	}

	if decimalPoint != '.' {
		p1p2 = strings.Replace(p1p2, ".", string(decimalPoint), 1)
	}

	return sign + p1p2 + e + expSign + expStr
}
