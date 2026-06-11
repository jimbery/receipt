package extract

import "regexp"

var (
	htmlTotalRe   = regexp.MustCompile(`(?i)total[:\s]*£\s*(-?[0-9]+(?:\.[0-9]{2})?)`)
	totalIncVATRe = regexp.MustCompile(`(?i)total\s*\(inc\.?\s*vat\)[^£\d]{0,24}£\s*([0-9]+(?:\.[0-9]{2})?)`)
	totalPaidRe   = regexp.MustCompile(`(?i)total\s+paid[:\s]*£?\s*(-?[0-9]+(?:\.[0-9]{2})?)`)
	gbpTotalRe    = regexp.MustCompile(`(?is)total\s+(-?[0-9]+(?:\.[0-9]{2})?)\s+gbp`)
	refundTotalRe = regexp.MustCompile(`(?i)refund[:\s]*£\s*(-?[0-9]+(?:\.[0-9]{2})?)`)
	orderRefRe    = regexp.MustCompile(`(?i)order\s*(?:#|no\.?|number)[:\s]*([A-Z0-9\-]+)`)
	lineItemRe    = regexp.MustCompile(`(?i)([A-Za-z0-9][^\n]{2,40})\s+£\s*(-?[0-9]+\.[0-9]{2})`)
)

func parseHTMLTotal(body string) (int64, bool) {
	m := htmlTotalRe.FindStringSubmatch(body)
	if len(m) < 2 {
		return 0, false
	}
	return ParsePoundsToMinor(m[1])
}

// ParseReceiptTotal tries inline, table, paid, and GBP total patterns.
func ParseReceiptTotal(body string) (int64, bool) {
	if m := totalIncVATRe.FindStringSubmatch(body); len(m) > 1 {
		if minor, ok := ParsePoundsToMinor(m[1]); ok {
			return minor, true
		}
	}
	if minor, ok := parseHTMLTotal(body); ok {
		return minor, true
	}
	if minor, ok := ParseHTMLTableTotal(body); ok {
		return minor, true
	}
	if m := totalPaidRe.FindStringSubmatch(body); len(m) > 1 {
		return ParsePoundsToMinor(m[1])
	}
	if m := gbpTotalRe.FindStringSubmatch(body); len(m) > 1 {
		return ParsePoundsToMinor(m[1])
	}
	if m := refundTotalRe.FindStringSubmatch(body); len(m) > 1 {
		return ParsePoundsToMinor(m[1])
	}
	return 0, false
}

func parseOrderRef(body string) (string, bool) {
	m := orderRefRe.FindStringSubmatch(body)
	if len(m) < 2 {
		return "", false
	}
	return m[1], true
}
