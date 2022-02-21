package control

import (
	"context"

	"github.com/mredolatti/tf/codigo/common/is2fs"
	"github.com/mredolatti/tf/codigo/common/log"
)

// UserManagementServer sarasa
type ControlServer struct {
	is2fs.UnimplementedFileRefSyncServer
	logger log.Interface
}

func New(logger log.Interface) (*ControlServer, error) {
	return &ControlServer{logger: logger}, nil
}

func (c *ControlServer) SyncUser(ctx context.Context, request *is2fs.SyncUserRequest) (*is2fs.Updates, error) {
	c.logger.Info("incoming request: %+v, ", request)
	return &is2fs.Updates{
		Updates:            []*is2fs.Update{},
		PreviousCheckpoint: request.GetCheckpoint(),
		NewCheckpoint:      request.GetCheckpoint() + 1,
	}, nil
}
