import {createSlice} from "@reduxjs/toolkit";
import {User} from "../model/model";
import {fetchChannel} from "./channelsSlice";
import * as _ from "lodash";

export type UserLookup = { [key: string]: User }

const usersSlice = createSlice({
	name: 'users',
	initialState: {} as UserLookup,
	reducers: {},
	extraReducers: builder => {
		builder.addCase(fetchChannel.fulfilled, (state, action) => {
			return _.merge(state, action.payload.users)
		})
	}
})

export default usersSlice
