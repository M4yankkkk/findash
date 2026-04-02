package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/M4yankkkk/findash/internal/database"
	"github.com/M4yankkkk/findash/internal/models"
)

// EntryRepository handles all database operations for financial entries.
type EntryRepository struct {
	db *database.DB
}

// NewEntryRepository creates a new EntryRepository.
func NewEntryRepository(db *database.DB) *EntryRepository {
	return &EntryRepository{db: db}
}

// Create inserts a new financial entry and returns the created record.
func (r *EntryRepository) Create(entry *models.FinancialEntry) (*models.FinancialEntry, error) {
	query := `
		INSERT INTO financial_entries
			(id, user_id, title, amount, type, category, description, date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
		RETURNING id, user_id, title, amount, type, category, description, date, created_at, updated_at`

	created := &models.FinancialEntry{}
	err := r.db.QueryRow(query,
		entry.ID, entry.UserID, entry.Title, entry.Amount,
		entry.Type, entry.Category, entry.Description, entry.Date,
	).Scan(
		&created.ID, &created.UserID, &created.Title, &created.Amount,
		&created.Type, &created.Category, &created.Description, &created.Date,
		&created.CreatedAt, &created.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create entry: %w", err)
	}

	return created, nil
}

// FindByID returns a single non-deleted entry by ID.
func (r *EntryRepository) FindByID(id string) (*models.FinancialEntry, error) {
	query := `
		SELECT id, user_id, title, amount, type, category, description,
		       date, created_at, updated_at, deleted_at
		FROM financial_entries
		WHERE id = $1 AND deleted_at IS NULL`

	e := &models.FinancialEntry{}
	err := r.db.QueryRow(query, id).Scan(
		&e.ID, &e.UserID, &e.Title, &e.Amount,
		&e.Type, &e.Category, &e.Description, &e.Date,
		&e.CreatedAt, &e.UpdatedAt, &e.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("find entry by id: %w", err)
	}

	return e, nil
}

// List returns a paginated, filtered list of entries.
// If userID is non-empty, only that user's entries are returned (viewer/manager).
// If userID is empty, all entries are returned (admin).
func (r *EntryRepository) List(filter models.EntryFilter, userID string) ([]*models.FinancialEntry, int, error) {
	// Build WHERE clause dynamically based on filter fields
	conditions := []string{"deleted_at IS NULL"}
	args := []interface{}{}
	argIdx := 1

	if userID != "" {
		conditions = append(conditions, fmt.Sprintf("user_id = $%d", argIdx))
		args = append(args, userID)
		argIdx++
	}

	if filter.Category != "" {
		conditions = append(conditions, fmt.Sprintf("LOWER(category) = LOWER($%d)", argIdx))
		args = append(args, filter.Category)
		argIdx++
	}

	if filter.Type != "" {
		conditions = append(conditions, fmt.Sprintf("type = $%d", argIdx))
		args = append(args, filter.Type)
		argIdx++
	}

	if filter.DateFrom != nil {
		conditions = append(conditions, fmt.Sprintf("date >= $%d", argIdx))
		args = append(args, *filter.DateFrom)
		argIdx++
	}

	if filter.DateTo != nil {
		conditions = append(conditions, fmt.Sprintf("date <= $%d", argIdx))
		args = append(args, *filter.DateTo)
		argIdx++
	}

	where := "WHERE " + strings.Join(conditions, " AND ")

	// Count total matching records for pagination metadata
	var total int
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM financial_entries %s`, where)
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count entries: %w", err)
	}

	// Fetch page
	offset := (filter.Page - 1) * filter.PageSize
	listArgs := append(args, filter.PageSize, offset)
	listQuery := fmt.Sprintf(`
		SELECT id, user_id, title, amount, type, category, description,
		       date, created_at, updated_at, deleted_at
		FROM financial_entries
		%s
		ORDER BY date DESC, created_at DESC
		LIMIT $%d OFFSET $%d`,
		where, argIdx, argIdx+1,
	)

	rows, err := r.db.Query(listQuery, listArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("list entries: %w", err)
	}
	defer rows.Close()

	var entries []*models.FinancialEntry
	for rows.Next() {
		e := &models.FinancialEntry{}
		if err := rows.Scan(
			&e.ID, &e.UserID, &e.Title, &e.Amount,
			&e.Type, &e.Category, &e.Description, &e.Date,
			&e.CreatedAt, &e.UpdatedAt, &e.DeletedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan entry: %w", err)
		}
		entries = append(entries, e)
	}

	return entries, total, rows.Err()
}

// Update applies partial updates to an entry. Only non-nil fields are changed.
func (r *EntryRepository) Update(id string, input models.UpdateEntryInput) (*models.FinancialEntry, error) {
	setClauses := []string{"updated_at = NOW()"}
	args := []interface{}{}
	argIdx := 1

	if input.Title != nil {
		setClauses = append(setClauses, fmt.Sprintf("title = $%d", argIdx))
		args = append(args, *input.Title)
		argIdx++
	}
	if input.Amount != nil {
		setClauses = append(setClauses, fmt.Sprintf("amount = $%d", argIdx))
		args = append(args, *input.Amount)
		argIdx++
	}
	if input.Type != nil {
		setClauses = append(setClauses, fmt.Sprintf("type = $%d", argIdx))
		args = append(args, *input.Type)
		argIdx++
	}
	if input.Category != nil {
		setClauses = append(setClauses, fmt.Sprintf("category = $%d", argIdx))
		args = append(args, *input.Category)
		argIdx++
	}
	if input.Description != nil {
		setClauses = append(setClauses, fmt.Sprintf("description = $%d", argIdx))
		args = append(args, *input.Description)
		argIdx++
	}
	if input.Date != nil {
		setClauses = append(setClauses, fmt.Sprintf("date = $%d", argIdx))
		args = append(args, *input.Date)
		argIdx++
	}

	args = append(args, id)
	query := fmt.Sprintf(`
		UPDATE financial_entries
		SET %s
		WHERE id = $%d AND deleted_at IS NULL
		RETURNING id, user_id, title, amount, type, category, description,
		          date, created_at, updated_at, deleted_at`,
		strings.Join(setClauses, ", "), argIdx,
	)

	e := &models.FinancialEntry{}
	err := r.db.QueryRow(query, args...).Scan(
		&e.ID, &e.UserID, &e.Title, &e.Amount,
		&e.Type, &e.Category, &e.Description, &e.Date,
		&e.CreatedAt, &e.UpdatedAt, &e.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("update entry: %w", err)
	}

	return e, nil
}

// SoftDelete marks an entry as deleted without removing the row.
// Financial records are never hard-deleted for audit integrity.
func (r *EntryRepository) SoftDelete(id string) error {
	now := time.Now()
	query := `
		UPDATE financial_entries
		SET deleted_at = $1, updated_at = NOW()
		WHERE id = $2 AND deleted_at IS NULL`

	result, err := r.db.Exec(query, now, id)
	if err != nil {
		return fmt.Errorf("soft delete entry: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("entry not found or already deleted")
	}

	return nil
}
