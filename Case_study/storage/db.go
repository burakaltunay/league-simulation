package storage

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"io/ioutil"
	"log"
)

var db *sql.DB

func InitDB(filepath string, schemaPath string) {
	var err error
	db, err = sql.Open("sqlite3", filepath)
	if err != nil {
		log.Fatalf("Failed to open DB: %v", err)
	}
	// Read and execute schema
	schema, err := ioutil.ReadFile(schemaPath)
	if err != nil {
		log.Fatalf("Failed to read schema: %v", err)
	}
	_, err = db.Exec(string(schema))
	if err != nil {
		log.Fatalf("Failed to execute schema: %v", err)
	}
}

func GetDB() *sql.DB {
	return db
} 