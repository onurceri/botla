package httputil

import "github.com/google/uuid"

func ParseUUID(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}

func IsValidUUID(s string) bool {
	_, err := uuid.Parse(s)
	return err == nil
}
