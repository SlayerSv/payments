package service

import (
	"context"
	"log/slog"
	"strconv"
	"time"

	"github.com/SlayerSv/payments/internal/shared/kafka"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OutboxWorker struct {
	pool     *pgxpool.Pool
	producer *kafka.Producer
}

type OutboxEvent struct {
	ID      int64
	Topic   string
	Payload []byte
}

func NewOutboxWorker(pool *pgxpool.Pool, producer *kafka.Producer) *OutboxWorker {
	return &OutboxWorker{pool: pool, producer: producer}
}

func (w *OutboxWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			w.processEvents(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func (w *OutboxWorker) processEvents(ctx context.Context) {
	// Начинаем транзакцию Postgres
	tx, err := w.pool.Begin(ctx)
	if err != nil {
		slog.Error("Outbox processing: starting transaction", slog.String("error", err.Error()))
		return
	}
	defer tx.Rollback(ctx)

	// 1. Выбираем 10 необработанных событий.
	// SKIP LOCKED пропускает заблокированные другими процессами строчки
	query := `
		SELECT id, topic, payload 
		FROM outbox 
		WHERE status = 'PENDING' 
		ORDER BY created_at ASC 
		LIMIT 10 
		FOR UPDATE SKIP LOCKED`

	rows, err := tx.Query(ctx, query)
	if err != nil {
		slog.Error("Outbox processing: executing query", slog.String("error", err.Error()))
		return
	}
	defer rows.Close()

	var events []OutboxEvent
	var eventIDs []int64

	for rows.Next() {
		var ev OutboxEvent
		if err := rows.Scan(&ev.ID, &ev.Topic, &ev.Payload); err != nil {
			slog.Error("Outbox processing: scanning row", slog.String("error", err.Error()))
			continue
		}
		events = append(events, ev)
		eventIDs = append(eventIDs, ev.ID)
	}

	if len(events) == 0 {
		return
	}

	// 2. Шлем пачку в Кафку
	for _, ev := range events {
		// Используем ID события как ключ для обеспечения порядка (order guarantee)
		err = w.producer.Publish(ctx, ev.Topic, []byte(strconv.FormatInt(ev.ID, 10)), ev.Payload)
		if err != nil {
			slog.Error("Sending message to kafka", slog.String("error", err.Error()))
			return
		}
	}

	// 3. Если всё ушло успешно — помечаем в БД как обработанные (PROCESSED)
	updateQuery := "UPDATE outbox SET status = 'PROCESSED' WHERE id = any($1)"
	_, err = tx.Exec(ctx, updateQuery, eventIDs)
	if err != nil {
		slog.Error("Outbox processing: updating status", slog.String("error", err.Error()))
		return
	}

	// Фиксируем транзакцию БД
	err = tx.Commit(ctx)
	if err != nil {
		slog.Error("Outbox processing: commiting", slog.String("error", err.Error()))
	}
}
