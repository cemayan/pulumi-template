package pubsubproducer

import (
	"cloud.google.com/go/pubsub"
	"context"
	"encoding/json"
	"fmt"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"log"
	"net/http"
	"os"
)

var (
	ctx       = context.Background()
	topic     *pubsub.Topic
	projectId = os.Getenv("PROJECT_ID")
	topicId   = os.Getenv("TOPIC_ID")
)

func init() {
	// Create a Pub/Sub client
	client, err := pubsub.NewClient(ctx, projectId)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	topic = client.Topic(topicId)
	functions.HTTP("PubsubProducer", PubsubProducer)
}

type Payload struct {
	GameName  string          `json:"game_name"`
	EventName string          `json:"event_name"`
	EventData json.RawMessage `json:"event_data"`
}

// PubsubProducer is an HTTP Cloud Function with a request parameter.
func PubsubProducer(w http.ResponseWriter, r *http.Request) {
	var payload Payload

	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	marshal, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	msg := &pubsub.Message{
		Data: marshal,
	}

	//topic.PublishSettings = pubsub.PublishSettings{
	//	CountThreshold: 0,
	//}

	if _, err := topic.Publish(ctx, msg).Get(ctx); err != nil {
		http.Error(w, fmt.Sprintf("Could not publish message: %v", err), 500)
		return
	}
}
