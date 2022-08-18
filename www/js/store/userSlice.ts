import {createAsyncThunk, createSlice} from "@reduxjs/toolkit";
import {User} from "../model/user";
import axios from "../axios";

export const createUser = createAsyncThunk(
	'users/createUser',
	async (params: { screenname: string, password: string }) => {
		try {
			const response = await axios.post('/api/v1/users', JSON.stringify(params))
			return response.data.user as User
		} catch(error) {
			return null
		}
	}
)

export const createSession = createAsyncThunk(
	'users/createSession',
	async (params: { screenname: string, password: string }) => {
		try {
			const response = await axios.post('/api/v1/session', JSON.stringify(params))
			return response.data.user as User
		} catch(error) {
			return null
		}
	}
)

export const destroySession = createAsyncThunk(
	'users/destroySession',
	async () => {
		try {
			await axios.delete('/api/v1/session')
		} catch(error) {
			return null
		}
	}
)

export const fetchCurrentUser = createAsyncThunk(
	'users/fetchCurrentUser',
	async () => {
		try {
			const response = await axios.get('/api/v1/user')
			return response.data.user as User
		} catch(error) {
			return null
		}
	}
)

type SliceState = { status: 'not checked', user: null } | { status: 'checking', user: null } | { status: 'checked', user?: User }

const userSlice = createSlice({
	name: 'user',
	initialState: {status: 'not checked'} as SliceState,
	reducers: {},
	extraReducers: builder => {
		builder.addCase(createUser.fulfilled, (state, action) => {
			state.user = action.payload
		})

		builder.addCase(createSession.fulfilled, (state, action) => {
			state.user = action.payload
		})

		builder.addCase(destroySession.fulfilled, (state, action) => {
			state.user = null
		})

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
