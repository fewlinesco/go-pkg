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

type ConsumerScheduler struct {
	PollingInterval time.Duration
	DispatchTimeout time.Duration
	BatchSize       int
	Handlers        map[string]Handler

	identifier string
	db         *sqlx.DB
	logger     *log.Logger
	err        error
	shutdown   chan bool
	stopped    chan bool
}

var NoMatchingHandlerError = errors.New("no handler matching this event")

// Shutdown gracefully stop the event consumer
func (c *ConsumerScheduler) Shutdown() {
	<-c.stopped
}

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
					c.logger.Printf(`duration="%s" eventid="%s" message="%s"`, time.Since(start), eventid, message)
				}

				ctx, cancel := context.WithTimeout(context.Background(), c.DispatchTimeout)
				defer cancel()

				evs, err := ScheduleNextEventsToConsume(ctx, c.db, c.identifier)
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
							if _, err := MarkConsumerEventAsFailed(ctx, c.db, ev, err.Error()); err != nil {
								log(ev.ID, fmt.Sprintf("could not mark consumer event as failed: %v", err))
							}
							return
						}

						err := handler(ctx, ev)
						if err != nil {
							log(ev.ID, fmt.Sprintf("handling failed: %v", err))

							_, err := MarkConsumerEventAsFailed(ctx, c.db, ev, err.Error())
							if err != nil {
								log(ev.ID, fmt.Sprintf("could not mark consumer event as failed: %v", err))
							}
						}

						_, err = MarkConsumedEventAsProcessed(ctx, c.db, ev)
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
