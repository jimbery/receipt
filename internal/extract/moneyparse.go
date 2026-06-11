package extract

import (
	"strconv"
	"strings"
)

// ParsePoundsToMinor converts a decimal pound string to minor units without float arithmetic.
// Supports negative amounts (credit notes). Examples: "12.99", "-5.00", "45".
func ParsePoundsToMinor(s string) (int64, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, false
	}
	negative := false
	if strings.HasPrefix(s, "-") {
		negative = true
		s = strings.TrimSpace(s[1:])
	}
	poundsPart, pencePart, _ := strings.Cut(s, ".")
	poundsPart = strings.TrimSpace(poundsPart)
	if poundsPart == "" && pencePart == "" {
		return 0, false
	}
	if poundsPart == "" {
		poundsPart = "0"
	}
	pounds, err := strconv.ParseInt(poundsPart, 10, 64)
	if err != nil {
		return 0, false
	}
	var pence int64
	if pencePart != "" {
		if len(pencePart) > 2 {
			pencePart = pencePart[:2]
		}
		for len(pencePart) < 2 {
			pencePart += "0"
		}
		pence, err = strconv.ParseInt(pencePart, 10, 64)
		if err != nil {
			return 0, false
		}
	}
	minor := pounds*100 + pence
	if negative {
		minor = -minor
	}
	return minor, true
}
