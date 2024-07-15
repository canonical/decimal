// Copyright 2015-2024 Canonical Ltd.

package decimal_test

import (
	"fmt"
	"math"
	"math/big"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/canonical/decimal"
)

func TestSumMaintainsContext(t *testing.T) {
	c := qt.New(t)
	ctx := decimal.NewContext(5)

	a := testCtx.New(1, 1) // 10 = 1 * 10.
	b := ctx.New(5, 2)     // 500 = 5 * (10 ^ 2).
	sum := a.Add(b)

	// The context is taken from the first operand.
	c.Assert(sum.Context, qt.Equals, a.Context)
	c.Assert(sum.StringFixed(), qt.Equals, "510.000000") // Preserving 6 decimal points from textCtx.
}

func TestNegativeContext(t *testing.T) {
	c := qt.New(t)
	ctx := decimal.NewContext(-2) // Precision to hundreds.

	a := ctx.MustNewFromString("1234.056")
	c.Assert(a.StringFixed(), qt.Equals, "1200")
}

func TestNew(t *testing.T) {
	t.Parallel()
	c := qt.New(t)
	tests := []struct {
		value    int64
		exp      int32
		expected string
	}{
		{1, 0, "1.000000"},
		{2, -5, "0.000020"},
		{123456, -6, "0.123456"},
		{1, -7, "0.000000"},
	}
	for idx, test := range tests {
		c.Run(fmt.Sprintf("test case %d: %s", idx, test.expected), func(c *qt.C) {
			d := testCtx.New(test.value, test.exp)
			c.Assert(d.StringFixed(), qt.Equals, test.expected)
			c.Assert(d.Context, qt.Equals, testCtx)
		})
	}
}

var stringTests = map[string]string{
	"3.141592653589793":   "3.141593",
	"3":                   "3.000000",
	"1234567890123456":    "1234567890123456.000000",
	"1234567890123456000": "1234567890123456000.000000",
	"1234.567890123456":   "1234.567890",
	"0.1234567890123456":  "0.123457",
	"0":                   "0.000000",
	"0.111111111111111":   "0.111111",
	"0.1111111111111111":  "0.111111",
	"0.1111111111111119":  "0.111111",
	"0.0000000000000001":  "0.000000",
	"0.0000000000000002":  "0.000000",
	"0.0000000000000003":  "0.000000",
	"0.0000000000000005":  "0.000000",
	"0.0000000000000008":  "0.000000",
	"0.1000000000000001":  "0.100000",
	"0.1000000000000002":  "0.100000",
	"0.1000000000000003":  "0.100000",
	"0.1000000000000005":  "0.100000",
	"0.1000000000000008":  "0.100000",
}

var stringErrorTests = []string{
	"",
	"qwert",
	"-",
	".",
	"-.",
	".-",
	"234-.56",
	"234-56",
	"2-",
	"..",
	"2..",
	"..2",
	".5.2",
	"8..2",
	"8.1.",
}

func TestNewFromStringError(t *testing.T) {
	t.Parallel()
	c := qt.New(t)
	for idx, s := range stringErrorTests {
		c.Run(fmt.Sprintf("test case %d: %s", idx, s), func(c *qt.C) {
			_, err := testCtx.NewFromString(s)
			c.Assert(err, qt.ErrorMatches, "can't convert .* to decimal.*")
			c.Assert(func() {
				testCtx.MustNewFromString(s)
			}, qt.PanicMatches, "can't convert .* to decimal.*")
		})
	}
}

func TestMustNewFromString(t *testing.T) {
	t.Parallel()
	c := qt.New(t)
	for src, expected := range stringTests {
		c.Run(fmt.Sprintf("test %s -> %s", src, expected), func(c *qt.C) {
			d := testCtx.MustNewFromString(src)
			c.Assert(d.StringFixed(), qt.Equals, expected)
		})
	}
}

func TestMonetaryContextContextFromString(t *testing.T) {
	t.Parallel()
	c := qt.New(t)
	tests := map[string]string{
		"3.141592653589793":   "3.14",
		"3":                   "3.00",
		"1234567890123456":    "1234567890123456.00",
		"1234567890123456000": "1234567890123456000.00",
		"1234.567890123456":   "1234.57",
		"0.1234567890123456":  "0.12",
		"0":                   "0.00",
		"0.111111111111111":   "0.11",
		"0.1111111111111119":  "0.11",
		"0.0000000000000001":  "0.00",
		"0.1000000000000001":  "0.10",
	}

	for src, expected := range tests {
		c.Run(fmt.Sprintf("test %s -> %s", src, expected), func(c *qt.C) {
			d, err := decimal.MonetaryContext.NewFromString(src)
			c.Assert(err, qt.IsNil)
			c.Assert(d.StringFixed(), qt.Equals, expected)
		})
	}
}

func TestMonetaryContextContextFromStringErrors(t *testing.T) {
	t.Parallel()
	c := qt.New(t)
	for idx, s := range stringErrorTests {
		c.Run(fmt.Sprintf("test case %d: %s", idx, s), func(c *qt.C) {
			_, err := decimal.MonetaryContext.NewFromString(s)
			c.Assert(err, qt.ErrorMatches, "can't convert .* to decimal.*")
			c.Assert(func() {
				decimal.MonetaryContext.MustNewFromString(s)
			}, qt.PanicMatches, "can't convert .* to decimal.*")
		})
	}
}

func TestNewFromRat(t *testing.T) {
	t.Parallel()
	c := qt.New(t)
	tests := []struct {
		about  string
		num    int64
		denom  int64
		expect string
	}{{
		about:  "integer",
		num:    15,
		denom:  1,
		expect: "15.000000",
	}, {
		about:  "decimal",
		num:    15,
		denom:  1000,
		expect: "0.015000",
	}, {
		about:  "irrational",
		num:    10,
		denom:  6,
		expect: "1.666667",
	}, {
		about:  "negative integer",
		num:    -15,
		denom:  1,
		expect: "-15.000000",
	}}
	for _, test := range tests {
		c.Run(test.about, func(c *qt.C) {
			d := testCtx.NewFromRat(big.NewRat(test.num, test.denom))
			c.Assert(d.StringFixed(), qt.Equals, test.expect)
		})
	}
}

var floatStringTests = map[float64]string{
	3.141592653589793:   "3.141593",
	3:                   "3.000000",
	1234567890123456:    "1234567890123456.000000",
	1234567890123456000: "1234567890123456000.000000",
	1234.567890123456:   "1234.567890",
	.1234567890123456:   "0.123457",
	0:                   "0.000000",
	.1111111111111110:   "0.111111",
	.1111111111111111:   "0.111111",
	.1111111111111119:   "0.111111",
	.0000000000000001:   "0.000000",
	.0000000000000002:   "0.000000",
	.0000000000000003:   "0.000000",
	.0000000000000005:   "0.000000",
	.0000000000000008:   "0.000000",
	.1000000000000001:   "0.100000",
	.1000000000000002:   "0.100000",
	.1000000000000003:   "0.100000",
	.1000000000000005:   "0.100000",
	.1000000000000008:   "0.100000",

	-3.141592653589793:   "-3.141593",
	-3:                   "-3.000000",
	-1234567890123456:    "-1234567890123456.000000",
	-1234567890123456000: "-1234567890123456000.000000",
	-1234.567890123456:   "-1234.567890",
	-.1234567890123456:   "-0.123457",
	-.1111111111111110:   "-0.111111",
	-.1111111111111111:   "-0.111111",
	-.1111111111111119:   "-0.111111",
	-.0000000000000001:   "0.000000",
	-.0000000000000002:   "0.000000",
	-.0000000000000003:   "0.000000",
	-.0000000000000005:   "0.000000",
	-.0000000000000008:   "0.000000",
	-.1000000000000001:   "-0.100000",
	-.1000000000000002:   "-0.100000",
	-.1000000000000003:   "-0.100000",
	-.1000000000000005:   "-0.100000",
	-.1000000000000008:   "-0.100000",
}

func TestNewFromFloat(t *testing.T) {
	t.Parallel()
	c := qt.New(t)
	for flt, expected := range floatStringTests {
		c.Run(fmt.Sprintf("test %f -> %s", flt, expected), func(c *qt.C) {
			d, err := testCtx.NewFromFloat(flt)
			c.Assert(err, qt.IsNil)
			c.Assert(d.StringFixed(), qt.Equals, expected)
			c.Assert(d.Context, qt.Equals, testCtx)
		})
	}
}

func TestNewFromString(t *testing.T) {
	t.Parallel()
	c := qt.New(t)
	for src, expected := range stringTests {
		c.Run(fmt.Sprintf("test %s -> %s", src, expected), func(c *qt.C) {
			d, err := testCtx.NewFromString(src)
			c.Assert(err, qt.IsNil)
			c.Assert(d.StringFixed(), qt.Equals, expected)
			c.Assert(d.Context, qt.Equals, testCtx)
		})
	}
}

var floatStringTestsLong = map[float64]string{
	3.141592653589793:   "3.1415926535897930",
	3:                   "3.0000000000000000",
	1234567890123456:    "1234567890123456.0000000000000000",
	1234567890123456000: "1234567890123456000.0000000000000000",
	1234.567890123456:   "1234.5678901234560000",
	.1234567890123456:   "0.1234567890123456",
	0:                   "0.0000000000000000",
	.1111111111111110:   "0.1111111111111110",
	.1111111111111111:   "0.1111111111111111",
	.1111111111111119:   "0.1111111111111119",
	.0000000000000001:   "0.0000000000000001",
	.0000000000000002:   "0.0000000000000002",
	.0000000000000003:   "0.0000000000000003",
	.0000000000000005:   "0.0000000000000005",
	.00000000000000008:  "0.0000000000000001",
	.1000000000000001:   "0.1000000000000001",
	.1000000000000002:   "0.1000000000000002",
	.1000000000000003:   "0.1000000000000003",
	.1000000000000005:   "0.1000000000000005",
	.1000000000000008:   "0.1000000000000008",
}

func TestNewFromFloatDefaultContext(t *testing.T) {
	t.Parallel()
	c := qt.New(t)
	for src, expected := range floatStringTestsLong {
		c.Run(fmt.Sprintf("test %f -> %s", src, expected), func(c *qt.C) {
			d, err := decimal.DefaultContext.NewFromFloat(src)
			c.Assert(err, qt.IsNil)
			c.Assert(d.StringFixed(), qt.Equals, expected)
		})
	}
}

func TestNewFromFloatError(t *testing.T) {
	c := qt.New(t)
	shouldPanicOn := []float64{
		math.NaN(),
		math.Inf(1),
		math.Inf(-1),
	}

	for idx, flt := range shouldPanicOn {
		c.Run(fmt.Sprintf("test case %d", idx), func(c *qt.C) {
			_, err := testCtx.NewFromFloat(flt)
			c.Assert(err, qt.ErrorMatches, "can't convert .* to decimal")
			c.Assert(func() {
				testCtx.MustNewFromFloat(flt)
			}, qt.PanicMatches, "can't convert .* to decimal")
		})
	}
}
