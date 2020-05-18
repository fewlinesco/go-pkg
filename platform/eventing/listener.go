package eventing

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/cloudevents/sdk-go/v2/client"
	"github.com/cloudevents/sdk-go/v2/event"
	cloudeventsnats "github.com/cloudevents/sdk-go/v2/protocol/nats"
	"github.com/jmoiron/sqlx"
)

type listenerSubject string

// Listener defines what a nats listener looks like
type Listener struct {
	config   Config
	subjects []string
	db       *sqlx.DB
	logger   *log.Logger
}

// NewListener creates a new config for a listener
func NewListener(config Config, subjects []string, db *sqlx.DB, logger *log.Logger) *Listener {
	return &Listener{
		config:   config,
		subjects: subjects,
		db:       db,
		logger:   logger,
	}
}

// Start starts a consumer for the provided subjects
func (listener *Listener) Start() {
	for _, subject := range listener.subjects {
		ctx := context.WithValue(context.Background(), listenerSubject(subject), subject)

		natsConsumer, err := cloudeventsnats.NewConsumer(listener.config.URL, subject, cloudeventsnats.NatsOptions())
		if err != nil {
			listener.logger.Printf("failed to create nats consumer, %v", err)
		}

		natsClient, err := client.New(natsConsumer)
		if err != nil {
			listener.logger.Printf("failed to create client, %v", err)
		}

		listener.logger.Printf("Consumer started for: %s", subject)

		go func(ctx context.Context) {
			for {
				if err := natsClient.StartReceiver(ctx, func(ctx context.Context, ev event.Event) error {
					start := time.Now()
					log := func(eventid string, message string) {
						listener.logger.Printf(`duration=%s eventid="%s" message="%s"`, time.Since(start), eventid, message)
					}

					if _, err := CreateConsumerEvent(ctx, listener.db, ev.Subject(), ev.Type(), ev.DataSchema(), ev.Data()); err != nil {
						log(ev.ID(), fmt.Sprintf("can't queue event: %v", err))
					}

					log(ev.ID(), "Event queued")

					return nil
				}); err != nil {
					listener.logger.Printf("failed to start nats receiver, %s", err.Error())
				}
			}
		}(ctx)
	}
}
