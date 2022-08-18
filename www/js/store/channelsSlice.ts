import {createAsyncThunk, createSlice} from "@reduxjs/toolkit";
import {Channel} from "../model/model";
import axios from "../axios";
import {UserLookup} from "./usersSlice";

export type ChannelLookup = { [key: string]: Channel }

type ChannelResponse = {
	channel: Channel,
	users: UserLookup,
}

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

export const fetchChannel = createAsyncThunk(
	'channels/fetchChannel',
	async (channelID: string) => {
		try {
			const response = await axios.get(`/api/v1/channels/${channelID}`)
			return response.data as ChannelResponse
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
	}
})

export default channelsSlice
