package models

import (
	"time"

	"github.com/google/uuid"
)

type FeedbackRequest struct {
	Rating  int       `json:"rating"`
	Comment string    `json:"comment"`
	OrderId uuid.UUID `json:"order_id"`
}

type FeedbackResponse struct {
	Id        uuid.UUID `json:"id"`
	OrderId   uuid.UUID `json:"order_id"`
	CourierId uuid.UUID `json:"courier_id"`
	Rating    int       `json:"rating"`
	Comment   string    `json:"comment"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
