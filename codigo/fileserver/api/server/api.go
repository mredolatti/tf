package server

import (
	"fmt"
	"net"

	"github.com/mredolatti/tf/codigo/common/is2fs"
	"github.com/mredolatti/tf/codigo/common/log"
	"github.com/mredolatti/tf/codigo/fileserver/api/server/control"
	"github.com/mredolatti/tf/codigo/fileserver/filemanager"
	"google.golang.org/grpc"
)

type Options struct {
	Logger      log.Interface
	Port        int
	FileManager filemanager.Interface
}

type ServerAPI struct {
	logger log.Interface
	server *grpc.Server
	port   int
}

func New(options *Options) (*ServerAPI, error) {

	controlServer, err := control.New(options.Logger, options.FileManager)
	if err != nil {
		return nil, fmt.Errorf("error instantiating control server: %w", err)
	}

	server := grpc.NewServer()
	is2fs.RegisterFileRefSyncServer(server, controlServer)

	return &ServerAPI{
		logger: options.Logger,
		server: server,
		port:   options.Port,
	}, nil
}

func (s *ServerAPI) Start() error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return fmt.Errorf("failed to open TCP connection on port %d", s.port)
	}

	return s.server.Serve(lis)
}
