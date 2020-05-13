package eventing

import (
	"context"
	"fmt"

	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/cloudevents/sdk-go/v2/client"
)

func ListenAndSave(client client.Client) {
	fmt.Printf("ListenAndSave\n")
	for {
		client.StartReceiver(context.Background(), saveReceivedEvent)
	}
}

func saveReceivedEvent(event cloudevents.Event) {
	fmt.Printf("saveReceivedEvent: %v\n", event)
}
