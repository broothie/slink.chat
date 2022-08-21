package model

import "time"

const (
	TypeChannel   Type = "channel"
	WorldChatName      = "World Chat"
)

type Channel struct {
	ChannelID string    `firestore:"channel_id" json:"channelID"`
	Type      Type      `firestore:"type" json:"type"`
	CreatedAt time.Time `firestore:"created_at" json:"createdAt"`
	UpdatedAt time.Time `firestore:"updated_at" json:"updatedAt"`

	Name    string   `firestore:"name" json:"name"`
	UserID  string   `firestore:"user_id" json:"userID"`
	UserIDs []string `firestore:"user_ids" json:"userIDs"`
	Private bool     `firestore:"private" json:"private"`
}

func (Channel) ModelType() Type {
	return TypeChannel
}
