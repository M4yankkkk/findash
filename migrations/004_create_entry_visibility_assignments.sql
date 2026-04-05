-- 004_create_entry_visibility_assignments.sql
-- Maps viewer users to the entries they are allowed to see.

CREATE TABLE IF NOT EXISTS entry_visibility_assignments (
    viewer_id   UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    entry_id    UUID        NOT NULL REFERENCES financial_entries(id) ON DELETE CASCADE,
    assigned_by UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (viewer_id, entry_id)
);

CREATE INDEX IF NOT EXISTS idx_visibility_viewer_id ON entry_visibility_assignments(viewer_id);
CREATE INDEX IF NOT EXISTS idx_visibility_entry_id ON entry_visibility_assignments(entry_id);
