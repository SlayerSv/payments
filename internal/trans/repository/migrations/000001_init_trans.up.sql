-- 1. Создаем перечисления (ENUM)
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
    operation_type operation_type NOT NULL,
    donor_wallet_id UUID,
    receiver_wallet_id UUID,
    amount BIGINT NOT NULL CHECK (amount > 0),
    status transaction_status NOT NULL DEFAULT 'CREATED',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 3. Создаем индексы для быстрого поиска по отправителю и получателю
CREATE INDEX idx_transactions_donor_wallet_id ON transactions(donor_wallet_id);
CREATE INDEX idx_transactions_receiver_wallet_id ON transactions(receiver_wallet_id);
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
    (operation_type = 'DEPOSIT'  AND donor_wallet_id IS NULL     AND receiver_wallet_id IS NOT NULL) OR
    (operation_type = 'WITHDRAW' AND donor_wallet_id IS NOT NULL AND receiver_wallet_id IS NULL)     OR
    (operation_type = 'TRANSFER' AND donor_wallet_id IS NOT NULL AND receiver_wallet_id IS NOT NULL)
);