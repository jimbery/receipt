package match

import "strings"

const (
	neutralMCCScore = 0.5
	weakMCCScore    = 0.3
)

// mccCorroboration is a weak signal only — never filters candidates.
func mccCorroboration(mcc, supplier, hint string) float64 {
	if mcc == "" {
		return neutralMCCScore
	}
	keywords, ok := mccKeywordHints()[mcc]
	if !ok {
		return neutralMCCScore
	}
	combined := strings.ToLower(supplier + " " + hint)
	for _, kw := range keywords {
		if strings.Contains(combined, kw) {
			return 1.0
		}
	}
	return weakMCCScore
}

// MTD-cohort oriented MCC hints for Phase 0 synthetic evaluation.
func mccKeywordHints() map[string][]string {
	return map[string][]string{
		"5251": {"screwfix", "toolstation", "build", "trade", "wickes", "bq"},
		"5541": {"shell", "bp", "esso", "fuel", "petrol"},
		"5542": {"fuel", "petrol"},
		"5399": {"amazon"},
		"4816": {"subscription", "software"},
	}
}
