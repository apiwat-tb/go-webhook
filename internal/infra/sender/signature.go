package sender

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// Sign computes an HMAC-SHA256 signature for payload using secret and returns "sha256=<hex>".
func Sign(secret string, payload []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

// Verify performs constant-time validation of the provided signature.
func Verify(secret string, payload []byte, sig string) bool {
	expected := Sign(secret, payload)
	return hmac.Equal([]byte(expected), []byte(sig))
}
