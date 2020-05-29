package eventing

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/cloudevents/sdk-go/v2/client"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/cloudevents/sdk-go/v2/protocol"
	cloudeventsnats "github.com/cloudevents/sdk-go/v2/protocol/nats"
	"github.com/fewlinesco/go-pkg/platform/database"
	"github.com/fewlinesco/go-pkg/platform/monitoring"
	"go.opencensus.io/trace"
)

var (
	// ErrConsumerEventCanNotBePersisted is an error indicatin that an incoming event could not be persisted in the database
	ErrConsumerEventCanNotBePersisted = errors.New("the event could not be persisted in the database")
)

type listenerSubject string

// Listener subscribes to subjects/queues in the broker and persists them in a local database
type Listener struct {
	URL      string
	subjects []string
	db       *database.DB
	logger   *log.Logger

	stop               chan bool
	maxNumberOfRetries int
}

// NewListener creates a new config for a listener
func NewListener(URL string, subjects []string, db *database.DB, logger *log.Logger) *Listener {
	return &Listener{
		URL:      URL,
		subjects: subjects,
		db:       db,
		logger:   logger,

		stop:               make(chan bool),
		maxNumberOfRetries: 5,
	}
}

// Stop is a function which can schedule all the listeners to stop
func (listener *Listener) Stop() {
	listener.stop <- true
}

// Start starts a listener for the provided subjects
// It will save all events for the subjects in the DB
func (listener *Listener) Start() error {
	shutdown := make(chan error)
	ctx, cancel := context.WithCancel(context.Background())

	for _, subject := range listener.subjects {
		go func(ctx context.Context) {
			ctx = context.WithValue(ctx, listenerSubject(subject), subject)

			listener.logger.Printf("consumer started for: %s", subject)

			if err := startReceiver(ctx, listener, subject); err != nil {
				shutdown <- fmt.Errorf("receiver '%s': %v", subject, err)
			}
		}(ctx)
	}

	select {
	case <-listener.stop:
		cancel()
		return nil
	case err := <-shutdown:
		cancel()
		return fmt.Errorf("all listeners stopped: %v", err)
	}
}

func startReceiver(ctx context.Context, listener *Listener, subject string) error {
	natsConsumer, err := cloudeventsnats.NewConsumer(listener.URL, subject, cloudeventsnats.NatsOptions())
	if err != nil {
		return fmt.Errorf("failed to create nats consumer, %v", err)
	}

	natsClient, err := client.New(natsConsumer)
	if err != nil {
		return fmt.Errorf("failed to create client, %v", err)
	}

	go func(ctx context.Context) {
		<-ctx.Done()
		natsConsumer.Close(ctx)
	}(ctx)

	// do we need to keep a timeframe?
	for i := 0; i < listener.maxNumberOfRetries; i++ {
		select {
		case <-ctx.Done():
			return nil
		default:
			if err := natsClient.StartReceiver(ctx, receiverHandler(listener.db)); err != nil {
				monitoring.CaptureException(err).SetLevel(monitoring.LogLevels.Error).Log()

				listener.logger.Printf("nats receiver failed (%d/%d), %s", (i + 1), listener.maxNumberOfRetries, err)
			}
		}
	}

	return fmt.Errorf("failed to start too many times")
}

func receiverHandler(db *database.DB) func(context.Context, event.Event) protocol.Result {
	return func(ctx context.Context, ev event.Event) protocol.Result {
		ctx, span := trace.StartSpan(ctx, "eventing.EventReceived")
		defer span.End()

		if _, err := CreateConsumerEvent(ctx, db, ev.Subject(), ev.Type(), ev.DataSchema(), ev.Data()); err != nil {
			errorAttribute := trace.StringAttribute("error", err.Error())
			span.AddAttributes(errorAttribute)

			return protocol.NewResult("%w: %v", ErrConsumerEventCanNotBePersisted, err)
		}

		return nil
	}
}
