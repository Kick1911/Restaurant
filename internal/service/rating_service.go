package service

import (
	"context"
	"errors"
	"log/slog"

	"github.com/google/uuid"
	"github.com/kick/sigma-connected/internal/domain"
	"github.com/kick/sigma-connected/internal/dto"
	"github.com/kick/sigma-connected/internal/repository"
)

type RatingService struct {
	ratingRepo *repository.RatingRepository
}

func NewRatingService(ratingRepo *repository.RatingRepository) *RatingService {
	return &RatingService{ratingRepo: ratingRepo}
}

func (s *RatingService) Create(ctx context.Context, dishID, userID uuid.UUID, req dto.CreateRatingRequest) (*dto.RatingResponse, error) {
	rating := &domain.Rating{
		ID:     uuid.New(),
		DishID: dishID,
		UserID: userID,
		Rating: req.Rating,
	}

	if err := s.ratingRepo.Create(ctx, rating); err != nil {
		slog.Error("create rating", "error", err, "dish_id", dishID, "user_id", userID)
		return nil, errors.New("failed to create rating")
	}

	return &dto.RatingResponse{
		ID:     rating.ID.String(),
		DishID: rating.DishID.String(),
		UserID: rating.UserID.String(),
		Rating: rating.Rating,
	}, nil
}

func (s *RatingService) GetByDishID(ctx context.Context, dishID uuid.UUID) ([]dto.RatingResponse, error) {
	ratings, err := s.ratingRepo.FindByDishID(ctx, dishID)
	if err != nil {
		slog.Error("get ratings by dish", "error", err, "dish_id", dishID)
		return nil, errors.New("failed to fetch ratings")
	}

	responses := make([]dto.RatingResponse, len(ratings))
	for i, r := range ratings {
		responses[i] = dto.RatingResponse{
			ID:     r.ID.String(),
			DishID: r.DishID.String(),
			UserID: r.UserID.String(),
			Rating: r.Rating,
		}
	}

	return responses, nil
}
