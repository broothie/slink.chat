import * as React from 'react'
import {useAppDispatch, useAppSelector} from "../hooks";
import {destroySession} from "../store/userSlice";
import {useEffect} from "react";
import {destroyChannel, fetchChannels} from "../store/channelsSlice";
import * as _ from 'lodash'
import TitleBar from "./TitleBar";

export default function ChannelList({ addChannel, openCreateChannel }: {
	addChannel: { (channelID: string): void },
	openCreateChannel: { (): void }
}) {
	const dispatch = useAppDispatch()
	const user = useAppSelector(state => state.user.user)
	const channels = useAppSelector(state => state.channels)

	function signOff() {
		dispatch(destroySession())
	}

	function removeChannel(channelID) {
		dispatch(destroyChannel(channelID))
	}

	useEffect(() => {
		dispatch(fetchChannels())
	}, [])

	const privateChannels = _.filter(channels, 'private')
	const publicChannels = _.reject(channels, 'private')

	return (
		<div className="window p-1 flex flex-col w-fit" style={{ height: 800 }}>
			<div className="draggable-handle">
				<TitleBar title="Channel List" close={signOff}/>
			</div>

			<div className="px-2 py-1 font-sans flex-grow flex flex-col">
				<div className="text-sm flex flex-row justify-between">
					<p>Welcome, {user.screenname}!</p>
					<a className="link" onClick={signOff}>Sign Off</a>
				</div>

				<div className="hr mb-1"></div>

				<div className="bg-logo-tile flex flex-row items-center py-3 px-10 space-x-2">
					<div className="">
						<img src="/static/logo.png" alt="Slink logo" className="w-24 h-auto"/>
					</div>

					<p className="flex flex-col leading-tight">
						<span className="text-white">Slink</span>
						<span className="text-white italic">Instant</span>
						<span className="text-white italic">Messenger</span>
					</p>
				</div>

				<div className="hr my-0.5"></div>

				<div className="bg-white inset px-2 py-1 text-sm flex-grow">
					<div>
						<div className="p-1 border-b border-black">
							<p>Chats</p>
						</div>

						<div>
							{_.map(privateChannels, channel => (
								<div
									key={channel.channelID}
									className="pl-3 pr-0.5 py-0.5 cursor-pointer"
									onDoubleClick={() => addChannel(channel.channelID)}
								>
									<div className="hover:bg-logo-tile hover:text-white p-0.5 select-none flex flex-row justify-between">
										<p>{channel.name}</p>
										<button className="button px-1 cursor-pointer" onClick={() => removeChannel(channel.channelID)}>-</button>
									</div>
								</div>
							))}
						</div>
					</div>

					<div>
						<div className="p-1 border-b border-black flex flex-row justify-between">
							<p>Channels</p>

							<div>
								<a className="link" onClick={openCreateChannel}>Create</a>
							</div>
						</div>

						<div>
							{_.map(publicChannels, channel => (
								<div
									key={channel.channelID}
									className="pl-3 pr-0.5 py-0.5 cursor-pointer"
									onDoubleClick={() => addChannel(channel.channelID)}
								>
									<div className="hover:bg-logo-tile hover:text-white p-0.5 select-none flex flex-row justify-between">
										<p>{channel.name}</p>
										<button className="button px-1 cursor-pointer" onClick={() => removeChannel(channel.channelID)}>-</button>
									</div>
								</div>
							))}
						</div>
					</div>
				</div>
			</div>
		</div>
	)
}
