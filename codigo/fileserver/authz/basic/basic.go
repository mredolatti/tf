package basic

import (
	"sync"

	"github.com/mredolatti/tf/codigo/fileserver/authz"
)

// InMemoryAuthz in-memory implementation of Authorization
type InMemoryAuthz struct {
	permissions permissionMap
	mutex       sync.Mutex
}

// NewInMemoryAuthz creates a new in-memory authorization manager
func NewInMemoryAuthz() *InMemoryAuthz {
	return &InMemoryAuthz{permissions: make(permissionMap)}
}

// Can returns whether a subject an perform a certain operation on an object
func (i *InMemoryAuthz) Can(subject string, operation int, object string) (bool, error) {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	p := i.permissions.forSubjectAndObject(subject, object)
	if p == nil {
		return false, nil // TODO: we shold return whether the user or the object wasn't found
	}

	return p.Can(operation)
}

// AllForSubject returns all the objects that the provided subject has access to and it's permissions
func (i *InMemoryAuthz) AllForSubject(subject string) map[string]authz.Permission {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	res := i.permissions.forSubject(subject)
	delete(res, authz.AnyObject)
	return res
}

// AllForObject returns all the subjects that can perform at least one operation on the provided object
func (i *InMemoryAuthz) AllForObject(object string) map[string]authz.Permission {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	return i.permissions.forObject(object)
}

// Grant access to subject on object to perform operation
func (i *InMemoryAuthz) Grant(subject string, operation int, object string) error {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	i.permissions.grant(subject, operation, object)
	return nil
}

// Revoke access to subject on object to perform operation
func (i *InMemoryAuthz) Revoke(subject string, operation int, object string) error {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	i.permissions.revoke(subject, operation, object)
	return nil
}

type permissionMap map[string]map[string]authz.Permission

func (p *permissionMap) forSubjectAndObject(subject string, object string) *authz.Permission {
	if *p == nil {
		return nil
	}

	forSubject, ok := (*p)[subject]
	if !ok {
		return nil
	}

	forObject, ok := forSubject[object]
	if !ok {
		return nil
	}

	return &forObject
}

func (p *permissionMap) forSubject(subject string) map[string]authz.Permission {
	if *p == nil {
		return nil
	}

	forSubject, ok := (*p)[subject]
	if !ok {
		return nil
	}

	return forSubject
}

func (p *permissionMap) forObject(object string) map[string]authz.Permission {
	if *p == nil {
		return nil
	}

	tmp := make(map[string]authz.Permission)
	for subject, forObject := range *p {
		for obj, permission := range forObject {
			if obj == object {
				tmp[subject] = permission
			}
		}
	}

	return tmp
}

func (p *permissionMap) grant(subject string, operation int, object string) error {
	if *p == nil {
		*p = make(permissionMap)
	}

	forUser, ok := (*p)[subject]
	if !ok {
		forUser = make(map[string]authz.Permission)
	}

	forObject := forUser[object] // create if it doesn't exist
	forObject.Grant(operation)

	forUser[object] = forObject
	(*p)[subject] = forUser

	return nil
}

func (p *permissionMap) revoke(subject string, operation int, object string) error {
	if *p == nil {
		return nil // Nothing to be done
	}

	forUser, ok := (*p)[subject]
	if !ok {
		return nil // Nothing to be done
	}

	forObject, ok := forUser[object] // create if it doesn't exist
	if !ok {
		return nil // Nothing to be done
	}

	forObject.Revoke(operation)
	forUser[object] = forObject
	(*p)[subject] = forUser

	return nil
}

var _ authz.Authorization = (*InMemoryAuthz)(nil)
