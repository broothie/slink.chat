import axios from "axios";

export default axios.create({
	headers: { 'X-Csrf-Token': document.getElementsByName('gorilla.csrf.Token')[0].getAttribute('content') }
})
