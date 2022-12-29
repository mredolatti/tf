package psql

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/jackc/pgx/v4/stdlib"

	"github.com/mredolatti/tf/codigo/common/config"
	"github.com/mredolatti/tf/codigo/indexsrv/repository"
)

type Factory struct {
	db *sqlx.DB
}

// Accounts implements repository.Factory
func (f *Factory) Accounts() repository.UserAccountRepository {
	return NewUserAccountRepository(f.db)
}

// FileServers implements repository.Factory
func (f *Factory) FileServers() repository.FileServerRepository {
	return NewFileServerRepository(f.db)
}

// Mappings implements repository.Factory
func (f *Factory) Mappings() repository.MappingRepository {
	return NewMappingRepository(f.db)
}

// Organizations implements repository.Factory
func (f *Factory) Organizations() repository.OrganizationRepository {
	return NewOrganizationRepository(f.db)
}

// PendingOAuth implements repository.Factory
func (f *Factory) PendingOAuth() repository.PendingOAuth2Repository {
	return NewPendingOAuth2Repository(f.db)
}

// Users implements repository.Factory
func (f *Factory) Users() repository.UserRepository {
	return NewUserRepository(f.db)
}

func NewFactory(cfg *config.Postgres) (*Factory, error) {
	db, err := sqlx.Connect("pgx", fmt.Sprintf("postgres://%s:%s@%s:%d/%s", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DB))
	if err != nil {
		return nil, fmt.Errorf("error connecting to postgres: %w", err)
	}
	return &Factory{db: db}, nil
}

var _ repository.Factory = (*Factory)(nil)
