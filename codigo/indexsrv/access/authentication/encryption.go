package authentication

import (
	"golang.org/x/crypto/bcrypt"
)

type PasswordHasher interface {
	Hash(string) (string, error)
}

type BCryptHasher struct {}

func (h *BCryptHasher) Hash(p string) (string, error) {
	// TODO(mredolatti): parametrize cost to a struct prop.
	hash, err := bcrypt.GenerateFromPassword([]byte(p), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), err
}

var _ PasswordHasher = (*BCryptHasher)(nil)
