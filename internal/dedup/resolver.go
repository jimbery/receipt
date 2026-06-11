package dedup

import (
	"fmt"
	"sort"
	"strings"

	"github.com/jimbery/receipt/internal/emailtypes"
)

// DocumentPrecedence ranks document kinds for family collapse.
func DocumentPrecedence(kind emailtypes.DocumentKind) int {
	switch kind {
	case emailtypes.KindPurchaseReceipt:
		return 3
	case emailtypes.KindCreditNote:
		return 2
	case emailtypes.KindDispatchNotice:
		return 1
	case emailtypes.KindMarketing, emailtypes.KindStatement, emailtypes.KindUnknown:
		return 0
	}
	return 0
}

// ExtractedWithProvenance pairs an extraction with source metadata.
type ExtractedWithProvenance struct {
	Extracted emailtypes.ExtractedReceipt
	MessageID string
	Kind      emailtypes.DocumentKind
}

// ResolveFamilies collapses purchase families per ADR-002 D4.
func ResolveFamilies(items []ExtractedWithProvenance) []emailtypes.ExtractedReceipt {
	byFamily := make(map[string][]ExtractedWithProvenance)
	for _, it := range items {
		key := it.Extracted.FamilyKey
		if key == "" {
			key = "singleton:" + it.MessageID
		}
		byFamily[key] = append(byFamily[key], it)
	}

	var families []string
	for k := range byFamily {
		families = append(families, k)
	}
	sort.Strings(families)

	out := make([]emailtypes.ExtractedReceipt, 0, len(items))
	for _, key := range families {
		group := byFamily[key]
		out = append(out, resolveGroup(group)...)
	}
	return out
}

func resolveGroup(group []ExtractedWithProvenance) []emailtypes.ExtractedReceipt {
	if len(group) == 1 {
		return []emailtypes.ExtractedReceipt{group[0].Extracted}
	}

	// Credit notes never dropped — always pass through alongside purchase docs.
	var credits []emailtypes.ExtractedReceipt
	var purchasable []ExtractedWithProvenance
	for _, g := range group {
		if g.Kind == emailtypes.KindCreditNote {
			credits = append(credits, g.Extracted)
			continue
		}
		purchasable = append(purchasable, g)
	}

	if len(purchasable) == 0 {
		return credits
	}

	if ambiguousFamily(purchasable) {
		var all []emailtypes.ExtractedReceipt
		for _, g := range purchasable {
			all = append(all, g.Extracted)
		}
		all = append(all, credits...)
		return all
	}

	// Amazon shipment-level: distinct shipment suffixes stay separate.
	if hasShipmentSplit(purchasable) {
		var out []emailtypes.ExtractedReceipt
		for _, g := range purchasable {
			out = append(out, g.Extracted)
		}
		out = append(out, credits...)
		return out
	}

	// Same message forwarded twice -> one receipt (same MessageID).
	byMsg := make(map[string][]ExtractedWithProvenance)
	for _, g := range purchasable {
		byMsg[g.MessageID] = append(byMsg[g.MessageID], g)
	}
	if len(byMsg) == 1 && len(purchasable) > 1 {
		return []emailtypes.ExtractedReceipt{collapseToBest(purchasable)}
	}

	// Invoice beats confirmation for same order ref.
	out := []emailtypes.ExtractedReceipt{collapseToBest(purchasable)}
	out = append(out, credits...)
	return out
}

func collapseToBest(group []ExtractedWithProvenance) emailtypes.ExtractedReceipt {
	best := pickBest(group)
	best.MergedFrom = mergedMessageIDs(group, best.SourceMessageID)
	return best
}

func mergedMessageIDs(group []ExtractedWithProvenance, winnerID string) []string {
	ids := make([]string, 0, len(group)-1)
	for _, g := range group {
		if g.MessageID != winnerID {
			ids = append(ids, g.MessageID)
		}
	}
	sort.Strings(ids)
	return ids
}

func pickBest(group []ExtractedWithProvenance) emailtypes.ExtractedReceipt {
	best := group[0]
	bestScore := DocumentPrecedence(best.Kind)
	bestGrade := gradeRank(best.Extracted.Grade)
	for _, g := range group[1:] {
		score := DocumentPrecedence(g.Kind)
		gr := gradeRank(g.Extracted.Grade)
		if score > bestScore || (score == bestScore && gr > bestGrade) {
			best = g
			bestScore = score
			bestGrade = gr
		}
	}
	return best.Extracted
}

func gradeRank(g emailtypes.CompletenessGrade) int {
	switch g {
	case emailtypes.GradeItemised:
		return 4
	case emailtypes.GradePartial:
		return 3
	case emailtypes.GradeEnvelope:
		return 2
	case emailtypes.GradeRequiresOCR:
		return 1
	}
	return 1
}

func ambiguousFamily(group []ExtractedWithProvenance) bool {
	if len(group) < 2 {
		return false
	}
	noRefs := true
	for _, g := range group {
		if g.Extracted.OrderRef != "" {
			noRefs = false
			break
		}
	}
	if !noRefs {
		return false
	}
	// Same supplier+day+total fingerprint but distinct messages -> ambiguous.
	fp := fingerprint(group[0].Extracted)
	for _, g := range group[1:] {
		if fingerprint(g.Extracted) != fp {
			return false
		}
	}
	return true
}

func fingerprint(e emailtypes.ExtractedReceipt) string {
	return fmt.Sprintf("%s|%s|%d", strings.ToLower(e.Supplier), e.IssuedAt.Format("2006-01-02"), e.TotalMinor)
}

func hasShipmentSplit(group []ExtractedWithProvenance) bool {
	count := 0
	for _, g := range group {
		if strings.Contains(g.Extracted.OrderRef, "-SHIP-") {
			count++
		}
	}
	return count >= 2
}
