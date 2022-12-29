package mongodb

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestPendingOAuthIntegration(t *testing.T) {

	ctx := context.Background()
	conn, err := mongo.Connect(ctx)
	assert.Nil(t, err)

	db := conn.Database("mifs_test_useraccounts")
	defer db.Drop(ctx)

	userID := primitive.NewObjectID()
	fs1ID := primitive.NewObjectID()

	repo := NewPendingOAuth2Repository(db)

	inserted , err := repo.Put(ctx, userID.Hex(), fs1ID.Hex(), "someState")
	assert.Nil(t, err)

	assert.Equal(t, inserted.FileServerID(), fs1ID.Hex())
	assert.Equal(t, inserted.UserID(), userID.Hex())
	assert.Equal(t, inserted.State(), "someState")

	popped, err := repo.Pop(ctx, "someState")
	assert.Nil(t, err)
	assert.Equal(t, inserted, popped)
}

