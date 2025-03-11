import * as React from 'react'
import { useAppDispatch, useAppSelector } from "../hooks";
import { destroySession } from "../store/userSlice";
import { useEffect } from "react";
import { destroyChannel, fetchChannels } from "../store/channelsSlice";
import * as _ from 'lodash'
import TitleBar from "./TitleBar";
import {playDoorSlam} from "../audio";
import {Channel} from "../model/model";
import useSocket from "../useSocket";

export default function ChannelList({ addChannel, openCreateChannel, openCreateChat, openSearchChannels }: {
	addChannel: { (channelID: string, ring?: boolean) },
	openCreateChannel: { () },
	openCreateChat: { () },
	openSearchChannels: { () },
}) {
	const dispatch = useAppDispatch()
	const user = useAppSelector(state => state.user.user)
	const channels = useAppSelector(state => state.channels)

	function signOff() {
		dispatch(destroySession())
			.unwrap()
			.then(playDoorSlam)
	}

	function removeChannel(channelID) {
		dispatch(destroyChannel(channelID))
	}

	useEffect(() => { dispatch(fetchChannels()) }, [])

	useSocket('ChannelList', 'api/v1/channels/chats/messages', (channel: Channel) => {
		addChannel(channel.channelID, true)
	})

	const privateChannels = _.filter(channels, 'private')
	const publicChannels = _.reject(channels, 'private')

	return user && (
		<div className="window p-1 flex flex-col w-fit" style={{ height: 800 }}>
			<div className="draggable-handle">
				<TitleBar title="Channel List" close={signOff} />
			</div>

			<div className="px-2 py-1 font-sans flex-grow flex flex-col">
				<div className="text-sm flex flex-row justify-between">
					<p>Welcome, {user.screenname}!</p>
					<a className="link" onClick={signOff}>Sign Off</a>
				</div>

				<div className="hr mb-1"></div>

				<div className="bg-logo-tile flex flex-row items-center py-3 px-10 space-x-2">
					<div className="">
						<img src="/static/img/logo.png" alt="Slink logo" className="w-24 h-auto" />
					</div>

					<p className="flex flex-col leading-tight">
						<span className="text-white">Slink</span>
						<span className="text-white italic">Instant</span>
						<span className="text-white italic">Messenger</span>
					</p>
				</div>

				<div className="hr my-0.5"/>

				<div className="bg-white inset px-2 py-1 text-sm flex-grow h-0 overflow-y-scroll">
					<div>
						<div className="p-1 border-b border-black flex flex-row justify-between">
							<p>Chats</p>

							<div>
								<a className="link" onClick={openCreateChat}>Create</a>
							</div>
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
				</div>
			</div>
		</div>
	)
}
