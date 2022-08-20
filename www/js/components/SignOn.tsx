import * as React from "react";
import { useAppDispatch } from "../hooks";
import { createSession } from "../store/userSlice";
import AuthWindow from "./AuthWindow";
import { playDoorOpen } from "../audio";

export default function SignOn() {
	const dispatch = useAppDispatch()

	function signOn(screenname: string, password: string) {
		dispatch(createSession({ screenname, password }))
			.unwrap()
			.then(playDoorOpen)
	}

	return <AuthWindow title="Sign On" swapText="Get a Screen Name" swapLink="/signup" submit={signOn} />
}
