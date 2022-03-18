package server

import (
	"crypto/tls"
	"fmt"
	"net"

	"github.com/mredolatti/tf/codigo/common/is2fs"
	"github.com/mredolatti/tf/codigo/common/log"
	"github.com/mredolatti/tf/codigo/fileserver/api/server/control"
	"github.com/mredolatti/tf/codigo/fileserver/filemanager"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Options to configure server-side gRPC API
type Options struct {
	Logger                   log.Interface
	Port                     int
	FileManager              filemanager.Interface
	ServerCertificateChainFN string
	ServerPrivateKeyFN       string
	RootCAFn                 string
}

// ServerAPI is the gRPC server
type ServerAPI struct {
	logger log.Interface
	server *grpc.Server
	port   int
}

// New constructs a new server-side API
func New(options *Options) (*ServerAPI, error) {

	credentials, err := parseTLSCredentials(options)
	if err != nil {
		return nil, fmt.Errorf("failed to build TLS credentials: %w", err)
	}

	controlServer, err := control.New(options.Logger, options.FileManager)
	if err != nil {
		return nil, fmt.Errorf("error instantiating control server: %w", err)
	}

	var auth authInterceptor

	server := grpc.NewServer(
		grpc.Creds(credentials),
		grpc.UnaryInterceptor(auth.Unary()),
		grpc.StreamInterceptor(auth.Stream()),
	)
	is2fs.RegisterFileRefSyncServer(server, controlServer)

	return &ServerAPI{
		logger: options.Logger,
		server: server,
		port:   options.Port,
	}, nil
}

// Start listening for incoming connections
func (s *ServerAPI) Start() error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return fmt.Errorf("failed to open TCP connection on port %d", s.port)
	}

	return s.server.Serve(lis)
}

func parseTLSCredentials(options *Options) (credentials.TransportCredentials, error) {
	serverCert, err := tls.LoadX509KeyPair(options.ServerCertificateChainFN, options.ServerPrivateKeyFN)
	if err != nil {
		return nil, err
	}

	return credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.NoClientCert,
	}), nil
}
