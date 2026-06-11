package mail

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/mail"
	"net/textproto"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jimbery/receipt/internal/emailtypes"
)

// MaildirSource reads messages from a Maildir layout (cur/, new/, tmp/).
type MaildirSource struct {
	root      string
	mailboxID string
}

// NewMaildirSource returns a MailSource rooted at path with the given mailbox ID.
func NewMaildirSource(path, mailboxID string) *MaildirSource {
	return &MaildirSource{root: path, mailboxID: mailboxID}
}

func (m *MaildirSource) ID() string { return m.mailboxID }

// Fetch returns all messages with Date >= since, sorted by ID for determinism.
func (m *MaildirSource) Fetch(ctx context.Context, since time.Time) ([]emailtypes.RawMessage, error) {
	_ = ctx
	var paths []string
	for _, sub := range []string{"cur", "new"} {
		dir := filepath.Join(m.root, sub)
		entries, err := os.ReadDir(dir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("maildir read %s: %w", dir, err)
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			paths = append(paths, filepath.Join(dir, e.Name()))
		}
	}
	sort.Strings(paths)

	out := make([]emailtypes.RawMessage, 0, len(paths))
	for _, p := range paths {
		msg, err := parseMaildirFile(p, m.mailboxID)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", p, err)
		}
		if !msg.Date.Before(since) {
			out = append(out, msg)
		}
	}
	return out, nil
}

// ParseRFC822 parses scrubbed fixture bytes into a RawMessage.
func ParseRFC822(data []byte, id, mailboxID string) (emailtypes.RawMessage, error) {
	parsed, err := mail.ReadMessage(strings.NewReader(string(data)))
	if err != nil {
		return emailtypes.RawMessage{}, err
	}
	headers := make(map[string]string)
	for k, v := range parsed.Header {
		if len(v) > 0 {
			headers[k] = v[0]
		}
	}
	date := time.Time{}
	if ds := headers["Date"]; ds != "" {
		if t, parseErr := mail.ParseDate(ds); parseErr == nil {
			date = t
		}
	}
	body, err := readBody(parsed)
	if err != nil {
		return emailtypes.RawMessage{}, err
	}
	if id == "" {
		id = "fixture"
	}
	return emailtypes.RawMessage{
		ID:          id,
		MailboxID:   mailboxID,
		From:        headers["From"],
		To:          splitAddresses(headers["To"]),
		Subject:     headers["Subject"],
		Date:        date,
		BodyText:    body.text,
		BodyHTML:    body.html,
		Headers:     headers,
		Attachments: body.attachments,
	}, nil
}

func parseMaildirFile(path, mailboxID string) (emailtypes.RawMessage, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return emailtypes.RawMessage{}, err
	}
	id := filepath.Base(path)
	if i := strings.Index(id, ":2,"); i >= 0 {
		id = id[:i]
	}
	return ParseRFC822(data, id, mailboxID)
}

type bodyParts struct {
	text        string
	html        string
	attachments []emailtypes.Attachment
}

func readBody(msg *mail.Message) (bodyParts, error) {
	ct := msg.Header.Get("Content-Type")
	mediaType, params, err := mime.ParseMediaType(ct)
	if err != nil || !strings.HasPrefix(mediaType, "multipart/") {
		raw, readErr := io.ReadAll(msg.Body)
		if readErr != nil {
			return bodyParts{}, readErr
		}
		content := decodePartContent(raw, textproto.MIMEHeader(msg.Header))
		if strings.Contains(strings.ToLower(ct), "html") {
			return bodyParts{html: content}, nil
		}
		return bodyParts{text: content}, nil
	}
	boundary := params["boundary"]
	if boundary == "" {
		return bodyParts{}, errors.New("multipart without boundary")
	}
	raw, readErr := io.ReadAll(msg.Body)
	if readErr != nil {
		return bodyParts{}, readErr
	}
	var out bodyParts
	if collectErr := collectBodyParts(multipart.NewReader(bytes.NewReader(raw), boundary), &out); collectErr != nil {
		return bodyParts{}, collectErr
	}
	return out, nil
}

func splitAddresses(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
