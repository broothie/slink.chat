import {createSlice} from "@reduxjs/toolkit";
import {User} from "../model/model";

export type UserLookup = { [key: string]: User }

const usersSlice = createSlice({
	name: 'users',
	initialState: {} as UserLookup,
	reducers: {},
})

export default usersSlice
