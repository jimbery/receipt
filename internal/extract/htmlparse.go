package extract

import (
	"encoding/json"
	"regexp"
	"strings"
)

var (
	jsonLDScriptRe = regexp.MustCompile(`(?is)<script[^>]*type=["']application/ld\+json["'][^>]*>(.*?)</script>`)
	htmlTagRe      = regexp.MustCompile(`<[^>]+>`)
	tableRowRe     = regexp.MustCompile(`(?is)<tr[^>]*>(.*?)</tr>`)
	tableCellRe    = regexp.MustCompile(`(?is)<t[dh][^>]*>(.*?)</t[dh]>`)
)

// ExtractVisibleText strips HTML tags and collapses whitespace.
func ExtractVisibleText(html string) string {
	text := htmlTagRe.ReplaceAllString(html, " ")
	text = strings.ReplaceAll(text, "&nbsp;", " ")
	text = strings.ReplaceAll(text, "&pound;", "£")
	fields := strings.Fields(text)
	return strings.Join(fields, " ")
}

// ParseJSONLDTotal extracts a total price from schema.org JSON-LD script blocks.
func ParseJSONLDTotal(html string) (int64, bool) {
	scripts := jsonLDScriptRe.FindAllStringSubmatch(html, -1)
	for _, m := range scripts {
		if len(m) < 2 {
			continue
		}
		var doc any
		if err := json.Unmarshal([]byte(strings.TrimSpace(m[1])), &doc); err != nil {
			continue
		}
		if minor, ok := jsonLDPriceMinor(doc); ok {
			return minor, true
		}
	}
	return 0, false
}

func jsonLDPriceMinor(v any) (int64, bool) {
	switch t := v.(type) {
	case map[string]any:
		if price, ok := t["price"]; ok {
			return priceToMinor(price)
		}
		if offers, hasOffers := t["offers"]; hasOffers {
			if minor, found := jsonLDPriceMinor(offers); found {
				return minor, true
			}
		}
		for _, child := range t {
			if minor, found := jsonLDPriceMinor(child); found {
				return minor, true
			}
		}
	case []any:
		for _, item := range t {
			if minor, found := jsonLDPriceMinor(item); found {
				return minor, true
			}
		}
	}
	return 0, false
}

func priceToMinor(price any) (int64, bool) {
	switch p := price.(type) {
	case string:
		return ParsePoundsToMinor(strings.TrimPrefix(p, "£"))
	case float64:
		// JSON numbers for prices are avoided in fixtures; reject float path.
		return 0, false
	case json.Number:
		return ParsePoundsToMinor(p.String())
	default:
		return 0, false
	}
}

// ParseHTMLTableTotal finds a cell labelled "total" in an HTML table and parses the amount.
func ParseHTMLTableTotal(html string) (int64, bool) {
	rows := tableRowRe.FindAllStringSubmatch(html, -1)
	var lastMinor int64
	var found bool
	for _, row := range rows {
		cells := tableCellRe.FindAllStringSubmatch(row[1], -1)
		if len(cells) < 2 {
			continue
		}
		label := strings.ToLower(ExtractVisibleText(cells[0][1]))
		if !strings.Contains(label, "total") || strings.Contains(label, "subtotal") {
			continue
		}
		if strings.Contains(label, "ex.") && strings.Contains(label, "vat") {
			continue
		}
		value := ExtractVisibleText(cells[len(cells)-1][1])
		if minor, ok := parseMoneyToken(value); ok {
			lastMinor = minor
			found = true
		}
	}
	return lastMinor, found
}

var moneyTokenRe = regexp.MustCompile(`£\s*(-?[0-9]+(?:\.[0-9]{2})?)`)

func parseMoneyToken(s string) (int64, bool) {
	m := moneyTokenRe.FindStringSubmatch(s)
	if len(m) < 2 {
		return 0, false
	}
	return ParsePoundsToMinor(m[1])
}
