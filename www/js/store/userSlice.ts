import {createAsyncThunk, createSlice, PayloadAction} from "@reduxjs/toolkit";
import {User} from "../model/user";
import axios from "../axios";

export const fetchCurrentUser = createAsyncThunk(
	'users/fetchCurrentUser',
	async () => {
		try {
			const response = await axios.get('/api/v1/user')
			return response.data.user as User
		} catch(error) {
			// console.error(error)
			return null
		}
	}
)

type SliceState = { status: 'not checked', user: null } | { status: 'checking', user: null } | { status: 'checked', user?: User }

const userSlice = createSlice({
	name: 'user',
	initialState: {status: 'not checked'} as SliceState,
	reducers: {
		// update: (state, action: PayloadAction<SliceState>) => action.payload,
	},
	extraReducers: builder => {
		builder.addCase(fetchCurrentUser.pending, (state, action) => {
      state.status = 'checking'
		})

		builder.addCase(fetchCurrentUser.fulfilled, (state, action) => {
      state.status = 'checked'
      state.user = action.payload
		})
	}
})

// export const { update } = userSlice.actions

export default userSlice
