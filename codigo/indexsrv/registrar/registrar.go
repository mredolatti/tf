package registrar

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/mredolatti/tf/codigo/indexsrv/models"
	"github.com/mredolatti/tf/codigo/indexsrv/repository"
)

const (
	clientID     = "0000000"
	clientSecret = "1234567890"
)

// Public errors
var (
	ErrAccountExists = errors.New("account already exists")
)

// Interface defines the set of methods for managing user <-> server links
type Interface interface {
	InitiateLinkProcess(ctx context.Context, userID string, serverID string, force bool) (string, error)
	CompleteLinkProcess(ctx context.Context, state string, code string) error
}

// Impl is an implementation of the registar interface
type Impl struct {
	randGen       *randGenerator
	fileServers   repository.FileServerRepository
	userAccoounts repository.UserAccountRepository
	oauth2Flows   repository.PendingOAuth2Repository
	httpClient    http.Client
}

// New constructs a new registrar
func New(
	fileServers repository.FileServerRepository,
	userAccoounts repository.UserAccountRepository,
	oauth2Flows repository.PendingOAuth2Repository,
	tlsConfig *tls.Config,
) *Impl {
	return &Impl{
		randGen:       newRandGenerator(),
		fileServers:   fileServers,
		userAccoounts: userAccoounts,
		oauth2Flows:   oauth2Flows,
		httpClient: http.Client{
			Transport: &http.Transport{TLSClientConfig: tlsConfig},
		},
	}
}

// InitiateLinkProcess sets up the initial parameters to authenticate againsta a file-server,
// and returns a URL to redirect the user to
func (i *Impl) InitiateLinkProcess(ctx context.Context, userID string, serverID string, force bool) (string, error) {
	if acc, _ := i.userAccoounts.Get(ctx, userID, serverID); !force && acc != nil {
		return "", ErrAccountExists
	}

	server, err := i.fileServers.Get(ctx, serverID)
	if err != nil {
		return "", fmt.Errorf("error fetching server from repository: %w", err)
	}

	state := i.randGen.RandStringRunes(50)
	redirectURL, err := buildRedirectURL(server, clientID, state)
	if err != nil {
		return "", fmt.Errorf("error building redirect URL: %w", err)
	}

	if _, err := i.oauth2Flows.Put(ctx, userID, serverID, state); err != nil {
		return "", fmt.Errorf("error persisting oauth2 flow init parameters: %w", err)

	}

	return redirectURL.String(), nil
}

// CompleteLinkProcess effectively sets up a user account after the reception of an auth code
func (i *Impl) CompleteLinkProcess(ctx context.Context, state string, code string) error {

	flow, err := i.oauth2Flows.Pop(ctx, state)
	if err != nil {
		return fmt.Errorf("error fetching pending flow from repository: %w", err)
	}

	tokenResp, err := i.exchangeCode(ctx, flow.FileServerID(), code)
	if err != nil {
		return fmt.Errorf("error exchanging code for token: %w", err)
	}

	if _, err := i.userAccoounts.Add(ctx, flow.UserID(), flow.FileServerID(), tokenResp.AccessToken, tokenResp.RefreshToken); err != nil {
		return fmt.Errorf("error creating user account with received tokens: %w", err)
	}

	return nil
}

func (i *Impl) exchangeCode(ctx context.Context, serverID string, code string) (*tokenResponse, error) {

	server, err := i.fileServers.Get(ctx, serverID)
	if err != nil {
		return nil, fmt.Errorf("error fetching server from repository: %w", err)
	}

	req, err := http.NewRequest("GET", server.TokenURL(), nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request for code exchange: %w", err)
	}

	// ?grant_type=authorization_code&client_id=${cid}&client_secret=${secret}&scope=read&code=${code//[$'\t\r\n ']}&redirect_uri=${redirect}"
	qps := req.URL.Query()
	qps.Add("grant_type", "authorization_code")
	qps.Add("client_id", clientID)
	qps.Add("client_secret", clientSecret)
	qps.Add("scope", "read")
	qps.Add("code", code)
	qps.Add("redirect_uri", "http://index-server:9876/user_accounts/callback")
	req.URL.RawQuery = qps.Encode()

	resp, err := i.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making code exchange request: %w", err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	var tokenResponse tokenResponse
	err = json.Unmarshal(body, &tokenResponse)
	if err != nil {
		return nil, fmt.Errorf("error parsing resopnse json: %w", err)
	}

	return &tokenResponse, nil

}

func buildRedirectURL(server models.FileServer, clientID string, state string) (*url.URL, error) {
	redirectURL, err := url.Parse(server.AuthURL())
	if err != nil {
		return nil, fmt.Errorf("error parsing URL '%s': %w", server.AuthURL(), err)
	}

	queryString := url.Values{}
	queryString.Add("client_id", clientID)
	queryString.Add("state", state)
	queryString.Add("response_type", "code")
	redirectURL.RawQuery = queryString.Encode()

	return redirectURL, nil
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
}
