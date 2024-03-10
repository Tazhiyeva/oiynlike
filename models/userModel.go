package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID           primitive.ObjectID `bson:"_id" json:"id"`
	FirstName    string             `bson:"first_name" json:"first_name" validate:"required,omitempty"`
	LastName     string             `bson:"last_name" json:"last_name" validate:"required,omitempty"`
	Password     string             `bson:"password" json:"password" validate:"required,min=8"`
	Email        string             `bson:"email" json:"email" validate:"email,omitempty,required"`
	Token        string             `bson:"token" json:"token"`
	UserType     string             `bson:"user_type" json:"user_type" validate:"required,eq=ADMIN|eq=USER"`
	RefreshToken string             `bson:"refresh_token" json:"refresh_token"`
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time          `bson:"updated_at" json:"updated_at"`
	UserId       string             `bson:"user_id" json:"user_id"`
	PhotoURL     string             `bson:"photo_url" json:"photo_url" validate:"omitempty"`
	City         string             `bson:"city" json:"city" validate:"omitempty"`
	AboutUser    string             `bson:"about_user" json:"about_user" validate:"omitempty"`
}
