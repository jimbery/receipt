package mail

import (
	"context"
	"net/http"
	"time"

	"github.com/jimbery/receipt/internal/emailtypes"
)

// GraphSource fetches messages via Microsoft Graph using a local pilot OAuth token.
type GraphSource struct {
	mailboxID string
	tokenPath string
	client    *http.Client
	baseURL   string
}

// NewGraphSource returns a Graph mail source (requires tokenPath at runtime).
func NewGraphSource(mailboxID, tokenPath string) *GraphSource {
	return &GraphSource{mailboxID: mailboxID, tokenPath: tokenPath}
}

func (g *GraphSource) ID() string { return g.mailboxID }

// Fetch returns messages received on or after since.
func (g *GraphSource) Fetch(ctx context.Context, since time.Time) ([]emailtypes.RawMessage, error) {
	if g.tokenPath == "" {
		return nil, ErrNotConfigured
	}
	token, err := LoadPilotToken(g.tokenPath)
	if err != nil {
		return nil, err
	}
	return fetchGraph(ctx, g.client, g.baseURL, token, g.mailboxID, since)
}
