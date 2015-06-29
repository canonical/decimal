// This file uses tests provided by Mike Cowlishaw's decimal arithmetic test suite.
// Cowlishaw's tests can be downloaded from:
//
//    http://speleotrove.com/decimal/dectest.zip
//

package decimal_test

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/CanonicalLtd/decimal"
)

var DirectiveRe = regexp.MustCompile(`(\w+):\s+(\d+)`)

// TODO (mattyw) Need to support conditions
var TestRe = regexp.MustCompile(`(\w+)\s+(\w+)\s+(\S+)\s+(\S+)\s+->\s+(\S+)\S*\s*(.*)`)

// A decimal context for testing with a sane number of decimal places.
//var testCtx = decimal.NewContext(6)

type Test struct {
	Context        *decimal.Context
	Name           string
	Operation      string
	OpperandA      string
	OpperandB      string
	ExpectedResult string
}

func (test *Test) Validate() error {
	if _, err := test.Context.NewFromString(test.OpperandA); err != nil {
		return fmt.Errorf("failed to parse string %v: %v", test.OpperandA, err)
	}
	if _, err := test.Context.NewFromString(test.OpperandB); err != nil {
		return fmt.Errorf("failed to parse string %v: %v", test.OpperandB, err)
	}
	if _, err := test.Context.NewFromString(test.ExpectedResult); err != nil {
		return fmt.Errorf("failed to parse string %v: %v", test.ExpectedResult, err)
	}
	return nil
}
func (test *Test) Run() error {
	a, err := test.Context.NewFromString(test.OpperandA)
	if err != nil {
		return fmt.Errorf("failed to parse string %v: %v", test.OpperandA, err)
	}
	b, err := test.Context.NewFromString(test.OpperandB)
	if err != nil {
		return fmt.Errorf("failed to parse string %v: %v", test.OpperandB, err)
	}
	expected, err := test.Context.NewFromString(test.ExpectedResult)
	if err != nil {
		return fmt.Errorf("failed to parse string %v: %v", test.ExpectedResult, err)
	}
	var result decimal.Decimal
	if test.Operation == "add" {
		result = a.Add(b)
	} else if test.Operation == "subtract" || test.Operation == "minus" {
		result = a.Sub(b)
	} else if test.Operation == "multiply" {
		result = a.Mul(b)
	} else {
		return fmt.Errorf("operation (%v) not supported", test.Operation)
	}
	if !result.Equals(expected) {
		return fmt.Errorf("test %s %v != expected value %v", test.Name, result, expected)
	}
	return nil
}

func NewTest(t *testing.T, context *decimal.Context, line string) (*Test, *decimal.Context) {
	line = strings.Replace(line, "\"", "", -1)
	line = strings.Replace(line, "'", "", -1)
	split := strings.Split(line, " ")
	switch split[0] {
	case "--":
		//comment
		return nil, nil
	case " ":
		//empty line
		return nil, nil
	}
	if DirectiveRe.MatchString(line) {
		groups := DirectiveRe.FindAllStringSubmatch(line, -1)
		if groups[0][1] == "precision" {
			precision, err := strconv.Atoi(groups[0][2])
			if err != nil {
				t.Fatalf("failed to convert precision %v %v", line, err)
			}
			ctx := decimal.NewContext(int32(precision))
			return nil, &ctx
		}
		return nil, nil
	}
	if TestRe.MatchString(line) {
		groups := TestRe.FindAllStringSubmatch(line, -1)
		return &Test{context, groups[0][1], groups[0][2], groups[0][3], groups[0][4], groups[0][5]}, nil
	}
	return nil, nil
}

func ReadFile(t *testing.T, filename string) []*Test {
	file, err := os.Open(filename)
	if err != nil {
		t.Fatalf("failed to read test file %v", err)
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	scanner := bufio.NewScanner(reader)

	tests := []*Test{}
	currentContext := &decimal.DefaultContext
	for scanner.Scan() {
		test, newContext := NewTest(t, currentContext, scanner.Text())
		if test != nil {
			tests = append(tests, test)
		}
		if newContext != nil {
			currentContext = newContext
		}
	}
	return tests
}

func runTests(t *testing.T, filename string) {
	tests := ReadFile(t, filename)
	pass := 0
	skipped := 0
	fail := 0
	total := 0
	for i, test := range tests {
		total += 1
		if err := test.Validate(); err != nil {
			//t.Logf("skipping test with validation error: %v", err)
			skipped += 1
			continue
		}
		err := test.Run()
		if err != nil {
			fail += 1
			t.Errorf("Test %d %s failed: %s", i, test.Name, err)
		} else {
			pass += 1
		}
	}
	t.Logf("Tests %v Pass %v Fail %v Skipped %v", total, pass, fail, skipped)
}

func TestDecTestAdd(t *testing.T) {
	runTests(t, "add.decTest")
}

func TestDecTestSub(t *testing.T) {
	runTests(t, "subtract.decTest")
}

func TestDecTestMultiply(t *testing.T) {
	runTests(t, "multiply.decTest")
}
