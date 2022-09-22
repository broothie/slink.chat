import * as React from 'react'
import {useEffect, useRef, useState} from "react";
import * as _ from "lodash";
import {useAppDispatch, useAppSelector} from "../hooks";
import {Message} from "../model/model";
import {fetchUser} from "../store/usersSlice";
import axios from "../axios";
import TitleBar from "./TitleBar";
import {playMessageReceive, playMessageSend} from "../audio";
import {createChat, fetchChannel, fetchChannelUsers} from "../store/channelsSlice";
import {fetchMessages, receiveMessage} from "../store/messagesSlice";

export default function Chat({ channelID, close, addChannel }: {
	channelID: string,
	addChannel: { (channelID: string) }
	close: { () },
}) {
	const user = useAppSelector(state => state.user.user)
	const channel = useAppSelector(state => state.channels[channelID])
	const messages = useAppSelector(state => _.filter(state.messages, message => message.channelID === channelID))
	const windowRef = useRef()
	const dispatch = useAppDispatch()

	const [message, setMessage] = useState('')

	function onTextareaKeyDown(event) {
		if (event.key === 'Enter' && !event.shiftKey) {
			event.preventDefault()
			sendMessage()
		}
	}

	function sendMessage() {
		axios.post(`/api/v1/channels/${channelID}/messages`, {body: message})
			.then(() => setMessage(''))
	}

	function addMessage(message: Message) {
		dispatch(receiveMessage(message))
		if (message.userID === user.userID) {
			playMessageSend().catch(console.error)
		} else {
			playMessageReceive().catch(console.error)
		}
	}

	useEffect(() => { dispatch(fetchChannel(channelID)) }, [])
	useEffect(() => {
		dispatch(fetchChannelUsers(channelID))
			.unwrap()
			.then(() => dispatch(fetchMessages(channelID)))
	}, [])

	useEffect(() => {
		if (!!windowRef.current) {
			const window = windowRef.current as HTMLElement
			window?.scrollTo(0, window?.scrollHeight)
		}
	}, [messages])

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

			<div className="hr my-1"/>

			<div className="px-3 font-sans">
				<div
					className="bg-white inset w-80 h-52 font-serif text-sm p-1 overflow-y-auto whitespace-pre-wrap"
					ref={windowRef}
				>
					{messages.map(message => (
						<MessageItem key={message.messageID} message={message} addChannel={addChannel}/>
					))}
				</div>

				<div className="hr my-0.5"/>

				<div>
					<textarea
						name="message"
						id="message"
						rows={2}
						className="bg-white inset resize-none w-full p-1 outline-0 font-serif text-sm overflow-y-auto"
						autoFocus={true}
						value={message}
						onChange={e => setMessage(e.target.value)}
						onKeyDown={onTextareaKeyDown}
					/>

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

function MessageItem({ message, addChannel }: {
	message: Message,
	addChannel: { (channelID: string) },
}) {
	const currentUser = useAppSelector(state => state.user.user)
	const messageUser = useAppSelector(state => state.users[message.userID])
	const dispatch = useAppDispatch()

	useEffect(() => {
		if (!messageUser) {
			dispatch(fetchUser(message.userID))
		}
	}, [])

	function onScreennameClick() {
		dispatch(createChat([message.userID, currentUser.userID]))
			.unwrap()
			.then(channel => { addChannel(channel.channelID) })
	}

	return messageUser && (
		<p>
			{message.userID === currentUser.userID ? (
				<span className="text-indigo-700">{messageUser.screenname}:</span>
			) : (
				<a className="text-red-500 cursor-pointer" onClick={onScreennameClick}>
					{messageUser.screenname}:
				</a>
			)}

			<span>&nbsp;{message.body}</span>
		</p>
	)
}
