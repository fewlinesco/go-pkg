package eventing

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
)

// ConsumerScheduler describes how the events which need to be consumed
// will be handled
type ConsumerScheduler struct {
	PollingInterval time.Duration
	DispatchTimeout time.Duration
	BatchSize       int
	Handlers        map[string]Handler

	Identifier string
	DB         *sqlx.DB
	Logger     *log.Logger
	shutdown   chan bool
	stopped    chan bool
}

var NoMatchingHandlerError = errors.New("no handler matching this event")

// Shutdown gracefully stop the event consumer
func (c *ConsumerScheduler) Shutdown() {
	<-c.stopped
}

// Start starts the consumer scheduler which will poll the DB at certain intervals
// for new queued events which needs to be consumed
func (c *ConsumerScheduler) Start() error {
	ticker := time.NewTicker(c.PollingInterval)
	done := make(chan error)

	go func() {
		for {
			select {
			case <-c.shutdown:
				c.stopped <- true
				return

			case <-ticker.C:
				start := time.Now()

				log := func(eventid string, message string) {
					c.Logger.Printf(`duration="%s" eventid="%s" message="%s"`, time.Since(start), eventid, message)
				}

				ctx, cancel := context.WithTimeout(context.Background(), c.DispatchTimeout)
				defer cancel()

				evs, err := ScheduleNextEventsToConsume(ctx, c.DB, c.Identifier)
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

						handler, ok := c.Handlers[ev.EventType]
						if !ok {
							err := fmt.Errorf("no handler matching this event: %v", NoMatchingHandlerError)
							log(ev.ID, err.Error())
							if _, err := MarkConsumerEventAsDiscarded(ctx, c.DB, ev); err != nil {
								log(ev.ID, fmt.Sprintf("could not mark consumer event as failed: %v", err))
							}
							return
						}

						err := handler(ctx, ev)
						if err != nil {
							log(ev.ID, fmt.Sprintf("handling failed: %v", err))

							_, err := MarkConsumerEventAsFailed(ctx, c.DB, ev, err.Error())
							if err != nil {
								log(ev.ID, fmt.Sprintf("could not mark consumer event as failed: %v", err))
							}
						}

						_, err = MarkConsumedEventAsProcessed(ctx, c.DB, ev)
						if err != nil {
							log(ev.ID, fmt.Sprintf("could not mark consumed event as processed %v", err))
						}

						log(ev.ID, "event consumed")
					}(ctx, ev)
				}

				wg.Wait()
			}
		}
	}()

	return <-done
}
