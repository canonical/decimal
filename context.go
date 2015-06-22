package decimal

import (
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"
)

// Context stores the explicitly set precision for decimal number calculations.
//
// Each Decimal created using a context retains that context's information:
//
//     c := NewContext(4) // context with precision to four decimal places.
//     d := c.New(15, -1) // 1.5
//     d = d.Mul(d)
//
type Context struct {
	precision int32
}

// Precision returns the precision of the context.
func (c Context) Precision() int32 {
	return c.precision
}

// NewContext creates a new context with the specified precision.
func NewContext(precision int32) Context {
	return Context{
		precision: precision,
	}
}

// Adapt adapts the provided decimal number to the context.
// If the decimal number has a higher accuracy than that of the context, it will
// be rounded to the accuracy of the context.
func (c Context) Adapt(d Decimal) Decimal {
	r := d
	if d.exp < -c.precision {
		r = d.Round(-c.precision)
	}
	r.Context = c
	return r
}

const defaultPrecision = 16

// DefaultContext is the decimal context with the default precision of 16 decimal places.
var DefaultContext = NewContext(defaultPrecision)

// Monetary is a decimal context for financial calculations,
// maintaining precision to two decimal places.
var MonetaryContext = NewContext(2)

// Zero returns the zero decimal value using the specified context.
func (c Context) Zero() Decimal {
	return c.New(0, 0)
}

// New returns a new fixed-point decimal, value * 10 ^ exp.
func (c Context) New(value int64, exp int32) Decimal {
	return c.Adapt(Decimal{
		value: big.NewInt(value),
		exp:   exp,
	})
}

// NewFromString returns a new Decimal from a string representation.
func (c Context) NewFromString(value string) (Decimal, error) {
	var intString string
	var exp int32
	parts := strings.Split(value, ".")
	switch len(parts) {
	case 1:
		// There is no decimal point, we can just parse the original string as
		// an int.
		intString = value
		exp = 0
	case 2:
		intString = parts[0] + parts[1]
		expInt := -len(parts[1])
		if expInt < math.MinInt32 {
			// NOTE(vadim): I doubt a string could realistically be this long
			return Decimal{}, fmt.Errorf("can't convert %s to decimal: fractional part too long", value)
		}
		exp = int32(expInt)
	default:
		return Decimal{}, fmt.Errorf("can't convert %s to decimal: too many .s", value)
	}

	dValue := new(big.Int)
	_, ok := dValue.SetString(intString, 10)
	if !ok {
		return Decimal{}, fmt.Errorf("can't convert %s to decimal", value)
	}

	return c.Adapt(Decimal{
		value: dValue,
		exp:   exp,
	}), nil
}

// MustNewFromString parses the provided string representation into a Decimal or panics if an error occurs.
func (c Context) MustNewFromString(value string) Decimal {
	d, err := c.NewFromString(value)
	if err != nil {
		panic(err)
	}
	return d
}

// NewFromRat returns a new Decimal from a big.Rat.
func (c Context) NewFromRat(value *big.Rat) Decimal {
	var str string
	if value.IsInt() {
		str = value.Num().String()
	} else {
		str = value.FloatString(int(c.precision))
	}
	d, err := c.NewFromString(str)
	if err != nil {
		// Panic here, because we just failed to parse a string
		// we constructed.
		panic(err)
	}
	return d
}

// NewFromFloat converts a float64 to Decimal.
// NOTE: this will return an error on NaN, +/-inf.
func (c Context) NewFromFloat(value float64) (Decimal, error) {
	floor := math.Floor(value)

	// fast path, where float is an int
	if floor == value && !math.IsInf(value, 0) {
		return c.New(int64(value), 0), nil
	}

	// slow path: float is a decimal
	// HACK(vadim): do this the slow hacky way for now because the logic to
	// convert a base-2 float to base-10 properly is not trivial
	str := strconv.FormatFloat(value, 'f', -1, 64)
	return c.NewFromString(str)
}

// MustNewFromFloat creates a decimal number equivalent to the provided float.
// It will panic on error.
func (c Context) MustNewFromFloat(value float64) Decimal {
	d, err := c.NewFromFloat(value)
	if err != nil {
		panic(err)
	}
	return d
}
