package registrar

import (
	"errors"

	"github.com/golang-jwt/jwt"
)

var ErrInvalidToken = errors.New("could not parse token")

// Token is a type mapping to jwt.Token
type Token struct {
	raw string
}

// TokenClaims is a type alias to jwt.StandardClaims
type TokenClaims = jwt.StandardClaims

// ParseClaims type-asserts the token's claims and returns them
func (t *Token) ParseClaims() (*TokenClaims, error) {
	var parser jwt.Parser
	parsed, _, err := parser.ParseUnverified(t.raw, &TokenClaims{})
	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := parsed.Claims.(*TokenClaims)
	if !ok {
		return nil, ErrInvalidClaims
	}
	return claims, nil
}

// Raw returns the full token string
func (t *Token) Raw() string {
	return t.raw
}


