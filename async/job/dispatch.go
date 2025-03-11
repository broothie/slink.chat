package job

import (
	"context"
	"encoding/json"
	"reflect"

	"github.com/broothie/slink.chat/async"
	"github.com/pkg/errors"
)

func (s *Server) dispatch(ctx context.Context, message async.Message) error {
	switch message.Name {
	case ResetDatabase{}.Name():
		return s.ResetDatabase(ctx)

	case NewUserJob{}.Name():
		var payload NewUserJob
		if err := json.Unmarshal(message.Payload, &payload); err != nil {
			return errors.Wrap(err, "failed to unmarshal payload")
		}

		return s.NewUserJob(ctx, payload)

	case NewChannelJob{}.Name():
		var payload NewChannelJob
		if err := json.Unmarshal(message.Payload, &payload); err != nil {
			return errors.Wrap(err, "failed to unmarshal payload")
		}

		return s.NewChannelJob(ctx, payload)
	}

	return nil
}

func typeName(job any) string {
	return reflect.TypeOf(job).Name()
}
