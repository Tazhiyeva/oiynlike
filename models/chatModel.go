package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Chat struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Title      string             `bson:"title" json:"title"`
	GameCardID primitive.ObjectID `bson:"gamecard_id" json:"gamecard_id"`
	Members    []Sender           `bson:"members" json:"members"`
	Messages   []Message          `bson:"messages" json:"messages"`
}

type Message struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Sender    Sender             `bson:"sender_id" json:"sender_id"`
	Content   string             `bson:"content" json:"content"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}

type Sender struct {
	FirstName string `json:"first_name,omitempty" bson:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty" bson:"last_name,omitempty"`
	UserID    string `json:"user_id,omitempty" bson:"user_id,omitempty"`
	PhotoURL  string `json:"photo_url,omitempty" bson:"photo_url,omitempty"`
}
