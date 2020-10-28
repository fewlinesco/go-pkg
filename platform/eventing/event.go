package eventing

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/fewlinesco/go-pkg/platform/database"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx/types"
)

// EventStatus holds all the possible states for an event
type EventStatus string

// Possible event status
const (
	EventStatusQueued    EventStatus = "queued"
	EventStatusScheduled EventStatus = "scheduled"
	EventStatusFailed    EventStatus = "failed"
	EventStatusProcessed EventStatus = "processed"
	EventStatusDiscarded EventStatus = "discarded"
)

// Event stores all the information required in order to dispatch an event to the Broker
type Event struct {
	ID           string         `db:"id"`
	Worker       *string        `db:"worker"`
	Status       EventStatus    `db:"status"`
	Subject      string         `db:"subject"`
	Type         string         `db:"type"`
	Source       string         `db:"source"`
	DataSchema   string         `db:"dataschema"`
	Data         types.JSONText `db:"data"`
	DispatchedAt time.Time      `db:"dispatched_at"`
	ScheduledAt  *time.Time     `db:"scheduled_at"`
	FinishedAt   *time.Time     `db:"finished_at"`
	Error        *string        `db:"error"`
}

// CreatePublisherEvent creates a new events that we'll store inside the publisher_events table.
// subject: the resource bound to the event (e.g current user id, etc...)
// type: is the name of the event (e.g `application.created`)
// source: name of the application that created the event
// dataschema: is the JSON-Schema ID of the event (e.g. https://github.com/fewlinesco/myapp/jsonschema/application.created.json)
// data: is the payload of the event itself
func CreatePublisherEvent(ctx context.Context, tx *database.Tx, subject string, eventType string, source string, dataschema string, data interface{}) (Event, error) {
	rawData, err := json.Marshal(data)
	if err != nil {
		return Event{}, fmt.Errorf("can't marshal event: %w", err)
	}

	ev := Event{
		ID:           uuid.New().String(),
		Status:       EventStatusQueued,
		Subject:      subject,
		DataSchema:   dataschema,
		Type:         eventType,
		Source:       source,
		Data:         rawData,
		DispatchedAt: time.Now(),
	}

	_, err = tx.NamedExecContext(ctx, `
		INSERT INTO publisher_events
		(id, status, subject, type, source, dataschema, data, dispatched_at)
		VALUES
		(:id, :status, :subject, :type, :source, :dataschema, :data, :dispatched_at)
	`, ev)

	if err != nil {
		return ev, fmt.Errorf("can't insert: %w", err)
	}

	return ev, nil
}
