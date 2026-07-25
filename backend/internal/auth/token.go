package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
)

// GenerateToken returns a random, URL-safe token to send to the user and its
// SHA-256 hash to store in the database. Only the hash is persisted, so a
// database leak does not reveal usable tokens. Used for password-reset and
// email-change confirmation links.
func GenerateToken() (token, hash string, err error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", "", err
	}
	token = base64.RawURLEncoding.EncodeToString(b)
	return token, HashToken(token), nil
}

// HashToken returns the SHA-256 hex digest of a token. Used both when storing a
// freshly generated token and when looking one up for verification.
func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
