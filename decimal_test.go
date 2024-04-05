package decimal_test

import (
	"math"
	"testing"

	"github.com/canonical/decimal"
)

// A decimal context for testing with a sane number of decimal places.
var testCtx = decimal.NewContext(6)

func TestDecimalRound(t *testing.T) {
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
		// Rounding up is a nop
		{1234, -3, -4, "1.234000"},
		{1245, -3, -4, "1.245000"},
		{1229, -3, -4, "1.229000"},
		{123, 0, -1, "123.000000"},
	}

	for i, test := range tests {
		t.Logf("test %d: decimal (%d, %d) rounded to %d",
			i, test.val, test.exp, test.roundTo)
		d := testCtx.New(test.val, test.exp).Round(test.roundTo)

		if d.StringFixed() != test.expected {
			t.Errorf("expected %s, got %s", test.expected, d.StringFixed())
		}
	}
}

func TestDecimalString(t *testing.T) {
	tests := []struct {
		val      int64
		exp      int32
		expected string
	}{
		{50, 0, "50"},
		{-5, 1, "-50"},
		{1234, -2, "12.34"},
		{-4321, -2, "-43.21"},
		// The precision of this number is 6 fractional digits, hence the result
		// is 0.
		{1, -7, "0"},
		{12034, -3, "12.034"},
	}

	for i, test := range tests {
		t.Logf("test %d: %d, %d", i, test.val, test.exp)
		d := testCtx.New(test.val, test.exp)
		if d.String() != test.expected {
			t.Fatalf("expected %q, go %q", test.expected, d.String())
		}
	}
}

func TestDecimalStringFixed(t *testing.T) {
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
		t.Logf("test %d: %d, %d", i, test.val, test.exp)
		d := testCtx.New(test.val, test.exp)
		if d.StringFixed() != test.expected {
			t.Fatalf("expected %q, go %q", test.expected, d.StringFixed())
		}
	}
}

func TestAdd(t *testing.T) {
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
		t.Logf("test %d", i)
		a := testCtx.New(test.aVal, test.aExp)
		b := testCtx.New(test.bVal, test.bExp)

		c := a.Add(b)
		if c.StringFixed() != test.expected {
			t.Errorf("expected %s, got %s", test.expected, c.StringFixed())
		}
		if c.Context != a.Context {
			t.Errorf("context not maintained")
		}

		// Run same test with different context.
		b = ctx2.New(test.bVal, test.bExp)
		c = a.Add(b)
		if c.StringFixed() != test.expected {
			t.Errorf("expected %s, got %s", test.expected, c.StringFixed())
		}
		if c.Context != a.Context {
			t.Errorf("context not maintained")
		}

	}
}

func TestSub(t *testing.T) {
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
		t.Logf("test %d", i)
		a := testCtx.New(test.aVal, test.aExp)
		b := testCtx.New(test.bVal, test.bExp)

		c := a.Sub(b)
		if c.StringFixed() != test.expected {
			t.Errorf("expected %s, got %s", test.expected, c.StringFixed())
		}
		if c.Context != a.Context {
			t.Errorf("context not maintainted during operation")
		}

		// Run same test with different context.
		b = ctx2.New(test.bVal, test.bExp)
		c = a.Sub(b)
		if c.StringFixed() != test.expected {
			t.Errorf("expected %s, got %s", test.expected, c.StringFixed())
		}
		if c.Context != a.Context {
			t.Errorf("context not maintained")
		}

	}
}

func TestMul(t *testing.T) {
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
		t.Logf("test %d", i)
		a := testCtx.New(test.aVal, test.aExp)
		b := testCtx.New(test.bVal, test.bExp)

		c := a.Mul(b)
		if c.StringFixed() != test.expected {
			t.Errorf("expected %s, got %s", test.expected, c.StringFixed())
		}
		if c.Context != a.Context {
			t.Errorf("context not maintainted during operation")
		}

		// Run same test with different context.
		b = ctx2.New(test.bVal, test.bExp)
		c = a.Mul(b)
		if c.StringFixed() != test.expected {
			t.Errorf("expected %s, got %s", test.expected, c.StringFixed())
		}
		if c.Context != a.Context {
			t.Errorf("context not maintained")
		}

	}
}

func TestDecimalDiv(t *testing.T) {
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
		t.Logf("test %d", i)
		a := testCtx.New(test.aVal, test.aExp)
		b := testCtx.New(test.bVal, test.bExp)

		c := a.Div(b)
		if c.StringFixed() != test.expected {
			t.Errorf("expected %s, got %s", test.expected, c.StringFixed())
		}
		if c.Context != a.Context {
			t.Errorf("context not maintainted during operation")
		}

		// Run same test with different context.
		b = ctx2.New(test.bVal, test.bExp)
		c = a.Div(b)
		if c.StringFixed() != test.expected {
			t.Errorf("expected %s, got %s", test.expected, c.StringFixed())
		}
		if c.Context != a.Context {
			t.Errorf("context not maintained")
		}

	}
}

func TestUninitialized(t *testing.T) {
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
		t.Logf("test %d", i)
		if d.String() != "0" {
			t.Errorf("expected 0, got %s", d.String())
		}
		if d.Cmp(b) != 0 {
			t.Errorf("%s != %s", d.String(), b.String())
		}
		if d.Exponent() != 0 {
			t.Errorf("a.Exponent() != 0")
		}
	}
}

func TestOverflow(t *testing.T) {
	c := decimal.NewContext(math.MaxInt32)
	if !panicMatches(t, func() { c.New(1, math.MinInt32).Mul(c.New(1, math.MinInt32)) }, "exponent .* overflows an int32!") {
		t.Fatalf("should have gotten an overflow panic")
	}
	if !panicMatches(t, func() { c.New(1, math.MaxInt32).Mul(c.New(1, math.MaxInt32)) }, "exponent .* overflows an int32!") {
		t.Fatalf("should have gotten an overflow panic")
	}
}

func TestCmp(t *testing.T) {
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
		{1, -12, 2, -12, 0}, // equal due to rounding
	}

	ctx2 := decimal.NewContext(10)

	for i, test := range tests {
		t.Logf("test %d", i)
		a := testCtx.New(test.aVal, test.aExp)
		b := testCtx.New(test.bVal, test.bExp)
		r := a.Cmp(b)
		if r != test.expected {
			t.Errorf("comparing %s and %s expected %d, got %d", a, b, test.expected, r)
		}
		b = ctx2.New(test.bVal, test.bExp)
		r = a.Cmp(b)
		if r != test.expected {
			t.Errorf("comparing %s and %s expected %d, got %d", a, b, test.expected, r)
		}

	}
}

func TestAbs(t *testing.T) {
	a := testCtx.New(-1234, -4)
	b := testCtx.New(1234, -4)

	c := a.Abs()
	if c.Cmp(b) != 0 {
		t.Errorf("error")
	}
}
