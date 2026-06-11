package scrub

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
)

const maxFixtureSlugLen = 64

var (
	slugSanitizer     = regexp.MustCompile(`[^a-z0-9]+`)
	orderSlugStripper = regexp.MustCompile(`(?i)(?:yww\d+|a\d{8,}|confirmation-of-your-order-a\d+)`)
)

func blockedFixturePatterns() []string {
	return []string{
		"jayimbery",
		"jay imbery",
		"@hotmail.com",
		"@outlook.com",
		"@live.com",
	}
}

// ContainsPII reports whether dictionary literals or pattern PII remain.
func ContainsPII(content string) bool {
	return len(DetectPIIViolations(content)) > 0
}

// MerchantFromEML guesses the merchant folder name from RFC822 content.
func MerchantFromEML(content string) (string, error) {
	lower := strings.ToLower(content)
	switch {
	case strings.Contains(lower, "toolstation.com"):
		return "toolstation", nil
	case strings.Contains(lower, "screwfix.com"):
		return "screwfix", nil
	case strings.Contains(lower, "amazon.co.uk"), strings.Contains(lower, "amazon.com"):
		return "amazon", nil
	case strings.Contains(lower, "diy.com"), strings.Contains(lower, "bandq"):
		return "bandq", nil
	case strings.Contains(lower, "shell.com"):
		return "shell", nil
	case strings.Contains(lower, "bp.com"):
		return "bp", nil
	case strings.Contains(lower, "octopus.energy"):
		return "octopus", nil
	case strings.Contains(lower, "edfenergy"):
		return "edf", nil
	default:
		return "", errors.New("unknown merchant in message")
	}
}

// SlugifyFilename produces a stable fixture basename with order identifiers stripped.
func SlugifyFilename(name string) string {
	base := strings.TrimSuffix(filepath.Base(name), filepath.Ext(name))
	base = strings.ToLower(base)
	base = orderSlugStripper.ReplaceAllString(base, "order")
	base = slugSanitizer.ReplaceAllString(base, "-")
	base = strings.Trim(base, "-")
	if base == "" {
		return "fixture"
	}
	if len(base) > maxFixtureSlugLen {
		base = base[:maxFixtureSlugLen]
		base = strings.TrimRight(base, "-")
	}
	return base
}

// ExportEML scrubs raw RFC822 content and validates it is safe to commit.
func ExportEML(raw []byte, s *Scrubber) ([]byte, error) {
	if s == nil {
		s = New(nil)
	}
	scrubbed := s.ScrubFixture(string(raw))
	if err := AuditFixture(scrubbed); err != nil {
		return nil, err
	}
	return []byte(scrubbed), nil
}

// WriteFixture writes scrubbed content to destDir/merchant/name.eml.
func WriteFixture(destDir, merchant, basename string, content []byte) (string, error) {
	dir := filepath.Join(destDir, merchant)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return "", err
	}
	path := filepath.Join(dir, basename+".eml")
	if err := os.WriteFile(path, content, 0o600); err != nil {
		return "", err
	}
	return path, nil
}

// IsPrintableASCII reports whether s is a reasonable fixture slug fragment.
func IsPrintableASCII(s string) bool {
	for _, r := range s {
		if r > unicode.MaxASCII || (!unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '-') {
			return false
		}
	}
	return true
}

// FixtureRelPath joins merchant and basename for manifest entries.
func FixtureRelPath(merchant, basename string) string {
	return filepath.ToSlash(filepath.Join("merchants", merchant, basename+".eml"))
}
