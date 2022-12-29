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

// FileServer is a MongoDB-compatible struct implementing models.FileServer interface
type FileServer struct {
	IDField              primitive.ObjectID `bson:"_id"`
	NameField            string             `bson:"name"`
	OrgField             primitive.ObjectID `bson:"orgId"`
	AuthURLField         string             `bson:"authUrl"`
	TokenURLField        string             `bson:"tokenUrl"`
	FetchURLField        string             `bson:"fetchUrl"`
	ControlEndpointField string             `bson:"controlEndpoint"`
}

// ID returns the id of the fileServer
func (f *FileServer) ID() string {
	return f.IDField.Hex()
}

// Name returns the name of the fileServer
func (f *FileServer) Name() string {
	return f.NameField
}

// OrganizationID returns the ID of the organization this server belongs to
func (f *FileServer) OrganizationID() string {
	return f.OrgField.Hex()
}

// AuthURL returns the URL used to authorize users when linking a server to their account
func (f *FileServer) AuthURL() string {
	return f.AuthURLField
}

// TokenURL returns the URL used to get a token based on an auth code or a refresh token
func (f *FileServer) TokenURL() string {
	return f.TokenURLField
}

// FetchURL returns the URL to be used when returning fetch recipes
func (f *FileServer) FetchURL() string {
	return f.FetchURLField
}

// ControlEndpoint returns the control endpoiunt used to make RPC calls
func (f *FileServer) ControlEndpoint() string {
	return f.ControlEndpointField
}

type FileServerRepository struct {
	collection *mongo.Collection
}

// Add implements repository.FileServerRepository
func (r *FileServerRepository) Add(
	ctx context.Context,
	id string,
	name string,
	orgID string,
	authURL string,
	tokenURL string,
	fetchURL string,
	controlEndpoint string,
) (models.FileServer, error) {
	oid, err := primitive.ObjectIDFromHex(orgID)
	if err != nil {
		return nil, fmt.Errorf("error constructing objectID for fileServer with id=%s: %w", orgID, err)
	}
	u := FileServer{
		IDField:              primitive.NewObjectID(),
		NameField:            name,
		OrgField:             oid,
		AuthURLField:         authURL,
		TokenURLField:        tokenURL,
		FetchURLField:        fetchURL,
		ControlEndpointField: controlEndpoint,
	}

	res, err := r.collection.InsertOne(ctx, &u)
	if err != nil {
		return nil, fmt.Errorf("error inserting fileServer in mongodb: %w", err)
	}

	u.IDField = res.InsertedID.(primitive.ObjectID)
	return &u, nil
}

// Get implements repository.FileServerRepository
func (r *FileServerRepository) List(ctx context.Context, orgID string) ([]models.FileServer, error) {
	oid, err := primitive.ObjectIDFromHex(orgID)
	if err != nil {
		return nil, fmt.Errorf("error constructing objectID for fileServer with id=%s: %w", orgID, err)
	}
	cursor, err := r.collection.Find(ctx, bson.D{{Key: "orgId", Value: oid}})
	if err != nil {
		return nil, fmt.Errorf("error fetching fileServer from mongodb: %w", err)
	}

	var orgs []models.FileServer
	for cursor.Next(ctx) {
		o := &FileServer{}
		if err := cursor.Decode(o); err != nil {
			return nil, fmt.Errorf("error deserializing fileServer from mongo result: %w", err)
		}
		orgs = append(orgs, o)
	}

	return orgs, nil
}

// Get implements repository.FileServerRepository
func (r *FileServerRepository) Get(ctx context.Context, id string) (models.FileServer, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("error constructing objectID for fileServer with id=%s: %w", id, err)
	}
	res := r.collection.FindOne(ctx, bson.D{{Key: "_id", Value: oid}})
	if err := res.Err(); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, repository.ErrNotFound
		}
		return nil, fmt.Errorf("error fetching fileServer from mongodb: %w", err)
	}

	var u FileServer
	if err := res.Decode(&u); err != nil {
		return nil, fmt.Errorf("error deserializing fileServer from mongo result: %w", err)
	}

	return &u, nil
}

// Remove implements repository.FileServerRepository
func (r *FileServerRepository) Remove(ctx context.Context, fileServerID string) error {
	oid, err := primitive.ObjectIDFromHex(fileServerID)
	if err != nil {
		return fmt.Errorf("error constructing objectID for fileServer with id=%s: %w", fileServerID, err)
	}
	res, err := r.collection.DeleteOne(ctx, bson.D{{Key: "_id", Value: oid}})
	if err != nil {
		return fmt.Errorf("error fetching fileServer from mongodb: %w", err)
	}

	if res.DeletedCount != 1 {
		return errors.New("no items deleted") // TODO(mredolatti): mover a un error generico
	}

	return nil
}

func NewFileServerRepository(db *mongo.Database) *FileServerRepository {
	return &FileServerRepository{
		collection: db.Collection("FileServers"),
	}
}

var _ repository.FileServerRepository = (*FileServerRepository)(nil)
var _ models.FileServer = (*FileServer)(nil)
