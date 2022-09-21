import * as React from "react";
import { useAppDispatch } from "../hooks";
import { createUser } from "../store/userSlice";
import AuthWindow from "./AuthWindow";
import { playDoorOpen } from "../audio";
import {useState} from "react";

export default function SignUp() {
	const dispatch = useAppDispatch()
	const [messages, setMessages] = useState([])

	async function signUp(screenname: string, password: string) {
		try {
			await dispatch(createUser({ screenname, password })).unwrap()
			await playDoorOpen()
		} catch (error) {
			if (error instanceof Array) {
				setMessages(error)
			}
		}
	}

	return <AuthWindow title="Sign Up" swapText="Sign On" swapLink="/signon" submit={signUp} messages={messages} />
}
