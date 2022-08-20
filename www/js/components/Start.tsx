import * as React from "react";
import ChannelList from "./ChannelList";
import { useEffect, useState } from "react";
import Chat from "./Chat";
import * as _ from "lodash";
import CreateChannel from "./CreateChannel";
import CreateChat from "./CreateChat";
import SearchChannels from "./SearchChannels";

export default function Start() {
	const [windowIDs, setWindowIDs] = useState([])
	const [windows, setWindows] = useState({})

	function addWindow(id: string, style: Object, element: React.ReactNode) {
		setWindowIDs(windowIDs => _.uniq(windowIDs.concat([id])))
		setWindows(windows => _.merge({}, windows, { [id]: { style, element } }))
		activateWindow(id)
	}

	function removeWindow(id: string) {
		setWindowIDs(windowIDs => _.without(windowIDs, id))
		setWindows(windows => {
			const copy = _.merge({}, windows)
			delete copy.id
			return copy
		})
	}

	function activateWindow(id: string) {
		setWindowIDs(windowIDs => _.sortBy(windowIDs, windowID => windowID === id))
	}

	function addChannel(channelID: string) {
		addWindow(channelID, { left: 50, top: 50 }, (
			<Chat channelID={channelID} close={() => removeWindow(channelID)} />
		))
	}

	function openCreateChat() {
		addWindow('CreateChat', { right: 400, top: 50 }, (
			<CreateChat close={() => removeWindow('CreateChat')} addChannel={addChannel} />
		))
	}

	function openCreateChannel() {
		addWindow('CreateChannel', { right: 400, top: 50 }, (
			<CreateChannel close={() => removeWindow('CreateChannel')} addChannel={addChannel} />
		))
	}

	function openSearchChannels() {
		addWindow('SearchChannels', { right: 400, top: 50 }, (
			<SearchChannels close={() => removeWindow('SearchChannels')} addChannel={addChannel} />
		))
	}

	useEffect(() => {
		addWindow('ChannelList', { right: 50, top: 50 }, (
			<ChannelList
				addChannel={addChannel}
				openCreateChannel={openCreateChannel}
				openCreateChat={openCreateChat}
				openSearchChannels={openSearchChannels}
			/>
		))
	}, [])

	return (
		<div className="w-full h-full relative">
			{windowIDs.map(windowID => (
				<div
					key={windowID}
					className="absolute draggable"
					style={windows[windowID].style}
					onClick={() => activateWindow(windowID)}
				>
					{windows[windowID].element}
				</div>
			))}
		</div>
	)
}
