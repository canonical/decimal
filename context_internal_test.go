package decimal

import (
	"testing"
)

func TestNewContext(t *testing.T) {
	var precision int32 = 10

	context := NewContext(precision)
	if context.precision != precision {
		t.Fatalf("context precision (%d) not equal to %d", context.precision, precision)
	}
}
