package mongodb

import (
	"context"
	"errors"
	"fmt"

	"github.com/mredolatti/tf/codigo/indexsrv/models"
	"github.com/mredolatti/tf/codigo/indexsrv/repository"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// User is a MongoDB-compatible struct implementing models.User interface
type User struct {
	IDField           primitive.ObjectID `bson:"_id"`
	NameField         string             `bson:"name"`
	EmailField        string             `bson:"email"`
	PasswordHashField string             `bson:"password"`
	TFASecretField    string             `bson:"tfaSecret"`
}

// TFASecret implements models.User
func (f *User) TFASecret() string {
	return f.TFASecretField
}

// ID returns the id of the user
func (f *User) ID() string {
	return f.IDField.Hex()
}

// Name returns the name of the user
func (f *User) Name() string {
	return f.NameField
}

// Email returns the email of the user
func (f *User) Email() string {
	return f.EmailField
}

// AccessToken returns the last access token of the user
func (f *User) PasswordHash() string {
	return f.PasswordHashField
}

type UserRepository struct {
	collection *mongo.Collection
}

// Add implements repository.UserRepository
func (r *UserRepository) Add(ctx context.Context, name string, email string, passwordHash string) (models.User, error) {
	u := User{
		IDField:           primitive.NewObjectID(),
		NameField:         name,
		EmailField:        email,
		PasswordHashField: passwordHash,
	}

	res, err := r.collection.InsertOne(ctx, &u)
	if err != nil {
		return nil, fmt.Errorf("error inserting user in mongodb: %w", err)
	}

	u.IDField = res.InsertedID.(primitive.ObjectID)
	return &u, nil
}

// Get implements repository.UserRepository
func (r *UserRepository) Get(ctx context.Context, id string) (models.User, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("error constructing objectID for user with id=%s: %w", id, err)
	}
	res := r.collection.FindOne(ctx, bson.D{{Key: "_id", Value: oid}})
	if err := res.Err(); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, repository.ErrNotFound
		}
		return nil, fmt.Errorf("error fetching user from mongodb: %w", err)
	}

	var u User
	if err := res.Decode(&u); err != nil {
		return nil, fmt.Errorf("error deserializing user from mongo result: %w", err)
	}

	return &u, nil
}

// GetByEmail
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (models.User, error) {
	res := r.collection.FindOne(ctx, bson.D{{Key: "email", Value: email}})
	if err := res.Err(); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, repository.ErrNotFound
		}
		return nil, fmt.Errorf("error fetching user from mongodb: %w", err)
	}

	var u User
	if err := res.Decode(&u); err != nil {
		return nil, fmt.Errorf("error deserializing user from mongo result: %w", err)
	}

	return &u, nil
}

// Remove implements repository.UserRepository
func (r *UserRepository) Remove(ctx context.Context, userID string) error {
	oid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return fmt.Errorf("error constructing objectID for user with id=%s: %w", userID, err)
	}
	res, err := r.collection.DeleteOne(ctx, bson.D{{Key: "_id", Value: oid}})
	if err != nil {
		return fmt.Errorf("error fetching user from mongodb: %w", err)
	}

	if res.DeletedCount != 1 {
		return errors.New("no items deleted") // TODO(mredolatti): mover a un error generico
	}

	return nil
}

// UpdateTokens implements repository.UserRepository
func (r *UserRepository) UpdatePassword(ctx context.Context, id string, passwordHash string) (models.User, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("error constructing objectID for user with id=%s: %w", id, err)
	}
	res := r.collection.FindOneAndUpdate(
		ctx,
		bson.D{{Key: "_id", Value: oid}},
		bson.D{{Key: "$set", Value: bson.D{{Key: "password", Value: passwordHash}}}},
	)

	if err := res.Err(); err != nil {
		return nil, fmt.Errorf("error updating tokens in mongodb: %w", err)
	}

	var u User
	if err := res.Decode(&u); err != nil {
		return nil, fmt.Errorf("error de-serializing updated user: %w", err)
	}

	// findOneAndUpdate returns the document without the updated fields.
	u.PasswordHashField = passwordHash
	return &u, nil
}

// Update2FA implements repository.UserRepository
func (r *UserRepository) Update2FA(ctx context.Context, id string, keySecret string) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("error constructing objectID for user with id=%s: %w", id, err)
	}

	res := r.collection.FindOneAndUpdate(
		ctx,
		bson.D{{Key: "_id", Value: oid}},
		bson.D{{Key: "$set", Value: bson.D{{Key: "tfaSecret", Value: keySecret}}}},
	)
	if err := res.Err(); err != nil {
		return fmt.Errorf("error updating 2fa key secret in mongodb: %w", err)
	}
	return nil
}

func NewUserRepository(db *mongo.Database) *UserRepository {
	return &UserRepository{
		collection: db.Collection("Users"),
	}
}

var _ repository.UserRepository = (*UserRepository)(nil)
var _ models.User = (*User)(nil)
