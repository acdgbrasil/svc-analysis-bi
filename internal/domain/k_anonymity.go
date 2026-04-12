package domain

// KThreshold is the minimum equivalence class size required by the
// anonymization policy (ADR-001 section 2.5).
const KThreshold = 5

// CheckKAnonymity iterates over the provided groups and marks any group
// whose Value is strictly less than k by setting BelowThreshold to true.
// Returns a new slice; the input slice is not mutated.
func CheckKAnonymity(groups []IndicatorGroup, k int) []IndicatorGroup {
	result := make([]IndicatorGroup, len(groups))
	for i, g := range groups {
		labels := make(map[string]string, len(g.Labels))
		for key, val := range g.Labels {
			labels[key] = val
		}
		result[i] = IndicatorGroup{
			Labels:         labels,
			Value:          g.Value,
			BelowThreshold: g.Value < k,
		}
	}
	return result
}

// CountSuppressed returns the number of groups that have BelowThreshold
// set to true.
func CountSuppressed(groups []IndicatorGroup) int {
	count := 0
	for _, g := range groups {
		if g.BelowThreshold {
			count++
		}
	}
	return count
}
