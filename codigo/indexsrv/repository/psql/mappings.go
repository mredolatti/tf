package psql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/mredolatti/tf/codigo/indexsrv/models"
	"github.com/mredolatti/tf/codigo/indexsrv/repository"

	"github.com/jmoiron/sqlx"
)

const (
	_all                    = " RETURNING *"
	mappingListQuery        = "SELECT * FROM mappings WHERE user_id = $1"
	mappingListByPathQuery  = "SELECT * FROM mappings WHERE user_id = $1 AND path <@ $2"
	_mappingAddTpl          = "INSERT INTO mappings(user_id, server_id, path, ref, updated, deleted) VALUES ($1, $2, $3, $4, $5, $6)"
	mappingAddQuery         = _mappingAddTpl + _all
	mappingAddOrUpdateQuery = ("INSERT INTO mappings(user_id, server_id, path, ref, updated, deleted) " +
		"VALUES (:user_id, :server_id, :path, :ref, :updated, :deleted) " +
		"ON CONFLICT (user_id, server_id, ref) DO UPDATE SET updated=EXCLUDED.updated, deleted=EXCLUDED.deleted")
	mappingDelQuery = "DELETE FROM mappings WHERE user_id = $1 AND path = $2"
)

var formatterForDisplay *strings.Replacer = strings.NewReplacer(
	".", "/",
	"__DOT__", ".",
)

var formatterForStorage *strings.Replacer = strings.NewReplacer(
	"/", ".",
	".", "__DOT__",
)

// Mapping is a postgres-compatible struct implementing models.Mapping interface
type Mapping struct {
	UserIDField   string `db:"user_id"`
	ServerIDField string `db:"server_id"`
	PathField     string `db:"path"`
	RefField      string `db:"ref"`
	DeletedField  bool   `db:"deleted"`
	UpdatedField  int64  `db:"updated"`
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
	return formatterForDisplay.Replace(m.PathField)
}

// Updated returns the time when this mapping was last updated
func (m *Mapping) Updated() time.Time {
	return time.Unix(0, m.UpdatedField)
}

// Deleted returns whether the mapping references a no-longer available file or not
func (m *Mapping) Deleted() bool {
	return m.DeletedField
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
func (r *MappingRepository) List(ctx context.Context, userID string, query models.MappingQuery) ([]models.Mapping, error) {
	// TODO: User or stop accepting query
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
func (r *MappingRepository) Add(ctx context.Context, userID string, mapping models.Mapping) (models.Mapping, error) {
	var fetched Mapping
	err := r.db.QueryRowxContext(
		ctx,
		mappingAddQuery,
		userID,
		mapping.FileServerID(),
		mapping.Ref(),
		formatterForDisplay.Replace(mapping.Path()),
		mapping.Updated(),
		mapping.Deleted(),
	).StructScan(&fetched)
	if err != nil {
		return nil, fmt.Errorf("error executing users::add in postgres: %w", err)
	}
	return &fetched, nil
}

// HandleServerUpdates adds/updates mappings from an incoming set of changes from a file server
func (r *MappingRepository) HandleServerUpdates(ctx context.Context, userID string, updates []models.Update) error {
	if _, err := r.db.NamedExecContext(ctx, mappingAddOrUpdateQuery, mappingsFromUpdates(userID, updates)); err != nil {
		return fmt.Errorf("error inserting/updating mappings: %w", err)
	}
	return nil
}

// Update TODO!
func (r *MappingRepository) Update(ctx context.Context, userID string, mappingID string, mapping models.Mapping) (models.Mapping, error) {
	// TODO!
	return nil, nil
}

// Remove deletes a file server that matches the supplied id
func (r *MappingRepository) Remove(ctx context.Context, userID string, path string) error {
	_, err := r.db.ExecContext(ctx, mappingDelQuery, userID, path)
	if err != nil {
		return fmt.Errorf("error executiong users::del in postgres: %w", err)
	}
	return nil
}

func mappingsFromUpdates(userID string, updates []models.Update) []Mapping {
	mappings := make([]Mapping, 0, len(updates))
	for _, update := range updates {
		mappings = append(mappings, Mapping{
			UserIDField:   userID,
			ServerIDField: update.ServerID,
			RefField:      update.FileRef,
			DeletedField:  update.ChangeType == models.UpdateTypeFileDelete,
			UpdatedField:  update.Checkpoint,
			PathField:     formatterForStorage.Replace(fmt.Sprintf("unnasigned/%s/%s", update.ServerID, update.FileRef)),
		})
	}

	return mappings
}

var _ repository.MappingRepository = (*MappingRepository)(nil)
