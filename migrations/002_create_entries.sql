-- 002_create_entries.sql
-- Creates the financial_entries table with soft delete support.

CREATE TYPE entry_type AS ENUM ('income', 'expense');

CREATE TABLE IF NOT EXISTS financial_entries (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id     UUID                NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title       VARCHAR(255)        NOT NULL,
    amount      NUMERIC(15, 2)      NOT NULL CHECK (amount > 0),
    type        entry_type          NOT NULL,
    category    VARCHAR(100)        NOT NULL,
    description TEXT,
    date        DATE                NOT NULL,
    created_at  TIMESTAMPTZ         NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ         NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ                              -- NULL = active, set = soft deleted
);

-- Indexes for common query patterns
CREATE INDEX IF NOT EXISTS idx_entries_user_id    ON financial_entries(user_id);
CREATE INDEX IF NOT EXISTS idx_entries_type       ON financial_entries(type);
CREATE INDEX IF NOT EXISTS idx_entries_category   ON financial_entries(category);
CREATE INDEX IF NOT EXISTS idx_entries_date       ON financial_entries(date);
CREATE INDEX IF NOT EXISTS idx_entries_deleted_at ON financial_entries(deleted_at);