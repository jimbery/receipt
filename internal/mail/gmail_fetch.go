package mail

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/jimbery/receipt/internal/emailtypes"
)

const defaultGmailBase = "https://gmail.googleapis.com"

type gmailListResponse struct {
	Messages []struct {
		ID string `json:"id"`
	} `json:"messages"`
}

type gmailGetResponse struct {
	ID      string `json:"id"`
	Raw     string `json:"raw"`
	Snippet string `json:"snippet"`
}

func fetchGmail(
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
		baseURL = defaultGmailBase
	}
	now := time.Now()
	if token.expired(now) {
		return nil, ErrTokenExpired
	}

	ids, err := gmailListIDs(ctx, client, baseURL, token, since)
	if err != nil {
		return nil, err
	}

	out := make([]emailtypes.RawMessage, 0, len(ids))
	for _, id := range ids {
		msg, getErr := gmailGetRaw(ctx, client, baseURL, token, id, mailboxID)
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

func gmailListIDs(
	ctx context.Context,
	client *http.Client,
	baseURL string,
	token PilotToken,
	since time.Time,
) ([]string, error) {
	q := url.Values{}
	q.Set("maxResults", "100")
	if !since.IsZero() {
		q.Set("q", "after:"+since.Format("2006/01/02"))
	}
	listURL := strings.TrimRight(baseURL, "/") + "/gmail/v1/users/me/messages?" + q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, listURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", token.authHeader())

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gmail list: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, gmailAPIError(resp)
	}

	var parsed gmailListResponse
	if decodeErr := json.NewDecoder(resp.Body).Decode(&parsed); decodeErr != nil {
		return nil, decodeErr
	}
	ids := make([]string, 0, len(parsed.Messages))
	for _, m := range parsed.Messages {
		if m.ID != "" {
			ids = append(ids, m.ID)
		}
	}
	return ids, nil
}

func gmailGetRaw(
	ctx context.Context,
	client *http.Client,
	baseURL string,
	token PilotToken,
	id, mailboxID string,
) (emailtypes.RawMessage, error) {
	getURL := strings.TrimRight(baseURL, "/") + "/gmail/v1/users/me/messages/" + url.PathEscape(id) + "?format=raw"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, getURL, nil)
	if err != nil {
		return emailtypes.RawMessage{}, err
	}
	req.Header.Set("Authorization", token.authHeader())

	resp, err := client.Do(req)
	if err != nil {
		return emailtypes.RawMessage{}, fmt.Errorf("gmail get %s: %w", id, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return emailtypes.RawMessage{}, gmailAPIError(resp)
	}

	var parsed gmailGetResponse
	if decodeErr := json.NewDecoder(resp.Body).Decode(&parsed); decodeErr != nil {
		return emailtypes.RawMessage{}, decodeErr
	}
	raw, decErr := base64.URLEncoding.DecodeString(parsed.Raw)
	if decErr != nil {
		raw, decErr = base64.RawURLEncoding.DecodeString(parsed.Raw)
		if decErr != nil {
			return emailtypes.RawMessage{}, fmt.Errorf("gmail raw decode: %w", decErr)
		}
	}
	return ParseRFC822(raw, id, mailboxID)
}

func gmailAPIError(resp *http.Response) error {
	body, _ := io.ReadAll(io.LimitReader(resp.Body, apiErrorBodyLimit))
	return fmt.Errorf("gmail api %s: %s", resp.Status, strings.TrimSpace(string(body)))
}
