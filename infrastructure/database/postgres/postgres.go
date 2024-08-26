package postgres

import (
	"database/sql"
	"log"
)

func New(connection string) *sql.DB {
	db, err := sql.Open("postgres", connection)
	if err != nil {
		log.Fatal(err)
	}
	return db
}
