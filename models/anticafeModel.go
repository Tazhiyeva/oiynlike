package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AnticafeModel struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Title       string             `json:"title" bson:"title" validate:"required"`
	Rating      string             `json:"rating" bson:"rating"`
	Address     string             `json:"address" bson:"address" validate:"required"`
	OpeningTime time.Time          `json:"openingTime" bson:"openingTime" validate:"required"`
	ClosingTime time.Time          `json:"closingTime" bson:"closingTime" validate:"required"`
	PhoneNumber string             `json:"phoneNumber" bson:"phoneNumber" validate:"required"`
	Description string             `json:"description" bson:"description"`
	Photos      []string           `json:"photos" bson:"photos"`
	Latitude    string             `json:"latitude" bson:"latitude"`
	Longitude   string             `json:"longitude" bson:"longitude"`
	CreatedAt   time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt   time.Time          `json:"updatedAt" bson:"updatedAt"`
}
