package errs

import (
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	NotFound      = errors.New("not found")
	AlreadyExists = errors.New("already exists")
	Internal      = errors.New("internal server error")
)

func WrapErr(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return NotFound
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505": // unique_violation
			return AlreadyExists
		}
	}
	return Internal
}
