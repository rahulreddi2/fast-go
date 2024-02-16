package dbconnectt

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

var Db *sql.DB
var ErrNoRows = errors.New("no rows found")

func GetPostgreSQLDB() (*sql.DB, error) {
	db, err := sql.Open("postgres", "postgres://postgres:Rahulreddy@7@localhost:2669/rahul?sslmode=disable")
	if err != nil {
		return nil, err
	}

	Db = db
	Db.SetConnMaxLifetime(time.Minute * 5)
	Db.SetMaxOpenConns(100)

	if err := Db.Ping(); err != nil {
		return nil, err
	}

	fmt.Println("Connected to PostgreSQL database")
	return Db, nil
}
