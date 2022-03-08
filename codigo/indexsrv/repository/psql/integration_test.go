package psql

import (
	"context"
	"fmt"
	"math/rand"
	"reflect"
	"sort"
	"strconv"
	"testing"
	"time"

	"github.com/mredolatti/tf/codigo/indexsrv/models"
	"github.com/mredolatti/tf/codigo/indexsrv/repository"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
)

func TestOrgRepoIntegration(t *testing.T) {
	db, err := sqlx.Connect("pgx", "postgres://postgres:mysecretpassword@localhost:5432/indexsrv")
	if err != nil {
		t.Error("a postgres db is required for these tests: ", err)
	}

	repo, err := NewOrganizationRepository(db)
	if err != nil {
		t.Error("no error shold be returned with a non-nil db. Got: ", err)
	}

	// Cleanup
	defer db.Query("DELETE FROM organizations WHERE name = 'test_org_1'")

	added, err := repo.Add(context.Background(), &Organization{NameField: "test_org_1"})
	if err != nil {
		t.Error("expected no error. Got: ", err)
	}
	if added.Name() != "test_org_1" {
		t.Error("incorrect name. Got: ", added.Name())
	}

	fetched, err := repo.Get(context.Background(), added.ID())
	if err != nil {
		t.Error("expected no error. Got: ", err)
	}
	if fetched.Name() != "test_org_1" {
		t.Error("wrong name. Got: ", fetched.Name())
	}

	list, err := repo.List(context.Background())
	if err != nil {
		t.Error("expected no error. Got: ", err)
	}

	found := false
	for _, org := range list {
		if org.Name() == "test_org_1" {
			found = true
		}
	}
	if !found {
		t.Errorf("test_org_1 not found in list. Got: %+v", list)
	}

	err = repo.Remove(context.Background(), fetched.ID())
	if err != nil {
		t.Error("expected no error. Got: ", err)
	}

	fetched, err = repo.Get(context.Background(), added.ID())
	if fetched != nil {
		t.Errorf("fetched shold be nil. Is: %T %+v", fetched, fetched)
	}
	if err != repository.ErrNotFound {
		t.Error("expected ErrNotFOund. Got: ", err)
	}
}

func TestFileServerRepoIntegration(t *testing.T) {
	db, err := sqlx.Connect("pgx", "postgres://postgres:mysecretpassword@localhost:5432/indexsrv")
	if err != nil {
		t.Error("a postgres db is required for these tests: ", err)
	}

	orgRepo, err := NewOrganizationRepository(db)
	if err != nil {
		t.Error("no error shold be returned with a non-nil db. Got: ", err)
	}

	// Cleanup
	defer func() {
		db.Query("DELETE FROM organizations WHERE name = 'test_org_1'")
	}()

	newOrg, err := orgRepo.Add(context.Background(), &Organization{NameField: "test_org_1"})
	if err != nil {
		t.Error("expected no error. Got: ", err)
	}
	if newOrg.Name() != "test_org_1" {
		t.Error("incorrect name. Got: ", newOrg.Name())
	}

	// Begin testing the file server repo

	fsRepo, err := NewFileServerRepository(db)
	if err != nil {
		t.Error("there shold be no error. got: ", err)
	}

	rand.Seed(time.Now().UnixNano())
	newID := strconv.FormatInt(rand.Int63(), 10)
	newServer, err := fsRepo.Add(context.Background(), newID, "server1", newOrg.ID(), "https://auth.server1", "sftp://fetch.server1", "control.server1:1234")
	if err != nil {
		t.Error("there shold be no error. got: ", err)
	}

	if id := newServer.ID(); id != newID {
		t.Error("wrong id. Got: ", id)
	}

	if name := newServer.Name(); name != "server1" {
		t.Error("wrong name. Got: ", name)
	}

	if orgID := newServer.OrganizationID(); orgID != newOrg.ID() {
		t.Error("wrong org. Got: ", orgID)
	}

	if url := newServer.AuthURL(); url != "https://auth.server1" {
		t.Error("wrong auth url. Got: ", url)
	}

	if url := newServer.FetchURL(); url != "sftp://fetch.server1" {
		t.Error("wrong fetch url. Got: ", url)
	}

	if url := newServer.ControlEndpoint(); url != "control.server1:1234" {
		t.Error("wrong control endpoint. Got: ", url)
	}

	server, err := fsRepo.Get(context.Background(), newServer.ID())
	if err != nil {
		t.Error("there shold be no error. got: ", err)
	}
	if name := server.Name(); name != "server1" {
		t.Error("wrong name: ", name)
	}

	servers, err := fsRepo.List(context.Background(), newOrg.ID())
	found := false
	for _, s := range servers {
		if s.ID() == server.ID() {
			found = true
		}
	}
	if !found {
		t.Error("server not found in List() result.")
	}

	err = fsRepo.Remove(context.Background(), server.ID())
	if err != nil {
		t.Error("there shold be no error. got: ", err)
	}
	server, err = fsRepo.Get(context.Background(), newServer.ID())
	if err != repository.ErrNotFound || server != nil {
		t.Error("there shold be no error. got: ", server, err)
	}
}

func TestIntegrationUsers(t *testing.T) {
	db, err := sqlx.Connect("pgx", "postgres://postgres:mysecretpassword@localhost:5432/indexsrv")
	if err != nil {
		t.Error("a postgres db is required for these tests: ", err)
	}

	repo, err := NewUserRepository(db)
	if err != nil {
		t.Error("no error shold be returned with a non-nil db. Got: ", err)
	}

	// Cleanup
	defer db.Query("DELETE FROM users WHERE name = 'user_1'")

	added, err := repo.Add(context.Background(), "some_id", "user1", "some@pepe.com", "qwerty", "asdfg")
	if err != nil {
		t.Error("expected no error. Got: ", err)
	}
	if added.Name() != "user1" {
		t.Error("incorrect name. Got: ", added.Name())
	}

	fetched, err := repo.Get(context.Background(), added.ID())
	if err != nil {
		t.Error("expected no error. Got: ", err)
	}
	if fetched.Name() != "user1" {
		t.Error("wrong name. Got: ", fetched.Name())
	}

	if fetched.Email() != "some@pepe.com" || fetched.AccessToken() != "qwerty" || fetched.RefreshToken() != "asdfg" {
		t.Error("wrong mail or tokens. Got: ", fetched)
	}

	updated, err := repo.UpdateTokens(context.Background(), fetched.ID(), "ytrewq", "gfdsa")
	if err != nil {
		t.Error("expected no error. Got: ", err)
	}

	if updated.AccessToken() != "ytrewq" || updated.RefreshToken() != "gfdsa" {
		t.Error("wrong tokens. Got: ", updated)
	}

	list, err := repo.List(context.Background())
	if err != nil {
		t.Error("expected no error. Got: ", err)
	}

	found := false
	for _, user := range list {
		if user.Name() == "user1" {
			found = true
		}
	}
	if !found {
		t.Errorf("user1 not found in list. Got: %+v", list)
	}

	err = repo.Remove(context.Background(), fetched.ID())
	if err != nil {
		t.Error("expected no error. Got: ", err)
	}

	fetched, err = repo.Get(context.Background(), added.ID())
	if fetched != nil {
		t.Errorf("fetched shold be nil. Is: %T %+v", fetched, fetched)
	}
	if err != repository.ErrNotFound {
		t.Error("expected ErrNotFOund. Got: ", err)
	}
}

func TestIntegrationUserAccounts(t *testing.T) {
	bg := context.Background()
	db, err := sqlx.Connect("pgx", "postgres://postgres:mysecretpassword@localhost:5432/indexsrv")
	if err != nil {
		t.Error("a postgres db is required for these tests: ", err)
	}

	orgRepo, err := NewOrganizationRepository(db)
	if err != nil {
		t.Error("no error shold be returned with a non-nil db. Got: ", err)
	}
	org, err := orgRepo.Add(bg, &Organization{NameField: "test_org_1"})
	if err != nil {
		t.Error("expected no error. Got: ", err)
	}
	defer db.Query("DELETE FROM organizations WHERE name = 'test_org_1'") // cleanup

	fsRepo, err := NewFileServerRepository(db)
	if err != nil {
		t.Error("there shold be no error. got: ", err)
	}
	fs, err := fsRepo.Add(bg, "server_123", "server1", org.ID(), "https://auth.server1", "sftp://fetch.server1", "control.server1:1234")
	if err != nil {
		t.Error("there shold be no error. got: ", err)
	}
	defer db.Query("DELETE FROM file_servers WHERE id = 'server_123'") // cleanup

	userRepo, err := NewUserRepository(db)
	if err != nil {
		t.Error("no error shold be returned with a non-nil db. Got: ", err)
	}
	user, err := userRepo.Add(bg, "user_123", "user1", "user@some.com", "", "")
	if err != nil {
		t.Error("erro creating user: ", err)
	}
	defer db.Query("DELETE FROM users WHERE id = 'user_123'") // cleanup

	// initial population done: now test!

	accountRepo, err := NewUserAccountRepository(db)
	if err != nil {
		t.Error("no error shold be returned with a non-nil db. Got: ", err)
	}

	account, err := accountRepo.Add(bg, user.ID(), fs.ID(), "access", "refresh")
	if err != nil {
		t.Error("there should be no error on creation: ", err)
	}
	defer db.Query("DELETE FROM user_accounts WHERE user_id = $1 and  server_id = $2", user.ID(), fs.ID())

	if uid := account.UserID(); uid != user.ID() {
		t.Error("wrong user: ", uid)
	}

	if sid := account.FileServerID(); sid != fs.ID() {
		t.Error("wrong file server ID: ", sid)
	}

	if at := account.Token(); at != "access" {
		t.Error("wrong access token: ", at)
	}

	if rt := account.RefreshToken(); rt != "refresh" {
		t.Error("wrong refresh token: ", rt)
	}

	forUser, err := accountRepo.List(bg, user.ID())
	if err != nil {
		t.Error("list should not error. Got: ", err)
	}

	if l := len(forUser); l != 1 {
		t.Error("there should be 1 element only. Got: ", l)
	}

	if !reflect.DeepEqual(forUser[0], account) {
		t.Errorf("list[0] and account fetched by Get() should be the same. Got:\nfrom list={%+v}\nfrom get:{%+v}",
			forUser[0], account)
	}

	if err := accountRepo.UpdateCheckpoint(bg, user.ID(), fs.ID(), 123); err != nil {
		t.Error("no error should have been returned. got: ", err)
	}

	acc, err := accountRepo.Get(bg, user.ID(), fs.ID())
	if err != nil {
		t.Error("should not fail. Got: ", err)
	}

	if cp := acc.Checkpoint(); cp != 123 {
		t.Error("checkpoint shold have been updated to 123. Got: ", cp)
	}

	if err := accountRepo.Remove(bg, user.ID(), fs.ID()); err != nil {
		t.Error("there should be no error when deleteing an account. Got: ", err)
	}

	if list, err := accountRepo.List(bg, user.ID()); err != nil || len(list) != 0 {
		t.Error("there should be no error, and the list should be empty. Got: ", err, list)
	}
}

func TestIntegrationMappings(t *testing.T) {
	bg := context.Background()
	db, err := sqlx.Connect("pgx", "postgres://postgres:mysecretpassword@localhost:5432/indexsrv")
	if err != nil {
		t.Error("a postgres db is required for these tests: ", err)
	}

	orgRepo, err := NewOrganizationRepository(db)
	if err != nil {
		t.Error("no error shold be returned with a non-nil db. Got: ", err)
	}
	org, err := orgRepo.Add(context.Background(), &Organization{NameField: "test_org_1"})
	if err != nil {
		t.Error("expected no error. Got: ", err)
	}
	defer db.Query("DELETE FROM organizations WHERE name = 'test_org_1'") // cleanup

	fsRepo, err := NewFileServerRepository(db)
	if err != nil {
		t.Error("there shold be no error. got: ", err)
	}
	fs, err := fsRepo.Add(context.Background(), "server_123", "server1", org.ID(), "https://auth.server1", "sftp://fetch.server1", "control.server1:1234")
	if err != nil {
		t.Error("there shold be no error. got: ", err)
	}
	defer db.Query("DELETE FROM file_servers WHERE id = 'server_123'") // cleanup

	userRepo, err := NewUserRepository(db)
	if err != nil {
		t.Error("no error shold be returned with a non-nil db. Got: ", err)
	}
	user, err := userRepo.Add(bg, "user_123", "user1", "user@some.com", "", "")
	if err != nil {
		t.Error("erro creating user: ", err)
	}
	defer db.Query("DELETE FROM users WHERE id = 'user_123'") // cleanup

	// DB population done
	// Test begins below:

	repo, err := NewMappingRepository(db)
	if err != nil {
		t.Error("no error shold be returned with a non-nil db. Got: ", err)
	}

	updates := []models.Update{
		{OrganizationID: org.ID(), ServerID: fs.ID(), FileRef: "file1", Checkpoint: 1, ChangeType: models.UpdateTypeFileAdd},
		{OrganizationID: org.ID(), ServerID: fs.ID(), FileRef: "file2", Checkpoint: 2, ChangeType: models.UpdateTypeFileAdd},
		{OrganizationID: org.ID(), ServerID: fs.ID(), FileRef: "file3", Checkpoint: 3, ChangeType: models.UpdateTypeFileAdd},
		{OrganizationID: org.ID(), ServerID: fs.ID(), FileRef: "file4", Checkpoint: 4, ChangeType: models.UpdateTypeFileAdd},
	}

	if err := repo.HandleServerUpdates(bg, user.ID(), updates); err != nil {
		t.Error("error updating mappings: ", err)
	}
	defer db.Query("DELETE from mappings where server_id = 'server_123'")

	mappings, err := repo.List(bg, user.ID(), models.MappingQuery{})
	if err != nil {
		t.Error("error fetching mappings inserted")
	}
	defer db.Query("DELETE FROM mappings WHERE server_id = $1 and user_id = $2", fs.ID(), user.ID())

	if l := len(mappings); l != 4 {
		t.Error("4 new mappings should have been added. Got: ", l)
	}

	sort.Slice(mappings, func(i, j int) bool {
		return mappings[i].(models.Mapping).Ref() < mappings[j].(models.Mapping).Ref()
	})

	for idx, mapping := range mappings {
		expectedPath := fmt.Sprintf("unnasigned/%s/file%d", fs.ID(), idx+1)
		if p := mapping.Path(); p != expectedPath {
			t.Errorf("expected path = %s. got: %s", expectedPath, p)
		}
		var expectedCP int64 = int64(idx) + 1
		if cp := mapping.Updated().UnixNano(); cp != expectedCP {
			t.Errorf("expected checkpoint %d. got: %d", expectedCP, cp)
		}

		if mapping.Deleted() {
			t.Errorf("mapping for ref %s should not be marked as deleted", mapping.Ref())
		}
	}

	// update them and validate chekpoint is updated, rest remains the same

	updates = []models.Update{
		{OrganizationID: org.ID(), ServerID: fs.ID(), FileRef: "file1", Checkpoint: 5, ChangeType: models.UpdateTypeFileUpdate},
		{OrganizationID: org.ID(), ServerID: fs.ID(), FileRef: "file2", Checkpoint: 6, ChangeType: models.UpdateTypeFileUpdate},
		{OrganizationID: org.ID(), ServerID: fs.ID(), FileRef: "file3", Checkpoint: 7, ChangeType: models.UpdateTypeFileUpdate},
		{OrganizationID: org.ID(), ServerID: fs.ID(), FileRef: "file4", Checkpoint: 8, ChangeType: models.UpdateTypeFileUpdate},
	}

	if err := repo.HandleServerUpdates(bg, user.ID(), updates); err != nil {
		t.Error("error updating mappings: ", err)
	}

	mappings, err = repo.List(bg, user.ID(), models.MappingQuery{})
	if err != nil {
		t.Error("error fetching mappings inserted")
	}

	sort.Slice(mappings, func(i, j int) bool {
		return mappings[i].(models.Mapping).Ref() < mappings[j].(models.Mapping).Ref()
	})

	for idx, mapping := range mappings {
		expectedPath := fmt.Sprintf("unnasigned/%s/file%d", fs.ID(), idx+1)
		if p := mapping.Path(); p != expectedPath {
			t.Errorf("expected path = %s. got: %s", expectedPath, p)
		}
		var expectedCP int64 = int64(idx) + 5
		if cp := mapping.Updated().UnixNano(); cp != expectedCP {
			t.Errorf("expected checkpoint %d. got: %d", expectedCP, cp)
		}

		if mapping.Deleted() {
			t.Errorf("mapping for ref %s should not be marked as deleted", mapping.Ref())
		}
	}

	// delete them and validate checkpoint and status is updated

	updates = []models.Update{
		{OrganizationID: org.ID(), ServerID: fs.ID(), FileRef: "file1", Checkpoint: 9, ChangeType: models.UpdateTypeFileDelete},
		{OrganizationID: org.ID(), ServerID: fs.ID(), FileRef: "file2", Checkpoint: 10, ChangeType: models.UpdateTypeFileDelete},
		{OrganizationID: org.ID(), ServerID: fs.ID(), FileRef: "file3", Checkpoint: 11, ChangeType: models.UpdateTypeFileDelete},
		{OrganizationID: org.ID(), ServerID: fs.ID(), FileRef: "file4", Checkpoint: 12, ChangeType: models.UpdateTypeFileDelete},
	}

	if err := repo.HandleServerUpdates(bg, user.ID(), updates); err != nil {
		t.Error("error updating mappings: ", err)
	}

	mappings, err = repo.List(bg, user.ID(), models.MappingQuery{})
	if err != nil {
		t.Error("error fetching mappings inserted")
	}

	sort.Slice(mappings, func(i, j int) bool {
		return mappings[i].(models.Mapping).Ref() < mappings[j].(models.Mapping).Ref()
	})

	for idx, mapping := range mappings {
		expectedPath := fmt.Sprintf("unnasigned/%s/file%d", fs.ID(), idx+1)
		if p := mapping.Path(); p != expectedPath {
			t.Errorf("expected path = %s. got: %s", expectedPath, p)
		}
		var expectedCP int64 = int64(idx) + 9
		if cp := mapping.Updated().UnixNano(); cp != expectedCP {
			t.Errorf("expected checkpoint %d. got: %d", expectedCP, cp)
		}

		if !mapping.Deleted() {
			t.Errorf("mapping for ref %s SHOULD be marked as deleted", mapping.Ref())
		}
	}
}
