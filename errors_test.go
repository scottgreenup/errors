package errors

import (
	stderrs "errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewWithStackTrace_Formatting(t *testing.T) {
	tests := []struct {
		name     string
		format   string
		expected string
	}{
		{
			name:     "%s",
			format:   "%s",
			expected: "MyError",
		},
		{
			name:     "%q",
			format:   "%q",
			expected: "\"MyError\"",
		},
		{
			name:     "%v",
			format:   "%v",
			expected: "MyError",
		},
		{
			name:   "%+v",
			format: "%+v",
			expected: strings.Join(
				[]string{
					"MyError",
					"errors.TestNewWithStackTrace_Formatting.func1",
					"\t/.+/github.com/scottgreenup/errors/errors_test.go:[0-9]+",
					"testing.tRunner",
					"\t/.+:.+",
					"runtime.goexit",
					"\t/.+:.+",
					"",
				},
				"\n",
			),
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Arrange
			err := NewWithStackTrace("MyError")

			// Act
			result := fmt.Sprintf(testCase.format, err)

			// Assert
			assert.Regexp(t, testCase.expected, result)
		})
	}
}

func createErrorAtDepth(depth int) error {
	if depth == 0 {
		return NewWithStackTrace("DepthError")
	}

	return createErrorAtDepth(depth - 1)
}

func repeatArray(elements []string, n int) []string {
	result := make([]string, len(elements)*n)
	for i := 0; i < n; i++ {
		for j := 0; j < len(elements); j++ {
			result[(i*len(elements))+j] = elements[j]
		}
	}
	return result
}

func TestNewWithStackTrace_StackTraceDepth(t *testing.T) {
	createDepthStackTrace := func(depth int) string {
		var expectedResult []string
		expectedMaximum := 32
		expectedResult = append(
			expectedResult,
			"DepthError",
		)
		expectedResult = append(
			expectedResult,
			repeatArray([]string{
				"errors.createErrorAtDepth",
				"\t/.+/github.com/scottgreenup/errors/errors_test.go:[0-9]+",
			}, depth+1)...,
		)
		expectedResult = append(
			expectedResult,
			"errors.TestNewWithStackTrace_StackTraceDepth.func[0-9]+",
			"\t/.+/github.com/scottgreenup/errors/errors_test.go:[0-9]+",
			"testing.tRunner",
			"\t/.+:.+",
			"runtime.goexit",
			"\t/.+:.+",
		)

		cappedLength := min(expectedMaximum, len(expectedResult))
		return strings.Join(expectedResult[0:cappedLength], "\n")
	}

	tests := []struct {
		name     string
		depth    int
		expected string
	}{
		{
			name:     "%+v",
			depth:    0,
			expected: createDepthStackTrace(0),
		},
		{
			name:     "%+v",
			depth:    1,
			expected: createDepthStackTrace(1),
		},
		{
			name:     "%+v",
			depth:    28,
			expected: createDepthStackTrace(28),
		},
		{
			name:     "%+v",
			depth:    29,
			expected: createDepthStackTrace(29),
		},
		{
			name:     "%+v",
			depth:    50,
			expected: createDepthStackTrace(50),
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Arrange
			err := createErrorAtDepth(testCase.depth)

			// Act
			result := fmt.Sprintf("%+v", err)

			// Assert
			assert.Regexp(t, testCase.expected, result)
		})
	}
}

func TestWrap(t *testing.T) {
	tests := []struct {
		name     string
		provider func() []error
	}{
		{
			name: "New > Wrap",
			provider: func() []error {
				a := New("inner")
				b := Wrap(a, "outer")
				return []error{a, b}
			},
		},
		{
			name: "NewWithStackTrace > Wrap",
			provider: func() []error {
				a := NewWithStackTrace("inner")
				b := Wrap(a, "outer")
				return []error{a, b}
			},
		},
		{
			name: "New > WrapWithStackTrace",
			provider: func() []error {
				a := New("inner")
				b := WrapWithStackTrace(a, "outer")
				return []error{a, b}
			},
		},
		{
			name: "NewWithStackTrace > WrapWithStackTrace",
			provider: func() []error {
				a := NewWithStackTrace("inner")
				b := WrapWithStackTrace(a, "outer")
				return []error{a, b}
			},
		},
		{
			name: "Standard New > Standard Wrap",
			provider: func() []error {
				a := stderrs.New("inner")
				b := fmt.Errorf("outer: %w", a)
				return []error{a, b}
			},
		},
		{
			name: "Standard New > Wrap",
			provider: func() []error {
				a := stderrs.New("inner")
				b := Wrap(a, "outer")
				return []error{a, b}
			},
		},
		{
			name: "Standard New > Wrap > Standard Wrap",
			provider: func() []error {
				a := stderrs.New("inner")
				b := Wrap(a, "outer")
				c := fmt.Errorf("even outerer: %w", b)
				return []error{a, b, c}
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			errs := testCase.provider()
			currA := errs[len(errs)-1]
			currB := errs[len(errs)-1]
			for i := len(errs) - 1; i >= 0; i-- {
				assert.Equal(t, errs[i], currA)
				currA = Unwrap(currA)

				// Regression test to ensure that the we work with the standard library
				assert.Equal(t, errs[i], currB)
				currB = stderrs.Unwrap(currB)
			}
		})
	}
}

func TestIs(t *testing.T) {
	tests := []struct {
		name     string
		provider func() []error
	}{
		{
			name: "New > Wrap",
			provider: func() []error {
				a := New("inner")
				b := Wrap(a, "outer")
				return []error{a, b}
			},
		},
		{
			name: "New > Wrap > Wrap",
			provider: func() []error {
				a := New("inner")
				b := Wrap(a, "outer")
				c := Wrap(b, "outerer")
				return []error{a, b, c}
			},
		},
		{
			name: "New > Wrap > Wrap > Wrap",
			provider: func() []error {
				a := New("inner")
				b := Wrap(a, "outer")
				c := Wrap(b, "outerer")
				d := Wrap(c, "outererer")
				return []error{a, b, c, d}
			},
		},
		{
			name: "NewWithStackTrace > Wrap",
			provider: func() []error {
				a := NewWithStackTrace("inner")
				b := Wrap(a, "outer")
				return []error{a, b}
			},
		},
		{
			name: "NewWithStackTrace > Wrap > Wrap",
			provider: func() []error {
				a := NewWithStackTrace("inner")
				b := Wrap(a, "outer")
				c := Wrap(b, "outerer")
				return []error{a, b, c}
			},
		},
		{
			name: "New > WrapWithStackTrace",
			provider: func() []error {
				a := New("inner")
				b := WrapWithStackTrace(a, "outer")
				return []error{a, b}
			},
		},
		{
			name: "New > WrapWithStackTrace > WrapWithStackTrace",
			provider: func() []error {
				a := New("inner")
				b := WrapWithStackTrace(a, "outer")
				c := WrapWithStackTrace(b, "outerer")
				return []error{a, b, c}
			},
		},
		{
			name: "NewWithStackTrace > WrapWithStackTrace",
			provider: func() []error {
				a := NewWithStackTrace("inner")
				b := WrapWithStackTrace(a, "outer")
				return []error{a, b}
			},
		},
		{
			name: "NewWithStackTrace > WrapWithStackTrace > WrapWithStackTrace",
			provider: func() []error {
				a := NewWithStackTrace("inner")
				b := WrapWithStackTrace(a, "outer")
				c := WrapWithStackTrace(b, "outer")
				return []error{a, b, c}
			},
		},
		{
			name: "Standard New > Standard Wrap",
			provider: func() []error {
				a := stderrs.New("inner")
				b := fmt.Errorf("outer: %w", a)
				return []error{a, b}
			},
		},
		{
			name: "Standard New > Wrap",
			provider: func() []error {
				a := stderrs.New("inner")
				b := Wrap(a, "outer")
				return []error{a, b}
			},
		},
		{
			name: "Standard New > Wrap > Standard Wrap",
			provider: func() []error {
				a := stderrs.New("inner")
				b := Wrap(a, "outer")
				c := fmt.Errorf("even outerer: %w", b)
				return []error{a, b, c}
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			errs := testCase.provider()

			// For every error in the chain, it should equal each error below it AND itself.
			for i := len(errs) - 1; i >= 0; i-- {
				outerError := errs[i]
				for j := i; j >= 0; j-- {
					innerError := errs[j]
					assert.True(t, Is(outerError, innerError))
					assert.True(t, stderrs.Is(outerError, innerError))
				}
			}
		})
	}
}
