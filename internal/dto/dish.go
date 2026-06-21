package dto

import "github.com/go-playground/validator/v10"

type CreateDishRequest struct {
	Name        string  `json:"name" validate:"required,min=1,max=200"`
	Description string  `json:"description" validate:"required,max=2000"`
	Price       float64 `json:"price" validate:"required,gt=0"`
	ImageURL    string  `json:"image_url" validate:"omitempty,url"`
}

type UpdateDishRequest struct {
	Name        string  `json:"name" validate:"omitempty,min=1,max=200"`
	Description string  `json:"description" validate:"omitempty,max=2000"`
	Price       float64 `json:"price" validate:"omitempty,gt=0"`
	ImageURL    string  `json:"image_url" validate:"omitempty,url"`
}

type DishSearchParams struct {
	Query string `json:"q" validate:"omitempty,max=200"`
	Page  int    `json:"page" validate:"omitempty,min=1"`
	Limit int    `json:"limit" validate:"omitempty,min=1,max=100"`
}

type DishResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	ImageURL    string  `json:"image_url"`
	AvgRating   float64 `json:"avg_rating,omitempty"`
}

func (r CreateDishRequest) Validate(validate *validator.Validate) error {
	return validate.Struct(r)
}

func (r UpdateDishRequest) Validate(validate *validator.Validate) error {
	return validate.Struct(r)
}
