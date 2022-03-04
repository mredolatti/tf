package fslinks

import (
	"fmt"
	"sync"

	"github.com/mredolatti/tf/codigo/common/is2fs"
	"github.com/mredolatti/tf/codigo/indexsrv/models"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

type fsClient struct {
	client is2fs.FileRefSyncClient
	conn   *grpc.ClientConn
}

type fsClientMap map[string]*fsClient

type connTracker struct {
	servers fsClientMap
	mutex   sync.Mutex
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
		conn, err := grpc.Dial(server.ControlEndpoint())
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
