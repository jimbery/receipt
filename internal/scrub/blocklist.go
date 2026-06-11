package scrub

import (
	"encoding/json"
	"os"
	"regexp"
	"strings"
)

// PilotBlocklist holds mailbox-specific literals loaded from gitignored emls/pilot-blocklist.json.
// Committed source and CI audit use an empty blocklist; scrub-eml loads the local file.
type PilotBlocklist struct {
	EmailLocalParts []string `json:"email_local_parts"`
	AreaNames       []string `json:"area_names"`
	LiteralNames    []string `json:"literal_names"`
}

// LoadPilotBlocklist reads optional JSON. A missing file yields an empty blocklist.
func LoadPilotBlocklist(path string) (PilotBlocklist, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return PilotBlocklist{}, nil
		}
		return PilotBlocklist{}, err
	}
	var b PilotBlocklist
	if unmarshalErr := json.Unmarshal(data, &b); unmarshalErr != nil {
		return PilotBlocklist{}, unmarshalErr
	}
	return b, nil
}

func (b PilotBlocklist) scrubAreas(input string) string {
	out := input
	for _, area := range b.AreaNames {
		if area == "" {
			continue
		}
		re := regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(area) + `\b`)
		out = re.ReplaceAllString(out, redactedArea)
	}
	return out
}

func (b PilotBlocklist) detectAreaViolations(content string) []string {
	var out []string
	for _, area := range b.AreaNames {
		if area == "" {
			continue
		}
		re := regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(area) + `\b`)
		if re.FindStringIndex(content) != nil {
			out = append(out, "collection_area")
			break
		}
	}
	return out
}

func blockedFixturePatterns(pilot PilotBlocklist) []string {
	out := []string{"@hotmail.com", "@outlook.com", "@live.com"}
	for _, p := range pilot.EmailLocalParts {
		if p != "" {
			out = append(out, strings.ToLower(p))
		}
	}
	for _, n := range pilot.LiteralNames {
		if n != "" {
			out = append(out, strings.ToLower(n))
		}
	}
	return out
}
