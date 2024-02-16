package dbwallets

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fast/dbmodelss"
	golbal "fast/gobal"
	"fmt"
	"math/rand"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

func Newuser(ctx *fiber.Ctx) (string, string, error) {
	if ctx.Method() != fiber.MethodPost {
		return "", "", fmt.Errorf("method not allowed")
	}

	var newUser dbmodelss.Wallet

	if err := json.Unmarshal(ctx.Request().Body(), &newUser); err != nil {
		return "", "", fmt.Errorf("error decoding JSON request body: %v", err)
	}

	var countWalletID, countUsername int
	err := golbal.Wallet_db.QueryRow("SELECT 1 FROM wallet WHERE wallet_id = $1 LIMIT 1", newUser.Wallet_ID).Scan(&countWalletID)
	if err != nil && err != sql.ErrNoRows {
		return "", "", fmt.Errorf("error checking for existing wallet_id: %v", err)
	}

	err = golbal.Wallet_db.QueryRow("SELECT 1 FROM wallet WHERE username = $1 LIMIT 1", newUser.Username).Scan(&countUsername)
	if err != nil && err != sql.ErrNoRows {
		return "", "", fmt.Errorf("error checking for existing username: %v", err)
	}

	if countWalletID > 0 {
		return "", "", errors.New("user with the same wallet_ID already exists")
	}
	if countUsername > 0 {
		return "", "", errors.New("user with the same username already exists")
	}

	result, err := golbal.Wallet_db.Exec("INSERT INTO wallet (wallet_id, balance, user_id, username) VALUES ($1, $2, $3, $4)",
		newUser.Wallet_ID, newUser.Balance, newUser.User_ID, newUser.Username)
	if err != nil {
		return "", "", fmt.Errorf("error creating new user: %v", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return "", "", fmt.Errorf("error getting rows affected: %v", err)
	}

	if affected > 0 {
		ctx.Status(fiber.StatusCreated)
		ctx.JSON(dbmodelss.NewResponse("success", "new user created successfully"))

		return newUser.User_ID, newUser.Username, nil
	}

	return "", "", errors.New("failed to create new user")
}

func CreditWallet(ctx *fiber.Ctx) (*dbmodelss.WalletPubSubMessage, error) {
	if ctx.Method() != fiber.MethodPost {
		return nil, fiber.NewError(fiber.StatusMethodNotAllowed, "Method Not Allowed")
	}

	var creditInfo struct {
		Wallet_ID        int     `json:"wallet_id"`
		User_ID          string  `json:"user_id"`
		Username         string  `json:"username"`
		Amount           float64 `json:"amount"`
		Clients_Trans_ID string  `json:"clients_trans_id"`
	}

	if err := ctx.BodyParser(&creditInfo); err != nil {
		HandleError(ctx, err, "Bad Request")
		return nil, err
	}

	var currentBalance float64
	err := golbal.Wallet_db.QueryRow("SELECT balance FROM wallet WHERE wallet_id = $1 AND user_id = $2 AND username = $3", creditInfo.Wallet_ID, creditInfo.User_ID, creditInfo.Username).Scan(&currentBalance)
	if err != nil {
		handleUserExistenceError(ctx, err, creditInfo)
		return nil, err
	}

	newBalance := currentBalance + creditInfo.Amount
	var count int
	selectErr := golbal.Wallet_db.QueryRow("SELECT COUNT(*) FROM wallet_ledger WHERE Clients_Trans_ID = $1", creditInfo.Clients_Trans_ID).Scan(&count)
	if selectErr != nil {
		fmt.Println("error:", selectErr)
		return nil, selectErr
	}

	if count > 0 {
		HandleError(ctx, nil, "User with the same clients transaction exists, choose another")
		return nil, errors.New("user with the same clients transaction exists, choose another")
	}

	_, err = golbal.Wallet_db.Exec("UPDATE wallet SET balance = $1 WHERE wallet_id = $2", newBalance, creditInfo.Wallet_ID)
	if err != nil {
		HandleError(ctx, err, "Error updating wallet balance")
		return nil, err
	}

	_, err = golbal.Wallet_db.Exec("INSERT INTO wallet_ledger(wallet_id, transaction_id, user_id, username, previous_balance, amount, current_balance, transaction_status, clients_trans_id) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)",
		creditInfo.Wallet_ID, generateTransactionID(), creditInfo.User_ID, creditInfo.Username, currentBalance, creditInfo.Amount, newBalance, "success", creditInfo.Clients_Trans_ID)
	if err != nil {
		HandleError(ctx, err, "Error inserting into wallet_ledger table")
		return nil, err
	}

	response := dbmodelss.NewResponse("success", fmt.Sprintf("Wallet credited with %.6f rupees. New balance: %.6f Your transaction_id: %s", creditInfo.Amount, newBalance, creditInfo.Clients_Trans_ID))

	ctx.Status(fiber.StatusOK).JSON(response)

	return &dbmodelss.WalletPubSubMessage{
		WalletID: creditInfo.Wallet_ID,
		Balance:  newBalance,
		UserID:   creditInfo.User_ID,
		Username: creditInfo.Username,
	}, nil
}

func handleUserExistenceError(ctx *fiber.Ctx, err error, creditInfo struct {
	Wallet_ID        int     `json:"wallet_id"`
	User_ID          string  `json:"user_id"`
	Username         string  `json:"username"`
	Amount           float64 `json:"amount"`
	Clients_Trans_ID string  `json:"clients_trans_id"`
}) error {
	return HandleError(ctx, err, "Error checking user existence")
}

func DebitWallet(ctx *fiber.Ctx) (*dbmodelss.WalletPubSubMessage, error) {
	if ctx.Method() != fiber.MethodPost {
		return nil, fiber.NewError(fiber.StatusMethodNotAllowed, "Method Not Allowed")
	}

	var debitInfo struct {
		Wallet_ID        int     `json:"wallet_id"`
		User_ID          string  `json:"user_id"`
		Username         string  `json:"username"`
		Amount           float64 `json:"amount"`
		Clients_Trans_ID string  `json:"clients_trans_id"`
	}

	if err := ctx.BodyParser(&debitInfo); err != nil {
		return nil, fiber.NewError(fiber.StatusBadRequest, "Bad Request")
	}

	var currentBalance float64
	err := golbal.Wallet_db.QueryRow("SELECT balance FROM wallet WHERE wallet_id = $1 AND user_id = $2 AND username = $3", debitInfo.Wallet_ID, debitInfo.User_ID, debitInfo.Username).Scan(&currentBalance)
	if err != nil {
		return nil, handleDebitUserExistenceError(ctx, err, debitInfo)
	}

	newBalance := currentBalance - debitInfo.Amount
	var count int
	selectErr := golbal.Wallet_db.QueryRow("SELECT COUNT(*) FROM wallet_ledger WHERE Clients_Trans_ID = $1", debitInfo.Clients_Trans_ID).Scan(&count)
	if selectErr != nil {
		fmt.Println("error:", selectErr)
		return nil, HandleError(ctx, selectErr, "Error checking clients transaction existence")
	}

	if count > 0 {
		return nil, HandleError(ctx, errors.New("user with the same clients transaction exists, choose another"), "user with the same clients transaction exists, choose another")
	}

	_, err = golbal.Wallet_db.Exec("UPDATE wallet SET balance = $1 WHERE wallet_id = $2", newBalance, debitInfo.Wallet_ID)
	if err != nil {
		return nil, HandleError(ctx, err, "Error updating wallet balance")
	}

	_, err = golbal.Wallet_db.Exec("INSERT INTO wallet_ledger(wallet_id, transaction_id, user_id, username, previous_balance, amount, current_balance, transaction_status, clients_trans_id) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)",
		debitInfo.Wallet_ID, generateTransactionID(), debitInfo.User_ID, debitInfo.Username, currentBalance, debitInfo.Amount, newBalance, "success", debitInfo.Clients_Trans_ID)
	if err != nil {
		return nil, HandleError(ctx, err, "Error inserting into wallet_ledger table")
	}

	response := dbmodelss.NewResponse("success", fmt.Sprintf("Wallet debited with %.6f rupees. New balance: %.6f Your transaction_id: %s", debitInfo.Amount, newBalance, debitInfo.Clients_Trans_ID))

	ctx.Status(fiber.StatusOK).JSON(response)

	return &dbmodelss.WalletPubSubMessage{
		WalletID: debitInfo.Wallet_ID,
		Balance:  newBalance,
		UserID:   debitInfo.User_ID,
		Username: debitInfo.Username,
	}, nil
}

func HandleError(ctx *fiber.Ctx, err error, message string) error {
	if err != nil {
		dbmodelss.HandleError(ctx, err, message)
		return err
	}

	fmt.Println("Non-fatal error:", message)
	return nil
}

func handleDebitUserExistenceError(ctx *fiber.Ctx, err error, debitInfo struct {
	Wallet_ID        int     `json:"wallet_id"`
	User_ID          string  `json:"user_id"`
	Username         string  `json:"username"`
	Amount           float64 `json:"amount"`
	Clients_Trans_ID string  `json:"clients_trans_id"`
}) error {
	var crtwallet dbmodelss.Wallet
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return HandleError(ctx, err, "Error checking user existence or fetching current balance")
	} else {
		errCheck := golbal.Wallet_db.QueryRow("SELECT * FROM wallet WHERE wallet_id = $1", debitInfo.Wallet_ID).Scan(&crtwallet.Wallet_ID, &crtwallet.Balance, &crtwallet.User_ID, &crtwallet.Username)

		if errCheck != nil {
			return HandleError(ctx, errCheck, "Error checking user existence")
		}

		if debitInfo.Username != crtwallet.Username {
			errorMessage := fmt.Sprintf("Incorrect username (%s). Please check the username and try again.", debitInfo.Username)
			return HandleError(ctx, errors.New(errorMessage), errorMessage)
		} else if debitInfo.User_ID != crtwallet.User_ID {
			errorMessage := fmt.Sprintf("Incorrect user_ID (%s). Please check the user ID and try again.", debitInfo.User_ID)
			return HandleError(ctx, errors.New(errorMessage), errorMessage)
		} else {
			return HandleError(ctx, nil, "Please enter the proper credentials")
		}
	}
}

func generateTransactionID() string {

	return strconv.FormatInt(rand.Int63(), 10)
}
