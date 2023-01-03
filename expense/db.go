package expense

import (
	"database/sql"
	"log"

	"github.com/brown-kaew/assessment/config"
	_ "github.com/lib/pq"
)

func InitDB(conf config.Config) (*sql.DB, func()) {
	db, err := sql.Open("postgres", conf.DatabaseUrl)
	if err != nil {
		log.Fatal("connect to database error", err)
	}

	initTable(db)

	return db, func() { db.Close() }
}

func initTable(db *sql.DB) {
	createTable := `
	CREATE TABLE IF NOT EXISTS expenses (
		id SERIAL PRIMARY KEY,
		title TEXT,
		amount FLOAT,
		note TEXT,
		tags TEXT[]
	);`

	if _, err := db.Exec(createTable); err != nil {
		log.Fatal("can't create table", err)
	}
}
