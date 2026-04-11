-- Типы состояний Саги
DO $$ BEGIN
    CREATE TYPE saga_status AS ENUM (
        'CREATED', 
        'DEBIT_PENDING', 
        'DEBIT_SUCCESS', 
        'CREDIT_PENDING', 
        'COMPLETED', 
        'ROLLBACK_PENDING', 
        'FAILED'
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- Типы счетов
DO $$ BEGIN
    CREATE TYPE account_type AS ENUM ('WALLET', 'SAVINGS');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- Основная таблица транзакций (Саг)
CREATE TABLE IF NOT EXISTS saga_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sender_id UUID NOT NULL,
    sender_type account_type NOT NULL,
    receiver_id UUID NOT NULL,
    receiver_type account_type NOT NULL,
    amount BIGINT NOT NULL CHECK (amount > 0),
    status saga_status NOT NULL DEFAULT 'CREATED',
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_saga_status ON saga_transactions(status);