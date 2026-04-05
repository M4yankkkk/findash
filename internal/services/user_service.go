package services

import (
	"fmt"

	"github.com/M4yankkkk/findash/internal/models"
	"github.com/M4yankkkk/findash/internal/repository"
	"github.com/google/uuid"
)

// UpdateRoleInput holds validated data for a role change request.
type UpdateRoleInput struct {
	Role models.Role `json:"role" binding:"required"`
}

// UpdateStatusInput holds validated data for active/inactive status changes.
type UpdateStatusInput struct {
	IsActive bool `json:"is_active"`
}

// UpdateViewerVisibilityInput holds payload for replacing viewer-visible entries.
type UpdateViewerVisibilityInput struct {
	EntryIDs []string `json:"entry_ids" binding:"required"`
}

// UserService handles user management operations.
type UserService struct {
	userRepo       *repository.UserRepository
	auditRepo      *repository.AuditRepository
	entryRepo      *repository.EntryRepository
	visibilityRepo *repository.VisibilityRepository
}

// NewUserService creates a new UserService.
func NewUserService(
	userRepo *repository.UserRepository,
	auditRepo *repository.AuditRepository,
	entryRepo *repository.EntryRepository,
	visibilityRepo *repository.VisibilityRepository,
) *UserService {
	return &UserService{userRepo: userRepo, auditRepo: auditRepo, entryRepo: entryRepo, visibilityRepo: visibilityRepo}
}

// ListUsers returns a paginated list of all users.
func (s *UserService) ListUsers(page, pageSize int) ([]models.UserResponse, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	users, total, err := s.userRepo.ListAll(page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("listing users: %w", err)
	}

	responses := make([]models.UserResponse, len(users))
	for i, u := range users {
		responses[i] = u.ToResponse()
	}

	return responses, total, nil
}

// UpdateRole changes a user's role.
// An admin cannot demote themselves — this prevents accidental lockout.
func (s *UserService) UpdateRole(targetUserID, requestingUserID string, newRole models.Role) error {
	if !newRole.IsValid() {
		return fmt.Errorf("invalid role: must be one of admin, manager, viewer")
	}

	// Prevent self-demotion
	if targetUserID == requestingUserID {
		return fmt.Errorf("you cannot change your own role")
	}

	target, err := s.userRepo.FindByID(targetUserID)
	if err != nil {
		return fmt.Errorf("finding user: %w", err)
	}
	if target == nil {
		return fmt.Errorf("user not found")
	}

	if err := s.userRepo.UpdateRole(targetUserID, newRole); err != nil {
		return fmt.Errorf("updating role: %w", err)
	}

	// Audit the role change
	_ = s.auditRepo.Log(&models.AuditLog{
		ID:         uuid.New().String(),
		UserID:     requestingUserID,
		Action:     models.AuditActionRoleChange,
		Resource:   "user",
		ResourceID: targetUserID,
		Detail: fmt.Sprintf(
			"Role changed from %s to %s for user %s",
			target.Role, newRole, target.Email,
		),
	})

	return nil
}

// UpdateStatus changes a user's active/inactive status.
// An admin cannot deactivate themselves to prevent lockout.
func (s *UserService) UpdateStatus(targetUserID, requestingUserID string, isActive bool) error {
	if targetUserID == requestingUserID && !isActive {
		return fmt.Errorf("you cannot deactivate your own account")
	}

	target, err := s.userRepo.FindByID(targetUserID)
	if err != nil {
		return fmt.Errorf("finding user: %w", err)
	}
	if target == nil {
		return fmt.Errorf("user not found")
	}

	if err := s.userRepo.UpdateStatus(targetUserID, isActive); err != nil {
		return fmt.Errorf("updating status: %w", err)
	}

	actionWord := "deactivated"
	if isActive {
		actionWord = "activated"
	}

	_ = s.auditRepo.Log(&models.AuditLog{
		ID:         uuid.New().String(),
		UserID:     requestingUserID,
		Action:     models.AuditActionUpdate,
		Resource:   "user",
		ResourceID: targetUserID,
		Detail:     fmt.Sprintf("User %s was %s", target.Email, actionWord),
	})

	return nil
}

// GetUser returns a single user by ID.
func (s *UserService) GetUser(userID string) (*models.UserResponse, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, fmt.Errorf("finding user: %w", err)
	}
	if user == nil {
		return nil, nil
	}

	resp := user.ToResponse()
	return &resp, nil
}

// ListViewers returns paginated viewer users for visibility assignment workflows.
func (s *UserService) ListViewers(page, pageSize int) ([]models.UserResponse, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	viewers, total, err := s.userRepo.ListByRole(models.RoleViewer, page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("listing viewers: %w", err)
	}

	responses := make([]models.UserResponse, len(viewers))
	for i, u := range viewers {
		responses[i] = u.ToResponse()
	}

	return responses, total, nil
}

// GetViewerVisibility returns entry IDs visible to a viewer user.
func (s *UserService) GetViewerVisibility(viewerID string) ([]string, error) {
	viewer, err := s.userRepo.FindByID(viewerID)
	if err != nil {
		return nil, fmt.Errorf("finding viewer: %w", err)
	}
	if viewer == nil {
		return nil, fmt.Errorf("viewer not found")
	}
	if viewer.Role != models.RoleViewer {
		return nil, fmt.Errorf("target user is not a viewer")
	}

	ids, err := s.visibilityRepo.ListViewerEntryIDs(viewerID)
	if err != nil {
		return nil, fmt.Errorf("fetching viewer visibility: %w", err)
	}

	return ids, nil
}

// ReplaceViewerVisibility replaces all entries visible to a viewer.
func (s *UserService) ReplaceViewerVisibility(viewerID string, entryIDs []string, actorID string, actorRole models.Role) error {
	viewer, err := s.userRepo.FindByID(viewerID)
	if err != nil {
		return fmt.Errorf("finding viewer: %w", err)
	}
	if viewer == nil {
		return fmt.Errorf("viewer not found")
	}
	if viewer.Role != models.RoleViewer {
		return fmt.Errorf("target user is not a viewer")
	}

	uniqueIDs := make([]string, 0, len(entryIDs))
	seen := make(map[string]struct{}, len(entryIDs))
	for _, entryID := range entryIDs {
		if entryID == "" {
			continue
		}
		if _, ok := seen[entryID]; ok {
			continue
		}
		seen[entryID] = struct{}{}
		uniqueIDs = append(uniqueIDs, entryID)
	}

	for _, entryID := range uniqueIDs {
		entry, err := s.entryRepo.FindByID(entryID)
		if err != nil {
			return fmt.Errorf("finding entry: %w", err)
		}
		if entry == nil {
			return fmt.Errorf("entry not found: %s", entryID)
		}

		if actorRole == models.RoleManager {
			owned, err := s.entryRepo.IsOwnedBy(entryID, actorID)
			if err != nil {
				return fmt.Errorf("checking entry ownership: %w", err)
			}
			if !owned {
				return fmt.Errorf("forbidden entry assignment")
			}
		}
	}

	if err := s.visibilityRepo.ReplaceViewerEntries(viewerID, uniqueIDs, actorID); err != nil {
		return fmt.Errorf("replacing viewer visibility: %w", err)
	}

	_ = s.auditRepo.Log(&models.AuditLog{
		ID:         uuid.New().String(),
		UserID:     actorID,
		Action:     models.AuditActionUpdate,
		Resource:   "viewer_visibility",
		ResourceID: viewerID,
		Detail:     fmt.Sprintf("Updated viewer visibility for %s with %d entries", viewer.Email, len(uniqueIDs)),
	})

	return nil
}
