
const doorOpenSound = new Audio('/static/audio/dooropen.wav')
const doorSlamSound = new Audio('/static/audio/doorslam.wav')
const messageReceiveSound = new Audio('/static/audio/imrcv.wav')
const messageSendSound = new Audio('/static/audio/imsend.wav')
const messageRingSound = new Audio('/static/audio/ring.wav')

export async function playDoorOpen() {
	return await doorOpenSound.play()
}

export async function playDoorSlam() {
	return await doorSlamSound.play()
}

export async function playMessageReceive() {
	return await messageReceiveSound.play()
}

export async function playMessageSend() {
	return await messageSendSound.play()
}

export async function playMessageRing() {
	return await messageRingSound.play()
}
