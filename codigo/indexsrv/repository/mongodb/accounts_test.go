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
	org1Name, fs1Name := "someOrg1", "someFS1"
	org2Name, fs2Name := "someOrg2", "someFS2"

	repo := NewUserAccountRepository(db)

	inserted1, err := repo.Add(ctx, userID.Hex(), org1Name, fs1Name, "someAccessToken", "someRefreshToken")
	assert.Nil(t, err)
	assert.Equal(t, inserted1.UserID(), userID.Hex())
	assert.Equal(t, inserted1.OrganizationName(), org1Name)
	assert.Equal(t, inserted1.FileServerName(), fs1Name)
	assert.Equal(t, inserted1.Token(), "someAccessToken")
	assert.Equal(t, inserted1.RefreshToken(), "someRefreshToken")
	assert.Equal(t, inserted1.Checkpoint(), int64(0))

	fetched1, err := repo.Get(ctx, userID.Hex(), org1Name, fs1Name)
	assert.Nil(t, err)
	assert.Equal(t, inserted1, fetched1)

	inserted2, err := repo.Add(ctx, userID.Hex(), org2Name, fs2Name, "someOtherAccessToken", "someOtherRefreshToken")
	assert.Nil(t, err)
	assert.Equal(t, inserted2.UserID(), userID.Hex())
	assert.Equal(t, inserted2.OrganizationName(), org2Name)
	assert.Equal(t, inserted2.FileServerName(), fs2Name)
	assert.Equal(t, inserted2.Token(), "someOtherAccessToken")
	assert.Equal(t, inserted2.RefreshToken(), "someOtherRefreshToken")
	assert.Equal(t, inserted2.Checkpoint(), int64(0))

	fetched2, err := repo.Get(ctx, userID.Hex(), org2Name, fs2Name)
	assert.Nil(t, err)
	assert.Equal(t, inserted2, fetched2)

	list, err := repo.List(ctx, userID.Hex())
	assert.Nil(t, err)
	assert.Contains(t, list, inserted1)
	assert.Contains(t, list, inserted2)

	err = repo.UpdateCheckpoint(ctx, userID.Hex(), org1Name, fs1Name, 123)
	assert.Nil(t, err)
	fetched1, err = repo.Get(ctx, userID.Hex(), org1Name, fs1Name)
	assert.Nil(t, err)
	assert.Equal(t, fetched1.Checkpoint(), int64(123))

	err = repo.UpdateTokens(ctx, userID.Hex(), org2Name, fs2Name, "newAccessToken", "newRefreshToken")
	assert.Nil(t, err)
	fetched2, err = repo.Get(ctx, userID.Hex(), org2Name, fs2Name)
	assert.Equal(t, fetched2.Token(), "newAccessToken")
	assert.Equal(t, fetched2.RefreshToken(), "newRefreshToken")
}
