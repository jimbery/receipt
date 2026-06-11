package scrub_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jimbery/receipt/internal/scrub"
)

func TestExportEML_ScrubsVolunteerEmail(t *testing.T) {
	pilot := scrub.PilotBlocklist{EmailLocalParts: []string{"pilotuser"}}
	s := scrub.NewWithPilotBlocklist(nil, pilot)
	raw := []byte("From: noreply@toolstation.com\r\nTo: pilotuser@hotmail.com\r\n\r\nOrder total £12.99\r\n")
	out, err := scrub.ExportEML(raw, s)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(strings.ToLower(string(out)), "pilotuser") {
		t.Fatalf("volunteer email remains: %s", out)
	}
	if !strings.Contains(string(out), "REDACTED_EMAIL@example.com") {
		t.Fatalf("expected redacted email: %s", out)
	}
}

func TestScrubbedVolunteerEMLs_AuditClean(t *testing.T) {
	root := filepath.Join("..", "..", "test", "testdata", "email", "merchants")
	_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(strings.ToLower(info.Name()), ".eml") {
			return nil
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		if auditErr := scrub.AuditFixture(string(data)); auditErr != nil {
			t.Errorf("%s: %v", path, auditErr)
		}
		return nil
	})
}
