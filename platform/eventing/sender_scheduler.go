package eventing

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/cloudevents/sdk-go/v2/client"
	"github.com/cloudevents/sdk-go/v2/event"

	"github.com/fewlinesco/go-pkg/platform/database"
)

// SenderScheduler represents the datastructure in charge of dispatching events
type SenderScheduler struct {
	PollingInterval time.Duration
	DispatchTimeout time.Duration
	ShutdownTimeout time.Duration
	BatchSize       int

	cloudEventClient client.Client
	identifier       string
	db               *database.DB
	eventSourceName  string
	logger           *log.Logger
	err              error
	shutdown         chan bool
	stopped          chan bool
}

// NewSenderScheduler initializes a new event sender scheduler
func NewSenderScheduler(identifier string, cloudeventClient client.Client, logger *log.Logger, db *database.DB, source string) *SenderScheduler {
	return &SenderScheduler{
		BatchSize:       150,
		PollingInterval: 500 * time.Millisecond,
		DispatchTimeout: 400 * time.Millisecond,

		cloudEventClient: cloudeventClient,
		identifier:       identifier,
		logger:           logger,
		eventSourceName:  source,
		db:               db,
		shutdown:         make(chan bool, 1),
		stopped:          make(chan bool, 1),
	}
}

// Shutdown gracefully stop the event processor
func (s *SenderScheduler) Shutdown() {
	<-s.stopped
}

// Start a new goroutine to send awaiting events using CloudEvents
func (s *SenderScheduler) Start() error {
	ticker := time.NewTicker(s.PollingInterval)
	done := make(chan error)

	if err := ReenqueWorkerPublisherEvents(context.Background(), s.db, s.identifier); err != nil {
		return err
	}

	go func() {
		for {
			select {
			case <-s.shutdown:
				s.stopped <- true
				return

			case <-ticker.C:
				start := time.Now()

				log := func(eventid string, message string) {
					s.logger.Printf(`duration="%s" eventid="%s" message="%s"`, time.Since(start), eventid, message)
				}

				ctx, cancel := context.WithTimeout(context.Background(), s.DispatchTimeout)
				defer cancel()

				evs, err := ScheduleNextEventsToPublish(ctx, s.db, s.identifier)
				if err != nil {
					if errors.Is(err, ErrNoEventsToSchedule) {
						continue
					}

					log("", fmt.Sprintf("can't fetch new events: %v", err))
					continue
				}

				var wg sync.WaitGroup
				wg.Add(len(evs))

				for _, ev := range evs {
					go func(ctx context.Context, ev Event) {
						defer wg.Done()
						cloudevent := event.New()
						cloudevent.SetID(ev.ID)
						cloudevent.SetSubject(ev.Subject)
						cloudevent.SetSource(s.eventSourceName)
						cloudevent.SetTime(ev.DispatchedAt)
						cloudevent.SetType(ev.EventType)
						cloudevent.SetDataSchema(ev.DataSchema)
						if err := cloudevent.SetData("application/json", ev.Data); err != nil {
							if _, err := MarkPublisherEventAsFailed(ctx, s.db, ev, err.Error()); err != nil {
								log(ev.ID, fmt.Sprintf("can't mark event as failed: %v", err))
								return
							}
						}

						if err := s.cloudEventClient.Send(ctx, cloudevent); err != nil {
							log(ev.ID, fmt.Sprintf("re-enqueue because can't send event to the broker: %v.", err))

							if err := ReenquePublisherEvent(ctx, s.db, ev); err != nil {
								log(ev.ID, fmt.Sprintf("can't re-enqueue event: %v", err))

								return
							}

							return
						}

						if _, err := MarkPublishedEventAsProcessed(ctx, s.db, ev); err != nil {
							log(ev.ID, fmt.Sprintf("can't mark event as sent: %v", err))
							return
						}

						log(ev.ID, "event sent")
					}(ctx, ev)
				}

				wg.Wait()
			}
		}
	}()

	return <-done
}
