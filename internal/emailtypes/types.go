package emailtypes

import (
	"time"

	"github.com/jimbery/receipt/internal/model"
)

// CompletenessGrade describes how much purchase detail was extracted.
type CompletenessGrade string

const (
	GradeItemised    CompletenessGrade = "itemised"
	GradePartial     CompletenessGrade = "partial"
	GradeEnvelope    CompletenessGrade = "envelope"
	GradeRequiresOCR CompletenessGrade = "requires_ocr"
)

// DocumentKind is the classifier output before extraction.
type DocumentKind string

const (
	KindPurchaseReceipt DocumentKind = "purchase_receipt"
	KindCreditNote      DocumentKind = "credit_note"
	KindDispatchNotice  DocumentKind = "dispatch_notice"
	KindMarketing       DocumentKind = "marketing"
	KindStatement       DocumentKind = "statement"
	KindUnknown         DocumentKind = "unknown"
)

// ShouldExtract reports whether the pipeline should run extractors.
func (k DocumentKind) ShouldExtract() bool {
	switch k {
	case KindPurchaseReceipt, KindCreditNote, KindDispatchNotice, KindUnknown:
		return true
	case KindMarketing, KindStatement:
		return false
	}
	return false
}

// Provenance records where a field value came from.
type Provenance string

const (
	ProvSchemaOrg Provenance = "schema_org"
	ProvHTML      Provenance = "html"
	ProvPDFText   Provenance = "pdf_text"
	ProvSubject   Provenance = "subject"
	ProvHeader    Provenance = "header"
)

// RawMessage is a normalised email before classification.
type RawMessage struct {
	ID          string            `json:"id"`
	MailboxID   string            `json:"mailbox_id,omitempty"`
	From        string            `json:"from"`
	To          []string          `json:"to,omitempty"`
	Subject     string            `json:"subject"`
	Date        time.Time         `json:"date"`
	BodyHTML    string            `json:"body_html,omitempty"`
	BodyText    string            `json:"body_text,omitempty"`
	Attachments []Attachment      `json:"attachments,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
}

// Attachment is metadata for a MIME part.
type Attachment struct {
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	Content     []byte `json:"content,omitempty"`
}

// ExtractedReceipt is the intermediate representation before canonicalisation.
type ExtractedReceipt struct {
	SourceMessageID string
	Supplier        string
	OrderRef        string
	FamilyKey       string
	Grade           CompletenessGrade
	DocumentKind    DocumentKind
	TotalMinor      int64
	Currency        string
	VATMinor        int64
	VATExempt       bool
	IssuedAt        time.Time
	LineItems       []model.LineItem
	Category        string
	FieldProvenance map[string]Provenance
	AttachmentRefs  []string
	// MergedFrom lists source message IDs collapsed into this receipt (ADR-002 D4).
	MergedFrom []string
}
