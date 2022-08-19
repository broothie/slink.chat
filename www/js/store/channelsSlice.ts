import {createAsyncThunk, createSlice} from "@reduxjs/toolkit";
import {Channel, User} from "../model/model";
import axios from "../axios";

export type ChannelLookup = { [key: string]: Channel }

export const fetchChannels = createAsyncThunk(
	'channels/fetchChannels',
	async () => {
		try {
			const response = await axios.get('/api/v1/channels')
			return response.data.channels as ChannelLookup
		} catch(error) {
			return null
		}
	}
)

export const createChannel = createAsyncThunk(
	'channels/createChannel',
	async ({ name, isPrivate }: { name: string, isPrivate: boolean }) => {
		try {
			const response = await axios.post('/api/v1/channels', { name, 'private': isPrivate })
			return response.data.channel as Channel
		} catch(error) {
			return null
		}
	}
)

export const createChat = createAsyncThunk(
	'channels/createChat',
	async (users: User[]) => {
		try {
			const response = await axios.post('/api/v1/chats', users)
			return response.data.channel as Channel
		} catch(error) {
			return null
		}
	}
)

export const destroyChannel = createAsyncThunk(
	'channels/destroyChannel',
	async (channelID: string) => {
		try {
			const response = await axios.delete(`/api/v1/channels/${channelID}`)
			return response.data.channelID
		} catch(error) {
			return null
		}
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
