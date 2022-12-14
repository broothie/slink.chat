import * as React from "react";
import { useRef, useState } from "react";
import { Link } from "react-router-dom";
import TitleBar from "./TitleBar";
import * as _ from "lodash";

type Submit = { (screenname: string, password: string) }

export default function AuthWindow({ title, swapText, swapLink, submit, messages }: {
	title: string,
	swapText: string,
	swapLink: string,
	submit: Submit,
	messages: string[],
}) {
	const [screenname, setScreenname] = useState('')
	const [password, setPassword] = useState('')

	const formRef = useRef()

	function submitForm(event) {
		event.preventDefault()

		if (!!formRef.current) {
			const form = formRef.current as HTMLFormElement
			if (form?.checkValidity()) {
				try {
					submit(screenname, password)
				} catch (error) {
					console.error('error!!!', error)
				}
			}
		}
	}

	return (
		<div className="p-1 flex flex-col self-center mx-auto window draggable">
			<div className="draggable-handle">
				<TitleBar title={title} />
			</div>

			<div className="px-2 py-1 font-sans">
				<div className="bg-logo-tile flex flex-col items-center p-2">
					<div className="p-5">
						<img src="/static/img/logo.png" alt="Slink logo" className="w-48 h-auto" />
					</div>
					<p className="text-xl font-bold text-white">Slink Instant Messenger</p>
					<p className="text-xs text-white">Inspired by AIM and Windows 95</p>
				</div>

				<div className="hr my-2"></div>

				<form ref={formRef} className="space-y-2" onSubmit={submitForm}>
					<div className="flex flex-col">
						<label htmlFor="screenname" className="text-sm italic font-bold text-emphasis-text">ScreenName</label>
						<input
							id="screenname"
							type="text"
							className="input"
							minLength={6}
							required={true}
							autoFocus={true}
							value={screenname}
							onChange={e => setScreenname(e.target.value)}
						/>

						<div>
							<Link to={swapLink} className="link text-sm">{swapText}</Link>
						</div>
					</div>

					<div className="flex flex-col">
						<label htmlFor="password" className="text-sm font-bold">Password</label>
						<input
							id="password"
							type="password"
							className="input"
							minLength={6}
							required={true}
							value={password}
							onChange={e => setPassword(e.target.value)}
						/>
					</div>

					{!_.isEmpty(messages) && (
						<div className="text-sm">
							{messages.map((message, index) => (
								<p key={index} className="text-rose-600">{message}</p>
							))}
						</div>
					)}

					<div className="flex flex-row justify-between items-center py-1">
						<a href="https://github.com/broothie/slink.chat" className="text-sm link">GitHub</a>
						<button
							type="submit"
							className="button py-0.5 px-1 text-sm"
							disabled={screenname === '' || password.length < 6}
						>
							{title}
						</button>
					</div>
				</form>
			</div>
		</div>
	)
}
