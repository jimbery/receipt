package scrub_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jimbery/receipt/internal/scrub"
)

func TestScrubber_ReplacesPII(t *testing.T) {
	s := scrub.New(nil)
	in := "Contact Volunteer Name at volunteer@example.com or +44 7700 900123"
	out := s.Scrub(in)
	if s.ContainsPII(out) {
		t.Fatalf("scrubbed output still contains dictionary literals: %q", out)
	}
	if strings.Contains(out, "Volunteer Name") {
		t.Fatalf("name not redacted: %q", out)
	}
}

func TestScrubber_PreservesStructure(t *testing.T) {
	s := scrub.New(nil)
	in := "Order total: £12.99\nOrder #ABC-123"
	out := s.Scrub(in)
	if !strings.Contains(out, "Order total") || !strings.Contains(out, "ABC-123") {
		t.Fatalf("structure lost: %q", out)
	}
}

func TestScrubFixture_RedactsPatternPII(t *testing.T) {
	pilot := scrub.PilotBlocklist{AreaNames: []string{"Northgate"}}
	s := scrub.NewWithPilotBlocklist(nil, pilot)
	in := strings.Join([]string{
		"Dear Mr Alex Smith",
		"Hi, Alex",
		"Thanks for your order, Alex!",
		"Alex – BIRMINGHAM",
		"BS1 4TR",
		"Northgate",
		"Order YWW074091737",
		"Confirmation A16237359139",
	}, "\n")
	out := s.ScrubFixture(in)
	if err := scrub.AuditFixture(out); err != nil {
		t.Fatalf("audit: %v\nout=%q", err, out)
	}
}

// Property: committed email fixtures pass pattern-based PII audit (not dictionary-only).
func TestScrubber_NoPIIInCommittedFixtures(t *testing.T) {
	root := filepath.Join("..", "..", "test", "testdata", "email")
	var files []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.IsDir() || strings.HasSuffix(path, "manifest.json") {
			return nil
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Fatal("no fixtures under test/testdata/email")
	}
	for _, path := range files {
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			t.Fatal(readErr)
		}
		if auditErr := scrub.AuditFixture(string(data)); auditErr != nil {
			t.Fatalf("%s: %v", path, auditErr)
		}
	}
}

func TestDetectPIIViolations_FindsRealLeak(t *testing.T) {
	v := scrub.DetectPIIViolations("Dear Mr Alex Smith\nBS1 4TR", scrub.PilotBlocklist{})
	if len(v) == 0 {
		t.Fatal("expected violations")
	}
}
