package difflibgo_test

import (
	"strings"
	"testing"

	"github.com/carlmontanari/difflibgo/difflibgo"
)

func failOutput(t *testing.T, actual, expected []string) {
	t.Helper()

	t.Fatalf(
		"actual and expectd do not match...\n"+
			"*** actual   >>>\n"+
			"%s\n"+
			"<<< actual   ***\n"+
			"*** expected >>>\n"+
			"%s\n"+
			"<<< expected ***",
		strings.Join(actual, "\n"),
		strings.Join(expected, "\n"),
	)
}

func TestDifferCompare(t *testing.T) {
	cases := []struct {
		name     string
		a        []string
		b        []string
		expected []string
	}{
		{
			name: "simple-no-diff",
			a: []string{
				"abc",
				"def",
				"123",
				"xyz",
			},
			b: []string{
				"abc",
				"def",
				"123",
				"xyz",
			},
			expected: []string{
				"  abc",
				"  def",
				"  123",
				"  xyz",
			},
		},
		{
			name: "simple-diff",
			a: []string{
				"abc",
				"defq",
				"123",
				"xyz",
			},
			b: []string{
				"abc",
				"def",
				"123",
				"xyz9",
			},
			expected: []string{
				"  abc",
				"- defq",
				"+ def",
				"  123",
				"- xyz",
				"+ xyz9",
			},
		},
		{
			name: "simple-diff-different-length-inputs",
			a: []string{
				"abc",
				"defq",
				"123",
			},
			b: []string{
				"abc",
				"def",
				"123",
				"xyz9",
			},
			expected: []string{
				"  abc",
				"- defq",
				"+ def",
				"  123",
				"+ xyz9",
			},
		},
		{
			name: "simple-diff-different-length-inputs-the-other-way",
			a: []string{
				"abc",
				"defq",
				"123",
				"xyz",
			},
			b: []string{
				"abc",
				"def",
				"123",
			},
			expected: []string{
				"  abc",
				"- defq",
				"+ def",
				"  123",
				"- xyz",
			},
		},
		{
			name: "regression-index-out-of-range-in-differ-dump",
			a: []string{
				"aaaa",
			},
			b: []string{
				"bbbb",
			},
			expected: []string{
				"- aaaa",
				"+ bbbb",
			},
		},
	}

	for _, testCase := range cases {
		t.Run(
			testCase.name,
			func(t *testing.T) {
				actual := difflibgo.Compare(testCase.a, testCase.b)

				if len(actual) != len(testCase.expected) {
					t.Log("actual and expected lengths differ...")

					failOutput(t, actual, testCase.expected)
				}

				var fail bool

				for idx, actualLine := range actual {
					if testCase.expected[idx] != actualLine {
						t.Logf("actual and expected lines at index %d differ...", idx)
						t.Logf("actual  : %s", actualLine)
						t.Logf("expected: %s", testCase.expected[idx])

						fail = true
					}
				}

				if fail {
					failOutput(t, actual, testCase.expected)
				}
			},
		)
	}
}
