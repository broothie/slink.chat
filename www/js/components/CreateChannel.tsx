import * as React from 'react'
import TitleBar, {CloseFunction} from "./TitleBar";
import {useState} from "react";
import {useAppDispatch} from "../hooks";
import {createChannel} from "../store/channelsSlice";

export default function CreateChannel({close}: { close: CloseFunction }) {
	const [name, setName] = useState('')

	const dispatch = useAppDispatch()

	function create() {
		dispatch(createChannel({ name, isPrivate: false }))
		close()
	}

	function onInputChange(event) {
		if (event.key === 'enter') {
			create()
		} else {
			setName(event.target.value)
		}
	}

	return (
		<div className="window w-72">
			<div className="draggable-handle">
				<TitleBar title="Create Channel" close={close}/>
			</div>

			<div className="font-sans p-2 flex flex-row space-x-1">
				<input
					type="text"
					placeholder="Channel name"
					className="bg-white input text-sm flex-grow"
					autoFocus={true}
					onChange={onInputChange}
				/>
				<button className="button px-1 text-sm" onClick={create}>Create</button>
			</div>
		</div>
	)
}
