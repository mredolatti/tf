package mapper

import (
	"context"

	"github.com/mredolatti/tf/codigo/indexsrv/models"
	"github.com/mredolatti/tf/codigo/indexsrv/repository"
)

type Interface interface {
	Get(ctx context.Context, userID string, query *models.MappingQuery) ([]models.Mapping, error)
}

type Impl struct {
	repo repository.MappingRepository
}

func New(repo repository.MappingRepository) *Impl {
	return &Impl{repo: repo}
}

func (i *Impl) Get(ctx context.Context, userID string, query *models.MappingQuery) ([]models.Mapping, error) {
	if query == nil {
		return i.repo.List(ctx, userID, models.MappingQuery{})
	}
	return i.repo.List(ctx, userID, *query)
}

var _ Interface = (*Impl)(nil)
