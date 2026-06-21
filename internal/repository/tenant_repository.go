package repository

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/kick/sigma-connected/internal/domain"
)

type TenantRepository struct {
	db *sqlx.DB
}

func NewTenantRepository(db *sqlx.DB) *TenantRepository {
	return &TenantRepository{db: db}
}

func (r *TenantRepository) Create(ctx context.Context, tenant *domain.Tenant) error {
	query := `INSERT INTO tenants (id, name, slug, created_at, updated_at)
	          VALUES (:id, :name, :slug, :created_at, :updated_at)
	          RETURNING id, created_at, updated_at`
	rows, err := r.db.NamedQueryContext(ctx, query, tenant)
	if err != nil {
		return err
	}
	defer rows.Close()
	if rows.Next() {
		return rows.Scan(&tenant.ID, &tenant.CreatedAt, &tenant.UpdatedAt)
	}
	return nil
}

func (r *TenantRepository) FindBySlug(ctx context.Context, slug string) (*domain.Tenant, error) {
	var tenant domain.Tenant
	err := r.db.GetContext(ctx, &tenant, "SELECT id, name, slug, created_at, updated_at FROM tenants WHERE slug = $1", slug)
	if err != nil {
		return nil, err
	}
	return &tenant, nil
}
