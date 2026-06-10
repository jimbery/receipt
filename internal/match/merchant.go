package match

import (
	"regexp"
	"strings"

	"github.com/jimbery/receipt/internal/match/similarity"
)

// MerchantComparer scores descriptor similarity. Phase 3 replaces BasicMerchantComparer.
type MerchantComparer interface {
	Similarity(txnDescriptor, receiptSupplier, receiptHint string) float64
}

// BasicMerchantComparer uses prefix stripping, token overlap, and Jaro-Winkler.
type BasicMerchantComparer struct{}

func (BasicMerchantComparer) Similarity(txnDescriptor, receiptSupplier, receiptHint string) float64 {
	txn := normaliseMerchant(txnDescriptor)
	if txn == "" {
		return 0
	}

	best := 0.0
	for _, candidate := range []string{receiptSupplier, receiptHint} {
		rec := normaliseMerchant(candidate)
		if rec == "" {
			continue
		}
		token := tokenOverlap(txn, rec)
		jw := similarity.JaroWinkler(txn, rec)
		score := token
		if jw > score {
			score = jw
		}
		if strings.Contains(txn, rec) || strings.Contains(rec, txn) {
			if contained := 0.85; contained > score {
				score = contained
			}
		}
		if score > best {
			best = score
		}
	}
	return best
}

const minMerchantTokenLen = 3

var nonAlnum = regexp.MustCompile(`[^a-z0-9]+`)

func normaliseMerchant(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	for _, prefix := range []string{"sq *", "sumup*", "paypal *", "pp*"} {
		if after, ok := strings.CutPrefix(s, prefix); ok {
			s = after
			break
		}
	}
	s = nonAlnum.ReplaceAllString(s, "")
	for len(s) > 0 && s[len(s)-1] >= '0' && s[len(s)-1] <= '9' {
		s = s[:len(s)-1]
	}
	return s
}

func tokenOverlap(a, b string) float64 {
	if a == b {
		return 1
	}
	aTokens := splitTokens(a)
	bTokens := splitTokens(b)
	if len(aTokens) == 0 || len(bTokens) == 0 {
		return 0
	}

	matches := 0
	for _, at := range aTokens {
		for _, bt := range bTokens {
			if at == bt || strings.Contains(at, bt) || strings.Contains(bt, at) {
				matches++
				break
			}
		}
	}
	denom := max(len(aTokens), len(bTokens))
	return float64(matches) / float64(denom)
}

func splitTokens(s string) []string {
	parts := nonAlnum.Split(s, -1)
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if len(p) >= minMerchantTokenLen {
			out = append(out, p)
		}
	}
	return out
}
