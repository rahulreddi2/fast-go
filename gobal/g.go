package golbal

import (
	"database/sql"
	"fast/dbconnectt"
	"fmt"
)

var Wallet_db *sql.DB

func init() {
	fmt.Println("Inside init function .... ")
	var err error
	Wallet_db, err = dbconnectt.GetPostgreSQLDB()

	if err != nil {

		panic(fmt.Sprintf("Failed to connect to database: %v", err))
	}
}
