// langur/object/range.go

package object

import (
	"fmt"
	"langur/common"
)

type Range struct {
	Start Object
	End   Object
}

func (r *Range) Copy() Object {
	return &Range{Start: r.Start.Copy(), End: r.End.Copy()}
}

func (l *Range) Equal(r2 Object) bool {
	r, ok := r2.(*Range)
	if !ok {
		return false
	}
	return Equal(l.Start, r.Start) && Equal(l.End, r.End)
}

func (r *Range) String() string {
	return ComposedOrRegularString(r.Start) + " .. " + ComposedOrRegularString(r.End)
}

func (r *Range) ReplString() string {
	return common.RangeTypeName + " " + r.Start.ReplString() + " .. " + r.End.ReplString()
}

func (r *Range) Type() ObjectType {
	return RANGE_OBJ
}
func (r *Range) TypeString() string {
	return common.RangeTypeName
}

func (r *Range) IsTruthy() bool {
	return r.IsFlatOrAscending()
}

func (r *Range) IsFlatOrAscending() bool {
	g, _ := GreaterOrEqual(r.End, r.Start)
	return g
}

func NewRange(start, end Object) *Range {
	if RangeValid(start, end) {
		return &Range{Start: start, End: end}
	}
	return nil
}

func RangeValid(start, end Object) bool {
	_, ok := start.(IGreaterThan)
	return ok && start.Type() == end.Type()
}

func WithinRange(value Object, rng *Range) (bool, error) {
	return WithinValueRange(value, rng.Start, rng.End)
}

func WithinValueRange(value, start, end Object) (bool, error) {
	// only determines if within the range of 2 values (not considering an increment)

	low, high := start, end
	gt, comparable := GreaterThan(start, end)
	if !comparable {
		return false, fmt.Errorf("Could not compare start/end of range")
	}
	if gt {
		low, high = high, low
	}

	gt, comparable = GreaterThan(low, value)
	if !comparable {
		return false, fmt.Errorf("Could not compare lowest value of range with value")
	}
	if gt {
		return false, nil
	}

	gt, comparable = GreaterThan(value, high)
	if !comparable {
		return false, fmt.Errorf("Could not compare highest value of range with value")
	}
	if gt {
		return false, nil
	}

	return true, nil
}

func (r *Range) ToList(inc *Number, correctIncrementSign bool) (*List, error) {
	var ints []int64
	i, err := inc.ToInt64()
	if err == nil {
		ints, err = r.toInt64Slice(i, correctIncrementSign)
	}
	if err != nil {
		// try using the decimal library for bigger numbers
		list, err := r.toSlice(inc, correctIncrementSign)
		if err != nil {
			return nil, err
		}
		return &List{Elements: list}, nil
	}
	return ListFromInt64Slice(ints), nil
}

func (r *Range) toInt64Slice(inc int64, correctIncrementSign bool) ([]int64, error) {
	var start, end int64
	var err error

	switch e := r.Start.(type) {
	case *Number:
		start, err = e.ToInt64()
		if err != nil {
			return nil, fmt.Errorf("Expected int64 range")
		}
	}
	switch e := r.End.(type) {
	case *Number:
		end, err = e.ToInt64()
		if err != nil {
			return nil, fmt.Errorf("Expected int64 range")
		}
	}

	return int64PairToSlice(start, end, inc, correctIncrementSign)
}

// convert to slice using Object types from start
func (r *Range) toSlice(inc *Number, correctIncrementSign bool) ([]Object, error) {
	var start, end *Number
 
	switch e := r.Start.(type) {
	case *Number:
		start = e
	default:
		return nil, fmt.Errorf("Expected numeric range")
	}
	switch e := r.End.(type) {
	case *Number:
		end = e
	default:
		return nil, fmt.Errorf("Expected numeric range")
	}

	return numberPairToSlice(start, end, inc, correctIncrementSign)
}

func (right *Range) ToNumber() (*Number, bool) {
	// gives hypothetical count for an integer range converted to list elements
	start, ok := right.Start.(*Number)
	end, ok2 := right.End.(*Number)
	if ok && ok2 {
		n := end.Subtract(start)
		if n == nil {
			return nil, false
		}
		n = n.(*Number).Abs()
		n2 := n.(*Number).Add(One)
		if n2 == nil {
			return nil, false
		}
		return n2.(*Number), true
	}
	return nil, false
}

func (r *Range) ToString() (*String, error) {
	rSlc, err := r.toRuneSlice()
	if err == nil {
		return NewStringFromParts(rSlc)
	}
	return nil, err
}

func (r *Range) toRuneSlice() ([]rune, error) {
	var start, end, num rune
	var err error

	switch e := r.Start.(type) {
	case *Number:
		start, err = e.ToRune()
		if err != nil {
			return nil, fmt.Errorf("Invalid code point for start")
		}
	}
	switch e := r.End.(type) {
	case *Number:
		end, err = e.ToRune()
		if err != nil {
			return nil, fmt.Errorf("Invalid code point for end")
		}
	}

	num = start
	if start > end {
		// descending range
		numbers := make([]rune, 0, start-end+1)
		for {
			numbers = append(numbers, num)
			num--
			if num < end {
				break
			}
		}

		return numbers, nil

	} else {
		numbers := make([]rune, 0, end-start+1)
		for {
			numbers = append(numbers, num)
			num++
			if num > end {
				break
			}
		}

		return numbers, nil
	}
}
