package mongodb

import (
	"context"
	"testing"

	"github.com/mredolatti/tf/codigo/indexsrv/repository"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestOrganizationIntegration(t *testing.T) {

	ctx := context.Background()
	conn, err := mongo.Connect(ctx)
	assert.Nil(t, err)

	db := conn.Database("mifs_test_organizations")
	defer db.Drop(ctx)

	repo := NewOrganizationRepository(db)

	// Insertion
	inserted1, err := repo.Add(ctx, &Organization{NameField: "someOrg1"})
	assert.Nil(t, err)
	inserted2, err := repo.Add(ctx, &Organization{NameField: "someOrg2"})
	assert.Nil(t, err)
	inserted3, err := repo.Add(ctx, &Organization{NameField: "someOrg3"})
	assert.Nil(t, err)
	inserted4, err := repo.Add(ctx, &Organization{NameField: "someOrg4"})
	assert.Nil(t, err)

	assert.Equal(t, inserted1.Name(), "someOrg1")
	assert.Equal(t, inserted2.Name(), "someOrg2")
	assert.Equal(t, inserted3.Name(), "someOrg3")
	assert.Equal(t, inserted4.Name(), "someOrg4")

	// Fetching
	fetched1, err := repo.Get(ctx, inserted1.ID())
	assert.Nil(t, err)
	assert.Equal(t, inserted1, fetched1)

	// Listing
	list, err := repo.List(ctx)
	assert.Nil(t, err)
	assert.Contains(t, list, inserted1)
	assert.Contains(t, list, inserted2)
	assert.Contains(t, list, inserted3)
	assert.Contains(t, list, inserted4)


	// Removing
	err = repo.Remove(ctx, inserted1.ID())
	assert.Nil(t, err)
	fetchedAfterRemoval, err := repo.Get(ctx, inserted1.ID())
	assert.Nil(t, fetchedAfterRemoval)
	assert.Equal(t, err, repository.ErrNotFound)
}

