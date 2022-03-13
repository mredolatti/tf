package fslinks

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"sync"

	"github.com/mredolatti/tf/codigo/common/is2fs"
	"github.com/mredolatti/tf/codigo/indexsrv/models"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials"
)

type fsClient struct {
	client is2fs.FileRefSyncClient
	conn   *grpc.ClientConn
}

type fsClientMap map[string]*fsClient

type connTracker struct {
	servers fsClientMap
	creds   credentials.TransportCredentials
	mutex   sync.Mutex
}

func newConnTracker(rootCA string) (*connTracker, error) {

	creds, err := parseCredentials(rootCA)
	if err != nil {
		return nil, fmt.Errorf("error setting up gRPC client TLS credentials: ", creds)
	}

	return &connTracker{
		servers: make(fsClientMap),
		creds:   creds,
	}, nil
}

func (t *connTracker) get(server models.FileServer) (*fsClient, error) {

	t.mutex.Lock()
	defer t.mutex.Unlock()
	packed, exists := t.servers[server.ID()]

	if exists && !shouldRecycle(packed.conn) {
		exists = false // recreate the connection
	}

	if !exists {
		var err error
		conn, err := grpc.Dial(server.ControlEndpoint(), grpc.WithTransportCredentials(t.creds))
		if err != nil {
			return nil, fmt.Errorf("failed to connect to server: %w", err)
		}

		packed = &fsClient{
			client: is2fs.NewFileRefSyncClient(conn),
			conn:   conn,
		}

		t.servers[server.ID()] = packed
	}

	return packed, nil
}

func shouldRecycle(conn *grpc.ClientConn) bool {
	switch conn.GetState() {
	case connectivity.Idle, connectivity.Connecting, connectivity.Ready, connectivity.TransientFailure:
		// Either everything is ok, or is in the way to become so.
		return false
	default:
		// Either a shutdown or an invalid state means we shoulr replace this connection with a new one
		return true
	}
}

func parseCredentials(rootCA string) (credentials.TransportCredentials, error) {
	certData, err := ioutil.ReadFile(rootCA)
	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(certData) {
		return nil, errors.New("error adding certificate to pool")
	}

	return credentials.NewTLS(&tls.Config{
		RootCAs: certPool,
	}), nil
}
