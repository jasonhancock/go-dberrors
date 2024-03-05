// Package sqlite provides error wrapping capabilities for use with the
// modernc.org/sqlite sqlite driver.
package sqlite

import (
	"database/sql"
	"fmt"

	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

type Err struct {
	resource  string
	err       error
	notFound  bool
	conflict  bool
	retryable bool
}

var conflicterCodes = map[int]struct{}{
	sqlite3.SQLITE_CONSTRAINT_CHECK:      {},
	sqlite3.SQLITE_CONSTRAINT_PRIMARYKEY: {},
	sqlite3.SQLITE_CONSTRAINT_UNIQUE:     {},
}

var retryableCodes = map[int]struct{}{
	sqlite3.SQLITE_LOCKED:     {},
	sqlite3.SQLITE_STATE_BUSY: {},
}

func NewError(resource string, err error) error {
	if err == nil {
		return nil
	}

	e := Err{
		resource: resource,
		err:      err,
		notFound: err == sql.ErrNoRows,
	}

	if slErr, ok := err.(*sqlite.Error); ok {
		code := slErr.Code()
		_, e.conflict = conflicterCodes[code]
		_, e.retryable = retryableCodes[code]
	}

	return &e
}

func (e Err) Error() string {
	if e.notFound {
		return fmt.Sprintf("%s not found", e.resource)
	}
	return e.err.Error()
}

func (c Err) NotFound() bool {
	return c.notFound
}

func (c Err) Conflict() bool {
	return c.conflict
}

func (c *Err) Unwrap() error {
	return c.err
}

func (c *Err) Retry() bool {
	return c.retryable
}
