package models

import "time"

// EntryType distinguishes between money coming in and going out.
type EntryType string

const (
	EntryTypeIncome  EntryType = "income"
	EntryTypeExpense EntryType = "expense"
)

// IsValid returns true if the entry type is recognised.
func (t EntryType) IsValid() bool {
	return t == EntryTypeIncome || t == EntryTypeExpense
}

// FinancialEntry represents a single financial record in the system.
// Soft deletes are used (DeletedAt) so records are never permanently lost.

type FinancialEntry struct {
	ID          string     `json:"id"`
	UserID      string     `json:"user_id"`
	Title       string     `json:"title"`
	Amount      float64    `json:"amount"`
	Type        EntryType  `json:"type"`
	Category    string     `json:"category"`
	Description string     `json:"description,omitempty"`
	Date        time.Time  `json:"date"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"` // nil = active record
}

// IsDeleted returns true if the entry has been soft-deleted.
func (e *FinancialEntry) IsDeleted() bool {
	return e.DeletedAt != nil
}

// CreateEntryInput holds validated data for creating a new entry.
type CreateEntryInput struct {
	Title       string    `json:"title"       binding:"required,min=1,max=255"`
	Amount      float64   `json:"amount"      binding:"required,gt=0"`
	Type        EntryType `json:"type"        binding:"required"`
	Category    string    `json:"category"    binding:"required,min=1,max=100"`
	Description string    `json:"description" binding:"max=1000"`
	Date        time.Time `json:"date"        binding:"required"`
}

// UpdateEntryInput holds validated data for updating an existing entry.
// All fields are optional — only non-zero values will be applied.
type UpdateEntryInput struct {
	Title       *string    `json:"title"       binding:"omitempty,min=1,max=255"`
	Amount      *float64   `json:"amount"      binding:"omitempty,gt=0"`
	Type        *EntryType `json:"type"        binding:"omitempty"`
	Category    *string    `json:"category"    binding:"omitempty,min=1,max=100"`
	Description *string    `json:"description" binding:"omitempty,max=1000"`
	Date        *time.Time `json:"date"        binding:"omitempty"`
}

// EntryFilter holds optional query parameters for listing entries.
type EntryFilter struct {
	Category string
	Type     EntryType
	DateFrom *time.Time
	DateTo   *time.Time
	Page     int
	PageSize int
}
