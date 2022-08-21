import {configureStore} from '@reduxjs/toolkit'
import logger from 'redux-logger'

import userSlice from './userSlice'
import channelsSlice from "./channelsSlice";
import usersSlice from "./usersSlice";
import messagesSlice from "./messagesSlice";

const store = configureStore({
	reducer: {
		user: userSlice.reducer,
		users: usersSlice.reducer,
		channels: channelsSlice.reducer,
		messages: messagesSlice.reducer,
	},
	middleware: (getDefaultMiddleware) => getDefaultMiddleware().concat(logger),
})

export default store
