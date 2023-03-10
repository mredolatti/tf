package basic

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/mredolatti/tf/codigo/fileserver/models"
	"github.com/mredolatti/tf/codigo/fileserver/storage"
)

// InMemoryFileStore implements the storage interface in memory
type InMemoryFileStore struct {
	files map[string]InMemoryFile
	mtx   sync.RWMutex
}

// NewInMemoryFileStore creates a new in-memory file store
func NewInMemoryFileStore() *InMemoryFileStore {
	return &InMemoryFileStore{files: make(map[string]InMemoryFile)}
}

// Read attempts to get the contents of a file
func (s *InMemoryFileStore) Read(id string) ([]byte, error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	r, ok := s.files[id]
	if !ok {
		return nil, storage.ErrNoSuchFile
	}

	cpy := make([]byte, len(r.contents))
	copy(cpy, r.contents)
	return cpy, nil
}

// Write attempts to update the contents of a file
func (s *InMemoryFileStore) Write(id string, data []byte, force bool) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	if !force {
		_, ok := s.files[id]
		if ok {
			return storage.ErrFileExists
		}
	}

	cpy := make([]byte, len(data))
	copy(cpy, data)
	s.files[id] = InMemoryFile{
		id:       id,
		contents: cpy,
	}
	return nil
}

// Del deletes a file
func (s *InMemoryFileStore) Del(id string) error {
	s.mtx.Lock()
	delete(s.files, id)
	s.mtx.Unlock()
	return nil
}

// InMemoryFile is an in-memory implentation of a file contents record
type InMemoryFile struct {
	id       string
	contents []byte
}

// ID returns the id
func (i *InMemoryFile) ID() string {
	return i.id
}

// Contents returns the contents of the file
func (i *InMemoryFile) Contents() []byte {
	return i.contents
}

// InMemoryFileMetadataStore is an in-memory implementation of a file-metadata store
type InMemoryFileMetadataStore struct {
	metas  map[string]InMemoryMetadata
	lastID int
	mutex  sync.Mutex
}

// NewInMemoryFileMetadataStore creates a new in-memory file-metadata store
func NewInMemoryFileMetadataStore() *InMemoryFileMetadataStore {
	return &InMemoryFileMetadataStore{
		metas:  make(map[string]InMemoryMetadata),
		lastID: 0,
	}
}

// Get fetches file-metadata for a specific id
func (i *InMemoryFileMetadataStore) Get(id string) (models.FileMetadata, error) {
	i.mutex.Lock()
	m, ok := i.metas[id]
	i.mutex.Unlock()

	if !ok {
		return nil, storage.ErrNoSuchFile
	}

	return &m, nil
}

// GetMany fetches file-metadata for multiple ids
func (i *InMemoryFileMetadataStore) GetMany(filter *storage.Filter) (map[string]models.FileMetadata, error) {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	if filter == nil { // No filter: return everything in a new map
		result := make(map[string]models.FileMetadata, len(i.metas))
		for id, m := range i.metas {
			result[id] = &m
		}
		return result, nil
	}

	return i.getByFilter(filter), nil
}

// Create adds a new file-metadata record
func (i *InMemoryFileMetadataStore) Create(name string, notes string, patient string, typ string, whenNs int64) (models.FileMetadata, error) {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	i.lastID++
	id := strconv.Itoa(i.lastID)

	_, ok := i.metas[id]
	if ok {
		return nil, fmt.Errorf("this is most likely a bug, another file exists with the last generated id: %s", id)
	}

	m := InMemoryMetadata{id: id, name: name, notes: notes, patientID: patient, typ: typ, lastUpdated: whenNs}
	i.metas[id] = m
	return &m, nil
}

// Update modifies an existing metadata record
func (i *InMemoryFileMetadataStore) Update(id string, updated models.FileMetadata, whenNs int64) (models.FileMetadata, error) {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	m, ok := i.metas[id]
	if !ok {
		return nil, storage.ErrNoSuchFile
	}

	m.name = updated.Name()
	m.notes = updated.Notes()
	m.patientID = updated.PatientID()
	m.typ = updated.Type()
	m.lastUpdated = whenNs
	i.metas[id] = m

	return &m, nil
}

// Remove deletes a metadata record
func (i *InMemoryFileMetadataStore) Remove(id string, whenNs int64) error {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	toDelete, ok := i.metas[id]
	if !ok {
		return nil
	}

	toDelete.lastUpdated = whenNs
	toDelete.deleted = true
	i.metas[id] = toDelete

	return nil
}

func (i *InMemoryFileMetadataStore) getByFilter(filter *storage.Filter) map[string]models.FileMetadata {
	if length := len(filter.IDs); length > 0 { // If ID list is specified
		result := make(map[string]models.FileMetadata, length)
		for _, id := range filter.IDs {
			m, ok := i.metas[id]
			if ok && filterMatches(filter, &m) {
				result[id] = &m
			}
		}
		return result
	}

	// no id filter, iterate all collection
	result := make(map[string]models.FileMetadata, len(i.metas)/2) // approx
	for id, m := range i.metas {
		if filterMatches(filter, &m) {
			result[id] = &m
		}
	}
	return result
}

func filterMatches(filter *storage.Filter, item *InMemoryMetadata) bool {
	if filter.UpdatedAfter == nil {
		return true
	}

	return item.LastUpdated() > *filter.UpdatedAfter
}

// InMemoryMetadata is an in-memory representation of a file metadata
type InMemoryMetadata struct {
	id          string
	name        string
	notes       string
	patientID   string
	sizeBytes   int64
	typ         string
	contentID   string
	lastUpdated int64
	deleted     bool
}

// ID returns the file id
func (i *InMemoryMetadata) ID() string {
	return i.id
}

// Name returns the file name
func (i *InMemoryMetadata) Name() string {
	return i.name
}

// Notes returns the notes associated to the file
func (i *InMemoryMetadata) Notes() string {
	return i.notes
}

func (i *InMemoryMetadata) SizeBytes() int64 {
	return i.sizeBytes
}

// PatientID returns the id of the patient associated to this file
func (i *InMemoryMetadata) PatientID() string {
	return i.patientID
}

// Type returns the type of file
func (i *InMemoryMetadata) Type() string {
	return i.typ
}

// ContentID returns the id of the content of the file (if any)
func (i *InMemoryMetadata) ContentID() string {
	return i.contentID
}

// LastUpdated returns the timestamp of the last update
func (i *InMemoryMetadata) LastUpdated() int64 {
	return i.lastUpdated
}

// Deleted returns true if the file has been deleted
func (i *InMemoryMetadata) Deleted() bool {
	return i.deleted
}

var _ models.File = (*InMemoryFile)(nil)
var _ models.FileMetadata = (*InMemoryMetadata)(nil)
var _ storage.FilesMetadata = (*InMemoryFileMetadataStore)(nil)
var _ storage.Files = (*InMemoryFileStore)(nil)
