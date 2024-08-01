package utils

// ArrContainsStr checks if a string is in a string array
func ArrContainsStr(array []string, str string) bool {
	m := make(map[string]struct{}, len(array))
	for _, s := range array {
		m[s] = struct{}{}
	}
	_, exists := m[str]
	return exists
}
