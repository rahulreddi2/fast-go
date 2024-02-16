package pubsubpublish

import (
	"context"
	"encoding/json"
	"fast/pubc"
	"fmt"
	"log"

	"cloud.google.com/go/pubsub"
)

const (
	projectID = "rahulreddy2139"
	topicID   = "mytopic"
)

var pubsubClient *pubsub.Client
var topic *pubsub.Topic

func InitPublisher(ctx context.Context) error {
	var err error
	pubsubClient, err = pubc.CreatePubSubClient(ctx, projectID)
	if err != nil {
		log.Printf("Failed to create Pub/Sub client: %v", err)
		return err
	}

	topic = pubsubClient.Topic(topicID)
	ok, err := topic.Exists(ctx)
	if err != nil {
		log.Printf("Failed to check if topic exists: %v", err)
		return err
	}
	if !ok {
		if _, err := pubsubClient.CreateTopic(ctx, topicID); err != nil {
			log.Printf("Failed to create topic: %v", err)
			return err
		}
		log.Printf("Topic %s created.\n", topicID)
	} else {
		log.Printf("Topic %s already exists.\n", topicID)
	}

	return nil
}

func PublishMessage(ctx context.Context, message interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error encoding JSON: %v", err)
		return err
	}

	result := topic.Publish(ctx, &pubsub.Message{
		Data: data,
	})
	if _, err := result.Get(ctx); err != nil {
		log.Printf("Error publishing message: %v", err)
		return err
	}

	fmt.Printf("Published message to topic %s: %s\n", topicID, data)

	return nil
}
