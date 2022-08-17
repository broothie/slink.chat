import * as React from 'react'
import * as ReactDOM from 'react-dom/client'
import Root from "./components/Root"

document.addEventListener('DOMContentLoaded', () => {
	const root = document.getElementById('root')
	ReactDOM.createRoot(root).render(<Root/>)
})
