package mail_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jimbery/receipt/internal/mail"
)

func TestGmailSource_Fetch_MockAPI(t *testing.T) {
	rawMIME := strings.Join([]string{
		"From: orders@screwfix.com",
		"To: volunteer@example.com",
		"Subject: Your Screwfix receipt",
		"Date: Mon, 10 Jun 2026 12:00:00 +0000",
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=utf-8",
		"",
		"Order total £12.99",
		"Order #ABC-123",
	}, "\r\n")
	encoded := base64.URLEncoding.EncodeToString([]byte(rawMIME))

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/messages") && !strings.Contains(r.URL.Path, "/messages/"):
			_, _ = w.Write([]byte(`{"messages":[{"id":"msg-1"}]}`))
		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/messages/msg-1"):
			_ = json.NewEncoder(w).Encode(map[string]string{"id": "msg-1", "raw": encoded})
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	tokenPath := writePilotToken(t, time.Now().Add(time.Hour))
	src := mail.NewGmailSourceWithHTTP("pilot-gmail", tokenPath, mail.PilotHTTPConfig{
		Client:  srv.Client(),
		BaseURL: srv.URL,
	})

	msgs, err := src.Fetch(context.Background(), time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if len(msgs) != 1 {
		t.Fatalf("got %d messages, want 1", len(msgs))
	}
	if msgs[0].ID != "msg-1" {
		t.Fatalf("id %q", msgs[0].ID)
	}
	if !strings.Contains(msgs[0].Subject, "Screwfix") {
		t.Fatalf("subject %q", msgs[0].Subject)
	}
}

func TestGmailSource_Fetch_ExpiredToken(t *testing.T) {
	tokenPath := writePilotToken(t, time.Now().Add(-time.Hour))
	src := mail.NewGmailSource("pilot-gmail", tokenPath)
	_, err := src.Fetch(context.Background(), time.Time{})
	if !errors.Is(err, mail.ErrTokenExpired) {
		t.Fatalf("got %v, want ErrTokenExpired", err)
	}
}

func TestGraphSource_Fetch_MockAPI(t *testing.T) {
	rawMIME := strings.Join([]string{
		"From: orders@toolstation.com",
		"To: volunteer@example.com",
		"Subject: Toolstation order confirmed",
		"Date: Tue, 10 Jun 2026 09:00:00 +0000",
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=utf-8",
		"",
		"Order total £45.00",
		"Order YWW123456789",
	}, "\r\n")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/messages") && !strings.Contains(r.URL.Path, "/messages/"):
			_, _ = w.Write([]byte(`{"value":[{"id":"graph-msg-1"}]}`))
		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/messages/graph-msg-1/$value"):
			_, _ = w.Write([]byte(rawMIME))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	tokenPath := writePilotToken(t, time.Now().Add(time.Hour))
	src := mail.NewGraphSourceWithHTTP("pilot-graph", tokenPath, mail.PilotHTTPConfig{
		Client:  srv.Client(),
		BaseURL: srv.URL,
	})

	msgs, err := src.Fetch(context.Background(), time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if len(msgs) != 1 {
		t.Fatalf("got %d messages, want 1", len(msgs))
	}
	if msgs[0].ID != "graph-msg-1" {
		t.Fatalf("id %q", msgs[0].ID)
	}
	if !strings.Contains(msgs[0].From, "toolstation") {
		t.Fatalf("from %q", msgs[0].From)
	}
}

func TestGraphSource_Fetch_ExpiredToken(t *testing.T) {
	tokenPath := writePilotToken(t, time.Now().Add(-time.Hour))
	src := mail.NewGraphSource("pilot-graph", tokenPath)
	_, err := src.Fetch(context.Background(), time.Time{})
	if !errors.Is(err, mail.ErrTokenExpired) {
		t.Fatalf("got %v, want ErrTokenExpired", err)
	}
}

func TestLoadPilotToken_MissingAccessToken(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(path, []byte(`{"token_type":"Bearer"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := mail.LoadPilotToken(path); err == nil {
		t.Fatal("expected error for missing access_token")
	}
}

func writePilotToken(t *testing.T, expiry time.Time) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "token.json")
	payload, err := json.Marshal(struct {
		AccessToken string    `json:"access_token"`
		TokenType   string    `json:"token_type"`
		Expiry      time.Time `json:"expiry"`
	}{
		AccessToken: "test-access-token",
		TokenType:   "Bearer",
		Expiry:      expiry,
	})
	if err != nil {
		t.Fatal(err)
	}
	if writeErr := os.WriteFile(path, payload, 0o600); writeErr != nil {
		t.Fatal(writeErr)
	}
	return path
}
