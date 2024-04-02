package decimal_test

import (
	"fmt"
	"math/big"

	"github.com/canonical/decimal"
)

func ExampleNewContext() {
	context := decimal.NewContext(6)
	fmt.Printf("%d\n", context.Precision())
	d := context.New(6, -5)
	fmt.Println(d.String())
	// Output:
	// 6
	// 0.000060
}

func ExampleDecimal_Round() {
	d := decimal.New(1234545, -3)
	fmt.Println(d.Round(-4).String())
	fmt.Println(d.Round(-2).String())
	fmt.Println(d.Round(-1).String())
	fmt.Println(d.Round(0).String())
	fmt.Println(d.Round(1).String())
	// Output:
	// 1234.5450000000000000
	// 1234.5500000000000000
	// 1234.6000000000000000
	// 1235.0000000000000000
	// 1240.0000000000000000
}

func ExampleDecimal_String() {
	d := decimal.New(-12345, -3)
	fmt.Println(d.String())
	// Output: -12.3450000000000000
}

func ExampleDecimal_Truncate() {
	d := decimal.MustNewFromString("123.456").Truncate(2)
	fmt.Println(d.String())
	// Output: 123.4500000000000000
}

func ExampleContext_NewFromString() {
	d, err := decimal.NewFromString("-123.45")
	if err != nil {
		panic(err)
	}
	fmt.Println(d.String())
	d, err = decimal.NewFromString(".0001")
	if err != nil {
		panic(err)
	}
	fmt.Println(d.String())
	// Output:
	// -123.4500000000000000
	// 0.0001000000000000
}

func ExampleContext_NewFromRat() {
	d := decimal.NewFromRat(big.NewRat(10, 5))
	fmt.Println(d)
	// Output:
	// 2.0000000000000000
}

func ExampleContext_NewFromFloat() {
	d, err := decimal.NewFromFloat(123.45678901234567)
	if err != nil {
		panic(err)
	}
	fmt.Println(d.String())
	d, err = decimal.NewFromFloat(.0000000000000001)
	if err != nil {
		panic(err)
	}
	fmt.Println(d.String())
	// Output:
	// 123.4567890123456700
	// 0.0000000000000001
}
