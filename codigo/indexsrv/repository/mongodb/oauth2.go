package mongodb

import (
	"context"
	"fmt"

	"github.com/mredolatti/tf/codigo/indexsrv/models"
	"github.com/mredolatti/tf/codigo/indexsrv/repository"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// PendingOAuth2 represents an in-progress oauth2 flow
type PendingOAuth2 struct {
	IDField           primitive.ObjectID
	UserIDField       primitive.ObjectID `bson:"userId"`
	FileServerIDField primitive.ObjectID `bson:"serverId"`
	StateField        string `bson:"state"`
}

// FileServerID returns the if of the file server we're trying au authenticate in
func (p *PendingOAuth2) FileServerID() string {
	return p.FileServerIDField.Hex()
}

// UserID returns the user we're trying to authenticate
func (p *PendingOAuth2) UserID() string {
	return p.UserIDField.Hex()
}

// State returns the randomized-code used to secure the request (and map to user_id, server_id)
func (p *PendingOAuth2) State() string {
	return p.StateField
}

// PendingOAuth2Repository is a postgres-based implementation of an in-progress oauth2 flow repository
type PendingOAuth2Repository struct {
	collection *mongo.Collection
}

// NewPendingOAuth2Repository constructs a new PendingOAuth2Repository
func NewPendingOAuth2Repository(db *mongo.Database) *PendingOAuth2Repository {
	return &PendingOAuth2Repository{collection: db.Collection("PendingOAuth2")}
}

// Put starts tracking a new flow
func (r *PendingOAuth2Repository) Put(ctx context.Context, userID string, serverID string, state string) (models.PendingOAuth2, error) {
	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("error constructing objectID for user with id=%s: %w", userID, err)
	}
	fsid, err := primitive.ObjectIDFromHex(serverID)
	if err != nil {
		return nil, fmt.Errorf("error constructing objectID for server with id=%s: %w", serverID, err)
	}

	toInsert := &PendingOAuth2{
		IDField: primitive.NewObjectID(),
		FileServerIDField: fsid,
		UserIDField: uid,
		StateField: state,
	}

	_, err = r.collection.InsertOne(ctx, toInsert)
	if err != nil {
		return nil, fmt.Errorf("error inserting in mongo: %w", err)
	}

	return toInsert, nil
}

// Pop fetches & deletes an oauth2 flow by state
func (r *PendingOAuth2Repository) Pop(ctx context.Context, state string) (models.PendingOAuth2, error) {
	res := r.collection.FindOneAndDelete(ctx, bson.D{{Key: "state", Value: state}})
	if err := res.Err(); err != nil {
		return nil, fmt.Errorf("error querying mongo: %w", err)
	}

	var po PendingOAuth2
	if err := res.Decode(&po); err != nil {
		return nil, fmt.Errorf("error decoding fetched value: %w", err)
	}

	return &po, nil
}

var _ repository.PendingOAuth2Repository = (*PendingOAuth2Repository)(nil)
