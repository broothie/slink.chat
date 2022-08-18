import {createAsyncThunk, createSlice} from "@reduxjs/toolkit";
import {Channel} from "../model/model";
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
