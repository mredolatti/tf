package memory

import (
	"testing"

	"github.com/mredolatti/tf/codigo/common/refutil"
	"github.com/mredolatti/tf/codigo/indexsrv/mapper"
)

func TestStorageByPath(t *testing.T) {
	var s byPathStorage
	if err := s.add("/path/to/something.doc", &MMapping{id: "someId"}); err != nil {
		t.Error("should not return an error. Got: ", err)
	}

	elem, err := s.get("/path/to/something.doc")
	if err != nil {
		t.Error("should not return an error. Got: ", err)
	}

	if elem != nil && elem.id != "someId" {
		t.Errorf("id should be 123. is: %+v", *elem)
	}
}

func TestStorageByUser(t *testing.T) {
	var s byUserStorage
	if err := s.add("user1", "/path/to/file1", &MMapping{id: "id1"}); err != nil {
		t.Error("shold not return an error. Got: ", err)
	}

	if err := s.add("user1", "/path/to/file1", &MMapping{id: "id2"}); err != errPathAlreadyInUse {
		t.Error("shold notify that path is already in use. Got: ", err)
	}

	if err := s.add("user1", "/path/to/file2", &MMapping{id: "id3"}); err != nil {
		t.Error("shold not return an error. Got: ", err)
	}

	if err := s.add("user2", "/path/to/file2", &MMapping{id: "id4"}); err != nil {
		t.Error("shold not return an error. Got: ", err)
	}

	u1f1, err := s.get("user1", "/path/to/file1")
	if err != nil {
		t.Error("shold not return an error. Got: ", err)
	}
	if u1f1 != nil && u1f1.id != "id1" {
		t.Error("id should be id1. Is: ", u1f1.id)
	}
	u1f3, err := s.get("user1", "/path/to/file2")
	if err != nil {
		t.Error("shold not return an error. Got: ", err)
	}
	if u1f3 != nil && u1f3.id != "id3" {
		t.Error("id should be id3. Is: ", u1f3.id)
	}
	u2f2, err := s.get("user2", "/path/to/file2")
	if err != nil {
		t.Error("shold not return an error. Got: ", err)
	}
	if u2f2 != nil && u2f2.id != "id4" {
		t.Error("id should be id4. Is: ", u2f2.id)
	}
}

func TestMemoryMapper(t *testing.T) {
	var s Impl

	// populate the file storage first, so that the mapper is aware of some files
	source := &FileSource{}
	s.source = source
	source.storage.add("user1", "id1", &MFile{id: "id1"})
	source.storage.add("user1", "id2", &MFile{id: "id2"})
	source.storage.add("user1", "id3", &MFile{id: "id3"})
	source.storage.add("user2", "id3", &MFile{id: "id3"})
	source.storage.add("user2", "id4", &MFile{id: "id4"})
	source.storage.add("user3", "id4", &MFile{id: "id4"})

	if err := s.Add("user1", "id1", "/my/path/someFile1"); err != nil {
		t.Error("err should be nil. Got: ", err)
	}
	if err := s.Add("user1", "id2", "/my/path/someFile2"); err != nil {
		t.Error("err should be nil. Got: ", err)
	}
	if err := s.Add("user1", "id3", "/my/otherPath/someFile3"); err != nil {
		t.Error("err should be nil. Got: ", err)
	}
	if err := s.Add("user2", "id3", "/my/otherPath/someFile3"); err != nil {
		t.Error("err should be nil. Got: ", err)
	}
	if err := s.Add("user2", "id4", "/my/otherPath/someFile4"); err != nil {
		t.Error("err should be nil. Got: ", err)
	}
	if err := s.Add("user3", "id4", "/my/otherPath/someFile4"); err != nil {
		t.Error("err should be nil. Got: ", err)
	}

	// Error cases
	if err := s.Add("user1", "id1", "/my/path/someFile1"); err != mapper.ErrMappingExists {
		t.Error("shoud have got `MappingExits` error. Got: ", err)
	}
	if err := s.Add("user1", "id4", "/my/otherPath/someFile4"); err != mapper.ErrFileNotFound {
		t.Error("err should be `FileNotFound`. Got: ", err)
	}

	if res, err := s.Get("user1", &mapper.Query{}); len(res) != 3 || err != nil {
		t.Error("there should be no no error and 3 results. Got: ", res, err)
	}
	if res, err := s.Get("user1", &mapper.Query{Path: refutil.StrRef("/my/path")}); len(res) != 2 || err != nil {
		t.Error("should have 2 results and no error. Got: ", res, err)
	}

	if res, err := s.Get("user1", &mapper.Query{Path: refutil.StrRef("/my/otherPath")}); len(res) != 1 || err != nil {
		t.Error("should have 1 result and no error. Got: ", res, err)
	}

	// TODO: test other query filter criteria!
}
