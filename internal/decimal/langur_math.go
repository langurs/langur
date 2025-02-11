// langur/decimal/langur_math.go

package decimal

import (
	"fmt"
	"strings"
)

// bʸ = x
// root: have x and y; find b (determine base)
// NOTE(davis): I tried calculation, following an example from rosettacode.org for the nth root, ...
// ... but the following proved to be faster and more useful.

// This takes a divide and conquer approach. There might be a name for that, but I don't know.
// The simplest case is the square root of 4. Divide 4 by 2 and you already have the result.
// With 34 digits of precision, it typically takes about 111 iterations to find the result.
func (x Decimal) Root(y Decimal) Decimal {
	var low, high, lastHigh, lastLow, useX, useY, expTest Decimal

	if y.IsZero() {
		// indeterminate
		decThrow("Cannot calculate 0 root")
		return Zero
	}
	if x.IsZero() {
		// What should the result be?
		// 0? NaN? error?
		decThrow("Cannot calculate root for 0")
		return Zero
	}

	if !y.IsInteger() {
		// return x.Pow(One.Div(y))
		
		decThrow("Cannot calculate non-integer root in the current implementation")
		return Zero
	}
	isNegX := x.LessThan(Zero)
	if isNegX && y.IsEvenInteger() {
		decThrow("Cannot calculate an even root on a negative number")
		return Zero
	}

	isNegY := y.LessThan(Zero)
	if isNegY {
		useY = y.Neg()
	} else {
		useY = y
	}
	isXunderOne := x.Abs().LessThan(One)
	if isXunderOne {
		useX = One.DivWithMinMaxScale(x)
	} else {
		useX = x
	}

	if isNegX {
		high = Zero
		low = useX
	} else {
		high = useX
		low = Zero
	}
	b := useX.DivWithMinMaxScale(Two)

	maxIterations := DivisionPrecision * 100 // ?

	for i := 1; ; i++ {
		if i > 100 {
			lastHigh = high
			lastLow = low
		}

		// expTest, err = exponent(b, useY)
		// if err != nil {
		// 	return Zero, err
		// }
		expTest = b.Pow(useY)

		if expTest.Equal(useX) {
			// success!
			break
		}
		if expTest.GreaterThan(useX) {
			// midpoint too high; is new high point
			high = b
		} else {
			// midpoint too low; is new low point
			low = b
		}
		if i > 100 && high.Equal(lastHigh) && low.Equal(lastLow) {
			// starts checking after 100 iterations; to not waste time checking too early
			// no more changes; stop here
			break
		}

		b = midFromTwo(low, high)

		if i > maxIterations {
			// This is to prevent a runaway.
			decThrow(fmt.Sprintf("Failed to calculate root; too many iterations (%d)", maxIterations))
			return Zero
		}
	}

	if isNegY && !isXunderOne ||
		isXunderOne && !isNegY {

		b = One.DivWithMinMaxScale(b)
	} else {
		b = b.RescaleMin(0, true)
	}

	return b
}

// TODO
// bʸ = x
// logarithm: have x and b; find y (determine exponent)
// func (d Decimal) Logarithm() {}

func midFromTwo(low, high Decimal) Decimal {
	return high.Sub(low).DivWithMinMaxScale(Two).Add(low)
}

func Mid(nums ...Decimal) Decimal {
	high := nums[0]
	low := nums[0]

	for i := 1; i < len(nums); i++ {
		// if nums[i].IsInfinite() || nums[i].IsNaN() {
		// 	return nums[i], fmt.Errorf("Cannot calculate midpoint with NaN or Infinity")
		// }
		if nums[i].GreaterThan(high) {
			high = nums[i]
		}
		if nums[i].LessThan(low) {
			low = nums[i]
		}
	}

	return midFromTwo(low, high)
}

func Mean(nums ...Decimal) Decimal {
	total := Zero
	for _, n := range nums {
		// if n.IsInfinite() || n.IsNaN() {
		// 	return n, fmt.Errorf("Cannot calculate mean with NaN or Infinity")
		// }
		total = total.Add(n)
	}
	return total.DivWithMinMaxScale(NewFromInt(int64(len(nums))))
}

func (d Decimal) ToFraction() (numerator Decimal, denominator Decimal) {
	parts := d.StringParts()
	if parts[1] == "" {
		// done; no fractional; whole number over 1
		return d, One
	}
	numerator = RequireFromString(parts[0]+parts[1])
	denominator = RequireFromString("1"+strings.Repeat("0", len(parts[1])))
	
	// simplify fraction with greatest common divisor
	gcd := Gcd(numerator, denominator).Abs()
	if !gcd.Equal(One) {
		numerator = numerator.DivTruncate(gcd, 0)
		denominator = denominator.DivTruncate(gcd, 0)
	}
	
	return
}

// greatest common divisor
// use the Euclidian method for fast calculation with large numbers
func Gcd(d1, d2 Decimal) Decimal {
	for !d1.IsZero() {
		d1, d2 = d2.Mod(d1), d1
	}
	return d2
}

// least common muliple
func Lcm(d1, d2 Decimal) Decimal {
	return d1.Mul(d2).DivTruncate(Gcd(d1, d2), 0)
}

// Why Pow2()?: The decimal library Pow() function is short-changing fractional exponents.
// Since we have a Root() function, we can create a fraction and do this in 2 steps to get a better result.
// 2 ^ 3.5 == (2 ^ 35) ^/ 10
func (d Decimal) Pow2(exp Decimal) Decimal {
	if exp.IsInteger() {
		return d.Pow(exp)
	}
	expNumerator, expDenominator := exp.ToFraction()
	return d.Pow(expNumerator).Root(expDenominator)

	// if exp.IsInteger() {
	// 	// integer denominator
	// 	power, ok := exp.ToInt(true)
	// 	if !ok {
	// 		decThrow("Cannot calculate integer power outside range of int")
	// 	}
	// 	inverse := power < 0
	// 	if inverse {
	// 		power = -power
	// 	}

	// 	result := One
	// 	for i := 0; i < power; i++ {
	// 		result = result.Mul(d)
	// 	}

	// 	if inverse {			
	// 		result = One.Div(result)
	// 	}
	// 	return result

	// } else {
	// 	// convert to fraction and call self...
	// 	expNumerator, expDenominator := exp.ToFraction()
	// 	return d.Pow2(expNumerator).Root(expDenominator)
	// }
}
