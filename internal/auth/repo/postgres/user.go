package postgres

import (
	"context"

	"github.com/SlayerSv/payments/internal/auth/models"
	"github.com/SlayerSv/payments/internal/shared/errs"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
	pool *pgxpool.Pool
}

func NewUser(pool *pgxpool.Pool) *User {
	return &User{pool: pool}
}

func (r *User) Create(ctx context.Context, email string) (uuid.UUID, error) {
	query := `INSERT INTO users (email) VALUES ($1) RETURNING id`
	var id uuid.UUID
	err := r.pool.QueryRow(ctx, query, email).Scan(&id)
	return id, errs.WrapErr(err)
}

func (r *User) Get(ctx context.Context, userID uuid.UUID) (models.User, error) {
	u := models.User{}
	query := `SELECT id, email, name, password_hash, created_at, updated_at FROM users WHERE id = $1`

	err := r.pool.QueryRow(ctx, query, userID).Scan(&u.ID, &u.Email, &u.Name, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt)
	return u, errs.WrapErr(err)
}

func (r *User) GetByEmail(ctx context.Context, email string) (models.User, error) {
	u := models.User{}
	query := `SELECT id, email, name, password_hash, created_at, updated_at FROM users WHERE email = $1`

	err := r.pool.QueryRow(ctx, query, email).Scan(&u.ID, &u.Email, &u.Name, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt)
	return u, errs.WrapErr(err)
}

func (r *User) GetEmails(ctx context.Context, ids []string) (map[string]string, error) {
	idemail := map[string]string{}
	query := `SELECT id, email FROM users WHERE id = ANY($1)`

	rows, err := r.pool.Query(ctx, query, ids)
	if err != nil {
		return nil, errs.WrapErr(err)
	}
	defer rows.Close()

	var id, email string
	for rows.Next() {
		if err := rows.Scan(&id, &email); err != nil {
			return nil, errs.WrapErr(err)
		}
		idemail[id] = email
	}
	if err = rows.Err(); err != nil {
		return nil, errs.WrapErr(err)
	}
	return idemail, nil
}

func (r *User) UpdateName(ctx context.Context, userID uuid.UUID, newName string) (models.User, error) {
	query := `UPDATE users SET name = $1 WHERE id = $2 RETURNING *`
	u := models.User{}
	err := r.pool.QueryRow(ctx, query, newName, userID).Scan(&u.ID, &u.Email, &u.Name, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt)
	return u, errs.WrapErr(err)
}

func (r *User) UpdatePassword(ctx context.Context, userID uuid.UUID, newPassword string) (models.User, error) {
	query := `UPDATE users SET password_hash = $1 WHERE id = $2 RETURNING *`
	u := models.User{}
	err := r.pool.QueryRow(ctx, query, newPassword, userID).Scan(&u.ID, &u.Email, &u.Name, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt)
	return u, errs.WrapErr(err)
}

func (r *User) Delete(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	var delid uuid.UUID
	err := r.pool.QueryRow(ctx, "DELETE FROM users WHERE id = $1 RETURNING id", id).Scan(&delid)
	return delid, errs.WrapErr(err)
}
