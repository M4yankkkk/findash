package models

import "time"

// AuditAction represents the type of action recorded in the audit log.
type AuditAction string

const (
	AuditActionCreate     AuditAction = "CREATE"
	AuditActionUpdate     AuditAction = "UPDATE"
	AuditActionDelete     AuditAction = "DELETE"
	AuditActionRoleChange AuditAction = "ROLE_CHANGE"
	AuditActionLogin      AuditAction = "LOGIN"
)

// AuditLog records every significant action performed in the system.
// This is especially important in finance applications for traceability.

type AuditLog struct {
	ID         string      `json:"id"`
	UserID     string      `json:"user_id"`     // who performed the action
	Action     AuditAction `json:"action"`      // what they did
	Resource   string      `json:"resource"`    // which entity type (e.g. "entry", "user")
	ResourceID string      `json:"resource_id"` // which specific record
	Detail     string      `json:"detail"`      // human-readable description
	CreatedAt  time.Time   `json:"created_at"`
}
