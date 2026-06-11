package mail_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jimbery/receipt/internal/mail"
)

func TestParseScrubbedScrewfixA162(t *testing.T) {
	path := filepath.Join("..", "..", "test", "testdata", "email", "merchants", "screwfix",
		"confirmation-of-your-order-order.eml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	msg, err := mail.ParseRFC822(data, "a162", "t")
	if err != nil {
		t.Fatal(err)
	}
	if msg.BodyText == "" && msg.BodyHTML == "" {
		t.Fatal("expected decoded text or html body")
	}
	t.Logf("text=%d html=%d", len(msg.BodyText), len(msg.BodyHTML))
}
