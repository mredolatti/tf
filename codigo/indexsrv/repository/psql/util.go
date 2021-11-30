package psql

import "errors"

// ErrNilDB is returned when constructing a postgresql-based repository with a nil connection
var ErrNilDB = errors.New("db cannot be nil")
