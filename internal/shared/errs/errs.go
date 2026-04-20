package errs

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	NotFound           = errors.New("not found")
	AlreadyExists      = errors.New("already exists")
	Internal           = errors.New("internal server error")
	IncorrectEmail     = errors.New("incorrect email")
	InvalidCredentials = errors.New("invalid credentials")
	BadRequest         = errors.New("bad request")
	Unauthorized       = errors.New("unauthorized")
	Forbidden          = errors.New("forbidden")
	ConcurrentUpdate   = errors.New("concurrent update conflict: version mismatch")
	InsufficientFunds  = errors.New("insufficient funds")
	MaxRetriesReached  = errors.New("max retries reached due to high contention")
)

type Response struct {
	Error string `json:"error"`
}

func WrapErr(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("%w: %w", NotFound, err)
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505": // unique_violation
			return fmt.Errorf("%w: %w", AlreadyExists, err)
		}
	}
	return fmt.Errorf("%w: %w", Internal, err)
}
