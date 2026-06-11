package scrub

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	redactedName     = "REDACTED_NAME"
	redactedPostcode = "REDACTED_POSTCODE"
	redactedAddress  = "REDACTED_ADDRESS"
	redactedArea     = "REDACTED_AREA"
	redactedOrder    = "REDACTED_ORDER"
)

var (
	ukPostcodePattern = regexp.MustCompile(
		`(?i)\b([A-Z]{1,2}\d[A-Z\d]?\s+\d[A-Z]{2})\b`,
	)
	dearNamePattern = regexp.MustCompile(
		`(?i)\bDear\s+(?:Mr|Mrs|Ms|Miss|Dr)\.?\s+[A-Z][a-z]+(?:\s+[A-Z][a-z]+)+`,
	)
	hiNamePattern = regexp.MustCompile(
		`(?i)\bHi,?\s+[A-Z][a-z]+\b`,
	)
	thanksOrderNamePattern = regexp.MustCompile(
		`(?i)(?:Thanks for your order|anks for your order),?\s+[A-Z][a-z]+!`,
	)
	nameDashCityPattern = regexp.MustCompile(
		`\b[A-Z][a-z]+\s+(?:=E2=80=93\s*|[–—-]\s*)[A-Z]{2,}\b`,
	)
	streetAddressPattern = regexp.MustCompile(
		`(?i)\b\d+\s+[A-Z][A-Za-z\s,]{3,40}(?:Road|Street|Lane|Way|Close|Avenue|Drive|Hill)\b[^,\n]{0,80}` +
			`(?:,\s*[A-Za-z\s]{2,30})?\s*,?\s*[A-Z]{1,2}\d[A-Z\d]?\s+\d[A-Z]{2}\b`,
	)
	encodedNameDashPattern = regexp.MustCompile(
		`\b[A-Z][a-z]+\s*=E2=80=93\s*[A-Za-z=]+`,
	)
	collectionAreaPattern = regexp.MustCompile(
		`(?i)\bKings Heath\b`,
	)
	toolstationOrderPattern = regexp.MustCompile(`(?i)\bYWW\d+\b`)
	screwfixOrderPattern    = regexp.MustCompile(`(?i)\bA\d{8,}\b`)
	amazonOrderIDPattern    = regexp.MustCompile(`(?i)orderID=3D?[A-Z0-9]{8,}`)
	orderNumberLinePattern  = regexp.MustCompile(
		`(?i)(?:Order\s+(?:number|#|ref)[:\s]*)(?:A\d{8,}|YWW\d+)`,
	)
)

// DetectPIIViolations reports pattern-based PII still present after scrubbing.
func DetectPIIViolations(content string) []string {
	var out []string
	check := func(label string, re *regexp.Regexp) {
		if re.FindStringIndex(content) != nil {
			out = append(out, label)
		}
	}
	check("uk_postcode", ukPostcodePattern)
	check("dear_name", dearNamePattern)
	check("hi_name", hiNamePattern)
	check("thanks_order_name", thanksOrderNamePattern)
	check("name_dash_city", nameDashCityPattern)
	check("street_address", streetAddressPattern)
	check("encoded_name_dash", encodedNameDashPattern)
	check("collection_area", collectionAreaPattern)
	check("toolstation_order", toolstationOrderPattern)
	check("screwfix_order", screwfixOrderPattern)
	for _, m := range amazonOrderIDPattern.FindAllString(content, -1) {
		if !strings.Contains(strings.ToUpper(m), "REDACTED") {
			out = append(out, "amazon_order_id")
			break
		}
	}
	for _, p := range blockedFixturePatterns() {
		if strings.Contains(strings.ToLower(content), strings.ToLower(p)) {
			out = append(out, "blocked:"+p)
		}
	}
	if s := New(nil); s.ContainsPII(content) {
		out = append(out, "dictionary_literal")
	}
	return out
}

func scrubPatternPII(input string) string {
	out := input
	out = dearNamePattern.ReplaceAllString(out, "Dear "+redactedName)
	out = hiNamePattern.ReplaceAllString(out, "Hi, "+redactedName)
	out = thanksOrderNamePattern.ReplaceAllString(out, "Thanks for your order, "+redactedName+"!")
	out = nameDashCityPattern.ReplaceAllString(out, redactedName+" – REDACTED_CITY")
	out = streetAddressPattern.ReplaceAllString(out, redactedAddress)
	out = encodedNameDashPattern.ReplaceAllString(out, redactedName+" =E2=80=93 REDACTED_CITY")
	out = collectionAreaPattern.ReplaceAllString(out, redactedArea)
	out = ukPostcodePattern.ReplaceAllString(out, redactedPostcode)
	out = toolstationOrderPattern.ReplaceAllString(out, redactedOrder)
	out = screwfixOrderPattern.ReplaceAllString(out, redactedOrder)
	out = amazonOrderIDPattern.ReplaceAllString(out, "orderID=REDACTED")
	out = orderNumberLinePattern.ReplaceAllString(out, "Order number: "+redactedOrder)
	return out
}

// AuditFixture reports whether scrubbed content is safe to commit under test/testdata/email/.
func AuditFixture(content string) error {
	if violations := DetectPIIViolations(content); len(violations) > 0 {
		return fmt.Errorf("PII patterns remain: %s", strings.Join(violations, ", "))
	}
	return nil
}
