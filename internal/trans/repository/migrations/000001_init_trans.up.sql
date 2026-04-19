-- 1. Создаем перечисления (ENUM)
CREATE TYPE account_type AS ENUM ('WALLET', 'SAVINGS');
CREATE TYPE transaction_status AS ENUM (
    'CREATED',
    'DEBIT_SUCCESS',
    'COMPLETED',
    'ROLLBACK_PENDING',
    'FAILED'
);
CREATE TYPE operation_type AS ENUM ('DEPOSIT', 'WITHDRAW', 'TRANSFER');

-- 2. Создаем таблицу transactions
CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_type operation_type NOT NULL,
    sender_id UUID,
    sender_type account_type,
    receiver_id UUID,
    receiver_type account_type,
    amount BIGINT NOT NULL CHECK (amount > 0),
    status transaction_status NOT NULL DEFAULT 'CREATED',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 3. Создаем индексы для быстрого поиска по отправителю и получателю
CREATE INDEX idx_transactions_sender_id ON transactions(sender_id);
CREATE INDEX idx_transactions_receiver_id ON transactions(receiver_id);
CREATE INDEX idx_transactions_status ON transactions(status);

-- 4. Автоматизация обновления поля updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER set_updated_at
BEFORE UPDATE ON transactions
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

ALTER TABLE transactions ADD CONSTRAINT check_transaction_participants CHECK (
    (transaction_type = 'DEPOSIT'  AND sender_id IS NULL     AND receiver_id IS NOT NULL) OR
    (transaction_type = 'WITHDRAW' AND sender_id IS NOT NULL AND receiver_id IS NULL)     OR
    (transaction_type = 'TRANSFER' AND sender_id IS NOT NULL AND receiver_id IS NOT NULL)
);