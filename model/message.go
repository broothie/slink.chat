package model

import (
	"encoding/json"
	"time"

	"github.com/TwiN/go-away"
)

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

func (m Message) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{
		"messageID": m.ID,
		"createdAt": m.CreatedAt,
		"updatedAt": m.UpdatedAt,
		"userID":    m.UserID,
		"channelID": m.ChannelID,
		"body":      goaway.Censor(m.Body),
	})
}
