import {useEffect} from "react";

export default function useSocket(context: string, path: string, onmessage: { (data) }) {
	useEffect(() => {
		let closedByClient = false
		let socket
		function startSocket() {
			const protocol = location.protocol === 'https:' ? 'wss' : 'ws'
			socket = new WebSocket(`${protocol}://${location.host}/${path}`)

			socket.onopen = () => { console.log(context, 'socket opened') }
			socket.onmessage = event => onmessage(JSON.parse(event.data))
			socket.onclose = event => {
				if (closedByClient) return

				console.log(context, 'server closed socket', {event})
				setTimeout(startSocket, 1000)
			}
		}

		startSocket()

		return () => {
			closedByClient = true
			socket.close()
			console.log(context, 'socket closed')
		}
	}, [])
}

