// This code is forked from Regen Ledger.
// Their code is under the Apache2.0 License
// https://github.com/regen-network/regen-ledger/blob/3d818cf6e01af92eed25de5c17728a79070f56a3/types/math/dec_test.go

package math

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

func TestDec(t *testing.T) {

	// Property tests
	t.Run("TestNewDecFromInt64", rapid.MakeCheck(testDecInt64))

	// Properties about *FromString functions
	t.Run("TestInvalidNewDecFromString", rapid.MakeCheck(testInvalidNewDecFromString))
	t.Run(
		"TestInvalidNewNonNegativeDecFromString",
		rapid.MakeCheck(testInvalidNewNonNegativeDecFromString),
	)
	t.Run(
		"TestInvalidNewNonNegativeFixedDecFromString",
		rapid.MakeCheck(testInvalidNewNonNegativeFixedDecFromString),
	)
	t.Run(
		"TestInvalidNewPositiveDecFromString",
		rapid.MakeCheck(testInvalidNewPositiveDecFromString),
	)
	t.Run(
		"TestInvalidNewPositiveFixedDecFromString",
		rapid.MakeCheck(testInvalidNewPositiveFixedDecFromString),
	)

	// Properties about addition
	t.Run("TestAddLeftIdentity", rapid.MakeCheck(testAddLeftIdentity))
	t.Run("TestAddRightIdentity", rapid.MakeCheck(testAddRightIdentity))
	t.Run("TestAddCommutative", rapid.MakeCheck(testAddCommutative))
	t.Run("TestAddAssociative", rapid.MakeCheck(testAddAssociative))

	// Properties about subtraction
	t.Run("TestSubRightIdentity", rapid.MakeCheck(testSubRightIdentity))
	t.Run("TestSubZero", rapid.MakeCheck(testSubZero))

	// Properties about multiplication
	t.Run("TestMulLeftIdentity", rapid.MakeCheck(testMulLeftIdentity))
	t.Run("TestMulRightIdentity", rapid.MakeCheck(testMulRightIdentity))
	t.Run("TestMulCommutative", rapid.MakeCheck(testMulCommutative))
	t.Run("TestMulAssociative", rapid.MakeCheck(testMulAssociative))
	t.Run("TestZeroIdentity", rapid.MakeCheck(testMulZero))

	// Properties about division
	t.Run("TestDivisionBySelf", rapid.MakeCheck(testSelfQuo))
	t.Run("TestDivisionByOne", rapid.MakeCheck(testQuoByOne))

	// Properties combining operations
	t.Run("TestSubAdd", rapid.MakeCheck(testSubAdd))
	t.Run("TestAddSub", rapid.MakeCheck(testAddSub))
	t.Run("TestMulQuoA", rapid.MakeCheck(testMulQuoA))
	t.Run("TestMulQuoB", rapid.MakeCheck(testMulQuoB))
	t.Run("TestMulQuoExact", rapid.MakeCheck(testMulQuoExact))
	t.Run("TestQuoMulExact", rapid.MakeCheck(testQuoMulExact))

	// Properties about comparison and equality
	t.Run("TestCmpInverse", rapid.MakeCheck(testCmpInverse))
	t.Run("TestEqualCommutative", rapid.MakeCheck(testEqualCommutative))

	// Properties about tests on a single Dec
	t.Run("TestIsZero", rapid.MakeCheck(testIsZero))
	t.Run("TestIsNegative", rapid.MakeCheck(testIsNegative))
	t.Run("TestIsPositive", rapid.MakeCheck(testIsPositive))
	t.Run("TestNumDecimalPlaces", rapid.MakeCheck(testNumDecimalPlaces))

	// Unit tests
	zero := Dec{}
	one := NewDecFromInt64(1)
	two := NewDecFromInt64(2)
	three := NewDecFromInt64(3)
	four := NewDecFromInt64(4)
	five := NewDecFromInt64(5)
	minusOne := NewDecFromInt64(-1)

	onePointOneFive, err := NewDecFromString("1.15")
	require.NoError(t, err)
	twoPointThreeFour, err := NewDecFromString("2.34")
	require.NoError(t, err)
	threePointFourNine, err := NewDecFromString("3.49")
	require.NoError(t, err)
	onePointFourNine, err := NewDecFromString("1.49")
	require.NoError(t, err)
	minusFivePointZero, err := NewDecFromString("-5.0")
	require.NoError(t, err)

	twoThousand := NewDecFinite(2, 3)
	require.True(t, twoThousand.Equal(NewDecFromInt64(2000)))

	res, err := two.Add(zero)
	require.NoError(t, err)
	require.True(t, res.Equal(two))

	res, err = five.Sub(two)
	require.NoError(t, err)
	require.True(t, res.Equal(three))

	/*
		res, err = SafeSubBalance(five, two)
		require.NoError(t, err)
		require.True(t, res.Equal(three))

		_, err = SafeSubBalance(two, five)
		require.Error(t, err, "Expected insufficient funds error")

		res, err = SafeAddBalance(three, two)
		require.NoError(t, err)
		require.True(t, res.Equal(five))

		_, err = SafeAddBalance(minusFivePointZero, five)
		require.Error(t, err, "Expected ErrInvalidRequest")
	*/

	res, err = four.Quo(two)
	require.NoError(t, err)
	require.True(t, res.Equal(two))

	res, err = five.QuoInteger(two)
	require.NoError(t, err)
	require.True(t, res.Equal(two))

	res, err = five.Rem(two)
	require.NoError(t, err)
	require.True(t, res.Equal(one))

	x, err := four.Int64()
	require.NoError(t, err)
	require.Equal(t, int64(4), x)

	require.Equal(t, "5", five.String())

	res, err = onePointOneFive.Add(twoPointThreeFour)
	require.NoError(t, err)
	require.True(t, res.Equal(threePointFourNine))

	res, err = threePointFourNine.Sub(two)
	require.NoError(t, err)
	require.True(t, res.Equal(onePointFourNine))

	res, err = minusOne.Sub(four)
	require.NoError(t, err)
	require.True(t, res.Equal(minusFivePointZero))

	require.True(t, zero.IsZero())
	require.False(t, zero.IsPositive())
	require.False(t, zero.IsNegative())

	require.False(t, one.IsZero())
	require.True(t, one.IsPositive())
	require.False(t, one.IsNegative())

	require.False(t, minusOne.IsZero())
	require.False(t, minusOne.IsPositive())
	require.True(t, minusOne.IsNegative())

	res, err = one.MulExact(two)
	require.NoError(t, err)
	require.True(t, res.Equal(two))

	ten := NewDecFromInt64(10)
	oneHundred := NewDecFromInt64(100)
	logTenOneHundred, err := Log10(oneHundred)
	require.NoError(t, err)
	require.True(t, two.Equal(logTenOneHundred))
	logTenTen, err := Log10(ten)
	require.NoError(t, err)
	require.True(t, one.Equal(logTenTen))
	logTenOne, err := Log10(one)
	require.NoError(t, err)
	require.True(t, zero.Equal(logTenOne))

	logEOne, err := Ln(one)
	require.NoError(t, err)
	require.True(t, zero.Equal(logEOne))

	eight := NewDecFromInt64(8)
	twoCubed, err := Pow(two, three)
	require.NoError(t, err)
	require.True(t, eight.Equal(twoCubed))

	oneThousand := NewDecFromInt64(1000)
	tenSquared, err := Exp10(two)
	require.NoError(t, err)
	require.True(t, oneHundred.Equal(tenSquared))
	tenCubed, err := Exp10(three)
	require.NoError(t, err)
	require.True(t, oneThousand.Equal(tenCubed))

	cielOnePointFourNine, err := onePointFourNine.Ceil()
	require.NoError(t, err)
	require.True(t, two.Equal(cielOnePointFourNine))

	onePointFiveOne, err := NewDecFromString("1.51")
	require.NoError(t, err)
	floorOnePointFiveOne, err := onePointFiveOne.Floor()
	require.NoError(t, err)
	require.True(t, one.Equal(floorOnePointFiveOne))
}

// TODO: Think a bit more about the probability distribution of Dec
var genDec *rapid.Generator[Dec] = rapid.Custom(func(t *rapid.T) Dec {
	f := rapid.Float64().Draw(t, "f")
	dec, err := NewDecFromString(fmt.Sprintf("%g", f))
	require.NoError(t, err)
	return dec
})

// A Dec value and the float used to create it
type floatAndDec struct {
	float float64
	dec   Dec
}

// Generate a Dec value along with the float used to create it
var genFloatAndDec *rapid.Generator[floatAndDec] = rapid.Custom(func(t *rapid.T) floatAndDec {
	f := rapid.Float64().Draw(t, "f")
	dec, err := NewDecFromString(fmt.Sprintf("%g", f))
	require.NoError(t, err)
	return floatAndDec{f, dec}
})

// Property: n == NewDecFromInt64(n).Int64()
func testDecInt64(t *rapid.T) {
	nIn := rapid.Int64().Draw(t, "n")
	nOut, err := NewDecFromInt64(nIn).Int64()

	require.NoError(t, err)
	require.Equal(t, nIn, nOut)
}

// Property: invalid_number_string(s) => NewDecFromString(s) == err
func testInvalidNewDecFromString(t *rapid.T) {
	s := rapid.StringMatching("[[:alpha:]]+").Draw(t, "s")
	_, err := NewDecFromString(s)
	require.Error(t, err)
}

// Property: invalid_number_string(s) || IsNegative(s)
// => NewNonNegativeDecFromString(s) == err
func testInvalidNewNonNegativeDecFromString(t *rapid.T) {
	s := rapid.OneOf(
		rapid.StringMatching("[[:alpha:]]+"),
		rapid.StringMatching(`^-\d*\.?\d+$`).Filter(
			func(s string) bool { return !strings.HasPrefix(s, "-0") && !strings.HasPrefix(s, "-.0") },
		),
	).Draw(t, "s")
	_, err := NewNonNegativeDecFromString(s)
	require.Error(t, err)
}

// Property: invalid_number_string(s) || IsNegative(s) || NumDecimals(s) > n
// => NewNonNegativeFixedDecFromString(s, n) == err
func testInvalidNewNonNegativeFixedDecFromString(t *rapid.T) {
	n := rapid.Uint32Range(0, 999).Draw(t, "n")
	s := rapid.OneOf(
		rapid.StringMatching("[[:alpha:]]+"),
		rapid.StringMatching(`^-\d*\.?\d+$`).Filter(
			func(s string) bool { return !strings.HasPrefix(s, "-0") && !strings.HasPrefix(s, "-.0") },
		),
		rapid.StringMatching(fmt.Sprintf(`\d*\.\d{%d,}`, n+1)),
	).Draw(t, "s")
	_, err := NewNonNegativeFixedDecFromString(s, n)
	require.Error(t, err)
}

// Property: invalid_number_string(s) || IsNegative(s) || IsZero(s)
// => NewPositiveDecFromString(s) == err
func testInvalidNewPositiveDecFromString(t *rapid.T) {
	s := rapid.OneOf(
		rapid.StringMatching("[[:alpha:]]+"),
		rapid.StringMatching(`^-\d*\.?\d+|0$`),
	).Draw(t, "s")
	_, err := NewPositiveDecFromString(s)
	require.Error(t, err)
}

// Property: invalid_number_string(s) || IsNegative(s) || IsZero(s) || NumDecimals(s) > n
// => NewPositiveFixedDecFromString(s) == err
func testInvalidNewPositiveFixedDecFromString(t *rapid.T) {
	n := rapid.Uint32Range(0, 999).Draw(t, "n")
	s := rapid.OneOf(
		rapid.StringMatching("[[:alpha:]]+"),
		rapid.StringMatching(`^-\d*\.?\d+|0$`),
		rapid.StringMatching(fmt.Sprintf(`\d*\.\d{%d,}`, n+1)),
	).Draw(t, "s")
	_, err := NewPositiveFixedDecFromString(s, n)
	require.Error(t, err)
}

// Property: 0 + a == a
func testAddLeftIdentity(t *rapid.T) {
	a := genDec.Draw(t, "a")
	zero := NewDecFromInt64(0)

	b, err := zero.Add(a)
	require.NoError(t, err)

	require.True(t, a.Equal(b))
}

// Property: a + 0 == a
func testAddRightIdentity(t *rapid.T) {
	a := genDec.Draw(t, "a")
	zero := NewDecFromInt64(0)

	b, err := a.Add(zero)
	require.NoError(t, err)

	require.True(t, a.Equal(b))
}

// Property: a + b == b + a
func testAddCommutative(t *rapid.T) {
	a := genDec.Draw(t, "a")
	b := genDec.Draw(t, "b")

	c, err := a.Add(b)
	require.NoError(t, err)

	d, err := b.Add(a)
	require.NoError(t, err)

	require.True(t, c.Equal(d))
}

// Property: (a + b) + c == a + (b + c)
func testAddAssociative(t *rapid.T) {
	a := genDec.Draw(t, "a")
	b := genDec.Draw(t, "b")
	c := genDec.Draw(t, "c")

	// (a + b) + c
	d, err := a.Add(b)
	require.NoError(t, err)

	e, err := d.Add(c)
	require.NoError(t, err)

	// a + (b + c)
	f, err := b.Add(c)
	require.NoError(t, err)

	g, err := a.Add(f)
	require.NoError(t, err)

	require.True(t, e.Equal(g))
}

// Property: a - 0 == a
func testSubRightIdentity(t *rapid.T) {
	a := genDec.Draw(t, "a")
	zero := NewDecFromInt64(0)

	b, err := a.Sub(zero)
	require.NoError(t, err)

	require.True(t, a.Equal(b))
}

// Property: a - a == 0
func testSubZero(t *rapid.T) {
	a := genDec.Draw(t, "a")
	zero := NewDecFromInt64(0)

	b, err := a.Sub(a)
	require.NoError(t, err)

	require.True(t, b.Equal(zero))
}

// Property: 1 * a == a
func testMulLeftIdentity(t *rapid.T) {
	a := genDec.Draw(t, "a")
	one := NewDecFromInt64(1)

	b, err := one.Mul(a)
	require.NoError(t, err)

	require.True(t, a.Equal(b))
}

// Property: a * 1 == a
func testMulRightIdentity(t *rapid.T) {
	a := genDec.Draw(t, "a")
	one := NewDecFromInt64(1)

	b, err := a.Mul(one)
	require.NoError(t, err)

	require.True(t, a.Equal(b))
}

// Property: a * b == b * a
func testMulCommutative(t *rapid.T) {
	a := genDec.Draw(t, "a")
	b := genDec.Draw(t, "b")

	c, err := a.Mul(b)
	require.NoError(t, err)

	d, err := b.Mul(a)
	require.NoError(t, err)

	require.True(t, c.Equal(d))
}

// Property: (a * b) * c == a * (b * c)
func testMulAssociative(t *rapid.T) {
	a := genDec.Draw(t, "a")
	b := genDec.Draw(t, "b")
	c := genDec.Draw(t, "c")

	// (a * b) * c
	d, err := a.Mul(b)
	require.NoError(t, err)

	e, err := d.Mul(c)
	require.NoError(t, err)

	// a * (b * c)
	f, err := b.Mul(c)
	require.NoError(t, err)

	g, err := a.Mul(f)
	require.NoError(t, err)

	require.True(t, e.Equal(g))
}

// Property: (a - b) + b == a
func testSubAdd(t *rapid.T) {
	a := genDec.Draw(t, "a")
	b := genDec.Draw(t, "b")

	c, err := a.Sub(b)
	require.NoError(t, err)

	d, err := c.Add(b)
	require.NoError(t, err)

	require.True(t, a.Equal(d))
}

// Property: (a + b) - b == a
func testAddSub(t *rapid.T) {
	a := genDec.Draw(t, "a")
	b := genDec.Draw(t, "b")

	c, err := a.Add(b)
	require.NoError(t, err)

	d, err := c.Sub(b)
	require.NoError(t, err)

	require.True(t, a.Equal(d))
}

// Property: a * 0 = 0
func testMulZero(t *rapid.T) {
	a := genDec.Draw(t, "a")
	zero := Dec{}

	c, err := a.Mul(zero)
	require.NoError(t, err)
	require.True(t, c.IsZero())
}

// Property: a/a = 1
func testSelfQuo(t *rapid.T) {
	decNotZero := func(d Dec) bool { return !d.IsZero() }
	a := genDec.Filter(decNotZero).Draw(t, "a")
	one := NewDecFromInt64(1)

	b, err := a.Quo(a)
	require.NoError(t, err)
	require.True(t, one.Equal(b))
}

// Property: a/1 = a
func testQuoByOne(t *rapid.T) {
	a := genDec.Draw(t, "a")
	one := NewDecFromInt64(1)

	b, err := a.Quo(one)
	require.NoError(t, err)
	require.True(t, a.Equal(b))
}

// Property: (a * b) / a == b
func testMulQuoA(t *rapid.T) {
	decNotZero := func(d Dec) bool { return !d.IsZero() }
	a := genDec.Filter(decNotZero).Draw(t, "a")
	b := genDec.Draw(t, "b")

	c, err := a.Mul(b)
	require.NoError(t, err)

	d, err := c.Quo(a)
	require.NoError(t, err)

	require.True(t, b.Equal(d))
}

// Property: (a * b) / b == a
func testMulQuoB(t *rapid.T) {
	decNotZero := func(d Dec) bool { return !d.IsZero() }
	a := genDec.Draw(t, "a")
	b := genDec.Filter(decNotZero).Draw(t, "b")

	c, err := a.Mul(b)
	require.NoError(t, err)

	d, err := c.Quo(b)
	require.NoError(t, err)

	require.True(t, a.Equal(d))
}

// Property: (a * 10^b) / 10^b == a using MulExact and QuoExact
// and a with no more than b decimal places (b <= 32).
func testMulQuoExact(t *rapid.T) {
	b := rapid.Uint32Range(0, 32).Draw(t, "b")
	decPrec := func(d Dec) bool { return d.NumDecimalPlaces() <= b }
	a := genDec.Filter(decPrec).Draw(t, "a")

	c := NewDecFinite(1, int32(b))

	d, err := a.MulExact(c)
	require.NoError(t, err)

	e, err := d.QuoExact(c)
	require.NoError(t, err)

	require.True(t, a.Equal(e))
}

// Property: (a / b) * b == a using QuoExact and MulExact and
// a as an integer.
func testQuoMulExact(t *rapid.T) {
	a := rapid.Uint64().Draw(t, "a")
	aDec, err := NewDecFromString(fmt.Sprintf("%d", a))
	require.NoError(t, err)
	b := rapid.Uint32Range(0, 32).Draw(t, "b")
	c := NewDecFinite(1, int32(b))

	require.NoError(t, err)

	d, err := aDec.QuoExact(c)
	require.NoError(t, err)

	e, err := d.MulExact(c)
	require.NoError(t, err)

	require.True(t, aDec.Equal(e))
}

// Property: Cmp(a, b) == -Cmp(b, a)
func testCmpInverse(t *rapid.T) {
	a := genDec.Draw(t, "a")
	b := genDec.Draw(t, "b")

	require.Equal(t, a.Cmp(b), -b.Cmp(a))
}

// Property: Equal(a, b) == Equal(b, a)
func testEqualCommutative(t *rapid.T) {
	a := genDec.Draw(t, "a")
	b := genDec.Draw(t, "b")

	require.Equal(t, a.Equal(b), b.Equal(a))
}

// Property: isZero(f) == isZero(NewDecFromString(f.String()))
func testIsZero(t *rapid.T) {
	floatAndDec := genFloatAndDec.Draw(t, "floatAndDec")
	f, dec := floatAndDec.float, floatAndDec.dec

	require.Equal(t, f == 0, dec.IsZero())

}

// Property: isNegative(f) == isNegative(NewDecFromString(f.String()))
func testIsNegative(t *rapid.T) {
	floatAndDec := genFloatAndDec.Draw(t, "floatAndDec")
	f, dec := floatAndDec.float, floatAndDec.dec

	require.Equal(t, f < 0, dec.IsNegative())
}

// Property: isPositive(f) == isPositive(NewDecFromString(f.String()))
func testIsPositive(t *rapid.T) {
	floatAndDec := genFloatAndDec.Draw(t, "floatAndDec")
	f, dec := floatAndDec.float, floatAndDec.dec

	require.Equal(t, f > 0, dec.IsPositive())
}

// Property: floatDecimalPlaces(f) == NumDecimalPlaces(NewDecFromString(f.String()))
func testNumDecimalPlaces(t *rapid.T) {
	floatAndDec := genFloatAndDec.Draw(t, "floatAndDec")
	f, dec := floatAndDec.float, floatAndDec.dec

	require.Equal(t, floatDecimalPlaces(t, f), dec.NumDecimalPlaces())
}

func floatDecimalPlaces(t *rapid.T, f float64) uint32 {
	reScientific := regexp.MustCompile(
		`^\-?(?:[[:digit:]]+(?:\.([[:digit:]]+))?|\.([[:digit:]]+))(?:e?(?:\+?([[:digit:]]+)|(-[[:digit:]]+)))?$`,
	)
	fStr := fmt.Sprintf("%g", f)
	matches := reScientific.FindAllStringSubmatch(fStr, 1)
	if len(matches) != 1 {
		t.Fatalf("Didn't match float: %g", f)
	}

	// basePlaces is the number of decimal places in the decimal part of the
	// string
	basePlaces := 0
	if matches[0][1] != "" {
		basePlaces = len(matches[0][1])
	} else if matches[0][2] != "" {
		basePlaces = len(matches[0][2])
	}
	t.Logf("Base places: %d", basePlaces)

	// exp is the exponent
	exp := 0
	if matches[0][3] != "" {
		var err error
		exp, err = strconv.Atoi(matches[0][3])
		require.NoError(t, err)
	} else if matches[0][4] != "" {
		var err error
		exp, err = strconv.Atoi(matches[0][4])
		require.NoError(t, err)
	}

	// Subtract exponent from base and check if negative
	res := basePlaces - exp
	if res <= 0 {
		return 0
	}

	return uint32(res)
}

func TestIsFinite(t *testing.T) {
	a, err := NewDecFromString("1.5")
	require.NoError(t, err)

	require.True(t, a.IsFinite())

	b, err := NewDecFromString("NaN")
	require.NoError(t, err)

	require.False(t, b.IsFinite())
}

func TestReduce(t *testing.T) {
	a, err := NewDecFromString("1.30000")
	require.NoError(t, err)
	b, n := a.Reduce()
	require.Equal(t, 4, n)
	require.True(t, a.Equal(b))
	require.Equal(t, "1.3", b.String())
}

func TestMulExactGood(t *testing.T) {
	a, err := NewDecFromString("1.000001")
	require.NoError(t, err)
	b := NewDecFinite(1, 6)
	c, err := a.MulExact(b)
	require.NoError(t, err)
	d, err := c.Int64()
	require.NoError(t, err)
	require.Equal(t, int64(1000001), d)
}

func TestMulExactBad(t *testing.T) {
	a, err := NewDecFromString("1.000000000000000000000000000000000000123456789")
	require.NoError(t, err)
	b := NewDecFinite(1, 10)
	_, err = a.MulExact(b)
	require.ErrorIs(t, err, ErrUnexpectedRounding)
}

func TestQuoExactGood(t *testing.T) {
	a, err := NewDecFromString("1000001")
	require.NoError(t, err)
	b := NewDecFinite(1, 6)
	c, err := a.QuoExact(b)
	require.NoError(t, err)
	fmt.Println(c.String())
	require.Equal(t, "1.000001000000000000000000000000000", c.String())
}

func TestQuoExactBad(t *testing.T) {
	a, err := NewDecFromString("1000000000000000000000000000000000000123456789")
	require.NoError(t, err)
	b := NewDecFinite(1, 10)
	_, err = a.QuoExact(b)
	require.ErrorIs(t, err, ErrUnexpectedRounding)
}

func TestToBigInt(t *testing.T) {
	i1 := "1000000000000000000000000000000000000123456789"
	tcs := []struct {
		intStr  string
		out     string
		isError error
	}{
		{i1, i1, nil},
		{"1000000000000000000000000000000000000123456789.00000000", i1, nil},
		{"123.456e6", "123456000", nil},
		{"12345.6", "", ErrNonIntegeral},
	}
	for idx, tc := range tcs {
		a, err := NewDecFromString(tc.intStr)
		require.NoError(t, err)
		b, err := a.BigInt()
		if tc.isError == nil {
			require.NoError(t, err, "test_%d", idx)
			require.Equal(t, tc.out, b.String(), "test_%d", idx)
		} else {
			require.ErrorIs(t, err, tc.isError, "test_%d", idx)
		}
	}
}

func TestToSdkInt(t *testing.T) {
	i1 := "1000000000000000000000000000000000000123456789"
	tcs := []struct {
		intStr string
		out    string
	}{
		{i1, i1},
		{"1000000000000000000000000000000000000123456789.00000000", i1},
		{"123.456e6", "123456000"},
		{"123.456e1", "1234"},
		{"123.456", "123"},
		{"123.956", "123"},
		{"-123.456", "-123"},
		{"-123.956", "-123"},
		{"-0.956", "0"},
		{"-0.9", "0"},
	}
	for idx, tc := range tcs {
		a, err := NewDecFromString(tc.intStr)
		require.NoError(t, err)
		b := a.SdkIntTrim()
		require.Equal(t, tc.out, b.String(), "test_%d", idx)
	}
}

func TestToSdkLegacy(t *testing.T) {
	i1 := "1000000000000000000000000000000000000123456789.321000000000000000"
	tcs := []struct {
		intStr string
		out    string
	}{
		{i1, i1},
		{"1000000000000000000000000000000000000123456789.321000000000000000", i1},
		{"123.456e6", "123456000.000000000000000000"},
		{"123.456e1", "1234.560000000000000000"},
		{"123.456", "123.456000000000000000"},
		{"123.956", "123.956000000000000000"},
		{"-123.456", "-123.456000000000000000"},
		{"-123.956", "-123.956000000000000000"},
		{"-0.956", "-0.956000000000000000"},
		{"-0.9", "-0.900000000000000000"},
	}
	for idx, tc := range tcs {
		a, err := NewDecFromString(tc.intStr)
		require.NoError(t, err)
		b := a.SdkLegacyDec()
		require.Equal(t, tc.out, b.String(), "test_%d", idx)
	}
}
func TestInfDecString(t *testing.T) {
	_, err := NewDecFromString("iNf")
	require.Error(t, err)
	require.ErrorIs(t, err, ErrInfiniteString)
}

func TestInDelta(t *testing.T) {
	// Test cases
	testCases := []struct {
		expected       Dec
		result         Dec
		epsilon        Dec
		expectedResult bool
	}{
		{
			NewDecFromInt64(10),
			NewDecFromInt64(10),
			NewDecFromInt64(1),
			true,
		}, // expected == result, within epsilon
		{
			NewDecFromInt64(10),
			NewDecFromInt64(9),
			NewDecFromInt64(1),
			true,
		}, // expected != result, but within epsilon
		{
			NewDecFromInt64(10),
			NewDecFromInt64(8),
			NewDecFromInt64(1),
			false,
		}, // expected != result, outside epsilon
		{
			NewDecFromInt64(10),
			NewDecFromInt64(10),
			NewDecFromInt64(0),
			true,
		}, // expected == result, epsilon is zero
		{
			NewDecFromInt64(10),
			NewDecFromInt64(-10),
			NewDecFromInt64(1),
			false,
		}, // expected != result, outside epsilon
		{
			NewDecFromInt64(-10),
			NewDecFromInt64(-10),
			NewDecFromInt64(0),
			true,
		}, // expected == result, epsilon is zero
	}

	// Run test cases
	for _, tc := range testCases {
		result := InDelta(tc.expected, tc.result, tc.epsilon)
		require.Equal(t, tc.expectedResult, result)
	}
}

func TestSlicesInDelta(t *testing.T) {
	// Test cases
	testCases := []struct {
		name     string
		a        []Dec
		b        []Dec
		epsilon  Dec
		expected bool
	}{
		{
			name:     "Equal slices within epsilon",
			a:        []Dec{NewDecFromInt64(1), NewDecFromInt64(2), NewDecFromInt64(3)},
			b:        []Dec{NewDecFromInt64(1), NewDecFromInt64(2), NewDecFromInt64(3)},
			epsilon:  NewDecFromInt64(0),
			expected: true,
		},
		{
			name:     "Equal slices within epsilon",
			a:        []Dec{NewDecFromInt64(0), NewDecFromInt64(-1), NewDecFromInt64(4)},
			b:        []Dec{NewDecFromInt64(-1), NewDecFromInt64(-2), NewDecFromInt64(3)},
			epsilon:  NewDecFromInt64(1),
			expected: true,
		},
		{
			name:     "Equal slices NOT within epsilon",
			a:        []Dec{NewDecFromInt64(1), NewDecFromInt64(2), NewDecFromInt64(5)},
			b:        []Dec{NewDecFromInt64(1), NewDecFromInt64(2), NewDecFromInt64(3)},
			epsilon:  NewDecFromInt64(1),
			expected: false,
		},
		{
			name:     "Different slices within epsilon",
			a:        []Dec{NewDecFromInt64(-1), NewDecFromInt64(2), NewDecFromInt64(3)},
			b:        []Dec{NewDecFromInt64(2), NewDecFromInt64(5), NewDecFromInt64(6)},
			epsilon:  NewDecFromInt64(3),
			expected: true,
		},
		{
			name:     "Different slices outside epsilon",
			a:        []Dec{NewDecFromInt64(1), NewDecFromInt64(2), NewDecFromInt64(3)},
			b:        []Dec{NewDecFromInt64(4), NewDecFromInt64(5), NewDecFromInt64(6)},
			epsilon:  NewDecFromInt64(1),
			expected: false,
		},
		{
			name:     "Different slice lengths",
			a:        []Dec{NewDecFromInt64(1), NewDecFromInt64(2), NewDecFromInt64(3)},
			b:        []Dec{NewDecFromInt64(1), NewDecFromInt64(2)},
			epsilon:  NewDecFromInt64(0),
			expected: false,
		},
		{
			name:     "Empty slice",
			a:        []Dec{},
			b:        []Dec{},
			epsilon:  NewDecFromInt64(0),
			expected: true,
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := SlicesInDelta(tc.a, tc.b, tc.epsilon)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestSumDecSlice(t *testing.T) {
	// Test case 1: Empty slice
	x := []Dec{}
	expectedSum := ZeroDec()

	sum, err := SumDecSlice(x)
	require.NoError(t, err)
	require.True(t, sum.Equal(expectedSum), "Expected sum to be zero")

	// Test case 2: Slice with positive values
	x = []Dec{
		NewDecFromInt64(1),
		NewDecFromInt64(2),
		NewDecFromInt64(3),
	}
	expectedSum = NewDecFromInt64(6)

	sum, err = SumDecSlice(x)
	require.NoError(t, err)
	require.True(t, sum.Equal(expectedSum), "Expected sum to be 6")

	// Test case 3: Slice with negative values
	x = []Dec{
		NewDecFromInt64(-1),
		NewDecFromInt64(-2),
		NewDecFromInt64(-3),
	}
	expectedSum = NewDecFromInt64(-6)

	sum, err = SumDecSlice(x)
	require.NoError(t, err)
	require.True(t, sum.Equal(expectedSum), "Expected sum to be -6")

	// Test case 4: Slice with mixed positive and negative values
	x = []Dec{
		NewDecFromInt64(1),
		NewDecFromInt64(-2),
		NewDecFromInt64(3),
	}
	expectedSum = NewDecFromInt64(2)

	sum, err = SumDecSlice(x)
	require.NoError(t, err)
	require.True(t, sum.Equal(expectedSum), "Expected sum to be 2")
}

func TestDecReduce(t *testing.T) {
	// Test case 1
	x := NewDecFromInt64(12345678900)
	expectedY := NewDecFromInt64(123456789)
	expectedN := 2

	y, n := x.Reduce()

	require.Equal(t, expectedY.dec.Coeff, y.dec.Coeff)
	require.Equal(t, expectedN, n)

	// Test case 2
	x = NewDecFromInt64(0)
	expectedY = NewDecFromInt64(0)
	expectedN = 0

	y, n = x.Reduce()

	require.Equal(t, expectedY.dec.Coeff, y.dec.Coeff)
	require.Equal(t, expectedN, n)

	// Test case 3
	x = NewDecFromInt64(10000.000)
	expectedY = NewDecFromInt64(1)
	expectedN = 4

	y, n = x.Reduce()

	require.Equal(t, expectedY.dec.Coeff, y.dec.Coeff)
	require.Equal(t, expectedN, n)

	// Test case 4
	x = NewDecFromInt64(0000.000)
	expectedY = NewDecFromInt64(0)
	expectedN = 0

	y, n = x.Reduce()

	require.Equal(t, expectedY.dec.Coeff, y.dec.Coeff)
	require.Equal(t, expectedN, n)

	// Test case 5
	x = NewDecFromInt64(-1234560000.000)
	expectedY = NewDecFromInt64(123456)
	expectedN = 4

	y, n = x.Reduce()

	require.Equal(t, expectedY.dec.Coeff, y.dec.Coeff)
	require.True(t, y.dec.Negative)
	require.Equal(t, expectedN, n)

	// Test case 6
	x = NewDecFromInt64(-123456000000.000)
	expectedN = 6
	strX := "-123456000000"

	y, n = x.Reduce()

	require.Equal(t, strX, y.String())
	require.Equal(t, expectedN, n)
}
