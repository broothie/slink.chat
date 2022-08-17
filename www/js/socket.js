const channelID = 'cbu7ljlvqc7sc7sk95j0'

document.addEventListener('DOMContentLoaded', () => {
  const socket = new WebSocket(`ws://${location.host}/api/v1/channels/${channelID}/subscribe`)

  socket.onopen = event => {
    console.log('open', {event})
  }

  socket.onmessage = event => {
    console.log('message', {event})
    const message = JSON.parse(event.data)

    const p = document.createElement('p')
    p.innerText = message.body

    document.getElementById('messages').appendChild(p)
  }

  window.addEventListener('beforeunload', () => {
    socket.close()
  })

  document.getElementById('button').addEventListener('click', () => {
    const input = document.getElementById('input')
    const value = input.value

    fetch(`/api/v1/channels/${channelID}/messages`, {
      method: 'POST',
      body: JSON.stringify({body: value}),
    })
      .then(() => input.value = '')
      .catch(console.error)
  })
})
