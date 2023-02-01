package apiv1

import (
	"github.com/mredolatti/tf/codigo/fileserver/authz"
	"github.com/mredolatti/tf/codigo/fileserver/models"
	"github.com/mredolatti/tf/codigo/fileserver/storage"
)

// These wrappers (though not necessary for current version) are used to decouple external interfaces
// from internal ones. enhancing plugin compatibility via the use of adapters for each version

type FilesWrapper struct {
	w Files
}

// Read implements storage.Files
func (fw *FilesWrapper) Read(id string) ([]byte, error) {
	return fw.w.Read(id)
}

// Del implements storage.Files
func (fw *FilesWrapper) Del(id string) error {
	return fw.w.Del(id)
}

// Write implements storage.Files
func (fw *FilesWrapper) Write(id string, data []byte, force bool) error {
	return fw.w.Write(id, data, force)
}

type FilesMetaWrapper struct {
	w FilesMetadata
}

// Create implements storage.FilesMetadata
func (fmw *FilesMetaWrapper) Create(name string, notes string, patient string, typ string, whenNs int64) (models.FileMetadata, error) {
	n, err := fmw.w.Create(name, notes, patient, typ, whenNs)
	return n.(models.FileMetadata), err
}

// Get implements storage.FilesMetadata
func (fmw *FilesMetaWrapper) Get(id string) (models.FileMetadata, error) {
	c, err := fmw.w.Get(id)
	return c.(models.FileMetadata), err
}

// GetMany implements storage.FilesMetadata
func (fmw *FilesMetaWrapper) GetMany(filter *storage.Filter) (map[string]models.FileMetadata, error) {
	f := Filter(*filter)
	res, err := fmw.w.GetMany(&f)
	if err != nil {
		return nil, err
	}

	toRet := make(map[string]models.FileMetadata)
	for k, v := range res {
		toRet[k] = v
	}
	return toRet, nil
}

// Remove implements storage.FilesMetadata
func (fmw *FilesMetaWrapper) Remove(id string, whenNs int64) error {
	return fmw.w.Remove(id, whenNs)
}

// Update implements storage.FilesMetadata
func (fmw *FilesMetaWrapper) Update(id string, updated models.FileMetadata, whenNs int64) (models.FileMetadata, error) {
	res, err := fmw.w.Update(id, updated.(FileMetadata), whenNs)
	return res.(FileMetadata), err
}

type AuthorizationWrapper struct {
	w Authorization
}

// AllForObject implements authz.Authorization
func (aw *AuthorizationWrapper) AllForObject(object string) map[string]authz.Permission {
	res := aw.w.AllForObject(object)
	if len(res) == 0 {
		return nil
	}

	toRet := make(map[string]authz.Permission, len(res))
	for k, v := range toRet {
		toRet[k] = v.(authz.Permission)
	}
	return toRet
}

// AllForSubject implements authz.Authorization
func (aw *AuthorizationWrapper) AllForSubject(subject string) map[string]authz.Permission {
	res := aw.w.AllForSubject(subject)
	if len(res) == 0 {
		return nil
	}

	toRet := make(map[string]authz.Permission, len(res))
	for k, v := range toRet {
		toRet[k] = v.(authz.Permission)
	}
	return toRet
}

// Can implements authz.Authorization
func (aw *AuthorizationWrapper) Can(subject string, operation int, object string) (bool, error) {
	return aw.w.Can(subject, operation, object)
}

// Grant implements authz.Authorization
func (aw *AuthorizationWrapper) Grant(subject string, operation int, object string) error {
	return aw.w.Grant(subject, operation, object)
}

// Revoke implements authz.Authorization
func (aw *AuthorizationWrapper) Revoke(subject string, operation int, object string) error {
	return aw.w.Revoke(subject, operation, object)
}

var _ storage.Files = (*FilesWrapper)(nil)
var _ storage.FilesMetadata = (*FilesMetaWrapper)(nil)
var _ authz.Authorization = (*AuthorizationWrapper)(nil)
