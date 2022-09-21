package async

import (
	"context"
	"encoding/json"
	"net/http"
)

type Dispatch func(context.Context, Message) error

func (a *Async) Handler(dispatch Dispatch) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var pubSubMessage struct {
			Message struct {
				Data []byte `json:"data,omitempty"`
			} `json:"message"`
		}

		if err := json.NewDecoder(r.Body).Decode(&pubSubMessage); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var message Message
		if err := json.Unmarshal(pubSubMessage.Message.Data, &message); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := dispatch(r.Context(), message); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
