import * as React from 'react'
import TitleBar from "./TitleBar";
import { useEffect, useMemo, useState } from "react";
import { User } from "../model/model";
import * as _ from "lodash";
import axios from "../axios";
import { UserLookup } from "../store/usersSlice";
import { useAppDispatch } from "../hooks";
import { createChat } from "../store/channelsSlice";

export default function CreateChat({ addChannel, close }: {
	addChannel: { (channelID: string) },
	close: { () },
}) {
	const [query, setQuery] = useState('')
	const [users, setUsers] = useState([] as User[])
	const [addedUsers, setAddedUsers] = useState({} as UserLookup)

	const dispatch = useAppDispatch()

	function create() {
		dispatch(createChat(_.keys(addedUsers)))
			.unwrap()
			.then(channel => {
				addChannel(channel.channelID)
				close()
			})
	}

	function searchUsers(query) {
		axios.get(`/api/v1/users/search?query=${query}`)
			.then(({ data }: { data: { users: User[] } }) => {
				setUsers(data.users)
			})
	}

	const throttledSearchUsers = useMemo(() => _.throttle(searchUsers, 100), [])
	useEffect(() => { throttledSearchUsers(query) }, [query])

	return (
		<div className="window flex flex-col w-80 p-1">
			<div className="draggable-handle">
				<TitleBar title="Create Chat" close={close} />
			</div>

			<div className="flex flex-col p-1 font-sans space-y-1">
				<div className="flex flex-row space-x-1">
					<input
						type="text"
						placeholder="Search for users by screenname..."
						className="flex-grow input p-1"
						autoFocus={true}
						value={query}
						onChange={e => setQuery(e.target.value)}
					/>

					<button className="button px-1">Add</button>
				</div>

				<div className="flex flex-col inset h-52 overflow-y-auto bg-white text-sm">
					{_.reject(users, user => _.includes(_.keys(addedUsers), user.userID)).map(user => (
						<div key={user.userID} className="flex flex-row justify-between p-1 hover:bg-logo-tile hover:text-white">
							<p>{user.screenname}</p>
							<button
								className="button px-2"
								onClick={() => setAddedUsers(_.merge({}, addedUsers, { [user.userID]: user }))}
							>
								+
							</button>
						</div>
					))}
				</div>

				<div className="hr my-1"></div>

				<div className="flex flex-col inset h-52 overflow-y-auto bg-white text-sm">
					{_.map(addedUsers, user => (
						<div key={user.userID} className="flex flex-row justify-between p-1 hover:bg-logo-tile hover:text-white">
							<p>{user.screenname}</p>
							<button
								className="button px-2"
								onClick={() => setAddedUsers(() => {
									const copy = _.merge({}, addedUsers)
									delete copy[user.userID]
									return copy
								})}
							>
								-
							</button>
						</div>
					))}
				</div>

				<div className="flex flex-row justify-end">
					<button className="button px-1" onClick={create}>Create</button>
				</div>
			</div>
		</div>
	)
}
