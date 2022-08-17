package model

import "time"

type Message struct {
	ID        string    `firestore:"id" json:"id"`
	CreatedAt time.Time `firestore:"created_at" json:"createdAt"`
	UpdatedAt time.Time `firestore:"updated_at" json:"updatedAt"`
	UserID    string    `firestore:"user_id" json:"userID"`
	Body      string    `firestore:"body" json:"body"`
}
