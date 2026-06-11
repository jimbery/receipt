package mail_test

import (
	stdmail "net/mail"
	"os"
	"path/filepath"
	"strings"
	"testing"

	receiptmail "github.com/jimbery/receipt/internal/mail"
)

func FuzzParseMaildirMessage_NoPanic(f *testing.F) {
	f.Add("From: a@b.com\nSubject: Hi\nDate: Mon, 10 Mar 2026 10:00:00 +0000\n\nbody")
	f.Add("not valid mail at all {{{")
	f.Fuzz(func(t *testing.T, raw string) {
		if len(raw) > 64*1024 {
			return
		}
		_, _ = stdmail.ReadMessage(strings.NewReader(raw))
	})
}

func FuzzParseRFC822_NoPanic(f *testing.F) {
	plain := "From: orders@screwfix.com\r\nSubject: Receipt\r\n" +
		"Date: Mon, 10 Jun 2026 12:00:00 +0000\r\nMIME-Version: 1.0\r\n" +
		"Content-Type: text/plain; charset=utf-8\r\n\r\nOrder total £12.99\r\n"
	f.Add([]byte(plain))
	multipart := "From: a@b.com\r\nSubject: multipart\r\n" +
		"Content-Type: multipart/mixed; boundary=----bb\r\n\r\n------bb\r\n" +
		"Content-Type: text/plain\r\n\r\npart one\r\n------bb--\r\n"
	f.Add([]byte(multipart))
	f.Add([]byte("not valid mime {{{"))
	seedRFC822FromFixtures(f)
	f.Fuzz(func(t *testing.T, raw []byte) {
		if len(raw) > 128*1024 {
			return
		}
		_, _ = receiptmail.ParseRFC822(raw, "fuzz-id", "fuzz-mailbox")
	})
}

func seedRFC822FromFixtures(f *testing.F) {
	root := filepath.Join("..", "..", "test", "testdata", "email", "merchants")
	err := filepath.Walk(root, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.IsDir() || !strings.HasSuffix(strings.ToLower(info.Name()), ".eml") {
			return nil
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		if len(data) > 32*1024 {
			data = data[:32*1024]
		}
		f.Add(data)
		return nil
	})
	if err != nil {
		f.Fatalf("seed fixtures: %v", err)
	}
}
