-- ============================================================
-- payflow-simulator database migrations (v2 - custom string IDs)
-- Run this on Supabase: SQL Editor > New Query > paste & run
-- ============================================================

-- 1. Users
CREATE TABLE IF NOT EXISTS users (
    id            VARCHAR(20)         PRIMARY KEY,   -- USR-A1B2C3D4E5F6
    full_name     VARCHAR(100)        NOT NULL,
    email         VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255)        NOT NULL,
    created_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 2. Wallets (1 user = 1 wallet, balance tidak boleh negatif)
CREATE TABLE IF NOT EXISTS wallets (
    id         VARCHAR(20)    PRIMARY KEY,            -- WLT-A1B2C3D4E5F6
    user_id    VARCHAR(20) UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    balance    DECIMAL(15, 2) DEFAULT 0.00 CHECK (balance >= 0),
    currency   VARCHAR(3)     DEFAULT 'IDR',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 3. Merchants
CREATE TABLE IF NOT EXISTS merchants (
    id            VARCHAR(20)         PRIMARY KEY,   -- MRC-A1B2C3D4E5F6
    merchant_name VARCHAR(100)        NOT NULL,
    api_key       VARCHAR(255) UNIQUE NOT NULL,
    webhook_url   VARCHAR(255),
    created_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 4. Transactions (core table)
CREATE TABLE IF NOT EXISTS transactions (
    id                   VARCHAR(25)    PRIMARY KEY,  -- TXN-20260307-A1B2C3D4
    reference_id         VARCHAR(50) UNIQUE NOT NULL,
    wallet_id            VARCHAR(20) REFERENCES wallets(id),
    sender_wallet_id     VARCHAR(20) REFERENCES wallets(id),
    receiver_merchant_id VARCHAR(20) REFERENCES merchants(id),
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
    id              VARCHAR(25)    PRIMARY KEY,       -- TUP-20260307-A1B2C3D4
    wallet_id       VARCHAR(20) REFERENCES wallets(id),
    amount          DECIMAL(15, 2) NOT NULL CHECK (amount > 0),
    payment_channel VARCHAR(50) CHECK (payment_channel IN ('BANK_TRANSFER', 'VIRTUAL_ACCOUNT')),
    status          VARCHAR(20) DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'SUCCESS', 'FAILED', 'EXPIRED')),
    expired_at      TIMESTAMP,
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 6. Webhook logs
CREATE TABLE IF NOT EXISTS webhook_logs (
    id              VARCHAR(20)    PRIMARY KEY,       -- WHL-A1B2C3D4E5F6
    merchant_id     VARCHAR(20) REFERENCES merchants(id),
    transaction_id  VARCHAR(25) REFERENCES transactions(id),
    event           VARCHAR(50)    NOT NULL,
    payload         JSONB          NOT NULL,
    response_status INT,                              -- HTTP status dari merchant endpoint
    response_body   TEXT,
    retry_count     INT            DEFAULT 0,
    is_delivered    BOOLEAN        DEFAULT FALSE,
    sent_at         TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ── Indexes untuk query performance ──────────────────────────
CREATE INDEX IF NOT EXISTS idx_transactions_wallet_id    ON transactions(wallet_id);
CREATE INDEX IF NOT EXISTS idx_transactions_reference_id ON transactions(reference_id);
CREATE INDEX IF NOT EXISTS idx_transactions_created_at   ON transactions(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_wallets_user_id           ON wallets(user_id);

-- ── Seed data: dummy merchants ────────────────────────────────
INSERT INTO merchants (id, merchant_name, api_key, webhook_url) VALUES
    ('MRC-TOKOPEDIA0001', 'Tokopedia Simulator', 'tok-api-key-001', 'http://localhost:8080/webhook/receive'),
    ('MRC-GOJEK0000002', 'Gojek Simulator',     'goj-api-key-002', 'http://localhost:8080/webhook/receive'),
    ('MRC-PLN00000003', 'PLN Simulator',        'pln-api-key-003', 'http://localhost:8080/webhook/receive')
ON CONFLICT DO NOTHING;
