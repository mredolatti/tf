package psql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/mredolatti/tf/codigo/indexsrv/models"
	"github.com/mredolatti/tf/codigo/indexsrv/repository"

	"github.com/jmoiron/sqlx"
)

const (
	mappingListQuery       = "SELECT * FROM mappings WHERE user_id = $1"
	mappingListByPathQuery = "SELECT * FROM mappings WHERE user_id = $1 AND path <@ $2"
	mappingAddQuery        = "INSERT INTO mappings(user_id, server_id, path, ref, updated) VALUES ($1, $2, $3, $4, $5) RETURNING *"
	mappingDelQuery        = "DELETE FROM mappings WHERE user_id = $1 AND path = $2"
)

// Mapping is a postgres-compatible struct implementing models.Mapping interface
type Mapping struct {
	UserIDField   string
	ServerIDField string
	PathField     string
	RefField      string
	UpdatedField  time.Time
}

// UserID returns the if of the user who has an mapping in a file server
func (m *Mapping) UserID() string {
	return m.UserIDField
}

// FileServerID returns the id of the server in which the user has the mapping
func (m *Mapping) FileServerID() string {
	return m.ServerIDField
}

// Ref returns the internal reference to the file in the server
func (m *Mapping) Ref() string {
	return m.RefField
}

// Path returns the virtual path as seen by the user
func (m *Mapping) Path() string {
	return m.PathField
}

// Updated returns the time when this mapping was last updated
func (m *Mapping) Updated() time.Time {
	return m.UpdatedField
}

// MappingRepository is a mapping to a table in postgres that allows enables operations on mappings
type MappingRepository struct {
	db *sqlx.DB
}

// NewMappingRepository constructs a new postgresql-based mappings repository
func NewMappingRepository(db *sqlx.DB) (*MappingRepository, error) {
	if db == nil {
		return nil, ErrNilDB
	}
	return &MappingRepository{db: db}, nil
}

// List returns a list of all mappings for a specific user
func (r *MappingRepository) List(ctx context.Context, userID string) ([]models.Mapping, error) {
	var mappings []Mapping
	err := r.db.SelectContext(ctx, &mappings, mappingListQuery, userID)
	if err != nil {
		return nil, fmt.Errorf("error executing mappings::list in postgres: %w", err)
	}

	ret := make([]models.Mapping, len(mappings))
	for idx := range mappings {
		ret[idx] = &mappings[idx]
	}

	return ret, nil
}

// ListByPath returns all the mappings within a path for a specific user
func (r *MappingRepository) ListByPath(ctx context.Context, userID string, path string) (models.Mapping, error) {
	var mapping Mapping
	err := r.db.QueryRowxContext(ctx, mappingListByPathQuery, userID, path).StructScan(&mapping)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, fmt.Errorf("error executing mappings::get in postgres: %w", err)
	}

	return &mapping, nil
}

// Add adds a mapping
func (r *MappingRepository) Add(ctx context.Context, userID string, serverID string, ref string, path string, created time.Time) (models.Mapping, error) {
	var mapping Mapping
	err := r.db.QueryRowxContext(ctx, mappingAddQuery, userID, serverID, ref, path, created).StructScan(&mapping)
	if err != nil {
		return nil, fmt.Errorf("error executing users::add in postgres: %w", err)
	}
	return &mapping, nil
}

// Remove deletes a file server that matches the supplied id
func (r *MappingRepository) Remove(ctx context.Context, userID string, path string) error {
	_, err := r.db.ExecContext(ctx, mappingDelQuery, userID, path)
	if err != nil {
		return fmt.Errorf("error executiong users::del in postgres: %w", err)
	}
	return nil
}
