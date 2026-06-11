package extract

import (
	"strings"

	"github.com/jimbery/receipt/internal/emailtypes"
)

// Extractor extracts structured receipt data from a raw message.
type Extractor interface {
	Name() string
	Matches(msg emailtypes.RawMessage) bool
	Extract(msg emailtypes.RawMessage) (emailtypes.ExtractedReceipt, error)
}

// Registry holds merchant-specific and generic extractors.
type Registry struct {
	extractors []Extractor
	version    string
}

// NewRegistry returns the default extractor set for Phase 1.
func NewRegistry() *Registry {
	return NewRegistryWith(nil)
}

// NewRegistryWith builds a registry, optionally prepending extractors before the generic fallback.
func NewRegistryWith(extra []Extractor) *Registry {
	generic := &GenericExtractor{}
	extractors := make([]Extractor, 0, len(extra)+7)
	extractors = append(extractors, extra...)
	extractors = append(extractors,
		&AmazonExtractor{},
		&ScrewfixExtractor{},
		&ToolstationExtractor{},
		&BandQExtractor{},
		&FuelExtractor{},
		&UtilityExtractor{},
		generic,
	)
	return &Registry{
		version:    "extract-v1-20260610",
		extractors: extractors,
	}
}

// Version identifies the committed extractor set.
func (r *Registry) Version() string { return r.version }

// Extract runs the first matching extractor.
func (r *Registry) Extract(
	msg emailtypes.RawMessage,
	kind emailtypes.DocumentKind,
) (emailtypes.ExtractedReceipt, bool) {
	if !kind.ShouldExtract() {
		return emailtypes.ExtractedReceipt{}, false
	}
	for _, ex := range r.extractors {
		if ex.Matches(msg) {
			out, err := ex.Extract(msg)
			if err != nil {
				continue
			}
			out.SourceMessageID = msg.ID
			out.DocumentKind = kind
			out.Grade = Grade(out)
			return out, true
		}
	}
	return emailtypes.ExtractedReceipt{}, false
}

// SupplierFromFrom extracts a display supplier from the From header.
func SupplierFromFrom(from string) string {
	from = strings.ToLower(from)
	at := strings.LastIndex(from, "@")
	if at < 0 {
		return strings.TrimSpace(from)
	}
	domain := from[at+1:]
	domain = strings.TrimSuffix(domain, ">")
	parts := strings.Split(domain, ".")
	if len(parts) >= 2 {
		p := parts[len(parts)-2]
		if p == "" {
			return domain
		}
		return strings.ToUpper(p[:1]) + p[1:]
	}
	return domain
}
