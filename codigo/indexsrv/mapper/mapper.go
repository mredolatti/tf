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

// Options parameters to configure the mapper
type Options struct {
	LastUpdateTolerance time.Duration
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
	serverLinks fslinks.Interface
}

// New constructs a new Mapper
func New(repo repository.MappingRepository) *Impl {
	return &Impl{mappings: repo}
}

// Get fetches mappings for a specific user based on a query
func (i *Impl) Get(ctx context.Context, userID string, forceUpdate bool, query *models.MappingQuery) ([]models.Mapping, error) {

	if query == nil {
		return i.mappings.List(ctx, userID, models.MappingQuery{})
	}
	return i.mappings.List(ctx, userID, *query)
}

// Update updates a mapping
func (i *Impl) Update(ctx context.Context, userID string, mapping models.Mapping) (models.Mapping, error) {
	return nil, nil
}

func (i *Impl) ensureUpdated(ctx context.Context, userID string, force bool) error {
	forUser, err := i.accounts.List(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to fetch user accounts for userID=%s: %w", userID, err)
	}

	thresholdNS := time.Now().Add(-defaultUpdateTolerance).UnixNano()

	var wg sync.WaitGroup
	var errCount int64
	for _, account := range forUser {
		if account.Checkpoint() < thresholdNS {
			go func(acc models.UserAccount) {
				wg.Add(1)
				defer wg.Done()
				updates, err := i.serverLinks.FetchUpdates(ctx, acc.UserID(), acc.FileServerID(), acc.Checkpoint())
				if err != nil {
					// TODO(mredolatti): Log!
					atomic.AddInt64(&errCount, 1)
				}

				err = i.handleUpdates(ctx, acc, updates)
				if err != nil {
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

	toArchive := make([]models.Update, 0, len(updates))
	toPublish := make([]models.Update, 0, len(updates))

	var newCheckpoint int64
	for _, update := range updates {
		switch update.ChangeType {
		case models.UpdateTypeFileAdd, models.UpdateTypeFileUpdate:
			toPublish = append(toPublish, update)
		case models.UpdateTypeFileDelete:
			toArchive = append(toArchive, update)
		}

		if current := update.Checkpoint; current > newCheckpoint {
			newCheckpoint = current
		}
	}

	if err := i.mappings.HandleServerUpdates(ctx, account.UserID(), toPublish); err != nil {
		return fmt.Errorf("error adding/updating valid mappings: %w", err)
	}

	if err := i.accounts.UpdateCheckpoint(ctx, account.UserID(), account.FileServerID(), newCheckpoint); err != nil {
		return fmt.Errorf("error updating checkpoint: %w", err)
	}
	// @}

	return nil
}

var _ Interface = (*Impl)(nil)
