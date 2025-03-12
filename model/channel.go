package model

import "time"

const (
	TypeChannel Type = "channel"

	ChannelNameWorldChat = "World Chat"
)

type Channel struct {
	ID        string    `firestore:"id" json:"channelID"`
	CreatedAt time.Time `firestore:"created_at" json:"createdAt"`
	UpdatedAt time.Time `firestore:"updated_at" json:"updatedAt"`

	Name              string    `firestore:"name" json:"name"`
	UserID            string    `firestore:"user_id" json:"userID"`
	UserIDs           []string  `firestore:"user_ids" json:"userIDs"`
	Private           bool      `firestore:"private" json:"private"`
	LastMessageSentAt time.Time `firestore:"last_message_sent_at" json:"lastMessageSentAt"`
}

func (Channel) Type() Type {
	return TypeChannel
}
