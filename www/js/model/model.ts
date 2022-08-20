
export type User = {
	userID: string,
	screenname: string,
}

export type Subscription = {
	subscriptionID: string,
	userID: string
	channelID: string
}

export type Channel = {
	channelID: string,
	name: string,
	private: boolean,
}

export type Message = {
	messageID: string,
	body: string,
	createdAt: string,
	userID: string,
	channelID: string,
}
