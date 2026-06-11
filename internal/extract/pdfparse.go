package extract

import (
	"bytes"
	"compress/flate"
	"io"
	"regexp"
	"strings"
)

var (
	pdfLiteralTextRe = regexp.MustCompile(`\(([^)\\]*(?:\\.[^)\\]*)*)\)\s*Tj`)
)

// ExtractPDFText returns machine-readable text from a text-layer PDF.
// Image-only or encrypted PDFs return empty string (caller routes to requires_ocr).
func ExtractPDFText(data []byte) string {
	if len(data) < 5 || !bytes.HasPrefix(data, []byte("%PDF")) {
		return ""
	}
	if bytes.Contains(data, []byte("/Encrypt")) {
		return ""
	}
	var out strings.Builder
	out.WriteString(extractPDFLiteralOperators(data))
	for _, chunk := range decompressPDFStreams(data) {
		out.WriteString(extractPDFLiteralOperators(chunk))
		_, _ = out.Write(chunk)
	}
	return out.String()
}

func extractPDFLiteralOperators(data []byte) string {
	var b strings.Builder
	for _, m := range pdfLiteralTextRe.FindAllStringSubmatch(string(data), -1) {
		if len(m) > 1 {
			b.WriteString(unescapePDFLiteral(m[1]))
			b.WriteByte(' ')
		}
	}
	return b.String()
}

func unescapePDFLiteral(s string) string {
	s = strings.ReplaceAll(s, `\n`, "\n")
	s = strings.ReplaceAll(s, `\r`, "\r")
	s = strings.ReplaceAll(s, `\t`, "\t")
	s = strings.ReplaceAll(s, `\(`, "(")
	s = strings.ReplaceAll(s, `\)`, ")")
	s = strings.ReplaceAll(s, `\\`, `\`)
	return s
}

func decompressPDFStreams(data []byte) [][]byte {
	var out [][]byte
	needle := []byte("stream")
	for {
		idx := bytes.Index(data, needle)
		if idx < 0 {
			break
		}
		start := idx + len(needle)
		if start < len(data) && data[start] == '\r' {
			start++
		}
		if start < len(data) && data[start] == '\n' {
			start++
		}
		end := bytes.Index(data[start:], []byte("endstream"))
		if end < 0 {
			break
		}
		chunk := data[start : start+end]
		if inflated, ok := flateInflate(chunk); ok {
			out = append(out, inflated)
		}
		data = data[start+end:]
	}
	return out
}

func flateInflate(data []byte) ([]byte, bool) {
	r := flate.NewReader(bytes.NewReader(data))
	defer r.Close()
	inflated, err := io.ReadAll(r)
	if err != nil {
		return nil, false
	}
	return inflated, true
}

// ParsePDFTotal scans extracted PDF text for a total amount.
func ParsePDFTotal(text string) (int64, bool) {
	lower := strings.ToLower(text)
	idx := strings.Index(lower, "total")
	if idx < 0 {
		return parseMoneyToken(text)
	}
	segment := text[idx:]
	if minor, ok := parseMoneyToken(segment); ok {
		return minor, true
	}
	return parseMoneyToken(text)
}
