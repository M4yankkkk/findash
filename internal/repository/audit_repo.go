package repository

import (
	"fmt"

	"github.com/M4yankkkk/findash/internal/database"
	"github.com/M4yankkkk/findash/internal/models"
)

// AuditRepository handles all database operations for audit logs.
type AuditRepository struct {
	db *database.DB
}

// NewAuditRepository creates a new AuditRepository.
func NewAuditRepository(db *database.DB) *AuditRepository {
	return &AuditRepository{db: db}
}

// Log inserts a new audit log entry. This is fire-and-forget in most cases —
// a failure here should not block the main operation.
func (r *AuditRepository) Log(entry *models.AuditLog) error {
	query := `
		INSERT INTO audit_logs (id, user_id, action, resource, resource_id, detail, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())`

	_, err := r.db.Exec(query,
		entry.ID,
		entry.UserID,
		entry.Action,
		entry.Resource,
		entry.ResourceID,
		entry.Detail,
	)
	if err != nil {
		return fmt.Errorf("insert audit log: %w", err)
	}

	return nil
}
