package fsbasic

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/mredolatti/tf/codigo/fileserver/extension/contracts/apiv1"
)

type Files struct {
	rootPath string
}

func NewFiles(path string) (*Files, error) {
	if stats, err := os.Stat(path); err != nil || !stats.IsDir() {
		return nil, fmt.Errorf("cannot use '%s' as path: %w", path, err)
	}

	return &Files{path}, nil
}

// Del implements apiv1.Files
func (f *Files) Del(id string) error {
	// truncate only, to keep "metadata" alive
	return os.Truncate(f.buildPath(id), 0)
}

// Read implements apiv1.Files
func (f *Files) Read(id string) ([]byte, error) {
	res, err := ioutil.ReadFile(f.buildPath(id))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, apiv1.ErrFileDoesNotExist
		}
		return nil, err
	}
	return res, nil
}

// Write implements apiv1.Files
func (f *Files) Write(id string, data []byte, force bool) error {
	fp := f.buildPath(id)

	exists := true
	if _, err := os.Stat(fp); errors.Is(err, os.ErrNotExist) {
		exists = false
	}

	if exists && !force {
		return apiv1.ErrFileExists
	}

	return os.WriteFile(fp, data, 0660)
}

func (f *Files) buildPath(id string) string {
	return path.Join(f.rootPath, id)
}

var _ apiv1.Files = (*Files)(nil)
