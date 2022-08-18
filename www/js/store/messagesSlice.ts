import {createAsyncThunk, createSlice, PayloadAction} from "@reduxjs/toolkit";
import {Message} from "../model/model";
import axios from "../axios";
import * as _ from "lodash";

export type MessageLookup = { [key: string]: Message }

export const createMessage = createAsyncThunk(
	'messages/createMessage',
	async ({ channelID, body }: {channelID: string, body: string}) => {
		try {
			const response = await axios.post(`/api/v1/channels/${channelID}/messages`, {body})
			return response.data.message as Message
		} catch(error) {
			return null
		}
	}
)

export const fetchMessages = createAsyncThunk(
	'messages/fetchMessages',
	async (channelID: string) => {
		try {
			const response = await axios.get(`/api/v1/channels/${channelID}/messages`)
			return response.data.messages as MessageLookup
		} catch(error) {
			return null
		}
	}
)

const messagesSlice = createSlice({
	name: 'messages',
	initialState: {} as MessageLookup,
	reducers: {
		receiveMessage: (state, action: PayloadAction<Message>) => {
			const message = action.payload
			state[message.id] = message
		}
	},
	extraReducers: builder => {
		builder.addCase(createMessage.fulfilled, (state, action) => {
			console.log('message created', action.payload)
		})

		builder.addCase(fetchMessages.fulfilled, (state, action) => {
			return _.merge(state, action.payload)
		})
	}
})

export const { receiveMessage } = messagesSlice.actions

export default messagesSlice
