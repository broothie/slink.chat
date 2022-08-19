import * as React from 'react'
import TitleBar from "./TitleBar";
import {useState} from "react";
import {User} from "../model/model";

export default function CreateChat({ close }: { close: { () } }) {
	const [query, setQuery] = useState('')
	const [users, setUsers] = useState([] as User[])

	return (
		<div className="window flex flex-col w-80 p-1">
			<TitleBar title="Create Chat" close={close}/>

			<div className="flex flex-col p-1 font-sans space-y-1">
				<div className="flex flex-row space-x-1">
					<input
						type="text"
						placeholder="Search for users by screenname..."
						className="flex-grow input"
						value={query}
						onChange={e => setQuery(e.target.value)}
					/>

					<button className="button px-1">Add</button>
				</div>

				<div className="hr my-1"></div>

				<div className="flex flex-col inset h-52 overflow-y-auto bg-white">
					{users.map(user => (
						<div>{user.screenname}</div>
					))}
				</div>

				<div className="flex flex-row justify-end">
					<button className="button px-1">Create</button>
				</div>
			</div>
		</div>
	)
}
