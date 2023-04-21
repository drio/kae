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

type Model interface {
	CreateToken(string, string, int) (string, error)
	GetTokens() (ListTokens, error)
	GetIdFromToken(string) (int, error)
	InsertHeartBeat(int) error
	LastHeartBeat(int) (time.Time, error)
	Fire(int, bool) error
	Disable(int, bool) error
	Remove(int) error
}

type ListTokens []*Token

type Token struct {
	ID          int
	Token       string
	Name        string
	Description string
	Interval    int
	Disabled    bool
	Fired       bool
	TimeCreated time.Time
}

func NewSQLModel(db *sql.DB) (*SQLModel, error) {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	model := &SQLModel{db, rnd}
	_, err := model.db.Exec(`
		CREATE TABLE IF NOT EXISTS tokens (
			id INTEGER NOT NULL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			token VARCHAR(20) NOT NULL,
      interval INTEGER,
			description VARCHAR(1000) NOT NULL,

      -- to disable the token temporarely
      disabled BOOLEAN NOT NULL DEFAULT TRUE,
      -- to indicate a token is in a fired state; will go back to false once we get a valid ping again
      fired BOOLEAN NOT NULL DEFAULT FALSE,

			time_created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		  time_deleted TIMESTAMP
		);
		
		CREATE TABLE IF NOT EXISTS pings (
			id INTEGER NOT NULL PRIMARY KEY,
			token_id INTEGER NOT NULL REFERENCES tokens(id),
			last_heartbeat TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
		
		CREATE INDEX IF NOT EXISTS tokens_list_id ON pings(token_id);
		`)
	return model, err
}

// Create a token and return the id which identifies the token uniquely
func (m *SQLModel) CreateToken(name, description string, interval int) (string, error) {
	token := m.makeTokenID(20)
	// Generate time here because SQLite's CURRENT_TIMESTAMP only returns seconds.
	timeCreated := time.Now().In(time.UTC).Format(time.RFC3339Nano)
	_, err := m.db.Exec(`INSERT INTO tokens 
    (token, name, interval, time_created, description) 
    VALUES (?, ?, ?, ?, ?)`,
		token, name, interval, timeCreated, description)
	return token, err
}

// GetLists fetches all the tokens  ordered with the most recent first.
func (m *SQLModel) GetTokens() (ListTokens, error) {
	rows, err := m.db.Query(`
		SELECT id, token, name, interval, disabled, fired, time_created, description
		FROM tokens
    WHERE time_deleted is NULL
		ORDER BY time_created DESC
		`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var listTokens ListTokens
	for rows.Next() {
		var t Token
		err = rows.Scan(&t.ID, &t.Token, &t.Name, &t.Interval, &t.Disabled, &t.Fired, &t.TimeCreated, &t.Description)
		if err != nil {
			return nil, err
		}
		listTokens = append(listTokens, &t)
	}
	return listTokens, rows.Err()
}

// Number of seconds since last heartbeat
func (m *SQLModel) LastHeartBeat(tokenId int) (time.Time, error) {
	rows, err := m.db.Query(`
    SELECT p.last_heartbeat
    FROM tokens as t
    JOIN pings as p
      ON t.id = p.token_id
    WHERE t.id = ?
    ORDER BY last_heartbeat
    DESC limit 1
    `, tokenId)
	if err != nil {
		return time.Time{}, err
	}
	defer rows.Close()

	var lastHB time.Time
	for rows.Next() {
		err = rows.Scan(&lastHB)
		if err != nil {
			return time.Time{}, err
		}
	}
	return lastHB, nil
}

func (m *SQLModel) Fire(id int, b bool) error {
	_, err := m.db.Exec("UPDATE tokens SET fired = ? WHERE id = ?", b, id)
	return err
}

func (m *SQLModel) Disable(id int, b bool) error {
	_, err := m.db.Exec("UPDATE tokens SET disabled = ? WHERE id = ?", b, id)
	return err
}

func (m *SQLModel) InsertHeartBeat(id int) error {
	_, err := m.db.Exec("INSERT INTO pings (token_id) VALUES (?)", id)
	return err
}

func (m *SQLModel) Remove(id int) error {
	_, err := m.db.Exec(`
			UPDATE tokens
			SET time_deleted = CURRENT_TIMESTAMP
			WHERE id = ?
		`, id)
	return err
}

func (m *SQLModel) GetIdFromToken(token string) (int, error) {
	rows, err := m.db.Query("select id from tokens WHERE token = ?", token)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var id int
	for rows.Next() {
		err = rows.Scan(&id)
		if err != nil {
			return 0, err
		}
	}

	return id, nil
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
