package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/kick/sigma-connected/internal/domain"
)

type RatingRepository struct {
	db *sqlx.DB
}

func NewRatingRepository(db *sqlx.DB) *RatingRepository {
	return &RatingRepository{db: db}
}

func (r *RatingRepository) Create(ctx context.Context, rating *domain.Rating) error {
	query := `INSERT INTO ratings (id, dish_id, user_id, rating, created_at)
	          VALUES (:id, :dish_id, :user_id, :rating, :created_at)
	          ON CONFLICT (dish_id, user_id) DO UPDATE SET rating = :rating
	          RETURNING id, created_at`
	rows, err := r.db.NamedQueryContext(ctx, query, rating)
	if err != nil {
		return fmt.Errorf("create rating: %w", err)
	}
	defer rows.Close()
	if rows.Next() {
		return rows.Scan(&rating.ID, &rating.CreatedAt)
	}
	return nil
}

func (r *RatingRepository) FindByDishID(ctx context.Context, dishID uuid.UUID) ([]domain.Rating, error) {
	var ratings []domain.Rating
	err := r.db.SelectContext(ctx, &ratings,
		"SELECT id, dish_id, user_id, rating, created_at FROM ratings WHERE dish_id = $1 ORDER BY created_at DESC", dishID)
	if err != nil {
		return nil, fmt.Errorf("find ratings by dish ID: %w", err)
	}
	return ratings, nil
}

func (r *RatingRepository) FindByUserAndDish(ctx context.Context, userID, dishID uuid.UUID) (*domain.Rating, error) {
	var rating domain.Rating
	err := r.db.GetContext(ctx, &rating,
		"SELECT id, dish_id, user_id, rating, created_at FROM ratings WHERE user_id = $1 AND dish_id = $2",
		userID, dishID)
	if err != nil {
		return nil, fmt.Errorf("find rating by user and dish: %w", err)
	}
	return &rating, nil
}
