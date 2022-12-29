package mongodb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/mredolatti/tf/codigo/indexsrv/models"
	"github.com/mredolatti/tf/codigo/indexsrv/repository"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Mapping is a MongoDB-compatible struct implementing models.Mapping interface
type Mapping struct {
	IDField       primitive.ObjectID `bson:"_id"`
	UserIDField   primitive.ObjectID `bson:"userId"`
	ServerIDField primitive.ObjectID `bson:"serverId"`
	PathField     string             `bson:"path"`
	RefField      string             `bson:"ref"`
	DeletedField  bool               `bson:"deleted"`
	UpdatedField  int64              `bson:"updated"`
}

// ID returns the id of the mapping
func (m *Mapping) ID() string {
	return m.IDField.Hex()
}

// UserID implements models.Mapping
func (m *Mapping) UserID() string {
	return m.UserIDField.Hex()
}

// FileServerID implements models.Mapping
func (m *Mapping) FileServerID() string {
	return m.ServerIDField.Hex()
}

// Ref implements models.Mapping
func (m *Mapping) Ref() string {
	return m.RefField
}

// Path implements models.Mapping
func (m *Mapping) Path() string {
	return m.PathField
}

// Deleted implements models.Mapping
func (m *Mapping) Deleted() bool {
	return m.DeletedField
}

// Updated implements models.Mapping
func (m *Mapping) Updated() time.Time {
	return time.Unix(0, m.UpdatedField).UTC()
}

func (m *Mapping) String() string {
	return fmt.Sprintf("{id=%s userId=%s serverId=%s path=%s ref=%s deleted=%t updated=%d}",
		m.ID(), m.UserID(), m.FileServerID(), m.Path(), m.Ref(), m.Deleted(), m.Updated().UnixNano(),
	)
}

type MappingRepository struct {
	collection *mongo.Collection
}

// HandleServerUpdates implements repository.MappingRepository
func (r *MappingRepository) HandleServerUpdates(ctx context.Context, userID string, updates []models.Update) error {
	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return fmt.Errorf("error constructing objectID for mapping with id=%s: %w", userID, err)
	}
	ops := make([]mongo.WriteModel, 0, len(updates))
	for _, update := range updates {

		fsid, err := primitive.ObjectIDFromHex(update.ServerID)
		if err != nil {
			return fmt.Errorf("error constructing objectID for mapping with id=%s: %w", update.ServerID, err)
		}

		switch update.ChangeType {
		case models.UpdateTypeFileAdd:
			ops = append(ops, mongo.NewInsertOneModel().SetDocument(Mapping{
				IDField:       primitive.NewObjectID(),
				UserIDField:   uid,
				ServerIDField: fsid,
				PathField:     fmt.Sprintf("unassigned/%s/%s", update.ServerID, update.FileRef),
				RefField:      update.FileRef,
				DeletedField:  false,
				UpdatedField:  update.Checkpoint,
			}))
		case models.UpdateTypeFileUpdate, models.UpdateTypeFileDelete:
			ops = append(ops, mongo.NewUpdateOneModel().
				SetFilter(bson.D{{Key: "userId", Value: uid}, {Key: "serverId", Value: fsid}, {Key: "ref", Value: update.FileRef}}).
				SetUpdate(bson.D{{Key: "$set", Value: bson.D{
					{Key: "updated", Value: update.Checkpoint},
					{Key: "deleted", Value: update.ChangeType == models.UpdateTypeFileDelete},
				}}}),
			)
		}
	}

	_, err = r.collection.BulkWrite(ctx, ops)
	if err != nil {
		return fmt.Errorf("error performing bulk update in mongodb: %w", err)
	}
	return nil
}

// Update implements repository.MappingRepository
func (*MappingRepository) Update(ctx context.Context, userID string, mappingID string, mapping models.Mapping) (models.Mapping, error) {
	panic("unimplemented")
}

// Add implements repository.MappingRepository
func (r *MappingRepository) Add(ctx context.Context, userID string, source models.Mapping) (models.Mapping, error) {
	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("error constructing objectID for mapping with user id=%s: %w", userID, err)
	}
	fsid, err := primitive.ObjectIDFromHex(source.FileServerID())
	if err != nil {
		return nil, fmt.Errorf("error constructing objectID for mapping with server id=%s: %w", source.FileServerID(), err)
	}
	u := Mapping{
		IDField:       primitive.NewObjectID(),
		UserIDField:   uid,
		ServerIDField: fsid,
		PathField:     source.Path(),
		RefField:      source.Ref(),
		DeletedField:  source.Deleted(),
		UpdatedField:  source.Updated().UnixNano(),
	}

	res, err := r.collection.InsertOne(ctx, &u)
	if err != nil {
		return nil, fmt.Errorf("error inserting mapping in mongodb: %w", err)
	}

	u.IDField = res.InsertedID.(primitive.ObjectID)
	return &u, nil
}

// Get implements repository.MappingRepository
func (r *MappingRepository) List(ctx context.Context, userID string, query models.MappingQuery) ([]models.Mapping, error) {
	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("error constructing objectID for mapping with id=%s: %w", userID, err)
	}
	cursor, err := r.collection.Find(ctx, bson.D{{Key: "userId", Value: uid}})
	if err != nil {
		return nil, fmt.Errorf("error fetching mapping from mongodb: %w", err)
	}

	var orgs []models.Mapping
	for cursor.Next(ctx) {
		o := &Mapping{}
		if err := cursor.Decode(o); err != nil {
			return nil, fmt.Errorf("error deserializing mapping from mongo result: %w", err)
		}
		orgs = append(orgs, o)
	}

	return orgs, nil
}

// Remove implements repository.MappingRepository
func (r *MappingRepository) Remove(ctx context.Context, userID string, fileServerID string) error {
	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return fmt.Errorf("error constructing objectID for mapping with id=%s: %w", userID, err)
	}
	fsid, err := primitive.ObjectIDFromHex(fileServerID)
	if err != nil {
		return fmt.Errorf("error constructing objectID for mapping with id=%s: %w", fileServerID, err)
	}
	res, err := r.collection.DeleteOne(ctx, bson.D{{Key: "userId", Value: uid}, {Key: "fileServerId", Value: fsid}})
	if err != nil {
		return fmt.Errorf("error fetching mapping from mongodb: %w", err)
	}

	if res.DeletedCount != 1 {
		return errors.New("no items deleted") // TODO(mredolatti): mover a un error generico
	}

	return nil
}

func NewMappingRepository(db *mongo.Database) *MappingRepository {
	return &MappingRepository{
		collection: db.Collection("Mappings"),
	}
}

var _ repository.MappingRepository = (*MappingRepository)(nil)
var _ models.Mapping = (*Mapping)(nil)
