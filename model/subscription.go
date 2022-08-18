package model

import "time"

const WorldChatName = "World Chat"

type Subscription struct {
	ID        string    `firestore:"id" json:"id"`
	CreatedAt time.Time `firestore:"created_at" json:"createdAt"`
	UpdatedAt time.Time `firestore:"updated_at" json:"updatedAt"`
	UserID    string    `firestore:"user_id" json:"userID"`
	ChannelID string    `firestore:"channel_id" json:"channelID"`
}
