package eventing

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/cloudevents/sdk-go/v2/client"
	"github.com/cloudevents/sdk-go/v2/event"
	cloudeventsnats "github.com/cloudevents/sdk-go/v2/protocol/nats"
	"github.com/fewlinesco/go-pkg/platform/monitoring"
	"github.com/jmoiron/sqlx"
)

var (
	ErrConsumerEventCanNotBePersisted = errors.New("the event could not be persisted in the database")
)

type listenerSubject string

// Listener subscribes to subjects/queues in the broker and persists them in a local database
type Listener struct {
	URL      string
	subjects []string
	db       *sqlx.DB
	logger   *log.Logger

	maxNumberOfRetries int
}

// NewListener creates a new config for a listener
func NewListener(URL string, subjects []string, db *sqlx.DB, logger *log.Logger) *Listener {
	return &Listener{
		URL:      URL,
		subjects: subjects,
		db:       db,
		logger:   logger,

		maxNumberOfRetries: 5,
	}
}

// Start starts a listener for the provided subjects
// It will save all events for the subjects in the DB
func (listener *Listener) Start() {
	for _, subject := range listener.subjects {
		ctx := context.WithValue(context.Background(), listenerSubject(subject), subject)

		natsConsumer, err := cloudeventsnats.NewConsumer(listener.URL, subject, cloudeventsnats.NatsOptions())
		if err != nil {
			listener.logger.Printf("failed to create nats consumer, %v", err)
			os.Exit(1)
		}

		natsClient, err := client.New(natsConsumer)
		if err != nil {
			listener.logger.Printf("failed to create client, %v", err)
			os.Exit(1)
		}

		listener.logger.Printf("consumer started for: %s", subject)

		go func(ctx context.Context, listener *Listener, client client.Client, subject string) {
			log := func(eventid string, message string, start time.Time) {
				listener.logger.Printf(`duration=%s eventid="%s" message="%s"`, time.Since(start), eventid, message)
			}

			for i := 0; i < listener.maxNumberOfRetries; i++ {
				if err := natsClient.StartReceiver(ctx, func(ctx context.Context, ev event.Event) error {
					start := time.Now()

					if _, err := CreateConsumerEvent(ctx, listener.db, ev.Subject(), ev.Type(), ev.DataSchema(), ev.Data()); err != nil {
						monitoring.CaptureException(err).SetLevel(monitoring.LogLevels.Error).AddTag("event", ev.String()).Log()

						return fmt.Errorf("%w: %v", ErrConsumerEventCanNotBePersisted, err)
					}

					log(ev.ID(), "event queued", start)

					return nil
				}); err != nil {
					if !errors.Is(err, ErrConsumerEventCanNotBePersisted) {
						monitoring.CaptureException(err).SetLevel(monitoring.LogLevels.Error).Log()
					}

					listener.logger.Printf("nats receiver failed, %s", err)
				}
			}

			listener.logger.Printf("nats receiver for: %s has failed to start too many times", subject)
			os.Exit(1)
		}(ctx, listener, natsClient, subject)
	}
}
