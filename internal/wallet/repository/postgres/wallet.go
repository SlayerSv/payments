package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/SlayerSv/payments/internal/shared/errs"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/SlayerSv/payments/internal/wallet/models"
)

type WalletRepo struct {
	pool *pgxpool.Pool
}

func NewWalletRepo(pool *pgxpool.Pool) *WalletRepo {
	return &WalletRepo{pool: pool}
}

// CreateAccount — создает новый кошелек для пользователя
func (r *WalletRepo) CreateAccount(ctx context.Context, ownerID uuid.UUID) (uuid.UUID, error) {
	query := `INSERT INTO accounts (owner_id) VALUES ($1) RETURNING id`
	var id uuid.UUID
	err := r.pool.QueryRow(ctx, query, ownerID).Scan(&id)
	return id, errs.WrapErr(err)
}

// GetAccount — получает стейт кошелька (включая его version)
func (r *WalletRepo) GetAccount(ctx context.Context, id uuid.UUID) (models.Account, error) {
	query := `
		SELECT id, owner_id, current_balance, version, created_at 
		FROM accounts 
		WHERE id = $1`

	var acc models.Account
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&acc.ID, &acc.OwnerID, &acc.CurrentBalance, &acc.Version, &acc.CreatedAt,
	)
	return acc, errs.WrapErr(err)
}

// GetIdempotencyResponse — проверяет, обрабатывали ли мы уже этот запрос
func (r *WalletRepo) GetIdempotencyResponse(ctx context.Context, key string) ([]byte, bool, error) {
	query := `SELECT response_body FROM idempotency_log WHERE key = $1`

	var responseBody []byte
	err := r.pool.QueryRow(ctx, query, key).Scan(&responseBody)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, false, nil // Запрос новый
		}
		return nil, false, errs.WrapErr(err)
	}

	return responseBody, true, nil // Запрос уже был обработан
}

// UpdateBalanceAtomic — сердце финтех-логики
func (r *WalletRepo) UpdateBalanceAtomic(ctx context.Context, args models.UpdateBalanceParams) error {
	// 1. Начинаем транзакцию
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return errs.WrapErr(fmt.Errorf("failed to begin tx: %w", err))
	}
	// Обязательный defer для отката в случае паники или ошибки
	defer tx.Rollback(ctx)

	// 2. Optimistic Locking: обновляем баланс только если версия совпадает
	updateAccQuery := `
		UPDATE accounts 
		SET current_balance = current_balance + $1, 
		    version = version + 1 
		WHERE id = $2 AND version = $3`

	res, err := tx.Exec(ctx, updateAccQuery, args.AmountDelta, args.AccountID, args.ExpectedVersion)
	if err != nil {
		return errs.WrapErr(fmt.Errorf("failed to update account: %w", err))
	}

	// Если ни одна строка не обновилась, значит кто-то успел изменить баланс до нас
	// (или аккаунта не существует, но обычно мы проверяем это на этапе сервиса)
	if res.RowsAffected() == 0 {
		return errs.ConcurrentUpdate // Сервис поймает эту ошибку и сделает Retry
	}

	// 3. Пишем в Ledger (история для аудита)
	ledgerQuery := `
		INSERT INTO ledger_entries (account_id, transaction_id, amount) 
		VALUES ($1, $2, $3)`

	_, err = tx.Exec(ctx, ledgerQuery, args.AccountID, args.TransactionID, args.AmountDelta)
	if err != nil {
		return errs.WrapErr(fmt.Errorf("failed to insert ledger entry: %w", err))
	}

	// 4. Пишем в Outbox (чтобы потом фоновый воркер отправил это в Kafka)
	outboxQuery := `
		INSERT INTO outbox (topic, payload) 
		VALUES ($1, $2)`

	_, err = tx.Exec(ctx, outboxQuery, args.OutboxTopic, args.OutboxPayload)
	if err != nil {
		return errs.WrapErr(fmt.Errorf("failed to insert outbox: %w", err))
	}

	// 5. Фиксируем ключ идемпотентности, чтобы не списать деньги дважды при ретрае по сети
	if args.IdempotencyKey != "" {
		idempotencyQuery := `
			INSERT INTO idempotency_log (key, response_body)
			VALUES ($1, $2)`

		_, err = tx.Exec(ctx, idempotencyQuery, args.IdempotencyKey, args.IdempotencyResponse)
		if err != nil {
			return errs.WrapErr(fmt.Errorf("failed to insert idempotency log: %w", err))
		}
	}

	// 6. Если дошли сюда без ошибок — коммитим всё разом
	if err := tx.Commit(ctx); err != nil {
		return errs.WrapErr(fmt.Errorf("failed to commit tx: %w", err))
	}

	return nil
}
