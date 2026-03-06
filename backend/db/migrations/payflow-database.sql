-- ============================================================
-- payflow-simulator database migrations
-- Run this on Supabase: SQL Editor > New Query > paste & run
-- ============================================================

-- 1. Users
CREATE TABLE IF NOT EXISTS users (
    id            SERIAL PRIMARY KEY,
    full_name     VARCHAR(100)        NOT NULL,
    email         VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255)        NOT NULL,
    created_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 2. Wallets (1 user = 1 wallet, balance tidak boleh negatif)
CREATE TABLE IF NOT EXISTS wallets (
    id         SERIAL PRIMARY KEY,
    user_id    INT UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    balance    DECIMAL(15, 2) DEFAULT 0.00 CHECK (balance >= 0),
    currency   VARCHAR(3)     DEFAULT 'IDR',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 3. Merchants
CREATE TABLE IF NOT EXISTS merchants (
    id            SERIAL PRIMARY KEY,
    merchant_name VARCHAR(100)        NOT NULL,
    api_key       VARCHAR(255) UNIQUE NOT NULL,
    webhook_url   VARCHAR(255),
    created_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 4. Transactions (core table)
CREATE TABLE IF NOT EXISTS transactions (
    id                   SERIAL PRIMARY KEY,
    reference_id         VARCHAR(50) UNIQUE NOT NULL,
    wallet_id            INT REFERENCES wallets(id),
    sender_wallet_id     INT REFERENCES wallets(id),  -- untuk transfer antar user
    receiver_merchant_id INT REFERENCES merchants(id),
    type                 VARCHAR(20)    NOT NULL CHECK (type IN ('PAYMENT', 'TOPUP', 'TRANSFER')),
    amount               DECIMAL(15, 2) NOT NULL CHECK (amount > 0),
    fee                  DECIMAL(15, 2) DEFAULT 0.00,
    status               VARCHAR(20)    DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'SUCCESS', 'FAILED', 'EXPIRED')),
    metadata             JSONB,
    expired_at           TIMESTAMP,
    created_at           TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 5. Top-up requests (simulasi payment channel)
CREATE TABLE IF NOT EXISTS top_up_requests (
    id              SERIAL PRIMARY KEY,
    wallet_id       INT REFERENCES wallets(id),
    amount          DECIMAL(15, 2) NOT NULL CHECK (amount > 0),
    payment_channel VARCHAR(50) CHECK (payment_channel IN ('BANK_TRANSFER', 'VIRTUAL_ACCOUNT')),
    status          VARCHAR(20) DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'SUCCESS', 'FAILED', 'EXPIRED')),
    expired_at      TIMESTAMP,
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 6. Webhook logs
CREATE TABLE IF NOT EXISTS webhook_logs (
    id              SERIAL PRIMARY KEY,
    merchant_id     INT REFERENCES merchants(id),
    transaction_id  INT REFERENCES transactions(id),
    payload         JSONB          NOT NULL,
    response_status INT,
    retry_count     INT            DEFAULT 0,
    sent_at         TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ── Indexes untuk query performance ──────────────────────
CREATE INDEX IF NOT EXISTS idx_transactions_wallet_id    ON transactions(wallet_id);
CREATE INDEX IF NOT EXISTS idx_transactions_reference_id ON transactions(reference_id);
CREATE INDEX IF NOT EXISTS idx_transactions_created_at   ON transactions(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_wallets_user_id           ON wallets(user_id);

-- ── Seed data: dummy merchants ────────────────────────────
INSERT INTO merchants (merchant_name, api_key, webhook_url) VALUES
    ('Tokopedia Simulator', 'tok-api-key-001', 'https://webhook.site/tokopedia'),
    ('Gojek Simulator',     'goj-api-key-002', 'https://webhook.site/gojek'),
    ('PLN Simulator',       'pln-api-key-003', 'https://webhook.site/pln')
ON CONFLICT DO NOTHING;
