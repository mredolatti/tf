package mongodb

import (
	"context"
	"testing"

	"github.com/mredolatti/tf/codigo/indexsrv/repository"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestUserIntegration(t *testing.T) {

	ctx := context.Background()
	conn, err := mongo.Connect(ctx)
	assert.Nil(t, err)

	db := conn.Database("mifs_test_users")
	defer db.Drop(ctx)

	repo := NewUserRepository(db)

	// Insertion
	userInserted, err := repo.Add(ctx, "", "user1", "a@b.com", "someHashedPass")
	assert.Nil(t, err)

	assert.Equal(t, userInserted.Name(), "user1")
	assert.Equal(t, userInserted.Email(), "a@b.com")
	assert.Equal(t, userInserted.PasswordHash(), "someHashedPass")

	// Fetching
	userFetched, err := repo.Get(ctx, userInserted.ID())
	assert.Nil(t, err)
	assert.Equal(t, userInserted, userFetched, "should be equal")

	// Updating tokens
	updated, err := repo.UpdatePassword(ctx, userFetched.ID(), "newHashedPass")
	assert.Nil(t, err)
	oid, _ := primitive.ObjectIDFromHex(userFetched.ID())
	assert.Equal(t, updated, &User{
		IDField: oid,
		NameField: userFetched.Name(),
		EmailField: updated.Email(),
		PasswordHashField: "newHashedPass",
	})
	fetchedAfterUpdate, err := repo.Get(ctx, userInserted.ID())
	assert.Nil(t, err)
	assert.Equal(t, fetchedAfterUpdate, updated)

	// Removing
	err = repo.Remove(ctx, userFetched.ID())
	assert.Nil(t, err)
	fetchedAfterRemoval, err := repo.Get(ctx, userFetched.ID())
	assert.Nil(t, fetchedAfterRemoval)
	assert.Equal(t, err, repository.ErrNotFound)
}

