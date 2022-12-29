package mongodb

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/mredolatti/tf/codigo/common/config"
	"github.com/mredolatti/tf/codigo/indexsrv/repository"
)

type Factory struct {
	conn *mongo.Client
	db   *mongo.Database
}

func NewFactory(cfg *config.Mongo) (*Factory, error) {
	opts := &options.ClientOptions{Hosts: cfg.Hosts}
	if len(cfg.User) > 0 && len(cfg.Password) > 0 {
		opts.Auth = &options.Credential{
			AuthMechanism: "PLAIN",
			Username:      cfg.User,
			Password:      cfg.Password,
		}
	}
	client, err := mongo.Connect(context.Background(), opts)
	if err != nil {
		return nil, fmt.Errorf("error instantiating mongo client: %w", err)
	}

	return &Factory{
		conn: client,
		db:   client.Database(cfg.DB),
	}, nil
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

var _ repository.Factory = (*Factory)(nil)
