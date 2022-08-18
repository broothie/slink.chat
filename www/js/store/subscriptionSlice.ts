import {createAsyncThunk, createSlice} from "@reduxjs/toolkit";
import {Subscription} from "../model/model";
import axios from "../axios";

type SubscriptionLookup = { [key: string]: Subscription }

export const fetchSubscriptions = createAsyncThunk(
	'subscriptions/fetchSubscriptions',
	async () => {
		try {
			const response = await axios.get('/api/v1/subscriptions')
			return response.data.subscriptions as SubscriptionLookup
		} catch(error) {
			return null
		}
	}
)

const subscriptionSlice = createSlice({
	name: 'subscriptions',
	initialState: {} as SubscriptionLookup,
	reducers: {},
	extraReducers: builder => {
		builder.addCase(fetchSubscriptions.fulfilled, (state, action) => action.payload)
	}
})

export default subscriptionSlice
