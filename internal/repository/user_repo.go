package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/M4yankkkk/findash/internal/database"
	"github.com/M4yankkkk/findash/internal/models"
)

// UserRepository handles all database operations for users.
type UserRepository struct {
	db *database.DB
}

// NewUserRepository creates a new UserRepository.
func NewUserRepository(db *database.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create inserts a new user and returns the created record.
func (r *UserRepository) Create(user *models.User) (*models.User, error) {
	query := `
		INSERT INTO users (id, name, email, password, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING id, name, email, password, role, created_at, updated_at`

	created := &models.User{}
	err := r.db.QueryRow(query,
		user.ID, user.Name, user.Email, user.Password, user.Role,
	).Scan(
		&created.ID, &created.Name, &created.Email,
		&created.Password, &created.Role,
		&created.CreatedAt, &created.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	return created, nil
}

// FindByEmail returns a user matching the given email, or nil if not found.
func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	query := `
		SELECT id, name, email, password, role, created_at, updated_at
		FROM users
		WHERE email = $1`

	user := &models.User{}
	err := r.db.QueryRow(query, email).Scan(
		&user.ID, &user.Name, &user.Email,
		&user.Password, &user.Role,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("find user by email: %w", err)
	}

	return user, nil
}

// FindByID returns a user matching the given ID, or nil if not found.
func (r *UserRepository) FindByID(id string) (*models.User, error) {
	query := `
		SELECT id, name, email, password, role, created_at, updated_at
		FROM users
		WHERE id = $1`

	user := &models.User{}
	err := r.db.QueryRow(query, id).Scan(
		&user.ID, &user.Name, &user.Email,
		&user.Password, &user.Role,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("find user by id: %w", err)
	}

	return user, nil
}

// ListAll returns all users with pagination.
func (r *UserRepository) ListAll(page, pageSize int) ([]*models.User, int, error) {
	// Get total count first
	var total int
	if err := r.db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count users: %w", err)
	}

	offset := (page - 1) * pageSize
	query := `
		SELECT id, name, email, password, role, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	rows, err := r.db.Query(query, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		u := &models.User{}
		if err := rows.Scan(
			&u.ID, &u.Name, &u.Email,
			&u.Password, &u.Role,
			&u.CreatedAt, &u.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, u)
	}

	return users, total, rows.Err()
}

// UpdateRole changes the role of a user by ID.
func (r *UserRepository) UpdateRole(userID string, role models.Role) error {
	query := `UPDATE users SET role = $1, updated_at = NOW() WHERE id = $2`
	result, err := r.db.Exec(query, role, userID)
	if err != nil {
		return fmt.Errorf("update role: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// CountAll returns the total number of users in the system.
func (r *UserRepository) CountAll() (int, error) {
	var count int
	if err := r.db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&count); err != nil {
		return 0, fmt.Errorf("count all users: %w", err)
	}
	return count, nil
}
