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

// UserAccount is a postgres-compatible struct implementing models.UserAccount interface
type UserAccount struct {
	IDField           primitive.ObjectID `bson:"_id"`
	UserIDField       primitive.ObjectID `bson:"userId"`
	ServerIDField     primitive.ObjectID `bson:"serverId"`
	AccessTokenField  string             `bson:"accessToken"`
	RefreshTokenField string             `bson:"refreshToken"`
	CheckpointField   int64              `bson:"checkpoint"`
}

// UserID returns the if of the user who has an account in a file server
func (u *UserAccount) UserID() string {
	return u.UserIDField.Hex()
}

// FileServerID returns the id of the server in which the user has the account
func (u *UserAccount) FileServerID() string {
	return u.ServerIDField.Hex()
}

// Token returns the token used to make request on behalf of this user to the server
func (u *UserAccount) Token() string {
	return u.AccessTokenField
}

// RefreshToken returns the token used to get new tokens when the current one expires
func (u *UserAccount) RefreshToken() string {
	return u.RefreshTokenField
}

// Checkpoint returns a nanosecond-granularity timestamp of the last update (or zero y if has never happend)
func (u *UserAccount) Checkpoint() int64 {
	return u.CheckpointField
}

// UserAccountRepository is a mapping to a table in postgres that allows enables operations on user accounts
type UserAccountRepository struct {
	collection *mongo.Collection
}

func NewUserAccountRepository(db *mongo.Database) *UserAccountRepository {
	return &UserAccountRepository{
		collection: db.Collection("UserAccounts"),
	}
}

// Add implements repository.UserAccountRepository
func (r *UserAccountRepository) Add(ctx context.Context, userID string, serverID string, accessToken string, refreshToken string) (models.UserAccount, error) {
	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("error constructing objectID for user with id=%s: %w", userID, err)
	}
	fsid, err := primitive.ObjectIDFromHex(serverID)
	if err != nil {
		return nil, fmt.Errorf("error constructing objectID for server with id=%s: %w", serverID, err)
	}

	a := UserAccount{
		IDField:           primitive.NewObjectID(),
		UserIDField:       uid,
		ServerIDField:     fsid,
		AccessTokenField:  accessToken,
		RefreshTokenField: refreshToken,
	}

	res, err := r.collection.InsertOne(ctx, &a)
	if err != nil {
		return nil, fmt.Errorf("error inserting mapping in mongodb: %w", err)
	}

	a.IDField = res.InsertedID.(primitive.ObjectID)
	return &a, nil
}

// Get implements repository.UserAccountRepository
func (r *UserAccountRepository) Get(ctx context.Context, userID string, serverID string) (models.UserAccount, error) {
	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("error constructing objectID for user with id=%s: %w", userID, err)
	}
	fsid, err := primitive.ObjectIDFromHex(serverID)
	if err != nil {
		return nil, fmt.Errorf("error constructing objectID for server wth id=%s: %w", serverID, err)
	}

	res := r.collection.FindOne(ctx, bson.D{{Key: "userId", Value: uid}, {Key: "serverId", Value: fsid}})
	if err := res.Err(); err != nil {
		return nil, fmt.Errorf("error querying mongo: %w", err)
	}

	var account UserAccount
	if err := res.Decode(&account); err != nil {
		return nil, fmt.Errorf("error parsing fetched account: %w", err)
	}

	return &account, nil
}

// List implements repository.UserAccountRepository
func (r *UserAccountRepository) List(ctx context.Context, userID string) ([]models.UserAccount, error) {
	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("error constructing objectID for user with id=%s: %w", userID, err)
	}

	cursor, err := r.collection.Find(ctx, bson.D{{Key: "userId", Value: uid}})
	if err != nil {
		return nil, fmt.Errorf("error querying mongo: %w", err)
	}

	var accounts []models.UserAccount
	for cursor.Next(ctx) {
		account := &UserAccount{}
		if err := cursor.Decode(account); err != nil {
			return nil, fmt.Errorf("error decoding item: %w", err)
		}
		accounts = append(accounts, account)

	}

	return accounts, nil
}

// Remove implements repository.UserAccountRepository
func (r *UserAccountRepository) Remove(ctx context.Context, userID string, serverID string) error {
	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return fmt.Errorf("error constructing objectID for user with id=%s: %w", userID, err)
	}
	fsid, err := primitive.ObjectIDFromHex(serverID)
	if err != nil {
		return fmt.Errorf("error constructing objectID for server wth id=%s: %w", serverID, err)
	}

	res, err := r.collection.DeleteOne(ctx, bson.D{{Key: "userId", Value: uid}, {Key: "serverId", Value: fsid}})
	if err != nil {
		return fmt.Errorf("error executing mongo operation: %w", err)
	}

	if res.DeletedCount == 0 {
		return repository.ErrNotFound
	}

	return nil
}

// UpdateCheckpoint implements repository.UserAccountRepository
func (r *UserAccountRepository) UpdateCheckpoint(ctx context.Context, userID string, serverID string, newCheckpoint int64) error {
	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return fmt.Errorf("error constructing objectID for user with id=%s: %w", userID, err)
	}
	fsid, err := primitive.ObjectIDFromHex(serverID)
	if err != nil {
		return fmt.Errorf("error constructing objectID for server wth id=%s: %w", serverID, err)
	}

	res, err := r.collection.UpdateOne(
		ctx,
		bson.D{{Key: "userId", Value: uid}, {Key: "serverId", Value: fsid}},
		bson.D{{Key: "$set", Value: bson.D{{Key: "checkpoint", Value: newCheckpoint}}}},
	)
	if err != nil {
		return fmt.Errorf("error querying mongo: %w", err)
	}

	if res.ModifiedCount == 0 {
		return repository.ErrNotFound
	}

	return nil
}

// UpdateTokens implements repository.UserAccountRepository
func (r *UserAccountRepository) UpdateTokens(ctx context.Context, userID string, serverID string, accessToken string, refreshToken string) error {
	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return fmt.Errorf("error constructing objectID for user with id=%s: %w", userID, err)
	}
	fsid, err := primitive.ObjectIDFromHex(serverID)
	if err != nil {
		return fmt.Errorf("error constructing objectID for server wth id=%s: %w", serverID, err)
	}

	res, err := r.collection.UpdateOne(
		ctx,
		bson.D{{Key: "userId", Value: uid}, {Key: "serverId", Value: fsid}},
		bson.D{{Key: "$set", Value: bson.D{
			{Key: "accessToken", Value: accessToken},
			{Key: "refreshToken", Value: refreshToken},
		}}},
	)
	if err != nil {
		return fmt.Errorf("error querying mongo: %w", err)
	}

	if res.ModifiedCount == 0 {
		return repository.ErrNotFound
	}

	return nil
}

var _ repository.UserAccountRepository = (*UserAccountRepository)(nil)
var _ models.UserAccount = (*UserAccount)(nil)
