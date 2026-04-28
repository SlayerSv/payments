package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/SlayerSv/payments/internal/shared/errs"
	"github.com/SlayerSv/payments/internal/trans/models"
)

type Transaction struct {
	pool *pgxpool.Pool
}

func NewTransaction(pool *pgxpool.Pool) *Transaction {
	return &Transaction{pool: pool}
}

// Create — Создание новой транзакции
func (r *Transaction) Create(ctx context.Context, tx models.Transaction) (uuid.UUID, error) {
	query := `
		INSERT INTO transactions (
			sender_id, sender_type, 
			receiver_id, receiver_type, amount
		) VALUES ($1, $2, $3, $4, $5) 
		RETURNING id`

	var id uuid.UUID
	err := r.pool.QueryRow(ctx, query,
		tx.SenderID, tx.SenderType,
		tx.ReceiverID, tx.ReceiverType, tx.Amount,
	).Scan(&id)
	return id, errs.WrapErr(err)
}

// GetByID — Получение одной транзакции
func (r *Transaction) GetByID(ctx context.Context, id uuid.UUID) (models.Transaction, error) {
	query := `
		SELECT id, sender_id, sender_type, receiver_id, receiver_type, 
		       amount, status, created_at, updated_at 
		FROM transactions 
		WHERE id = $1`

	var tx models.Transaction
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&tx.ID, &tx.SenderID, &tx.SenderType, &tx.ReceiverID, &tx.ReceiverType,
		&tx.Amount, &tx.Status, &tx.CreatedAt, &tx.UpdatedAt,
	)
	return tx, errs.WrapErr(err)
}

// UpdateStatus — Обновление статуса и сообщения об ошибке
// updated_at обновится в БД автоматически благодаря триггеру
func (r *Transaction) UpdateStatus(ctx context.Context, id uuid.UUID, newStatus models.TransactionStatus) error {
	query := `
		UPDATE transactions 
		SET status = $1
		WHERE id = $2`

	res, err := r.pool.Exec(ctx, query, newStatus, id)
	if err != nil {
		return errs.WrapErr(err)
	}
	if res.RowsAffected() == 0 {
		return errs.WrapErr(pgx.ErrNoRows)
	}
	return nil
}

// GetAccHistory — Получение истории транзакций конкретного аккаунта (и как отправителя, и как получателя)
func (r *Transaction) GetTransactionHistory(ctx context.Context, accountID uuid.UUID) ([]models.Transaction, error) {
	query := `
		SELECT id, sender_id, sender_type, receiver_id, receiver_type, 
		       amount, status, created_at, updated_at 
		FROM transactions 
		WHERE sender_id = $1 OR receiver_id = $1
		ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, query, accountID)
	if err != nil {
		return nil, errs.WrapErr(err)
	}
	defer rows.Close()

	var history []models.Transaction
	for rows.Next() {
		var tx models.Transaction
		err := rows.Scan(
			&tx.ID, &tx.SenderID, &tx.SenderType, &tx.ReceiverID, &tx.ReceiverType,
			&tx.Amount, &tx.Status, &tx.CreatedAt, &tx.UpdatedAt,
		)
		if err != nil {
			return nil, errs.WrapErr(err)
		}
		history = append(history, tx)
	}

	if err = rows.Err(); err != nil {
		return nil, errs.WrapErr(err)
	}
	return history, nil
}
