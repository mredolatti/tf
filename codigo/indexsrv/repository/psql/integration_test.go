package psql

import (
	"context"
	"math/rand"
	"strconv"
	"testing"
	"time"

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

	added, err := repo.Add(context.Background(), "test_org_1")
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

	newOrg, err := orgRepo.Add(context.Background(), "test_org_1")
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
