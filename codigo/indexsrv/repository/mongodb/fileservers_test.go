package mongodb

import (
	"context"
	"testing"

	"github.com/mredolatti/tf/codigo/indexsrv/repository"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestFileServersIntegration(t *testing.T) {

	ctx := context.Background()
	conn, err := mongo.Connect(ctx)
	assert.Nil(t, err)

	db := conn.Database("mifs_test_fileservers")
	defer db.Drop(ctx)

	repo := NewFileServerRepository(db)

	oid1 := primitive.NewObjectID().Hex()
	oid2 := primitive.NewObjectID().Hex()

	// Insertion
	inserted1, err := repo.Add(ctx, "", "fs1", oid1, "http://auth.org1.io", "http://token.org1.io", "http://fetch.org1.io", "ctrl1")
	assert.Nil(t, err)
	inserted2, err := repo.Add(ctx, "", "fs2", oid1, "http://auth.org2.io", "http://token.org2.io", "http://fetch.org2.io", "ctrl2")
	assert.Nil(t, err)
	inserted3, err := repo.Add(ctx, "", "fs3", oid2, "http://auth.org3.io", "http://token.org3.io", "http://fetch.org3.io", "ctrl3")
	assert.Nil(t, err)
	inserted4, err := repo.Add(ctx, "", "fs4", oid2, "http://auth.org4.io", "http://token.org4.io", "http://fetch.org4.io", "ctrl4")
	assert.Nil(t, err)

	assert.Equal(t, inserted1.Name(), "fs1")
	assert.Equal(t, inserted2.Name(), "fs2")
	assert.Equal(t, inserted3.Name(), "fs3")
	assert.Equal(t, inserted4.Name(), "fs4")

	// Fetching
	fetched1, err := repo.Get(ctx, inserted1.ID())
	assert.Nil(t, err)
	assert.Equal(t, inserted1, fetched1)

	// Listing
	list1, err := repo.List(ctx, oid1)
	assert.Nil(t, err)
	assert.Contains(t, list1, inserted1)
	assert.Contains(t, list1, inserted2)

	list2, err := repo.List(ctx, oid2)
	assert.Nil(t, err)
	assert.Contains(t, list2, inserted3)
	assert.Contains(t, list2, inserted4)

	// Removing
	err = repo.Remove(ctx, inserted1.ID())
	assert.Nil(t, err)
	fetchedAfterRemoval, err := repo.Get(ctx, inserted1.ID())
	assert.Nil(t, fetchedAfterRemoval)
	assert.Equal(t, err, repository.ErrNotFound)
}

