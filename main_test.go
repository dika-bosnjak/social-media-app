package main

import (
	"database/sql"
	"testing"
)

const (
	dbDriver = "mysql"
	dbSource = "dika:12345678@tcp(localhost:3306)/social_media?charset=utf8mb4&parseTime=True&loc=Local"
)

var DB *sql.DB

func Test_main(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "connect to the database"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn, err := sql.Open(dbDriver, dbSource)
			if err != nil {
				t.Errorf("FAILED: could not connect to the database")
			}

			DB = conn

			main()
		})
	}
}
