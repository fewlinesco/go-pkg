package eventing

import (
	"fmt"

	"github.com/cloudevents/sdk-go/v2/client"
	cloudeventsnats "github.com/cloudevents/sdk-go/v2/protocol/nats"
)

// NewNatsPublisher creates a new event publisher using nats.
func NewNatsPublisher(natserver string, natsubject string) (client.Client, error) {
	publisher, err := cloudeventsnats.NewSender(natserver, natsubject, cloudeventsnats.NatsOptions())
	if err != nil {
		return nil, fmt.Errorf("can't create nats publisher: %v", err)
	}

	natsClient, err := client.New(publisher)
	if err != nil {
		return nil, fmt.Errorf("can't create nats client: %v", err)
	}

	return natsClient, nil
}

// NewNatsConsumer creates a new event consumer using nats.
func NewNatsConsumer(natserver string, natsubject string) (client.Client, error) {
	consumer, err := cloudeventsnats.NewConsumer(natserver, natsubject, cloudeventsnats.NatsOptions())
	if err != nil {
		return nil, fmt.Errorf("can't create nats consumer: %v", err)
	}

	natsClient, err := client.New(consumer)
	if err != nil {
		return nil, fmt.Errorf("can't create nats client: %v", err)
	}

	return natsClient, nil
}
