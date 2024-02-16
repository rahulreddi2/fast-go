package dbmodelss

import (
	"encoding/json"
	"log"

	"github.com/gofiber/fiber/v2"
)

type Wallet struct {
	Wallet_ID int     `json:"wallet_id"`
	Balance   float64 `json:"balance"`
	User_ID   string  `json:"user_id"`
	Username  string  `json:"username"`
}

type Wallet_ledger struct {
	Wallet_ID          int     `json:"wallet_id"`
	Transaction_ID     string  `json:"transaction_id"`
	UserID             string  `json:"user_id"`
	Username           string  `json:"username"`
	PreviousBalance    float64 `json:"previous_balance"`
	Amount             float64 `json:"amount"`
	CurrentBalance     float64 `json:"current_balance"`
	Transaction_Status string  `json:"transaction_status"`
}

type Response struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type WalletPubSubMessage struct {
	WalletID int     `json:"wallet_id"`
	Balance  float64 `json:"balance"`
	UserID   string  `json:"user_id"`
	Username string  `json:"username"`
}

func HandleError(c *fiber.Ctx, err error, message string) {
	statusCode := fiber.StatusInternalServerError
	response := Response{Status: "error", Message: message + ": " + err.Error()}
	log.Println(response.Message)

	c.Status(statusCode)
	if err := json.NewEncoder(c).Encode(response); err != nil {
		log.Println("Error encoding error response:", err)
		c.SendString("Internal server error")
	}
}

func NewResponse(status, message string) *Response {
	return &Response{Status: status, Message: message}
}
