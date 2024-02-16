package pubc

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"cloud.google.com/go/pubsub"
)

func CreatePubSubClient(ctx context.Context, projectID string) (*pubsub.Client, error) {

	keyFilePath := "./rahulreddy2139-f59c942551df.json"
	absKeyFilePath, err := filepath.Abs(keyFilePath)
	if err != nil {
		log.Fatalf("Couldn't get absolute path: %v", err)
		return nil, err
	}

	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", absKeyFilePath)

	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		log.Printf("Failed to create Pub/Sub client: %v", err)
		return nil, err
	}

	fmt.Println("Client created for our sake")
	return client, nil
}
