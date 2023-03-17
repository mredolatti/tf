package registrar

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/mredolatti/tf/codigo/common/dtos"
	"github.com/mredolatti/tf/codigo/common/log"
)

const (
	registerPath string = "register"
)

type indexServerClient interface {
	RegisterServer(ctx context.Context, serverInfo *dtos.ServerInfoDTO) (*dtos.RegistrationResultDTO, error)
	CheckServerStatusRequest() (*dtos.ServerStatusDTO, error)
	PauseServer() (*dtos.ServerStatusDTO, error)
	ResumeServer() (*dtos.ServerStatusDTO, error)
}

type indexServerClientConfig struct {
	Logger   log.Interface
	BaseURL  string
	RootCAFN string
	CertFN   string
	KeyFN    string
}

type indexServerClientImpl struct {
	logger  log.Interface
	client  http.Client
	baseURL string
}

func newIndexServerClient(cfg *indexServerClientConfig) (*indexServerClientImpl, error) {

	tlsConfig, err := setupTLSConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("error setting up tls: %w", err)
	}

	return &indexServerClientImpl{
		baseURL: cfg.BaseURL,
		logger:  cfg.Logger,
		client: http.Client{
			Transport: &http.Transport{
				TLSClientConfig: tlsConfig,
			},
		},
	}, nil
}

// RegisterServer implements IndexServerClient
func (c *indexServerClientImpl) RegisterServer(ctx context.Context, serverInfo *dtos.ServerInfoDTO) (*dtos.RegistrationResultDTO, error) {
	body, err := json.Marshal(serverInfo)
	if err != nil {
		return nil, fmt.Errorf("error serializing server info: %w", err)
	}

	dstURL, err := url.JoinPath(c.baseURL, registerPath)
	if err != nil {
		return nil, fmt.Errorf("error building destintaion url: %w", err)
	}

	request, err := http.NewRequestWithContext(ctx, "POST", dstURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("error building request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")

	response, err := c.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("error performing request: %w", err)
	}

	if c := response.StatusCode; c < 200 || c >= 300 {
		return nil, fmt.Errorf("non-2xx (%d) status code returned", c)
	}

	respBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	var result dtos.RegistrationResultDTO
	if err = json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("error deserializing body: %w", err)
	}

	return &result, nil
}

// CheckServerStatusRequest implements IndexServerClient
func (*indexServerClientImpl) CheckServerStatusRequest() (*dtos.ServerStatusDTO, error) {
	panic("unimplemented")
}

// PauseServer implements IndexServerClient
func (*indexServerClientImpl) PauseServer() (*dtos.ServerStatusDTO, error) {
	panic("unimplemented")
}

// ResumeServer implements IndexServerClient
func (*indexServerClientImpl) ResumeServer() (*dtos.ServerStatusDTO, error) {
	panic("unimplemented")
}

func setupTLSConfig(config *indexServerClientConfig) (*tls.Config, error) {
	certBytes, err := ioutil.ReadFile(config.RootCAFN)
	if err != nil {
		return nil, fmt.Errorf("error reading root cert file: %w", err)
	}
	caPool := x509.NewCertPool()
	caPool.AppendCertsFromPEM(certBytes)

	certs, err := tls.LoadX509KeyPair(config.CertFN, config.KeyFN)
	if err != nil {
		return nil, fmt.Errorf("cannot read server certficate chain / private key files: %w", err)
	}

	return &tls.Config{
		Certificates: []tls.Certificate{certs},
		RootCAs:      caPool,
		ClientCAs:    caPool,
	}, nil
}

var _ indexServerClient = (*indexServerClientImpl)(nil)
