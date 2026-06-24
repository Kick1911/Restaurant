package dto

import "github.com/go-playground/validator/v10"

type CreateRatingRequest struct {
	Rating int `json:"rating" validate:"required,min=1,max=5"`
}

type RatingResponse struct {
	ID     string `json:"id"`
	DishID string `json:"dish_id"`
	UserID string `json:"user_id"`
	Rating int    `json:"rating"`
}

func (r CreateRatingRequest) Validate(validate *validator.Validate) error {
	return validate.Struct(r)
}
