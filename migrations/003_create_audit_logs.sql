-- 003_create_audit_logs.sql
-- Creates the audit_logs table for tracking all significant system actions.
-- Audit logs are append-only — no updates or deletes are ever performed on this table.

CREATE TYPE audit_action AS ENUM ('CREATE', 'UPDATE', 'DELETE', 'ROLE_CHANGE', 'LOGIN');

CREATE TABLE IF NOT EXISTS audit_logs (
    id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id      UUID            NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    action       audit_action    NOT NULL,
    resource     VARCHAR(100)    NOT NULL,   -- e.g. 'entry', 'user'
    resource_id  VARCHAR(255)    NOT NULL,   -- UUID of the affected record
    detail       TEXT,                       -- human-readable description
    created_at   TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

-- Index for querying logs by user or by time range
CREATE INDEX IF NOT EXISTS idx_audit_user_id   ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_action    ON audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_audit_created   ON audit_logs(created_at DESC);