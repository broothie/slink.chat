import * as React from 'react'
import {useAppDispatch, useAppSelector} from "../hooks";
import {destroySession} from "../store/userSlice";

export default function ChannelList() {
	const user = useAppSelector(state => state.user.user)
	const dispatch = useAppDispatch()

	function signOff() {
		dispatch(destroySession())
	}

	return (
		<div className="window p-1 flex flex-col w-fit absolute top-10 right-10">
			<div className="title-bar">
				<p className="px-0.5 text-sm">Channel List</p>
			</div>

			<div className="px-2 py-1 font-sans">
				<div className="text-sm flex flex-row justify-between">
					<p>Welcome, {user.screenname}!</p>
					<a className="link" onClick={signOff}>Sign Off</a>
				</div>

				<div className="hr mb-1"></div>

				<div className="bg-logo-tile flex flex-row items-center px-2">
					<div className="p-3">
						<img src="/static/logo.png" alt="Slink logo" className="w-32 h-auto"/>
					</div>
					<p className="flex flex-col leading-tight">
						<span className="text-white">Slink</span>
						<span className="text-white italic">Instant</span>
						<span className="text-white italic">Messenger</span>
					</p>
				</div>

				<div className="hr my-0.5"></div>
			</div>
		</div>
	)
}
