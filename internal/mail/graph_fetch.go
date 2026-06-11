package mail

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/jimbery/receipt/internal/emailtypes"
)

const defaultGraphBase = "https://graph.microsoft.com"

type graphListResponse struct {
	Value []struct {
		ID string `json:"id"`
	} `json:"value"`
}

func fetchGraph(
	ctx context.Context,
	client *http.Client,
	baseURL string,
	token PilotToken,
	mailboxID string,
	since time.Time,
) ([]emailtypes.RawMessage, error) {
	if client == nil {
		client = http.DefaultClient
	}
	if baseURL == "" {
		baseURL = defaultGraphBase
	}
	if token.expired(time.Now()) {
		return nil, ErrTokenExpired
	}

	ids, err := graphListIDs(ctx, client, baseURL, token, since)
	if err != nil {
		return nil, err
	}

	out := make([]emailtypes.RawMessage, 0, len(ids))
	for _, id := range ids {
		msg, getErr := graphGetMIME(ctx, client, baseURL, token, id, mailboxID)
		if getErr != nil {
			return nil, getErr
		}
		if !since.IsZero() && msg.Date.Before(since) {
			continue
		}
		out = append(out, msg)
	}
	return out, nil
}

func graphListIDs(
	ctx context.Context,
	client *http.Client,
	baseURL string,
	token PilotToken,
	since time.Time,
) ([]string, error) {
	q := url.Values{}
	q.Set("$select", "id")
	q.Set("$top", "100")
	if !since.IsZero() {
		q.Set("$filter", "receivedDateTime ge "+since.UTC().Format(time.RFC3339))
	}
	listURL := strings.TrimRight(baseURL, "/") + "/v1.0/me/messages?" + q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, listURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", token.authHeader())

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("graph list: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, graphAPIError(resp)
	}

	var parsed graphListResponse
	if decodeErr := json.NewDecoder(resp.Body).Decode(&parsed); decodeErr != nil {
		return nil, decodeErr
	}
	ids := make([]string, 0, len(parsed.Value))
	for _, m := range parsed.Value {
		if m.ID != "" {
			ids = append(ids, m.ID)
		}
	}
	return ids, nil
}

func graphGetMIME(
	ctx context.Context,
	client *http.Client,
	baseURL string,
	token PilotToken,
	id, mailboxID string,
) (emailtypes.RawMessage, error) {
	getURL := strings.TrimRight(baseURL, "/") + "/v1.0/me/messages/" + url.PathEscape(id) + "/$value"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, getURL, nil)
	if err != nil {
		return emailtypes.RawMessage{}, err
	}
	req.Header.Set("Authorization", token.authHeader())

	resp, err := client.Do(req)
	if err != nil {
		return emailtypes.RawMessage{}, fmt.Errorf("graph get %s: %w", id, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return emailtypes.RawMessage{}, graphAPIError(resp)
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return emailtypes.RawMessage{}, err
	}
	return ParseRFC822(raw, id, mailboxID)
}

func graphAPIError(resp *http.Response) error {
	body, _ := io.ReadAll(io.LimitReader(resp.Body, apiErrorBodyLimit))
	return fmt.Errorf("graph api %s: %s", resp.Status, strings.TrimSpace(string(body)))
}
