package mail

import (
	"bytes"
	"encoding/base64"
	"io"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net/textproto"
	"strings"

	"github.com/jimbery/receipt/internal/emailtypes"
)

func headerValue(headers textproto.MIMEHeader, key string) string {
	if v := headers.Get(key); v != "" {
		return v
	}
	return ""
}

func decodePartContent(raw []byte, headers textproto.MIMEHeader) string {
	enc := strings.ToLower(strings.TrimSpace(headerValue(headers, "Content-Transfer-Encoding")))
	switch enc {
	case "quoted-printable":
		decoded, err := io.ReadAll(quotedprintable.NewReader(bytes.NewReader(raw)))
		if err != nil {
			return string(raw)
		}
		return string(decoded)
	case "base64":
		decoded := make([]byte, base64.StdEncoding.DecodedLen(len(raw)))
		n, err := base64.StdEncoding.Decode(decoded, raw)
		if err != nil {
			return string(raw)
		}
		return string(decoded[:n])
	default:
		return string(raw)
	}
}

func nextMultipartPart(reader *multipart.Reader) *multipart.Part {
	part, err := reader.NextPart()
	if err != nil {
		return nil
	}
	return part
}

func collectBodyParts(reader *multipart.Reader, out *bodyParts) error {
	for part := nextMultipartPart(reader); part != nil; part = nextMultipartPart(reader) {
		pct, pparams, _ := mime.ParseMediaType(part.Header.Get("Content-Type"))
		disp := part.Header.Get("Content-Disposition")
		if strings.HasPrefix(pct, "multipart/") {
			boundary := pparams["boundary"]
			if boundary != "" {
				if err := collectBodyParts(multipart.NewReader(part, boundary), out); err != nil {
					return err
				}
			}
			continue
		}
		raw, readErr := io.ReadAll(part)
		if readErr != nil {
			return readErr
		}
		if strings.Contains(disp, "attachment") || pparams["name"] != "" {
			name := pparams["name"]
			if _, dispParams, parseErr := mime.ParseMediaType(disp); parseErr == nil {
				if n := dispParams["filename"]; n != "" {
					name = n
				}
			}
			out.attachments = append(out.attachments, emailtypes.Attachment{
				Filename:    name,
				ContentType: pct,
				Content:     raw,
			})
			continue
		}
		content := decodePartContent(raw, part.Header)
		switch {
		case strings.HasPrefix(pct, "text/html"):
			if out.html == "" || len(content) > len(out.html) {
				out.html = content
			}
		case strings.HasPrefix(pct, "text/plain"):
			if out.text == "" || len(content) > len(out.text) {
				out.text = content
			}
		}
	}
	return nil
}
