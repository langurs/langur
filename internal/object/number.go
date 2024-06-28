// langur/object/number.go

package object

import (
	"fmt"
	"langur/common"
	"langur/cpoint"
	dec "langur/decimal"
	"langur/native"
	"langur/str"
	"math"
	"strings"
)

type decType = dec.Decimal

const maxDivisionMaxScaleMode int = math.MaxInt32 - 2

// TODO: make safe way to change across goroutines or langur processes
func SetDivisionMaxScaleMode(p int) error {
	// This is the number of digits after the decimal point.
	// I don't suppose this is the same as "precision" (total number of digits).
	// set max scale for division (has to stop somewhere)

	if p < 0 || p > maxDivisionMaxScaleMode {
		return fmt.Errorf("Cannot Set Division Max Scale to %d (out of range)", p)
	}
	dec.DivisionPrecision = p
	return nil
}
func GetDivisionMaxScaleMode() int {
	return dec.DivisionPrecision
}

// build once
var Zero = NumberFromInt(0)
var One = NumberFromInt(1)
var Two = NumberFromInt(2)

type Number struct {
	decimal decType
	integer int64

	// is using the integer optimization ...
	// 1. is within int64 limits
	// 2. has no trailing zeros after the decimal point
	usingIntOptimization bool
}

func (n *Number) HashKey() Object {
	// simplifies number key so 1.0 hashes the same 1
	return n.Simplify()
}

func (n *Number) Type() ObjectType {
	return NUMBER_OBJ
}
func (n *Number) TypeString() string {
	return common.NumberTypeName
}

func (n *Number) Copy() Object {
	return &Number{
		decimal:              n.decimal, // okay b/c decimal library generates a new decimal if it does calculation
		integer:              n.integer,
		usingIntOptimization: n.usingIntOptimization,
	}
}

func (l *Number) Equal(n2 Object) bool {
	r, ok := n2.(*Number)
	if !ok {
		return false
	}
	if l.usingIntOptimization && r.usingIntOptimization {
		return l.integer == r.integer
	}
	return l.ToDecimal().Equal(r.ToDecimal())
}

func (l *Number) Same(n2 Object) bool {
	// 1 equals 1.0, but they are not the same.
	r, ok := n2.(*Number)
	if !ok {
		return false
	}
	if l.usingIntOptimization && r.usingIntOptimization {
		return l.integer == r.integer
	}
	return l.ToDecimal().Same(r.ToDecimal())
}

func (l *Number) GreaterThan(n2 Object) (bool, bool) {
	r, ok := n2.(*Number)
	if !ok {
		return false, false
	}
	if l.usingIntOptimization && r.usingIntOptimization {
		return l.integer > r.integer, true
	}
	return l.ToDecimal().GreaterThan(r.ToDecimal()), true
}

func (n *Number) IsTruthy() bool {
	if n.usingIntOptimization {
		return n.integer != 0
	}
	return !n.decimal.IsZero()
}

func (n *Number) ToDecimal() decType {
	if n.usingIntOptimization {
		return dec.NewFromInt(n.integer)
	}
	return n.decimal
}

func (n *Number) UseDecimal() *Number {
	if n.usingIntOptimization {
		return &Number{decimal: dec.NewFromInt(n.integer)}
	}
	return n
}

func (n *Number) Optimize() *Number {
	if n.usingIntOptimization {
		return n
	}
	d, ok := n.decimal.ToInt64(false)
	if ok {
		return NumberFromInt64(d)
	}
	return n
}

func NumberToInt(n Object) (int, bool) {
	switch n := n.(type) {
	case *Number:
		new, err := n.ToInt()
		if err == nil {
			return new, true
		}
	}
	return 0, false
}

func NumberToRune(n Object) (rune, bool) {
	switch n := n.(type) {
	case *Number:
		new, err := n.ToRune()
		if err == nil {
			return new, true
		}
	}
	return 0, false
}

func (l *Number) ToInt() (int, error) {
	if l.usingIntOptimization {
		num, ok := native.ToInt(l.integer)
		if ok {
			return num, nil
		}

	} else if l.decimal.IsInteger() {
		d, ok := l.decimal.ToInt(true)
		if ok {
			return d, nil
		}
	}

	return 0, fmt.Errorf("Unable to convert %s to int (outside of range or not an integer)", l.String())
}

func (l *Number) ToInt64() (int64, error) {
	if l.usingIntOptimization {
		return l.integer, nil

	} else if l.decimal.IsInteger() {
		i, ok := l.decimal.ToInt64(true)
		if ok {
			return i, nil
		}
	}

	return 0, fmt.Errorf("Unable to convert %s to int64 (outside of range or not an integer)", l.String())
}

func (l *Number) ToInt32() (int32, error) {
	if l.usingIntOptimization {
		i, ok := native.ToInt32(l.integer)
		if ok {
			return i, nil
		}

	} else if l.decimal.IsInteger() {
		i, ok := l.decimal.ToInt32(true)
		if ok {
			return i, nil
		}
	}

	return 0, fmt.Errorf("Unable to convert %s to int32 (outside of range or not an integer)", l.String())
}

func (l *Number) ToRune() (rune, error) {
	if l.usingIntOptimization {
		i, ok := native.ToRune(l.integer)
		if ok {
			return i, nil
		}

	} else if l.decimal.IsInteger() {
		i, ok := l.decimal.ToRune(true)
		if ok {
			return i, nil
		}
	}

	return 0, fmt.Errorf("Unable to convert %s to rune (outside of range or not an integer)", l.String())
}

func (n *Number) String() string {
	if n.usingIntOptimization {
		return str.Int64ToStr(n.integer, 10)
	}
	return n.decimal.StringWithTrailingZeros()
}

func (n *Number) ReplString() string {
	return common.NumberTypeName + " " + n.String()
}

func NumberFromRune(r rune) *Number {
	return &Number{integer: int64(r), usingIntOptimization: true}
}
func NumberFromInt(n int) *Number {
	return &Number{integer: int64(n), usingIntOptimization: true}
}
func NumberFromInt64(n int64) *Number {
	return &Number{integer: n, usingIntOptimization: true}
}
func numberFromDecimal(d decType) *Number {
	n := &Number{decimal: d}
	return n.Optimize()
}

func NumberFromString(s string) (*Number, error) {
	var err error
	var i int64
	i, err = str.StrToInt64(s, 10)
	if err == nil {
		return NumberFromInt64(i), nil
	}
	var d decType
	d, err = dec.NewFromString(s)
	return numberFromDecimal(d), err
}

func NumberFromStringBase(s string, base int) (*Number, error) {
	if base == 10 || base == 0 {
		// base 10 may be bigger number than can be converted from other bases
		return NumberFromString(s)
	}
	if base < 2 || base > 36 {
		return Zero, fmt.Errorf("Base must be between 2 and 36")
	}
	i64, err := str.StrToInt64(s, base)
	if err != nil {
		return Zero, fmt.Errorf("Out of range of int64 or non-base 10 used for fractional numbers or for e-notation")
	}
	return NumberFromInt64(i64), nil
}

func (n *Number) ScientificNotation(
	capitalize, requireSign, requireExpSign, rescale bool, scale, scaleExp int) string {

	// convert to decimal to use already developed method
	return n.UseDecimal().decimal.ScientificNotation(
		capitalize, requireSign, requireExpSign, rescale, scale, scaleExp)
}

func ToNumber(obj Object, base int) (*Number, bool) {
	var s string
	switch obj := obj.(type) {
	case *Number:
		return obj, true

	case *String:
		s = obj.String()

	case *Boolean:
		if obj.Value {
			return One, true
		} else {
			return Zero, true
		}

	case *DateTime:
		// Unix nanoseconds since January 1, 1970 UTC
		return NumberFromInt64(obj.UnixNano()), true

	case *Duration:
		return NumberFromInt64(obj.ToNanoseconds()), true

	case *Range:
		return obj.ToNumber()

	default:
		s = obj.String()
	}

	n, err := NumberFromStringBase(s, base)
	if err == nil {
		return n, true
	}
	return n, false
}

func (n *Number) ToByte() (byte, error) {
	num, err := n.ToInt()
	if err != nil {
		return 0, fmt.Errorf("Expected integer for unsigned byte")
	}
	if num < 0 {
		return 0, fmt.Errorf("Cannot use negative number for unsigned byte")
	}
	if num > 255 {
		return 0, fmt.Errorf("Cannot use number above 255 for unsigned byte")
	}
	return byte(num), nil
}

func CodePointsToFlatRuneSlice(cp Object) ([]rune, error) {
	switch arg := cp.(type) {
	case *Number:
		r, err := arg.ToRune()
		if err != nil {
			return nil, fmt.Errorf("Invalid code point")
		}
		return []rune{r}, nil

	case *List:
		rSlc := make([]rune, 0, len(arg.Elements))
		for _, v := range arg.Elements {
			s, err := CodePointsToFlatRuneSlice(v)
			if err != nil {
				return nil, err
			}
			rSlc = append(rSlc, []rune(s)...)
		}
		return rSlc, nil

	case *Range:
		from, err := arg.ToList()
		if err != nil {
			return nil, err
		}
		rSlc := make([]rune, 0, len(from.Elements))
		for _, v := range from.Elements {
			s, err := CodePointsToFlatRuneSlice(v)
			if err != nil {
				return nil, err
			}
			rSlc = append(rSlc, []rune(s)...)
		}
		return rSlc, nil
	}

	return nil, fmt.Errorf("Expected integer, range, or list")
}

func ToBaseString(
	original Object,
	uppercase, requireSign, signCountsForPadding bool,
	integerMin, fracRound, base int,
	padWith rune) (
	*String, error) {

	var parts []string
	var intErr error
	var integer int64

	switch numObj := original.(type) {
	case *Number:
		// NOTE: base 10 rounding; would not be sufficient if fractionals possible on other bases
		n, err := numObj.Round(fracRound)
		if err != nil {
			return nil, err
		}
		parts = strings.Split(n.String(), ".")
		integer, intErr = n.ToInt64()

	default:
		return nil, fmt.Errorf("Expected number")
	}

	var s, intPadding string
	var isNeg bool
	if base == 10 && len(parts) == 2 {
		if parts[0][0] == '-' {
			parts[0] = parts[0][1:]
			isNeg = true
		}
		if len(parts[0]) < integerMin {
			intPadding = strings.Repeat(string(padWith), integerMin-len(parts[0]))
		}
		s = parts[0] + "." + parts[1]

	} else {
		if intErr != nil {
			return nil, intErr
		}

		s = str.Int64ToStr(integer, base)
		if s[0] == '-' {
			s = s[1:]
			isNeg = true
		}
		if len(s) < integerMin {
			intPadding = strings.Repeat(string(padWith), integerMin-len(s))
		}

		if uppercase {
			s = strings.ToUpper(s)
		}
	}

	sign := ""
	if isNeg {
		sign = "-"
	} else if requireSign {
		sign = "+"
	}

	if signCountsForPadding && sign != "" && len(intPadding) > 0 {
		intPadding = intPadding[1:]
	}

	if padWith == ' ' {
		s = intPadding + sign + s
	} else {
		s = sign + intPadding + s
	}

	return NewString(s), nil
}

func (n *Number) Reverse() *Number {
	rSlc := []rune(n.String())
	isNeg := false

	if rSlc[0] == '-' {
		isNeg = true
		rSlc = cpoint.ReverseSlice(rSlc[1:])
	} else {
		rSlc = cpoint.ReverseSlice(rSlc)
	}

	newStr := string(rSlc)
	if isNeg {
		newStr = "-" + newStr
	}
	num, _ := NumberFromString(newStr)
	return num
}
