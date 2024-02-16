package subs

import (
	"context"
	"encoding/json"
	"fast/dbmodelss"
	"fast/pubc"
	"fmt"
	"log"

	"cloud.google.com/go/pubsub"
)

const (
	projectID   = "rahulreddy2139"
	topicID     = "mytopic"
	subName     = "mysub"
	maxMessages = 1
)

var subscription *pubsub.Subscription

func InitSubscriber(ctx context.Context) error {
	var err error
	pubsubClient, err := pubc.CreatePubSubClient(ctx, projectID)
	if err != nil {
		log.Printf("Failed to create Pub/Sub client: %v", err)
		return err
	}

	subscription = pubsubClient.Subscription(subName)
	ok, err := subscription.Exists(ctx)
	if err != nil {
		log.Printf("Failed to check if subscription exists: %v", err)
		return err
	}

	if !ok {
		_, err := pubsubClient.CreateSubscription(ctx, subName, pubsub.SubscriptionConfig{
			Topic: pubsubClient.Topic(topicID),
		})
		if err != nil {
			log.Printf("Failed to create subscription: %v", err)
			return err
		}
		log.Printf("Subscription %s created.\n", subName)
	} else {
		log.Printf("Subscription %s already exists.\n", subName)
	}

	return nil
}

func StartSubscription(ctx context.Context) {
	cctx, cancel := context.WithCancel(ctx)
	defer cancel()
	fmt.Println("i am there ")
	go func() {
		err := subscription.Receive(cctx, func(ctx context.Context, msg *pubsub.Message) {
			var walletInfo dbmodelss.WalletPubSubMessage
			if err := json.Unmarshal(msg.Data, &walletInfo); err != nil {
				log.Printf("Error decoding JSON: %v", err)
				msg.Nack()
				return
			}
			log.Println("Received message:", walletInfo)
			msg.Ack()
		})

		if err != nil {
			log.Printf("Error receiving message: %v", err)
		}
	}()
}
