package memory

import (
	"testing"
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
