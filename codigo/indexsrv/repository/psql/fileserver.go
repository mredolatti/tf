package psql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/mredolatti/tf/codigo/indexsrv/models"
	"github.com/mredolatti/tf/codigo/indexsrv/repository"

	"github.com/jmoiron/sqlx"
)

const (
	fsListByOrg = "SELECT * FROM file_servers WHERE org_id = $1"
	fsGetQuery  = "SELECT * FROM file_servers WHERE id = $1"
	fsAddQuery  = "INSERT INTO file_servers(name, org_id, auth_url, fetch_url) VALUES ($1, $2, $3, $4) RETURNING *"
	fsDelQuery  = "DELETE FROM file_servers WHERE id = $1"
)

// FileServer is a postgres-compatible struct implementing models.FileServer interface
type FileServer struct {
	IDField       int64  `db:"id"`
	NameField     string `db:"name"`
	OrgField      string `db:"org_id"`
	AuthURLField  string `db:"auth_url"`
	FetchURLField string `db:"fetch_url"`
}

// ID returns the id of the file server
func (f *FileServer) ID() int64 {
	return f.IDField
}

// Name returns the name of the file server
func (f *FileServer) Name() string {
	return f.NameField
}

// OrganizationID returns the ID of the organization this server belongs to
func (f *FileServer) OrganizationID() string {
	return f.OrgField
}

// AuthURL returns the URL used to authorize users when linking a server to their account
func (f *FileServer) AuthURL() string {
	return f.AuthURLField
}

// FetchURL returns the URL to be used when returning fetch recipes
func (f *FileServer) FetchURL() string {
	return f.FetchURLField
}

// FileServerRepository is a mapping to a table in postgres that allows enables operations
// on file server
type FileServerRepository struct {
	db *sqlx.DB
}

// NewFileServerRepository constructs a new postgresql-based file server repository
func NewFileServerRepository(db *sqlx.DB) (*FileServerRepository, error) {
	if db == nil {
		return nil, ErrNilDB
	}
	return &FileServerRepository{db: db}, nil
}

// List returns a list of all the file server
func (r *FileServerRepository) List(ctx context.Context, orgID string) ([]models.FileServer, error) {
	var servers []FileServer
	err := r.db.SelectContext(ctx, &servers, fsListByOrg, orgID)
	if err != nil {
		return nil, fmt.Errorf("error executing file_servers::list in postgres: %w", err)
	}

	ret := make([]models.FileServer, len(servers))
	for idx := range servers {
		ret[idx] = &servers[idx]
	}

	return ret, nil
}

// Get returns an file server that matches the supplied id
func (r *FileServerRepository) Get(ctx context.Context, id int64) (models.FileServer, error) {
	var server FileServer
	err := r.db.QueryRowxContext(ctx, fsGetQuery, id).StructScan(&server)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, fmt.Errorf("error executing file_servers::get in postgres: %w", err)
	}

	return &server, nil
}

// Add adds a file server with the supplied name
func (r *FileServerRepository) Add(ctx context.Context, name string, orgID string, authURL string, fetchURL string) (models.FileServer, error) {
	var server FileServer
	err := r.db.QueryRowxContext(ctx, fsAddQuery, name, orgID, authURL, fetchURL).StructScan(&server)
	if err != nil {
		return nil, fmt.Errorf("error executing file_server::add in postgres: %w", err)
	}
	return &server, nil
}

// Remove deletes a file server that matches the supplied id
func (r *FileServerRepository) Remove(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, fsDelQuery, id)
	if err != nil {
		return fmt.Errorf("error executiong file_server::del in postgres: %w", err)
	}
	return nil
}