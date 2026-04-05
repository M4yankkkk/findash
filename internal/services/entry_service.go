package services

import (
	"fmt"

	"github.com/M4yankkkk/findash/internal/models"
	"github.com/M4yankkkk/findash/internal/repository"
	"github.com/google/uuid"
)

// EntryService handles all business logic for financial entries.
type EntryService struct {
	entryRepo      *repository.EntryRepository
	auditRepo      *repository.AuditRepository
	visibilityRepo *repository.VisibilityRepository
}

// NewEntryService creates a new EntryService.
func NewEntryService(
	entryRepo *repository.EntryRepository,
	auditRepo *repository.AuditRepository,
	visibilityRepo *repository.VisibilityRepository,
) *EntryService {
	return &EntryService{entryRepo: entryRepo, auditRepo: auditRepo, visibilityRepo: visibilityRepo}
}

// CreateEntry creates a new financial entry for the requesting user.
func (s *EntryService) CreateEntry(input models.CreateEntryInput, userID string) (*models.FinancialEntry, error) {
	if !input.Type.IsValid() {
		return nil, fmt.Errorf("type must be 'income' or 'expense'")
	}

	entry := &models.FinancialEntry{
		ID:          uuid.New().String(),
		UserID:      userID,
		Title:       input.Title,
		Amount:      input.Amount,
		Type:        input.Type,
		Category:    input.Category,
		Description: input.Description,
		Date:        input.Date,
	}

	created, err := s.entryRepo.Create(entry)
	if err != nil {
		return nil, fmt.Errorf("creating entry: %w", err)
	}

	_ = s.auditRepo.Log(&models.AuditLog{
		ID:         uuid.New().String(),
		UserID:     userID,
		Action:     models.AuditActionCreate,
		Resource:   "entry",
		ResourceID: created.ID,
		Detail:     fmt.Sprintf("Created %s entry: %s (%.2f)", created.Type, created.Title, created.Amount),
	})

	return created, nil
}

// GetEntry returns a single entry by ID.
// Viewers and managers can only access their own entries.
// Admins can access any entry.
func (s *EntryService) GetEntry(entryID, requestingUserID string, role models.Role) (*models.FinancialEntry, error) {
	entry, err := s.entryRepo.FindByID(entryID)
	if err != nil {
		return nil, fmt.Errorf("finding entry: %w", err)
	}
	if entry == nil {
		return nil, nil
	}

	if role == models.RoleViewer {
		visible, err := s.visibilityRepo.IsEntryVisibleToViewer(requestingUserID, entryID)
		if err != nil {
			return nil, fmt.Errorf("checking viewer visibility: %w", err)
		}
		if !visible {
			return nil, nil
		}
		return entry, nil
	}

	// Ownership check — non-admins can only see their own entries
	if role != models.RoleAdmin && entry.UserID != requestingUserID {
		return nil, nil // Return nil to surface as 404, not 403 (avoids leaking record existence)
	}

	return entry, nil
}

// ListEntries returns a paginated, filtered list of entries.
// Admins see all entries; others see only their own.
func (s *EntryService) ListEntries(
	filter models.EntryFilter,
	requestingUserID string,
	role models.Role,
) ([]*models.FinancialEntry, int, error) {
	// Clamp pagination defaults
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 || filter.PageSize > 100 {
		filter.PageSize = 20
	}

	if role == models.RoleViewer {
		entries, total, err := s.entryRepo.ListForViewer(filter, requestingUserID)
		if err != nil {
			return nil, 0, fmt.Errorf("listing viewer entries: %w", err)
		}
		return entries, total, nil
	}

	// Admins see everything; managers are scoped to their own entries.
	scopedUserID := requestingUserID
	if role == models.RoleAdmin {
		scopedUserID = ""
	}

	entries, total, err := s.entryRepo.List(filter, scopedUserID)
	if err != nil {
		return nil, 0, fmt.Errorf("listing entries: %w", err)
	}

	return entries, total, nil
}

// UpdateEntry applies a partial update to an entry.
// Only the owner or an admin may update an entry.
func (s *EntryService) UpdateEntry(
	entryID string,
	input models.UpdateEntryInput,
	requestingUserID string,
	role models.Role,
) (*models.FinancialEntry, error) {
	existing, err := s.entryRepo.FindByID(entryID)
	if err != nil {
		return nil, fmt.Errorf("finding entry: %w", err)
	}
	if existing == nil {
		return nil, nil
	}

	// Only owner or admin can update
	if role != models.RoleAdmin && existing.UserID != requestingUserID {
		return nil, fmt.Errorf("forbidden")
	}

	// Validate type if being changed
	if input.Type != nil && !input.Type.IsValid() {
		return nil, fmt.Errorf("type must be 'income' or 'expense'")
	}

	updated, err := s.entryRepo.Update(entryID, input)
	if err != nil {
		return nil, fmt.Errorf("updating entry: %w", err)
	}

	_ = s.auditRepo.Log(&models.AuditLog{
		ID:         uuid.New().String(),
		UserID:     requestingUserID,
		Action:     models.AuditActionUpdate,
		Resource:   "entry",
		ResourceID: entryID,
		Detail:     fmt.Sprintf("Updated entry: %s", entryID),
	})

	return updated, nil
}

// DeleteEntry soft-deletes an entry.
// Only the owner or an admin may delete an entry.
func (s *EntryService) DeleteEntry(entryID, requestingUserID string, role models.Role) error {
	existing, err := s.entryRepo.FindByID(entryID)
	if err != nil {
		return fmt.Errorf("finding entry: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("not found")
	}

	// Only owner or admin can delete
	if role != models.RoleAdmin && existing.UserID != requestingUserID {
		return fmt.Errorf("forbidden")
	}

	if err := s.entryRepo.SoftDelete(entryID); err != nil {
		return fmt.Errorf("deleting entry: %w", err)
	}

	_ = s.auditRepo.Log(&models.AuditLog{
		ID:         uuid.New().String(),
		UserID:     requestingUserID,
		Action:     models.AuditActionDelete,
		Resource:   "entry",
		ResourceID: entryID,
		Detail:     fmt.Sprintf("Soft deleted entry: %s", entryID),
	})

	return nil
}
