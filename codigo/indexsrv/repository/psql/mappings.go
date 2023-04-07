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
	_mappingAddTpl          = "INSERT INTO mappings(user_id, organization_name, server_name, size_bytes, path, ref, updated) VALUES ($1, $2, $3, $4, $5, $6)"
	mappingAddQuery         = _mappingAddTpl + _all
	mappingAddOrUpdateQuery = ("INSERT INTO mappings(user_id, organization_name, server_name, size_bytes, path, ref, updated, deleted) " +
		"VALUES (:user_id, :organization_name, :server_name, :path, :ref, :updated, :deleted) " +
		"ON CONFLICT (user_id, organization_name, server_name, ref) DO UPDATE SET updated=EXCLUDED.updated, size_bytes=EXCLUDED.size_bytes, deleted=EXCLUDED.deleted")
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
	IDField         string `db:"id"`
	UserIDField     string `db:"user_id"`
	OrgNameField    string `db:"organization_name"`
	ServerNameField string `db:"server_name"`
	SizeBytesField  int64  `db:"size_bytes"`
	PathField       string `db:"path"`
	RefField        string `db:"ref"`
	UpdatedField    int64  `db:"updated"`
	DeletedField    bool   `db:"deleted"`
}

func (m *Mapping) ID() string {
    return m.IDField
}

// UserID returns the if of the user who has an mapping in a file server
func (m *Mapping) UserID() string {
	return m.UserIDField
}

func (m *Mapping) OrganizationName() string {
	return m.OrganizationName()
}

func (m *Mapping) ServerName() string {
	return m.ServerNameField
}

func (m *Mapping) SizeBytes() int64 {
	return m.SizeBytesField
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
func NewMappingRepository(db *sqlx.DB) *MappingRepository {
	return &MappingRepository{db: db}
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
		mapping.OrganizationName(),
		mapping.ServerName(),
		mapping.SizeBytes(),
		formatterForStorage.Replace(mapping.Path()),
		mapping.Ref(),
		mapping.Updated().UnixNano(),
	).StructScan(&fetched)
	if err != nil {
		return nil, fmt.Errorf("error executing users::add in postgres: %w", err)
	}
	return &fetched, nil
}

// HandleServerUpdates adds/updates mappings from an incoming set of changes from a file server
func (r *MappingRepository) HandleServerUpdates(ctx context.Context, userID string, orgName string, serverName string, updates []models.Update) error {
	if _, err := r.db.NamedExecContext(ctx, mappingAddOrUpdateQuery, mappingsFromUpdates(userID, orgName, serverName, updates)); err != nil {
		return fmt.Errorf("error inserting/updating mappings: %w", err)
	}
	return nil
}

// Update TODO!
func (r *MappingRepository) AddPath(
	ctx context.Context,
	userID string,
	org string,
	server string,
	ref string,
	newPath string,
) (models.Mapping, error) {
	// TODO!
	return nil, nil
}

func (r *MappingRepository) UpdatePathByID(ctx context.Context, userID string, id string, newPath string) (models.Mapping, error) {
	// TODO
	return nil, nil
}

func (r *MappingRepository) RemovePathByID(ctx context.Context, userID string, id string) error {
	// TODO
	return nil
}

// Remove deletes a file server that matches the supplied id
func (r *MappingRepository) Remove(ctx context.Context, userID string, path string) error {
	_, err := r.db.ExecContext(ctx, mappingDelQuery, userID, path)
	if err != nil {
		return fmt.Errorf("error executiong users::del in postgres: %w", err)
	}
	return nil
}

func mappingsFromUpdates(userID string, orgName string, serverName string, updates []models.Update) []Mapping {
	mappings := make([]Mapping, 0, len(updates))
	for _, update := range updates {
		mappings = append(mappings, Mapping{
			UserIDField:     userID,
			OrgNameField:    orgName,
			ServerNameField: serverName,
			SizeBytesField:  update.SizeBytes,
			RefField:        update.FileRef,
			DeletedField:    update.ChangeType == models.UpdateTypeFileDelete,
			UpdatedField:    update.Checkpoint,
			PathField:       formatterForStorage.Replace(update.UnmappedPath(orgName, serverName)),
		})
	}

	return mappings
}

var _ repository.MappingRepository = (*MappingRepository)(nil)
var _ models.Mapping = (*Mapping)(nil)
