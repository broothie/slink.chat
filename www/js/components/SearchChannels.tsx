import * as React from 'react'
import TitleBar from './TitleBar'
import { useState, useEffect } from "react";
import { Channel } from '../model/model';
import axios from '../axios';
import { useAppDispatch } from "../hooks";
import { fetchChannels } from '../store/channelsSlice';

export default function SearchChannels({ addChannel, close }: {
  addChannel: { (channelID: string) },
  close: { () },
}) {
  const [query, setQuery] = useState('')
  const [channels, setChannels] = useState([] as Channel[])

  const dispatch = useAppDispatch()

  function channelClicked(channelID: string) {
    axios.post(`/api/v1/channels/${channelID}/subscriptions`)
      .then(() => {
        dispatch(fetchChannels())
        addChannel(channelID)
        close()
      })
  }

  useEffect(() => {
    axios.get(`/api/v1/channels/search?query=${query}`)
      .then(({ data }: { data: { channels: Channel[] } }) => {
        setChannels(data.channels)
      })
  }, [query])

  return (
    <div className="window flex flex-col w-80 p-1">
      <div className="draggable-handle">
        <TitleBar title="Search Channels" close={close} />
      </div>

      <div className="flex flex-col p-1 font-sans space-y-1">
        <div className="flex flex-row space-x-1">
          <input
            type="text"
            placeholder="Search for channels..."
            className="flex-grow input p-1"
            autoFocus={true}
            value={query}
            onChange={e => setQuery(e.target.value)}
          />
        </div>

        <div className="hr"></div>

        <div className="bg-white inset p-1 text-sm h-52">
          {channels.map(channel => (
            <div key={channel.channelID} className="p-1 hover:bg-logo-tile hover:text-white flex flex-row justify-between">
              <p>{channel.name}</p>
              <button
                className="button px-2"
                onClick={() => channelClicked(channel.channelID)}
              >
                +
              </button>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
