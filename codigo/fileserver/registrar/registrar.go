package registrar

import (
	"context"
	"fmt"
	"net/url"

	"github.com/mredolatti/tf/codigo/common/dtos"
	"github.com/mredolatti/tf/codigo/common/log"
)

const (
	authPath  = "auth"
	tokenPath = "token"
	fetchPath = "files"
)

type Interface interface {
	EnsureThisServerIsRegistered(ctx context.Context) error
}

type Config struct {
	ServerName         string
	OrgName            string
	ThisHost           string
	ClientPort         int
	ControlPort        int
	IndexServerBaseURL string
	RootCAFn           string
	CertChainFn        string
	PrivateKeyFn       string
}

type Impl struct {
	isClient   indexServerClient
	serverInfo dtos.ServerInfoDTO
}

func New(cfg *Config, logger log.Interface) (*Impl, error) {
	isc, err := newIndexServerClient(&indexServerClientConfig{
		Logger:   logger,
		BaseURL:  cfg.IndexServerBaseURL,
		RootCAFN: cfg.RootCAFn,
		KeyFN:    cfg.PrivateKeyFn,
		CertFN:   cfg.CertChainFn,
	})
	if err != nil {
		return nil, fmt.Errorf("error constructing index-server api client: %w", err)
	}

	authURL, tokenURL, fetchURL, controlURL, err := buildURLs(cfg)
	if err != nil {
		return nil, fmt.Errorf("error building service URLs for registration purposes: %w", err)
	}

	return &Impl{
		isClient: isc,
		serverInfo: dtos.ServerInfoDTO{
			OrgName:         cfg.OrgName,
			Name:            cfg.ServerName,
			AuthURL:         authURL,
			TokenURL:        tokenURL,
			FetchURL:        fetchURL,
			ControlEndpoint: controlURL,
		},
	}, nil
}

// EnsureThisServerIsRegistered implements Interface
func (r *Impl) EnsureThisServerIsRegistered(ctx context.Context) error {

	response, err := r.isClient.RegisterServer(ctx, &r.serverInfo)
	if err != nil {
		return fmt.Errorf("error attempting registration: %w", err)
	}

	switch response.Result {
	case dtos.ResultOK, dtos.ResultAlreadyRegistered:
	default:
		return fmt.Errorf("registration failed with result: %d", response.Result)
	}

	return nil
}

func buildURLs(cfg *Config) (auth, token, fetch, control string, err error) {
	clientBaseURL := fmt.Sprintf("https://%s:%d/", cfg.ThisHost, cfg.ClientPort)
	if auth, err = url.JoinPath(clientBaseURL, authPath); err != nil {
		return "", "", "", "", err
	}

	if token, err = url.JoinPath(clientBaseURL, tokenPath); err != nil {
		return "", "", "", "", err
	}

	if fetch, err = url.JoinPath(clientBaseURL, fetchPath); err != nil {
		return "", "", "", "", err
	}

	control = fmt.Sprintf("%s:%d", cfg.ThisHost, cfg.ControlPort)
	return auth, token, fetch, control, nil
}

var _ Interface = (*Impl)(nil)
