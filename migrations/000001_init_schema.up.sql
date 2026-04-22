CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS vouchers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code VARCHAR(255) UNIQUE NOT NULL,
    discount_type VARCHAR(50) NOT NULL, 
    discount_value NUMERIC(15, 2) NOT NULL,
    max_uses INT NOT NULL,
    total_claims INT DEFAULT 0 NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    min_transaction_amount NUMERIC(15, 2) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS claims (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    voucher_id UUID REFERENCES vouchers(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    status VARCHAR(50) NOT NULL, 
    transaction_amount NUMERIC(15, 2),
    discount_applied NUMERIC(15, 2),
    final_amount NUMERIC(15, 2),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(voucher_id, user_id)
);

CREATE INDEX idx_vouchers_code ON vouchers(code);
CREATE INDEX idx_claims_status_voucher ON claims(status, voucher_id);
CREATE INDEX idx_claims_user_voucher ON claims(user_id, voucher_id);
