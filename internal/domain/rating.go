package domain

import (
	"time"

	"github.com/google/uuid"
)

type Rating struct {
	ID        uuid.UUID `json:"id"`
	DishID    uuid.UUID `json:"dish_id"`
	UserID    uuid.UUID `json:"user_id"`
	Rating    int       `json:"rating"`
	CreatedAt time.Time `json:"created_at"`
}
