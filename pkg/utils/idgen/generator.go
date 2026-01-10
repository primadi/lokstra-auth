package idgen

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

// GenerateID generates a unique ID with a prefix
func GenerateID(prefix string) string {
	// Generate 16 random bytes (128 bits)
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based ID if random fails
		return fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())
	}

	// Convert to hex string
	return fmt.Sprintf("%s_%s", prefix, hex.EncodeToString(bytes))
}
