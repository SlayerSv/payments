-- Таблица кошельков
CREATE TABLE IF NOT EXISTS accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL, -- Ссылка на UserID из Auth сервиса
    balance BIGINT NOT NULL DEFAULT 0,
    version BIGINT NOT NULL DEFAULT 0, -- Для Optimistic Locking
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Бухгалтерская книга (Ledger) — история всех движений
CREATE TABLE IF NOT EXISTS ledger_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID NOT NULL REFERENCES accounts(id),
    transaction_id UUID NOT NULL, -- ID из сервиса транзакций (Correlation ID)
    amount BIGINT NOT NULL,       -- Положительное (credit) или отрицательное (debit)
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Таблица для обеспечения идемпотентности gRPC вызовов
CREATE TABLE IF NOT EXISTS idempotency_log (
    key VARCHAR(255) PRIMARY KEY, -- Обычно это TransactionID от Саги
    response_body JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Таблица Outbox для гарантированной отправки событий в Kafka
DO $$ BEGIN
    CREATE TYPE outbox_status AS ENUM ('PENDING', 'PROCESSED', 'FAILED');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

CREATE TABLE IF NOT EXISTS outbox (
    id BIGSERIAL PRIMARY KEY,
    topic VARCHAR(100) NOT NULL,
    payload JSONB NOT NULL,
    status outbox_status DEFAULT 'PENDING',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_outbox_status ON outbox(status);