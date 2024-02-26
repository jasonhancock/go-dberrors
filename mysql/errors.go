package mysql

import (
	"database/sql"
	"fmt"

	"github.com/go-sql-driver/mysql"
)

type Err struct {
	resource string
	err      error
	notFound bool
	conflict bool
	exists   bool
}

var conflicterCodes = map[uint16]struct{}{
	1062: {},
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

	if mErr, ok := err.(*mysql.MySQLError); ok {
		_, e.conflict = conflicterCodes[mErr.Number]
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

func (c Err) Exists() bool {
	return c.exists
}

func (e *Err) Unwrap() error {
	return e.err
}
