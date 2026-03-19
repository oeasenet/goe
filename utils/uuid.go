package utils

import (
	"strings"

	"github.com/google/uuid"
)

// GenerateUUIDv7 generates a UUID version 7 string without dashes and returns it. Returns an empty string on error.
func GenerateUUIDv7() string {
	uuidWithDashes, err := uuid.NewV7()
	if err != nil {
		return ""
	}
	uuidWithoutDashes := strings.ReplaceAll(uuidWithDashes.String(), "-", "")
	return uuidWithoutDashes
}
