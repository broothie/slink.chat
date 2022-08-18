import * as React from 'react'
import {useAppDispatch, useAppSelector} from "../hooks";
import {destroySession} from "../store/userSlice";
import {useEffect} from "react";
import {fetchSubscriptions} from "../store/subscriptionSlice";
import _ from 'lodash'

export default function ChannelList() {
	const dispatch = useAppDispatch()
	const user = useAppSelector(state => state.user.user)
	const subscriptions = useAppSelector(state => state.subscriptions)

	function signOff() {
		dispatch(destroySession())
	}

	useEffect(() => {
		dispatch(fetchSubscriptions())
	}, [])

	const privateSubscriptions = _.filter(_.values(subscriptions), subscription => subscription.private)
	const publicSubscriptions = _.filter(_.values(subscriptions), subscription => !subscription.private)

	return (
		<div className="window p-1 flex flex-col w-fit absolute top-10 right-10 h-5/6">
			<div className="title-bar">
				<p className="px-0.5 text-sm">Channel List</p>
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
							{_.map(privateSubscriptions, subscription => (
								<div key={subscription.id} className="pl-3 pr-0.5 py-0.5 cursor-pointer">
									<p className="hover:bg-logo-tile hover:text-white p-0.5">{subscription.name}</p>
								</div>
							))}
						</div>
					</div>

					<div>
						<div className="p-1 border-b border-black">
							<p>Channels</p>
						</div>

						<div>
							{_.map(publicSubscriptions, subscription => (
								<div key={subscription.id} className="pl-3 pr-0.5 py-0.5 cursor-pointer">
									<p className="hover:bg-logo-tile hover:text-white p-0.5">{subscription.name}</p>
								</div>
							))}
						</div>
					</div>
				</div>
			</div>
		</div>
	)
}
