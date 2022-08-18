
document.addEventListener('mousedown', event => {
	const node = event.target as HTMLElement

	if (!node.matches('.draggable-handle *')) return
	let currentNode = node
	while (!currentNode.matches('.draggable')) {
		currentNode = currentNode.parentNode as HTMLElement
		if (!currentNode) return
	}

	const draggable = currentNode
	const rect = draggable.getBoundingClientRect()

	const mouseStart = { x: event.clientX, y: event.clientY }
	const draggableStart = { left: rect.left, right: rect.right, top: rect.top, bottom: rect.bottom }

	function onMouseMove(event) {
		draggable.style.position = 'absolute'
		draggable.style.left = `${draggableStart.left + (event.clientX - mouseStart.x)}px`
		draggable.style.top = `${draggableStart.top + (event.clientY - mouseStart.y)}px`
		draggable.style.right = null
		draggable.style.bottom = null
	}

	document.addEventListener('mousemove', onMouseMove)
	draggable.addEventListener('mouseup', () => {
		document.removeEventListener('mousemove', onMouseMove)
	})
})
