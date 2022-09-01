package model

import "time"

const TypeMessage Type = "message"

type Message struct {
	ID        string    `firestore:"id" json:"messageID"`
	CreatedAt time.Time `firestore:"created_at" json:"createdAt"`
	UpdatedAt time.Time `firestore:"updated_at" json:"updatedAt"`

	UserID    string `firestore:"user_id" json:"userID"`
	ChannelID string `firestore:"channel_id" json:"channelID"`
	Body      string `firestore:"body" json:"body"`
}

func (Message) Type() Type {
	return TypeMessage
}
