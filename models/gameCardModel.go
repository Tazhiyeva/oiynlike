package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type HostUser struct {
	FirstName string `json:"first_name,omitempty" bson:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty" bson:"last_name,omitempty"`
	UserID    string `json:"user_id,omitempty" bson:"user_id,omitempty"`
	PhotoURL  string `json:"photo_url,omitempty" bson:"photo_url,omitempty"`
	City      string `json:"city,omitempty" bson:"city,omitempty"`
}

type MatchedPlayer struct {
	FirstName string `json:"first_name,omitempty" bson:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty" bson:"last_name,omitempty"`
	UserID    string `json:"user_id,omitempty" bson:"user_id,omitempty"`
	PhotoURL  string `json:"photo_url,omitempty" bson:"photo_url,omitempty"`
	City      string `json:"city,omitempty" bson:"city,omitempty"`
}

type GameCard struct {
	ID             primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	HostUser       HostUser           `json:"host_user" bson:"host_user"`
	Title          string             `json:"title" bson:"title" validate:"required"`
	Description    string             `json:"description" bson:"description" validate:"required"`
	MaxPlayers     int                `json:"max_players" bson:"max_players" validate:"gt=0"`
	Status         string             `json:"status" bson:"status" validate:"oneof=active inactive moderating"`
	CurrentPlayers int                `json:"current_players" bson:"current_players"`
	MatchedPlayers []MatchedPlayer    `json:"matched_players" bson:"matched_players"`
	CreatedAt      time.Time          `json:"created_at,omitempty" bson:"created_at,omitempty"`
	UpdatedAt      time.Time          `json:"updated_at,omitempty" bson:"updated_at,omitempty"`
	ScheduledTime  time.Time          `json:"scheduled_time,omitempty" bson:"scheduled_time,omitempty" validate:"required"`
}
