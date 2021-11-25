package memory

import (
	"errors"
	"time"

	"github.com/mredolatti/tf/codigo/indexsrv/mapper"
)

type MFile struct {
	id        string
	ref       string
	size      int64
	patientID string
	updated   time.Time
}

func (m *MFile) ID() string {
	return m.id
}

func (m *MFile) Ref() string {
	return m.ref
}

func (m *MFile) Size() int64 {
	return m.size
}

func (m *MFile) PatientID() string {
	return m.patientID
}

func (m *MFile) Updated() time.Time {
	return m.updated
}

type MMapping struct {
	id      string
	userID  string
	path    string
	file    *mapper.File
	updated time.Time
}

func (m *MMapping) ID() string {
	return m.id
}

func (m *MMapping) UserId() string {
	return m.userID
}

func (m *MMapping) Path() string {
	return m.path
}

func (m *MMapping) File() *mapper.File {
	return m.file
}

func (m *MMapping) Updated() time.Time {
	return m.updated
}

type Impl struct {
}

func (i *Impl) Get(userID string, query *mapper.Query) ([]mapper.Mapping, error) {
	return nil, nil
}

func (i *Impl) Add(userID string, fileID string, path string) error {

	return nil
}

func (m *Impl) Remove(userID, mappingID string) error {
	return nil
}

// ------------------ Auxiliary package-private types

type byPathStorage map[string]MMapping // path -> mapping

var errPathAlreadyInUse = errors.New("path already in use")
var errNoSuchPath = errors.New("no such path")

func (s *byPathStorage) add(path string, mapping *MMapping) error {
	if *s == nil {
		*s = make(byPathStorage)
	}

	_, exists := (*s)[path]
	if exists {
		return errPathAlreadyInUse
	}

	(*s)[path] = *mapping
	return nil
}

func (s *byPathStorage) get(path string) (*MMapping, error) {
	curr, exists := (*s)[path]
	if !exists {
		return nil, errNoSuchPath
	}
	return &curr, nil
}

type byUserStorage map[string]byPathStorage // user -> (path -> mapping)

var errNoSuchUser = errors.New("no such user")

func (s *byUserStorage) add(user string, path string, mapping *MMapping) error {
	if *s == nil {
		*s = make(byUserStorage)
	}
	curr := (*s)[user]
	err := curr.add(path, mapping)
	(*s)[user] = curr
	return err
}

func (s *byUserStorage) get(user string, path string) (*MMapping, error) {
	curr, exists := (*s)[user]
	if !exists {
		return nil, errNoSuchUser
	}
	return curr.get(path)
}

var _ mapper.File = (*MFile)(nil)
var _ mapper.Mapping = (*MMapping)(nil)
var _ mapper.Interface = (*Impl)(nil)
