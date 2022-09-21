import * as React from "react";
import { useAppDispatch } from "../hooks";
import { createSession } from "../store/userSlice";
import AuthWindow from "./AuthWindow";
import { playDoorOpen } from "../audio";
import {useState} from "react";

export default function SignOn() {
	const dispatch = useAppDispatch()
	const [messages, setMessages] = useState([])

	async function signOn(screenname: string, password: string) {
		setMessages([])

		try {
			await dispatch(createSession({ screenname, password })).unwrap()
			await playDoorOpen()
		} catch (error) {
			if (error instanceof Array) {
				setMessages(error)
			}
		}
	}

	return <AuthWindow title="Sign On" swapText="Get a Screen Name" swapLink="/signup" submit={signOn} messages={messages} />
}
