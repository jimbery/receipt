package scrub

import (
	"regexp"
	"strings"
)

// Scrubber replaces PII tokens in email content for fixture export.
type Scrubber struct {
	replacements map[string]string
	pilot        PilotBlocklist
}

// New returns a scrubber with default UK-oriented PII patterns.
func New(dict map[string]string) *Scrubber {
	return NewWithPilotBlocklist(dict, PilotBlocklist{})
}

// NewWithPilotBlocklist returns a scrubber with optional gitignored pilot literals.
func NewWithPilotBlocklist(dict map[string]string, pilot PilotBlocklist) *Scrubber {
	if dict == nil {
		dict = map[string]string{
			"Volunteer Name":                     "REDACTED_NAME",
			"volunteer@example.com":              "REDACTED_EMAIL@example.com",
			"1 Example Street, London, SW1A 1AA": "REDACTED_ADDRESS",
			"+44 7700 900123":                    "REDACTED_PHONE",
			"12345678":                           "REDACTED_ACCOUNT",
		}
	}
	return &Scrubber{replacements: dict, pilot: pilot}
}

// Scrub replaces dictionary literals and common email/phone patterns.
func (s *Scrubber) Scrub(input string) string {
	out := s.replaceDictionary(input)
	out = scrubPatternPII(out, s.pilot)
	out = emailPattern.ReplaceAllString(out, "REDACTED_EMAIL@example.com")
	out = phonePattern.ReplaceAllString(out, "REDACTED_PHONE")
	return out
}

// ScrubFixture redacts volunteer PII but preserves merchant sender addresses for routing tests.
func (s *Scrubber) ScrubFixture(input string) string {
	out := s.replaceDictionary(input)
	out = scrubVolunteerEmails(out, s.pilot)
	out = scrubPatternPII(out, s.pilot)
	out = phonePattern.ReplaceAllString(out, "REDACTED_PHONE")
	return out
}

func (s *Scrubber) replaceDictionary(input string) string {
	out := input
	for literal, replacement := range s.replacements {
		out = strings.ReplaceAll(out, literal, replacement)
	}
	return out
}

func scrubVolunteerEmails(input string, pilot PilotBlocklist) string {
	return emailPattern.ReplaceAllStringFunc(input, func(email string) string {
		lower := strings.ToLower(email)
		local := strings.Split(lower, "@")[0]
		for _, part := range pilot.EmailLocalParts {
			if part != "" && strings.Contains(local, strings.ToLower(part)) {
				return "REDACTED_EMAIL@example.com"
			}
		}
		if strings.Contains(lower, "@hotmail.") ||
			strings.Contains(lower, "@outlook.") ||
			strings.Contains(lower, "@live.") {
			return "REDACTED_EMAIL@example.com"
		}
		return email
	})
}

// ContainsPII reports whether any scrub-dictionary source literal remains in output.
func (s *Scrubber) ContainsPII(scrubbed string) bool {
	for literal := range s.replacements {
		if strings.Contains(scrubbed, literal) {
			return true
		}
	}
	return false
}

var (
	emailPattern = regexp.MustCompile(`[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`)
	phonePattern = regexp.MustCompile(`\+?\d[\d\s\-()]{8,}\d`)
)
