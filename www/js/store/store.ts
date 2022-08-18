import {configureStore} from '@reduxjs/toolkit'

import userSlice from './userSlice'
import subscriptionSlice from "./subscriptionSlice";

const store = configureStore({
	reducer: {
		user: userSlice.reducer,
		subscriptions: subscriptionSlice.reducer,
	},
})

export default store
