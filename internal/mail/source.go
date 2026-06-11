package mail

import (
	"context"
	"errors"
	"time"

	"github.com/jimbery/receipt/internal/emailtypes"
)

// ErrNotConfigured indicates a pilot source lacks OAuth credentials.
var ErrNotConfigured = errors.New("mail source not configured")

// MailSource fetches raw messages from a mailbox.
type MailSource interface {
	ID() string
	Fetch(ctx context.Context, since time.Time) ([]emailtypes.RawMessage, error)
}

var (
	_ MailSource = (*MaildirSource)(nil)
	_ MailSource = (*GmailSource)(nil)
	_ MailSource = (*GraphSource)(nil)
)
