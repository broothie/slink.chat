
export type User = {
	id: string,
	screenname: string,
}

export type Channel = {
	id: string,
	name: string,
	private: boolean,
}

export type Message = {
	id: string,
	body: string,
	createdAt: string,
	userID: string,
	channelID: string,
}
