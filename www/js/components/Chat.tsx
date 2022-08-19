import * as React from 'react'
import {useEffect, useRef, useState} from "react";
import * as _ from "lodash";
import {useAppSelector} from "../hooks";
import {Channel, Message, User} from "../model/model";
import classNames from "classnames";
import {UserLookup} from "../store/usersSlice";
import axios from "../axios";
import TitleBar from "./TitleBar";
import {playMessageReceive, playMessageSend} from "../audio";

type CloseFunction = { (): void }

type MessageLookup = { [key: string]: Message }

type ChannelResponse = {
	channel: Channel,
	messages: MessageLookup,
	users: UserLookup,
}

export default function Chat({ channelID, close }: { channelID: string, close: CloseFunction }) {
	const user = useAppSelector(state => state.user.user)
	const windowRef = useRef()

	const [message, setMessage] = useState('')
	const [channel, setChannel] = useState(null as Channel)
	const [messages, setMessages] = useState([] as Message[])
	const [users, setUsers] = useState({} as UserLookup)

	function textareaOnChange(event) {
		if (event.key === 'Enter' && !event.shiftKey) {
			event.preventDefault()
			sendMessage()
		}
	}

	function sendMessage() {
		axios.post(`/api/v1/channels/${channelID}/messages`, {body: message})
			.then(() => setMessage(''))
			.then(playMessageSend)
	}

	function addMessage(message: Message) {
		setMessages(messages => _.sortBy(messages.concat([message]), 'createdAt'))

		if (message.userID !== user.userID) playMessageReceive().catch(console.error)

		if (!users[message.userID]) {
			const userID = message.userID

			axios.get(`/api/v1/users/${userID}`)
				.then(({ data }: { data: { user: User } }) => {
					setUsers(users => _.merge({}, users, { [userID]: data.user }))
				})
		}
	}

	let closedByClient = false
	let socket
	function startSocket() {
		const protocol = location.protocol === 'https:' ? 'wss' : 'ws'
		socket = new WebSocket(`${protocol}://${location.host}/api/v1/channels/${channelID}/messages/subscribe`)

		socket.onopen = () => { console.log('socket opened', channelID) }
		socket.onmessage = event => {
			const message = JSON.parse(event.data) as Message
			addMessage(message)
		}

		socket.onclose = event => {
			if (closedByClient) return

			console.log('server closed socket', {channelID, event})
			setTimeout(startSocket, 1000)
		}
	}

	useEffect(() => {
		if (!!windowRef.current) {
			const window = windowRef.current as HTMLElement
			window?.scrollTo(0, window?.scrollHeight)
		}
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

		return () => {
			closedByClient = true
			socket.close()
			console.log('socket closed', channelID)
		}
	}, [])

	return channel && (
		<div className="window p-1 flex flex-col w-fit">
			<div className="draggable-handle">
				<TitleBar title={`${channel.name} - Instant Message`} close={close}/>
			</div>

			<div className="hr my-1"></div>

			<div className="px-3 font-sans">
				<div
					className="bg-white inset w-80 h-52 font-serif text-sm p-1 overflow-y-auto whitespace-pre-wrap"
					ref={windowRef}
				>
					{messages.map(message => (
						<p key={message.messageID}>
							<span className={classNames({
								'text-indigo-700': message.userID === user.userID,
								'text-red-500': message.userID !== user.userID,
							})}>
								{users[message.userID]?.screenname}:
							</span>
							<span> {message.body}</span>
						</p>
					))}
				</div>

				<div className="hr my-0.5"></div>

				<div>
					<textarea
						name="message"
						id="message"
						rows={2}
						className="bg-white inset resize-none w-full p-1 outline-0 font-serif text-sm overflow-y-auto"
						autoFocus={true}
						value={message}
						onChange={e => setMessage(e.target.value)}
						onKeyDown={textareaOnChange}
					></textarea>

					<div className="flex flex-row justify-end pb-2">
						<button
							type="submit"
							className="button px-1 py-0.5 text-sm"
							disabled={message === ''}
							onClick={sendMessage}
						>
							Send
						</button>
					</div>
				</div>
			</div>
		</div>
	)
}
