package difflibgo

import (
	"strings"
)

func getDiffLines(a, b string) []string {
	return Compare(
		strings.Split(a, "\n"),
		strings.Split(b, "\n"),
	)
}

// UnifiedDiff accepts a and b which can be strings, bytes, or something you can marshal to json
// and returns a python difflib style diff string or an error (if there is a marshall error).
func UnifiedDiff(a, b string) string {
	return strings.Join(getDiffLines(a, b), "\n")
}

// UnifiedDiffColorized is the same as UnifiedDiff but instead of the diff symbols (+, -, ?) the
// line is rewritten with green, red, yellow (respectively) colorization.
func UnifiedDiffColorized(a, b string) string {
	diffLines := getDiffLines(a, b)

	unifiedDiffLines := make([]string, len(diffLines))

	for idx, line := range diffLines {
		var diffLine string

		switch line[:2] {
		case diffUnknown:
			diffLine = yellow + line[2:] + end
		case diffSubtraction:
			diffLine = red + line[2:] + end
		case diffAddition:
			diffLine = green + line[2:] + end
		default:
			diffLine = line[2:]
		}

		unifiedDiffLines[idx] = diffLine
	}

	return strings.Join(unifiedDiffLines, "\n")
}
