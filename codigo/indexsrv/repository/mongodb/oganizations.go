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

// Organization is a MongoDB-compatible struct implementing models.Organization interface
type Organization struct {
	IDField   primitive.ObjectID `bson:"_id"`
	NameField string             `bson:"name"`
}

// ID returns the id of the organization
func (f *Organization) ID() string {
	return f.IDField.Hex()
}

// Name returns the name of the organization
func (f *Organization) Name() string {
	return f.NameField
}

type OrganizationRepository struct {
	collection *mongo.Collection
}

// Add implements repository.OrganizationRepository
func (r *OrganizationRepository) Add(ctx context.Context, source models.Organization) (models.Organization, error) {
	u := Organization{
		IDField:   primitive.NewObjectID(),
		NameField: source.Name(),
	}

	res, err := r.collection.InsertOne(ctx, &u)
	if err != nil {
		return nil, fmt.Errorf("error inserting organization in mongodb: %w", err)
	}

	u.IDField = res.InsertedID.(primitive.ObjectID)
	return &u, nil
}

// Get implements repository.OrganizationRepository
func (r *OrganizationRepository) List(ctx context.Context) ([]models.Organization, error) {
	cursor, err := r.collection.Find(ctx, bson.D{})
	if err != nil {
		return nil, fmt.Errorf("error fetching organization from mongodb: %w", err)
	}

	var orgs []models.Organization
	for cursor.Next(ctx) {
		o := &Organization{}
		if err := cursor.Decode(o); err != nil {
			return nil, fmt.Errorf("error deserializing organization from mongo result: %w", err)
		}
		orgs = append(orgs, o)
	}

	return orgs, nil
}

// Get implements repository.OrganizationRepository
func (r *OrganizationRepository) Get(ctx context.Context, id string) (models.Organization, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("error constructing objectID for organization with id=%s: %w", id, err)
	}
	res := r.collection.FindOne(ctx, bson.D{{Key: "_id", Value: oid}})
	if err := res.Err(); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, repository.ErrNotFound
		}
		return nil, fmt.Errorf("error fetching organization from mongodb: %w", err)
	}

	var u Organization
	if err := res.Decode(&u); err != nil {
		return nil, fmt.Errorf("error deserializing organization from mongo result: %w", err)
	}

	return &u, nil
}

// Remove implements repository.OrganizationRepository
func (r *OrganizationRepository) Remove(ctx context.Context, organizationID string) error {
	oid, err := primitive.ObjectIDFromHex(organizationID)
	if err != nil {
		return fmt.Errorf("error constructing objectID for organization with id=%s: %w", organizationID, err)
	}
	res, err := r.collection.DeleteOne(ctx, bson.D{{Key: "_id", Value: oid}})
	if err != nil {
		return fmt.Errorf("error fetching organization from mongodb: %w", err)
	}

	if res.DeletedCount != 1 {
		return errors.New("no items deleted") // TODO(mredolatti): mover a un error generico
	}

	return nil
}

func NewOrganizationRepository(db *mongo.Database) *OrganizationRepository {
	return &OrganizationRepository{
		collection: db.Collection("Organizations"),
	}
}

var _ repository.OrganizationRepository = (*OrganizationRepository)(nil)
var _ models.Organization = (*Organization)(nil)
