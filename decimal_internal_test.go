package decimal

import (
	"testing"
)

var testCtx = NewContext(6)

func TestDecimalRescale(t *testing.T) {

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
		t.Logf("test %d", i)

		d := testCtx.New(test.value, test.exp).rescale(test.rescale)
		if d.String() != test.before {
			t.Errorf("expected %s, got %s", test.before, d.String())
		}

		// test StringScaled
		s2 := testCtx.New(test.value, test.exp).StringScaled(test.rescale)
		if s2 != test.after {
			t.Errorf("expected %s, got %s", test.after, s2)
		}
	}
}
