package eventing

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx/types"
	nats "github.com/nats-io/nats.go"

	"github.com/fewlinesco/go-pkg/platform/database"
)

var (
	// ErrNoEventsToSchedule is sent when no events to schedule exists in database
	ErrNoEventsToSchedule = errors.New("no events to schedule")
)

// Config defines the properties needed to publish and consume events
type Config struct {
	URL string `json:"url"`
}

// DefaultConfig defines the default properties for publishing and consuming events
var DefaultConfig = Config{
	URL: nats.DefaultURL,
}

type eventStatus string

// Possible event status
const (
	EventStatusQueued    eventStatus = "queued"
	EventStatusScheduled             = "scheduled"
	EventStatusFailed                = "failed"
	EventStatusProcessed             = "processed"
	EventStatusDiscarded             = "discarded"
)

// Event stores all the information required in order to dispatch an event to the Broker
type Event struct {
	ID           string         `db:"id"`
	Worker       *string        `db:"worker"`
	Status       eventStatus    `db:"status"`
	Subject      string         `db:"subject"`
	EventType    string         `db:"event_type"`
	DataSchema   string         `db:"dataschema"`
	Data         types.JSONText `db:"data"`
	DispatchedAt time.Time      `db:"dispatched_at"`
	ScheduledAt  *time.Time     `db:"scheduled_at"`
	FinishedAt   *time.Time     `db:"finished_at"`
	Error        *string        `db:"error"`
}

// CreatePublisherEvent creates a new events that we'll store inside the publisher_events table.
// subject: the resource bound to the event (e.g current user id, etc...)
// eventtype: is the name of the event (e.g `application.created`)
// dataschema: is the JSON-Schema ID of the event (e.g. https://github.com/fewlinesco/myapp/jsonschema/application.created.json)
// data: is the payload of the event itself
func CreatePublisherEvent(ctx context.Context, tx *database.Tx, subject string, eventtype string, dataschema string, data interface{}) (Event, error) {
	rawdata, err := json.Marshal(data)
	if err != nil {
		return Event{}, fmt.Errorf("can't marshal event: %w", err)
	}

	ev := Event{
		ID:           uuid.New().String(),
		Status:       EventStatusQueued,
		Subject:      subject,
		DataSchema:   dataschema,
		EventType:    eventtype,
		Data:         types.JSONText(rawdata),
		DispatchedAt: time.Now(),
	}

	_, err = tx.NamedExecContext(ctx, `
		INSERT INTO publisher_events
		(id, status, subject, event_type, dataschema, data, dispatched_at)
		VALUES
		(:id, :status, :subject, :event_type, :dataschema, :data, :dispatched_at)
	`, ev)

	if err != nil {
		return ev, fmt.Errorf("can't insert: %w", err)
	}

	return ev, nil
}

// CreateConsumerEvent creates a new events that we'll store inside the consumer_events table.
// subject: the resource bound to the event (e.g current user id, etc...)
// eventtype: is the name of the event (e.g `application.created`)
// dataschema: is the JSON-Schema ID of the event (e.g. https://github.com/fewlinesco/myapp/jsonschema/application.created.json)
// data: is the payload of the event itself
func CreateConsumerEvent(ctx context.Context, db *database.DB, subject string, eventtype string, dataschema string, data []byte) (Event, error) {
	ev := Event{
		ID:           uuid.New().String(),
		Status:       EventStatusQueued,
		Subject:      subject,
		DataSchema:   dataschema,
		EventType:    eventtype,
		Data:         types.JSONText(data),
		DispatchedAt: time.Now(),
	}

	_, err := db.NamedExecContext(ctx, `
		INSERT INTO consumer_events
		(id, status, subject, event_type, dataschema, data, dispatched_at)
		VALUES
		(:id, :status, :subject, :event_type, :dataschema, :data, :dispatched_at)
	`, ev)

	if err != nil {
		return ev, fmt.Errorf("can't insert: %w", err)
	}

	return ev, nil
}

// ScheduleNextEventsToConsume find the next events to consumed, mark them as "scheduled" and send them back.
// It's done in a transaction to ensure the event is also marked as "scheduled" for other workers
func ScheduleNextEventsToConsume(ctx context.Context, db *database.DB, workerName string) ([]Event, error) {
	var evs []Event

	err := db.SelectContext(ctx, &evs, `
		UPDATE consumer_events
		SET status = $1,
			scheduled_at = NOW(),
			worker = $2
		WHERE id IN (
			SELECT consumer_events.id
			FROM consumer_events
			WHERE status = 'queued'
			ORDER BY dispatched_at ASC
			FOR UPDATE SKIP LOCKED
			LIMIT $3
		)
		RETURNING *
	`, EventStatusScheduled, workerName, 100)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return evs, fmt.Errorf("%w: %v", ErrNoEventsToSchedule, err)
		}

		return nil, fmt.Errorf("can't select for udpdate: %v", err)
	}

	return evs, nil
}

// ScheduleNextEventsToPublish find the next events to process, mark them as "scheduled" and send them back.
// It's done in a transaction to ensure the event is also marked as "scheduled" for other workers
func ScheduleNextEventsToPublish(ctx context.Context, db *database.DB, workerName string) ([]Event, error) {
	var evs []Event

	err := db.SelectContext(ctx, &evs, `
		UPDATE publisher_events
		SET status = $1,
			scheduled_at = NOW(),
			worker = $2
		WHERE id IN (
			SELECT publisher_events.id
			FROM publisher_events
			WHERE status = 'queued'
			ORDER BY dispatched_at ASC
			FOR UPDATE SKIP LOCKED
			LIMIT $3
		)
		RETURNING *
	`, EventStatusScheduled, workerName, 100)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return evs, fmt.Errorf("%w: %v", ErrNoEventsToSchedule, err)
		}

		return nil, fmt.Errorf("can't select for udpdate: %v", err)
	}

	return evs, nil
}

// MarkConsumerEventAsFailed logs the failure and the timestamp. It returns the new updated event
func MarkConsumerEventAsFailed(ctx context.Context, db *database.DB, ev Event, reason string) (Event, error) {
	ev.Status = EventStatusFailed
	ev.Error = &reason
	now := time.Now()
	ev.FinishedAt = &now

	if _, err := db.NamedExecContext(ctx, "UPDATE consumer_events SET status = :status, finished_at = :finished_at, error = :error WHERE id = :id", ev); err != nil {
		return ev, fmt.Errorf("can't update: %v", err)
	}

	return ev, nil
}

// MarkPublisherEventAsFailed logs the failure and the timestamp. It returns the new updated event
func MarkPublisherEventAsFailed(ctx context.Context, db *database.DB, ev Event, reason string) (Event, error) {
	ev.Status = EventStatusFailed
	ev.Error = &reason
	now := time.Now()
	ev.FinishedAt = &now

	if _, err := db.NamedExecContext(ctx, "UPDATE publisher_events SET status = :status, finished_at = :finished_at, error = :error WHERE id = :id", ev); err != nil {
		return ev, fmt.Errorf("can't update: %v", err)
	}

	return ev, nil
}

// ReenqueWorkerPublisherEvents changes all event status to make them ready to be picked-up again
func ReenqueWorkerPublisherEvents(ctx context.Context, db *database.DB, workerName string) error {
	if _, err := db.ExecContext(ctx, "UPDATE publisher_events SET status = $1 WHERE worker = $2 AND status != 'processed'", EventStatusQueued, workerName); err != nil {
		return fmt.Errorf("can't re-enqueue worker's publisher events: %v", err)
	}

	return nil
}

// ReenqueWorkerConsumerEvents changes all event status to make them ready to be picked-up again
func ReenqueWorkerConsumerEvents(ctx context.Context, db *database.DB, workerName string) error {
	if _, err := db.ExecContext(ctx, "UPDATE consumer_events SET status = $1 WHERE worker = $2", EventStatusQueued, workerName); err != nil {
		return fmt.Errorf("can't re-enqueue worker's consumer events: %v", err)
	}

	return nil
}

// ReenquePublisherEvent changes the event status to make it ready to be picked-up again
func ReenquePublisherEvent(ctx context.Context, db *database.DB, ev Event) error {
	ev.Status = EventStatusQueued

	if _, err := db.NamedExecContext(ctx, "UPDATE publisher_events SET status = :status WHERE id = :id", ev); err != nil {
		return fmt.Errorf("can't update: %v", err)
	}

	return nil
}

// ReenqueConsumerEvent changes the event status to make it ready to be picked-up again
func ReenqueConsumerEvent(ctx context.Context, db *database.DB, ev Event) error {
	ev.Status = EventStatusQueued

	if _, err := db.NamedExecContext(ctx, "UPDATE consumer_events SET status = :status WHERE id = :id", ev); err != nil {
		return fmt.Errorf("can't update: %v", err)
	}

	return nil
}

// MarkPublishedEventAsProcessed It returns the new updated event
func MarkPublishedEventAsProcessed(ctx context.Context, db *database.DB, ev Event) (Event, error) {
	ev.Status = EventStatusProcessed
	now := time.Now()
	ev.FinishedAt = &now

	if _, err := db.NamedExecContext(ctx, "UPDATE publisher_events SET status = :status, finished_at = :finished_at WHERE id = :id", ev); err != nil {
		return ev, fmt.Errorf("can't update: %v", err)
	}

	return ev, nil
}

// MarkConsumedEventAsProcessed It returns the new updated event
func MarkConsumedEventAsProcessed(ctx context.Context, db *database.DB, ev Event) (Event, error) {
	ev.Status = EventStatusProcessed
	now := time.Now()
	ev.FinishedAt = &now

	if _, err := db.NamedExecContext(ctx, "UPDATE consumer_events SET status = :status, finished_at = :finished_at WHERE id = :id", ev); err != nil {
		return ev, fmt.Errorf("can't update: %v", err)
	}

	return ev, nil
}

// MarkConsumerEventAsDiscarded will mark a consumer event as discarded
// This is the case if a certain event type does not have a handler registered for it
func MarkConsumerEventAsDiscarded(ctx context.Context, db *database.DB, ev Event) (Event, error) {
	ev.Status = EventStatusDiscarded
	now := time.Now()
	ev.FinishedAt = &now

	if _, err := db.NamedExecContext(ctx, "UPDATE consumer_events SET status = :status, finished_at = :finished_at WHERE id = :id", ev); err != nil {
		return ev, fmt.Errorf("can't mark event as discarded: %v", err)
	}

	return ev, nil
}
