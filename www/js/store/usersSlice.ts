import {createAsyncThunk, createSlice} from "@reduxjs/toolkit";
import {User} from "../model/model";
import axios from "../axios";
import * as _ from "lodash";

export type UserLookup = { [key: string]: User }

export const fetchUser = createAsyncThunk(
	'users/fetchUser',
	async (userID: string) => {
		const response = await axios.get(`/api/v1/users/${userID}`)
		return response.data.user as User
	}
)

export const fetchUsers = createAsyncThunk(
	'users/fetchUsers',
	async (userIDs: string[]) => {
		const response = await axios.get(`/api/v1/users?user_ids=${userIDs.join(',')}`)
		return response.data.users as UserLookup
	}
)

const usersSlice = createSlice({
	name: 'users',
	initialState: {} as UserLookup,
	reducers: {},
	extraReducers: builder => {
		builder.addCase(fetchUser.fulfilled, (state, action) => {
			const user = action.payload
			return _.merge({}, state, { [user.userID]: user })
		})

		builder.addCase(fetchUsers.fulfilled, (state, action) => {
			const users = action.payload
			return _.merge({}, state, users)
		})
	}
})

export default usersSlice
