package decimal

import (
	"fmt"
	"math"
	"math/big"
	"testing"
)

var tests = map[float64]string{
	3.141592653589793:   "3.141592653589793",
	3:                   "3",
	1234567890123456:    "1234567890123456",
	1234567890123456000: "1234567890123456000",
	1234.567890123456:   "1234.567890123456",
	.1234567890123456:   "0.1234567890123456",
	0:                   "0",
	.1111111111111110:   "0.111111111111111",
	.1111111111111111:   "0.1111111111111111",
	.1111111111111119:   "0.1111111111111119",
	.000000000000000001: "0.000000000000000001",
	.000000000000000002: "0.000000000000000002",
	.000000000000000003: "0.000000000000000003",
	.000000000000000005: "0.000000000000000005",
	.000000000000000008: "0.000000000000000008",
	.1000000000000001:   "0.1000000000000001",
	.1000000000000002:   "0.1000000000000002",
	.1000000000000003:   "0.1000000000000003",
	.1000000000000005:   "0.1000000000000005",
	.1000000000000008:   "0.1000000000000008",
}

func TestNewFromFloat(t *testing.T) {

	for flt, expected := range tests {
		t.Logf("running test: %f -> %s", flt, expected)
		d := NewFromFloat(flt)
		if d.String() != expected {
			t.Errorf("expected %s, got %s", expected, d.String())
		}
		// Test negative values too.
		if flt != 0 {
			flt = -flt
			expected = fmt.Sprintf("-%s", expected)
			t.Logf("running test: %f -> %s", flt, expected)
			d = NewFromFloat(flt)
			if d.String() != expected {
				t.Errorf("expected %s, got %s", expected, d.String())
			}
		}
	}
}

func TestNewFromFloatPanics(t *testing.T) {
	shouldPanicOn := []float64{
		math.NaN(),
		math.Inf(1),
		math.Inf(-1),
	}

	for i, flt := range shouldPanicOn {
		t.Logf("test %d: %f with NewFromFloat", i, flt)
		var d Decimal
		if !didPanic(func() { d = NewFromFloat(flt) }) {
			t.Fatalf("expected panic when creating a Decimal from %v, got %v instead", flt, d.String())
		}

		t.Logf("test %d: %f with NewFromFloatWithExponent", i, flt)
		if !didPanic(func() { d = NewFromFloatWithExponent(flt, 0) }) {
			t.Fatalf("Expected panic when creating a Decimal from %v, got %v instead", flt, d.String())
		}

	}
}

func TestNewFromString(t *testing.T) {
	for _, s := range tests {
		t.Logf("testing string parsing for %q", s)
		d, err := NewFromString(s)
		if err != nil {
			t.Errorf("error while parsing %s", s)
		} else if d.String() != s {
			t.Errorf("expected %s, got %s", s, d.String())
		}
	}
}

func TestMustNewFromString(t *testing.T) {
	for _, s := range tests {
		t.Logf("testing string parsing for %q", s)
		d := MustNewFromString(s)
		if d.String() != s {
			t.Errorf("expected %s, got %s", s, d.String())
		}
	}
}

func TestNewFromStringError(t *testing.T) {
	tests := []string{
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

	for i, s := range tests {
		t.Logf("test %d: %q", i, s)
		_, err := NewFromString(s)
		if err == nil {
			t.Errorf("error expected when parsing %s", s)
		}

		if !didPanic(func() { MustNewFromString(s) }) {
			t.Errorf("expected panic when parsing %s", s)
		}

	}
}

func TestNewFromFloatWithExponent(t *testing.T) {
	type Inp struct {
		float float64
		exp   int32
	}
	tests := map[Inp]string{
		Inp{123.4, -3}:      "123.400",
		Inp{123.4, -1}:      "123.4",
		Inp{123.412345, 1}:  "120",
		Inp{123.412345, 0}:  "123",
		Inp{123.412345, -5}: "123.41235",
		Inp{123.412345, -6}: "123.412345",
		Inp{123.412345, -7}: "123.4123450",
	}
	for input, expected := range tests {
		t.Logf("testing NewFromFloatWithExponent: %v -> %q", input, expected)
		d := NewFromFloatWithExponent(input.float, input.exp)
		if d.String() != expected {
			t.Errorf("expected %s, got %s", expected, d.String())
		}

		// Test negative values too.
		input.float = -input.float
		expected = fmt.Sprintf("-%s", expected)
		t.Logf("testing NewFromFloatWithExponent: %v -> %q", input, expected)
		d = NewFromFloatWithExponent(input.float, input.exp)
		if d.String() != expected {
			t.Errorf("expected %s, got %s", expected, d.String())
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
		expect: "15",
	}, {
		about:  "decimal",
		num:    15,
		denom:  1000,
		expect: "0.0150000000000000",
	}, {
		about:  "irrational",
		num:    10,
		denom:  6,
		expect: "1.6666666666666667",
	}, {
		about:  "negative integer",
		num:    -15,
		denom:  1,
		expect: "-15",
	}}
	for i, test := range tests {
		t.Logf("test %d: %s", i, test.about)
		d := NewFromRat(big.NewRat(test.num, test.denom))
		if d.String() != test.expect {
			t.Errorf("expected %q, got %q", test.expect, d.String())
		}
	}
}

func TestDecimalRescale(t *testing.T) {
	type Inp struct {
		int     int64
		exp     int32
		rescale int32
	}
	type Res struct {
		Before string
		After  string
	}
	tests := map[Inp]Res{
		Inp{1234, -3, -5}: Res{"1.23400", "1.23400"},
		Inp{1234, -3, 0}:  Res{"1", "1"},
		Inp{1234, 3, 0}:   Res{"1234000", "1234000"},
		Inp{1234, -4, -4}: Res{"0.1234", "0.1234"},
		Inp{1234, 0, 0}:   Res{"1234", "1234"},
	}

	// add negatives
	for p, s := range tests {

		if p.int > 0 {
			negs := Res{
				Before: "-" + s.Before,
				After:  "-" + s.After,
			}
			tests[Inp{-p.int, p.exp, p.rescale}] = negs
		}
	}

	for input, s := range tests {
		t.Logf("testing with decimal %d %d, rescaled to %d", input.int, input.exp, input.rescale)

		d := New(input.int, input.exp).rescale(input.rescale)
		if d.String() != s.Before {
			t.Errorf("expected %s, got %s", s.Before, d.String())
		}

		// test StringScaled
		s2 := New(input.int, input.exp).StringScaled(input.rescale)
		if s2 != s.After {
			t.Errorf("expected %s, got %s", s.After, s2)
		}
	}
}

func TestDecimalRound(t *testing.T) {
	type Inp struct {
		int     int64
		exp     int32
		rescale int32
	}
	tests := map[Inp]string{
		Inp{1234, -3, -2}: "1.23",
		Inp{1245, -3, -2}: "1.25",
		Inp{1229, -3, -2}: "1.23",
		Inp{123, 0, 1}:    "120",
		Inp{12345, -1, 0}: "1235",
		Inp{12345, -1, 1}: "1230",
		// Negative cases.
		Inp{-1234, -3, -2}: "-1.23",
		Inp{-2245, -3, -2}: "-2.25",
		Inp{-3229, -3, -2}: "-3.23",
		Inp{-423, 0, 1}:    "-420",
		// Rounding up is a nop
		Inp{1234, -3, -4}: "1.234",
		Inp{1245, -3, -4}: "1.245",
		Inp{1229, -3, -4}: "1.229",
		Inp{123, 0, -1}:   "123",
	}

	for input, s := range tests {
		t.Logf("test rounding with decimal (%d, %d) rounded to %d", input.int, input.exp, input.rescale)
		d := New(input.int, input.exp).Round(input.rescale)

		if d.String() != s {
			t.Errorf("expected %s, got %s (%s, %d)",
				s, d.String(),
				d.value.String(), d.exp)
		}
	}
}

func TestUninitialized(t *testing.T) {
	a := Decimal{}
	b := Decimal{}

	decs := []Decimal{
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

func TestAdd(t *testing.T) {
	type Inp struct {
		a string
		b string
	}

	inputs := map[Inp]string{
		Inp{"2", "3"}:     "5",
		Inp{".1", ".1"}:   "0.2",
		Inp{".1", "-.1"}:  "0.0",
		Inp{"0", "1.001"}: "1.001",

		Inp{"2454495034", "3451204593"}:   "5905699627",
		Inp{"24544.95034", ".3451204593"}: "24545.2954604593",
	}

	for input, expected := range inputs {
		t.Logf("test %s + %s = %s", input.a, input.b, expected)
		a := MustNewFromString(input.a)
		b := MustNewFromString(input.b)

		c := a.Add(b)
		if c.String() != expected {
			t.Errorf("expected %s, got %s", expected, c.String())
		}
	}
}

func TestSub(t *testing.T) {
	type Inp struct {
		a string
		b string
	}

	inputs := map[Inp]string{
		Inp{"2", "3"}:     "-1",
		Inp{"12", "3"}:    "9",
		Inp{"-2", "9"}:    "-11",
		Inp{".1", "-.1"}:  "0.2",
		Inp{".1", ".1"}:   "0.0",
		Inp{"0", "1.001"}: "-1.001",
		Inp{"1.001", "0"}: "1.001",
		Inp{"2.3", ".3"}:  "2.0",

		Inp{"2454495034", "3451204593"}:   "-996709559",
		Inp{"24544.95034", ".3451204593"}: "24544.6052195407",
	}

	for input, expected := range inputs {
		t.Logf("test %s - %s = %s", input.a, input.b, expected)
		a := MustNewFromString(input.a)
		b := MustNewFromString(input.b)

		c := a.Sub(b)
		if c.String() != expected {
			t.Errorf("expected %s, got %s", expected, c.String())
		}
	}
}

func TestMul(t *testing.T) {
	type Inp struct {
		a string
		b string
	}

	inputs := map[Inp]string{
		Inp{"2", "3"}:     "6",
		Inp{".1", ".1"}:   "0.01",
		Inp{"0", "1.001"}: "0.000",

		Inp{"2454495034", "3451204593"}:   "8470964534836491162",
		Inp{"24544.95034", ".3451204593"}: "8470.964534836491162",
	}

	for input, expected := range inputs {
		t.Logf("test %s * %s = %s", input.a, input.b, expected)
		a := MustNewFromString(input.a)
		b := MustNewFromString(input.b)

		c := a.Mul(b)
		if c.String() != expected {
			t.Errorf("expected %s, got %s", expected, c.String())
		}
	}

	// positive scale
	c := New(1234, 5).Mul(New(45, -1))
	if c.String() != "555300000" {
		t.Errorf("Expected %s, got %s", "555300000", c.String())
	}
}

func TestDecimalDiv(t *testing.T) {
	type Inp struct {
		a string
		b string
	}

	tests := map[Inp]string{
		Inp{"6", "3"}:       "2",
		Inp{"0", "1"}:       "0",
		Inp{"10", "2"}:      "5",
		Inp{"2.2", "1.1"}:   "2",
		Inp{"-2.2", "-1.1"}: "2",
		Inp{"12.88", "5.6"}: "2.3",
		Inp{"1", "3"}:       "0.3333333333333333",
		Inp{"2", "3"}:       "0.6666666666666667", // rounded
		Inp{"10000", "3"}:   "3333.3333333333333333",

		Inp{"10234274355545544493", "-3"}:        "-3411424785181848164.3333333333333333",
		Inp{"-4612301402398.4753343454", "23.5"}: "-196268144782.9138440146978723",
		Inp{"1023427554493", "43432632"}:         "23563.5628642767953828", // rounded
		Inp{"1", "434324545566634"}:              "0.0000000000000023",
	}

	for input, expected := range tests {
		t.Logf("test %s / %s = %s", input.a, input.b, expected)
		num := MustNewFromString(input.a)
		denom := MustNewFromString(input.b)

		got := num.Div(denom)
		if got.String() != expected {
			t.Errorf("expected %s, got %v", expected, got)
		}
	}
}

func TestDivWithLiterals(t *testing.T) {
	// Test code path where exp > 0.
	type Inp struct {
		a    int64
		expa int32
		b    int64
		expb int32
	}
	tests := map[Inp]string{
		Inp{124, 10, 3, 1}: "41333333333.3333333333333333",
		Inp{124, 10, 3, 0}: "413333333333.3333333333333333",
		Inp{124, 10, 6, 1}: "20666666666.6666666666666667",
		Inp{124, 10, 6, 0}: "206666666666.6666666666666667",
		Inp{10, 10, 10, 1}: "1000000000",
	}

	for input, expectedAbs := range tests {
		for i := -1; i <= 1; i += 2 {
			for j := -1; j <= 1; j += 2 {
				num := New(input.a*int64(i), input.expa)
				denom := New(input.b*int64(j), input.expb)
				expected := expectedAbs
				if i != j {
					expected = "-" + expectedAbs
				}
				t.Logf("test %s / %s = %s", num.String(), denom.String(), expected)

				got := num.Div(denom)
				if got.String() != expected {
					t.Errorf("expected %s, got %v", expected, got)
				}
			}
		}
	}
}

func TestOverflow(t *testing.T) {
	if !didPanic(func() { New(1, math.MinInt32).Mul(New(1, math.MinInt32)) }) {
		t.Fatalf("should have gotten an overflow panic")
	}
	if !didPanic(func() { New(1, math.MaxInt32).Mul(New(1, math.MaxInt32)) }) {
		t.Fatalf("should have gotten an overflow panic")
	}
}

func TestExponent(t *testing.T) {
	a := New(1234, -3)
	if a.Exponent() != -3 {
		t.Errorf("expected exponent to be -3, got %d", a.Exponent())
	}
}

func TestAbs(t *testing.T) {
	a := New(-1234, -4)
	b := New(1234, -4)

	c := a.Abs()
	if c.Cmp(b) != 0 {
		t.Errorf("error")
	}
}

func TestCmp(t *testing.T) {
	tests := []struct {
		a        Decimal
		b        Decimal
		expected int
	}{{
		a:        New(123, 3),
		b:        New(-1234, 2),
		expected: 1,
	}, {
		a:        New(123, 3),
		b:        New(1234, 2),
		expected: -1,
	}, {
		a:        New(1230, 2),
		b:        New(123, 3),
		expected: 0,
	}}

	for i, test := range tests {
		t.Logf("test %d: %s vs %s", i, test.a.String(), test.b.String())
		r := test.a.Cmp(test.b)
		if r != test.expected {
			t.Errorf("comparing %s and %s expected %d, got %d", test.a, test.b, test.expected, r)
		}
	}
}

// didPanic checks if the provided function panics.
func didPanic(f func()) (ret bool) {
	defer func() {
		if message := recover(); message != nil {
			ret = true
		}
	}()

	// call the target function
	f()

	return false
}
