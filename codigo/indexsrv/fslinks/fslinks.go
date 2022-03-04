package fslinks

import (
	"context"
	"fmt"
	"io"

	"github.com/mredolatti/tf/codigo/common/is2fs"
	"github.com/mredolatti/tf/codigo/common/log"
	"github.com/mredolatti/tf/codigo/indexsrv/models"
	"github.com/mredolatti/tf/codigo/indexsrv/repository"
)

// Interface defines the methods for a file-server links monitor
type Interface interface {
	NotifyServerUp(ctx context.Context, serverID string) error
	FetchUpdates(ctx context.Context, serverID string, userID string, checkpoint int64) ([]models.Update, error)
}

// Impl is an implementation of fslink.Interface
type Impl struct {
	logger  log.Interface
	users   repository.UserRepository
	orgs    repository.OrganizationRepository
	servers repository.FileServerRepository
	conns   connTracker
}

// New constructs a new file-server link monitor
func New(logger log.Interface, userRepo repository.UserRepository, orgRepo repository.OrganizationRepository) (*Impl, error) {
	return &Impl{
		logger: logger,
		users:  userRepo,
		orgs:   orgRepo,
	}, nil
}

// NotifyServerUp is meant to be called whenever a server announces itself
func (i *Impl) NotifyServerUp(ctx context.Context, serverID string, healthy bool, uptime int64) error {

	fs, err := i.servers.Get(ctx, serverID)
	if err != nil {
		return fmt.Errorf("error fetching server '%s': %w", serverID, err)
	}

	_, err = i.conns.get(fs)
	if err != nil {
		return fmt.Errorf("error connecting to server '%s': %w", serverID, err)
	}

	return nil
}

// FetchUpdates asks the server for the latest changes in file for a specific user
func (i *Impl) FetchUpdates(ctx context.Context, serverID string, userID string, checkpoint int64) ([]models.Update, error) {
	fs, err := i.servers.Get(ctx, serverID)
	if err != nil {
		return nil, fmt.Errorf("error fetching server '%s': %w", serverID, err)
	}

	pack, err := i.conns.get(fs)
	if err != nil {
		return nil, fmt.Errorf("error connecting to server '%s': %w", serverID, err)
	}

	stream, err := pack.client.SyncUser(ctx, &is2fs.SyncUserRequest{
		Checkpoint: checkpoint,
		UserID:     userID,
		KeepAlive:  false, //TODO(mredolatti): Either implement this or remove it
	})
	if err != nil {
		return nil, fmt.Errorf(
			"error syncing available files for user '%s' in server '%s': %w",
			userID, serverID, err)
	}

	var updates []models.Update
	for {
		update, err := stream.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, fmt.Errorf(
				"error received when reading from stream for user '%s' in server '%s': %w",
				userID, serverID, err)
		}

		updates = append(updates, models.Update{
			OrganizationID: fs.OrganizationID(),
			ServerID:       fs.ID(),
			FileRef:        update.FileReference,
			Checkpoint:     update.Checkpoint,
			ChangeType:     toUpdateType(update.ChangeType),
		})

	}

	return nil, nil
}

func toUpdateType(ct is2fs.ChangeType) models.UpdateType {
	switch ct {
	case is2fs.ChangeType_FileChangeAdd:
		return models.UpdateTypeFileAdd
	case is2fs.ChangeType_FileChangeDelete:
		return models.UpdateTypeFileDelete
	case is2fs.ChangeType_FileChangeUpdate:
		return models.UpdateTypeFileUpdate
	}

	panic(fmt.Sprintf("update type not defined for value %d", ct))
}
