// Package decimal implements an arbitrary precision fixed-point decimal.
//
// To use as part of a struct:
//
//	type Struct struct {
//	    Number Decimal
//	}
//
// The zero-value of a Decimal is 0, as you would expect.
//
// The best way to create a new Decimal is to use decimal.NewFromString, ex:
//
//	n, err := decimal.NewFromString("-123.4567")
//	n.String() // output: "-123.4567"
//
// NOTE: This can "only" represent numbers with a maximum of 2^31 digits
// after the decimal point.
package decimal

import (
	"database/sql/driver"
	"fmt"
	"math"
	"math/big"
	"strings"
)

// Zero constant.
var Zero = DefaultContext.New(0, 1)

var (
	zeroInt = big.NewInt(0)
	tenInt  = big.NewInt(10)
	oneInt  = big.NewInt(1)
)

// Decimal represents a fixed-point decimal. It is immutable.
// number = value * 10 ^ exp
type Decimal struct {
	Context

	value *big.Int

	// NOTE(vadim): this must be an int32, because we cast it to float64 during
	// calculations. If exp is 64 bit, we might lose precision.
	// If we cared about being able to represent every possible decimal, we
	// could make exp a *big.Int but it would hurt performance and numbers
	// like that are unrealistic.
	exp int32
}

// Round returns a rounded decimal number.
//
// Rounding can only be performed with an exponent larger than the
// exponent of the number. If the exponent is smaller, the number is
// returned unchanged.
func (d Decimal) Round(exp int32) Decimal {
	if exp <= d.exp {
		return d
	}
	d.ensureInitialized()

	expShift := exp - d.exp - 1
	rounding := new(big.Int).Exp(tenInt, big.NewInt(int64(expShift)), nil)
	rounding.Mul(rounding, big.NewInt(int64(d.value.Sign()*5)))

	rounded := d
	rounded.value.Add(rounded.value, rounding)

	return rounded.rescale(exp)
}

// String returns the string representation of the decimal with the fixed point.
func (d Decimal) String() string {
	return d.string(true)
}

// StringFixed returns the string representation of the decimal with the fixed
// point and a fixed number of digits after the decimal point, as instructed by
// the precision.
func (d Decimal) StringFixed() string {
	return d.string(false)
}

func (d Decimal) string(trim bool) string {
	d.ensureInitialized()

	var intPart, fractionalPart string

	if d.exp >= 0 {
		intPart = d.rescale(0).value.String()
		if d.precision > 0 {
			fractionalPart = strings.Repeat("0", int(d.precision))
		}
	} else {
		abs := new(big.Int).Abs(d.value)
		str := abs.String()

		dExpInt := int(d.exp)
		if len(str) > -dExpInt {
			intPart = str[:len(str)+dExpInt]
			fractionalPart = str[len(str)+dExpInt:]
		} else {
			intPart = "0"

			num0s := -dExpInt - len(str)
			fractionalPart = strings.Repeat("0", num0s) + str
		}

		if len(fractionalPart) < int(d.precision) {
			fractionalPart = fractionalPart + strings.Repeat("0", int(d.precision)-len(fractionalPart))
		}
		if d.value.Sign() < 0 {
			intPart = "-" + intPart
		}
	}

	if trim {
		i := len(fractionalPart) - 1
		for ; i >= 0 && fractionalPart[i] == '0'; i-- {
			fractionalPart = fractionalPart[:i]
		}
	}

	number := intPart
	if len(fractionalPart) > 0 {
		number += "." + fractionalPart
	}

	return number
}

// rescale returns a rescaled version of the decimal. Returned
// decimal may be less precise if the given exponent is bigger
// than the initial exponent of the Decimal.
// NOTE: this will truncate, NOT round
func (d Decimal) rescale(exp int32) Decimal {
	d.ensureInitialized()
	// NOTE(vadim): must convert exps to float64 before - to prevent overflow
	diff := math.Abs(float64(exp) - float64(d.exp))
	value := new(big.Int).Set(d.value)

	expScale := new(big.Int).Exp(tenInt, big.NewInt(int64(diff)), nil)
	if exp > d.exp {
		value = value.Quo(value, expScale)
	} else if exp < d.exp {
		value = value.Mul(value, expScale)
	}

	return Decimal{
		Context: d.Context,
		value:   value,
		exp:     exp,
	}
}

func (d *Decimal) ensureInitialized() {
	if d.value == nil {
		d.value = new(big.Int)
	}
}

// Add returns d + d2.
func (d Decimal) Add(d2 Decimal) Decimal {
	baseScale := min(d.exp, d2.exp)
	rd := d.rescale(baseScale)
	rd2 := d2.rescale(baseScale)

	d3Value := new(big.Int).Add(rd.value, rd2.value)
	return d.Context.Adapt(Decimal{
		value: d3Value,
		exp:   baseScale,
	})
}

// Abs returns the absolute value of the decimal.
func (d Decimal) Abs() Decimal {
	d.ensureInitialized()
	d2Value := new(big.Int).Abs(d.value)
	return Decimal{
		Context: d.Context,
		value:   d2Value,
		exp:     d.exp,
	}
}

// Sub returns d - d2.
func (d Decimal) Sub(d2 Decimal) Decimal {
	baseScale := min(d.exp, d2.exp)
	rd := d.rescale(baseScale)
	rd2 := d2.rescale(baseScale)

	d3Value := new(big.Int).Sub(rd.value, rd2.value)
	return d.Context.Adapt(Decimal{
		value: d3Value,
		exp:   baseScale,
	})
}

// Mul returns d * d2.
func (d Decimal) Mul(d2 Decimal) Decimal {
	d.ensureInitialized()
	d2.ensureInitialized()

	expInt64 := int64(d.exp) + int64(d2.exp)
	if expInt64 > math.MaxInt32 || expInt64 < math.MinInt32 {
		// NOTE(vadim): better to panic than give incorrect results, as
		// Decimals are usually used for money
		panic(fmt.Sprintf("exponent %v overflows an int32!", expInt64))
	}

	d3Value := new(big.Int).Mul(d.value, d2.value)
	return d.Context.Adapt(Decimal{
		value: d3Value,
		exp:   int32(expInt64),
	})
}

// Div returns a / b. If it doesn't divide exactly, the result will have
// DivisionPrecision digits after the decimal point.
func (a Decimal) Div(b Decimal) Decimal {
	if a.value.Cmp(zeroInt) == 0 {
		return a.Context.Zero()
	}

	precision := int32(a.precision)

	// Working representation of the numerator.
	num := big.NewInt(0).Set(a.value)

	// Power of 10 needed for division with required precision.
	pwrFloat := float64(precision) + 1 + float64(a.exp) - float64(b.exp)
	if math.IsNaN(pwrFloat) || math.IsInf(pwrFloat, 0) {
		panic(fmt.Sprintf("can't convert %v to decimal", pwrFloat))
	}
	pwr := round(pwrFloat)
	exp := big.NewInt(0).Exp(tenInt, big.NewInt(int64(pwr)), nil)

	num.Mul(num, exp)

	// Divide.
	num = num.Div(num, b.value)

	// Check remainder.
	rem := big.NewInt(0)
	num, rem = num.DivMod(num, tenInt, rem)

	// Round up.
	if rem.Int64() >= 5 {
		num.Add(num, big.NewInt(1))
	}

	return a.Context.Adapt(Decimal{
		Context: a.Context,
		value:   num,
		exp:     -precision,
	})
}

// Cmp compares the numbers represented by d and d2 and returns:
//
//	-1 if d <  d2
//	 0 if d == d2
//	+1 if d >  d2
func (d Decimal) Cmp(d2 Decimal) int {
	baseExp := min(d.exp, d2.exp)
	rd := d.rescale(baseExp)
	rd2 := d2.rescale(baseExp)

	return rd.value.Cmp(rd2.value)
}

// Equal returns whether the number represented by d and d2 are equal.
func (d Decimal) Equal(d2 Decimal) bool {
	return d.Cmp(d2) == 0
}

// Equals is an alias for Equal.
func (d Decimal) Equals(d2 Decimal) bool {
	return d.Equal(d2)
}

// Exponent returns the exponent, or scale component of the decimal.
func (d Decimal) Exponent() int32 {
	return d.exp
}

// IntPart returns the integer component of the decimal.
func (d Decimal) IntPart() int64 {
	scaledD := d.rescale(0)
	return scaledD.value.Int64()
}

// Rat returns a rational number representation of the decimal.
func (d Decimal) Rat() *big.Rat {
	d.ensureInitialized()
	if d.exp <= 0 {
		denom := new(big.Int).Exp(tenInt, big.NewInt(int64(-d.exp)), nil)
		return new(big.Rat).SetFrac(d.value, denom)
	} else {
		mul := new(big.Int).Exp(tenInt, big.NewInt(int64(d.exp)), nil)
		num := new(big.Int).Mul(d.value, mul)
		return new(big.Rat).SetFrac(num, oneInt)
	}
}

// Float64 returns the nearest float64 value for d and a bool indicating
// whether f represents d exactly.
// For more details, see the documentation for big.Rat.Float64
func (d Decimal) Float64() (f float64, exact bool) {
	return d.Rat().Float64()
}

// StringScaled first scales the decimal then calls .StringFixed() on it.
func (d Decimal) StringScaled(exp int32) string {
	return d.rescale(exp).StringFixed()
}

// Truncate truncates off digits from the number, without rounding.
// NOTE: precision is the last digit that will not be truncated (should be >= 0)
func (d Decimal) Truncate(precision int32) Decimal {
	d.ensureInitialized()
	if precision >= 0 && -precision > d.exp {
		return d.rescale(-precision)
	}
	return d
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (d *Decimal) UnmarshalJSON(decimalBytes []byte) error {
	str, err := unquoteIfQuoted(decimalBytes)
	if err != nil {
		return fmt.Errorf("error decoding string '%s': %s", decimalBytes, err)
	}

	decimal, err := d.Context.NewFromString(str)
	*d = decimal
	if err != nil {
		return fmt.Errorf("error decoding string '%s': %s", str, err)
	}
	return nil
}

// MarshalJSON implements the json.Marshaler interface.
func (d Decimal) MarshalJSON() ([]byte, error) {
	str := "\"" + d.String() + "\""
	return []byte(str), nil
}

// Scan implements the sql.Scanner interface for database deserialization.
func (d *Decimal) Scan(value interface{}) error {
	str, err := unquoteIfQuoted(value)
	if err != nil {
		return err
	}
	*d, err = NewFromString(str)

	return err
}

// UnmarshalText implements the encoding.TextUnmarshaler interface for XML
// deserialization.
func (d *Decimal) UnmarshalText(text []byte) error {
	str := string(text)

	dec, err := d.Context.NewFromString(str)
	*d = dec
	if err != nil {
		return fmt.Errorf("error decoding string '%s': %s", str, err)
	}

	return nil
}

// MarshalText implements the encoding.TextMarshaler interface for XML
// serialization.
func (d Decimal) MarshalText() (text []byte, err error) {
	return []byte(d.String()), nil
}

// Value implements the driver.Valuer interface for database serialization.
func (d Decimal) Value() (driver.Value, error) {
	return d.String(), nil
}

// New returns a new fixed-point decimal, value * 10 ^ exp. Using the default context.
func New(value int64, exp int32) Decimal {
	return DefaultContext.New(value, exp)
}

// NewFromString parses a string representation into a decimal using the default context.
func NewFromString(value string) (Decimal, error) {
	return DefaultContext.NewFromString(value)
}

// MustNewFromString parses the string representation into a decimal using the default representation or panics.
func MustNewFromString(value string) Decimal {
	return DefaultContext.MustNewFromString(value)
}

// NewFromRat returns a new decimal from a big.Rat using the default context.
func NewFromRat(value *big.Rat) Decimal {
	return DefaultContext.NewFromRat(value)
}

// NewFromInt converts an int64 to Decimal using the default context.
func NewFromInt(value int64) Decimal {
	return DefaultContext.New(value, 0)
}

// NewFromFloat converts a float64 to Decimal using the default context.
func NewFromFloat(value float64) (Decimal, error) {
	return DefaultContext.NewFromFloat(value)
}

// MustNewFromFloat creates a decimal number equivalent to the provided float or
// panics. It uses the default context.
func MustNewFromFloat(value float64) Decimal {
	return DefaultContext.MustNewFromFloat(value)
}

func round(n float64) int64 {
	if n < 0 {
		return int64(n - 0.5)
	}
	return int64(n + 0.5)
}

func min(x, y int32) int32 {
	if x >= y {
		return y
	}
	return x
}

func unquoteIfQuoted(value interface{}) (string, error) {
	bytes, ok := value.([]byte)
	if !ok {
		return "", fmt.Errorf("could not convert value '%+v' to byte array",
			value)
	}

	// If the amount is quoted, strip the quotes
	if len(bytes) > 2 && bytes[0] == '"' && bytes[len(bytes)-1] == '"' {
		bytes = bytes[1 : len(bytes)-1]
	}
	return string(bytes), nil
}
