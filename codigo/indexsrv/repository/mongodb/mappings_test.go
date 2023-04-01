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
	inserted1, err := repo.Add(ctx, uid1.Hex(), &Mapping{
		OrganizationNameField: "org1",
		FileServerNameField:   "fs1",
		SizeBytesField:        0,
		PathField:             "/path/to/f1",
		RefField:              "ref1",
		DeletedField:          false,
		UpdatedField:          currentTime.UnixNano(),
	})
	assert.Nil(t, err)
	assert.Equal(t, inserted1.UserID(), uid1.Hex())
	assert.Equal(t, inserted1.Path(), "/path/to/f1")
	assert.Equal(t, inserted1.Deleted(), false)
	assert.Equal(t, inserted1.Ref(), "ref1")
	assert.Equal(t, inserted1.SizeBytes(), int64(0))
	assert.Equal(t, inserted1.Updated(), currentTime)
	assert.Equal(t, inserted1.OrganizationName(), "org1")
	assert.Equal(t, inserted1.ServerName(), "fs1")

	inserted2, err := repo.Add(ctx, uid1.Hex(), &Mapping{
		OrganizationNameField: "org1",
		FileServerNameField:   "fs1",
		PathField:             "/path/to/f2",
		RefField:              "ref2",
		DeletedField:          false,
		UpdatedField:          currentTime.UnixNano(),
	})
	assert.Nil(t, err)
	assert.Equal(t, inserted2.UserID(), uid1.Hex())
	assert.Equal(t, inserted2.Path(), "/path/to/f2")
	assert.Equal(t, inserted2.Deleted(), false)
	assert.Equal(t, inserted2.Ref(), "ref2")
	assert.Equal(t, inserted2.Updated(), currentTime)
	assert.Equal(t, inserted2.OrganizationName(), "org1")
	assert.Equal(t, inserted2.ServerName(), "fs1")

	fetched, err := repo.List(ctx, uid1.Hex(), models.MappingQuery{})
	assert.Nil(t, err)
	assert.Contains(t, fetched, inserted1)
	assert.Contains(t, fetched, inserted2)

	newTime := currentTime.Add(time.Hour).UnixNano()
	err = repo.HandleServerUpdates(ctx, uid1.Hex(), "org1", "fs1", []models.Update{
		{
			FileRef:          "ref1",
			Checkpoint:       newTime,
			SizeBytes:        100,
			ChangeType:       models.UpdateTypeFileUpdate,
		},
		{
			FileRef:          "ref2",
			Checkpoint:       newTime,
			ChangeType:       models.UpdateTypeFileDelete,
		},
		{
			FileRef:          "ref3",
			Checkpoint:       newTime,
			ChangeType:       models.UpdateTypeFileAdd,
			SizeBytes:        200,
		},
	})
	assert.Nil(t, err)
	inserted1.(*Mapping).UpdatedField = newTime
	inserted1.(*Mapping).SizeBytesField = 100
	inserted2.(*Mapping).UpdatedField = newTime
	inserted2.(*Mapping).DeletedField = true

	fetched, err = repo.List(ctx, uid1.Hex(), models.MappingQuery{})
	assert.Nil(t, err)
	assert.Equal(t, 3, len(fetched))
	assert.Contains(t, fetched, inserted1)

	// manually assert 3rd item since we don't have a reference
	assert.Equal(t, uid1.Hex(), fetched[2].UserID())
	assert.Equal(t, "org1", fetched[2].OrganizationName())
	assert.Equal(t, "fs1", fetched[2].ServerName())
	assert.Equal(t, false, fetched[2].Deleted())
	assert.Equal(t, "unassigned/org1/fs1/ref3", fetched[2].Path())
	assert.Equal(t, "ref3", fetched[2].Ref())
	assert.Equal(t, newTime, fetched[2].Updated().UnixNano())
	assert.Equal(t, int64(200), fetched[2].SizeBytes())
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

	b.ResetTimer()

	for idx := 0; idx < b.N; idx++ {
		uid1 := primitive.NewObjectID()
		repo.Add(ctx, uid1.Hex(), &Mapping{
			OrganizationNameField: "org1",
			FileServerNameField:   "fs1",
			PathField:             fmt.Sprintf("path/to/f%d", idx),
			RefField:              fmt.Sprintf("ref%d", idx),
			DeletedField:          false,
			UpdatedField:          currentTime.UnixNano(),
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

	b.ResetTimer()
	// SetParallelism always multiplies `p` by number of CPUs. divide it again to get accurate thread count
	b.SetParallelism(concurrency / runtime.GOMAXPROCS(0))

	b.RunParallel(func(p *testing.PB) {
		for idx := 0; p.Next(); idx++ {
			uid1 := primitive.NewObjectID()
			_, err = repo.Add(ctx, uid1.Hex(), &Mapping{
				OrganizationNameField: "org1",
				FileServerNameField:   "fs1",
				PathField:             fmt.Sprintf("path/to/f%d", idx),
				RefField:              fmt.Sprintf("ref%d", idx),
				DeletedField:          false,
				UpdatedField:          currentTime.UnixNano(),
			})
			if err != nil {
				panic(err.Error())
			}
		}
	})
}
