package utils

import (
	"crypto/sha256"
	"fmt"
)

// Sha256Entry sha256加密
func Sha256Entry(password string) string {
	h := sha256.Sum256([]byte(password))
	return fmt.Sprintf("%x", h)
}
