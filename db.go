package main

import (
	"database/sql"
	"math/rand"
	"time"
)

type SQLModel struct {
	db  *sql.DB
	rnd *rand.Rand
}

func NewSQLModel(db *sql.DB) (*SQLModel, error) {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	model := &SQLModel{db, rnd}
	_, err := model.db.Exec(`
		CREATE TABLE IF NOT EXISTS tokens (
			id VARCHAR(12) NOT NULL PRIMARY KEY,
			value VARCHAR(100) NOT NULL,
			name VARCHAR(255) NOT NULL,
      interval INTEGER,
      -- to disable the token temporarely
      disabled BOOLEAN NOT NULL DEFAULT FALSE,
      -- to indicate a token is in a fired state; will go back to false once we get a valid ping again
      fired BOOLEAN NOT NULL DEFAULT FALSE,

			time_created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		  time_deleted TIMESTAMP
		);
		
		CREATE TABLE IF NOT EXISTS pings (
			id INTEGER NOT NULL PRIMARY KEY,
			token_id INTEGER NOT NULL REFERENCES tokens(id),
			time_created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	    time_deleted TIMESTAMP
		);
		
		CREATE INDEX IF NOT EXISTS tokens_list_id ON pings(token_id);
		`)
	return model, err
}

func (m *SQLModel) FooBar() error {
	return nil
}
