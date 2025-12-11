package model

import (
	"time"

	"github.com/google/uuid"
)

type Review struct {
	Id        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	GameID    uuid.UUID `json:"game_id"`
	Rating    int       `json:"rating"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
