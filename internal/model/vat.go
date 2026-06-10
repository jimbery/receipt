package model

// VATLine is per-line VAT detail on a receipt (unused for matching in Phase 0).
type VATLine struct {
	Rate      float64 // e.g. 0.20 for 20%
	NetAmount Money
	VATAmount Money
}
