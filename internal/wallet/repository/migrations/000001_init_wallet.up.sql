-- Состояния аутбокса
CREATE TYPE outbox_status AS ENUM ('PENDING', 'PROCESSED');

-- Таблица аккаунтов (кошельков)
CREATE TABLE accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL,
    current_balance BIGINT NOT NULL DEFAULT 0,
    version BIGINT NOT NULL DEFAULT 0, -- Для Optimistic Locking
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Таблица истории изменений (аудит)
CREATE TABLE ledger_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID NOT NULL REFERENCES accounts(id),
    transaction_id UUID NOT NULL, -- ID из сервиса транзакций
    amount BIGINT NOT NULL,       -- Может быть отрицательным (списание)
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Таблица Outbox для надежной отправки событий
CREATE TABLE outbox (
    id BIGSERIAL PRIMARY KEY,
    topic VARCHAR(100) NOT NULL,
    payload JSONB NOT NULL,
    status outbox_status NOT NULL DEFAULT 'PENDING',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Таблица для идемпотентности gRPC/HTTP запросов
CREATE TABLE idempotency_log (
    key VARCHAR(255) PRIMARY KEY,
    response_body JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Индексы
CREATE INDEX idx_accounts_owner ON accounts(owner_id);
CREATE INDEX idx_ledger_transaction ON ledger_entries(transaction_id);
CREATE INDEX idx_outbox_status ON outbox(status) WHERE status = 'PENDING';