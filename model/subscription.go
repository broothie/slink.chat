package model

import "time"

const (
	TypeSubscription Type = "subscription"
	WorldChatName         = "World Chat"
)

type Subscription struct {
	SubscriptionID string    `firestore:"subscription_id" json:"subscriptionID"`
	Type           Type      `firestore:"type" json:"type"`
	CreatedAt      time.Time `firestore:"created_at" json:"createdAt"`
	UpdatedAt      time.Time `firestore:"updated_at" json:"updatedAt"`

	UserID    string `firestore:"user_id" json:"userID"`
	ChannelID string `firestore:"channel_id" json:"channelID"`
}

func (Subscription) ModelType() Type {
	return TypeSubscription
}
