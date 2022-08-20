import { createAsyncThunk, createSlice } from "@reduxjs/toolkit";
import { Channel, Subscription, User } from "../model/model";
import axios from "../axios";

export type ChannelLookup = { [key: string]: Channel }

export const fetchChannels = createAsyncThunk(
	'channels/fetchChannels',
	async () => {
		const response = await axios.get('/api/v1/channels')
		return response.data.channels as ChannelLookup
	}
)

export const createChannel = createAsyncThunk(
	'channels/createChannel',
	async ({ name, isPrivate }: { name: string, isPrivate: boolean }) => {
		const response = await axios.post('/api/v1/channels', { name, 'private': isPrivate })
		return response.data.channel as Channel
	}
)

export const createChat = createAsyncThunk(
	'channels/createChat',
	async (userIDs: string[]) => {
		const response = await axios.post('/api/v1/chats', userIDs)
		return response.data.channel as Channel
	}
)

export const destroyChannel = createAsyncThunk(
	'channels/destroyChannel',
	async (channelID: string) => {
		const response = await axios.delete(`/api/v1/channels/${channelID}`)
		return response.data.channelID
	}
)

const channelsSlice = createSlice({
	name: 'channels',
	initialState: {} as ChannelLookup,
	reducers: {},
	extraReducers: builder => {
		builder.addCase(fetchChannels.fulfilled, (state, action) => {
			return action.payload
		})

		builder.addCase(createChannel.fulfilled, (state, action) => {
			const channel = action.payload
			state[channel.channelID] = channel
		})

		builder.addCase(createChat.fulfilled, (state, action) => {
			const channel = action.payload
			state[channel.channelID] = channel
		})

		builder.addCase(destroyChannel.fulfilled, (state, action) => {
			const channelID = action.payload
			delete state[channelID]
		})
	}
})

export default channelsSlice
