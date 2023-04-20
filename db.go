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

type List []*Token

type Token struct {
	ID       string
	Name     string
	Value    bool
	Interval int
	Disabled bool
	Fired    bool
}

func NewSQLModel(db *sql.DB) (*SQLModel, error) {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	model := &SQLModel{db, rnd}
	_, err := model.db.Exec(`
		CREATE TABLE IF NOT EXISTS tokens (
			id VARCHAR(20) NOT NULL PRIMARY KEY,
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

// Create a token and return the id which identifies the token uniquely
func (m *SQLModel) CreateToken(name string, interval int) (string, error) {
	id := m.makeTokenID(20)
	// Generate time here because SQLite's CURRENT_TIMESTAMP only returns seconds.
	timeCreated := time.Now().In(time.UTC).Format(time.RFC3339Nano)
	_, err := m.db.Exec("INSERT INTO tokens (id, name, interval, time_created) VALUES (?, ?, ?, ?)",
		id, name, interval, timeCreated)
	return id, err
}

var listIDChars = "bcdfghjklmnpqrstvwxyz"

func (m *SQLModel) makeTokenID(n int) string {
	id := make([]byte, n)
	for i := 0; i < n; i++ {
		index := m.rnd.Intn(len(listIDChars))
		id[i] = listIDChars[index]
	}
	return string(id)
}
