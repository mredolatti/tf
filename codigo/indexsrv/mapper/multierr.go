package mapper

import (
	"fmt"
	"strings"
	"sync"
)

type ServerView struct {
	Organization string
	Server       string
}

type MultiSyncError struct {
	errors map[ServerView]error
	mtx    sync.Mutex
}

// Error implements error
func (m *MultiSyncError) Error() string {
	m.mtx.Lock()
	var b strings.Builder
	for sv, err := range m.errors {
		b.WriteString(fmt.Sprintf("[%s/%s: %s]", sv.Organization, sv.Server, err.Error()))
	}
	m.mtx.Unlock()
	return "the following servers failed to sync: " + b.String()
}

func (m *MultiSyncError) HasErrors() bool {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	return len(m.errors) > 0
}

func (m *MultiSyncError) ForEach(fn func(org, server string, err error)) {
	m.mtx.Lock()
	for k, v := range m.errors {
		fn(k.Organization, k.Server, v)
	}
	m.mtx.Unlock()
}

func (m *MultiSyncError) Add(org, server string, err error) {
	m.mtx.Lock()
	m.errors[ServerView{org, server}] = err
	m.mtx.Unlock()
}

func (m *MultiSyncError) GetAll() map[ServerView]error {
	m.mtx.Lock()
	tr := make(map[ServerView]error, len(m.errors))
	for k, v := range m.errors {
		tr[k] = v
	}
	m.mtx.Unlock()
	return tr
}

func newMultiSyncErr() *MultiSyncError {
	return &MultiSyncError{errors: make(map[ServerView]error)}
}

var _ error = (*MultiSyncError)(nil)
