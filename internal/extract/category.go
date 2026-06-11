package extract

// SA category labels for MTD sole-trader reporting (Phase 1 rule table).
const (
	SACategoryMaterials     = "materials"
	SACategoryMotor         = "motor"
	SACategoryOverheads     = "overheads"
	SACategoryUncategorised = "uncategorised"
)

func saCategoryByMerchant() map[string]string {
	return map[string]string{
		"Amazon":      SACategoryMaterials,
		"Screwfix":    SACategoryMaterials,
		"Toolstation": SACategoryMaterials,
		"B&Q":         SACategoryMaterials,
		"Shell":       SACategoryMotor,
		"BP":          SACategoryMotor,
		"Esso":        SACategoryMotor,
		"Octopus":     SACategoryOverheads,
		"EDF":         SACategoryOverheads,
	}
}

// SACategoryForMerchant returns the SA category for a known merchant name.
func SACategoryForMerchant(merchant string) (string, bool) {
	cat, ok := saCategoryByMerchant()[merchant]
	return cat, ok
}
