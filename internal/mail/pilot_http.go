package mail

import "net/http"

const apiErrorBodyLimit = 4096

// PilotHTTPConfig overrides the HTTP client and API base URL (used by tests and local mocks).
type PilotHTTPConfig struct {
	Client  *http.Client
	BaseURL string
}

func applyPilotHTTP(src *GmailSource, cfg PilotHTTPConfig) *GmailSource {
	if cfg.Client != nil {
		src.client = cfg.Client
	}
	if cfg.BaseURL != "" {
		src.baseURL = cfg.BaseURL
	}
	return src
}

func applyGraphPilotHTTP(src *GraphSource, cfg PilotHTTPConfig) *GraphSource {
	if cfg.Client != nil {
		src.client = cfg.Client
	}
	if cfg.BaseURL != "" {
		src.baseURL = cfg.BaseURL
	}
	return src
}

// NewGmailSourceWithHTTP returns a Gmail source with optional HTTP overrides.
func NewGmailSourceWithHTTP(mailboxID, tokenPath string, cfg PilotHTTPConfig) *GmailSource {
	return applyPilotHTTP(NewGmailSource(mailboxID, tokenPath), cfg)
}

// NewGraphSourceWithHTTP returns a Graph source with optional HTTP overrides.
func NewGraphSourceWithHTTP(mailboxID, tokenPath string, cfg PilotHTTPConfig) *GraphSource {
	return applyGraphPilotHTTP(NewGraphSource(mailboxID, tokenPath), cfg)
}
