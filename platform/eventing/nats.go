package eventing

import (
	"fmt"

	"github.com/cloudevents/sdk-go/v2/client"
	cloudeventsnats "github.com/cloudevents/sdk-go/v2/protocol/nats"
)

func NewNatsClient(natserver string, natsubject string) (client.Client, error) {
	p, err := cloudeventsnats.NewSender(natserver, natsubject, cloudeventsnats.NatsOptions())
	if err != nil {
		return nil, fmt.Errorf("can't create nats sender: %v", err)
	}

	c, err := client.New(p)
	if err != nil {
		return nil, fmt.Errorf("can't create nats client: %v", err)
	}

	return c, nil
}
