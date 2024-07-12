# decimal

Arbitrary-precision fixed-point decimal numbers in Go.

This library supports numbers with a maximum of 2^31 digits
after the decimal point.

## Features

 * the zero-value is 0, and is safe to use without initialization
 * addition, subtraction, multiplication with no loss of precision
 * division with specified precision
 * database/sql serialization/deserialization
 * JSON and XML serialization/deserialization

## Install

Run `go get github.com/canonical/decimal`

## Usage

```go
package main

import (
	"fmt"

	"github.com/canonical/decimal"
)

func main() {
	price, err := decimal.NewFromString("136.02")
	if err != nil {
		panic(err)
	}

	quantity := decimal.MustNewFromFloat(3)

	fee := decimal.MustNewFromString(".035")
	taxRate := decimal.MustNewFromString(".08875")

	subtotal := price.Mul(quantity)

	preTax := subtotal.Mul(fee.Add(decimal.MustNewFromFloat(1)))

	total := preTax.Mul(taxRate.Add(decimal.MustNewFromFloat(1)))

	fmt.Println("Subtotal:", subtotal)
	fmt.Println("Pre-tax:", preTax)
	fmt.Println("Taxes:", total.Sub(preTax))
	fmt.Println("Total:", total)
	fmt.Println("Tax rate:", total.Sub(preTax).Div(preTax))
}
```

## License

This project is a derivative of works released under the MIT license.
All other rights are reserved. Copyright 2015-2024 Canonical Ltd.
