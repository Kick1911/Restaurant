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

type DishService struct {
	dishRepo *repository.DishRepository
}

func NewDishService(dishRepo *repository.DishRepository) *DishService {
	return &DishService{dishRepo: dishRepo}
}

func (s *DishService) Create(ctx context.Context, tenantID uuid.UUID, req dto.CreateDishRequest) (*dto.DishResponse, error) {
	dish := &domain.Dish{
		ID:          uuid.New(),
		TenantID:    tenantID,
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		ImageURL:    req.ImageURL,
	}

	if err := s.dishRepo.Create(ctx, dish); err != nil {
		slog.Error("create dish", "error", err, "tenant_id", tenantID)
		return nil, errors.New("failed to create dish")
	}

	return &dto.DishResponse{
		ID:          dish.ID.String(),
		Name:        dish.Name,
		Description: dish.Description,
		Price:       dish.Price,
		ImageURL:    dish.ImageURL,
	}, nil
}

func (s *DishService) GetByID(ctx context.Context, tenantID, dishID uuid.UUID) (*dto.DishResponse, error) {
	dish, err := s.dishRepo.FindByID(ctx, tenantID, dishID)
	if err != nil {
		slog.Warn("get dish by id: not found", "error", err, "dish_id", dishID, "tenant_id", tenantID)
		return nil, errors.New("dish not found")
	}

	avgRating, err := s.dishRepo.GetAverageRating(ctx, dish.ID)
	if err != nil {
		slog.Error("get dish: average rating", "error", err, "dish_id", dish.ID)
	}

	return &dto.DishResponse{
		ID:          dish.ID.String(),
		Name:        dish.Name,
		Description: dish.Description,
		Price:       dish.Price,
		ImageURL:    dish.ImageURL,
		AvgRating:   avgRating,
	}, nil
}

func (s *DishService) Update(ctx context.Context, tenantID, dishID uuid.UUID, req dto.UpdateDishRequest) (*dto.DishResponse, error) {
	dish, err := s.dishRepo.FindByID(ctx, tenantID, dishID)
	if err != nil {
		return nil, errors.New("dish not found")
	}

	if req.Name != "" {
		dish.Name = req.Name
	}
	if req.Description != "" {
		dish.Description = req.Description
	}
	if req.Price > 0 {
		dish.Price = req.Price
	}
	if req.ImageURL != "" {
		dish.ImageURL = req.ImageURL
	}

	if err := s.dishRepo.Update(ctx, dish); err != nil {
		slog.Error("update dish", "error", err, "dish_id", dishID, "tenant_id", tenantID)
		return nil, errors.New("failed to update dish")
	}

	return &dto.DishResponse{
		ID:          dish.ID.String(),
		Name:        dish.Name,
		Description: dish.Description,
		Price:       dish.Price,
		ImageURL:    dish.ImageURL,
	}, nil
}

func (s *DishService) Delete(ctx context.Context, tenantID, dishID uuid.UUID) error {
	if err := s.dishRepo.Delete(ctx, tenantID, dishID); err != nil {
		slog.Error("delete dish", "error", err, "dish_id", dishID, "tenant_id", tenantID)
		return errors.New("dish not found")
	}
	return nil
}

func (s *DishService) Search(ctx context.Context, tenantID uuid.UUID, query string, page, limit int) ([]dto.DishResponse, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	dishes, total, err := s.dishRepo.Search(ctx, tenantID, query, page, limit)
	if err != nil {
		slog.Error("search dishes", "error", err, "tenant_id", tenantID, "query", query)
		return nil, 0, errors.New("failed to search dishes")
	}

	responses := make([]dto.DishResponse, len(dishes))
	for i, dish := range dishes {
		avgRating, _ := s.dishRepo.GetAverageRating(ctx, dish.ID)
		responses[i] = dto.DishResponse{
			ID:          dish.ID.String(),
			Name:        dish.Name,
			Description: dish.Description,
			Price:       dish.Price,
			ImageURL:    dish.ImageURL,
			AvgRating:   avgRating,
		}
	}

	return responses, total, nil
}
