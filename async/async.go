package async

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"cloud.google.com/go/pubsub"
	"github.com/broothie/qst"
	"github.com/broothie/slink.chat/config"
	"github.com/broothie/slink.chat/util"
	"github.com/pkg/errors"
)

type Job interface {
	Name() string
}

type Async struct {
	config *config.Config
	pubsub *pubsub.Client
}

type Message struct {
	Name    string
	Payload []byte
}

func New(cfg *config.Config) (*Async, error) {
	pubsubClient, err := pubsub.NewClient(context.Background(), cfg.ProjectID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create pubsub client")
	}

	return &Async{
		config: cfg,
		pubsub: pubsubClient,
	}, nil
}

func (a *Async) Do(ctx context.Context, job Job) error {
	payload, err := json.Marshal(job)
	if err != nil {
		return errors.Wrap(err, "failed to marshal job")
	}

	data, err := json.Marshal(Message{Name: job.Name(), Payload: payload})
	if err != nil {
		return errors.Wrap(err, "failed to marshal message")
	}

	if a.config.AsyncTopic == "" {
		if err := a.sendToLocalJobServer(job.Name(), data); err != nil {
			return errors.Wrap(err, "failed to publish message to local job server")
		}

		return nil
	}

	if _, err := a.pubsub.Topic(a.config.AsyncTopic).Publish(ctx, &pubsub.Message{Data: data}).Get(ctx); err != nil {
		return errors.Wrap(err, "failed to publish message")
	}

	return nil
}

func (a *Async) sendToLocalJobServer(name string, data []byte) error {
	jobServerURL := os.Getenv("JOB_SERVER_URL")
	if jobServerURL == "" {
		fmt.Println("skipping job", name)
		return nil
	}

	response, err := qst.Post(jobServerURL, qst.BodyJSON(util.Map{"message": util.Map{"data": data}}))
	if err != nil {
		return errors.Wrap(err, "failed to send message")
	}

	if !strings.HasPrefix(strconv.Itoa(response.StatusCode), "2") {
		body, err := io.ReadAll(response.Body)
		if err != nil {
			return errors.Wrap(err, "failed to read body")
		}

		return errors.Wrapf(err, "%d bad response: %s", response.StatusCode, string(body))
	}

	return nil
}
