package memory

import (
	"errors"
	"testing"
)

func TestInMemoryFileSource(t *testing.T) {
	var s FileSource

	if res, err := s.List("user1"); res != nil || !errors.Is(err, errNoSuchUser) {
		t.Error("user1 should have nothing yet. Got: ", res, err)
	}

	s.storage.add("user1", "id1", &MFile{id: "id1"})
	s.storage.add("user1", "id2", &MFile{id: "id2"})
	s.storage.add("user1", "id3", &MFile{id: "id3"})
	s.storage.add("user1", "id4", &MFile{id: "id4"})

	if res, err := s.List("user1"); err != nil || len(res) != 4 {
		t.Error("there shold be no errors and 4 items. got: ", res, err)
	}

	if elem, err := s.GetByID("user1", "id3"); err != nil || elem == nil || elem.ID() != "id3" {
		t.Error("expected no error and an element with id=id3. Got: ", elem, err)
	}
}
