package difflibgo_test

import (
	"strings"
	"testing"

	"github.com/carlmontanari/difflibgo/difflibgo"
)

func TestUnifiedDiff(t *testing.T) {
	cases := []struct {
		name     string
		a        string
		b        string
		expected string
	}{
		{
			name: "simple-no-diff",
			a: `abc
def
123
xyz`,
			b: `abc
def
123
xyz`,
			expected: `  abc
  def
  123
  xyz`,
		},
		{
			name: "simple-diff",
			a: `abc
defq
123
xyz`,
			b: `abc
def
z123
xyz`,
			expected: `  abc
- defq
- 123
+ def
+ z123
  xyz`,
		},
	}

	for _, testCase := range cases {
		t.Run(
			testCase.name,
			func(t *testing.T) {
				actual := difflibgo.UnifiedDiff(testCase.a, testCase.b)

				if actual != testCase.expected {
					failOutput(
						t,
						strings.Split(actual, "\n"),
						strings.Split(testCase.expected, "\n"),
					)
				}
			},
		)
	}
}

func TestUnifiedDiffColorized(t *testing.T) {
	cases := []struct {
		name     string
		a        string
		b        string
		expected string
	}{
		{
			name: "simple-no-diff",
			a: `abc
def
123
xyz`,
			b: `abc
def
123
xyz`,
			expected: `abc
def
123
xyz`,
		},
		{
			name: "simple-diff",
			a: `abc
defq
123
xyz`,
			b: `abc
def
z123
xyz`,
			expected: `abc
[91mdefq[0m
[91m123[0m
[92mdef[0m
[92mz123[0m
xyz`,
		},
	}

	for _, testCase := range cases {
		t.Run(
			testCase.name,
			func(t *testing.T) {
				actual := difflibgo.UnifiedDiffColorized(testCase.a, testCase.b)

				if actual != testCase.expected {
					failOutput(
						t,
						strings.Split(actual, "\n"),
						strings.Split(testCase.expected, "\n"),
					)
				}
			},
		)
	}
}
