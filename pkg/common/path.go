package common

import "strings"

// sanitizePath removes a leading "./" from the path if present.
func SanitizePath(path string) string {
	return strings.TrimPrefix(path, "./")
}
