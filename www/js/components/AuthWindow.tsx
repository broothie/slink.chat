import * as React from "react";
import {useState} from "react";
import {Link} from "react-router-dom";

type Submit = {
	(screenname: string, password: string)
}

export default function AuthWindow({ title, swapText, swapLink, submit }: {
	title: string,
	swapText: string,
	swapLink: string,
  submit: Submit,
}) {
	const [screenname, setScreenname] = useState('')
	const [password, setPassword] = useState('')

	return (
		<div className="p-1 flex flex-col self-center mx-auto window">
			<div className="w-full title-bar"><p className="px-0.5 text-sm">{title}</p></div>

			<div className="px-2 py-1 font-sans">
				<div className="bg-logo-tile flex flex-col items-center p-2">
					<div className="p-5">
						<img src="/static/logo.png" alt="Slink logo" className="w-48 h-auto"/>
					</div>
					<p className="text-xl font-bold text-white">Slink Instant Messenger</p>
					<p className="text-xs text-white">Inspired by AIM and Windows 95</p>
				</div>

				<div className="hr my-2"></div>

				<div className="space-y-2">
					<div className="flex flex-col">
						<label htmlFor="screenname" className="text-sm italic font-bold text-emphasis-text">ScreenName</label>
						<input
							id="screenname"
							type="text"
							className="input"
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
							required={true}
							value={password}
							onChange={e => setPassword(e.target.value)}
						/>
					</div>

					<div className="flex flex-row justify-between items-center py-1">
						<a href="https://github.com/broothie/slink.chat" className="text-sm link">GitHub</a>
						<button className="button py-0.5 px-1 text-sm" onClick={() => submit(screenname, password)}>
							{title}
						</button>
					</div>
				</div>
			</div>
		</div>
	)
}
