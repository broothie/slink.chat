import {createAsyncThunk, createSlice, PayloadAction} from "@reduxjs/toolkit";
import {Message} from "../model/model";
import axios from "../axios";
import * as _ from "lodash";

export type MessageLookup = { [key: string]: Message }

export const fetchMessages = createAsyncThunk(
	'messages/fetchMessages',
	async (channelID: string) => {
		const response = await axios.get(`/api/v1/channels/${channelID}/messages`)
		return response.data.messages as MessageLookup
	}
)

const messagesSlice = createSlice({
	name: 'messages',
	initialState: {} as MessageLookup,
	reducers: {
		receiveMessage: (state, action: PayloadAction<Message>) => {
			const message = action.payload
			return _.merge({}, state, { [message.messageID]: message })
		}
	},
	extraReducers: builder => {
		builder.addCase(fetchMessages.fulfilled, (state, action) => {
			return _.merge({}, state, action.payload)
		})
	}
})

export default messagesSlice

export const { receiveMessage } = messagesSlice.actions
