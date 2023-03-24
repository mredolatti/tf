package psql

import (
	"context"
	"fmt"
	"math/rand"
	"runtime"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/mredolatti/tf/codigo/common/refutil"
	"github.com/mredolatti/tf/codigo/indexsrv/models"
	"github.com/mredolatti/tf/codigo/indexsrv/repository"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
)

func TestOrgRepoIntegration(t *testing.T) {
	db, err := sqlx.Connect("pgx", "postgres://postgres:mysecretpassword@localhost:5432/indexsrv")
	assert.Nil(t, err)

	repo := NewOrganizationRepository(db)

	// Cleanup
	defer db.Query("DELETE FROM organizations WHERE name = 'test_org_1'")

	added, err := repo.Add(context.Background(), "test_org_1")
	assert.Nil(t, err)
	assert.Equal(t, "test_org_1", added.Name())

	fetched, err := repo.Get(context.Background(), added.ID())
	assert.Nil(t, err)
	assert.Equal(t, "test_org_1", fetched.Name())

	list, err := repo.List(context.Background())
	assert.Nil(t, err)

	found := false
	for _, org := range list {
		if org.Name() == "test_org_1" {
			found = true
		}
	}
	assert.True(t, found)

	err = repo.Remove(context.Background(), fetched.ID())
	assert.Nil(t, err)

	fetched, err = repo.Get(context.Background(), added.ID())
	assert.Nil(t, fetched)
	assert.Equal(t, repository.ErrNotFound, err)
}

func TestFileServerRepoIntegration(t *testing.T) {
	db, err := sqlx.Connect("pgx", "postgres://postgres:mysecretpassword@localhost:5432/indexsrv")
	assert.Nil(t, err)

	orgRepo := NewOrganizationRepository(db)

	// Cleanup
	defer func() {
		db.Query("DELETE FROM organizations WHERE name = 'test_org_1'")
	}()

	newOrg, err := orgRepo.Add(context.Background(), "test_org_1")
	assert.Nil(t, err)
	assert.Equal(t, "test_org_1", newOrg.Name())

	// Begin testing the file server repo

	fsRepo := NewFileServerRepository(db)
	rand.Seed(time.Now().UnixNano())
	newServer, err := fsRepo.Add(context.Background(), "server1", newOrg.ID(), "https://auth.server1", "https://token.server1", "sftp://fetch.server1", "control.server1:1234")
	assert.Nil(t, err)
	assert.Equal(t, "server1", newServer.Name())
	assert.Equal(t, newOrg.ID(), newServer.OrganizationID())
	assert.Equal(t, "https://auth.server1", newServer.AuthURL())
	assert.Equal(t, "sftp://fetch.server1", newServer.FetchURL())
	assert.Equal(t, "control.server1:1234", newServer.ControlEndpoint())

	server, err := fsRepo.Get(context.Background(), newServer.ID())
	assert.Nil(t, err)
	assert.Equal(t, "server1", server.Name())

	servers, err := fsRepo.List(context.Background(), models.FileServersQuery{OrgID: refutil.Ref(newOrg.ID())})
	found := false
	for _, s := range servers {
		if s.ID() == server.ID() {
			found = true
		}
	}
	assert.True(t, found)

	err = fsRepo.Remove(context.Background(), server.ID())
	assert.Nil(t, err)
	server, err = fsRepo.Get(context.Background(), newServer.ID())
	assert.Nil(t, server)
	assert.Equal(t, repository.ErrNotFound, err)
}

func TestIntegrationUsers(t *testing.T) {
	db, err := sqlx.Connect("pgx", "postgres://postgres:mysecretpassword@localhost:5432/indexsrv")
	assert.Nil(t, err)

	repo := NewUserRepository(db)

	// Cleanup
	defer db.Query("DELETE FROM users WHERE name = 'user_1'")

	added, err := repo.Add(context.Background(), "user1", "some@pepe.com", "somePassHash")
	assert.Nil(t, err)
	assert.Equal(t, "user1", added.Name())

	fetched, err := repo.Get(context.Background(), added.ID())
	assert.Nil(t, err)
	assert.Equal(t, "user1", fetched.Name())
	assert.Equal(t, "some@pepe.com", fetched.Email())
	assert.Equal(t, "somePassHash", fetched.PasswordHash())

	updated, err := repo.UpdatePassword(context.Background(), fetched.ID(), "newPassHash")
	assert.Nil(t, err)
	assert.Equal(t, "newPassHash", updated.PasswordHash())

	list, err := repo.List(context.Background())
	assert.Nil(t, err)

	found := false
	for _, user := range list {
		if user.Name() == "user1" {
			found = true
		}
	}
	assert.True(t, found)

	err = repo.Remove(context.Background(), fetched.ID())
	assert.Nil(t, err)

	fetched, err = repo.Get(context.Background(), added.ID())
	assert.Equal(t, nil, fetched)
	assert.Equal(t, repository.ErrNotFound, err)
}

func TestIntegrationUserAccounts(t *testing.T) {
	bg := context.Background()
	db, err := sqlx.Connect("pgx", "postgres://postgres:mysecretpassword@localhost:5432/indexsrv")
	assert.Nil(t, err)

	orgRepo := NewOrganizationRepository(db)
	org, err := orgRepo.Add(bg, "test_org_1")
	assert.Nil(t, err)
	defer db.Query("DELETE FROM organizations WHERE name = 'test_org_1'") // cleanup

	fsRepo := NewFileServerRepository(db)
	fs, err := fsRepo.Add(bg, "server1", org.ID(), "https://auth.server1", "https://token.server1", "sftp://fetch.server1", "control.server1:1234")
	assert.Nil(t, err)
	defer db.Query("DELETE FROM file_servers WHERE name = 'server1'") // cleanup

	userRepo := NewUserRepository(db)
	user, err := userRepo.Add(bg, "user1", "user@some.com", "sph")
	assert.Nil(t, err)
	defer db.Query("DELETE FROM users WHERE name = 'user1'") // cleanup

	// initial population done: now test!

	accountRepo := NewUserAccountRepository(db)
	account, err := accountRepo.Add(bg, user.ID(), fs.ID(), "access", "refresh")
	assert.Nil(t, err)
	defer db.Query("DELETE FROM user_accounts WHERE user_id = $1 and  server_id = $2", user.ID(), fs.ID())

	assert.Equal(t, user.ID(), account.UserID())

	assert.Equal(t, fs.ID(), account.FileServerID())

	assert.Equal(t, "access", account.Token())

	assert.Equal(t, "refresh", account.RefreshToken())

	forUser, err := accountRepo.List(bg, user.ID())
	assert.Nil(t, err)

	assert.Equal(t, 1, len(forUser))
	assert.Equal(t, account, forUser[0])

	assert.Equal(t, nil, accountRepo.UpdateCheckpoint(bg, user.ID(), fs.ID(), 123))

	acc, err := accountRepo.Get(bg, user.ID(), fs.ID())
	assert.Nil(t, err)
	assert.Equal(t, int64(123), acc.Checkpoint())
	assert.Equal(t, nil, accountRepo.Remove(bg, user.ID(), fs.ID()))

	list, err := accountRepo.List(bg, user.ID())
	assert.Nil(t, err)
	assert.Equal(t, 0, len(list))
}

func TestIntegrationMappings(t *testing.T) {
	bg := context.Background()
	db, err := sqlx.Connect("pgx", "postgres://postgres:mysecretpassword@localhost:5432/indexsrv")
	assert.Nil(t, err)

	orgRepo := NewOrganizationRepository(db)
	org, err := orgRepo.Add(context.Background(), "test_org_1")
	assert.Nil(t, err)
	defer db.Query("DELETE FROM organizations WHERE name = 'test_org_1'") // cleanup

	fsRepo := NewFileServerRepository(db)
	fs, err := fsRepo.Add(bg, "server1", org.ID(), "https://auth.server1", "https://token.server1", "sftp://fetch.server1", "control.server1:1234")
	assert.Nil(t, err)
	defer db.Query("DELETE FROM file_servers WHERE name = 'server1'") // cleanup

	userRepo := NewUserRepository(db)
	user, err := userRepo.Add(bg, "user1", "user@some.com", "sph")
	assert.Nil(t, err)
	defer db.Query("DELETE FROM users WHERE name = 'user1'") // cleanup

	// DB population done
	// Test begins below:

	repo := NewMappingRepository(db)
	updates := []models.Update{
		{OrganizationID: org.ID(), ServerID: fs.ID(), FileRef: "file1", Checkpoint: 1, ChangeType: models.UpdateTypeFileAdd},
		{OrganizationID: org.ID(), ServerID: fs.ID(), FileRef: "file2", Checkpoint: 2, ChangeType: models.UpdateTypeFileAdd},
		{OrganizationID: org.ID(), ServerID: fs.ID(), FileRef: "file3", Checkpoint: 3, ChangeType: models.UpdateTypeFileAdd},
		{OrganizationID: org.ID(), ServerID: fs.ID(), FileRef: "file4", Checkpoint: 4, ChangeType: models.UpdateTypeFileAdd},
	}

	assert.Equal(t, nil, repo.HandleServerUpdates(bg, user.ID(), updates))
	defer db.Query("DELETE from mappings where server_id = 'server_123'")

	mappings, err := repo.List(bg, user.ID(), models.MappingQuery{})
	assert.Nil(t, err)
	defer db.Query("DELETE FROM mappings WHERE server_id = $1 and user_id = $2", fs.ID(), user.ID())

	assert.Equal(t, 4, len(mappings))

	sort.Slice(mappings, func(i, j int) bool {
		return mappings[i].(models.Mapping).Ref() < mappings[j].(models.Mapping).Ref()
	})

	for idx, mapping := range mappings {
		expectedPath := fmt.Sprintf("unassigned/%s/file%d", fs.ID(), idx+1)
		assert.Equal(t, expectedPath, mapping.Path())
		var expectedCP int64 = int64(idx) + 1
		assert.Equal(t, expectedCP, mapping.Updated().UnixNano())
		assert.False(t, mapping.Deleted())
	}

	// update them and validate chekpoint is updated, rest remains the same

	updates = []models.Update{
		{OrganizationID: org.ID(), ServerID: fs.ID(), FileRef: "file1", Checkpoint: 5, ChangeType: models.UpdateTypeFileUpdate},
		{OrganizationID: org.ID(), ServerID: fs.ID(), FileRef: "file2", Checkpoint: 6, ChangeType: models.UpdateTypeFileUpdate},
		{OrganizationID: org.ID(), ServerID: fs.ID(), FileRef: "file3", Checkpoint: 7, ChangeType: models.UpdateTypeFileUpdate},
		{OrganizationID: org.ID(), ServerID: fs.ID(), FileRef: "file4", Checkpoint: 8, ChangeType: models.UpdateTypeFileUpdate},
	}

	assert.Equal(t, nil, repo.HandleServerUpdates(bg, user.ID(), updates))

	mappings, err = repo.List(bg, user.ID(), models.MappingQuery{})
	assert.Nil(t, err)

	sort.Slice(mappings, func(i, j int) bool {
		return mappings[i].(models.Mapping).Ref() < mappings[j].(models.Mapping).Ref()
	})

	for idx, mapping := range mappings {
		expectedPath := fmt.Sprintf("unassigned/%s/file%d", fs.ID(), idx+1)
		assert.Equal(t, expectedPath, mapping.Path())
		var expectedCP int64 = int64(idx) + 5
		assert.Equal(t, expectedCP, mapping.Updated().UnixNano())
		assert.False(t, mapping.Deleted())
	}

	// delete them and validate checkpoint and status is updated

	updates = []models.Update{
		{OrganizationID: org.ID(), ServerID: fs.ID(), FileRef: "file1", Checkpoint: 9, ChangeType: models.UpdateTypeFileDelete},
		{OrganizationID: org.ID(), ServerID: fs.ID(), FileRef: "file2", Checkpoint: 10, ChangeType: models.UpdateTypeFileDelete},
		{OrganizationID: org.ID(), ServerID: fs.ID(), FileRef: "file3", Checkpoint: 11, ChangeType: models.UpdateTypeFileDelete},
		{OrganizationID: org.ID(), ServerID: fs.ID(), FileRef: "file4", Checkpoint: 12, ChangeType: models.UpdateTypeFileDelete},
	}

	assert.Equal(t, nil, repo.HandleServerUpdates(bg, user.ID(), updates))

	mappings, err = repo.List(bg, user.ID(), models.MappingQuery{})
	assert.Nil(t, err)

	sort.Slice(mappings, func(i, j int) bool {
		return mappings[i].(models.Mapping).Ref() < mappings[j].(models.Mapping).Ref()
	})

	for idx, mapping := range mappings {
		expectedPath := fmt.Sprintf("unassigned/%s/file%d", fs.ID(), idx+1)
		assert.Equal(t, expectedPath, mapping.Path())
		var expectedCP int64 = int64(idx) + 9
		assert.Equal(t, expectedCP, mapping.Updated().UnixNano())

		if !mapping.Deleted() {
			t.Errorf("mapping for ref %s SHOULD be marked as deleted", mapping.Ref())
		}
	}
}

func TestIntegrationPendingOAuth2(t *testing.T) {
	bg := context.Background()
	db, err := sqlx.Connect("pgx", "postgres://postgres:mysecretpassword@localhost:5432/indexsrv")
	assert.Nil(t, err)

	orgRepo := NewOrganizationRepository(db)
	org, err := orgRepo.Add(context.Background(), "test_org_1")
	assert.Nil(t, err)
	defer db.Query("DELETE FROM organizations WHERE name = 'test_org_1'") // cleanup

	fsRepo := NewFileServerRepository(db)
	fs, err := fsRepo.Add(bg, "server1", org.ID(), "https://auth.server1", "https://token.server1", "sftp://fetch.server1", "control.server1:1234")
	assert.Nil(t, err)
	defer db.Query("DELETE FROM file_servers WHERE name = 'server1'") // cleanup

	userRepo := NewUserRepository(db)
	user, err := userRepo.Add(bg, "user1", "user@some.com", "sph")
	assert.Nil(t, err)
	defer db.Query("DELETE FROM users WHERE name = 'user1'") // cleanup

	// DB population done
	// Test begins below:

	repo := NewPendingOAuth2Repository(db)
	inProgress, err := repo.Put(bg, user.ID(), fs.ID(), "qwertyuiop")
	assert.Nil(t, err)

	popped, err := repo.Pop(bg, "qwertyuiop")
	assert.Nil(t, err)
	assert.Equal(t, inProgress, popped)

	_, err = repo.Pop(bg, "qwertyuiop")
	assert.NotNil(t, err)
}

func BenchmarkPSQLMappingInsertion(b *testing.B) {

	rs := rand.NewSource(time.Now().UnixNano())
	r := rand.New(rs)

	db, err := sqlx.Connect("pgx", "postgres://postgres:mysecretpassword@localhost:5432/indexsrv")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	ctx := context.Background()
	userRepo := NewUserRepository(db)
	user, err := userRepo.Add(ctx, "name", fmt.Sprintf("mail_%d", r.Int()), "sph")
	if err != nil {
		panic(err.Error())
	}
	defer userRepo.Remove(ctx, user.ID())

	orgRepo := NewOrganizationRepository(db)
	org, err := orgRepo.Add(ctx, fmt.Sprintf("name_%d", r.Int()))
	if err != nil {
		panic(err.Error())
	}
	defer orgRepo.Remove(ctx, org.ID())

	fsRepo := NewFileServerRepository(db)
	fs, err := fsRepo.Add(ctx, fmt.Sprintf("name_%d", r.Int()), org.ID(), "authURL", "tokenURL", "fetchURL", "controlEndpoint")
	if err != nil {
		panic(err.Error())
	}
	defer fsRepo.Remove(ctx, fs.ID())

	repo := NewMappingRepository(db)

	b.ResetTimer()
	random := r.Int()

	defer db.Query(fmt.Sprintf("DELETE FROM mappings WHERE user_id = '%s'", user.ID()))

	for idx := 0; idx < b.N; idx++ {
		_, err = repo.Add(ctx, user.ID(), &Mapping{
			ServerIDField: fs.ID(),
			PathField:     fmt.Sprintf("%d/path/to/%d", random, idx),
			RefField:      fmt.Sprintf("ref%d_%d", idx, random),
			DeletedField:  false,
			UpdatedField:  time.Now().UnixNano(),
		})
		if err != nil {
			panic(err.Error())
		}

	}
}

func BenchmarkPSQLMappingInsertionConcurrent100(b *testing.B) {
	benchmarkPSQLMappingInsertionConcurrent(b, 100)
}

func BenchmarkPSQLMappingInsertionConcurrent250(b *testing.B) {
	benchmarkPSQLMappingInsertionConcurrent(b, 250)
}

func BenchmarkPSQLMappingInsertionConcurrent500(b *testing.B) {
	benchmarkPSQLMappingInsertionConcurrent(b, 500)
}

func benchmarkPSQLMappingInsertionConcurrent(b *testing.B, concurrency int) {
	b.Helper()

	rs := rand.NewSource(time.Now().UnixNano())
	r := rand.New(rs)

	db, err := sqlx.Connect("pgx", "postgres://postgres:mysecretpassword@localhost:5432/indexsrv")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	db.SetMaxOpenConns(75)

	ctx := context.Background()
	userRepo := NewUserRepository(db)
	user, err := userRepo.Add(ctx, "name", fmt.Sprintf("mail_%d", r.Int()), "sph")
	if err != nil {
		panic(err.Error())
	}
	defer userRepo.Remove(ctx, user.ID())

	orgRepo := NewOrganizationRepository(db)
	org, err := orgRepo.Add(ctx, fmt.Sprintf("name_%d", r.Int()))
	if err != nil {
		panic(err.Error())
	}
	defer orgRepo.Remove(ctx, org.ID())

	fsRepo := NewFileServerRepository(db)
	fs, err := fsRepo.Add(ctx, fmt.Sprintf("name_%d", r.Int()), org.ID(), "authURL", "tokenURL", "fetchURL", "controlEndpoint")
	if err != nil {
		panic(err.Error())
	}
	defer fsRepo.Remove(ctx, fs.ID())

	repo := NewMappingRepository(db)
	defer db.Query(fmt.Sprintf("DELETE FROM mappings WHERE user_id = '%s'", user.ID()))
	// Setup done
	b.ResetTimer()

	// SetParallelism always multiplies `p` by number of CPUs. divide it again to get accurate thread count
	b.SetParallelism(concurrency / runtime.GOMAXPROCS(0))

	mutex := &sync.Mutex{}
	b.RunParallel(func(p *testing.PB) {

		mutex.Lock()
		random := r.Int()
		mutex.Unlock()

		for idx := 0; p.Next(); idx++ {
			mapping := &Mapping{
				ServerIDField: fs.ID(),
				PathField:     fmt.Sprintf("%d/path/to/%d", random, idx),
				RefField:      fmt.Sprintf("ref%d_%d", idx, random),
				DeletedField:  false,
				UpdatedField:  time.Now().UnixNano(),
			}
			_, err = repo.Add(ctx, user.ID(), mapping)
			if err != nil {
				fmt.Printf("%+v\n", mapping)
				panic(err.Error())
			}
		}
	})

}
