package mongodb

import (
	"context"
	"fmt"
	"runtime"
	"testing"
	"time"

	"github.com/mredolatti/tf/codigo/indexsrv/models"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestMappingsIntegration(t *testing.T) {

	ctx := context.Background()
	conn, err := mongo.Connect(ctx)
	assert.Nil(t, err)

	db := conn.Database("mifs_test_mappings")
	db.Drop(ctx)

	repo := NewMappingRepository(db)
	currentTime := time.Now().UTC()

	uid1 := primitive.NewObjectID()
	fsid1 := primitive.NewObjectID()
	inserted1, err := repo.Add(ctx, uid1.Hex(), &Mapping{
		ServerIDField: fsid1,
		PathField:     "/path/to/f1",
		RefField:      "ref1",
		DeletedField:  false,
		UpdatedField:  currentTime.UnixNano(),
	})
	assert.Nil(t, err)
	assert.Equal(t, inserted1.UserID(), uid1.Hex())
	assert.Equal(t, inserted1.Path(), "/path/to/f1")
	assert.Equal(t, inserted1.Deleted(), false)
	assert.Equal(t, inserted1.Ref(), "ref1")
	assert.Equal(t, inserted1.Updated(), currentTime)
	assert.Equal(t, inserted1.FileServerID(), fsid1.Hex())

	inserted2, err := repo.Add(ctx, uid1.Hex(), &Mapping{
		ServerIDField: fsid1,
		PathField:     "/path/to/f2",
		RefField:      "ref2",
		DeletedField:  false,
		UpdatedField:  currentTime.UnixNano(),
	})
	assert.Nil(t, err)
	assert.Equal(t, inserted2.UserID(), uid1.Hex())
	assert.Equal(t, inserted2.Path(), "/path/to/f2")
	assert.Equal(t, inserted2.Deleted(), false)
	assert.Equal(t, inserted2.Ref(), "ref2")
	assert.Equal(t, inserted2.Updated(), currentTime)
	assert.Equal(t, inserted2.FileServerID(), fsid1.Hex())

	fetched, err := repo.List(ctx, uid1.Hex(), models.MappingQuery{})
	assert.Nil(t, err)
	assert.Contains(t, fetched, inserted1, inserted2)

	newTime := currentTime.Add(time.Hour).UnixNano()
	err = repo.HandleServerUpdates(ctx, uid1.Hex(), []models.Update{
		{
			ServerID:       fsid1.Hex(),
			OrganizationID: "someOrg",
			FileRef:        "ref1",
			Checkpoint:     newTime,
			ChangeType:     models.UpdateTypeFileUpdate,
		},
		{
			ServerID:       fsid1.Hex(),
			OrganizationID: "someOrg",
			FileRef:        "ref2",
			Checkpoint:     newTime,
			ChangeType:     models.UpdateTypeFileDelete,
		},
		{
			ServerID:       fsid1.Hex(),
			OrganizationID: "someOrg",
			FileRef:        "ref3",
			Checkpoint:     newTime,
			ChangeType:     models.UpdateTypeFileAdd,
		},
	})
	assert.Nil(t, err)
	inserted1.(*Mapping).UpdatedField = newTime
	inserted2.(*Mapping).UpdatedField = newTime
	inserted2.(*Mapping).DeletedField = true

	fetched, err = repo.List(ctx, uid1.Hex(), models.MappingQuery{})
	assert.Nil(t, err)
	assert.Equal(t, len(fetched), 3)
	assert.Contains(t, fetched, inserted1, inserted2)

	// manually assert 3rd item since we don't have a reference
	assert.Equal(t, fetched[2].UserID(), uid1.Hex())
	assert.Equal(t, fetched[2].FileServerID(), fsid1.Hex())
	assert.Equal(t, fetched[2].Deleted(), false)
	assert.Equal(t, fetched[2].Path(), fmt.Sprintf("unassigned/%s/ref3", fsid1.Hex()))
	assert.Equal(t, fetched[2].Ref(), "ref3")
	assert.Equal(t, fetched[2].Updated().UnixNano(), newTime)
}

func BenchmarkMongoMappingNoConcurrency(b *testing.B) {
	ctx := context.Background()
	conn, err := mongo.Connect(ctx)
	if err != nil {
		panic(err.Error())
	}

	db := conn.Database("mifs_bench_mappings")
	defer db.Drop(ctx)

	repo := NewMappingRepository(db)
	currentTime := time.Now().UTC()

	fsid1 := primitive.NewObjectID()

	b.ResetTimer()

	for idx := 0; idx < b.N; idx++ {
		uid1 := primitive.NewObjectID()
		repo.Add(ctx, uid1.Hex(), &Mapping{
			ServerIDField: fsid1,
			PathField:     fmt.Sprintf("path/to/f%d", idx),
			RefField:      fmt.Sprintf("ref%d", idx),
			DeletedField:  false,
			UpdatedField:  currentTime.UnixNano(),
		})
	}
}

func BenchmarkMongoMappingWriteConcurrent100(b *testing.B) {
	benchmarkMongoMappingConcurrency(b, 100)
}

func BenchmarkMongoMappingWriteConcurrent250(b *testing.B) {
	benchmarkMongoMappingConcurrency(b, 250)
}

func BenchmarkMongoMappingWriteConcurrent500(b *testing.B) {
	benchmarkMongoMappingConcurrency(b, 500)
}

func benchmarkMongoMappingConcurrency(b *testing.B, concurrency int) {
	b.Helper()
	ctx := context.Background()
	var max uint64 = 300
	conn, err := mongo.Connect(ctx, &options.ClientOptions{
		MaxPoolSize: &max,
		MinPoolSize: &max,
	})
	if err != nil {
		panic(err.Error())
	}

	db := conn.Database("mifs_test_mappings_concurrent")
	defer db.Drop(ctx)

	repo := NewMappingRepository(db)
	currentTime := time.Now().UTC()

	fsid1 := primitive.NewObjectID()

	b.ResetTimer()
	// SetParallelism always multiplies `p` by number of CPUs. divide it again to get accurate thread count
	b.SetParallelism(concurrency / runtime.GOMAXPROCS(0))

	b.RunParallel(func(p *testing.PB) {
		for idx := 0; p.Next(); idx++ {
			uid1 := primitive.NewObjectID()
			_, err = repo.Add(ctx, uid1.Hex(), &Mapping{
				ServerIDField: fsid1,
				PathField:     fmt.Sprintf("path/to/f%d", idx),
				RefField:      fmt.Sprintf("ref%d", idx),
				DeletedField:  false,
				UpdatedField:  currentTime.UnixNano(),
			})
			if err != nil {
				panic(err.Error())
			}
		}
	})
}
