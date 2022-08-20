import * as React from 'react'

export type CloseFunction = { (): void }

export default function TitleBar({ title, close }: { title: string, close?: CloseFunction }) {
	return (
		<div className="title-bar flex flex-row justify-between h-6 px-1 py-0.5 select-none">
			<p className="text-sm">{title}</p>

			{!!close && (
				<button className="button p-0.5" onClick={close}>
					<img src="/static/img/x.png" alt="close button" className="h-3 w-3.5 object-scale-down" />
				</button>
			)}
		</div>
	)
}
