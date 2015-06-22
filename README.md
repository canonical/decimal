# decimal

Arbitrary-precision fixed-point decimal numbers in go.

NOTE: can "only" represent numbers with a maximum of 2^31 digits after the decmial point.

## Features

 * the zero-value is 0, and is safe to use without initialization
 * addition, subtraction, multiplication with no loss of precision
 * division with specified precision
 * database/sql serialization/deserialization
 * json and xml serialization/deserialization

## Install

Run `go get github.com/CanonicalLtd/decimal`

## Usage

```go
package main

import (
    "fmt"
    "github.com/CanonicalLtd/decimal"
)

func main() {
	price, err := decimal.NewFromString("136.02")
    if err != nil {
        panic(err)
    }

	quantity := decimal.NewFromFloat(3)

	fee, _ := decimal.NewFromString(".035")
	taxRate, _ := decimal.NewFromString(".08875")

    subtotal := price.Mul(quantity)

    preTax := subtotal.Mul(fee.Add(decimal.NewFromFloat(1)))

    total := preTax.Mul(taxRate.Add(decimal.NewFromFloat(1)))

	fmt.Println("Subtotal:", subtotal)
	fmt.Println("Pre-tax:", preTax)
    fmt.Println("Taxes:", total.Sub(preTax))
	fmt.Println("Total:", total)
	fmt.Println("Tax rate:", total.Sub(preTax).Div(preTax))
}
```

## License

The MIT License (MIT)

This is a heavily modified fork of [fpd.Decimal](https://github.com/oguzbilgic/fpd), which was also released under the MIT License.
This is a heavily modified fork of [shopspring Decimal](https://github.com/shopspring/decimal), which was also released under the MIT License.
