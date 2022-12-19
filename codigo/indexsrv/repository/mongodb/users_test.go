package mongodb

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestIntegration(t *testing.T) {

	ctx := context.Background()
	conn, err := mongo.Connect(ctx)
	if err != nil {
		t.Fatal("could not connect to mongo: " + err.Error())
	}
	db := conn.Database("mifs")

	repo := NewUserRepository(db)

	userInserted, err := repo.Add(ctx, "", "user1", "a@b.com", "someAccess", "someRefresh")
	if err != nil {
		t.Error("error inserting: ", err)
	}

	assert.Equal(t, userInserted.Name(), "user1", "shold be equal")
	assert.Equal(t, userInserted.Email(), "a@b.com", "shold be equal")
	assert.Equal(t, userInserted.AccessToken(), "someAccess", "shold be equal")
	assert.Equal(t, userInserted.RefreshToken(), "someRefresh", "shold be equal")

	userFetched, err := repo.Get(ctx, userInserted.ID())
	if err != nil {
		t.Error("error inserting: ", err)
	}

		
	assert.Equal(t, userInserted, userFetched, "should be equal")

}

