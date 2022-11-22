package initializers

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

func ConnectToDB() {
	var err error
	dsn := os.Getenv("DSN")
	DB, err = sql.Open("mysql", dsn)

	if err != nil {
		panic("Failed to connect to the DB")
	}
	fmt.Println("Connected to the database")

}
