package mongodb

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestUserAccountsIntegration(t *testing.T) {

	ctx := context.Background()
	conn, err := mongo.Connect(ctx)
	assert.Nil(t, err)

	db := conn.Database("mifs_test_useraccounts")
	defer db.Drop(ctx)

	userID := primitive.NewObjectID()
	fs1ID := primitive.NewObjectID()
	fs2ID := primitive.NewObjectID()

	repo := NewUserAccountRepository(db)

	inserted1, err := repo.Add(ctx, userID.Hex(), fs1ID.Hex(), "someAccessToken", "someRefreshToken")
	assert.Nil(t, err)
	assert.Equal(t, inserted1.UserID(), userID.Hex())
	assert.Equal(t, inserted1.FileServerID(), fs1ID.Hex())
	assert.Equal(t, inserted1.Token(), "someAccessToken")
	assert.Equal(t, inserted1.RefreshToken(), "someRefreshToken")
	assert.Equal(t, inserted1.Checkpoint(), int64(0))

	fetched1, err := repo.Get(ctx, userID.Hex(), fs1ID.Hex())
	assert.Nil(t, err)
	assert.Equal(t, inserted1, fetched1)

	inserted2, err := repo.Add(ctx, userID.Hex(), fs2ID.Hex(), "someOtherAccessToken", "someOtherRefreshToken")
	assert.Nil(t, err)
	assert.Equal(t, inserted2.UserID(), userID.Hex())
	assert.Equal(t, inserted2.FileServerID(), fs2ID.Hex())
	assert.Equal(t, inserted2.Token(), "someOtherAccessToken")
	assert.Equal(t, inserted2.RefreshToken(), "someOtherRefreshToken")
	assert.Equal(t, inserted2.Checkpoint(), int64(0))

	fetched2, err := repo.Get(ctx, userID.Hex(), fs2ID.Hex())
	assert.Nil(t, err)
	assert.Equal(t, inserted2, fetched2)

	list, err := repo.List(ctx, userID.Hex())
	assert.Nil(t, err)
	assert.Contains(t, list, inserted1)
	assert.Contains(t, list, inserted2)

	err = repo.UpdateCheckpoint(ctx, userID.Hex(), fs1ID.Hex(), 123)
	assert.Nil(t, err)
	fetched1, err = repo.Get(ctx, userID.Hex(), fs1ID.Hex())
	assert.Nil(t, err)
	assert.Equal(t, fetched1.Checkpoint(), int64(123))

	err = repo.UpdateTokens(ctx, userID.Hex(), fs2ID.Hex(), "newAccessToken", "newRefreshToken")
	assert.Nil(t, err)
	fetched2, err = repo.Get(ctx, userID.Hex(), fs2ID.Hex())
	assert.Equal(t, fetched2.Token(), "newAccessToken")
	assert.Equal(t, fetched2.RefreshToken(), "newRefreshToken")
}
