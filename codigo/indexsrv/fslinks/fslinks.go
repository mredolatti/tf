package fslinks

import (
	"context"
	"fmt"
	"io"

	"github.com/mredolatti/tf/codigo/common/is2fs"
	"github.com/mredolatti/tf/codigo/common/log"
	"github.com/mredolatti/tf/codigo/indexsrv/models"
	"github.com/mredolatti/tf/codigo/indexsrv/registrar"
	"github.com/mredolatti/tf/codigo/indexsrv/repository"
)

type ctxKeyUserID struct{}
type ctxKeyOrgName struct{}
type ctxKeyServerName struct{}

// Interface defines the methods for a file-server links monitor
type Interface interface {
	NotifyServerUp(ctx context.Context, serverID string, healthy bool, uptime int64) error
	FetchUpdates(ctx context.Context, orgName string, serverName string, user models.User, checkpoint int64) ([]models.Update, error)
}

// Impl is an implementation of fslink.Interface
type Impl struct {
	logger  log.Interface
	users   repository.UserRepository
	orgs    repository.OrganizationRepository
	servers repository.FileServerRepository
	conns   *connTracker
}

// New constructs a new file-server link monitor
func New(
	logger log.Interface,
	userRepo repository.UserRepository,
	orgRepo repository.OrganizationRepository,
	servers repository.FileServerRepository,
	reg registrar.Interface,
	rootCA string,
) (*Impl, error) {

	connTracker, err := newConnTracker(rootCA, newAuthInterceptor(reg))
	if err != nil {
		return nil, fmt.Errorf("error setting up gRPC connection tracker: %w", err)
	}

	return &Impl{
		logger:  logger,
		conns:   connTracker,
		users:   userRepo,
		orgs:    orgRepo,
		servers: servers,
	}, nil
}

// NotifyServerUp is meant to be called whenever a server announces itself
func (i *Impl) NotifyServerUp(ctx context.Context, serverID string, healthy bool, uptime int64) error {

	// TODO(mredolatti)
	/*
		fs, err := i.servers.Get(ctx, serverID)
		if err != nil {
			return fmt.Errorf("error fetching server '%s': %w", serverID, err)
		}

		_, err = i.conns.get(fs)
		if err != nil {
			return fmt.Errorf("error connecting to server '%s': %w", serverID, err)
		}
	*/

	return nil
}

// FetchUpdates asks the server for the latest changes in file for a specific user
func (i *Impl) FetchUpdates(ctx context.Context, orgName string, serverName string, user models.User, checkpoint int64) ([]models.Update, error) {
	fs, err := i.servers.Get(ctx, orgName, serverName)
	if err != nil {
		return nil, fmt.Errorf("error fetching server '%s/%s': %w", orgName, serverName, err)
	}

	pack, err := i.conns.get(fs)
	if err != nil {
		return nil, fmt.Errorf("error connecting to server '%s/%s': %w", orgName, serverName, err)
	}

	// TODO(mredolatti): poner todo esto dentro de un struct en vez de wrappear 3 veces el contexto
	ctx = context.WithValue(ctx, ctxKeyUserID{}, user.ID())
	ctx = context.WithValue(ctx, ctxKeyOrgName{}, orgName)
	ctx = context.WithValue(ctx, ctxKeyServerName{}, serverName)
	stream, err := pack.client.SyncUser(ctx, &is2fs.SyncUserRequest{Checkpoint: checkpoint, UserID: user.Name(), KeepAlive: false})
	if err != nil {
		return nil, fmt.Errorf("error syncing available files for user '%s' in server '%s/%s': %w",
			user.ID(), orgName, serverName, err)
	}

	var updates []models.Update
	for {
		update, err := stream.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, fmt.Errorf(
				"error received when reading from stream for user '%s' in server '%s/%s': %w",
				user.ID(), orgName, serverName, err)
		}

		updates = append(updates, models.Update{
			FileRef:    update.FileReference,
			Checkpoint: update.Checkpoint,
			ChangeType: toUpdateType(update.ChangeType),
			SizeBytes:  update.SizeBytes,
		})

	}

	return updates, nil
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

var _ Interface = (*Impl)(nil)
