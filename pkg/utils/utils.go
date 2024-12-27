package utils

// SliceContains checks if a slice contains a specific element
func SliceContains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}
