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
	listQuery = "SELECT * FROM organizations"
	getQuery  = "SELECT * FROM organizations WHERE id = $1"
	addQuery  = "INSERT INTO organizations(name) VALUES ($1) RETURNING *"
	delQuery  = "DELETE FROM organizations WHERE id = $1"
)

// ErrNilDB is returned when constructing a postgresql-based repository with a nil connection
var ErrNilDB = errors.New("db cannot be nil")

// Organization is a postgres-compatible struct implementing models.Organization interface
type Organization struct {
	IDField   string `db:"id"`
	NameField string `db:"name"`
}

// ID returns the id of the organization
func (o *Organization) ID() string {
	return o.IDField
}

// Name returns the name of the organization
func (o *Organization) Name() string {
	return o.NameField
}

// OrganizationRepository is a mapping to a table in postgres that allows enables operations
// on organizations
type OrganizationRepository struct {
	db *sqlx.DB
}

// NewOrganizationRepository constructs a new postgresql-based organization repository
func NewOrganizationRepository(db *sqlx.DB) (*OrganizationRepository, error) {
	if db == nil {
		return nil, ErrNilDB
	}
	return &OrganizationRepository{db: db}, nil
}

// List returns a list of all the organizations
func (r *OrganizationRepository) List(ctx context.Context) ([]models.Organization, error) {
	var orgs []Organization
	err := r.db.SelectContext(ctx, &orgs, listQuery)
	if err != nil {
		return nil, fmt.Errorf("error executing organizations::list in postgres: %w", err)
	}

	ret := make([]models.Organization, len(orgs))
	for idx := range orgs {
		ret[idx] = &orgs[idx]
	}

	return ret, nil
}

// Get returns an organization that matches the supplied id
func (r *OrganizationRepository) Get(ctx context.Context, id string) (models.Organization, error) {
	var org Organization
	err := r.db.QueryRowxContext(ctx, getQuery, id).StructScan(&org)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, fmt.Errorf("error executing organizations::get in postgres: %w", err)
	}

	return &org, nil
}

// Add adds an organization with the supplied name
func (r *OrganizationRepository) Add(ctx context.Context, name string) (models.Organization, error) {
	var org Organization
	err := r.db.QueryRowxContext(ctx, addQuery, name).StructScan(&org)
	if err != nil {
		return nil, fmt.Errorf("error executing organizations::add in postgres: %w", err)
	}

	return &org, nil
}

// Remove deletes an organization that matches the supplied id
func (r *OrganizationRepository) Remove(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, delQuery, id)
	if err != nil {
		return fmt.Errorf("error executiong organizations::del in postgres: %w", err)
	}
	return nil
}
