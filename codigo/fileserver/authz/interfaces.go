package authz

import (
	"errors"
)

// AnyObject is used as an object for permissions that don't target a specific one (ie: Create)
const AnyObject = "__GLOBAL__"

// EveryOne is used as an object for permissions that don't target a specific user, but affects everyone
const EveryOne = "__EVERYONE__"

// Permission types bitmask constants
const (
	Read   = (1 << 0)
	Write  = (1 << 1)
	Create = (1 << 2)
	Admin  = (1 << 31)
)

// IsValidOperation returns true is the provided operation is valid
func IsValidOperation(operation int) bool {
	switch operation {
	case Read, Write, Create, Admin:
		return true
	}

	return false
}

// Public errors
var (
	ErrNoSuchUser       = errors.New("no such user")
	ErrNosuchObject     = errors.New("no such object")
	ErrNoSuchPermission = errors.New("no such permission type")
)

// Authorization defines the interface of an authorization handling component
type Authorization interface {
	Can(subject string, operation int, object string) (bool, error)
	Grant(subject string, operation int, object string) error
	Revoke(subject string, operation int, object string) error
	AllForSubject(subject string) map[string]Permission
	AllForObject(object string) map[string]Permission
}

// Permission contains the set of operations that a subject can perform on a certain object
type Permission int

// Can returns true if operation is allowed
func (p Permission) Can(operation int) (bool, error) {
	if !IsValidOperation(operation) {
		return false, ErrNoSuchPermission
	}
	return (p & Permission(operation)) != 0, nil
}

// Grant enables operation
func (p *Permission) Grant(operation int) error {
	if !IsValidOperation(operation) {
		return ErrNoSuchPermission
	}

	(*p) |= Permission(operation)
	return nil
}

// Revoke disables operation
func (p *Permission) Revoke(operation int) error {
	if !IsValidOperation(operation) {
		return ErrNoSuchPermission
	}
	(*p) &= (Permission(operation ^ 0xFFFFFFFF))
	return nil
}
