package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"

	_ "modernc.org/sqlite"
)

func main() {
	// Config defaults
	port := 3500
	dbPath := "keep-an-eye.sqlite"

	db, err := sql.Open("sqlite", fmt.Sprintf("file:%s?_foreign_keys=on", dbPath))
	exitOnError(err)
	model, err := NewSQLModel(db)
	exitOnError(err)
	server, err := NewServer(model, log.Default())
	exitOnError(err)

	log.Printf("listening on http://localhost:%d", port)
	http.ListenAndServe("127.0.0.1:3500", server)
	err = http.ListenAndServe(":"+strconv.Itoa(port), server)
	exitOnError(err)
}

func exitOnError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
