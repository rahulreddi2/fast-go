package main

import (
	"context"
	"fast/dbconnectt"
	"fast/dbmodelss"
	"fast/dbwallets"

	"fast/pubsubpublish"
	"fast/subs"
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	defer func() {
		if err := dbconnectt.Db.Close(); err != nil {
			log.Printf("error while closing the connection: %v", err)
		}
	}()

	fmt.Println("starting the application...")

	if err := pubsubpublish.InitPublisher(ctx); err != nil {
		log.Fatalf("error initializing Pub/Sub publisher: %v", err)
	}

	if err := subs.InitSubscriber(ctx); err != nil {
		fmt.Printf("error initializing Pub/Sub subscriber: %v", err)
	}
	subs.StartSubscription(ctx)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		defer func() {
			if r := recover(); r != nil {
				dbwallets.HandleError(c, r.(error), "internal server error")
			}
		}()
		return c.Next()
	})

	app.Post("/newuser", func(ctx *fiber.Ctx) error {
		userID, username, err := dbwallets.Newuser(ctx)
		if err != nil {
			return err
		}

		walletInfo := dbmodelss.WalletPubSubMessage{
			UserID:   userID,
			Username: username,
		}
		pubsubpublish.PublishMessage(ctx.Context(), walletInfo)

		return nil
	})

	app.Post("/creditwallet", func(ctx *fiber.Ctx) error {
		creditInfo, err := dbwallets.CreditWallet(ctx)
		if err != nil {
			log.Println("credit wallet function error", err)
			return err
		}

		walletInfo := dbmodelss.WalletPubSubMessage{
			WalletID: creditInfo.WalletID,
			Balance:  creditInfo.Balance,
			UserID:   creditInfo.UserID,
			Username: creditInfo.Username,
		}
		pubsubpublish.PublishMessage(ctx.Context(), walletInfo)

		return nil
	})

	app.Post("/debitwallet", func(ctx *fiber.Ctx) error {
		debitInfo, err := dbwallets.DebitWallet(ctx)
		if err != nil {
			return err
		}

		walletInfo := dbmodelss.WalletPubSubMessage{
			WalletID: debitInfo.WalletID,
			Balance:  debitInfo.Balance,
			UserID:   debitInfo.UserID,
			Username: debitInfo.Username,
		}
		pubsubpublish.PublishMessage(ctx.Context(), walletInfo)

		return nil
	})

	port := 3000
	err := app.Listen(fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("error starting the server: %v", err)
	}
}
