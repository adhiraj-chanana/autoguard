package main

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite"
)

var db *sql.DB

func initDB() {
	var err error
	db, err = sql.Open("sqlite", "./autoguard.db")
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}

	// Create tables
	createCommitTable := `
	CREATE TABLE IF NOT EXISTS commits (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		commit_id TEXT,
		repo_url TEXT,
		status TEXT,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	createIssuesTable := `
	CREATE TABLE IF NOT EXISTS issues (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		commit_id TEXT,
		type TEXT,
		filename TEXT,
		line INTEGER,
		message TEXT,
		retries INTEGER
	);`

	_, err = db.Exec(createCommitTable)
	if err != nil {
		log.Fatal("Failed to create commits table:", err)
	}

	_, err = db.Exec(createIssuesTable)
	if err != nil {
		log.Fatal("Failed to create issues table:", err)
	}
}
