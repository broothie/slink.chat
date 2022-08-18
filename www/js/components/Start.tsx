import * as React from "react";
import ChannelList from "./ChannelList";
import {useState} from "react";
import Chat from "./Chat";
import {useAppSelector} from "../hooks";
import * as _ from "lodash";
import {Channel} from "../model/model";

export type AddChannel = { (channelID: string) }

export default function Start() {
	const channels = useAppSelector(state => state.channels)
	const [openChannels, setOpenChannels] = useState([] as Channel[])

	function addChannel(channelID: string) {
		if (_.includes(openChannels, channels[channelID])) return

		setOpenChannels(openChannels.concat([channels[channelID]]))
	}

	function removeChannel(channelID: string) {
		setOpenChannels(_.without(openChannels, _.find(openChannels, { channelID })))
	}

	return (
		<div className="w-full h-full relative">
			<ChannelList addChannel={addChannel}/>

			<div>
				{_.map(openChannels, channel => (
					<Chat
						key={channel.channelID}
						channelID={channel.channelID}
						offset={_.size(openChannels)}
						close={() => removeChannel(channel.channelID)}
					/>
				))}
			</div>
		</div>
	)
}
