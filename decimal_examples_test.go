package decimal_test

import (
	"fmt"

	"github.com/CanonicalLtd/decimal"
)

func ExampleNewContext() {
	context := decimal.NewContext(6)
	fmt.Printf("%d\n", context.Precision())

	d := context.New(6, -5)
	fmt.Printf("%s\n", d.String())
}
