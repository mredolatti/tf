package mapper

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/mredolatti/tf/codigo/indexsrv/fslinks"
	"github.com/mredolatti/tf/codigo/indexsrv/models"
	"github.com/mredolatti/tf/codigo/indexsrv/repository"
)

const (
	defaultUpdateTolerance = 1 * time.Hour
)

// Config parameters to configure the mapper
type Config struct {
	LastUpdateTolerance time.Duration
	Users               repository.UserRepository
	Repo                repository.MappingRepository
	Accounts            repository.UserAccountRepository
	ServerLinks         fslinks.Interface
}

// Interface defines the set of methods exposed by a Mapper
type Interface interface {
	Get(ctx context.Context, userID string, forceUpdate bool, query *models.MappingQuery) ([]models.Mapping, error)
	AddPath(ctx context.Context, userID, org, server, ref, newPath string) (models.Mapping, error)
	UpdatePathByID(ctx context.Context, userID, id, newPath string) (models.Mapping, error)
    ResetPathByID(ctx context.Context, userID, id string) error
}

// Impl implements the Mapper interface
type Impl struct {
	mappings    repository.MappingRepository
	accounts    repository.UserAccountRepository
	users       repository.UserRepository
	serverLinks fslinks.Interface
}

// New constructs a new Mapper
func New(config Config) *Impl {
	return &Impl{
		mappings:    config.Repo,
		accounts:    config.Accounts,
		serverLinks: config.ServerLinks,
		users:       config.Users,
	}
}

// Get fetches mappings for a specific user based on a query
func (i *Impl) Get(ctx context.Context, userID string, forceUpdate bool, query *models.MappingQuery) ([]models.Mapping, error) {

	if query == nil {
		query = &models.MappingQuery{}
	}

	err := i.ensureUpdated(ctx, userID, forceUpdate)
	if err != nil {
		return nil, err // do not wrap to preserve underlying error type
	}
	return i.mappings.List(ctx, userID, *query)
}

func (i *Impl) AddPath(ctx context.Context, userID, org, server, ref, newPath string) (models.Mapping, error) {
    return i.mappings.AddPath(ctx, userID, org, server, ref, newPath)
}

func (i *Impl) UpdatePathByID(ctx context.Context, userID, id, newPath string) (models.Mapping, error) {
    return i.mappings.UpdatePathByID(ctx, userID, id, newPath)
}

func (i *Impl) ResetPathByID(ctx context.Context, userID, id string) error {
    return i.mappings.RemovePathByID(ctx, userID, id)
}

func (i *Impl) ensureUpdated(ctx context.Context, userID string, force bool) error {


	user, err := i.users.Get(ctx, userID)
	if err != nil {
		return fmt.Errorf("error fetching user information: %w", err)
	}
	

	forUser, err := i.accounts.List(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to fetch user accounts for userID=%s: %w", userID, err)
	}

	thresholdNS := time.Now().Add(-defaultUpdateTolerance).UnixNano()

	var wg sync.WaitGroup
    multiErr := newMultiSyncErr()
	for _, account := range forUser {
		if account.Checkpoint() < thresholdNS || force {
			wg.Add(1)
			go func(acc models.UserAccount) {
				defer wg.Done()
				updates, err := i.serverLinks.FetchUpdates(ctx, acc.OrganizationName(), acc.FileServerName(), user, acc.Checkpoint())
				if err != nil {
                    multiErr.Add(acc.OrganizationName(), acc.FileServerName(), err)
                    return
				}

				err = i.handleUpdates(ctx, acc, updates)
				if err != nil {
                    multiErr.Add(acc.OrganizationName(), acc.FileServerName(), err)
                    return
				}
			}(account)
		}
	}
	wg.Wait()

    if multiErr.HasErrors() {
        return multiErr
    }

	return nil
}

func (i *Impl) handleUpdates(ctx context.Context, account models.UserAccount, updates []models.Update) error {

	if len(updates) == 0 {
		return nil
	}

	var newCheckpoint int64
	for _, update := range updates {
		if current := update.Checkpoint; current > newCheckpoint {
			newCheckpoint = current
		}
	}

	if err := i.mappings.HandleServerUpdates(ctx, account.UserID(), account.OrganizationName(), account.FileServerName(), updates); err != nil {
		return fmt.Errorf("error adding/updating valid mappings: %w", err)
	}

	if err := i.accounts.UpdateCheckpoint(ctx, account.UserID(), account.OrganizationName(), account.FileServerName(), newCheckpoint); err != nil {
		return fmt.Errorf("error updating checkpoint: %w", err)
	}

	return nil
}

var _ Interface = (*Impl)(nil)
