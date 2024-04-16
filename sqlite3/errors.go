// Package sqlite provides error wrapping capabilities for use with the
// github.com/mattn/go-sqlite3 sqlite driver.
package sqlite3

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/mattn/go-sqlite3"
)

type Err struct {
	resource  string
	err       error
	notFound  bool
	conflict  bool
	retryable bool
	exists    bool
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

	var slErr sqlite3.Error
	if errors.As(err, &slErr) {
		if slErr.Code == sqlite3.ErrConstraint && slErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			e.exists = true
		} else {
			e.conflict = slErr.Code == sqlite3.ErrConstraint
		}
		e.retryable = slErr.Code == sqlite3.ErrLocked || slErr.Code == sqlite3.ErrBusy
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

func (c Err) Exists() bool {
	return c.exists
}
