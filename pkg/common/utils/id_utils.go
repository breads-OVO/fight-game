package utils

import (
	"strings"

	"github.com/google/uuid"
)

// GenUUID 生成UUID
func GenUUID() string {
	return strings.ReplaceAll(uuid.New().String(), "-", "")
}

// GenUUIDWithPrefix 带前缀UUID
func GenUUIDWithPrefix(prefix string) string {
	return prefix + GenUUID()
}
