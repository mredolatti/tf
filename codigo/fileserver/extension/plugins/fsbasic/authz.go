package fsbasic

import (
	"encoding/binary"
	"errors"
	"fmt"
	"strings"

	"github.com/dgraph-io/badger/v3"
	"github.com/mredolatti/tf/codigo/fileserver/extension/contracts/apiv1"
)

type Permission uint32

// Can implements apiv1.Permission
func (p *Permission) Can(operation apiv1.Operation) (bool, error) {
	return (*p & Permission(operation)) != 0, nil
}

// Grant implements apiv1.Permission
func (p *Permission) Grant(operation apiv1.Operation) error {
	*p |= Permission(operation)
	return nil
}

// Revoke implements apiv1.Permission
func (p *Permission) Revoke(operation apiv1.Operation) error {
	*p &= Permission(operation ^ 0xFFFFFFFF)
	return nil
}

type Authorization struct {
	db *badger.DB
}

func NewAuthz(dbdir string) (*Authorization, error) {
	db, err := badger.Open(badger.DefaultOptions(dbdir))
	if err != nil {
		return nil, fmt.Errorf("error opening db: %w", err)
	}
	return &Authorization{db: db}, nil
}

// AllForObject implements apiv1.Authorization
func (a *Authorization) AllForObject(object string) (map[string]apiv1.Permission, error) {
	panic("unimplemented")
}

// AllForSubject implements apiv1.Authorization
func (a *Authorization) AllForSubject(subject string) (map[string]apiv1.Permission, error) {
	prefix := []byte(subject + "::")
	m := make(map[string]apiv1.Permission)
	err := a.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			err := it.Item().Value(func(v []byte) error {
				m[objFromKey(string(it.Item().Key()))] = ref(decodePermission(v))
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return m, nil
}

// Can implements apiv1.Authorization
func (a *Authorization) Can(subject string, operation apiv1.Operation, object string) (bool, error) {
	var can bool
	key := []byte(makeKey(subject, object))
	err := a.db.View(func(t *badger.Txn) error {
		item, err := t.Get(key)
		if errors.Is(err, badger.ErrKeyNotFound) {
			return nil // no entry for subject::object -> false
		}
		if err != nil {
			return err
		}

		err = item.Value(func(val []byte) error {
			p := decodePermission(val)
			var err error
			can, err = p.Can(operation)
			return err
		})
		return err
	})

	if err != nil {
		return false, fmt.Errorf("error querying db: %w", err)
	}
	return can, nil
}

// Grant implements apiv1.Authorization
func (a *Authorization) Grant(subject string, operation apiv1.Operation, object string) error {
	key := []byte(makeKey(subject, object))
	err := a.db.Update(func(t *badger.Txn) error {
		item, err := t.Get(key)
		if err != nil && !errors.Is(err, badger.ErrKeyNotFound) {
			return err
		}

		var p Permission
		if item != nil {
			if err = item.Value(func(val []byte) error {
				p = decodePermission(val)
				return nil
			}); err != nil {
				return err
			}
		}

		p.Grant(operation)
		return t.Set(key, encodePermission(p))
	})
	if err != nil {
		return fmt.Errorf("error performing update on db: %w", err)
	}
	return nil
}

// Revoke implements apiv1.Authorization
func (a *Authorization) Revoke(subject string, operation apiv1.Operation, object string) error {
	key := []byte(makeKey(subject, object))
	err := a.db.Update(func(t *badger.Txn) error {
		var p Permission
		item, err := t.Get(key)
		if err != nil && !errors.Is(err, badger.ErrKeyNotFound) {
			return err
		}

		if item == nil {
			return nil // key doesn't exist, nothing to do
		}

		if err = item.Value(func(val []byte) error {
			p = decodePermission(val)
			return nil
		}); err != nil {
			return err
		}

		p.Revoke(operation)
		return t.Set(key, encodePermission(p))
	})
	if err != nil {
		return fmt.Errorf("error performing update on db: %w", err)
	}
	return nil
}

func makeKey(subject string, object string) string {
	return subject + "::" + object
}

func objFromKey(key string) string {
	return strings.Split(string(key), "::")[1]
}

func decodePermission(raw []byte) Permission {
	return Permission(binary.LittleEndian.Uint32(raw))
}

func encodePermission(p Permission) []byte {
	return binary.LittleEndian.AppendUint32(nil, uint32(p))
}

func ref[T any](t T) *T {
	return &t
}

var _ apiv1.Authorization = (*Authorization)(nil)
var _ apiv1.Permission = (*Permission)(nil)
