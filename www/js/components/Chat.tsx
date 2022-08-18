import * as React from 'react'
import {useEffect, useRef, useState} from "react";
import * as _ from "lodash";
import {useAppDispatch, useAppSelector} from "../hooks";
import {createMessage, fetchMessages, MessageLookup, receiveMessage} from "../store/messagesSlice";
import {Channel, Message} from "../model/model";
import classNames from "classnames";
import {fetchChannel} from "../store/channelsSlice";
import {UserLookup} from "../store/usersSlice";
import axios from "../axios";

type ChannelResponse = {
	channel: Channel,
	messages: MessageLookup,
	users: UserLookup,
}

const offsets = [
	{ top: 'top-10', left: 'left-10' },
	{ top: 'top-11', left: 'left-11' },
	{ top: 'top-12', left: 'left-12' },
	{ top: 'top-14', left: 'left-14' },
	{ top: 'top-16', left: 'left-16' },
]

export default function Chat({ channelID, offset }: { channelID: string, offset: number }) {
	const user = useAppSelector(state => state.user.user)
	const offsetValues = offsets[offset % offsets.length]
	const windowRef = useRef()

	const [message, setMessage] = useState('')
	const [channel, setChannel] = useState(null as Channel)
	const [messages, setMessages] = useState([] as Message[])
	const [users, setUsers] = useState({} as UserLookup)

	async function sendMessage() {
		try {
			await axios.post(`/api/v1/channels/${channelID}/messages`, {body: message})
			setMessage('')
		} catch (error) {
			console.error(error)
		}
	}

	let socket
	function startSocket() {
		socket = new WebSocket(`ws://${location.host}/api/v1/channels/${channelID}/messages/subscribe`)

		socket.onopen = () => { console.log('open', channelID) }
		socket.onmessage = event => {
			const message = JSON.parse(event.data) as Message
			setMessages(messages => _.sortBy(messages.concat([message]), 'createdAt'))
		}

		socket.onclose = event => {
			console.log('server closed socket', {channelID, event})
			setTimeout(startSocket, 1000)
		}
	}

	useEffect(() => {
		if (!!windowRef.current) windowRef.current?.scrollTo(0, windowRef.current?.scrollHeight)
	}, [messages])

	useEffect(() => {
		axios.get(`/api/v1/channels/${channelID}`)
			.then(({data}: { data: ChannelResponse } ) => {
				setChannel(data.channel)
				setMessages(_.sortBy(data.messages, 'createdAt'))
				setUsers(data.users)
			})
			.catch(console.error)
	}, [])

	useEffect(() => {
		startSocket()
		return () => { socket.close() }
	}, [])

	return channel && (
		<div className={`window p-1 flex flex-col w-fit absolute ${offsetValues.top} ${offsetValues.left}`}>
			<div className="title-bar">
				<p className="px-0.5 text-sm">{channel.name} - Instant Message</p>
			</div>

			<div className="hr my-1"></div>

			<div className="px-3 font-sans">
				<div
					className="bg-white inset w-80 h-52 font-serif text-sm p-1 overflow-y-auto whitespace-pre-wrap"
					ref={windowRef}
				>
					{_.sortBy(messages, 'createdAt').map(message => (
						<p key={message.id}>
							<span className={classNames({
								'text-indigo-700': message.userID === user.id,
								'text-red-500': message.userID !== user.id,
							})}>
								{users[message.userID]?.screenname}:
							</span>
							<span> {message.body}</span>
						</p>
					))}
				</div>

				<div className="hr my-0.5"></div>

				<textarea
					name="message"
					id="message"
					rows={2}
					className="bg-white inset resize-none w-full p-1 outline-0 font-serif text-sm overflow-y-auto"
					autoFocus={true}
					value={message}
					onChange={e => setMessage(e.target.value)}
				>
			</textarea>

				<div className="flex flex-row justify-end pb-2">
					<button
						className="button px-1 py-0.5 text-sm"
						onClick={sendMessage}
					>
						Send
					</button>
				</div>
			</div>
		</div>
	)
}
