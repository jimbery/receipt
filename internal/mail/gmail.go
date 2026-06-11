package mail

import (
	"context"
	"net/http"
	"time"

	"github.com/jimbery/receipt/internal/emailtypes"
)

// GmailSource fetches messages via the Gmail API using a local pilot OAuth token.
type GmailSource struct {
	mailboxID string
	tokenPath string
	client    *http.Client
	baseURL   string
}

// NewGmailSource returns a Gmail mail source (requires tokenPath at runtime).
func NewGmailSource(mailboxID, tokenPath string) *GmailSource {
	return &GmailSource{mailboxID: mailboxID, tokenPath: tokenPath}
}

func (g *GmailSource) ID() string { return g.mailboxID }

// Fetch returns messages received on or after since.
func (g *GmailSource) Fetch(ctx context.Context, since time.Time) ([]emailtypes.RawMessage, error) {
	if g.tokenPath == "" {
		return nil, ErrNotConfigured
	}
	token, err := LoadPilotToken(g.tokenPath)
	if err != nil {
		return nil, err
	}
	return fetchGmail(ctx, g.client, g.baseURL, token, g.mailboxID, since)
}
