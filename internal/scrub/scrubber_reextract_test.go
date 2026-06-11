package scrub_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jimbery/receipt/internal/classify"
	"github.com/jimbery/receipt/internal/emailtypes"
	"github.com/jimbery/receipt/internal/extract"
	"github.com/jimbery/receipt/internal/mail"
	"github.com/jimbery/receipt/internal/scrub"
)

// M1.1 exit criterion: scrub → re-parse → classify/extract still yields the same receipt fields.
func TestScrubber_PreservesStructureViaReExtraction(t *testing.T) {
	root := filepath.Join("..", "..", "test", "testdata", "email", "merchants")
	clf := classify.New()
	reg := extract.NewRegistry()
	s := scrub.New(nil)

	var checked int
	_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(info.Name()))
		if ext != ".eml" && ext != ".txt" {
			return nil
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		before, ok := pipelineExtract(clf, reg, data, info.Name())
		if !ok {
			return nil
		}

		scrubbed := s.ScrubFixture(string(data))
		if auditErr := scrub.AuditFixture(scrubbed); auditErr != nil {
			t.Errorf("%s: %v", path, auditErr)
		}
		after, ok := pipelineExtract(clf, reg, []byte(scrubbed), info.Name())
		if !ok {
			t.Errorf("%s: extraction failed after scrub", path)
			return nil
		}

		if before.kind != after.kind {
			t.Errorf("%s: kind %s → %s after scrub", path, before.kind, after.kind)
		}
		if before.supplier != after.supplier {
			t.Errorf("%s: supplier %q → %q after scrub", path, before.supplier, after.supplier)
		}
		if before.totalMinor > 0 && before.totalMinor != after.totalMinor {
			t.Errorf("%s: total %d → %d after scrub", path, before.totalMinor, after.totalMinor)
		}
		if before.orderRef != "" && before.orderRef != after.orderRef {
			t.Errorf("%s: order ref %q → %q after scrub", path, before.orderRef, after.orderRef)
		}
		checked++
		return nil
	})
	if checked == 0 {
		t.Fatal("no extractable fixtures under test/testdata/email/merchants")
	}
}

type extractSnapshot struct {
	kind       emailtypes.DocumentKind
	supplier   string
	orderRef   string
	totalMinor int64
}

func pipelineExtract(
	clf *classify.Classifier,
	reg *extract.Registry,
	raw []byte,
	id string,
) (extractSnapshot, bool) {
	msg, err := mail.ParseRFC822(raw, id, "scrub-reextract")
	if err != nil {
		return extractSnapshot{}, false
	}
	kind := clf.Classify(msg)
	if !kind.ShouldExtract() {
		return extractSnapshot{}, false
	}
	ex, ok := reg.Extract(msg, kind)
	if !ok || ex.Supplier == "" {
		return extractSnapshot{}, false
	}
	return extractSnapshot{
		kind:       kind,
		supplier:   ex.Supplier,
		orderRef:   ex.OrderRef,
		totalMinor: ex.TotalMinor,
	}, true
}
