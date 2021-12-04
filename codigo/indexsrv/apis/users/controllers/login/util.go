package login

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/golang-jwt/jwt"

	"github.com/mredolatti/tf/codigo/common/log"
)

const (
	issuerA = "accounts.google.com"
	issuerB = "https://accounts.google.com"
)

var errNoSuchKey = errors.New("no such key")
var errInvalidGoogleJWT = errors.New("Invalid Google JWT")
var errInvalidISS = errors.New("iss is invalid")
var errInvalidAudience = errors.New("aud is invalid")
var errExpiredJWT = errors.New("JWT is expired")

type googleKeyFetcher struct {
	mutex  sync.Mutex
	logger log.Interface
	keys   map[string]string // keyID -> public key
}

func newGoogleKeyFetcher(logger log.Interface) *googleKeyFetcher {
	return &googleKeyFetcher{
		keys:   make(map[string]string),
		logger: logger,
	}
}

func (f *googleKeyFetcher) Run() {
	for {
		resp, err := http.Get("https://www.googleapis.com/oauth2/v1/certs")
		if err != nil {
			f.logger.Error("error fetching google public key: ", err)
			continue
		}
		dat, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			f.logger.Error("error parsing google public key response body:", err)
			continue
		}

		myResp := map[string]string{}
		err = json.Unmarshal(dat, &myResp)
		if err != nil {
			f.logger.Error("error deserializing google public key response:", err)
			continue
		}

		f.mutex.Lock()
		for keyID, pubKey := range myResp {
			f.keys[keyID] = pubKey
		}
		f.mutex.Unlock()

		time.Sleep(30 * time.Second)
	}
}

func (f *googleKeyFetcher) Get(keyID string) (string, error) {
	f.mutex.Lock()
	key, ok := f.keys[keyID]
	f.mutex.Unlock()
	if !ok {
		return "", errNoSuchKey
	}

	return key, nil
}

type googleClaims struct {
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	FirstName     string `json:"given_name"`
	LastName      string `json:"family_name"`
	jwt.StandardClaims
}

func validateJWT(tokenStr string, clientID string, gpkf *googleKeyFetcher) (*googleClaims, error) {
	claimsStruct := googleClaims{}
	token, err := jwt.ParseWithClaims(
		tokenStr,
		&claimsStruct,
		func(token *jwt.Token) (interface{}, error) {
			pem, err := gpkf.Get(fmt.Sprintf("%s", token.Header["kid"]))
			if err != nil {
				return nil, err
			}
			key, err := jwt.ParseRSAPublicKeyFromPEM([]byte(pem))
			if err != nil {
				return nil, err
			}
			return key, nil
		},
	)
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*googleClaims)
	if !ok {
		return nil, errInvalidGoogleJWT
	}

	if claims.Issuer != issuerA && claims.Issuer != issuerB {
		return nil, errInvalidISS
	}

	if claims.Audience != clientID {
		return nil, errInvalidAudience
	}

	if claims.ExpiresAt < time.Now().UTC().Unix() {
		return nil, errExpiredJWT
	}

	return claims, nil
}
