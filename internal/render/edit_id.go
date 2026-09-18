package render

import "strings"

func normalizeEditID(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}
