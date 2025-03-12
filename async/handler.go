package async

import (
	"context"
	"encoding/json"
	"net/http"
)

const googleCloudScheduleUserAgent = "Google-Cloud-Scheduler"

type Dispatch func(context.Context, Message) error

func (a *Async) Handler(dispatch Dispatch) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var message Message
		if r.UserAgent() == googleCloudScheduleUserAgent {
			if err := json.NewDecoder(r.Body).Decode(&message); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		} else {
			var pubSubMessage struct {
				Message struct {
					Data []byte `json:"data,omitempty"`
				} `json:"message"`
			}

			if err := json.NewDecoder(r.Body).Decode(&pubSubMessage); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			if err := json.Unmarshal(pubSubMessage.Message.Data, &message); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		}

		if err := dispatch(r.Context(), message); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
