package repository

import (
	"fmt"

	"github.com/M4yankkkk/findash/internal/database"
)

// VisibilityRepository handles viewer-entry visibility assignments.
type VisibilityRepository struct {
	db *database.DB
}

// NewVisibilityRepository creates a new VisibilityRepository.
func NewVisibilityRepository(db *database.DB) *VisibilityRepository {
	return &VisibilityRepository{db: db}
}

// ReplaceViewerEntries replaces all visible entries for a viewer.
func (r *VisibilityRepository) ReplaceViewerEntries(viewerID string, entryIDs []string, assignedBy string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("begin visibility transaction: %w", err)
	}

	if _, err := tx.Exec(`DELETE FROM entry_visibility_assignments WHERE viewer_id = $1`, viewerID); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("clear viewer visibility: %w", err)
	}

	for _, entryID := range entryIDs {
		if _, err := tx.Exec(
			`INSERT INTO entry_visibility_assignments (viewer_id, entry_id, assigned_by) VALUES ($1, $2, $3)`,
			viewerID, entryID, assignedBy,
		); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("insert viewer visibility: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit visibility transaction: %w", err)
	}

	return nil
}

// ListViewerEntryIDs returns all entry IDs assigned to the viewer.
func (r *VisibilityRepository) ListViewerEntryIDs(viewerID string) ([]string, error) {
	rows, err := r.db.Query(
		`SELECT entry_id::text FROM entry_visibility_assignments WHERE viewer_id = $1 ORDER BY created_at DESC`,
		viewerID,
	)
	if err != nil {
		return nil, fmt.Errorf("list viewer entry ids: %w", err)
	}
	defer rows.Close()

	ids := make([]string, 0)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan viewer entry id: %w", err)
		}
		ids = append(ids, id)
	}

	return ids, rows.Err()
}

// IsEntryVisibleToViewer returns true if the entry is assigned to the viewer.
func (r *VisibilityRepository) IsEntryVisibleToViewer(viewerID, entryID string) (bool, error) {
	const query = `
		SELECT EXISTS (
			SELECT 1
			FROM entry_visibility_assignments
			WHERE viewer_id = $1 AND entry_id = $2
		)`

	var exists bool
	if err := r.db.QueryRow(query, viewerID, entryID).Scan(&exists); err != nil {
		return false, fmt.Errorf("check viewer entry visibility: %w", err)
	}

	return exists, nil
}
