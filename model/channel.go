package model

import "time"

const TypeChannel Type = "channel"

type Channel struct {
	ChannelID string    `firestore:"channel_id" json:"channelID"`
	Type      Type      `firestore:"type" json:"type"`
	CreatedAt time.Time `firestore:"created_at" json:"createdAt"`
	UpdatedAt time.Time `firestore:"updated_at" json:"updatedAt"`

	UserID  string `firestore:"user_id" json:"userID"`
	Name    string `firestore:"name" json:"name"`
	Private bool   `firestore:"private" json:"private"`
}

func (Channel) ModelType() Type {
	return TypeChannel
}
