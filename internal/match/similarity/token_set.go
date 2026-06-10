package similarity

import (
	"slices"
	"strings"
)

// TokenSetRatio returns a similarity score in [0, 1] using token-set intersection.
func TokenSetRatio(a, b string) float64 {
	aTokens := uniqueTokens(a)
	bTokens := uniqueTokens(b)
	if len(aTokens) == 0 || len(bTokens) == 0 {
		return 0
	}
	if len(aTokens) == 1 && len(bTokens) == 1 && aTokens[0] == bTokens[0] {
		return 1
	}

	intersect := 0
	for _, at := range aTokens {
		if slices.Contains(bTokens, at) {
			intersect++
		}
	}
	union := len(aTokens) + len(bTokens) - intersect
	if union == 0 {
		return 0
	}
	return float64(intersect) / float64(union)
}

func uniqueTokens(s string) []string {
	parts := strings.FieldsFunc(strings.ToLower(s), func(r rune) bool {
		return (r < 'a' || r > 'z') && (r < '0' || r > '9')
	})
	seen := make(map[string]struct{}, len(parts))
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if len(p) < 2 {
			continue
		}
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		out = append(out, p)
	}
	return out
}
