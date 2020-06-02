package eventing

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/cloudevents/sdk-go/v2/client"
	cloudeventsnats "github.com/cloudevents/sdk-go/v2/protocol/nats"
	"github.com/fewlinesco/go-pkg/platform/database"
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

// Handler defines what an event handler looks like
type Handler func(context.Context, Event) error

// NewNatsConsumer initializes the settings needed for a new Nats consumer
func NewNatsConsumer(URL string, identifier string, db *database.DB, logger *log.Logger) *ConsumerScheduler {
	return &ConsumerScheduler{
		PollingInterval: 500 * time.Millisecond,
		DispatchTimeout: 400 * time.Millisecond,
		BatchSize:       150,

		Handlers:   make(map[string]Handler),
		Identifier: identifier,
		DB:         db,
		Logger:     logger,
		shutdown:   make(chan bool, 1),
		stopped:    make(chan bool, 1),
	}
}

// HandleEvent will register any handlers for event
func (c *ConsumerScheduler) HandleEvent(eventType string, handler Handler) {
	c.Handlers[eventType] = handler
}
