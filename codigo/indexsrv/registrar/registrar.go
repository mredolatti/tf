package registrar

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/mredolatti/tf/codigo/indexsrv/models"
	"github.com/mredolatti/tf/codigo/indexsrv/repository"
)

const (
	clientID     = "0000000"
	clientSecret = "1234567890"

	tokenValidityMaxTolerance = 5 * time.Minute
)

// Public errors
var (
	ErrOrgNotFound             = errors.New("organization not found")
	ErrServerAlreadyRegistered = errors.New("server already registered")

	ErrAccountExists = errors.New("account already exists")
	ErrInvalidClaims = errors.New("unknown claims in jwt")

	ErrNeedReLink = errors.New("accounts needs to be re-linked")
)

// Interface defines the set of methods for managing user <-> server links
type Interface interface {
	AddNewOrganization(ctx context.Context, name string) error
	ListOrganizations(ctx context.Context) ([]models.Organization, error)
	GetOrganization(ctx context.Context, name string) (models.Organization, error)
	ListServers(ctx context.Context, query models.FileServersQuery) ([]models.FileServer, error)
	GetServer(ctx context.Context, orgName string, name string) (models.FileServer, error)
	RegisterServer(ctx context.Context, orgId string, name string, authURL string, tokenURL string, fetchURL string, controlEndpoint string) error

	InitiateLinkProcess(ctx context.Context, userID string, orgName string, serverName string, force bool) (string, error)
	CompleteLinkProcess(ctx context.Context, state string, code string) error
	GetValidToken(ctx context.Context, userID string, orgName string, serverName string) (*Token, error)
}

// Impl is an implementation of the registar interface
type Impl struct {
	redirectURL   string
	randGen       *randGenerator
	fileServers   repository.FileServerRepository
	organizations repository.OrganizationRepository
	userAccounts  repository.UserAccountRepository
	oauth2Flows   repository.PendingOAuth2Repository
	httpClient    http.Client
}

type Config struct {
	FileServers        repository.FileServerRepository
	UserAccounts       repository.UserAccountRepository
	Organizations      repository.OrganizationRepository
	Pauth2Flows        repository.PendingOAuth2Repository
	BaseURL            string
	RootCAFN           string
	ServerCertFN       string
	ServerPrivateKeyFN string
}

// New constructs a new registrar
func New(cfg *Config) *Impl {

	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = setupTLSConfig(cfg.RootCAFN, cfg.ServerCertFN, cfg.ServerPrivateKeyFN)
    url, _ := url.JoinPath(cfg.BaseURL, "user_accounts/callback")
	return &Impl{
		randGen:       newRandGenerator(),
		fileServers:   cfg.FileServers,
		userAccounts:  cfg.UserAccounts,
		organizations: cfg.Organizations,
		oauth2Flows:   cfg.Pauth2Flows,
		httpClient:    http.Client{Transport: transport},
		redirectURL:   url,
	}
}

func (i *Impl) AddNewOrganization(ctx context.Context, name string) error {
	if _, err := i.organizations.Add(ctx, name); err != nil {
		return fmt.Errorf("error storing new org in database: %w", err)
	}
	return nil
}

func (i *Impl) ListOrganizations(ctx context.Context) ([]models.Organization, error) {
	orgs, err := i.organizations.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("error reading orgs from db: %w", err)
	}
	return orgs, nil
}

func (i *Impl) GetOrganization(ctx context.Context, name string) (models.Organization, error) {
	org, err := i.organizations.Get(ctx, name)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrOrgNotFound
		}
		return nil, fmt.Errorf("error reading org from db: %w", err)
	}
	return org, nil
}

func (i *Impl) ListServers(ctx context.Context, query models.FileServersQuery) ([]models.FileServer, error) {
	fss, err := i.fileServers.List(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error reading servers from db: %w", err)
	}
	return fss, nil
}

func (i *Impl) GetServer(ctx context.Context, orgName string, name string) (models.FileServer, error) {
	server, err := i.fileServers.Get(ctx, orgName, name)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrOrgNotFound
		}
		return nil, fmt.Errorf("error reading server from db: %w", err)
	}
	return server, nil
}

// RegisterServer implements Interface
func (i *Impl) RegisterServer(
	ctx context.Context,
	orgName string,
	serverName string,
	authURL string,
	tokenURL string,
	fetchURL string,
	controlEndpoint string,
) error {

	org, err := i.organizations.GetByName(ctx, orgName)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrOrgNotFound
		}
		return fmt.Errorf("error fetching organization '%s': %w", orgName, err)
	}

	if _, err := i.fileServers.Add(ctx, serverName, org.Name(), authURL, tokenURL, fetchURL, controlEndpoint); err != nil {
		if errors.Is(err, repository.ErrAlreadyExists) {
			return ErrServerAlreadyRegistered
		}
		return fmt.Errorf("error storing server info in db: %w", err)
	}
	return nil
}

// InitiateLinkProcess sets up the initial parameters to authenticate againsta a file-server,
// and returns a URL to redirect the user to
func (i *Impl) InitiateLinkProcess(ctx context.Context, userID string, orgName string, serverName string, force bool) (string, error) {
	if acc, _ := i.userAccounts.Get(ctx, userID, orgName, serverName); !force && acc != nil {
		return "", ErrAccountExists
	}

	server, err := i.fileServers.Get(ctx, orgName, serverName)
	if err != nil {
		return "", fmt.Errorf("error fetching server from repository: %w", err)
	}

	state := i.randGen.RandStringRunes(50)
	redirectURL, err := buildRedirectURL(server, clientID, state)
	if err != nil {
		return "", fmt.Errorf("error building redirect URL: %w", err)
	}

	if _, err := i.oauth2Flows.Put(ctx, userID, orgName, serverName, state); err != nil {
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

	tokenResp, err := i.exchangeCode(ctx, flow.OrganizationName(), flow.ServerName(), code)
	if err != nil {
		return fmt.Errorf("error exchanging code for token: %w", err)
	}

	if _, err := i.userAccounts.AddOrUpdate(ctx, flow.UserID(), flow.OrganizationName(), flow.ServerName(), tokenResp.AccessToken, tokenResp.RefreshToken); err != nil {
		return fmt.Errorf("error creating user account with received tokens: %w", err)
	}

	return nil
}

func (i *Impl) exchangeCode(ctx context.Context, orgName string, serverName string, code string) (*tokenResponse, error) {

	server, err := i.fileServers.Get(ctx, orgName, serverName)
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

// GetValidToken returns the current token if still valid or a refreshed one otherwise
func (i *Impl) GetValidToken(ctx context.Context, userID string, orgName string, serverName string) (*Token, error) {
	acc, err := i.userAccounts.Get(ctx, userID, orgName, serverName)
	if err != nil {
		return nil, fmt.Errorf("error getting account from repository: %w", err)
	}

	if token := acc.Token(); isTokenStillValid(token) {
		return &Token{raw: token}, nil
	}

	newAccessToken, err := i.doRefreshToken(ctx, userID, orgName, serverName, acc.RefreshToken())
	if err != nil {
		return nil, fmt.Errorf("error refreshing token: %w", err)
	}

	return &Token{raw: newAccessToken}, nil
}

func (i *Impl) doRefreshToken(ctx context.Context, userID string, orgName string, serverName string, refreshToken string) (string, error) {
	server, err := i.fileServers.Get(ctx, orgName, serverName)
	if err != nil {
		return "", fmt.Errorf("error fetching server from repository: %w", err)
	}

	status, tokenResponse, err := i.makeTokenRefreshRequest(server.TokenURL(), refreshToken)
	switch status {
	case 200: // do nothing
	case 401:
		return "", ErrNeedReLink
	default:
		return "", fmt.Errorf("error making request: %w", err)
	}
	if err := i.userAccounts.UpdateTokens(ctx, userID, orgName, serverName, tokenResponse.AccessToken, tokenResponse.RefreshToken); err != nil {
		return "", fmt.Errorf("error storing new tokens in db: %w", err)
	}

	return tokenResponse.AccessToken, nil
}

func (i *Impl) makeTokenRefreshRequest(tokenURL string, refreshToken string) (int, *tokenResponse, error) {
	bodyForm := url.Values{}
	bodyForm.Add("grant_type", "refresh_token")
	bodyForm.Add("client_id", clientID)
	bodyForm.Add("client_secret", clientSecret)
	bodyForm.Add("refresh_token", refreshToken)
	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(bodyForm.Encode()))
	if err != nil {
		return 0, nil, fmt.Errorf("error creating request for token refresh: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := i.httpClient.Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("error making token refresh request: %w", err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return resp.StatusCode, nil, fmt.Errorf("error reading response body: %w", err)
	}

	var tokenResponse tokenResponse
	err = json.Unmarshal(body, &tokenResponse)
	if err != nil {
		return resp.StatusCode, nil, fmt.Errorf("error parsing response json: %w", err)

	}

	return resp.StatusCode, &tokenResponse, nil
}

func isTokenStillValid(token string) bool {
	var parser jwt.Parser
	parsed, _, err := parser.ParseUnverified(token, &jwt.StandardClaims{})
	if err != nil {
		return false
	}

	claims, ok := parsed.Claims.(*jwt.StandardClaims)
	if !ok {
		return false
	}

	return time.Now().Before(time.Unix(claims.ExpiresAt, 0))
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

func setupTLSConfig(rootCAFN string, serverCertFN string, serverPrivateKeyFN string) *tls.Config {
	certBytes, err := ioutil.ReadFile(rootCAFN)
	if err != nil {
		panic("cannot read root certificate file: " + err.Error())
	}
	caPool := x509.NewCertPool()
	caPool.AppendCertsFromPEM(certBytes)

	certs, err := tls.LoadX509KeyPair(serverCertFN, serverPrivateKeyFN)
	if err != nil {
		panic("cannot read server certficate chain / private key files: " + err.Error())
	}

	return &tls.Config{
		Certificates: []tls.Certificate{certs},
		RootCAs:      caPool,
		ClientAuth:   tls.RequestClientCert,
	}
}

var _ Interface = (*Impl)(nil)
