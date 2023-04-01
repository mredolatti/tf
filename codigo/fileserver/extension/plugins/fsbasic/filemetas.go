package fsbasic

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"

	"github.com/mredolatti/tf/codigo/fileserver/extension/contracts/apiv1"
)

type FileMetadata struct {
	id          string
	name        string
	notes       string
	patientID   string
	ftype       string
	contentID   string
	lastUpdated int64
	sizeBytes   int64
	deleted     bool
}

// ContentID implements apiv1.FileMetadata
func (m *FileMetadata) ContentID() string {
	return m.contentID
}

// Deleted implements apiv1.FileMetadata
func (m *FileMetadata) Deleted() bool {
	return m.deleted
}

// ID implements apiv1.FileMetadata
func (m *FileMetadata) ID() string {
	return m.id
}

// LastUpdated implements apiv1.FileMetadata
func (m *FileMetadata) LastUpdated() int64 {
	return m.lastUpdated
}

// Name implements apiv1.FileMetadata
func (m *FileMetadata) Name() string {
	return m.name
}

func (m *FileMetadata) SizeBytes() int64 {
	return m.sizeBytes
}

// Notes implements apiv1.FileMetadata
func (m *FileMetadata) Notes() string {
	return m.notes
}

// PatientID implements apiv1.FileMetadata
func (m *FileMetadata) PatientID() string {
	return m.patientID
}

// Type implements apiv1.FileMetadata
func (m *FileMetadata) Type() string {
	return m.ftype
}

type FilesMetadata struct {
	path string
}

func NewFilesMetadata(path string) (*FilesMetadata, error) {
	if stats, err := os.Stat(path); err != nil || !stats.IsDir() {
		return nil, fmt.Errorf("cannot use '%s' as path: %w", path, err)
	}

	return &FilesMetadata{path}, nil
}

// Create implements apiv1.FilesMetadata
func (f *FilesMetadata) Create(name string, notes string, patient string, typ string, whenNs int64) (apiv1.FileMetadata, error) {
	fname := path.Join(f.path, name)
	_, err := os.Stat(fname)
	if err != nil {
		if os.IsExist(err) {
			return nil, apiv1.ErrFileExists
		}
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("error checking if file exists when creating: %w", err)
		}
	}

	file, err := os.Create(fname)
	if err != nil {
		return nil, err
	}
	file.Close()

	return getMetaFromStats(fname)
}

// Get implements apiv1.FilesMetadata
func (f *FilesMetadata) Get(id string) (apiv1.FileMetadata, error) {
	fname := path.Join(f.path, id)
	m, err := getMetaFromStats(fname)
	if err != nil {
		return nil, fmt.Errorf("error getting file stats: %w", err)
	}

	m.ftype, err = getFileType(fname)
	if err != nil {
		return nil, fmt.Errorf("error getting file type: %w", err)
	}

	return m, nil
}

// GetMany implements apiv1.FilesMetadata
func (f *FilesMetadata) GetMany(filter *apiv1.Filter) (map[string]apiv1.FileMetadata, error) {
	metas := make(map[string]apiv1.FileMetadata)
	var names []string = filter.IDs
	if names == nil {
		// TODO(populate from list in dir)
	}

	for _, name := range names {
		fm, err := f.Get(name)
		if err != nil {
			return nil, fmt.Errorf("error fetching `%s`: %w", name, err) // consider returning partial list & error
		}

		if filter.UpdatedAfter == nil || *filter.UpdatedAfter < fm.LastUpdated() {
			metas[name] = fm
		}
	}

	return metas, nil
}

// Remove implements apiv1.FilesMetadata
func (f *FilesMetadata) Remove(id string, whenNs int64) error {
	fname := path.Join(f.path, id)
	stats, err := os.Stat(fname)
	if err != nil {
		return fmt.Errorf("error fetching stats: %w", err)
	}

	if stats.Size() != 0 {
		return fmt.Errorf("cannot delete metadata with associated contents") // TODO(mredolatti): make this a std error
	}

	if err := os.Remove(fname); err != nil {
		return fmt.Errorf("error removing file: %w", err)
	}

	return nil
}

// Update implements apiv1.FilesMetadata
func (f *FilesMetadata) Update(id string, updated apiv1.FileMetadata, whenNs int64) (apiv1.FileMetadata, error) {
	return getMetaFromStats(path.Join(f.path, id)) // do nothing
}

func getMetaFromStats(fn string) (*FileMetadata, error) {
	stats, err := os.Stat(fn)
	if err != nil {
		return nil, fmt.Errorf("error fetching stats of newly created file: %w", err)
	}

	return &FileMetadata{
		id:          stats.Name(),
		name:        stats.Name(),
		sizeBytes:   stats.Size(),
		notes:       "N/A",
		patientID:   "N/A",
		contentID:   stats.Name(),
		lastUpdated: stats.ModTime().UnixNano(),
		deleted:     false,
		ftype:       "N/A",
	}, nil
}

func getFileType(fn string) (string, error) {
	file, err := os.Open(fn)
	if err != nil {
		return "", fmt.Errorf("error opening file: %w", err)
	}

	buf := make([]byte, 512)
	_, err = file.Read(buf)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return "N/A", nil
		}
		return "", fmt.Errorf("error reading file: %w", err)
	}
	return http.DetectContentType(buf), nil
}

var _ apiv1.FileMetadata = (*FileMetadata)(nil)
var _ apiv1.FilesMetadata = (*FilesMetadata)(nil)
