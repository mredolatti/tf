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
	IDField           string `bson:"_id"`
	NameField         string `bson:"name"`
	EmailField        string `bson:"email"`
	AccessTokenField  string `bson:"access_token"`
	RefreshTokenField string `bson:"refresh_token"`
}

// ID returns the id of the user
func (f *User) ID() string {
	return f.IDField
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
func (f *User) AccessToken() string {
	return f.AccessTokenField
}

// RefreshToken returns the refresh token for the user
func (f *User) RefreshToken() string {
	return f.RefreshTokenField
}

type UserRepository struct {
	collection *mongo.Collection
}

// Add implements repository.UserRepository
func (r *UserRepository) Add(
	ctx context.Context,
	id string,
	name string,
	email string,
	accessToken string,
	refreshToken string,
) (models.User, error) {
	u := User{
		NameField:         name,
		EmailField:        email,
		AccessTokenField:  accessToken,
		RefreshTokenField: refreshToken,
	}

	res, err := r.collection.InsertOne(ctx, &u)
	if err != nil {
		return nil, fmt.Errorf("error inserting user in mongodb: %w", err)
	}

	u.IDField = res.InsertedID.(primitive.ObjectID).String()
	return &u, nil
}

// Get implements repository.UserRepository
func (r *UserRepository) Get(ctx context.Context, id string) (models.User, error) {
	res := r.collection.FindOne(ctx, bson.D{{Key: "_id", Value: id}})
	if err := res.Err(); err != nil {
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
	res, err := r.collection.DeleteOne(ctx, bson.D{{Key: "_id", Value: userID}})
	if err != nil {
		return fmt.Errorf("error fetching user from mongodb: %w", err)
	}

        if res.DeletedCount != 1 {
            return errors.New("no items deleted") // TODO(mredolatti): mover a un error generico
        }
        
        return nil
}

// UpdateTokens implements repository.UserRepository
func (r *UserRepository) UpdateTokens(ctx context.Context, id string, accessToken string, refreshToken string) (models.User, error) {
    res := r.collection.FindOneAndUpdate(
        ctx,
        bson.D{{Key: "_id", Value: id}}, 
        bson.D{{Key: "accessToken", Value: accessToken}, {Key: "refreshToken", Value: refreshToken}},
    )

    if err := res.Err(); err != nil {
        return nil, fmt.Errorf("error updating tokens in mongodb: %w", err)
    }

    var u User
    if err := res.Decode(&u); err != nil {
        return nil, fmt.Errorf("error de-serializing updated user: %w", err)
    }

    // findOneAndUpdate returns the document without the updated fields.
    u.AccessTokenField = accessToken
    u.RefreshTokenField = refreshToken
    return &u, nil
}

func NewUserRepository(db *mongo.Database) *UserRepository {
	return &UserRepository{
		collection: db.Collection("Users"),
	}
}

var _ repository.UserRepository = (*UserRepository)(nil)
var _ models.User = (*User)(nil)
