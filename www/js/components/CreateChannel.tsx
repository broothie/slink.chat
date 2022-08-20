import * as React from 'react'
import TitleBar, { CloseFunction } from "./TitleBar";
import { useRef, useState } from "react";
import { useAppDispatch } from "../hooks";
import { createChannel } from "../store/channelsSlice";

export default function CreateChannel({ addChannel, close }: {
	addChannel: { (channelID: string) },
	close: CloseFunction,
}) {
	const [name, setName] = useState('')

	const formRef = useRef()
	const dispatch = useAppDispatch()

	function create(event) {
		event.preventDefault()

		if (formRef.current !== undefined) {
			const form = formRef.current as HTMLFormElement
			if (form?.checkValidity()) {
				dispatch(createChannel({ name, isPrivate: false }))
					.unwrap()
					.then(channel => {
						addChannel(channel.channelID)
						close()
					})
			}
		}
	}

	return (
		<div className="window w-72 p-1">
			<div className="draggable-handle">
				<TitleBar title="Create Channel" close={close} />
			</div>

			<form ref={formRef} className="font-sans p-2 flex flex-row space-x-1" onSubmit={create}>
				<input
					type="text"
					placeholder="Channel name"
					className="bg-white input text-sm flex-grow"
					autoFocus={true}
					value={name}
					onChange={e => setName(e.target.value)}
				/>

				<button type="submit" className="button px-1 text-sm" disabled={name === ''}>Create</button>
			</form>
		</div>
	)
}
