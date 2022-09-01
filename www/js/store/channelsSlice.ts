import { createAsyncThunk, createSlice } from "@reduxjs/toolkit";
import { Channel } from "../model/model";
import axios from "../axios";
import * as _ from "lodash";
import {UserLookup} from "./usersSlice";

export type ChannelLookup = { [key: string]: Channel }

export const fetchChannels = createAsyncThunk(
	'channels/fetchChannels',
	async () => {
		const response = await axios.get('/api/v1/channels')
		return response.data.channels as ChannelLookup
	}
)

export const fetchChannel = createAsyncThunk(
	'channels/fetchChannel',
	async (channelID: string) => {
		const response = await axios.get(`/api/v1/channels/${channelID}`)
		return response.data.channel as Channel
	}
)

export const fetchChannelUsers = createAsyncThunk(
	'channels/fetchChannelUsers',
	async (channelID: string) => {
		const response = await axios.get(`/api/v1/channels/${channelID}/users`)
		return response.data.users as UserLookup
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
		const response = await axios.post('/api/v1/channels/chats', userIDs)
		return response.data.channel as Channel
	}
)

export const destroyChannel = createAsyncThunk(
	'channels/destroyChannel',
	async (channelID: string) => {
		const response = await axios.delete(`/api/v1/channels/${channelID}/leave`)
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

		builder.addCase(fetchChannel.fulfilled, (state, action) => {
			const channel = action.payload
			return _.merge({}, state, { [channel.channelID]: channel })
		})

		builder.addCase(createChannel.fulfilled, (state, action) => {
			const channel = action.payload
			return _.merge({}, state, { [channel.channelID]: channel })
		})

		builder.addCase(createChat.fulfilled, (state, action) => {
			const channel = action.payload
			return _.merge({}, state, { [channel.channelID]: channel })
		})

		builder.addCase(destroyChannel.fulfilled, (state, action) => {
			const channelID = action.payload
			const copy = _.merge({}, state)
			delete copy[channelID]
			return copy
		})
	}
})

export default channelsSlice
