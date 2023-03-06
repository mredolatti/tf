package authentication

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

type SessionKeyFactory interface {
	Generate(n int) (string, error)
}

// Adaptado de: http://blog.questionable.services/article/generating-secure-random-numbers-crypto-rand/

type CryptoBase64Generator struct{}

func (c *CryptoBase64Generator) Generate(n int) (string, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("error reading cryptographic random data: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), err
}

var _ SessionKeyFactory = (*CryptoBase64Generator)(nil)
