package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/kick/sigma-connected/internal/domain"
)

type DishRepository struct {
	db *sqlx.DB
}

func NewDishRepository(db *sqlx.DB) *DishRepository {
	return &DishRepository{db: db}
}

func (r *DishRepository) Create(ctx context.Context, dish *domain.Dish) error {
	query := `INSERT INTO dishes (id, tenant_id, name, description, price, image_url, created_at, updated_at)
	          VALUES (:id, :tenant_id, :name, :description, :price, :image_url, :created_at, :updated_at)
	          RETURNING id, created_at, updated_at`
	rows, err := r.db.NamedQueryContext(ctx, query, dish)
	if err != nil {
		return err
	}
	defer rows.Close()
	if rows.Next() {
		return rows.Scan(&dish.ID, &dish.CreatedAt, &dish.UpdatedAt)
	}
	return nil
}

func (r *DishRepository) FindByID(ctx context.Context, tenantID, id uuid.UUID) (*domain.Dish, error) {
	var dish domain.Dish
	err := r.db.GetContext(ctx, &dish,
		"SELECT id, tenant_id, name, description, price, image_url, created_at, updated_at FROM dishes WHERE tenant_id = $1 AND id = $2",
		tenantID, id)
	if err != nil {
		return nil, err
	}
	return &dish, nil
}

func (r *DishRepository) Update(ctx context.Context, dish *domain.Dish) error {
	query := `UPDATE dishes SET name = :name, description = :description, price = :price,
	          image_url = :image_url, updated_at = NOW()
	          WHERE id = :id AND tenant_id = :tenant_id`
	result, err := r.db.NamedExecContext(ctx, query, dish)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("dish not found")
	}
	return nil
}

func (r *DishRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	result, err := r.db.ExecContext(ctx,
		"DELETE FROM dishes WHERE tenant_id = $1 AND id = $2", tenantID, id)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("dish not found")
	}
	return nil
}

func (r *DishRepository) Search(ctx context.Context, tenantID uuid.UUID, query string, page, limit int) ([]domain.Dish, int, error) {
	offset := (page - 1) * limit

	whereClauses := []string{"tenant_id = :tenant_id"}
	args := map[string]interface{}{
		"tenant_id": tenantID,
		"limit":     limit,
		"offset":    offset,
	}

	if query != "" {
		whereClauses = append(whereClauses, "(name ILIKE :query OR description ILIKE :query)")
		args["query"] = "%" + query + "%"
	}

	where := strings.Join(whereClauses, " AND ")

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM dishes WHERE %s", where)
	rows, err := r.db.NamedQueryContext(ctx, countQuery, args)
	if err != nil {
		return nil, 0, err
	}
	var total int
	if rows.Next() {
		rows.Scan(&total)
	}
	rows.Close()

	dataQuery := fmt.Sprintf("SELECT id, tenant_id, name, description, price, image_url, created_at, updated_at FROM dishes WHERE %s ORDER BY created_at DESC LIMIT :limit OFFSET :offset", where)
	var dishes []domain.Dish
	err = r.db.SelectContext(ctx, &dishes, dataQuery, args)
	if err != nil {
		return nil, 0, err
	}

	return dishes, total, nil
}

func (r *DishRepository) GetAverageRating(ctx context.Context, dishID uuid.UUID) (float64, error) {
	var avg float64
	err := r.db.GetContext(ctx, &avg,
		"SELECT COALESCE(AVG(rating), 0) FROM ratings WHERE dish_id = $1", dishID)
	if err != nil {
		return 0, err
	}
	return avg, nil
}
