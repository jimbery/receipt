package scrub_test

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/jimbery/receipt/internal/scrub"
)

type fixtureManifest struct {
	Version       int             `json:"version"`
	SignedOffBy   string          `json:"signed_off_by"`
	SignedOffDate string          `json:"signed_off_date"`
	Note          string          `json:"note"`
	Fixtures      []manifestEntry `json:"fixtures"`
}

type manifestEntry struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

func TestFixtureManifest_SignedOffEMLs(t *testing.T) {
	root := filepath.Join("..", "..", "test", "testdata", "email")
	data, err := os.ReadFile(filepath.Join(root, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var m fixtureManifest
	if unmarshalErr := json.Unmarshal(data, &m); unmarshalErr != nil {
		t.Fatal(unmarshalErr)
	}
	if m.SignedOffBy == "" || m.SignedOffDate == "" {
		t.Fatal("manifest missing human sign-off")
	}
	if len(m.Fixtures) == 0 {
		t.Fatal("manifest has no fixtures")
	}
	for _, entry := range m.Fixtures {
		path := filepath.Join(root, filepath.FromSlash(entry.Path))
		raw, readErr := os.ReadFile(path)
		if readErr != nil {
			t.Fatalf("%s: %v", entry.Path, readErr)
		}
		sum := sha256.Sum256(raw)
		got := hex.EncodeToString(sum[:])
		if got != entry.SHA256 {
			t.Fatalf("%s: sha256 mismatch got %s want %s", entry.Path, got, entry.SHA256)
		}
		if auditErr := scrub.AuditFixture(string(raw)); auditErr != nil {
			t.Fatalf("%s: %v", entry.Path, auditErr)
		}
	}
}
