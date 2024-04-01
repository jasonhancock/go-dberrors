package postgres

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/lib/pq"
	pqerror "github.com/reiver/go-pqerror"
)

// Err provides meaningful behavior to the postgres error.
type Err struct {
	resource string
	msg      string
	notFound bool
	conflict bool
	exists   bool
	err      error

	msgParts []string
}

// NewError creates a new PG error from a provided error.
func NewError(resource string, err error) error {
	if err == nil {
		return nil
	}

	pg := &Err{
		resource: resource,
		err:      err,
		msgParts: []string{
			fmt.Sprintf("resource=%q", resource),
		},
	}

	if err == sql.ErrNoRows {
		pg.notFound = true
		return pg
	}

	if e, ok := err.(*pq.Error); ok {
		switch e.Code {
		case pqerror.CodeCaseNotFound:
			pg.notFound = true
		case pqerror.CodeIntegrityConstraintViolationUniqueViolation:
			pg.exists = true
		case pqerror.CodeIntegrityConstraintViolationForeignKeyViolation:
			pg.conflict = true
		}

		if e.Severity != "" {
			pg.msgParts = append(pg.msgParts, fmt.Sprintf("severity=%q", e.Severity))
		}

		if err := e.Message; err != "" {
			pg.msgParts = append(pg.msgParts, fmt.Sprintf("err=%q", err))
		}

		if code := e.Code; code != "" {
			pg.msgParts = append(pg.msgParts, fmt.Sprintf("code=%q", code))
		}

		if constraint := e.Constraint; constraint != "" {
			pg.msgParts = append(pg.msgParts, fmt.Sprintf("constraint=%q", constraint))
		}

		if column := e.Column; column != "" {
			pg.msgParts = append(pg.msgParts, fmt.Sprintf("column=%q", column))
		}

		if position := e.Position; position != "" {
			pg.msgParts = append(pg.msgParts, fmt.Sprintf("position=%q", position))
		}

		if table := e.Table; table != "" {
			pg.msgParts = append(pg.msgParts, fmt.Sprintf("table=%q", table))
		}

		if hint := e.Hint; hint != "" {
			pg.msgParts = append(pg.msgParts, fmt.Sprintf("hint=%q", hint))
		}

		if detail := e.Detail; detail != "" {
			pg.msgParts = append(pg.msgParts, fmt.Sprintf("detail=%q", detail))
		}

		if intQuery := e.InternalQuery; intQuery != "" {
			pg.msgParts = append(pg.msgParts, fmt.Sprintf("internal_query=%q", intQuery))
		}

		if dataType := e.DataTypeName; dataType != "" {
			pg.msgParts = append(pg.msgParts, fmt.Sprintf("data_type_name=%q", dataType))
		}

		if where := e.Where; where != "" {
			pg.msgParts = append(pg.msgParts, fmt.Sprintf("where=%q", where))
		}

		if schema := e.Schema; schema != "" {
			pg.msgParts = append(pg.msgParts, fmt.Sprintf("schema=%q", schema))
		}

		return pg
	}

	pg.msg = err.Error()
	return pg
}

// Error returns the error string.
func (e *Err) Error() string {
	if e.notFound {
		return fmt.Sprintf("%s not found", e.resource)
	}

	return strings.Join(e.msgParts, " ") + " " + e.err.Error()
}

// NotFound returns whether this error refers to the behavior of a resource that was not found.
func (e *Err) NotFound() bool {
	return e.notFound
}

// Exists returns whether this error refers to the behavior of a resource that already exists in the
// datastore and a violation is thrown. This will be true for a unique key violation in the PG store,
// but can be expanded in the future.
func (e *Err) Exists() bool {
	return e.exists
}

// Conflict returns where this error refers to the behavior of a resource that conflicts. At this
// time the conflict is determiend from a foreign key violation, but can be expanded in the future.
func (e *Err) Conflict() bool {
	return e.conflict
}
