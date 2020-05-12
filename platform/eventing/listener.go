package eventing

import (
	"context"
	"fmt"

	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/cloudevents/sdk-go/v2/client"
	"github.com/jmoiron/sqlx"
)

func ListenAndSave(client client.Client, tx *sqlx.Tx) {
	client.StartReceiver(context.Background(), saveReceivedEvent(tx))
}

func saveReceivedEvent(tx *sqlx.Tx) func(cloudevents.Event) {
	return func(event cloudevents.Event) {
		fmt.Printf("I received the event: %v\n", event)
	}
}
