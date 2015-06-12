package decimal_test

import (
	"math"
	"math/big"
	"regexp"
	"testing"

	"github.com/CanonicalLtd/decimal"
)

func TestSumMaintainsContext(t *testing.T) {
	ctx := decimal.NewContext(5)

	a := testCtx.New(1, 1)
	b := ctx.New(5, 2)

	c := a.Add(b)

	// The context is taken from the first operand.
	if c.Context != a.Context {
		t.Fatalf("expected result of sum to maintain context.")
	}
	if c.String() != "510.000000" {
		t.Fatalf("unexpected result %q", c.String())
	}
}

func TestNew(t *testing.T) {
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
	for i, test := range tests {
		t.Logf("test %d", i)
		d := testCtx.New(test.value, test.exp)
		if d.String() != test.expected {
			t.Fatalf("expected %q, got %q", test.expected, d.String())
		}
		if d.Context != testCtx {
			t.Fatalf("context of decimal does not match creation context")
		}
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
	for i, s := range stringErrorTests {
		t.Logf("test %d: %q", i, s)
		_, err := testCtx.NewFromString(s)
		if !errorMatches(t, err, "can't convert .* to decimal.*") {
			t.Errorf("error expected when parsing %s", s)
		}

		if !panicMatches(t, func() { testCtx.MustNewFromString(s) }, "can't convert .* to decimal.*") {
			t.Errorf("expected panic when parsing %s", s)
		}

	}

}

func TestMustNewFromString(t *testing.T) {
	for src, expect := range stringTests {
		t.Logf("testing string parsing for %q", src)
		d := testCtx.MustNewFromString(src)
		if d.String() != expect {
			t.Errorf("expected %s, got %s", expect, d.String())
		}
	}
}

func errorMatches(t *testing.T, err error, exp string) bool {
	if err == nil {
		return false
	}
	r := regexp.MustCompile("^" + exp + "$")
	var ok bool
	if ok = r.MatchString(err.Error()); !ok {
		t.Logf("%v does not match %q", err, exp)
	}
	return ok
}

// panicMatches checks if the provided function panics.
func panicMatches(t *testing.T, f func(), exp string) (ret bool) {
	ret = false
	r := regexp.MustCompile("^" + exp + "$")
	defer func() {
		if message := recover(); message != nil {
			switch message.(type) {
			case error:
				if ret = r.MatchString((message.(error)).Error()); !ret {
					t.Logf("%q did not match %q", (message.(error)).Error(), exp)
				}
			case string:
				if ret = r.MatchString((message.(string))); !ret {
					t.Logf("%q did not match %q", message.(string), exp)
				}
			default:
				ret = false
			}
		}
	}()

	// call the target function
	f()

	return false
}

func TestMonetaryContextContextFromString(t *testing.T) {
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

	for src, expect := range tests {
		t.Logf("testing string parsing for %q", src)
		d, err := decimal.MonetaryContext.NewFromString(src)
		if err != nil {
			t.Errorf("error while parsing %s: %v", src, err)
		} else if d.String() != expect {
			t.Errorf("expected %s, got %s", expect, d.String())
		}
	}
}

func TestMonetaryContextContextFromStringErrors(t *testing.T) {
	for i, s := range stringErrorTests {
		t.Logf("test %d: %q", i, s)
		_, err := decimal.MonetaryContext.NewFromString(s)
		if !errorMatches(t, err, "can't convert .* to decimal.*") {
			t.Errorf("error expected when parsing %s", s)
		}

		if !panicMatches(t, func() { decimal.MonetaryContext.MustNewFromString(s) }, "can't convert .* to decimal.*") {
			t.Errorf("expected panic when parsing %s", s)
		}

	}
}

func TestNewFromRat(t *testing.T) {
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
	for i, test := range tests {
		t.Logf("test %d: %s", i, test.about)
		d := testCtx.NewFromRat(big.NewRat(test.num, test.denom))
		if d.String() != test.expect {
			t.Errorf("expected %q, got %q", test.expect, d.String())
		}
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
	for flt, expected := range floatStringTests {
		t.Logf("running test: %f -> %s", flt, expected)
		d, err := testCtx.NewFromFloat(flt)
		if err != nil {
			t.Errorf("unexpected error %v", err)
		}
		if d.String() != expected {
			t.Errorf("expected %s, got %s", expected, d.String())
		}
		if d.Context != testCtx {
			t.Errorf("unknown context value")
		}
	}
}

func TestNewFromString(t *testing.T) {
	for src, expect := range stringTests {
		t.Logf("testing string parsing for %q", src)
		d, err := testCtx.NewFromString(src)
		if err != nil {
			t.Errorf("error while parsing %s", src)
		} else if d.String() != expect {
			t.Errorf("expected %s, got %s", expect, d.String())
		}
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
	for src, expect := range floatStringTestsLong {
		t.Logf("testing string parsing for %q", src)
		d, err := decimal.DefaultContext.NewFromFloat(src)
		if err != nil {
			t.Errorf("error while parsing %s", src)
		} else if d.String() != expect {
			t.Errorf("expected %s, got %s", expect, d.String())
		}
	}
}

func TestNewFromFloatError(t *testing.T) {
	shouldPanicOn := []float64{
		math.NaN(),
		math.Inf(1),
		math.Inf(-1),
	}

	for i, flt := range shouldPanicOn {
		t.Logf("test %d: %f with NewFromFloat", i, flt)
		_, err := testCtx.NewFromFloat(flt)
		if !errorMatches(t, err, `can't convert .* to decimal`) {
			t.Fatalf("expected error when creating a Decimal from %v", flt)
		}

		t.Logf("test %d: %f with MustNewFromFloat", i, flt)
		var d decimal.Decimal

		if !panicMatches(t, func() { d = testCtx.MustNewFromFloat(flt) }, `can't convert .* to decimal`) {
			t.Fatalf("expected panic when creating a Decimal from %v, got %v instead", flt, d.String())
		}
	}
}
