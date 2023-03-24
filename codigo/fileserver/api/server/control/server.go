package control

import (
	"fmt"

	"github.com/mredolatti/tf/codigo/common/is2fs"
	"github.com/mredolatti/tf/codigo/common/log"
	"github.com/mredolatti/tf/codigo/common/refutil"
	"github.com/mredolatti/tf/codigo/fileserver/filemanager"
)

const (
	defaultQueueSize = 10000
)

// Server provides a set of RPCs to get updates on changes in files
type Server struct {
	is2fs.UnimplementedFileRefSyncServer
	logger          log.Interface
	manager         filemanager.Interface
	incomingChanges chan filemanager.Change
}

// New constructs a new server
func New(logger log.Interface, manager filemanager.Interface) (*Server, error) {
	server := &Server{
		logger:          logger,
		manager:         manager,
		incomingChanges: make(chan filemanager.Change, defaultQueueSize),
	}

	manager.AddListener(func(c filemanager.Change) {
		server.incomingChanges <- c
	})

	return server, nil
}

// SyncUser implements the SycUser rpc
func (c *Server) SyncUser(request *is2fs.SyncUserRequest, stream is2fs.FileRefSync_SyncUserServer) error {

	forUser, err := c.manager.ListFileMetadata(
		request.GetUserID(),
		&filemanager.ListQuery{UpdatedAfter: refutil.Ref(request.GetCheckpoint())},
	)
	if err != nil {
		return fmt.Errorf("error getting files for user %s: %w", request.GetUserID(), err)
	}

	for idx := range forUser {
		stream.Send(&is2fs.Update{
			FileReference: forUser[idx].ID(),
			ChangeType:    is2fs.ChangeType_FileChangeUpdate,
			Checkpoint:    forUser[idx].LastUpdated(),
			SizeBytes:     forUser[idx].SizeBytes(),
		})
	}

	// TODO(mredolatti): Implement subscription mechanism
	// - We need a MUX that parses incoming changes, checks if theres a subscription for the affected user,
	//   and queues the messages
	// - Need to POC what happens on client breaking the connection, how to properly cleanup and de-register
	//   from the mux

	return nil
}
