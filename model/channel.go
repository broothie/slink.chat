package model

import "time"

type Channel struct {
	ID        string    `firestore:"id" json:"id"`
	CreatedAt time.Time `firestore:"created_at" json:"createdAt"`
	UpdatedAt time.Time `firestore:"updated_at" json:"updatedAt"`
	UserID    string    `firestore:"user_id" json:"userID"`
	Name      string    `firestore:"name" json:"name"`
	Private   bool      `firestore:"private" json:"private"`
}
