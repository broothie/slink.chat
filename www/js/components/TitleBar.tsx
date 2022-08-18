import * as React from 'react'

export default function TitleBar({ title, close }: { title: string, close?: React.MouseEventHandler }) {
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
