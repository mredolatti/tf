package mapper

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
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
	Update(ctx context.Context, userID string, mapping models.Mapping) (models.Mapping, error)
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
		return nil, fmt.Errorf("update required but failed: %w", err)
	}
	return i.mappings.List(ctx, userID, *query)
}

// Update updates a mapping
func (i *Impl) Update(ctx context.Context, userName string, mapping models.Mapping) (models.Mapping, error) {
	return nil, nil
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
	var errCount int64
	for _, account := range forUser {
		if account.Checkpoint() < thresholdNS || force {
			wg.Add(1)
			go func(acc models.UserAccount) {
				defer wg.Done()

				fmt.Println("pidiendo con: ", acc.FileServerID(), userID, acc.Checkpoint())
				updates, err := i.serverLinks.FetchUpdates(ctx, acc.FileServerID(), user, acc.Checkpoint())
				fmt.Printf("updates: %+v\n", updates)
				if err != nil {
					// TODO(mredolatti): Log!
					atomic.AddInt64(&errCount, 1)
				}

				err = i.handleUpdates(ctx, acc, updates)
				if err != nil {
					fmt.Println("EEEEE: ", err)
					// TODO(mredolatti): Log!
					atomic.AddInt64(&errCount, 1)
				}
			}(account)
		}
	}
	wg.Wait()

	if errCount > 0 {
		return fmt.Errorf("%d accounts filed to sync", atomic.LoadInt64(&errCount))
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

	if err := i.mappings.HandleServerUpdates(ctx, account.UserID(), updates); err != nil {
		return fmt.Errorf("error adding/updating valid mappings: %w", err)
	}

	if err := i.accounts.UpdateCheckpoint(ctx, account.UserID(), account.FileServerID(), newCheckpoint); err != nil {
		return fmt.Errorf("error updating checkpoint: %w", err)
	}

	return nil
}

var _ Interface = (*Impl)(nil)
