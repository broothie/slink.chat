package model

import "time"

const TypeMessage Type = "message"

type Message struct {
	MessageID string    `firestore:"message_id" json:"messageID"`
	Type      Type      `firestore:"type" json:"type"`
	CreatedAt time.Time `firestore:"created_at" json:"createdAt"`
	UpdatedAt time.Time `firestore:"updated_at" json:"updatedAt"`

	UserID    string `firestore:"user_id" json:"userID"`
	ChannelID string `firestore:"channel_id" json:"channelID"`
	Body      string `firestore:"body" json:"body"`
}

func (Message) ModelType() Type {
	return TypeMessage
}
