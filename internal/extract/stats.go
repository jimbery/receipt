package extract

import "github.com/jimbery/receipt/internal/emailtypes"

// CountRequiresOCR returns extracted messages routed to requires_ocr.
func CountRequiresOCR(extracted []emailtypes.ExtractedReceipt) int {
	n := 0
	for _, e := range extracted {
		if Grade(e) == emailtypes.GradeRequiresOCR {
			n++
		}
	}
	return n
}
