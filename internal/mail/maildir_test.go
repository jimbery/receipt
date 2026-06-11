package mail_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/jimbery/receipt/internal/mail"
)

func TestMaildirSource_RoundTrip(t *testing.T) {
	root := filepath.Join("..", "..", "test", "testdata", "email", "maildir")
	src := mail.NewMaildirSource(root, "volunteer-1")
	msgs, err := src.Fetch(context.Background(), time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if len(msgs) != 1 {
		t.Fatalf("got %d messages, want 1", len(msgs))
	}
	m := msgs[0]
	if m.From != "orders@amazon.co.uk" {
		t.Fatalf("from: %q", m.From)
	}
	if m.Subject == "" {
		t.Fatal("expected subject")
	}
	if m.BodyText == "" {
		t.Fatal("expected body")
	}
}
