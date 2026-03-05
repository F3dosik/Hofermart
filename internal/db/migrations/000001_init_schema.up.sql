CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    login CITEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TYPE order_status AS ENUM (
    'NEW',
    'PROCESSING',
    'INVALID',
    'PROCESSED'
);

CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    number TEXT NOT NULL UNIQUE,
    status order_status NOT NULL DEFAULT 'NEW',
    accrual NUMERIC(10,2),
    uploaded_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_orders_status ON orders(status)
        WHERE status IN ('NEW', 'PROCESSING');

CREATE TABLE balance (
    user_id UUID PRIMARY KEY REFERENCES users(id),
    current NUMERIC(10,2) NOT NULL DEFAULT 0 CHECK (current >= 0),
    withdrawn NUMERIC(10,2) NOT NULL DEFAULT 0 CHECK (withdrawn >=0),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE withdrawals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID  NOT NULL REFERENCES users(id),
    order_number TEXT NOT NULL,
    sum NUMERIC(10,2) NOT NULL CHECK (sum >= 0),
    processed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_withdrawals_user_id ON withdrawals(user_id);