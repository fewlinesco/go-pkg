package eventing

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/cloudevents/sdk-go/v2/client"
	cloudeventsnats "github.com/cloudevents/sdk-go/v2/protocol/nats"
	"github.com/jmoiron/sqlx"
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

// Consumer describes how a consumer looks like
type Consumer struct {
	Scheduler ConsumerScheduler
	Listener  Listener
}

// Handler defines what an event handler looks like
type Handler func(context.Context, Event) error

// NewNatsConsumer initializes the settings needed for a new Nats consumer
func NewNatsConsumer(URL string, identifier string, subjects []string, db *sqlx.DB, logger *log.Logger) *Consumer {
	return &Consumer{
		Scheduler: ConsumerScheduler{
			PollingInterval: 500 * time.Millisecond,
			DispatchTimeout: 400 * time.Millisecond,
			BatchSize:       150,

			Handlers:   make(map[string]Handler, 0),
			identifier: identifier,
			db:         db,
			logger:     logger,
			shutdown:   make(chan bool, 1),
			stopped:    make(chan bool, 1),
		},
		Listener: Listener{
			URL:      URL,
			subjects: subjects,
			db:       db,
			logger:   logger,
		},
	}
}

// HandleEvent will register any handlers for event
func (c *Consumer) HandleEvent(eventType string, handler Handler) {
	c.Scheduler.Handlers[eventType] = handler
}

func (c *Consumer) Start() error {
	c.Listener.Start()

	c.Scheduler.Start()
	return nil
}
