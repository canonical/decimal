// Copyright 2015-2024 Canonical Ltd.

package decimal_test

import (
	"encoding/json"
	"fmt"
	"math"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/canonical/decimal"
)

// testCtx is a decimal context for testing
// with a sane number of decimal places.
var testCtx = decimal.NewContext(6)

func TestDecimalRound(t *testing.T) {
	t.Parallel()
	c := qt.New(t)
	tests := []struct {
		val      int64
		exp      int32
		roundTo  int32
		expected string
	}{
		{1234, -3, -2, "1.230000"},
		{1245, -3, -2, "1.250000"},
		{1229, -3, -2, "1.230000"},
		{123, 0, 1, "120.000000"},
		{12345, -1, 0, "1235.000000"},
		{12345, -1, 1, "1230.000000"},
		// Negative cases.
		{-1234, -3, -2, "-1.230000"},
		{-2245, -3, -2, "-2.250000"},
		{-3229, -3, -2, "-3.230000"},
		{-423, 0, 1, "-420.000000"},
		// Rounding up is no-op.
		{1234, -3, -4, "1.234000"},
		{1245, -3, -4, "1.245000"},
		{1229, -3, -4, "1.229000"},
		{123, 0, -1, "123.000000"},
	}

	for i, test := range tests {
		c.Run(fmt.Sprintf("test %d: decimal (%d, %d) rounded to %d", i, test.val, test.exp, test.roundTo), func(c *qt.C) {
			d := testCtx.New(test.val, test.exp).Round(test.roundTo)
			c.Assert(d.StringFixed(), qt.Equals, test.expected)
		})
	}
}

func TestDecimalString(t *testing.T) {
	t.Parallel()
	c := qt.New(t)
	tests := []struct {
		val      int64
		exp      int32
		expected string
	}{
		{50, 0, "50"},
		{-5, 1, "-50"},
		{1234, -2, "12.34"},
		{-4321, -2, "-43.21"},
		// The precision of this number is 6 fractional digits,
		// hence the result is 0.
		{1, -7, "0"},
		{12034, -3, "12.034"},
	}

	for i, test := range tests {
		c.Run(fmt.Sprintf("test %d: %d, %d", i, test.val, test.exp), func(c *qt.C) {
			d := testCtx.New(test.val, test.exp)
			c.Assert(d.String(), qt.Equals, test.expected)
		})
	}
}

func TestDecimalStringFixed(t *testing.T) {
	t.Parallel()
	c := qt.New(t)
	tests := []struct {
		val      int64
		exp      int32
		expected string
	}{
		{50, 0, "50.000000"},
		{-5, 1, "-50.000000"},
		{1234, -2, "12.340000"},
		{-4321, -2, "-43.210000"},
		{1, -7, "0.000000"},
	}

	for i, test := range tests {
		c.Run(fmt.Sprintf("test %d: %d, %d", i, test.val, test.exp), func(c *qt.C) {
			d := testCtx.New(test.val, test.exp)
			c.Assert(d.StringFixed(), qt.Equals, test.expected)
		})
	}
}

func TestAdd(t *testing.T) {
	t.Parallel()
	c := qt.New(t)
	tests := []struct {
		aVal     int64
		aExp     int32
		bVal     int64
		bExp     int32
		expected string
	}{
		{2, 0, 3, 0, "5.000000"},
		{1, -1, 1, -1, "0.200000"},
		{1, -1, -1, -1, "0.000000"},
		{0, 0, 1001, -3, "1.001000"},

		{2454495034, 0, 3451204593, 0, "5905699627.000000"},
		{2454495034, -5, 345120, -6, "24545.295460"},
	}

	ctx2 := decimal.NewContext(10)

	for i, test := range tests {
		c.Run(fmt.Sprintf("test case %d", i), func(c *qt.C) {
			a := testCtx.New(test.aVal, test.aExp)
			b := testCtx.New(test.bVal, test.bExp)
			sum := a.Add(b)
			c.Assert(sum.StringFixed(), qt.Equals, test.expected)
			c.Assert(sum.Context, qt.Equals, a.Context)

			// Run same test with different context.
			b = ctx2.New(test.bVal, test.bExp)
			sum = a.Add(b)
			c.Assert(sum.StringFixed(), qt.Equals, test.expected)
			c.Assert(sum.Context, qt.Equals, a.Context)
		})
	}
}

func TestSub(t *testing.T) {
	t.Parallel()
	c := qt.New(t)
	tests := []struct {
		aVal     int64
		aExp     int32
		bVal     int64
		bExp     int32
		expected string
	}{
		{2, 0, 3, 0, "-1.000000"},
		{12, 0, 3, 0, "9.000000"},
		{-2, 0, 9, 0, "-11.000000"},
		{1, -1, -1, -1, "0.200000"},
		{1, -1, 1, -1, "0.000000"},
		{0, 0, 1001, -3, "-1.001000"},
		{1001, -3, 0, 0, "1.001000"},
		{23, -1, 3, -1, "2.000000"},

		{2454495034, 0, 3451204593, 0, "-996709559.000000"},
		{2454495034, -5, 3451204593, -10, "24544.605220"},
	}

	ctx2 := decimal.NewContext(10)

	for i, test := range tests {
		c.Run(fmt.Sprintf("test case %d", i), func(c *qt.C) {
			a := testCtx.New(test.aVal, test.aExp)
			b := testCtx.New(test.bVal, test.bExp)
			sub := a.Sub(b)
			c.Assert(sub.StringFixed(), qt.Equals, test.expected)
			c.Assert(sub.Context, qt.Equals, a.Context)

			// Run same test with different context.
			b = ctx2.New(test.bVal, test.bExp)
			sub = a.Sub(b)
			c.Assert(sub.StringFixed(), qt.Equals, test.expected)
			c.Assert(sub.Context, qt.Equals, a.Context)
		})
	}
}

func TestMul(t *testing.T) {
	t.Parallel()
	c := qt.New(t)
	tests := []struct {
		aVal     int64
		aExp     int32
		bVal     int64
		bExp     int32
		expected string
	}{
		{2, 0, 3, 0, "6.000000"},
		{1, -1, 1, -1, "0.010000"},
		{0, 0, 1001, -3, "0.000000"},

		{2454495034, 0, 3451204593, 0, "8470964534836491162.000000"},
		{2454495034, -5, 3451204593, -3, "84709645348.364912"},
	}

	ctx2 := decimal.NewContext(10)

	for i, test := range tests {
		c.Run(fmt.Sprintf("test case %d", i), func(c *qt.C) {
			a := testCtx.New(test.aVal, test.aExp)
			b := testCtx.New(test.bVal, test.bExp)
			mul := a.Mul(b)
			c.Assert(mul.StringFixed(), qt.Equals, test.expected)
			c.Assert(mul.Context, qt.Equals, a.Context)

			// Run same test with different context.
			b = ctx2.New(test.bVal, test.bExp)
			mul = a.Mul(b)
			c.Assert(mul.StringFixed(), qt.Equals, test.expected)
			c.Assert(mul.Context, qt.Equals, a.Context)
		})
	}
}

func TestDecimalDiv(t *testing.T) {
	t.Parallel()
	c := qt.New(t)
	tests := []struct {
		aVal     int64
		aExp     int32
		bVal     int64
		bExp     int32
		expected string
	}{
		{6, 0, 3, 0, "2.000000"},
		{0, 0, 1, 0, "0.000000"},
		{10, 0, 2, 0, "5.000000"},
		{22, -1, 11, -1, "2.000000"},
		{-22, -1, -11, -1, "2.000000"},
		{1288, -2, 56, -1, "2.300000"},
		{1, 0, 3, 0, "0.333333"},
		{2, 0, 3, 0, "0.666667"}, // rounded
		{1, 4, 3, 0, "3333.333333"},

		{102342743555455, 0, -3, 0, "-34114247851818.333333"},
		{-4612301402398475334, -6, 235, -1, "-196268144782.913844"},
		{1023427554493, 0, 43432632, 0, "23563.562864"}, // rounded
		{1, 0, 434324545566634, 0, "0.000000"},

		{10, 10, 10, 1, "1000000000.000000"},
		{124, 10, 3, 1, "41333333333.333333"},
		{124, 10, 3, 0, "413333333333.333333"},
		{124, 10, 6, 1, "20666666666.666667"},
		{124, 10, 6, 0, "206666666666.666667"},
	}

	ctx2 := decimal.NewContext(10)

	for i, test := range tests {
		c.Run(fmt.Sprintf("test case %d", i), func(c *qt.C) {
			a := testCtx.New(test.aVal, test.aExp)
			b := testCtx.New(test.bVal, test.bExp)
			div := a.Div(b)
			c.Assert(div.StringFixed(), qt.Equals, test.expected)
			c.Assert(div.Context, qt.Equals, a.Context)

			// Run same test with different context.
			b = ctx2.New(test.bVal, test.bExp)
			div = a.Div(b)
			c.Assert(div.StringFixed(), qt.Equals, test.expected)
			c.Assert(div.Context, qt.Equals, a.Context)
		})
	}
}

func TestUninitialized(t *testing.T) {
	t.Parallel()
	c := qt.New(t)
	a := decimal.Decimal{}
	b := decimal.Decimal{}

	decs := []decimal.Decimal{
		a,
		a.Abs(),
		a.Add(b),
		a.Sub(b),
		a.Mul(b),
	}

	for i, d := range decs {
		c.Run(fmt.Sprintf("test case %d", i), func(c *qt.C) {
			c.Assert(d.String(), qt.Equals, "0")
			c.Assert(d.Cmp(b), qt.Equals, 0)
			c.Assert(d.Exponent(), qt.Equals, int32(0))
		})
	}
}

func TestMarshalUnmarshalJSON(t *testing.T) {
	t.Parallel()
	c := qt.New(t)
	tests := []struct {
		num      decimal.Decimal
		expected string
	}{
		{testCtx.MustNewFromString("1.234567"), "1.234567"},
		{decimal.New(12345678987654321, -16), "1.2345678987654321"},
	}

	for i, test := range tests {
		c.Run(fmt.Sprintf("test case %d", i), func(c *qt.C) {
			str, err := json.Marshal(test.num)
			c.Assert(err, qt.IsNil)
			c.Assert(string(str), qt.Equals, fmt.Sprintf(`"%s"`, test.expected))

			num2 := decimal.Decimal{}
			err = json.Unmarshal(str, &num2)
			c.Assert(err, qt.IsNil)
			c.Assert(test.num.Cmp(num2), qt.Equals, 0)
		})
	}
}

func TestOverflow(t *testing.T) {
	c := qt.New(t)
	ctx := decimal.NewContext(math.MaxInt32)
	c.Assert(func() {
		ctx.New(1, math.MinInt32).Mul(ctx.New(1, math.MinInt32))
	}, qt.PanicMatches, "exponent .* overflows an int32")
	c.Assert(func() {
		ctx.New(1, math.MaxInt32).Mul(ctx.New(1, math.MaxInt32))
	}, qt.PanicMatches, "exponent .* overflows an int32")
}

func TestCmp(t *testing.T) {
	t.Parallel()
	c := qt.New(t)
	tests := []struct {
		aVal     int64
		aExp     int32
		bVal     int64
		bExp     int32
		expected int
	}{
		{123, 3, -1234, 2, 1},
		{123, 3, 1234, 2, -1},
		{1230, 2, 123, 3, 0},
		{1, -12, 2, -12, 0}, // Equal due to rounding.
	}

	ctx2 := decimal.NewContext(10)

	for i, test := range tests {
		c.Run(fmt.Sprintf("test case %d", i), func(c *qt.C) {
			a := testCtx.New(test.aVal, test.aExp)
			b := testCtx.New(test.bVal, test.bExp)
			r := a.Cmp(b)
			c.Assert(r, qt.Equals, test.expected)

			// Run the same test with another context.
			b = ctx2.New(test.bVal, test.bExp)
			r = a.Cmp(b)
			c.Assert(r, qt.Equals, test.expected)
		})
	}
}

func TestComparisonMethods(t *testing.T) {
	t.Parallel()
	c := qt.New(t)
	type cmpFunc func(d1, d2 decimal.Decimal) bool
	gt := func(d1, d2 decimal.Decimal) bool { return d1.GT(d2) }
	gte := func(d1, d2 decimal.Decimal) bool { return d1.GTE(d2) }
	lt := func(d1, d2 decimal.Decimal) bool { return d1.LT(d2) }
	lte := func(d1, d2 decimal.Decimal) bool { return d1.LTE(d2) }

	tests := []struct {
		cmp      cmpFunc
		aVal     int64
		aExp     int32
		bVal     int64
		bExp     int32
		expected bool
	}{
		// Test GT.
		{gt, 123, 3, -1234, 2, true},
		{gt, 123, 3, 1234, 2, false},
		{gt, 1230, 2, 123, 3, false},
		{gt, 1, -12, 2, -12, false}, // Equal due to rounding.

		// Test GTE.
		{gte, 123, 3, -1234, 2, true},
		{gte, 123, 3, 1234, 2, false},
		{gte, 1230, 2, 123, 3, true},
		{gte, 1, -12, 2, -12, true}, // Equal due to rounding.

		// Test LT.
		{lt, 123, 3, -1234, 2, false},
		{lt, 123, 3, 1234, 2, true},
		{lt, 1230, 2, 123, 3, false},
		{lt, 1, -12, 2, -12, false}, // Equal due to rounding.

		// Test LTE.
		{lte, 123, 3, -1234, 2, false},
		{lte, 123, 3, 1234, 2, true},
		{lte, 1230, 2, 123, 3, true},
		{lte, 1, -12, 2, -12, true}, // Equal due to rounding.
	}

	ctx2 := decimal.NewContext(10)

	for i, test := range tests {
		c.Run(fmt.Sprintf("test case %d", i), func(c *qt.C) {
			a := testCtx.New(test.aVal, test.aExp)
			b := testCtx.New(test.bVal, test.bExp)
			r := test.cmp(a, b)
			c.Assert(r, qt.Equals, test.expected)

			// Run the same test with another context.
			b = ctx2.New(test.bVal, test.bExp)
			r = test.cmp(a, b)
			c.Assert(r, qt.Equals, test.expected)
		})
	}
}

func TestMaxMin(t *testing.T) {
	t.Parallel()
	c := qt.New(t)

	type method string
	const (
		min method = "Min"
		max method = "Max"
	)

	tests := []struct {
		about    string
		d1, d2   decimal.Decimal
		method   method
		expected decimal.Decimal
	}{{
		about:    "min compare when d1 < d2",
		d1:       testCtx.MustNewFromString("2"),
		d2:       testCtx.MustNewFromString("3"),
		method:   min,
		expected: testCtx.MustNewFromString("2"),
	}, {
		about:    "min compare when d1 == d2",
		d1:       testCtx.MustNewFromString("4"),
		d2:       testCtx.MustNewFromString("4"),
		method:   min,
		expected: testCtx.MustNewFromString("4"),
	}, {
		about:    "min compare when d1 > d2",
		d1:       testCtx.MustNewFromString("3"),
		d2:       testCtx.MustNewFromString("2"),
		method:   min,
		expected: testCtx.MustNewFromString("2"),
	}, {
		about:    "max compare when d1 > d2",
		d1:       testCtx.MustNewFromString("7"),
		d2:       testCtx.MustNewFromString("6"),
		method:   max,
		expected: testCtx.MustNewFromString("7"),
	}, {
		about:    "max compare when d1 == d2",
		d1:       testCtx.MustNewFromString("8"),
		d2:       testCtx.MustNewFromString("8"),
		method:   max,
		expected: testCtx.MustNewFromString("8"),
	}, {
		about:    "max compare when d1 < d2",
		d1:       testCtx.MustNewFromString("9"),
		d2:       testCtx.MustNewFromString("10"),
		method:   max,
		expected: testCtx.MustNewFromString("10"),
	}}
	for _, test := range tests {
		c.Run(test.about, func(c *qt.C) {
			var result decimal.Decimal
			switch test.method {
			case min:
				result = test.d1.Min(test.d2)
			case max:
				result = test.d1.Max(test.d2)
			default:
				panic("unexpected method value")
			}

			c.Assert(result.Equal(test.expected), qt.IsTrue)
		})
	}
}

func TestAbs(t *testing.T) {
	a := testCtx.New(-1234, -4)
	b := testCtx.New(1234, -4)

	c := a.Abs()
	qt.Assert(t, c.Cmp(b), qt.Equals, 0)
}

func TestStringScaled(t *testing.T) {
	t.Parallel()
	c := qt.New(t)
	tests := []struct {
		value   int64
		exp     int32
		rescale int32
		before  string
		after   string
	}{
		{1234, -3, -5, "1.234000", "1.234000"},
		{1234, -3, 0, "1.000000", "1.000000"},
		{1234, 3, 0, "1234000.000000", "1234000.000000"},
		{1234, -4, -4, "0.123400", "0.123400"},
		{1234, 0, 0, "1234.000000", "1234.000000"},

		{-1234, -3, -5, "-1.234000", "-1.234000"},
		{-1234, -3, 0, "-1.000000", "-1.000000"},
		{-1234, 3, 0, "-1234000.000000", "-1234000.000000"},
		{-1234, -4, -4, "-0.123400", "-0.123400"},
		{-1234, 0, 0, "-1234.000000", "-1234.000000"},
	}

	for i, test := range tests {
		c.Run(fmt.Sprintf("test case %d", i), func(c *qt.C) {
			s2 := testCtx.New(test.value, test.exp).StringScaled(test.rescale)
			c.Assert(s2, qt.Equals, test.after)
		})
	}
}

func TestTruncatePanic(t *testing.T) {
	a := testCtx.New(1, 1)
	qt.Assert(t, func() {
		a.Truncate(-1)
	}, qt.PanicMatches, "precision must be non-negative")
}

func TestScan(t *testing.T) {
	t.Parallel()
	c := qt.New(t)
	d := decimal.MustNewFromString("123.01")
	value, err := d.Value()
	c.Assert(err, qt.IsNil)

	// Serialized value must be able to be scanned back properly.
	got := new(decimal.Decimal)
	err = got.Scan(value)
	c.Assert(err, qt.IsNil)
	c.Assert(got, qt.DeepEquals, &d)
}
