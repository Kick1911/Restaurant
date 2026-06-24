package domain

import (
	"time"

	"github.com/google/uuid"
)

type Rating struct {
	ID        uuid.UUID `json:"id" db:"id"`
	DishID    uuid.UUID `json:"dish_id" db:"dish_id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	Rating    int       `json:"rating" db:"rating"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
