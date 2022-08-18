import * as React from 'react'

export type CloseFunction = { (): void }

export default function TitleBar({ title, close }: { title: string, close?: CloseFunction }) {
	return (
		<div className="title-bar flex flex-row justify-between h-6 px-1 py-0.5">
			<p className="text-sm">{title}</p>

			{!!close && (
				<button className="button p-0.5" onClick={close}>
					<img src="/static/x.png" alt="close button" className="h-full w-auto object-scale-down"/>
				</button>
			)}
		</div>
	)
}
