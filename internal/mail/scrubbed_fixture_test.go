package mail_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jimbery/receipt/internal/classify"
	"github.com/jimbery/receipt/internal/emailtypes"
	"github.com/jimbery/receipt/internal/extract"
	"github.com/jimbery/receipt/internal/mail"
)

func TestScrubbedEMLFixtures_ParseAndExtract(t *testing.T) {
	root := filepath.Join("..", "..", "test", "testdata", "email", "merchants")
	clf := classify.New()
	reg := extract.NewRegistry()

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
		msg, parseErr := mail.ParseRFC822(data, info.Name(), "fixture-lab")
		if parseErr != nil {
			t.Errorf("%s: parse: %v", path, parseErr)
			return nil
		}
		kind := clf.Classify(msg)
		if kind != emailtypes.KindPurchaseReceipt {
			t.Errorf("%s: kind %s want purchase_receipt", path, kind)
			return nil
		}
		ex, ok := reg.Extract(msg, kind)
		if !ok {
			t.Errorf("%s: extract failed", path)
			return nil
		}
		if ex.Supplier == "" {
			t.Errorf("%s: missing supplier", path)
		}
		return nil
	})
}
