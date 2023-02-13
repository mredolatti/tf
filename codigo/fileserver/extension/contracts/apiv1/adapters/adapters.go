package adapters

import (
	"github.com/mredolatti/tf/codigo/fileserver/authz"
	"github.com/mredolatti/tf/codigo/fileserver/extension/contracts/apiv1"
	"github.com/mredolatti/tf/codigo/fileserver/models"
	"github.com/mredolatti/tf/codigo/fileserver/storage"
)

// These wrappers (though not necessary for current version) are used to decouple external interfaces
// from internal ones. enhancing plugin compatibility via the use of adapters for each version

type FilesWrapper struct {
	w apiv1.Files
}

func NewFilesWrapper(w apiv1.Files) *FilesWrapper {
	return &FilesWrapper{w: w}
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
	w apiv1.FilesMetadata
}

func NewFilesMetaWrapper(w apiv1.FilesMetadata) *FilesMetaWrapper {
	return &FilesMetaWrapper{w: w}
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
	f := apiv1.Filter(*filter)
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
	res, err := fmw.w.Update(id, updated.(apiv1.FileMetadata), whenNs)
	return res.(apiv1.FileMetadata), err
}

type AuthorizationWrapper struct {
	w apiv1.Authorization
}

func NewAuthWrapper(w apiv1.Authorization) *AuthorizationWrapper {
	return &AuthorizationWrapper{w: w}
}

// AllForObject implements authz.Authorization
func (aw *AuthorizationWrapper) AllForObject(object string) (map[string]authz.Permission, error) {
	res, err := aw.w.AllForObject(object)
	if err != nil {
		return nil, err
	}
	if len(res) == 0 {
		return nil, nil
	}

	toRet := make(map[string]authz.Permission, len(res))
	for k, v := range toRet {
		toRet[k] = v.(authz.Permission)
	}
	return toRet, nil
}

// AllForSubject implements authz.Authorization
func (aw *AuthorizationWrapper) AllForSubject(subject string) (map[string]authz.Permission, error) {
	res, err := aw.w.AllForSubject(subject)
	if err != nil {
		return nil, err
	}
	if len(res) == 0 {
		return nil, nil
	}

	toRet := make(map[string]authz.Permission, len(res))
	for k, v := range res {
		toRet[k] = &PermissionWrapper{p: v}
	}
	return toRet, nil
}

// Can implements authz.Authorization
func (aw *AuthorizationWrapper) Can(subject string, operation authz.Operation, object string) (bool, error) {
	return aw.w.Can(subject, apiv1.Operation(operation), object)
}

// Grant implements authz.Authorization
func (aw *AuthorizationWrapper) Grant(subject string, operation authz.Operation, object string) error {
	return aw.w.Grant(subject, apiv1.Operation(operation), object)
}

// Revoke implements authz.Authorization
func (aw *AuthorizationWrapper) Revoke(subject string, operation authz.Operation, object string) error {
	return aw.w.Revoke(subject, apiv1.Operation(operation), object)
}

type PermissionWrapper struct {
	p apiv1.Permission
}

// Can implements authz.Permission
func (p *PermissionWrapper) Can(operation authz.Operation) (bool, error) {
	return p.p.Can(apiv1.Operation(operation))
}

// Grant implements authz.Permission
func (p *PermissionWrapper) Grant(operation authz.Operation) error {
	return p.p.Grant(apiv1.Operation(operation))
}

// Revoke implements authz.Permission
func (p *PermissionWrapper) Revoke(operation authz.Operation) error {
	return p.p.Revoke(apiv1.Operation(operation))
}

var _ storage.Files = (*FilesWrapper)(nil)
var _ storage.FilesMetadata = (*FilesMetaWrapper)(nil)
var _ authz.Authorization = (*AuthorizationWrapper)(nil)
var _ authz.Permission = (*PermissionWrapper)(nil)
