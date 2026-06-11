package mail_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jimbery/receipt/internal/mail"
)

func TestPilotSources_NotConfigured(t *testing.T) {
	sources := []mail.MailSource{
		mail.NewGmailSource("g1", ""),
		mail.NewGraphSource("m1", ""),
	}
	for _, s := range sources {
		_, err := s.Fetch(context.Background(), time.Time{})
		if !errors.Is(err, mail.ErrNotConfigured) {
			t.Fatalf("%s: got %v, want ErrNotConfigured", s.ID(), err)
		}
	}
}

func TestMailSource_InterfaceConformance(t *testing.T) {
	sources := []mail.MailSource{
		mail.NewMaildirSource(t.TempDir(), "md1"),
		mail.NewGmailSource("g1", ""),
		mail.NewGraphSource("m1", ""),
	}
	for _, s := range sources {
		if s.ID() == "" {
			t.Fatalf("empty ID from %T", s)
		}
	}
}
