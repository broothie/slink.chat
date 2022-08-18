import * as React from "react";
import {useAppDispatch} from "../hooks";
import {createUser} from "../store/userSlice";
import AuthWindow from "./AuthWindow";

export default function SignUp() {
	const dispatch = useAppDispatch()

	function signUp(screenname: string, password: string) {
		dispatch(createUser({ screenname, password }))
	}

	return <AuthWindow title="Sign Up" swapText="Sign On" swapLink="/signon" submit={signUp}/>
}
