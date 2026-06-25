package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/kick/sigma-connected/internal/domain"
)

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `INSERT INTO users (id, tenant_id, name, email, password_hash, role, created_at, updated_at)
						VALUES (:id, :tenant_id, :name, :email, :password_hash, :role, :created_at, :updated_at)
						RETURNING id, created_at, updated_at`
	rows, err := r.db.NamedQueryContext(ctx, query, user)
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	defer rows.Close()
	if rows.Next() {
		return rows.Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	}
	return nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	err := r.db.GetContext(ctx, &user,
		"SELECT id, tenant_id, name, email, password_hash, role, created_at, updated_at FROM users WHERE email = $1",
		email)
	if err != nil {
		return nil, fmt.Errorf("find user by email: %w", err)
	}
	return &user, nil
}

func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	var user domain.User
	err := r.db.GetContext(ctx, &user,
		"SELECT id, tenant_id, name, email, password_hash, role, created_at, updated_at FROM users WHERE id = $1", id)
	if err != nil {
		return nil, fmt.Errorf("find user by ID: %w", err)
	}
	return &user, nil
}
