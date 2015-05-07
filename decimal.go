// Package decimal implements an arbitrary precision fixed-point decimal.
//
// To use as part of a struct:
//
//     type Struct struct {
//         Number Decimal
//     }
//
// The zero-value of a Decimal is 0, as you would expect.
//
// The best way to create a new Decimal is to use decimal.NewFromString, ex:
//
//     n, err := decimal.NewFromString("-123.4567")
//     n.String() // output: "-123.4567"
//
// NOTE: This can "only" represent numbers with a maximum of 2^31 digits
// after the decimal point.
package decimal

import (
	"database/sql/driver"
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

const defaultPrecision = 16

var defaultContext = NewContext(defaultPrecision)

// Monetary is a decimal context for financial calculations,
// maintaining precision to two decimal places.
var Monetary = NewContext(2)

// Zero constant, to make computations faster.
var Zero = New(0, 1)

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

// New returns a new fixed-point decimal, value * 10 ^ exp.
func (c Context) New(value int64, exp int32) Decimal {
	return c.Adapt(Decimal{
		value: big.NewInt(value),
		exp:   exp,
	})
}

// New returns a new fixed-point decimal, value * 10 ^ exp. Using the default context.
func New(value int64, exp int32) Decimal {
	return defaultContext.New(value, exp)
}

// NewFromString returns a new Decimal from a string representation.
//
// Example:
//     d, err := NewFromString("-123.45")
//     d2, err := NewFromString(".0001")
func (c Context) NewFromString(value string) (Decimal, error) {
	var intString string
	var exp int32
	parts := strings.Split(value, ".")
	if len(parts) == 1 {
		// There is no decimal point, we can just parse the original string as
		// an int
		intString = value
		exp = 0
	} else if len(parts) == 2 {
		intString = parts[0] + parts[1]
		expInt := -len(parts[1])
		if expInt < math.MinInt32 {
			// NOTE(vadim): I doubt a string could realistically be this long
			return Decimal{}, fmt.Errorf("can't convert %s to decimal: fractional part too long", value)
		}
		exp = int32(expInt)
	} else {
		return Decimal{}, fmt.Errorf("can't convert %s to decimal: too many .s", value)
	}

	dValue := new(big.Int)
	_, ok := dValue.SetString(intString, 10)
	if !ok {
		return Decimal{}, fmt.Errorf("can't convert %s to decimal", value)
	}

	return c.Adapt(Decimal{
		Context: c,
		value:   dValue,
		exp:     exp,
	}), nil
}

// NewFromString parses a string representation into a decimal using the default context.
func NewFromString(value string) (Decimal, error) {
	return defaultContext.NewFromString(value)
}

// MustNewFromString parses the provided string representation into a Decimal or panics if an error occurs.
//
// Example:
//     d := MustNewFromString("-123.45")
//     d2 := MustNewFromString(".0001")
func (c Context) MustNewFromString(value string) Decimal {
	d, err := c.NewFromString(value)
	if err != nil {
		panic(err)
	}
	return d
}

// MustNewFromString parses the string representation into a decimal using the default representation or panics.
func MustNewFromString(value string) Decimal {
	return defaultContext.MustNewFromString(value)
}

// NewFromRat returns a new Decimal from a big.Rat.
//
// Example:
//    d := NewFromRat(big.NewRat(10, 5))
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

// NewFromRate returns a new decimal from a big.Rat using the default context.
func NewFromRat(value *big.Rat) Decimal {
	return defaultContext.NewFromRat(value)
}

// NewFromFloat converts a float64 to Decimal.
// NOTE: this will return an error on NaN, +/-inf.
//
// Example:
//     NewFromFloat(123.45678901234567).String() // output: "123.4567890123456"
//     NewFromFloat(.00000000000000001).String() // output: "0.00000000000000001"
func (c Context) NewFromFloat(value float64) (Decimal, error) {
	floor := math.Floor(value)

	// fast path, where float is an int
	if floor == value && !math.IsInf(value, 0) {
		return New(int64(value), 0), nil
	}

	// slow path: float is a decimal
	// HACK(vadim): do this the slow hacky way for now because the logic to
	// convert a base-2 float to base-10 properly is not trivial
	str := strconv.FormatFloat(value, 'f', -1, 64)
	return c.NewFromString(str)
}

// NewFromFloat converts a float64 to Decimal using the default context.
func NewFromFloat(value float64) (Decimal, error) {
	return defaultContext.NewFromFloat(value)
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

// MustNewFromFloat creates a decimal number equivalent to the provided float or
// panics. It uses the default context.
func MustNewFromFloat(value float64) Decimal {
	return defaultContext.MustNewFromFloat(value)
}

// rescale returns a rescaled version of the decimal. Returned
// decimal may be less precise if the given exponent is bigger
// than the initial exponent of the Decimal.
// NOTE: this will truncate, NOT round
//
// Example:
// 	d := New(12345, -4)
//	d2 := d.rescale(-1)
//	d3 := d2.rescale(-4)
//	println(d1)
//	println(d2)
//	println(d3)
//
// Output:
//	1.2345
//	1.2
//	1.2000
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
		value: value,
		exp:   exp,
	}
}

// Round returns a rounded decimal number.
//
// Rounidng can only be performed with an exponent larger than the
// exponent of the number. If the exponent is smaller, the number is
// returned unchanged.
//
// Example:
//     d := New(1234545, -3)
//     d.Round(-4) returns 1234.545
//     d.Round(-2) returns 1234.55
//     d.Round(-1) returns 1234.5
//     d.Round(0) returns 1235
//     d.Round(1) returns 1230
//
//     d := New(-1235, -3)
//     d.Round(-2) returns 1.24
func (d Decimal) Round(exp int32) Decimal {
	if exp == d.exp {
		return d
	} else if exp < d.exp { // nothing to round
		return d
	}
	d.ensureInitialized()
	rounding := Decimal{
		value: big.NewInt(int64(d.value.Sign()) * 5),
		exp:   exp - 1}
	rounded := d.Add(rounding)
	return rounded.rescale(exp)
}

// Abs returns the absolute value of the decimal.
func (d Decimal) Abs() Decimal {
	d.ensureInitialized()
	d2Value := new(big.Int).Abs(d.value)
	return Decimal{
		value: d2Value,
		exp:   d.exp,
	}
}

// Add returns d + d2.
func (d Decimal) Add(d2 Decimal) Decimal {
	baseScale := min(d.exp, d2.exp)
	rd := d.rescale(baseScale)
	rd2 := d2.rescale(baseScale)

	d3Value := new(big.Int).Add(rd.value, rd2.value)
	return Decimal{
		value: d3Value,
		exp:   baseScale,
	}
}

// Sub returns d - d2.
func (d Decimal) Sub(d2 Decimal) Decimal {
	baseScale := min(d.exp, d2.exp)
	rd := d.rescale(baseScale)
	rd2 := d2.rescale(baseScale)

	d3Value := new(big.Int).Sub(rd.value, rd2.value)
	return Decimal{
		value: d3Value,
		exp:   baseScale,
	}
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
	return Decimal{
		value: d3Value,
		exp:   int32(expInt64),
	}
}

// Div returns a / b. If it doesn't divide exactly, the result will have
// DivisionPrecision digits after the decimal point.
func (a Decimal) Div(b Decimal) Decimal {
	if a.value.Cmp(zeroInt) == 0 {
		return Zero
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

	// Remove insignificant decimal places.
	// TODO: this will be simplified once decimal context is introduced.
	m := big.NewInt(0)
	temp := big.NewInt(0).Set(num)
	var expShift int32
	for m.Cmp(zeroInt) == 0 {
		expShift++
		temp.DivMod(temp, tenInt, m)
	}
	expShift--
	expScale := big.NewInt(0).Exp(tenInt, big.NewInt(int64(expShift)), nil)
	num.Div(num, expScale)

	return Decimal{
		value: num,
		exp:   -(precision - expShift),
	}
}

// Cmp compares the numbers represented by d and d2 and returns:
//     -1 if d <  d2
//      0 if d == d2
//     +1 if d >  d2
func (d Decimal) Cmp(d2 Decimal) int {
	baseExp := min(d.exp, d2.exp)
	rd := d.rescale(baseExp)
	rd2 := d2.rescale(baseExp)

	return rd.value.Cmp(rd2.value)
}

// Equals returns whether the numbers represented by d and d2 are equal.
func (d Decimal) Equals(d2 Decimal) bool {
	return d.Cmp(d2) == 0
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

// String returns the string representation of the decimal
// with the fixed point.
//
// Example:
//     d := New(-12345, -3)
//     println(d.String())
//
// Output:
//     -12.345
func (d Decimal) String() string {
	if d.exp >= 0 {
		return d.rescale(0).value.String()
	}

	abs := new(big.Int).Abs(d.value)
	str := abs.String()

	var intPart, fractionalPart string
	dExpInt := int(d.exp)
	if len(str) > -dExpInt {
		intPart = str[:len(str)+dExpInt]
		fractionalPart = str[len(str)+dExpInt:]
	} else {
		intPart = "0"

		num0s := -dExpInt - len(str)
		fractionalPart = strings.Repeat("0", num0s) + str
	}

	number := intPart
	if len(fractionalPart) > 0 {
		number += "." + fractionalPart
	}

	if d.value.Sign() < 0 {
		return "-" + number
	}
	return number
}

// StringScaled first scales the decimal then calls .String() on it.
func (d Decimal) StringScaled(exp int32) string {
	return d.rescale(exp).String()
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (d *Decimal) UnmarshalJSON(decimalBytes []byte) error {
	str, err := unquoteIfQuoted(decimalBytes)
	if err != nil {
		return fmt.Errorf("Error decoding string '%s': %s", decimalBytes, err)
	}

	decimal, err := NewFromString(str)
	*d = decimal
	if err != nil {
		return fmt.Errorf("Error decoding string '%s': %s", str, err)
	}
	return nil
}

// MarshalJSON implements the json.Marshaler interface.
func (d Decimal) MarshalJSON() ([]byte, error) {
	str := "\"" + d.String() + "\""
	return []byte(str), nil
}

// Truncate truncates off digits from the number, without rounding.
// NOTE: precision is the last digit that will not be truncated (should be >= 0)
//
// Example:
//     decimal.NewFromString("123.456").Truncate(2).String() // "123.45"
func (d Decimal) Truncate(precision int32) Decimal {
	d.ensureInitialized()
	if precision >= 0 && -precision > d.exp {
		return d.rescale(-precision)
	}
	return d
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

// Value implements the driver.Valuer interface for database serialization.
func (d Decimal) Value() (driver.Value, error) {
	return d.String(), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface for XML
// deserialization.
func (d *Decimal) UnmarshalText(text []byte) error {
	str := string(text)

	dec, err := NewFromString(str)
	*d = dec
	if err != nil {
		return fmt.Errorf("Error decoding string '%s': %s", str, err)
	}

	return nil
}

// MarshalText implements the encoding.TextMarshaler interface for XML
// serialization.
func (d Decimal) MarshalText() (text []byte, err error) {
	return []byte(d.String()), nil
}

func (d *Decimal) ensureInitialized() {
	if d.value == nil {
		d.value = new(big.Int)
	}
}

func min(x, y int32) int32 {
	if x >= y {
		return y
	}
	return x
}

func round(n float64) int64 {
	if n < 0 {
		return int64(n - 0.5)
	}
	return int64(n + 0.5)
}

func unquoteIfQuoted(value interface{}) (string, error) {
	bytes, ok := value.([]byte)
	if !ok {
		return "", fmt.Errorf("Could not convert value '%+v' to byte array",
			value)
	}

	// If the amount is quoted, strip the quotes
	if len(bytes) > 2 && bytes[0] == '"' && bytes[len(bytes)-1] == '"' {
		bytes = bytes[1 : len(bytes)-1]
	}
	return string(bytes), nil
}
