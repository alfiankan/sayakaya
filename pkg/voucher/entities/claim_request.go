package entities

import "github.com/google/uuid"

type ClaimRequest struct {
	UserID uuid.UUID `json:"user_id"`
}
