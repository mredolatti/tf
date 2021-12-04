package memory

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/mredolatti/tf/codigo/indexsrv/mapper"
)

// MMapping is an inmplementation of a mapping stored completely in memory
type MMapping struct {
	id      string
	userID  string
	path    string
	file    mapper.File
	updated time.Time
}

// ID returns the id of the mapping
func (m *MMapping) ID() string {
	return m.id
}

// UserID returns the id of the user who owns this mapping
func (m *MMapping) UserID() string {
	return m.userID
}

// Path returns the virtual path as perceived by the user
func (m *MMapping) Path() string {
	return m.path
}

// File returns a pointer to a File-reference object
func (m *MMapping) File() mapper.File {
	return m.file
}

// Updated returns the last time this mapping was updated
func (m *MMapping) Updated() time.Time {
	return m.updated
}

// Impl is a fully in-memory implementation of a mapper
type Impl struct {
	source  mapper.FileSource
	storage byUserStorage
}

// New constructs an in-memory implementation based mapper
func New(source mapper.FileSource) (*Impl, error) {
	if source == nil {
		return nil, fmt.Errorf("source cannot be nil")
	}

	return &Impl{source: source}, nil
}

// Get retrieves a mapping from the in-memory storage
func (i *Impl) Get(userID string, query *mapper.Query) ([]mapper.Mapping, error) {
	forUser, err := i.storage.all(userID)
	if err != nil {
		if errors.Is(err, errNoSuchUser) {
			return nil, mapper.ErrUnknownUser
		}
		return nil, fmt.Errorf("error fetching mapping: %w", err)
	}

	return filterByQuery(forUser, query)
}

// Add adds a new mapping for a tuple(user, file, vpath)
func (i *Impl) Add(userID string, fileID string, path string) error {
	fref, err := i.source.GetByID(userID, fileID)
	if err != nil {
		if errors.Is(err, errNoSuchUser) {
			return mapper.ErrUnknownUser
		} else if errors.Is(err, errNoSuchID) {
			return mapper.ErrFileNotFound
		}
		return fmt.Errorf("error getting file record from source: %w", err)

	}

	err = i.storage.add(userID, path, &MMapping{
		id:      fref.ID(),
		userID:  userID,
		path:    path,
		file:    fref,
		updated: time.Now(),
	})
	if err != nil {
		if errors.Is(err, errPathAlreadyInUse) {
			return mapper.ErrMappingExists
		}
		return fmt.Errorf("error storing mapping: %w", err)
	}

	return nil
}

// Remove deletes a specific mapping for a certain user
func (i *Impl) Remove(userID, mappingID string) error {
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
	if *s == nil {
		return nil, errNoSuchPath
	}
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
	if *s == nil {
		return nil, errNoSuchUser
	}

	curr, exists := (*s)[user]
	if !exists {
		return nil, errNoSuchUser
	}
	return curr.get(path)
}

func (s *byUserStorage) all(user string) ([]mapper.Mapping, error) {
	if *s == nil {
		return nil, errNoSuchUser
	}

	curr, exists := (*s)[user]
	if !exists {
		return nil, errNoSuchUser
	}

	res := make([]mapper.Mapping, 0, len(curr))
	for path := range curr {
		mapping := curr[path]
		res = append(res, &mapping)
	}

	return res, nil
}

type filterFunc = func(mapper.Mapping) bool
type filterSlice []filterFunc

func (f *filterSlice) add(ff filterFunc) {
	if *f == nil {
		*f = make([]filterFunc, 0, 5)
	}

	*f = append(*f, ff)
}

func (f *filterSlice) apply(m mapper.Mapping) bool {
	if *f == nil {
		return true
	}

	for _, filter := range *f {
		if !filter(m) {
			return false
		}
	}
	return true
}

func filterByQuery(input []mapper.Mapping, query *mapper.Query) ([]mapper.Mapping, error) {

	if query == nil {
		return nil, fmt.Errorf("query cannot be nil")
	}

	var filters filterSlice
	if query.FileID != nil {
		filters.add(func(m mapper.Mapping) bool { return m.File().ID() == *query.FileID })
	}
	if query.Path != nil {
		filters.add(func(m mapper.Mapping) bool { return strings.HasPrefix(m.Path(), *query.Path) })
	}
	if query.PatientID != nil {
		// TODO!
	}
	if query.UpdatedAfter != nil {
		filters.add(func(m mapper.Mapping) bool { return m.Updated().After(*query.UpdatedAfter) })
	}
	if query.UpdatedBefore != nil {
		filters.add(func(m mapper.Mapping) bool { return m.Updated().Before(*query.UpdatedAfter) })
	}

	output := make([]mapper.Mapping, 0, len(input))
	for _, m := range input {
		if filters.apply(m) {
			output = append(output, m)
		}
	}
	return output, nil
}

// interface compliance compile-time checks
var _ mapper.Mapping = (*MMapping)(nil)
var _ mapper.Interface = (*Impl)(nil)
