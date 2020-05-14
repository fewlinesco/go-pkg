package eventing

import (
	"context"
	"fmt"
	"log"

	"github.com/cloudevents/sdk-go/v2/client"
	"github.com/cloudevents/sdk-go/v2/event"
)

func ListenAndSave(client client.Client) {
	for {
		if err := client.StartReceiver(context.Background(), saveReceivedEvent); err != nil {
			log.Printf("failed to start nats receiver, %s", err.Error())
		}
	}
}

func saveReceivedEvent(ctx context.Context, event event.Event) {
	fmt.Printf("saveReceivedEvent: %v\n", event)
}
