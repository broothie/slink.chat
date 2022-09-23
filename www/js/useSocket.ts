import {useEffect, useState} from "react";

export default function useSocket(context: string, path: string, onmessage: { (data: Object) }): WebSocket {
	const [closedByClient, setClosedByClient] = useState(false)
	const [socket, setSocket] = useState(null)

	useEffect(() => {
		if (!socket) {
			const protocol = location.protocol === 'https:' ? 'wss' : 'ws'
			const socket = new WebSocket(`${protocol}://${location.host}/${path}`)

			socket.onopen = () => { console.log(context, 'socket opened') }
			socket.onmessage = event => onmessage(JSON.parse(event.data))
			socket.onclose = event => {
				if (closedByClient) return

				console.log(context, 'server closed socket', {event})
				setTimeout(() => setSocket(null), 1000)
			}

			setSocket(socket)
		}

		return () => {
			setClosedByClient(true)
			if (!!socket) socket.close()
			console.log(context, 'socket closed')
		}
	}, [socket])

	return socket
}

