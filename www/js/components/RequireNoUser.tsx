import * as React from "react";
import {useEffect} from "react";
import {useAppDispatch, useAppSelector} from "../hooks";
import {fetchCurrentUser} from "../store/userSlice";
import {Navigate} from "react-router-dom";

export default function RequireNoUser({children}) {
	const userState = useAppSelector(state => state.user)
	const dispatch = useAppDispatch()

	useEffect(() => {
		if (userState.status === 'not checked') {
			dispatch(fetchCurrentUser())
		}
	}, [])

	switch (userState.status) {
		case "not checked":
		case "checking":
			return <div></div>
		case "missing":
			return children

		case "checked":
			return <Navigate to="/" />
	}
}
