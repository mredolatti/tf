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
	orgListQuery = "SELECT * FROM organizations"
	orgGetQuery  = "SELECT * FROM organizations WHERE id = $1"
	ogAddQuery   = "INSERT INTO organizations(name) VALUES ($1) RETURNING *"
	orgDelQuery  = "DELETE FROM organizations WHERE id = $1"
)

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
func NewOrganizationRepository(db *sqlx.DB) *OrganizationRepository {
	return &OrganizationRepository{db: db}
}

// List returns a list of all the organizations
func (r *OrganizationRepository) List(ctx context.Context) ([]models.Organization, error) {
	var orgs []Organization
	err := r.db.SelectContext(ctx, &orgs, orgListQuery)
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
	err := r.db.QueryRowxContext(ctx, orgGetQuery, id).StructScan(&org)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, fmt.Errorf("error executing organizations::get in postgres: %w", err)
	}

	return &org, nil
}

// Add adds an organization with the supplied name
func (r *OrganizationRepository) Add(ctx context.Context, org models.Organization) (models.Organization, error) {
	var read Organization
	err := r.db.QueryRowxContext(ctx, ogAddQuery, org.Name()).StructScan(&read)
	if err != nil {
		return nil, fmt.Errorf("error executing organizations::add in postgres: %w", err)
	}

	return &read, nil
}

// Remove deletes an organization that matches the supplied id
func (r *OrganizationRepository) Remove(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, orgDelQuery, id)
	if err != nil {
		return fmt.Errorf("error executiong organizations::del in postgres: %w", err)
	}
	return nil
}

var _ repository.OrganizationRepository = (*OrganizationRepository)(nil)
