package memory

import (
	"errors"
	"fmt"
	"time"

	"github.com/mredolatti/tf/codigo/indexsrv/mapper"
)

// FileSource is an in-memory implementation of a file source
type FileSource struct {
	storage byUserFileStorage
}

// List returns a list of file-description record for a specific user
func (f *FileSource) List(userID string) ([]mapper.File, error) {
	all, err := f.storage.all(userID)
	if err != nil {
		return nil, fmt.Errorf("error fetching all files for user: %w", err)
	}

	return all, nil
}

// GetByID fetches a file-description record by it's id
func (f *FileSource) GetByID(userID string, id string) (mapper.File, error) {
	file, err := f.storage.get(userID, id)
	if err != nil {
		return nil, fmt.Errorf("error fetching filedesc record from storage: %w", err)
	}

	return file, nil
}

// MFile is an in-memory implementation of a file description record
type MFile struct {
	id        string
	ref       string
	size      int64
	patientID string
	updated   time.Time
}

// ID returns the id of a file record
func (m *MFile) ID() string {
	return m.id
}

// Ref (?)
func (m *MFile) Ref() string {
	return m.ref
}

// Size returns the size in bytes of the file
func (m *MFile) Size() int64 {
	return m.size
}

// PatientID returns the id of the patient associated with the file
func (m *MFile) PatientID() string {
	return m.patientID
}

// Updated returns the last time this file was updated in it's server
func (m *MFile) Updated() time.Time {
	return m.updated
}

// ------------------ Auxiliary package-private types

type byIDStorage map[string]mapper.File // id -> file
var errIDAlreadyInUse = errors.New("id already in use")
var errNoSuchID = errors.New("no such id")

func (s *byIDStorage) add(id string, file mapper.File) error {
	if *s == nil {
		*s = make(byIDStorage)
	}

	_, exists := (*s)[id]
	if exists {
		return errIDAlreadyInUse
	}

	(*s)[id] = file
	return nil
}

func (s *byIDStorage) get(id string) (mapper.File, error) {
	if *s == nil {
		return nil, errNoSuchID
	}

	curr, exists := (*s)[id]
	if !exists {
		return nil, errNoSuchID
	}
	return curr, nil
}

func (s *byIDStorage) all() []mapper.File {
	if *s == nil {
		return nil
	}

	all := make([]mapper.File, 0, len(*s))
	for _, v := range *s {
		all = append(all, v)
	}

	return all
}

type byUserFileStorage map[string]byIDStorage // user -> (id -> file)

func (s *byUserFileStorage) add(user string, id string, file mapper.File) error {
	if *s == nil {
		*s = make(byUserFileStorage)
	}
	curr := (*s)[user]
	err := curr.add(id, file)
	(*s)[user] = curr
	return err
}

func (s *byUserFileStorage) get(user string, id string) (mapper.File, error) {
	if *s == nil {
		return nil, errNoSuchUser
	}
	curr, exists := (*s)[user]
	if !exists {
		return nil, errNoSuchUser
	}
	return curr.get(id)
}

func (s *byUserFileStorage) all(user string) ([]mapper.File, error) {
	if *s == nil {
		return nil, errNoSuchUser
	}

	forUser, ok := (*s)[user]
	if !ok {
		return nil, errNoSuchUser
	}

	return forUser.all(), nil
}

// interface compliance compile-time checks
var _ mapper.File = (*MFile)(nil)
var _ mapper.FileSource = (*FileSource)(nil)
