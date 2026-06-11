package extract

import (
	"fmt"
	"strings"
	"time"
)

// FamilyKey returns a deduplication key preferring order reference.
func FamilyKey(supplier, orderRef string, issuedAt time.Time, totalMinor int64, currency string) string {
	if orderRef != "" {
		return strings.ToLower(supplier) + "|" + strings.ToUpper(orderRef)
	}
	date := issuedAt.Format("2006-01-02")
	return fmt.Sprintf("%s|%s|%d|%s", strings.ToLower(supplier), date, totalMinor, strings.ToUpper(currency))
}
