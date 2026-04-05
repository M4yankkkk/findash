package services

import (
	"fmt"

	"github.com/M4yankkkk/findash/internal/models"
	"github.com/M4yankkkk/findash/internal/repository"
	"github.com/M4yankkkk/findash/pkg/utils"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// RegisterInput holds validated registration data.
type RegisterInput struct {
	Name     string `json:"name"     binding:"required,min=2,max=100"`
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=72"`
}

// LoginInput holds validated login data.
type LoginInput struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// AuthResponse is returned after successful register or login.
type AuthResponse struct {
	Token string              `json:"token"`
	User  models.UserResponse `json:"user"`
}

// AuthService handles user registration and authentication.
type AuthService struct {
	userRepo  *repository.UserRepository
	auditRepo *repository.AuditRepository
	jwtSecret string
	jwtExpiry int
}

// NewAuthService creates a new AuthService.
func NewAuthService(
	userRepo *repository.UserRepository,
	auditRepo *repository.AuditRepository,
	jwtSecret string,
	jwtExpiry int,
) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		auditRepo: auditRepo,
		jwtSecret: jwtSecret,
		jwtExpiry: jwtExpiry,
	}
}

// Register creates a new user account with a hashed password.
// The first user registered is automatically assigned the admin role.
func (s *AuthService) Register(input RegisterInput) (*AuthResponse, error) {
	// Check if email is already taken
	existing, err := s.userRepo.FindByEmail(input.Email)
	if err != nil {
		return nil, fmt.Errorf("checking email: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("email already registered")
	}

	// Hash password with bcrypt (cost 12 is a good balance of security vs speed)
	hashed, err := bcrypt.GenerateFromPassword([]byte(input.Password), 12)
	if err != nil {
		return nil, fmt.Errorf("hashing password: %w", err)
	}

	// First registered user becomes admin automatically
	role := models.RoleViewer
	count, err := s.userRepo.CountAll()
	if err != nil {
		return nil, fmt.Errorf("counting users: %w", err)
	}
	if count == 0 {
		role = models.RoleAdmin
	}

	user := &models.User{
		ID:       uuid.New().String(),
		Name:     input.Name,
		Email:    input.Email,
		Password: string(hashed),
		Role:     role,
		IsActive: true,
	}

	created, err := s.userRepo.Create(user)
	if err != nil {
		return nil, fmt.Errorf("creating user: %w", err)
	}

	// Issue JWT
	token, err := utils.GenerateToken(created.ID, created.Email, created.Role, s.jwtSecret, s.jwtExpiry)
	if err != nil {
		return nil, fmt.Errorf("generating token: %w", err)
	}

	// Log the registration (non-fatal)
	_ = s.auditRepo.Log(&models.AuditLog{
		ID:         uuid.New().String(),
		UserID:     created.ID,
		Action:     models.AuditActionCreate,
		Resource:   "user",
		ResourceID: created.ID,
		Detail:     fmt.Sprintf("New user registered: %s", created.Email),
	})

	return &AuthResponse{
		Token: token,
		User:  created.ToResponse(),
	}, nil
}

// Login validates credentials and returns a JWT on success.
func (s *AuthService) Login(input LoginInput) (*AuthResponse, error) {
	user, err := s.userRepo.FindByEmail(input.Email)
	if err != nil {
		return nil, fmt.Errorf("finding user: %w", err)
	}

	// Use a generic error message to avoid leaking whether an email exists
	if user == nil {
		return nil, fmt.Errorf("invalid email or password")
	}
	if !user.IsActive {
		return nil, fmt.Errorf("account is inactive")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		return nil, fmt.Errorf("invalid email or password")
	}

	token, err := utils.GenerateToken(user.ID, user.Email, user.Role, s.jwtSecret, s.jwtExpiry)
	if err != nil {
		return nil, fmt.Errorf("generating token: %w", err)
	}

	// Log the login (non-fatal)
	_ = s.auditRepo.Log(&models.AuditLog{
		ID:         uuid.New().String(),
		UserID:     user.ID,
		Action:     models.AuditActionLogin,
		Resource:   "user",
		ResourceID: user.ID,
		Detail:     fmt.Sprintf("User logged in: %s", user.Email),
	})

	return &AuthResponse{
		Token: token,
		User:  user.ToResponse(),
	}, nil
}
